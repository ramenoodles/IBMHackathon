package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractFileSymbolsPython(t *testing.T) {
	path := filepath.Join("flow", "testdata", "setup.py")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	symbols := ExtractFileSymbols(string(b), "setup.py")
	if len(symbols) == 0 {
		t.Fatal("expected symbols")
	}
	found := false
	for _, s := range symbols {
		if s.Name == "setUp" {
			found = true
			if s.Line < 1 {
				t.Errorf("expected line number")
			}
			if s.Kind != "function" {
				t.Errorf("kind = %q", s.Kind)
			}
			if s.Signature == "" {
				t.Error("expected signature")
			}
		}
	}
	if !found {
		t.Error("setUp not found")
	}
}

func TestExtractFileSymbolsDedupes(t *testing.T) {
	src := "def foo():\n    pass\ndef foo():\n    pass\n"
	symbols := ExtractFileSymbols(src, "x.py")
	if len(symbols) != 1 {
		t.Fatalf("got %d symbols", len(symbols))
	}
}
