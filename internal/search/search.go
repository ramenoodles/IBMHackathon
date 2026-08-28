// Package search wraps ripgrep and turns declaration matches into structured data.
package search

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Query describes a function implementation search.
type Query struct {
	Name     string
	Root     string
	Language string
	Limit    int
}

// Match is one likely implementation location reported by ripgrep.
type Match struct {
	Path string
	Line int
	Text string
}

// Finder invokes a ripgrep executable.
type Finder struct {
	binary string
}

// NewFinder creates a Finder. An empty binary name defaults to "rg".
func NewFinder(binary string) *Finder {
	if strings.TrimSpace(binary) == "" {
		binary = "rg"
	}
	return &Finder{binary: binary}
}

// Find returns likely declaration locations for Query.Name.
func (finder *Finder) Find(ctx context.Context, query Query) ([]Match, error) {
	patterns, globs, err := resolvePatterns(query.Language, query.Name)
	if err != nil {
		return nil, err
	}

	root, err := validateRoot(query.Root)
	if err != nil {
		return nil, err
	}

	args := []string{"--json", "--line-number", "--color=never", "--case-sensitive"}
	for _, glob := range globs {
		args = append(args, "--glob", glob)
	}
	for _, pattern := range patterns {
		args = append(args, "--regexp", pattern)
	}
	args = append(args, "--", ".")

	command := exec.CommandContext(ctx, finder.binary, args...)
	command.Dir = root
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, runErr := command.Output()
	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		if errors.Is(runErr, exec.ErrNotFound) {
			return nil, fmt.Errorf("could not find ripgrep executable %q; install ripgrep or pass -rg PATH", finder.binary)
		}
		detail := strings.TrimSpace(stderr.String())
		if detail != "" {
			return nil, fmt.Errorf("ripgrep failed: %s", detail)
		}
		return nil, fmt.Errorf("ripgrep failed: %w", runErr)
	}

	matches, err := parseMatches(output)
	if err != nil {
		return nil, fmt.Errorf("parse ripgrep output: %w", err)
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Path == matches[j].Path {
			return matches[i].Line < matches[j].Line
		}
		return matches[i].Path < matches[j].Path
	})
	if query.Limit > 0 && len(matches) > query.Limit {
		matches = matches[:query.Limit]
	}
	return matches, nil
}

func validateRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		root = "."
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve search root %q: %w", root, err)
	}
	info, err := os.Stat(absoluteRoot)
	if err != nil {
		return "", fmt.Errorf("open search root %q: %w", root, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("search root %q is not a directory", root)
	}
	return absoluteRoot, nil
}

type rgEvent struct {
	Type string `json:"type"`
	Data struct {
		Path       rgText `json:"path"`
		Lines      rgText `json:"lines"`
		LineNumber int    `json:"line_number"`
	} `json:"data"`
}

type rgText struct {
	Text  string `json:"text"`
	Bytes string `json:"bytes"`
}

func (value rgText) string() (string, error) {
	if value.Text != "" {
		return value.Text, nil
	}
	if value.Bytes == "" {
		return "", nil
	}
	decoded, err := base64.StdEncoding.DecodeString(value.Bytes)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func parseMatches(output []byte) ([]Match, error) {
	decoder := json.NewDecoder(bytes.NewReader(output))
	var matches []Match
	for decoder.More() {
		var event rgEvent
		if err := decoder.Decode(&event); err != nil {
			return nil, err
		}
		if event.Type != "match" {
			continue
		}
		path, err := event.Data.Path.string()
		if err != nil {
			return nil, fmt.Errorf("decode match path: %w", err)
		}
		line, err := event.Data.Lines.string()
		if err != nil {
			return nil, fmt.Errorf("decode matching line: %w", err)
		}
		matches = append(matches, Match{
			Path: filepath.Clean(strings.TrimPrefix(path, "./")),
			Line: event.Data.LineNumber,
			Text: strings.TrimRight(line, "\r\n"),
		})
	}
	return matches, nil
}
