package scanner

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// GrepLiteral searches the workspace for a fixed string (not a symbol definition pattern).
func (s *Scanner) GrepLiteral(workspacePath, pattern string, max int) ([]Match, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" || max <= 0 {
		return nil, nil
	}

	args := []string{
		"--json",
		"--max-count", fmt.Sprintf("%d", max),
		"-F", pattern,
		workspacePath,
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
		if len(matches) >= max {
			break
		}
	}
	_ = cmd.Wait()
	return matches, nil
}
