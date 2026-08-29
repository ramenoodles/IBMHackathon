package scanner

import (
	"path/filepath"
	"regexp"
	"strings"
)

// SymbolDef locates a symbol definition in source.
type SymbolDef struct {
	File string
	Line int
}

// CalleeRef is a function call extracted from source.
type CalleeRef struct {
	Name       string
	Line       int
	Confidence string
}

// BranchRef is a control-flow branch point in source.
type BranchRef struct {
	Label string
	Line  int
	Kind  string
}

var pythonDefPattern = regexp.MustCompile(`(?m)^\s*(?:async\s+)?def\s+(\w+)`)
var pythonCallPattern = regexp.MustCompile(`\b([A-Za-z_]\w*)\s*\(`)
var goCallPattern = regexp.MustCompile(`\b([A-Za-z_]\w*)\s*\(`)
var branchIfPattern = regexp.MustCompile(`(?m)^\s*(if|elif|else|switch|case)\b`)

// FindSymbolDefinition locates the first definition line for a symbol in a file.
func (s *Scanner) FindSymbolDefinition(workspace, filePath, symbol string) (SymbolDef, bool) {
	content, _, err := s.ReadFile(workspace, filePath)
	if err != nil {
		return SymbolDef{}, false
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, "def "+symbol) || strings.Contains(line, "func "+symbol) ||
			strings.Contains(line, symbol+"(") && strings.Contains(line, "{") {
			return SymbolDef{File: filePath, Line: i + 1}, true
		}
	}
	return SymbolDef{}, false
}

// FindCalleesInSnippet extracts likely function calls from a code snippet.
func FindCalleesInSnippet(snippet, lang string, max int) []CalleeRef {
	pattern := pythonCallPattern
	if lang == "go" || lang == "c" || lang == "cpp" {
		pattern = goCallPattern
	}

	seen := map[string]bool{}
	var refs []CalleeRef
	lines := strings.Split(snippet, "\n")
	for i, line := range lines {
		for _, m := range pattern.FindAllStringSubmatch(line, -1) {
			name := m[1]
			if isNoiseCallee(name) || seen[name] {
				continue
			}
			seen[name] = true
			refs = append(refs, CalleeRef{Name: name, Line: i + 1, Confidence: "verified"})
			if len(refs) >= max {
				return refs
			}
		}
	}
	return refs
}

// FindBranchesInSnippet detects branch points in a snippet.
func FindBranchesInSnippet(snippet string, max int) []BranchRef {
	var refs []BranchRef
	lines := strings.Split(snippet, "\n")
	for i, line := range lines {
		m := branchIfPattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		label := strings.TrimSpace(line)
		if len(label) > 40 {
			label = label[:40] + "..."
		}
		refs = append(refs, BranchRef{Label: label, Line: i + 1, Kind: m[1]})
		if len(refs) >= max {
			break
		}
	}
	return refs
}

// LanguageFromPath returns a language id from a file path.
func LanguageFromPath(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".py":
		return "python"
	case ".go":
		return "go"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc":
		return "cpp"
	case ".rs":
		return "rust"
	default:
		return "text"
	}
}

func isNoiseCallee(name string) bool {
	noise := map[string]bool{
		"if": true, "for": true, "while": true, "return": true, "print": true,
		"len": true, "str": true, "int": true, "dict": true, "list": true,
		"range": true, "super": true, "type": true, "set": true, "get": true,
	}
	return noise[name] || strings.ToUpper(name) == name
}
