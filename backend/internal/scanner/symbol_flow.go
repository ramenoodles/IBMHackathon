package scanner

import (
	"fmt"
	"regexp"
	"strings"
)

// FlowStep is one ordered statement inside a symbol body.
type FlowStep struct {
	Line       int
	Kind       string
	Label      string
	Summary    string
	Code       string
	Confidence string
}

var (
	pythonDefLine = regexp.MustCompile(`(?m)^(\s*)(?:async\s+)?def\s+(\w+)\s*\(`)
	goFuncLine    = regexp.MustCompile(`(?m)^func\s+(?:\([^)]*\)\s+)?(\w+)\s*\(`)
	qualifiedCall = regexp.MustCompile(`([A-Za-z_][\w]*(?:\.[A-Za-z_][\w]*)*)\s*\(`)
	assignPattern = regexp.MustCompile(`^\s*(\w+)\s*=\s*(.+)$`)
)

// ExtractSymbolSteps walks a symbol body in source order and returns traceable steps.
func ExtractSymbolSteps(content, filePath, symbol, lang string) []FlowStep {
	if content == "" || symbol == "" {
		return nil
	}
	lines := strings.Split(content, "\n")
	start, body, ok := locateSymbolBody(lines, symbol, lang)
	if !ok || len(body) == 0 {
		return nil
	}

	steps := []FlowStep{{
		Line:       start,
		Kind:       "entry",
		Label:      symbol + "()",
		Summary:    fmt.Sprintf("Entry: %s", symbol),
		Code:       strings.TrimSpace(lines[start-1]),
		Confidence: "verified",
	}}

	for _, bl := range body {
		lineNum := bl.lineNum
		raw := bl.text
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if br := branchIfPattern.FindStringSubmatch(raw); br != nil {
			label := strings.TrimSpace(trimmed)
			if len(label) > 48 {
				label = label[:45] + "..."
			}
			steps = append(steps, FlowStep{
				Line: lineNum, Kind: "branch", Label: label,
				Summary: describeStep(trimmed, "Branch"), Code: trimmed, Confidence: "verified",
			})
			continue
		}

		if strings.HasPrefix(trimmed, "return") {
			steps = append(steps, FlowStep{
				Line: lineNum, Kind: "return", Label: "return",
				Summary: describeStep(trimmed, "Return"), Code: trimmed, Confidence: "verified",
			})
			continue
		}

		callee, callLabel := primaryCallee(trimmed)
		if callee == "" {
			if m := assignPattern.FindStringSubmatch(trimmed); m != nil {
				steps = append(steps, FlowStep{
					Line: lineNum, Kind: "assign", Label: m[1] + " = …",
					Summary: describeStep(trimmed, "Assign"), Code: trimmed, Confidence: "verified",
				})
			}
			continue
		}

		steps = append(steps, FlowStep{
			Line: lineNum, Kind: "call", Label: callLabel,
			Summary: describeStep(trimmed, "Call "+callee), Code: trimmed, Confidence: "verified",
		})
	}

	return steps
}

type bodyLine struct {
	lineNum int
	text    string
}

func locateSymbolBody(lines []string, symbol, lang string) (startLine int, body []bodyLine, ok bool) {
	defIdx := -1
	baseIndent := 0

	for i, line := range lines {
		if lang == "python" || strings.HasSuffix(lang, "py") || lang == "" {
			m := pythonDefLine.FindStringSubmatch(line)
			if m == nil || m[2] != symbol {
				continue
			}
			defIdx = i
			baseIndent = len(m[1])
			break
		}
		m := goFuncLine.FindStringSubmatch(line)
		if m != nil && m[1] == symbol {
			defIdx = i
			baseIndent = leadingSpaces(line)
			break
		}
	}
	if defIdx < 0 {
		return 0, nil, false
	}

	for i := defIdx + 1; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			body = append(body, bodyLine{lineNum: i + 1, text: line})
			continue
		}
		indent := leadingSpaces(line)
		if indent <= baseIndent {
			break
		}
		body = append(body, bodyLine{lineNum: i + 1, text: line})
	}
	return defIdx + 1, body, true
}

func leadingSpaces(s string) int {
	n := 0
	for _, r := range s {
		if r == ' ' {
			n++
			continue
		}
		if r == '\t' {
			n += 4
			continue
		}
		break
	}
	return n
}

func primaryCallee(line string) (callee, label string) {
	matches := qualifiedCall.FindAllStringSubmatch(line, -1)
	if len(matches) == 0 {
		return "", ""
	}

	// Prefer callee on RHS of assignment, else last call on the line.
	pick := matches[len(matches)-1][1]
	if m := assignPattern.FindStringSubmatch(line); m != nil {
		rhs := m[2]
		rhsMatches := qualifiedCall.FindAllStringSubmatch(rhs, -1)
		if len(rhsMatches) > 0 {
			pick = rhsMatches[0][1]
		}
	}

	if isNoiseCallee(strings.TrimPrefix(pick, "self.")) {
		return "", ""
	}

	short := pick
	if len(short) > 36 {
		short = short[:33] + "..."
	}
	return pick, short + "()"
}

func describeStep(code, fallback string) string {
	c := strings.TrimSpace(code)
	if len(c) > 72 {
		return c[:69] + "..."
	}
	if c == "" {
		return fallback
	}
	return c
}

// FormatStepLabel returns a display label with line number prefix.
func FormatStepLabel(step FlowStep) string {
	if step.Line > 0 {
		return fmt.Sprintf("L%d %s", step.Line, step.Label)
	}
	return step.Label
}

// IsNoiseSymbol reports generic names that are poor graph labels alone.
func IsNoiseSymbol(name string) bool {
	noise := map[string]bool{
		"create": true, "session": true, "get": true, "set": true,
		"open": true, "close": true, "run": true, "init": true,
	}
	return noise[strings.ToLower(name)]
}
