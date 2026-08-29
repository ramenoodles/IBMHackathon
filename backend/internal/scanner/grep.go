package scanner

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// Match represents a single ripgrep hit for a symbol.
type Match struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

const maxMatches = 50

// GrepSymbol runs ripgrep to find function definitions and references for symbol.
// Returns an empty slice when ripgrep finds no matches or exits with code 1.
func (s *Scanner) GrepSymbol(workspacePath, filePath, symbol string) ([]Match, error) {
	if symbol == "" {
		return nil, nil
	}

	searchPath := workspacePath
	if filePath != "" {
		joined, err := SafeJoin(workspacePath, filePath)
		if err != nil {
			return nil, err
		}
		searchPath = joined
	}

	pattern := buildSymbolPattern(symbol)
	args := []string{
		"--json",
		"--max-count", fmt.Sprintf("%d", maxMatches),
		"-e", pattern,
		searchPath,
	}

	cmd := exec.Command("rg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return []Match{}, nil
	}

	if err := cmd.Start(); err != nil {
		return []Match{}, nil
	}

	var matches []Match
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		var line rgJSONLine
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			continue
		}
		if line.Type != "match" {
			continue
		}
		matches = append(matches, Match{
			File:    line.Data.Path.Text,
			Line:    int(line.Data.LineNumber),
			Content: strings.TrimSpace(line.Data.Lines.Text),
		})
		if len(matches) >= maxMatches {
			break
		}
	}

	_ = cmd.Wait()
	return matches, nil
}

// buildSymbolPattern creates a ripgrep regex for common definition patterns across languages.
func buildSymbolPattern(symbol string) string {
	escaped := regexp.QuoteMeta(symbol)
	return fmt.Sprintf(
		`(def\s+%s|async\s+def\s+%s|class\s+%s|fn\s+%s|func\s+%s|%s\s*\(|(?:static\s+)?(?:inline\s+)?[\w\s\*]+\s+%s\s*\()`,
		escaped, escaped, escaped, escaped, escaped, escaped, escaped,
	)
}

type rgJSONLine struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		LineNumber int `json:"line_number"`
		Lines      struct {
			Text string `json:"text"`
		} `json:"lines"`
	} `json:"data"`
}
