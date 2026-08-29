package graph

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseFlowGraph parses and validates LLM JSON output into a FlowGraph.
func ParseFlowGraph(raw string, symbol, filePath string) (FlowGraph, error) {
	cleaned := extractJSON(raw)
	var g FlowGraph
	if err := json.Unmarshal([]byte(cleaned), &g); err != nil {
		return FlowGraph{}, err
	}
	if g.Symbol == "" {
		g.Symbol = symbol
	}
	enforceLimits(&g, MaxRootNodes)
	if g.RootID == "" && len(g.Nodes) > 0 {
		g.RootID = g.Nodes[0].ID
	}
	for i := range g.Nodes {
		if g.Nodes[i].File == "" {
			g.Nodes[i].File = filePath
		}
	}
	return g, nil
}

// ParseNodeDetail parses LLM JSON into NodeDetail.
func ParseNodeDetail(raw string) (NodeDetail, error) {
	cleaned := extractJSON(raw)
	var d NodeDetail
	if err := json.Unmarshal([]byte(cleaned), &d); err != nil {
		return NodeDetail{}, err
	}
	return d, nil
}

func extractJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if idx := strings.Index(raw, "{"); idx >= 0 {
		raw = raw[idx:]
	}
	if idx := strings.LastIndex(raw, "}"); idx >= 0 {
		raw = raw[:idx+1]
	}
	return raw
}

func enforceLimits(g *FlowGraph, maxNodes int) {
	if len(g.Nodes) <= maxNodes {
		return
	}
	hidden := len(g.Nodes) - maxNodes + 1
	last := g.Nodes[maxNodes-1]
	last.Collapsed = true
	last.ChildCount = hidden
	last.Expandable = true
	last.Kind = "branch"
	last.Label = fmt.Sprintf("+%d branches", hidden)
	g.Nodes = g.Nodes[:maxNodes]
	allowed := map[string]bool{}
	for _, n := range g.Nodes {
		allowed[n.ID] = true
	}
	var edges []FlowEdge
	for _, e := range g.Edges {
		if allowed[e.From] && allowed[e.To] {
			edges = append(edges, e)
		}
	}
	g.Edges = edges
}
