package graph

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ibmhackathon/onbober/internal/llm"
	"github.com/ibmhackathon/onbober/internal/scanner"
	"github.com/ibmhackathon/onbober/internal/scanner/flow"
)

// Builder orchestrates scan-first graph construction and LLM enrichment.
type Builder struct {
	scanner *scanner.Scanner
	llm     *llm.OllamaClient
	cache   *Cache
	enrich  *EnrichCache
}

// NewBuilder creates a graph builder with scanner and LLM clients.
func NewBuilder(s *scanner.Scanner, client *llm.OllamaClient) *Builder {
	return &Builder{
		scanner: s,
		llm:     client,
		cache:   NewCache(10 * time.Minute),
		enrich:  NewEnrichCache(30 * time.Minute),
	}
}

// BuildRoot constructs the scan-first flow graph for a symbol.
func (b *Builder) BuildRoot(ctx context.Context, input BuildInput) (FlowGraph, error) {
	key := CacheKey(input.WorkspacePath, input.Symbol, input.FilePath, "", nil, 0)
	if cached, ok := b.cache.Get(key); ok {
		return cached, nil
	}

	g := b.buildScanGraph(input)
	b.cache.Set(key, g)
	return g, nil
}

// BuildExpand constructs child nodes by scanning a resolved callee symbol.
func (b *Builder) BuildExpand(ctx context.Context, input BuildInput) (FlowGraph, error) {
	if cycleDetected(input.NodeID, input.ParentPath) {
		return FlowGraph{}, fmt.Errorf("cycle detected in expansion path")
	}
	if input.Depth >= MaxDepth {
		return FlowGraph{}, fmt.Errorf("max depth exceeded")
	}

	limit := input.ExpandLimit
	if limit <= 0 || limit > MaxExpandNodes {
		limit = MaxExpandNodes
	}

	key := CacheKey(input.WorkspacePath, input.Symbol, input.FilePath, input.NodeID, input.ParentPath, input.Depth)
	if cached, ok := b.cache.Get(key); ok {
		return cached, nil
	}

	parent, parentGraph, ok := b.loadParentGraph(ctx, input)
	if !ok {
		return FlowGraph{}, fmt.Errorf("parent node not found")
	}

	expandSymbol := parent.CalleeSymbol
	expandFile := parent.CalleeFile
	if expandSymbol == "" {
		expandSymbol = input.Symbol
		expandFile = input.FilePath
	}
	if expandFile == "" {
		expandFile = input.FilePath
	}

	expandInput := input
	expandInput.FilePath = expandFile
	expandInput.Symbol = expandSymbol
	g := b.buildScanGraph(expandInput)
	g.RootID = input.NodeID
	g.Depth = input.Depth
	g = trimExpandGraph(g, limit)
	g = reRootFragment(g, input.NodeID, parentGraph)

	b.cache.Set(key, g)
	return g, nil
}

func (b *Builder) loadParentGraph(ctx context.Context, input BuildInput) (FlowNode, FlowGraph, bool) {
	parentGraph, _ := b.BuildRoot(ctx, input)
	node, ok := FindNodeByID(parentGraph, input.NodeID)
	return node, parentGraph, ok
}

func reRootFragment(fragment FlowGraph, parentID string, parentGraph FlowGraph) FlowGraph {
	if len(fragment.Nodes) == 0 {
		return fragment
	}
	// Connect parent to first child of expanded subgraph.
	firstChild := fragment.Nodes[0].ID
	if fragment.Nodes[0].Kind == "entry" && len(fragment.Nodes) > 1 {
		firstChild = fragment.Nodes[1].ID
	}
	fragment.Edges = append([]FlowEdge{{From: parentID, To: firstChild, Label: "expand"}}, fragment.Edges...)
	_ = parentGraph
	return fragment
}

func trimExpandGraph(g FlowGraph, limit int) FlowGraph {
	if len(g.Nodes) <= limit {
		return g
	}
	allowed := map[string]bool{}
	for i := 0; i < limit && i < len(g.Nodes); i++ {
		allowed[g.Nodes[i].ID] = true
	}
	var nodes []FlowNode
	var edges []FlowEdge
	for _, n := range g.Nodes {
		if allowed[n.ID] {
			nodes = append(nodes, n)
		}
	}
	for _, e := range g.Edges {
		if allowed[e.From] && allowed[e.To] {
			edges = append(edges, e)
		}
	}
	g.Nodes = nodes
	g.Edges = edges
	return g
}

func (b *Builder) buildScanGraph(input BuildInput) FlowGraph {
	if input.FilePath == "" || input.Symbol == "" {
		return MockRootGraph(input.Symbol, input.FilePath)
	}

	content, lang, err := b.scanner.ReadFile(input.WorkspacePath, input.FilePath)
	if err != nil {
		return MockRootGraph(input.Symbol, input.FilePath)
	}

	steps := flow.ExtractFlow(content, input.FilePath, input.Symbol, lang)
	if len(steps) < 1 {
		return FlowGraph{Symbol: input.Symbol, Depth: 1, Mock: true}
	}

	g := BuildCFGGraph(input.Symbol, input.FilePath, steps)
	g = AttachCalleeHints(g, input.WorkspacePath, b.scanner, lang)
	return g
}

// BuildNodeDetail returns an expanded explanation for a node.
func (b *Builder) BuildNodeDetail(ctx context.Context, input BuildInput, nodeID, title string, line int, confidence Confidence, code string) (NodeDetail, error) {
	snippet := code
	if snippet == "" {
		snippet, _, _, _, _ = b.gatherContext(input)
	}
	raw, err := b.llm.GenerateNodeDetail(ctx, llm.UserContext{
		PrimaryLanguage: input.Language,
		ExperienceLevel: input.Experience,
	}, nodeID, input.Symbol, snippet)
	if err != nil {
		d := MockNodeDetail(nodeID, title, input.FilePath, line)
		return d, nil
	}
	d, err := ParseNodeDetail(raw)
	if err != nil {
		d = MockNodeDetail(nodeID, title, input.FilePath, line)
		return d, nil
	}
	if d.Confidence == "" {
		d.Confidence = confidence
	}
	if d.File == "" {
		d.File = input.FilePath
	}
	if d.Line == 0 {
		d.Line = line
	}
	return d, nil
}

func (b *Builder) gatherContext(input BuildInput) (snippet, lang string, callees []scanner.CalleeRef, branches []scanner.BranchRef, matches []scanner.Match) {
	if input.FilePath != "" {
		content, language, err := b.scanner.ReadFile(input.WorkspacePath, input.FilePath)
		if err == nil {
			snippet = truncateStr(content, 4000)
			lang = language
			callees = scanner.FindCalleesInSnippet(snippet, lang, 10)
			branches = scanner.FindBranchesInSnippet(snippet, 8)
		}
	}
	matches, _ = b.scanner.GrepSymbol(input.WorkspacePath, input.FilePath, input.Symbol)
	return snippet, lang, callees, branches, matches
}

func cycleDetected(nodeID string, parentPath []string) bool {
	for _, p := range parentPath {
		if p == nodeID {
			return true
		}
	}
	return false
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n..."
}

// FindNode looks up a node by ID in a graph fragment list.
func FindNode(nodes []FlowNode, id string) (FlowNode, bool) {
	for _, n := range nodes {
		if n.ID == id {
			return n, true
		}
	}
	return FlowNode{}, false
}

// SanitizeLabel trims labels for display.
func SanitizeLabel(label string) string {
	label = strings.TrimSpace(label)
	if len(label) > 24 {
		return label[:21] + "..."
	}
	return label
}
