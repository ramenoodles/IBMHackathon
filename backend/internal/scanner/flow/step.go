package flow

// Step is one ordered statement inside a symbol body from deterministic parsing.
type Step struct {
	Line             int
	Kind             string // entry, call, branch, return, assign, loop, raise
	Label            string
	Summary          string
	Code             string
	Confidence       string
	CalleeSymbol     string
	CalleeQualified  string
	BranchKind       string // if, elif, else
	LoopKind         string // for, while
	Indent           int    // leading whitespace depth in the symbol body
	Source           string // regex, treesitter, merged
}
