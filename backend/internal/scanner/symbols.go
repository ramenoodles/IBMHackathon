package scanner

import (
	"regexp"
	"strings"
)

// FileSymbol is a traceable function or class in a source file.
type FileSymbol struct {
	Name      string `json:"name"`
	Line      int    `json:"line"`
	Kind      string `json:"kind"`
	Signature string `json:"signature"`
	Hint      string `json:"hint,omitempty"`
}

var reservedNames = map[string]bool{
	"if": true, "for": true, "while": true, "return": true,
	"import": true, "from": true, "const": true, "let": true, "var": true,
}

// ExtractFileSymbols returns functions and classes with line-anchored signatures.
func ExtractFileSymbols(content, filePath string) []FileSymbol {
	ext := fileExtension(filePath)
	lines := strings.Split(content, "\n")
	seen := make(map[string]bool)
	var symbols []FileSymbol

	var patterns []struct {
		re   *regexp.Regexp
		kind string
	}

	switch ext {
	case "py":
		patterns = []struct {
			re   *regexp.Regexp
			kind string
		}{
			{regexp.MustCompile(`^\s*(?:async\s+)?def\s+([A-Za-z_]\w*)\s*\(`), "function"},
			{regexp.MustCompile(`^\s*class\s+([A-Za-z_]\w*)`), "class"},
		}
	case "go":
		patterns = []struct {
			re   *regexp.Regexp
			kind string
		}{
			{regexp.MustCompile(`^\s*func\s+(?:\([^)]*\)\s+)?([A-Za-z_]\w*)\s*\(`), "function"},
		}
	case "rs":
		patterns = []struct {
			re   *regexp.Regexp
			kind string
		}{
			{regexp.MustCompile(`^\s*(?:pub\s+)?fn\s+([A-Za-z_]\w*)\s*\(`), "function"},
		}
	default:
		patterns = []struct {
			re   *regexp.Regexp
			kind string
		}{
			{regexp.MustCompile(`^\s*(?:static\s+|inline\s+)?[\w\s\*]+\s+([A-Za-z_]\w*)\s*\(`), "function"},
			{regexp.MustCompile(`^\s*class\s+([A-Za-z_]\w*)`), "class"},
		}
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		for _, p := range patterns {
			m := p.re.FindStringSubmatch(line)
			if len(m) < 2 {
				continue
			}
			name := m[1]
			if reservedNames[name] || seen[name] {
				continue
			}
			seen[name] = true
			symbols = append(symbols, FileSymbol{
				Name:      name,
				Line:      i + 1,
				Kind:      p.kind,
				Signature: strings.TrimSpace(line),
				Hint:      docstringHint(lines, i),
			})
		}
	}
	return symbols
}

func docstringHint(lines []string, defLineIdx int) string {
	for j := defLineIdx + 1; j < len(lines) && j < defLineIdx+4; j++ {
		t := strings.TrimSpace(lines[j])
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, `"""`) || strings.HasPrefix(t, "'''") {
			inner := strings.Trim(t, `"'`)
			if inner != "" {
				return truncateRunes(inner, 120)
			}
			if j+1 < len(lines) {
				return truncateRunes(strings.TrimSpace(lines[j+1]), 120)
			}
		}
		if strings.HasPrefix(t, "//") {
			return truncateRunes(strings.TrimPrefix(t, "//"), 120)
		}
		if strings.HasPrefix(t, "#") && !strings.HasPrefix(t, "#!") {
			return truncateRunes(strings.TrimPrefix(t, "#"), 120)
		}
		break
	}
	return ""
}

func fileExtension(path string) string {
	if idx := strings.LastIndex(path, "."); idx >= 0 {
		return strings.ToLower(path[idx+1:])
	}
	return ""
}

func truncateRunes(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
