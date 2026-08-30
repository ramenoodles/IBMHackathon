package analysis

import "fmt"

func buildCFG(file, symbol string, depth int, steps []flowStep, limit int, truncationMarker bool) Graph {
	nodes := make([]Node, 0, len(steps))
	ids := make([]string, len(steps))
	for i, step := range steps {
		id := makeNodeID(file, symbol, depth, step.Line, i)
		ids[i] = id
		nodes = append(nodes, Node{
			ID: id, Label: step.Label, Title: step.Label, Summary: step.Summary,
			Kind: step.Kind, File: file, Line: step.Line, Code: step.Code,
			CalleeSymbol: step.CalleeSymbol, Confidence: "verified",
		})
	}
	edges := buildCFGEdges(steps, ids)
	graph := Graph{Nodes: nodes, Edges: edges, Depth: depth, Symbol: symbol}
	if len(nodes) > 0 {
		graph.RootID = nodes[0].ID
	}
	return limitGraph(graph, limit, truncationMarker)
}

func buildCFGEdges(steps []flowStep, ids []string) []Edge {
	edges := make([]Edge, 0, len(steps))
	add := func(from, to int, label string) {
		if from < 0 || to < 0 || from >= len(steps) || to >= len(steps) {
			return
		}
		edges = append(edges, Edge{From: ids[from], To: ids[to], Label: label})
	}
	for i, step := range steps {
		switch step.Kind {
		case "return", "raise":
			continue
		case "entry":
			if len(steps) > 1 {
				add(i, 1, "start")
			}
			continue
		case "branch":
			if body := firstBodyIndex(steps, i); body >= 0 {
				label := "true"
				if step.BranchKind == "else" {
					label = "body"
				}
				add(i, body, label)
			}
			if step.BranchKind != "else" {
				if next := falseBranchIndex(steps, i); next >= 0 {
					add(i, next, "false")
				}
			}
			continue
		case "loop":
			if body := firstBodyIndex(steps, i); body >= 0 {
				add(i, body, "each")
			}
			if after := afterBlockIndex(steps, i); after >= 0 {
				add(i, after, "done")
			}
			continue
		}
		if next := sequentialNext(steps, i); next >= 0 {
			add(i, next, "then")
		}
	}
	return dedupeEdges(edges)
}

func firstBodyIndex(steps []flowStep, index int) int {
	if index+1 < len(steps) && steps[index+1].Indent > steps[index].Indent {
		return index + 1
	}
	return -1
}

func bodyEndIndex(steps []flowStep, index int) int {
	end := index
	for i := index + 1; i < len(steps); i++ {
		if steps[i].Indent <= steps[index].Indent {
			break
		}
		end = i
	}
	return end
}

func falseBranchIndex(steps []flowStep, index int) int {
	for i := bodyEndIndex(steps, index) + 1; i < len(steps); i++ {
		if steps[i].Indent < steps[index].Indent {
			break
		}
		if steps[i].Indent == steps[index].Indent {
			return i
		}
	}
	return -1
}

func afterBlockIndex(steps []flowStep, index int) int {
	for i := bodyEndIndex(steps, index) + 1; i < len(steps); i++ {
		if steps[i].Indent <= steps[index].Indent {
			return i
		}
	}
	return -1
}

func sequentialNext(steps []flowStep, index int) int {
	if index+1 >= len(steps) || steps[index+1].Indent < steps[index].Indent {
		return -1
	}
	return index + 1
}

func dedupeEdges(edges []Edge) []Edge {
	seen := make(map[string]bool)
	out := make([]Edge, 0, len(edges))
	for _, edge := range edges {
		key := edge.From + "->" + edge.To
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, edge)
	}
	return out
}

func limitGraph(graph Graph, limit int, marker bool) Graph {
	if limit <= 0 || len(graph.Nodes) <= limit {
		return graph
	}
	if marker && limit > 1 {
		hidden := len(graph.Nodes) - limit + 1
		originalEdges := graph.Edges
		kept := append([]Node(nil), graph.Nodes[:limit-1]...)
		last := kept[len(kept)-1]
		markerNode := Node{
			ID:    makeNodeID(last.File, graph.Symbol, graph.Depth, last.Line, len(graph.Nodes)),
			Label: fmt.Sprintf("+%d more steps", hidden), Title: "Flow continues",
			Summary: fmt.Sprintf("%d additional flow steps were omitted", hidden),
			Kind:    "branch", File: last.File, Line: last.Line, Confidence: "inferred",
		}
		kept = append(kept, markerNode)
		graph.Nodes = kept
		graph.Edges = boundedEdges(originalEdges, kept)
		keptIDs := make(map[string]bool, len(kept))
		for _, node := range kept {
			keptIDs[node.ID] = true
		}
		linked := false
		for _, edge := range originalEdges {
			if keptIDs[edge.From] && !keptIDs[edge.To] {
				graph.Edges = append(graph.Edges, Edge{From: edge.From, To: markerNode.ID, Label: edge.Label})
				linked = true
			}
		}
		if !linked && last.Kind != "return" && last.Kind != "raise" {
			graph.Edges = append(graph.Edges, Edge{From: last.ID, To: markerNode.ID, Label: "then"})
		}
		graph.Edges = dedupeEdges(graph.Edges)
		return graph
	}
	graph.Nodes = append([]Node(nil), graph.Nodes[:limit]...)
	graph.Edges = boundedEdges(graph.Edges, graph.Nodes)
	return graph
}

func boundedEdges(edges []Edge, nodes []Node) []Edge {
	allowed := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		allowed[node.ID] = true
	}
	out := make([]Edge, 0, len(edges))
	for _, edge := range edges {
		if allowed[edge.From] && allowed[edge.To] {
			out = append(out, edge)
		}
	}
	return out
}
