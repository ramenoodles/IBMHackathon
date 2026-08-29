package flow

import "testing"

func TestHumanizeBranch(t *testing.T) {
	steps := EnrichSteps([]Step{{
		Kind: "branch", BranchKind: "if",
		Code: "if not isinstance(entry, dict):",
	}})
	if steps[0].Summary == "" || steps[0].Label == "" {
		t.Fatalf("expected humanized branch, got %+v", steps[0])
	}
}
