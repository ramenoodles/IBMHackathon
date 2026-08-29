package graph

import (
	"fmt"
	"strings"

	"github.com/ibmhackathon/onbober/internal/scanner"
)

// NormalizeGraph prefers a scan-derived connected path when the LLM graph is weak.
func NormalizeGraph(g FlowGraph, steps []scanner.FlowStep, symbol, filePath string) FlowGraph {
	if len(steps) >= 2 {
		if !isConnectedFromRoot(g) || hasAmbiguousLabels(g) {
			return scanFirstGraph(symbol, filePath, steps)
		}
		g = enrichFromSteps(g, steps)
	} else if !isConnectedFromRoot(g) {
		g = connectOrphans(g)
	}
	g = ensureEntryNode(g, symbol, filePath, steps)
	return g
}

func scanFirstGraph(symbol, filePath string, steps []scanner.FlowStep) FlowGraph {
	if len(steps) == 0 {
		return FlowGraph{Symbol: symbol, RootID: "", Nodes: nil, Edges: nil, Depth: 1}
	}

	nodes := make([]FlowNode, 0, len(steps))
	edges := make([]FlowEdge, 0, len(steps)-1)
	rootID := ""

	for i, step := range steps {
		id := fmt.Sprintf("step_%d", step.Line)
		if i == 0 {
			rootID = id
		}
		nodes = append(nodes, FlowNode{
			ID:         id,
			Label:      scanner.FormatStepLabel(step),
			Summary:    step.Summary,
			Kind:       step.Kind,
			Confidence: Confidence(step.Confidence),
			File:       filePath,
			Line:       step.Line,
			Code:       step.Code,
		})
		if i > 0 {
			edgeLabel := ""
			if step.Kind == "branch" {
				edgeLabel = "branch"
			} else if step.Kind == "return" {
				edgeLabel = "return"
			} else {
				edgeLabel = "then"
			}
			edges = append(edges, FlowEdge{
				From:  nodes[i-1].ID,
				To:    id,
				Label: edgeLabel,
			})
		}
	}

	g := FlowGraph{
		RootID: rootID,
		Symbol: symbol,
		Depth:  1,
		Nodes:  nodes,
		Edges:  edges,
	}
	enforceLimits(&g, MaxRootNodes)
	return g
}

func enrichFromSteps(g FlowGraph, steps []scanner.FlowStep) FlowGraph {
	stepByLine := map[int]scanner.FlowStep{}
	for _, s := range steps {
		stepByLine[s.Line] = s
	}

	for i := range g.Nodes {
		n := &g.Nodes[i]
		if n.Line > 0 {
			if step, ok := stepByLine[n.Line]; ok {
				if n.Code == "" {
					n.Code = step.Code
				}
				if n.Summary == "" || len(n.Summary) < 8 {
					n.Summary = step.Summary
				}
				if !strings.HasPrefix(n.Label, "L") {
					n.Label = scanner.FormatStepLabel(step)
				}
				n.Confidence = ConfidenceVerified
				continue
			}
		}
		// Match by loose label when line missing.
		for _, step := range steps {
			if strings.Contains(step.Code, strings.TrimSuffix(n.Label, "()")) ||
				strings.EqualFold(step.Label, n.Label) {
				n.Line = step.Line
				n.Code = step.Code
				n.Label = scanner.FormatStepLabel(step)
				n.Summary = step.Summary
				n.Confidence = ConfidenceVerified
				break
			}
		}
	}
	return g
}

func ensureEntryNode(g FlowGraph, symbol, filePath string, steps []scanner.FlowStep) FlowGraph {
	if g.RootID != "" {
		return g
	}
	if len(g.Nodes) == 0 {
		return g
	}
	entryLine := 1
	entryCode := symbol + "()"
	if len(steps) > 0 {
		entryLine = steps[0].Line
		entryCode = steps[0].Code
	}
	entryID := fmt.Sprintf("entry_%s", symbol)
	entry := FlowNode{
		ID: entryID, Label: symbol + "()", Summary: "Entry point",
		Kind: "entry", Confidence: ConfidenceVerified,
		File: filePath, Line: entryLine, Code: entryCode,
	}
	g.Nodes = append([]FlowNode{entry}, g.Nodes...)
	if len(g.Nodes) > 1 {
		g.Edges = append([]FlowEdge{{From: entryID, To: g.Nodes[1].ID, Label: "start"}}, g.Edges...)
	}
	g.RootID = entryID
	return g
}

func isConnectedFromRoot(g FlowGraph) bool {
	if g.RootID == "" || len(g.Nodes) <= 1 {
		return len(g.Nodes) <= 1
	}
	adj := map[string][]string{}
	for _, e := range g.Edges {
		adj[e.From] = append(adj[e.From], e.To)
	}
	seen := map[string]bool{g.RootID: true}
	queue := []string{g.RootID}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, next := range adj[cur] {
			if seen[next] {
				continue
			}
			seen[next] = true
			queue = append(queue, next)
		}
	}
	return len(seen) == len(g.Nodes)
}

func hasAmbiguousLabels(g FlowGraph) bool {
	counts := map[string]int{}
	for _, n := range g.Nodes {
		key := strings.ToLower(strings.TrimSpace(n.Label))
		if key == "" {
			continue
		}
		counts[key]++
	}
	dupes := 0
	for _, c := range counts {
		if c > 1 {
			dupes++
		}
	}
	// Repeated generic labels like create/session without line context.
	return dupes >= 2 || (dupes >= 1 && !isConnectedFromRoot(g))
}

func connectOrphans(g FlowGraph) FlowGraph {
	if g.RootID == "" || len(g.Nodes) < 2 {
		return g
	}
	inTree := reachable(g.RootID, g.Edges)
	var last string = g.RootID
	for _, n := range g.Nodes {
		if inTree[n.ID] {
			last = n.ID
			continue
		}
		g.Edges = append(g.Edges, FlowEdge{From: last, To: n.ID, Label: "then"})
		inTree[n.ID] = true
		last = n.ID
	}
	return g
}

func reachable(root string, edges []FlowEdge) map[string]bool {
	adj := map[string][]string{}
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], e.To)
	}
	seen := map[string]bool{root: true}
	queue := []string{root}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, next := range adj[cur] {
			if seen[next] {
				continue
			}
			seen[next] = true
			queue = append(queue, next)
		}
	}
	return seen
}
