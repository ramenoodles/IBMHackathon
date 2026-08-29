package flow

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	pythonDefLine = regexp.MustCompile(`(?m)^(\s*)(?:async\s+)?def\s+(\w+)\s*\(`)
	goFuncLine    = regexp.MustCompile(`(?m)^func\s+(?:\([^)]*\)\s+)?(\w+)\s*\(`)
	jsFuncLine    = regexp.MustCompile(`(?m)(?:^|\s)(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*\(`)
	jsArrowLine   = regexp.MustCompile(`(?m)(?:^|\s)(?:export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s*)?\([^)]*\)\s*=>`)
	jsMethodLine  = regexp.MustCompile(`(?m)^(\s*)(\w+)\s*\([^)]*\)\s*\{`)
	qualifiedCall = regexp.MustCompile(`([A-Za-z_$][\w$]*(?:\.[A-Za-z_$][\w$]*)*)\s*\(`)
	assignPattern = regexp.MustCompile(`^\s*(\w+)\s*=\s*(.+)$`)
	branchPattern = regexp.MustCompile(`(?m)^\s*(if|elif|else|switch|case)\b`)
	loopPattern   = regexp.MustCompile(`(?m)^\s*(for|while)\b`)
)

type bodyLine struct {
	lineNum int
	text    string
	indent  int
}

// ExtractRegex walks a symbol body using indent + regex heuristics.
func ExtractRegex(content, symbol, lang string) []Step {
	if content == "" || symbol == "" {
		return nil
	}
	lines := strings.Split(content, "\n")
	start, body, ok := locateSymbolBody(lines, symbol, lang)
	if !ok {
		return nil
	}

	steps := []Step{{
		Line: start, Kind: "entry", Label: symbol + "()",
		Summary: fmt.Sprintf("Entry: %s", symbol),
		Code: strings.TrimSpace(lines[start-1]), Confidence: "verified", Source: "regex",
		Indent: leadingSpaces(lines[start-1]),
	}}

	for _, bl := range body {
		trimmed := strings.TrimSpace(bl.text)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}

		if br := branchPattern.FindStringSubmatch(bl.text); br != nil {
			label := trimmed
			if len(label) > 48 {
				label = label[:45] + "..."
			}
			steps = append(steps, Step{
				Line: bl.lineNum, Kind: "branch", Label: label,
				Summary: describe(trimmed), Code: trimmed, Confidence: "verified",
				BranchKind: br[1], Indent: bl.indent, Source: "regex",
			})
			continue
		}

		if lp := loopPattern.FindStringSubmatch(bl.text); lp != nil {
			label := trimmed
			if len(label) > 48 {
				label = label[:45] + "..."
			}
			steps = append(steps, Step{
				Line: bl.lineNum, Kind: "loop", Label: label,
				Summary: describe(trimmed), Code: trimmed, Confidence: "verified",
				LoopKind: lp[1], Indent: bl.indent, Source: "regex",
			})
			continue
		}

		if strings.HasPrefix(trimmed, "return") || strings.HasPrefix(trimmed, "raise") {
			kind := "return"
			if strings.HasPrefix(trimmed, "raise") {
				kind = "raise"
			}
			steps = append(steps, Step{
				Line: bl.lineNum, Kind: kind, Label: kind,
				Summary: describe(trimmed), Code: trimmed, Confidence: "verified",
				Indent: bl.indent, Source: "regex",
			})
			continue
		}

		callee, callLabel := primaryCallee(trimmed)
		if callee == "" {
			if m := assignPattern.FindStringSubmatch(trimmed); m != nil {
				steps = append(steps, Step{
					Line: bl.lineNum, Kind: "assign", Label: m[1] + " = …",
					Summary: describe(trimmed), Code: trimmed, Confidence: "verified",
					Indent: bl.indent, Source: "regex",
				})
			}
			continue
		}

		sym := calleeBaseName(callee)
		steps = append(steps, Step{
			Line: bl.lineNum, Kind: "call", Label: callLabel,
			Summary: describe(trimmed), Code: trimmed, Confidence: "verified",
			CalleeSymbol: sym, CalleeQualified: callee, Indent: bl.indent, Source: "regex",
		})
	}
	return steps
}

func locateSymbolBody(lines []string, symbol, lang string) (startLine int, body []bodyLine, ok bool) {
	defIdx := -1
	baseIndent := 0

	for i, line := range lines {
		found := false
		switch {
		case lang == "python" || strings.HasSuffix(lang, "py"):
			m := pythonDefLine.FindStringSubmatch(line)
			if m != nil && m[2] == symbol {
				defIdx, baseIndent, found = i, len(m[1]), true
			}
		case lang == "go":
			m := goFuncLine.FindStringSubmatch(line)
			if m != nil && m[1] == symbol {
				defIdx, baseIndent, found = i, leadingSpaces(line), true
			}
		case lang == "javascript" || lang == "typescript":
			if m := jsFuncLine.FindStringSubmatch(line); m != nil && m[1] == symbol {
				defIdx, baseIndent, found = i, leadingSpaces(line), true
			} else if m := jsArrowLine.FindStringSubmatch(line); m != nil && m[1] == symbol {
				defIdx, baseIndent, found = i, leadingSpaces(line), true
			} else if m := jsMethodLine.FindStringSubmatch(line); m != nil && m[2] == symbol {
				defIdx, baseIndent, found = i, len(m[1]), true
			}
		}
		if found {
			break
		}
	}
	if defIdx < 0 {
		return 0, nil, false
	}

	for i := defIdx + 1; i < len(lines); i++ {
		line := lines[i]
		indent := leadingSpaces(line)
		if strings.TrimSpace(line) != "" && indent <= baseIndent {
			break
		}
		body = append(body, bodyLine{lineNum: i + 1, text: line, indent: indent})
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
	pick := matches[len(matches)-1][1]
	if m := assignPattern.FindStringSubmatch(line); m != nil {
		rhsMatches := qualifiedCall.FindAllStringSubmatch(m[2], -1)
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

func calleeBaseName(qualified string) string {
	parts := strings.Split(qualified, ".")
	return parts[len(parts)-1]
}

func isNoiseCallee(name string) bool {
	noise := map[string]bool{
		"if": true, "for": true, "while": true, "return": true, "print": true,
		"len": true, "str": true, "int": true, "dict": true, "list": true,
		"range": true, "super": true, "type": true, "set": true, "get": true,
		"console": true, "expect": true, "assert": true,
	}
	return noise[name] || strings.ToUpper(name) == name
}

func describe(code string) string {
	c := strings.TrimSpace(code)
	if len(c) > 72 {
		return c[:69] + "..."
	}
	return c
}

// FormatLabel returns display label with line prefix.
func FormatLabel(s Step) string {
	if s.Line > 0 {
		return fmt.Sprintf("L%d %s", s.Line, s.Label)
	}
	return s.Label
}
