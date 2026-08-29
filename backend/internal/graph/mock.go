package graph

import "fmt"

// MockRootGraph returns a demo graph for offline demos and parse failures.
func MockRootGraph(symbol, filePath string) FlowGraph {
	rootID := fmt.Sprintf("entry_%s", symbol)
	branchID := fmt.Sprintf("branch_%s", symbol)
	callID := fmt.Sprintf("call_%s", symbol)
	returnID := fmt.Sprintf("return_%s", symbol)

	return FlowGraph{
		RootID: rootID,
		Symbol: symbol,
		Depth:  1,
		Mock:   true,
		Nodes: []FlowNode{
			{
				ID: rootID, Label: symbol, Summary: "Entry point",
				Kind: "entry", Confidence: ConfidenceVerified,
				File: filePath, Line: 1, Expandable: false,
			},
			{
				ID: branchID, Label: "validate input", Summary: "Checks preconditions",
				Kind: "branch", Confidence: ConfidenceInferred,
				File: filePath, Line: 5, Expandable: true, ChildCount: 3, Collapsed: true,
			},
			{
				ID: callID, Label: "helper_fn", Summary: "Delegates to helper",
				Kind: "call", Confidence: ConfidenceInferred,
				File: filePath, Line: 12, Expandable: true, ChildCount: 2, Collapsed: false,
			},
			{
				ID: returnID, Label: "return", Summary: "Returns result to caller",
				Kind: "return", Confidence: ConfidenceVerified,
				File: filePath, Line: 20,
			},
		},
		Edges: []FlowEdge{
			{From: rootID, To: branchID, Label: "next"},
			{From: branchID, To: callID, Label: "on success"},
			{From: callID, To: returnID, Label: "then"},
		},
	}
}

// MockExpandGraph returns child nodes for a collapsed branch expansion.
func MockExpandGraph(parentID, symbol, filePath string, limit int) FlowGraph {
	nodes := []FlowNode{}
	edges := []FlowEdge{}
	count := limit
	if count > 3 {
		count = 3
	}
	for i := 0; i < count; i++ {
		id := fmt.Sprintf("%s_child_%d", parentID, i)
		nodes = append(nodes, FlowNode{
			ID: id, Label: fmt.Sprintf("branch_%d", i+1),
			Summary: fmt.Sprintf("Path %d for %s", i+1, symbol),
			Kind: "branch", Confidence: ConfidenceInferred,
			File: filePath, Line: 10 + i*3,
		})
		edges = append(edges, FlowEdge{From: parentID, To: id, Label: fmt.Sprintf("case %d", i+1)})
	}
	return FlowGraph{RootID: parentID, Symbol: symbol, Depth: 2, Mock: true, Nodes: nodes, Edges: edges}
}

// MockNodeDetail returns a demo detail card for a node.
func MockNodeDetail(nodeID, symbol, filePath string, line int) NodeDetail {
	return NodeDetail{
		ID: nodeID, Title: symbol,
		Summary:     "Demo node detail (Ollama offline)",
		Explanation: "This is a placeholder explanation. When Ollama is running, this card shows a tailored walkthrough of what happens at this step in the execution path.",
		Confidence:  ConfidenceInferred,
		File:        filePath,
		Line:        line,
		RelatedSymbols: []string{"helper_fn"},
		Mock:        true,
	}
}
