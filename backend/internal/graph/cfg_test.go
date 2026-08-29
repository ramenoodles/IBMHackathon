package graph

import (
	"testing"

	"github.com/ibmhackathon/onbober/internal/scanner/flow"
)

func TestCFGGuardReturnBranch(t *testing.T) {
	// Models: if bad: return; for ...
	steps := []flow.Step{
		{Line: 68, Kind: "entry", Label: "compact_entry()", Indent: 4},
		{Line: 69, Kind: "branch", Label: "if not isinstance", BranchKind: "if", Indent: 4},
		{Line: 70, Kind: "return", Label: "return", Indent: 8},
		{Line: 72, Kind: "loop", Label: "for key, value", LoopKind: "for", Indent: 4},
		{Line: 73, Kind: "branch", Label: "if isinstance list", BranchKind: "if", Indent: 8},
	}
	g := BuildCFGGraph("compact_entry", "f.py", steps)

	hasFalseToLoop := false
	hasReturnToLoop := false
	for _, e := range g.Edges {
		if e.From == "step_69" && e.To == "step_72" && e.Label == "false" {
			hasFalseToLoop = true
		}
		if e.From == "step_70" && e.To == "step_72" {
			hasReturnToLoop = true
		}
	}
	if !hasFalseToLoop {
		t.Fatal("expected if -> false -> loop edge")
	}
	if hasReturnToLoop {
		t.Fatal("return should not connect to loop")
	}
}

func TestCFGBranchEdges(t *testing.T) {
	steps := []flow.Step{
		{Line: 1, Kind: "entry", Label: "msg()", Indent: 0},
		{Line: 2, Kind: "branch", Label: "if (status === 429)", BranchKind: "if", Indent: 0},
		{Line: 3, Kind: "call", Label: "parseRetryAfter()", Indent: 4},
		{Line: 4, Kind: "branch", Label: "else", BranchKind: "else", Indent: 0},
	}
	g := BuildCFGGraph("messageFromApiFailure", "api.js", steps)
	hasFalse := false
	for _, e := range g.Edges {
		if e.Label == "false" {
			hasFalse = true
		}
	}
	if !hasFalse {
		t.Fatal("expected false branch edge from if")
	}
}
