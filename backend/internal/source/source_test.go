package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewReaderDefaultsEmptyRoot(t *testing.T) {
	reader, err := NewReader("")
	if err != nil {
		t.Fatalf("NewReader() error = %v", err)
	}
	if reader.root == "" {
		t.Fatal("NewReader() returned an empty root")
	}
}

func TestNewReaderRejectsMissingAndFileRoots(t *testing.T) {
	if _, err := NewReader(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("NewReader() error = nil for missing root")
	}

	file := filepath.Join(t.TempDir(), "file.go")
	if err := os.WriteFile(file, []byte("package sample\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewReader(file); err == nil {
		t.Fatal("NewReader() error = nil for file root")
	}
}

func TestReaderReadFileAndContext(t *testing.T) {
	root := t.TempDir()
	content := "one\ntwo\nthree\nfour"
	if err := os.WriteFile(filepath.Join(root, "sample.go"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	reader, err := NewReader(root)
	if err != nil {
		t.Fatal(err)
	}

	got, err := reader.ReadFile("sample.go")
	if err != nil || got != content {
		t.Fatalf("ReadFile() = %q, %v; want %q", got, err, content)
	}
	snippet, err := reader.ReadContext("sample.go", 2, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if snippet.StartLine != 1 || snippet.EndLine != 3 || snippet.Content != "one\ntwo\nthree" {
		t.Fatalf("ReadContext() = %#v", snippet)
	}
}

func TestReaderRejectsOversizedFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "big.go"), make([]byte, 512), 0o600); err != nil {
		t.Fatal(err)
	}
	reader, err := NewReaderWithLimit(root, 256)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reader.ReadFile("big.go"); err == nil || !strings.Contains(err.Error(), "exceeds max size") {
		t.Fatalf("ReadFile(big.go) error = %v, want oversized error", err)
	}
}

func TestReaderRejectsUnsafeAndInvalidContext(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "sample.go"), []byte("one\ntwo"), 0o600); err != nil {
		t.Fatal(err)
	}
	reader, err := NewReader(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"../outside.go", filepath.Join(root, "sample.go")} {
		if _, err := reader.ReadFile(path); err == nil {
			t.Errorf("ReadFile(%q) error = nil", path)
		}
	}
	for _, test := range []struct {
		line, before, after int
	}{
		{0, 0, 0}, {1, -1, 0}, {1, 0, -1}, {3, 0, 0},
	} {
		if _, err := reader.ReadContext("sample.go", test.line, test.before, test.after); err == nil {
			t.Errorf("ReadContext(%+v) error = nil", test)
		}
	}
	if _, err := reader.ReadFile("missing.go"); err == nil || !strings.Contains(err.Error(), "read source file") {
		t.Fatalf("ReadFile(missing.go) error = %v", err)
	}
}
