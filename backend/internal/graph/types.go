// Package graph provides execution-flow graph models and builders for OnBober.
package graph

// Confidence indicates whether a node is backed by scanner evidence or LLM inference.
type Confidence string

const (
	ConfidenceVerified Confidence = "verified"
	ConfidenceInferred Confidence = "inferred"
)

// Limits enforced server-side to prevent graph explosion.
const (
	MaxRootNodes   = 8
	MaxExpandNodes = 6
	MaxDepth       = 4
)

// FlowNode represents a single step in an execution/call-flow graph.
type FlowNode struct {
	ID         string     `json:"id"`
	Label      string     `json:"label"`
	Summary    string     `json:"summary"`
	Kind       string     `json:"kind"`
	Confidence Confidence `json:"confidence"`
	File       string     `json:"file,omitempty"`
	Line       int        `json:"line,omitempty"`
	Code       string     `json:"code,omitempty"`
	Expandable bool       `json:"expandable"`
	ChildCount int        `json:"childCount"`
	Collapsed  bool       `json:"collapsed"`
}

// FlowEdge connects two nodes in the flow graph.
type FlowEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label,omitempty"`
}

// FlowGraph is a bounded subgraph returned by root or expand endpoints.
type FlowGraph struct {
	RootID string     `json:"rootId"`
	Nodes  []FlowNode `json:"nodes"`
	Edges  []FlowEdge `json:"edges"`
	Depth  int        `json:"depth"`
	Symbol string     `json:"symbol"`
	Mock   bool       `json:"mock,omitempty"`
}

// NodeDetail holds expanded explanation for a selected node.
type NodeDetail struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Summary        string     `json:"summary"`
	Explanation    string     `json:"explanation"`
	Confidence     Confidence `json:"confidence"`
	File           string     `json:"file,omitempty"`
	Line           int        `json:"line,omitempty"`
	RelatedSymbols []string   `json:"relatedSymbols,omitempty"`
	Mock           bool       `json:"mock,omitempty"`
}

// BuildInput carries context for graph construction.
type BuildInput struct {
	WorkspacePath string
	FilePath      string
	Symbol        string
	NodeID        string
	ParentPath    []string
	Depth         int
	ExpandLimit   int
	Experience    string
	Language      string
}
