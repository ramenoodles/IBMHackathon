package scanner

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	pyClassRe  = regexp.MustCompile(`^\s*class\s+([A-Za-z_]\w*)`)
	goTypeRe   = regexp.MustCompile(`^\s*type\s+([A-Za-z_]\w*)\s+struct`)
	jsClassRe  = regexp.MustCompile(`^\s*class\s+([A-Za-z_]\w*)`)
	stringLitRe = regexp.MustCompile(`"([^"\\]{2,})"|'([^'\\]{2,})'`)
)

// FindEnclosingSymbol returns the nearest class/type enclosing line (1-based).
func FindEnclosingSymbol(content string, line int) (name, kind string) {
	if line < 1 {
		return "", ""
	}
	lines := strings.Split(content, "\n")
	if line > len(lines) {
		line = len(lines)
	}

	bestIndent := -1
	for i := line - 1; i >= 0; i-- {
		trimmed := lines[i]
		if strings.TrimSpace(trimmed) == "" {
			continue
		}
		indent := leadingIndent(trimmed)
		if m := pyClassRe.FindStringSubmatch(trimmed); len(m) > 1 {
			if indent <= bestIndent || bestIndent < 0 {
				bestIndent = indent
				name, kind = m[1], "class"
			}
			continue
		}
		if m := goTypeRe.FindStringSubmatch(trimmed); len(m) > 1 {
			if indent <= bestIndent || bestIndent < 0 {
				bestIndent = indent
				name, kind = m[1], "type"
			}
			continue
		}
		if m := jsClassRe.FindStringSubmatch(trimmed); len(m) > 1 {
			if indent <= bestIndent || bestIndent < 0 {
				bestIndent = indent
				name, kind = m[1], "class"
			}
		}
	}
	return name, kind
}

func leadingIndent(line string) int {
	n := 0
	for _, r := range line {
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

// ExtractStringLiterals returns distinct string literals from a code snippet.
func ExtractStringLiterals(snippet string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range stringLitRe.FindAllStringSubmatch(snippet, -1) {
		val := m[1]
		if val == "" {
			val = m[2]
		}
		val = strings.TrimSpace(val)
		if len(val) < 2 || seen[val] {
			continue
		}
		seen[val] = true
		out = append(out, val)
	}
	return out
}

// DistinctiveDomainTerms harvests searchable terms from names and literals.
func DistinctiveDomainTerms(parts ...string) []string {
	seen := map[string]bool{}
	var out []string
	for _, part := range parts {
		for _, token := range splitDomainTokens(part) {
			if len(token) < 3 || seen[token] {
				continue
			}
			if isCommonWord(token) {
				continue
			}
			seen[token] = true
			out = append(out, token)
		}
	}
	return out
}

func splitDomainTokens(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var tokens []string
	tokens = append(tokens, s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsUpper(r) && b.Len() > 0 {
			tokens = append(tokens, b.String())
			b.Reset()
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	if b.Len() > 0 {
		tokens = append(tokens, b.String())
	}
	return tokens
}

func isCommonWord(s string) bool {
	switch strings.ToLower(s) {
	case "test", "data", "name", "type", "class", "return", "provider", "job", "file", "path", "self", "true", "false", "none", "str", "int", "def", "func":
		return true
	default:
		return false
	}
}
