package search

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type languageProfile struct {
	name     string
	aliases  []string
	globs    []string
	patterns []string
}

const nameToken = "{{name}}"

var languageProfiles = []languageProfile{
	{
		name:     "go",
		globs:    []string{"*.go"},
		patterns: []string{`^[[:space:]]*func[[:space:]]+(?:\([^)]*\)[[:space:]]*)?{{name}}[[:space:]]*\(`},
	},
	{
		name:    "python",
		aliases: []string{"py"},
		globs:   []string{"*.py"},
		patterns: []string{
			`^[[:space:]]*(?:async[[:space:]]+)?def[[:space:]]+{{name}}[[:space:]]*\(`,
		},
	},
	{
		name:    "javascript",
		aliases: []string{"js"},
		globs:   []string{"*.js", "*.jsx", "*.mjs", "*.cjs"},
		patterns: []string{
			`^[[:space:]]*(?:export[[:space:]]+)?(?:default[[:space:]]+)?(?:async[[:space:]]+)?function[[:space:]]+\*?[[:space:]]*{{name}}[[:space:]]*\(`,
			`^[[:space:]]*(?:export[[:space:]]+)?(?:const|let|var)[[:space:]]+{{name}}[[:space:]]*=[[:space:]]*(?:async[[:space:]]+)?(?:function\b|\(?[^=;]*\)?[[:space:]]*=>)`,
			`^[[:space:]]*(?:(?:static|async|get|set)[[:space:]]+)*\*?[[:space:]]*{{name}}[[:space:]]*\([^)]*\)[[:space:]]*\{`,
		},
	},
	{
		name:    "typescript",
		aliases: []string{"ts"},
		globs:   []string{"*.ts", "*.tsx", "*.mts", "*.cts"},
		patterns: []string{
			`^[[:space:]]*(?:export[[:space:]]+)?(?:default[[:space:]]+)?(?:async[[:space:]]+)?function[[:space:]]+\*?[[:space:]]*{{name}}[[:space:]]*(?:<[^>]+>)?[[:space:]]*\(`,
			`^[[:space:]]*(?:export[[:space:]]+)?(?:const|let|var)[[:space:]]+{{name}}(?:[[:space:]]*:[^=]+)?[[:space:]]*=[[:space:]]*(?:async[[:space:]]+)?(?:function\b|\(?[^=;]*\)?[[:space:]]*=>)`,
		},
	},
	{
		name:     "rust",
		aliases:  []string{"rs"},
		globs:    []string{"*.rs"},
		patterns: []string{`^[[:space:]]*(?:pub(?:\([^)]*\))?[[:space:]]+)?(?:async[[:space:]]+)?(?:unsafe[[:space:]]+)?fn[[:space:]]+{{name}}[[:space:]]*(?:<[^>]+>)?[[:space:]]*\(`},
	},
	{
		name:     "java",
		globs:    []string{"*.java"},
		patterns: []string{cStylePattern},
	},
	{
		name:     "c",
		globs:    []string{"*.c", "*.h"},
		patterns: []string{cStylePattern},
	},
	{
		name:    "cpp",
		aliases: []string{"c++", "cc"},
		globs:   []string{"*.cc", "*.cpp", "*.cxx", "*.hh", "*.hpp", "*.hxx"},
		patterns: []string{
			cStylePattern,
			`^[[:space:]]*(?:template[[:space:]]*<[^>]+>[[:space:]]*)?(?:[[:alnum:]_:$<>,.?*&\[\]]+[[:space:]]+)+[[:alnum:]_:]+::{{name}}[[:space:]]*\(`,
		},
	},
	{
		name:    "csharp",
		aliases: []string{"cs", "c#"},
		globs:   []string{"*.cs"},
		patterns: []string{
			cStylePattern,
		},
	},
}

const cStylePattern = `^[[:space:]]*(?:(?:public|protected|private|internal|static|final|virtual|override|abstract|async|inline|extern|constexpr|friend|synchronized|native|sealed|partial|unsafe|new)[[:space:]]+)*(?:[[:alnum:]_:$<>,.?*&\[\]]+[[:space:]]+)+{{name}}[[:space:]]*\(`

func resolvePatterns(language, name string) ([]string, []string, error) {
	if !validName(name) {
		return nil, nil, fmt.Errorf("%q is not a simple function identifier", name)
	}

	language = strings.ToLower(strings.TrimSpace(language))
	selected := languageProfiles
	if language != "" && language != "auto" {
		selected = nil
		for _, profile := range languageProfiles {
			if profile.name == language || contains(profile.aliases, language) {
				selected = []languageProfile{profile}
				break
			}
		}
		if len(selected) == 0 {
			return nil, nil, fmt.Errorf("unsupported language %q (choose %s)", language, strings.Join(languageNames(), ", "))
		}
	}

	escapedName := regexp.QuoteMeta(name)
	var patterns []string
	var globs []string
	seenPatterns := make(map[string]struct{})
	seenGlobs := make(map[string]struct{})
	for _, profile := range selected {
		for _, pattern := range profile.patterns {
			pattern = strings.ReplaceAll(pattern, nameToken, escapedName)
			if _, exists := seenPatterns[pattern]; !exists {
				seenPatterns[pattern] = struct{}{}
				patterns = append(patterns, pattern)
			}
		}
		for _, glob := range profile.globs {
			if _, exists := seenGlobs[glob]; !exists {
				seenGlobs[glob] = struct{}{}
				globs = append(globs, glob)
			}
		}
	}
	return patterns, globs, nil
}

func validName(name string) bool {
	if name == "" {
		return false
	}
	for index, character := range name {
		if character == '_' || character == '$' || character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || index > 0 && character >= '0' && character <= '9' {
			continue
		}
		return false
	}
	return true
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func languageNames() []string {
	names := make([]string, 0, len(languageProfiles)+1)
	names = append(names, "auto")
	for _, profile := range languageProfiles {
		names = append(names, profile.name)
	}
	sort.Strings(names[1:])
	return names
}
