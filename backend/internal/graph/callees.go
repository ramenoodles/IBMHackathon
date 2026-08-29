package graph

import (
	"github.com/ibmhackathon/onbober/internal/scanner"
	"github.com/ibmhackathon/onbober/internal/scanner/flow"
)

// AttachCalleeHints resolves call targets and sets expand metadata on nodes.
func AttachCalleeHints(g FlowGraph, workspace string, sc *scanner.Scanner, lang string) FlowGraph {
	for i := range g.Nodes {
		n := &g.Nodes[i]
		if n.Kind != "call" || n.CalleeSymbol == "" {
			continue
		}
		ref, ok := sc.ResolveCallee(workspace, n.File, n.CalleeSymbol, lang)
		if !ok {
			continue
		}
		n.CalleeFile = ref.DefFile
		n.CalleeLine = ref.DefLine
		n.Expandable = true

		if ref.DefFile != "" {
			content, fileLang, err := sc.ReadFile(workspace, ref.DefFile)
			if err == nil {
				steps := flow.ExtractFlow(content, ref.DefFile, n.CalleeSymbol, fileLang)
				childCount := len(steps)
				if childCount > 1 {
					childCount-- // exclude entry
				}
				n.ChildCount = childCount
				n.Collapsed = childCount > 3
			}
		}
	}
	return g
}

// FindNodeByID returns a node from the graph.
func FindNodeByID(g FlowGraph, id string) (FlowNode, bool) {
	return FindNode(g.Nodes, id)
}
