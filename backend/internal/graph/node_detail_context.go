package graph

import (
	"fmt"
	"strings"
)

// NodeDetailContext is structured CFG context for LLM node explanations.
type NodeDetailContext struct {
	NodeID         string
	Kind           string
	Line           int
	Code           string
	EnrichSummary  string
	Incoming       []string
	Outgoing       []string
	CalleeSymbol   string
	CalleeFile     string
}

func buildNodeDetailContext(g FlowGraph, nodeID string) (NodeDetailContext, FlowNode, bool) {
	node, ok := FindNodeByID(g, nodeID)
	if !ok {
		return NodeDetailContext{}, FlowNode{}, false
	}

	ctx := NodeDetailContext{
		NodeID:        node.ID,
		Kind:          node.Kind,
		Line:          node.Line,
		Code:          node.Code,
		CalleeSymbol:  node.CalleeSymbol,
		CalleeFile:    node.CalleeFile,
	}

	for _, e := range g.Edges {
		if e.To != nodeID {
			continue
		}
		parent, found := FindNodeByID(g, e.From)
		if !found {
			continue
		}
		label := e.Label
		if label == "" {
			label = "flow"
		}
		ctx.Incoming = append(ctx.Incoming, fmt.Sprintf("from %s (%s) via %q", parent.ID, parent.Kind, label))
	}

	for _, e := range g.Edges {
		if e.From != nodeID {
			continue
		}
		child, found := FindNodeByID(g, e.To)
		if !found {
			continue
		}
		label := e.Label
		if label == "" {
			label = "flow"
		}
		ctx.Outgoing = append(ctx.Outgoing, fmt.Sprintf("to %s (%s) via %q", child.ID, child.Kind, label))
	}

	return ctx, node, true
}

func (c NodeDetailContext) Format(symbol string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Function: %s\n", symbol)
	fmt.Fprintf(&b, "Node id: %s\n", c.NodeID)
	fmt.Fprintf(&b, "Kind: %s\n", c.Kind)
	if c.Line > 0 {
		fmt.Fprintf(&b, "Line: %d\n", c.Line)
	}
	if c.EnrichSummary != "" {
		fmt.Fprintf(&b, "Existing summary (do not contradict): %s\n", c.EnrichSummary)
	}
	if c.Code != "" {
		fmt.Fprintf(&b, "Step code:\n```\n%s\n```\n", c.Code)
	}
	if c.CalleeSymbol != "" {
		fmt.Fprintf(&b, "Callee: %s", c.CalleeSymbol)
		if c.CalleeFile != "" {
			fmt.Fprintf(&b, " (%s)", c.CalleeFile)
		}
		b.WriteString("\n")
	}
	if len(c.Incoming) > 0 {
		b.WriteString("Incoming:\n")
		for _, line := range c.Incoming {
			fmt.Fprintf(&b, "- %s\n", line)
		}
	}
	if len(c.Outgoing) > 0 {
		b.WriteString("Outgoing:\n")
		for _, line := range c.Outgoing {
			fmt.Fprintf(&b, "- %s\n", line)
		}
	}
	b.WriteString("\nExplain ONLY this step in 2-4 plain sentences for a developer onboarding to this function. Do not invent other branches or steps.")
	return b.String()
}
