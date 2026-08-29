package scanner

import "testing"

func TestLanguageForSearch(t *testing.T) {
	if got := languageForSearch(""); got != "auto" {
		t.Fatalf("empty path: got %q", got)
	}
	if got := languageForSearch("pkg/main.go"); got != "go" {
		t.Fatalf("go file: got %q", got)
	}
	if got := languageForSearch("init/main.c"); got != "c" {
		t.Fatalf("c file: got %q", got)
	}
}

func TestGrepSymbolViaWrapperEmptySymbol(t *testing.T) {
	s := New()
	matches, err := s.grepSymbolViaWrapper("/tmp", "", "", "auto")
	if err != nil {
		t.Fatal(err)
	}
	if matches != nil {
		t.Fatalf("expected nil matches, got %v", matches)
	}
}
