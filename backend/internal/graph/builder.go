package graph

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ibmhackathon/onbober/internal/llm"
	"github.com/ibmhackathon/onbober/internal/scanner"
)

// Builder orchestrates scan data, LLM generation, and graph limits.
type Builder struct {
	scanner *scanner.Scanner
	llm     *llm.OllamaClient
	cache   *Cache
}

// NewBuilder creates a graph builder with scanner and LLM clients.
func NewBuilder(s *scanner.Scanner, client *llm.OllamaClient) *Builder {
	return &Builder{
		scanner: s,
		llm:     client,
		cache:   NewCache(10 * time.Minute),
	}
}

// BuildRoot constructs the initial flow graph for a symbol.
func (b *Builder) BuildRoot(ctx context.Context, input BuildInput) (FlowGraph, error) {
	key := CacheKey(input.WorkspacePath, input.Symbol, input.FilePath, "", nil, 0)
	if cached, ok := b.cache.Get(key); ok {
		return cached, nil
	}

	snippet, lang, callees, branches, matches := b.gatherContext(input)
	steps := b.extractSteps(input)
	g, err := b.generateGraph(ctx, input, snippet, lang, callees, branches, matches, "", MaxRootNodes)
	if err != nil {
		if len(steps) >= 2 {
			g = scanFirstGraph(input.Symbol, input.FilePath, steps)
		} else {
			g = MockRootGraph(input.Symbol, input.FilePath)
			g.Mock = true
		}
	} else {
		g = NormalizeGraph(g, steps, input.Symbol, input.FilePath)
	}
	b.cache.Set(key, g)
	return g, nil
}

// BuildExpand constructs child nodes for a collapsed branch node.
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

	snippet, lang, callees, branches, matches := b.gatherContext(input)
	g, err := b.generateGraph(ctx, input, snippet, lang, callees, branches, matches, input.NodeID, limit)
	if err != nil {
		g = MockExpandGraph(input.NodeID, input.Symbol, input.FilePath, limit)
		g.Mock = true
	} else {
		steps := b.extractSteps(input)
		g = NormalizeGraph(g, steps, input.Symbol, input.FilePath)
	}
	b.cache.Set(key, g)
	return g, nil
}

// BuildNodeDetail returns an expanded explanation for a node.
func (b *Builder) BuildNodeDetail(ctx context.Context, input BuildInput, nodeID, title string, line int, confidence Confidence) (NodeDetail, error) {
	snippet, _, _, _, _ := b.gatherContext(input)
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

func (b *Builder) extractSteps(input BuildInput) []scanner.FlowStep {
	if input.FilePath == "" {
		return nil
	}
	content, lang, err := b.scanner.ReadFile(input.WorkspacePath, input.FilePath)
	if err != nil {
		return nil
	}
	return scanner.ExtractSymbolSteps(content, input.FilePath, input.Symbol, lang)
}

func (b *Builder) generateGraph(ctx context.Context, input BuildInput, snippet, lang string, callees []scanner.CalleeRef, branches []scanner.BranchRef, matches []scanner.Match, parentNode string, limit int) (FlowGraph, error) {
	matchRefs := make([]llm.MatchRef, len(matches))
	for i, m := range matches {
		matchRefs[i] = llm.MatchRef{File: m.File, Line: m.Line, Content: m.Content}
	}

	raw, err := b.llm.GenerateFlowGraph(ctx, llm.GraphBuildContext{
		Symbol:      input.Symbol,
		FilePath:    input.FilePath,
		Snippet:     snippet,
		Matches:     matchRefs,
		Callees:     callees,
		Branches:    branches,
		UserContext: llm.UserContext{PrimaryLanguage: input.Language, ExperienceLevel: input.Experience},
		ParentNode:  parentNode,
		ExpandLimit: limit,
	})
	if err != nil {
		return FlowGraph{}, err
	}

	g, err := ParseFlowGraph(raw, input.Symbol, input.FilePath)
	if err != nil {
		return FlowGraph{}, err
	}
	if parentNode != "" {
		g.Depth = input.Depth
	}
	return g, nil
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
