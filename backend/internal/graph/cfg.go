package graph

import (
	"fmt"

	"github.com/ibmhackathon/onbober/internal/scanner/flow"
)

// BuildCFGGraph constructs a control-flow graph from hybrid scan steps.
func BuildCFGGraph(symbol, filePath string, steps []flow.Step) FlowGraph {
	if len(steps) == 0 {
		return FlowGraph{Symbol: symbol, Depth: 1}
	}

	nodes := make([]FlowNode, 0, len(steps))
	idByLine := map[int]string{}
	rootID := ""

	for i, step := range steps {
		id := fmt.Sprintf("step_%d", step.Line)
		idByLine[step.Line] = id
		if i == 0 {
			rootID = id
		}

		node := FlowNode{
			ID:         id,
			Label:      flow.FormatLabel(step),
			Summary:    step.Summary,
			Kind:       step.Kind,
			Confidence: ConfidenceVerified,
			File:       filePath,
			Line:       step.Line,
			Code:       step.Code,
		}
		if step.Kind == "call" && step.CalleeSymbol != "" {
			node.CalleeSymbol = step.CalleeSymbol
			node.Expandable = true
		}
		if step.Kind == "loop" {
			node.Expandable = true
			node.Collapsed = true
			node.ChildCount = countLoopBody(steps, i)
		}
		nodes = append(nodes, node)
	}

	edges := buildCFGEdges(steps, idByLine)

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

func buildCFGEdges(steps []flow.Step, idByLine map[int]string) []FlowEdge {
	var edges []FlowEdge
	add := func(from, to int, label string) {
		if from < 0 || to < 0 || from >= len(steps) || to >= len(steps) {
			return
		}
		edges = append(edges, FlowEdge{
			From:  idByLine[steps[from].Line],
			To:    idByLine[steps[to].Line],
			Label: label,
		})
	}

	for i := range steps {
		cur := steps[i]
		switch cur.Kind {
		case "return", "raise":
			continue
		case "branch":
			if cur.BranchKind == "else" {
				if body := firstBodyIndex(steps, i); body >= 0 {
					add(i, body, "body")
				}
				continue
			}
			if cur.BranchKind == "if" || cur.BranchKind == "elif" {
				if body := firstBodyIndex(steps, i); body >= 0 {
					add(i, body, "true")
				}
				if falseIdx := falseBranchIndex(steps, i); falseIdx >= 0 {
					add(i, falseIdx, "false")
				}
				continue
			}
		case "loop":
			if body := firstBodyIndex(steps, i); body >= 0 {
				add(i, body, "each")
			}
			if after := afterBlockIndex(steps, i); after >= 0 {
				add(i, after, "done")
			}
			continue
		case "entry":
			if len(steps) > 1 {
				add(i, 1, "start")
			}
			continue
		}

		// Sequential steps within a block.
		if next := sequentialNext(steps, i); next >= 0 {
			add(i, next, "then")
		}
	}
	return dedupeEdges(edges)
}

func firstBodyIndex(steps []flow.Step, i int) int {
	if i+1 >= len(steps) {
		return -1
	}
	if steps[i+1].Indent > steps[i].Indent {
		return i + 1
	}
	return -1
}

func bodyEndIndex(steps []flow.Step, i int) int {
	end := i
	for j := i + 1; j < len(steps); j++ {
		if steps[j].Indent <= steps[i].Indent {
			break
		}
		end = j
	}
	return end
}

func falseBranchIndex(steps []flow.Step, i int) int {
	end := bodyEndIndex(steps, i)
	for j := end + 1; j < len(steps); j++ {
		if steps[j].Indent < steps[i].Indent {
			break
		}
		if steps[j].Indent == steps[i].Indent {
			if steps[j].BranchKind == "elif" || steps[j].BranchKind == "else" {
				return j
			}
			return j
		}
	}
	return -1
}

func afterBlockIndex(steps []flow.Step, i int) int {
	end := bodyEndIndex(steps, i)
	for j := end + 1; j < len(steps); j++ {
		if steps[j].Indent <= steps[i].Indent {
			return j
		}
	}
	return -1
}

func sequentialNext(steps []flow.Step, i int) int {
	if i+1 >= len(steps) {
		return -1
	}
	next := steps[i+1]
	// Do not fall through into a sibling block at lower indent.
	if next.Indent < steps[i].Indent {
		return -1
	}
	// Next line at same indent — continue in block.
	if next.Indent == steps[i].Indent {
		return i + 1
	}
	// Nested statement immediately following (e.g. if body).
	if next.Indent > steps[i].Indent {
		return i + 1
	}
	return -1
}

func countLoopBody(steps []flow.Step, i int) int {
	n := 0
	for j := i + 1; j < len(steps); j++ {
		if steps[j].Indent <= steps[i].Indent {
			break
		}
		n++
	}
	if n < 1 {
		n = 1
	}
	return n
}

func dedupeEdges(edges []FlowEdge) []FlowEdge {
	seen := map[string]bool{}
	var out []FlowEdge
	for _, e := range edges {
		key := e.From + "|" + e.Label + "|" + e.To
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, e)
	}
	return out
}
