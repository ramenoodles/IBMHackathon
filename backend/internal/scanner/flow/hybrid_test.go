package flow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractRegexPythonSetUp(t *testing.T) {
	src := readTestdata(t, "setup.py")
	steps := ExtractRegex(src, "setUp", "python")
	if len(steps) < 3 {
		t.Fatalf("expected >=3 steps, got %d", len(steps))
	}
	if steps[1].CalleeSymbol != "create_engine" {
		t.Fatalf("expected create_engine, got %s", steps[1].CalleeSymbol)
	}
}

func TestExtractFlowJSBranches(t *testing.T) {
	src := readTestdata(t, "api_error.js")
	steps := ExtractFlow(src, "api_error.js", "messageFromApiFailure", "javascript")
	branches := 0
	for _, s := range steps {
		if s.Kind == "branch" {
			branches++
		}
	}
	if branches < 3 {
		t.Fatalf("expected multiple branches, got %d", branches)
	}
}

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("testdata", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
