package search

import (
	"strings"
	"testing"
)

func TestResolvePatternsSupportsAliasesAndEscapesNames(t *testing.T) {
	patterns, globs, err := resolvePatterns("JS", "parse$Value")
	if err != nil {
		t.Fatal(err)
	}
	if len(patterns) == 0 || len(globs) == 0 || !contains(globs, "*.js") {
		t.Fatalf("resolvePatterns() = %v, %v", patterns, globs)
	}
	if !strings.Contains(patterns[0], `parse\$Value`) {
		t.Errorf("pattern %q does not escape function name", patterns[0])
	}
}

func TestResolvePatternsRejectsInvalidInputs(t *testing.T) {
	for _, name := range []string{"", "parse.value", "1parse"} {
		if _, _, err := resolvePatterns("go", name); err == nil {
			t.Errorf("resolvePatterns(%q) error = nil", name)
		}
	}
	if _, _, err := resolvePatterns("kotlin", "Parse"); err == nil {
		t.Fatal("resolvePatterns() accepted unsupported language")
	}
}

func TestResolvePatternsAutoDeduplicatesGlobsAndPatterns(t *testing.T) {
	patterns, globs, err := resolvePatterns("auto", "Parse")
	if err != nil {
		t.Fatal(err)
	}
	if len(patterns) == 0 || len(globs) == 0 {
		t.Fatal("resolvePatterns() returned no search patterns")
	}
	for i := range patterns {
		for j := i + 1; j < len(patterns); j++ {
			if patterns[i] == patterns[j] {
				t.Fatalf("duplicate pattern %q", patterns[i])
			}
		}
	}
}
