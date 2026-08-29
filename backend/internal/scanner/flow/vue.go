package flow

import (
	"regexp"
	"strings"
)

var scriptBlockPattern = regexp.MustCompile(`(?is)<script([^>]*)>([\s\S]*?)</script>`)

// ExtractVueScript returns script body and detected lang from a Vue SFC.
func ExtractVueScript(content string) (script string, lang string, ok bool) {
	m := scriptBlockPattern.FindStringSubmatch(content)
	if m == nil {
		return "", "", false
	}
	attrs := m[1]
	script = m[2]
	lang = "javascript"
	if strings.Contains(attrs, `lang="ts"`) || strings.Contains(attrs, `lang='ts'`) {
		lang = "typescript"
	}
	return strings.TrimSpace(script), lang, true
}

// PrepareSource normalizes file content for parsing (Vue SFC → script block).
func PrepareSource(content, filePath, lang string) (string, string) {
	if strings.HasSuffix(strings.ToLower(filePath), ".vue") {
		if script, scriptLang, ok := ExtractVueScript(content); ok {
			return script, scriptLang
		}
	}
	if lang == "" {
		lang = LangFromPath(filePath)
	}
	return content, lang
}

// LangFromPath returns language id from file extension.
func LangFromPath(filePath string) string {
	lower := strings.ToLower(filePath)
	switch {
	case strings.HasSuffix(lower, ".py"):
		return "python"
	case strings.HasSuffix(lower, ".go"):
		return "go"
	case strings.HasSuffix(lower, ".vue"), strings.HasSuffix(lower, ".js"), strings.HasSuffix(lower, ".jsx"):
		return "javascript"
	case strings.HasSuffix(lower, ".ts"), strings.HasSuffix(lower, ".tsx"):
		return "typescript"
	default:
		return "text"
	}
}
