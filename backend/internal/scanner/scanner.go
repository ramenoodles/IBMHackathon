// Package scanner provides filesystem and ripgrep utilities for codebase analysis.
//
// Symbol declaration search is delegated to the grepwrapper module (see
// grepwrapper_adapter.go and docs/GREPWRAPPER.md). Fixed-string literal
// search for deep-dive evidence lives in grep_literal.go.
package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Scanner wraps filesystem and ripgrep operations.
type Scanner struct{}

// New creates a Scanner instance.
func New() *Scanner {
	return &Scanner{}
}

// TreeEntry represents a single file or directory in a listing.
type TreeEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
}

// SafePath resolves and validates an absolute path, rejecting traversal attempts.
func SafePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("path traversal not allowed")
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("path does not exist: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path must be a directory")
	}
	return abs, nil
}

// SafeJoin resolves rel within workspace and ensures it stays inside the workspace root.
func SafeJoin(workspace, rel string) (string, error) {
	workspaceAbs, err := filepath.Abs(workspace)
	if err != nil {
		return "", err
	}
	clean := filepath.Clean(strings.ReplaceAll(rel, "/", string(filepath.Separator)))
	if clean == "." {
		clean = ""
	}
	if strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("path traversal not allowed")
	}
	full := filepath.Join(workspaceAbs, clean)
	relCheck, err := filepath.Rel(workspaceAbs, full)
	if err != nil || strings.HasPrefix(relCheck, "..") {
		return "", fmt.Errorf("path outside workspace")
	}
	return full, nil
}
