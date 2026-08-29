package analysis

import (
	"context"
	"fmt"
	"grepwrapper/internal/search"
	"grepwrapper/internal/source"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
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
	r, e := source.NewReader(root)
	if e != nil {
		return nil, e
	}
	return &Builder{search.NewFinder(rg), r}, nil
}

var callRE = regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`)

func (b *Builder) Root(ctx context.Context, file, symbol string) (Graph, error) {
	matches, err := b.finder.Find(ctx, search.Query{Name: symbol, Root: b.reader.Root(), Limit: 1})
	if err != nil {
		return Graph{}, err
	}
	if len(matches) == 0 {
		return Graph{}, fmt.Errorf("symbol %q not found", symbol)
	}
	return b.build(ctx, matches[0].Path, matches[0].Line, symbol, 0, map[string]bool{})
}
func (b *Builder) build(ctx context.Context, file string, line int, symbol string, depth int, seen map[string]bool) (Graph, error) {
	id := nodeID(file, line, symbol)
	if seen[id] {
		return Graph{RootID: id, Symbol: symbol, Depth: depth}, nil
	}
	seen[id] = true
	snip, err := b.reader.ReadContext(file, line, 0, 30)
	if err != nil {
		return Graph{}, err
	}
	n := Node{ID: id, Label: symbol, Title: symbol, Summary: strings.TrimSpace(snip.Content), Kind: "function", File: file, Line: line, Confidence: "verified", Code: snip.Content}
	g := Graph{RootID: id, Nodes: []Node{n}, Symbol: symbol, Depth: depth}
	if depth >= 4 {
		return g, nil
	}
	names := map[string]bool{}
	for _, m := range callRE.FindAllStringSubmatch(snip.Content, -1) {
		if m[1] != symbol {
			names[m[1]] = true
		}
	}
	list := make([]string, 0, len(names))
	for name := range names {
		list = append(list, name)
	}
	sort.Strings(list)
	for _, name := range list {
		ms, e := b.finder.Find(ctx, search.Query{Name: name, Root: b.reader.Root(), Limit: 1})
		if e != nil {
			continue
		}
		if len(ms) == 0 {
			continue
		}
		child, e := b.build(ctx, ms[0].Path, ms[0].Line, name, depth+1, seen)
		if e != nil {
			continue
		}
		g.Nodes = append(g.Nodes, child.Nodes...)
		g.Edges = append(g.Edges, Edge{From: id, To: child.RootID, Label: "calls"})
	}
	g.Nodes[0].ChildCount = len(g.Edges)
	g.Nodes[0].Expandable = len(g.Edges) > 0
	g.Nodes[0].Collapsed = false
	return g, nil
}
func nodeID(file string, line int, symbol string) string {
	return filepath.ToSlash(file) + ":" + fmt.Sprint(line) + ":" + symbol
}
func (b *Builder) Expand(ctx context.Context, g Graph, nodeID string, limit int) (Graph, error) {
	for _, n := range g.Nodes {
		if n.ID == nodeID {
			return b.build(ctx, n.File, n.Line, n.Label, g.Depth+1, map[string]bool{nodeID: true})
		}
	}
	return Graph{}, fmt.Errorf("node %q not found", nodeID)
}
