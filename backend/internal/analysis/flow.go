package analysis

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type flowStep struct {
	Line            int
	Kind            string
	Label           string
	Summary         string
	Code            string
	CalleeSymbol    string
	CalleeQualified string
	BranchKind      string
	Indent          int
}

type bodyLine struct {
	line   int
	text   string
	indent int
}

type Symbol struct {
	Name      string `json:"name"`
	Line      int    `json:"line"`
	Kind      string `json:"kind"`
	Signature string `json:"signature"`
}

var (
	callPattern    = regexp.MustCompile(`([A-Za-z_$][\w$]*(?:\.[A-Za-z_$][\w$]*)*)\s*\(`)
	assignPattern  = regexp.MustCompile(`^\s*(?:[A-Za-z_$][\w$]*(?:\s*,\s*[A-Za-z_$][\w$]*)*\s*(?::=|=)|(?:const|let|var)\s+[A-Za-z_$][\w$]*\s*=)`)
	pythonDecl     = regexp.MustCompile(`^\s*(?:async\s+)?def\s+([A-Za-z_]\w*)\s*\(`)
	goDecl         = regexp.MustCompile(`^\s*func\s+(?:\([^)]*\)\s+)?([A-Za-z_]\w*)\s*\(`)
	rustDecl       = regexp.MustCompile(`^\s*(?:pub(?:\([^)]*\))?\s+)?(?:async\s+)?(?:unsafe\s+)?fn\s+([A-Za-z_]\w*)\s*(?:<[^>]+>)?\s*\(`)
	jsFunctionDecl = regexp.MustCompile(`^\s*(?:export\s+)?(?:default\s+)?(?:async\s+)?function\s+\*?\s*([A-Za-z_$][\w$]*)\s*(?:<[^>]+>)?\s*\(`)
	jsArrowDecl    = regexp.MustCompile(`^\s*(?:export\s+)?(?:const|let|var)\s+([A-Za-z_$][\w$]*)(?:\s*:[^=]+)?\s*=\s*(?:async\s+)?(?:function\b|\(?[^=;]*\)?\s*=>)`)
	jsMethodDecl   = regexp.MustCompile(`^\s*(?:(?:public|private|protected|static|async|get|set)\s+)*\*?\s*([A-Za-z_$][\w$]*)\s*\([^)]*\)\s*(?::[^\{]+)?\s*\{`)
	cStyleDecl     = regexp.MustCompile(`^\s*(?:(?:public|protected|private|internal|static|final|virtual|inline|extern|constexpr|async)\s+)*(?:[\w:<>,.?*&\[\]]+\s+)+([A-Za-z_]\w*)\s*\(`)
	elseIfPattern  = regexp.MustCompile(`^(?:}\s*)?else\s+if\b|^elif\b`)
	ifPattern      = regexp.MustCompile(`^if\b`)
	elsePattern    = regexp.MustCompile(`^(?:}\s*)?else\b`)
	switchPattern  = regexp.MustCompile(`^(?:switch|case)\b`)
	loopPattern    = regexp.MustCompile(`^(?:for|while)\b`)
	exitPattern    = regexp.MustCompile(`^(?:return|raise|throw)\b`)
)

func LanguageFromPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".py":
		return "python"
	case ".go":
		return "go"
	case ".js", ".jsx", ".mjs", ".cjs", ".vue":
		return "javascript"
	case ".ts", ".tsx", ".mts", ".cts":
		return "typescript"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cc", ".cpp", ".cxx", ".hh", ".hpp", ".hxx":
		return "cpp"
	case ".cs":
		return "csharp"
	default:
		return "text"
	}
}

func ExtractSymbols(content, file string) []Symbol {
	language := LanguageFromPath(file)
	lines := strings.Split(content, "\n")
	seen := make(map[string]bool)
	out := make([]Symbol, 0)
	for i, line := range lines {
		name := declarationName(line, language)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, Symbol{Name: name, Line: i + 1, Kind: "function", Signature: strings.TrimSpace(line)})
	}
	return out
}

func extractFlow(content, file, symbol string) ([]flowStep, error) {
	language := LanguageFromPath(file)
	if language == "text" {
		return nil, fmt.Errorf("unsupported flow language for %q", file)
	}
	lines := strings.Split(content, "\n")
	declaration := -1
	for i, line := range lines {
		if declarationName(line, language) == symbol {
			declaration = i
			break
		}
	}
	if declaration < 0 {
		return nil, fmt.Errorf("symbol %q not found in %q", symbol, file)
	}

	var body []bodyLine
	if language == "python" {
		body = pythonBody(lines, declaration)
	} else {
		body = braceBody(lines, declaration)
	}
	steps := []flowStep{{
		Line: declaration + 1, Kind: "entry", Label: symbol + "()",
		Summary: "Function entry - execution starts here", Code: strings.TrimSpace(lines[declaration]),
		Indent: leadingSpaces(lines[declaration]),
	}}
	for _, line := range body {
		if step, ok := classifyLine(line); ok {
			steps = append(steps, step)
		}
	}
	return steps, nil
}

func declarationName(line, language string) string {
	var patterns []*regexp.Regexp
	switch language {
	case "python":
		patterns = []*regexp.Regexp{pythonDecl}
	case "go":
		patterns = []*regexp.Regexp{goDecl}
	case "rust":
		patterns = []*regexp.Regexp{rustDecl}
	case "javascript", "typescript":
		patterns = []*regexp.Regexp{jsFunctionDecl, jsArrowDecl, jsMethodDecl}
	default:
		patterns = []*regexp.Regexp{cStyleDecl}
	}
	for _, pattern := range patterns {
		if match := pattern.FindStringSubmatch(line); len(match) > 1 {
			return match[1]
		}
	}
	return ""
}

func pythonBody(lines []string, declaration int) []bodyLine {
	base := leadingSpaces(lines[declaration])
	body := make([]bodyLine, 0)
	for i := declaration + 1; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) != "" && leadingSpaces(line) <= base {
			break
		}
		body = append(body, bodyLine{line: i + 1, text: line, indent: leadingSpaces(line)})
	}
	return body
}

func braceBody(lines []string, declaration int) []bodyLine {
	body := make([]bodyLine, 0)
	depth := 0
	started := false
	inBlockComment := false
	for i := declaration; i < len(lines); i++ {
		delta, opened, closed := braceCounts(lines[i], &inBlockComment)
		if !started {
			if !opened {
				continue
			}
			started = true
			depth += delta
			if closed && depth <= 0 {
				break
			}
			continue
		}
		if depth <= 0 {
			break
		}
		depth += delta
		if depth <= 0 {
			break
		}
		body = append(body, bodyLine{line: i + 1, text: lines[i], indent: leadingSpaces(lines[i])})
	}
	return body
}

func braceCounts(line string, inBlockComment *bool) (delta int, opened, closed bool) {
	var quote rune
	escaped := false
	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		character := runes[i]
		if *inBlockComment {
			if character == '*' && i+1 < len(runes) && runes[i+1] == '/' {
				*inBlockComment = false
				i++
			}
			continue
		}
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if character == '\\' {
				escaped = true
				continue
			}
			if character == quote {
				quote = 0
			}
			continue
		}
		if character == '/' && i+1 < len(runes) {
			switch runes[i+1] {
			case '/':
				return delta, opened, closed
			case '*':
				*inBlockComment = true
				i++
				continue
			}
		}
		if character == '\'' || character == '"' || character == '`' {
			quote = character
			continue
		}
		switch character {
		case '{':
			delta++
			opened = true
		case '}':
			delta--
			closed = true
		}
	}
	return delta, opened, closed
}

func classifyLine(line bodyLine) (flowStep, bool) {
	code := strings.TrimSpace(line.text)
	if code == "" || code == "{" || code == "}" || strings.HasPrefix(code, "//") || strings.HasPrefix(code, "#") {
		return flowStep{}, false
	}
	clean := strings.TrimSpace(stripLineComment(code))
	branchKind := ""
	switch {
	case elseIfPattern.MatchString(clean):
		branchKind = "elif"
	case ifPattern.MatchString(clean):
		branchKind = "if"
	case elsePattern.MatchString(clean):
		branchKind = "else"
	case switchPattern.MatchString(clean):
		branchKind = "if"
	}
	if branchKind != "" {
		return newStep(line, "branch", shorten(clean, 48), "Conditional branch", branchKind, ""), true
	}
	if loopPattern.MatchString(clean) {
		return newStep(line, "loop", shorten(clean, 48), "Loop", "", ""), true
	}
	if exitPattern.MatchString(clean) {
		kind := "return"
		if strings.HasPrefix(clean, "raise") || strings.HasPrefix(clean, "throw") {
			kind = "raise"
		}
		return newStep(line, kind, shorten(clean, 44), "Exit the function", "", ""), true
	}
	if callee, qualified := primaryCallee(clean); callee != "" {
		return newStep(line, "call", qualified+"()", "Calls "+strings.ReplaceAll(callee, "_", " "), "", callee), true
	}
	if assignPattern.MatchString(clean) {
		return newStep(line, "assign", shorten(clean, 44), "Assign: "+shorten(clean, 60), "", ""), true
	}
	return flowStep{}, false
}

func newStep(line bodyLine, kind, label, summary, branchKind, callee string) flowStep {
	return flowStep{
		Line: line.line, Kind: kind, Label: label, Summary: summary,
		Code: strings.TrimSpace(line.text), BranchKind: branchKind, Indent: line.indent,
		CalleeSymbol: callee, CalleeQualified: strings.TrimSuffix(label, "()"),
	}
}

func primaryCallee(line string) (string, string) {
	matches := callPattern.FindAllStringSubmatch(line, -1)
	for i := len(matches) - 1; i >= 0; i-- {
		qualified := matches[i][1]
		parts := strings.Split(qualified, ".")
		base := parts[len(parts)-1]
		if !noiseCallee(base) {
			return base, qualified
		}
	}
	return "", ""
}

func noiseCallee(name string) bool {
	switch name {
	// Go / Python builtins
	case "if", "for", "while", "switch", "return", "len", "str", "int", "dict", "list", "range", "type", "set", "append", "make", "new", "delete", "panic", "print", "println":
		return true
	// JS/TS constructors — upper-case constructor names that appear on the RHS
	// of declarations (new Set(), new Map(), new Array(), …) and carry no
	// interesting call semantics of their own.
	case "Set", "Map", "Array", "Object", "Promise", "Error", "WeakMap", "WeakSet", "WeakRef", "Date", "RegExp", "URL", "URLSearchParams":
		return true
	// Vue / reactive primitives — these are type-wrapper calls, not
	// meaningful domain calls (ref<T>(), reactive<T>(), computed(() => …), …).
	case "ref", "reactive", "computed", "watch", "watchEffect", "readonly", "shallowRef", "shallowReactive", "toRef", "toRefs", "markRaw", "triggerRef":
		return true
	}
	return false
}

func stripLineComment(line string) string {
	var quote rune
	escaped := false
	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		character := runes[i]
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if character == '\\' {
				escaped = true
				continue
			}
			if character == quote {
				quote = 0
			}
			continue
		}
		if character == '\'' || character == '"' || character == '`' {
			quote = character
			continue
		}
		if character == '#' || character == '/' && i+1 < len(runes) && runes[i+1] == '/' {
			return string(runes[:i])
		}
	}
	return line
}

func leadingSpaces(line string) int {
	count := 0
	for _, character := range line {
		switch character {
		case ' ':
			count++
		case '\t':
			count += 4
		default:
			return count
		}
	}
	return count
}

func shorten(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	return value[:max-3] + "..."
}
