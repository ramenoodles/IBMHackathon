package scanner

import "testing"

func TestFindEnclosingSymbolPython(t *testing.T) {
	src := `class AdzunaJobProvider:
    def provider_name(self) -> str:
        return "adzuna"
`
	name, kind := FindEnclosingSymbol(src, 3)
	if name != "AdzunaJobProvider" || kind != "class" {
		t.Fatalf("got %q %q, want AdzunaJobProvider class", name, kind)
	}
}

func TestExtractStringLiterals(t *testing.T) {
	lits := ExtractStringLiterals(`return "adzuna"`)
	if len(lits) != 1 || lits[0] != "adzuna" {
		t.Fatalf("unexpected literals: %v", lits)
	}
}

func TestDistinctiveDomainTerms(t *testing.T) {
	terms := DistinctiveDomainTerms("AdzunaJobProvider", "adzuna.py", "adzuna")
	found := false
	for _, t := range terms {
		if t == "adzuna" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected adzuna in %v", terms)
	}
}
