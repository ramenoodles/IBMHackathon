package graph

import "testing"

func TestEnrichKeyDistinctForSameLine(t *testing.T) {
	a := EnrichKey("ws", "file.py", "node_a", 42)
	b := EnrichKey("ws", "file.py", "node_b", 42)
	if a == b {
		t.Fatalf("expected distinct keys for different node ids on same line, got %q", a)
	}
}

func TestValidateNodeDetailRejectsCodeLikeTitle(t *testing.T) {
	if ValidateNodeDetail(NodeDetail{Title: "return x", Explanation: "ok"}) {
		t.Fatal("expected code-like title to fail validation")
	}
	if !ValidateNodeDetail(NodeDetail{Title: "Check input", Explanation: "Validates the value."}) {
		t.Fatal("expected valid detail")
	}
}
