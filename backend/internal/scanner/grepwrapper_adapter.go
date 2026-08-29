package scanner

// grepwrapper_adapter wires OnBober's scanner to Jack's grepwrapper module.
//
// OnBober code should call GrepSymbol / GrepSymbolLang (grep.go), not this file
// directly. The adapter maps OnBober's Match type to bridge.Finder results and
// preserves soft-fail behavior when ripgrep is unavailable.
//
// See docs/GREPWRAPPER.md for architecture and ownership boundaries.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"grepwrapper/bridge"
)

const maxMatches = 50

var (
	defaultFinder     *bridge.Finder
	defaultFinderOnce sync.Once
)

func sharedFinder() *bridge.Finder {
	defaultFinderOnce.Do(func() {
		defaultFinder = bridge.NewFinder("")
	})
	return defaultFinder
}

func languageForSearch(filePath string) string {
	if filePath == "" {
		return "auto"
	}
	lang := LanguageFromPath(filePath)
	if lang == "" || lang == "text" {
		return "auto"
	}
	return lang
}

func (s *Scanner) grepSymbolViaWrapper(workspacePath, filePath, symbol, lang string) ([]Match, error) {
	if symbol == "" {
		return nil, nil
	}
	if lang == "" {
		lang = languageForSearch(filePath)
	}

	root := workspacePath
	var relFile string
	if filePath != "" {
		joined, err := SafeJoin(workspacePath, filePath)
		if err != nil {
			return nil, err
		}
		info, err := filepath.Abs(joined)
		if err == nil {
			root = info
			if st, statErr := statIsDir(info); statErr == nil && !st {
				root = filepath.Dir(info)
				relFile = filepath.ToSlash(filePath)
			}
		}
	}

	matches, err := sharedFinder().Find(context.Background(), bridge.Query{
		Name:     symbol,
		Root:     root,
		Language: lang,
		Limit:    maxMatches,
	})
	if err != nil {
		// Preserve soft-fail when ripgrep is unavailable (demo mode).
		return []Match{}, nil
	}

	out := make([]Match, 0, len(matches))
	for _, m := range matches {
		if relFile != "" && filepath.ToSlash(m.Path) != relFile &&
			!strings.HasSuffix(filepath.ToSlash(m.Path), "/"+relFile) {
			continue
		}
		out = append(out, Match{
			File:    filepath.ToSlash(m.Path),
			Line:    m.Line,
			Content: strings.TrimSpace(m.Text),
		})
	}
	return out, nil
}

func statIsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// ReadMatchContext returns source lines around a match using grepwrapper's path-safe reader.
func (s *Scanner) ReadMatchContext(workspacePath, relPath string, line, before, after int) (string, error) {
	reader, err := bridge.NewReader(workspacePath)
	if err != nil {
		return "", err
	}
	snippet, err := reader.ReadContext(relPath, line, before, after)
	if err != nil {
		return "", err
	}
	return snippet.Content, nil
}
