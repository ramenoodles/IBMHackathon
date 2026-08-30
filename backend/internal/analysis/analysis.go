package analysis

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ramenoodles/IBMHackathon/backend/internal/search"
	"github.com/ramenoodles/IBMHackathon/backend/internal/source"
)

const (
	MaxRootNodes   = 50
	MaxExpandNodes = 30
	MaxDepth       = 8
)

type Node struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	Kind         string `json:"kind"`
	File         string `json:"file"`
	CalleeSymbol string `json:"calleeSymbol,omitempty"`
	CalleeFile   string `json:"calleeFile,omitempty"`
	Line         int    `json:"line"`
	CalleeLine   int    `json:"calleeLine,omitempty"`
	ChildCount   int    `json:"childCount"`
	Expandable   bool   `json:"expandable"`
	Collapsed    bool   `json:"collapsed"`
	Confidence   string `json:"confidence"`
	Code         string `json:"code,omitempty"`
}

type Edge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label,omitempty"`
}

type Graph struct {
	RootID string `json:"rootId"`
	Nodes  []Node `json:"nodes"`
	Edges  []Edge `json:"edges"`
	Depth  int    `json:"depth"`
	Symbol string `json:"symbol"`
}

type Builder struct {
	finder *search.Finder
	reader *source.Reader
}

func New(root, rg string) (*Builder, error) {
	return NewWithLimit(root, rg, source.DefaultMaxFileBytes)
}

// NewWithLimit is New with an explicit per-file size cap passed to the reader.
func NewWithLimit(root, rg string, maxFileBytes int64) (*Builder, error) {
	reader, err := source.NewReaderWithLimit(root, maxFileBytes)
	if err != nil {
		return nil, err
	}
	return &Builder{finder: search.NewFinder(rg), reader: reader}, nil
}

// Root builds a bounded control-flow graph for symbol in the requested file.
func (b *Builder) Root(ctx context.Context, file, symbol string) (Graph, error) {
	if strings.TrimSpace(file) == "" || strings.TrimSpace(symbol) == "" {
		return Graph{}, fmt.Errorf("filePath and symbol are required")
	}
	return b.buildFunction(ctx, filepath.ToSlash(filepath.Clean(file)), symbol, 1, MaxRootNodes, true)
}

// Expand returns a bounded callee CFG fragment. The existing call node remains
// the fragment root so the frontend can merge the result without replacing it.
// When calleeFile and calleeSymbol are provided they are used directly so the
// caller CFG does not need to be rebuilt (avoids lookup failures on merged graphs).
func (b *Builder) Expand(ctx context.Context, nodeID string, limit int, calleeFile, calleeSymbol string) (Graph, error) {
	meta, err := parseNodeID(nodeID)
	if err != nil {
		return Graph{}, err
	}
	if meta.Depth >= MaxDepth {
		return Graph{}, fmt.Errorf("max flow depth reached")
	}

	callFile := strings.TrimSpace(calleeFile)
	callSym := strings.TrimSpace(calleeSymbol)
	if callFile != "" && callSym != "" {
		callFile = filepath.ToSlash(filepath.Clean(callFile))
	} else {
		container, buildErr := b.buildFunction(ctx, meta.File, meta.Symbol, meta.Depth, MaxRootNodes, false)
		if buildErr != nil {
			return Graph{}, buildErr
		}
		var call *Node
		for i := range container.Nodes {
			if container.Nodes[i].ID == nodeID {
				call = &container.Nodes[i]
				break
			}
		}
		if call == nil {
			return Graph{}, fmt.Errorf("flow node %q not found", nodeID)
		}
		if !call.Expandable || call.CalleeFile == "" || call.CalleeSymbol == "" {
			return Graph{}, fmt.Errorf("flow node %q is not expandable", nodeID)
		}
		callFile = call.CalleeFile
		callSym = call.CalleeSymbol
	}

	if limit <= 0 || limit > MaxExpandNodes {
		limit = MaxExpandNodes
	}
	fragment, err := b.buildFunction(ctx, callFile, callSym, meta.Depth+1, limit, false)
	if err != nil {
		return Graph{}, err
	}
	if len(fragment.Nodes) == 0 {
		return Graph{}, fmt.Errorf("callee %q has no flow steps", callSym)
	}
	fragment.Edges = append([]Edge{{From: nodeID, To: fragment.RootID, Label: "calls"}}, fragment.Edges...)
	fragment.RootID = nodeID
	return fragment, nil
}

func (b *Builder) buildFunction(ctx context.Context, file, symbol string, depth, limit int, truncationMarker bool) (Graph, error) {
	if err := ctx.Err(); err != nil {
		return Graph{}, err
	}
	content, err := b.reader.ReadFile(file)
	if err != nil {
		return Graph{}, err
	}
	steps, err := extractFlow(content, file, symbol)
	if err != nil {
		return Graph{}, err
	}
	graph := buildCFG(file, symbol, depth, steps, limit, truncationMarker)
	if depth < MaxDepth {
		b.resolveCalls(ctx, &graph, content)
	}
	return graph, nil
}

func (b *Builder) resolveCalls(ctx context.Context, graph *Graph, currentContent string) {
	language := LanguageFromPath(graph.Nodes[0].File)
	cache := make(map[string]calleeTarget)
	for i := range graph.Nodes {
		node := &graph.Nodes[i]
		if node.Kind != "call" || node.CalleeSymbol == "" {
			continue
		}
		if isInstanceMethodCall(node.Label) {
			continue
		}
		cacheKey := fmt.Sprintf("%s::%s::%d", node.File, node.Label, node.Line)
		target, ok := cache[cacheKey]
		if !ok {
			target = b.resolveCallee(ctx, node.File, currentContent, node.CalleeSymbol, language, node.Label)
			cache[cacheKey] = target
		}
		if target.File == "" || target.ChildCount == 0 {
			continue
		}
		node.CalleeFile = target.File
		node.CalleeLine = target.Line
		node.ChildCount = target.ChildCount
		node.Expandable = true
		node.Collapsed = true
	}
}

type calleeTarget struct {
	File       string
	Line       int
	ChildCount int
}

func (b *Builder) resolveCallee(ctx context.Context, currentFile, currentContent, symbol, language, label string) calleeTarget {
	qualified := strings.TrimSuffix(label, "()")
	if !strings.Contains(qualified, ".") {
		if steps, err := extractFlow(currentContent, currentFile, symbol); err == nil {
			return calleeTarget{File: currentFile, Line: steps[0].Line, ChildCount: flowChildCount(steps)}
		}
	}

	matches, err := b.finder.Find(ctx, search.Query{
		Name: symbol, Root: b.reader.Root(), Language: language, Limit: 10,
	})
	if err != nil {
		return calleeTarget{}
	}
	for _, match := range orderCalleeMatches(matches, qualified) {
		content, readErr := b.reader.ReadFile(match.Path)
		if readErr != nil {
			continue
		}
		steps, flowErr := extractFlow(content, match.Path, symbol)
		if flowErr == nil {
			return calleeTarget{File: filepath.ToSlash(match.Path), Line: steps[0].Line, ChildCount: flowChildCount(steps)}
		}
	}
	return calleeTarget{}
}

// orderCalleeMatches prefers paths that match a package prefix from a qualified call label.
func orderCalleeMatches(matches []search.Match, qualified string) []search.Match {
	parts := strings.Split(qualified, ".")
	if len(parts) < 2 {
		return matches
	}
	pkg := parts[0]
	var preferred, rest []search.Match
	for _, match := range matches {
		path := filepath.ToSlash(match.Path)
		dir := filepath.ToSlash(filepath.Dir(path))
		if strings.HasSuffix(dir, "/"+pkg) || strings.Contains(path, "/"+pkg+"/") {
			preferred = append(preferred, match)
		} else {
			rest = append(rest, match)
		}
	}
	return append(preferred, rest...)
}

func flowChildCount(steps []flowStep) int {
	if len(steps) <= 1 {
		return 0
	}
	return len(steps) - 1
}

type nodeMetadata struct {
	Depth   int
	File    string
	Symbol  string
	Line    int
	Ordinal int
}

func makeNodeID(file, symbol string, depth, line, ordinal int) string {
	encode := base64.RawURLEncoding.EncodeToString
	return fmt.Sprintf("flow:%d:%s:%s:%d:%d", depth, encode([]byte(filepath.ToSlash(file))), encode([]byte(symbol)), line, ordinal)
}

func parseNodeID(id string) (nodeMetadata, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 6 || parts[0] != "flow" {
		return nodeMetadata{}, fmt.Errorf("invalid flow node id")
	}
	depth, depthErr := strconv.Atoi(parts[1])
	line, lineErr := strconv.Atoi(parts[4])
	ordinal, ordinalErr := strconv.Atoi(parts[5])
	file, fileErr := base64.RawURLEncoding.DecodeString(parts[2])
	symbol, symbolErr := base64.RawURLEncoding.DecodeString(parts[3])
	if depthErr != nil || lineErr != nil || ordinalErr != nil || fileErr != nil || symbolErr != nil || depth < 1 || line < 1 || ordinal < 0 || len(file) == 0 || len(symbol) == 0 {
		return nodeMetadata{}, fmt.Errorf("invalid flow node id")
	}
	return nodeMetadata{Depth: depth, File: string(file), Symbol: string(symbol), Line: line, Ordinal: ordinal}, nil
}
