package search

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFinderFindsGoFunctionsAndMethods(t *testing.T) {
	requireRipgrep(t)
	root := t.TempDir()
	source := `package sample

func Parse(value string) error {
	return nil
}

type parser struct{}

func (parser) Parse(value string) error {
	return nil
}

func caller() {
	_ = Parse("example")
}
`
	writeTestFile(t, root, "sample.go", source)

	matches, err := NewFinder("rg").Find(context.Background(), Query{
		Name:     "Parse",
		Root:     root,
		Language: "go",
		Limit:    20,
	})
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("Find() returned %d matches, want 2: %#v", len(matches), matches)
	}
	if matches[0].Path != "sample.go" || matches[0].Line != 3 {
		t.Errorf("first match = %#v, want sample.go:3", matches[0])
	}
	if matches[1].Path != "sample.go" || matches[1].Line != 9 {
		t.Errorf("second match = %#v, want sample.go:9", matches[1])
	}
}

func TestFinderHonorsLanguageFilter(t *testing.T) {
	requireRipgrep(t)
	root := t.TempDir()
	writeTestFile(t, root, "sample.py", "def Parse(value):\n    return value\n")

	matches, err := NewFinder("rg").Find(context.Background(), Query{
		Name:     "Parse",
		Root:     root,
		Language: "go",
	})
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("Find() returned %#v, want no matches", matches)
	}
}

func TestFinderRejectsInvalidFunctionName(t *testing.T) {
	_, err := NewFinder("rg").Find(context.Background(), Query{
		Name: "Parse.*",
		Root: t.TempDir(),
	})
	if err == nil {
		t.Fatal("Find() error = nil, want invalid identifier error")
	}
}

func requireRipgrep(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep is not installed")
	}
}

func writeTestFile(t *testing.T, root, name, contents string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
