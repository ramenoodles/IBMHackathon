package source

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Snippet struct {
	Path      string
	StartLine int
	EndLine   int
	Content   string
}

type Reader struct {
	root string
}

func (reader *Reader) Root() string { return reader.root }

// Returns a new reader while doing basic validtion before hand
func NewReader(root string) (*Reader, error) {
	if strings.TrimSpace(root) == "" {
		root = "."
	}

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve source root %q: %w", root, err)
	}

	info, err := os.Stat(absoluteRoot)
	if err != nil {
		return nil, fmt.Errorf("open source root %q: %w", root, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("source root %q is not a directory", root)
	}

	return &Reader{
		root: absoluteRoot,
	}, nil
}

func (reader *Reader) ReadFile(relativePath string) (string, error) {
	path, err := reader.resolve(relativePath)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read source file %q: %w", relativePath, err)
	}

	return string(data), nil
}

// ReadContext reads source surrounding a particular line
//
// line is 1-based, matching ripgrep's line numbering
func (reader *Reader) ReadContext(
	relativePath string,
	line int,
	before int,
	after int,
) (Snippet, error) {
	if line < 1 {
		return Snippet{}, fmt.Errorf("line must be at least 1")
	}
	if before < 0 {
		return Snippet{}, fmt.Errorf("before must not be negative")
	}
	if after < 0 {
		return Snippet{}, fmt.Errorf("after must not be negative")
	}

	content, err := reader.ReadFile(relativePath)
	if err != nil {
		return Snippet{}, err
	}

	lines := strings.Split(content, "\n")

	if line > len(lines) {
		return Snippet{}, fmt.Errorf(
			"line %d is outside %q, which has %d lines",
			line,
			relativePath,
			len(lines),
		)
	}

	// Convert the 1-based match line to a 0-based index
	index := line - 1

	start := index - before
	if start < 0 {
		start = 0
	}

	end := index + after + 1
	if end > len(lines) {
		end = len(lines)
	}

	return Snippet{
		Path:      relativePath,
		StartLine: start + 1,
		EndLine:   end,
		Content:   strings.Join(lines[start:end], "\n"),
	}, nil
}

// resolve ensures relativePath remains underneath the configured root.
func (reader *Reader) resolve(relativePath string) (string, error) {
	if filepath.IsAbs(relativePath) {
		return "", fmt.Errorf("source path must be relative: %q", relativePath)
	}

	path := filepath.Join(reader.root, relativePath)
	path = filepath.Clean(path)

	relative, err := filepath.Rel(reader.root, path)
	if err != nil {
		return "", fmt.Errorf("resolve source path %q: %w", relativePath, err)
	}

	if relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf(
			"source path %q escapes codebase root",
			relativePath,
		)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		resolvedRelative, relErr := filepath.Rel(reader.root, resolved)
		if relErr != nil || resolvedRelative == ".." || strings.HasPrefix(resolvedRelative, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("source path %q escapes codebase root", relativePath)
		}
	}

	return path, nil
}
