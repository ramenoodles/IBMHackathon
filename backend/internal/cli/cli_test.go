package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPrintsMatches(t *testing.T) {
	requireRipgrep(t)
	root := cliTestRepository(t)
	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), []string{"-root", root, "Parse"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "sample.go:3: func Parse") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunPrintsContextAndFileModes(t *testing.T) {
	requireRipgrep(t)
	root := cliTestRepository(t)
	for _, mode := range []string{"context", "file"} {
		var stdout, stderr bytes.Buffer
		code := Run(context.Background(), []string{"-root", root, "-source", mode, "Parse"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("Run(%s) code = %d, stderr = %q", mode, code, stderr.String())
		}
		if !strings.Contains(stdout.String(), "sample.go") || !strings.Contains(stdout.String(), "func Parse") {
			t.Errorf("Run(%s) stdout = %q", mode, stdout.String())
		}
	}
}

func TestRunRejectsInvalidArguments(t *testing.T) {
	tests := [][]string{
		{"-source", "invalid", "Parse"},
		{"-before", "-1", "Parse"},
		{"-after", "-1", "Parse"},
		{"-max", "0", "Parse"},
		{},
	}
	for _, args := range tests {
		var stdout, stderr bytes.Buffer
		if code := Run(context.Background(), args, &stdout, &stderr); code != 2 {
			t.Errorf("Run(%v) code = %d, want 2", args, code)
		}
	}
}

func TestRunReturnsOneWhenNoMatch(t *testing.T) {
	requireRipgrep(t)
	root := cliTestRepository(t)
	var stdout, stderr bytes.Buffer
	if code := Run(context.Background(), []string{"-root", root, "Missing"}, &stdout, &stderr); code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "No likely implementation") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func requireRipgrep(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep is not installed")
	}
}

func cliTestRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "sample.go"), []byte("package sample\n\nfunc Parse() string {\n\treturn \"ok\"\n}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}
