package scanner

// Match represents a single ripgrep hit for a symbol.
type Match struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

// GrepSymbol runs language-aware ripgrep via the grepwrapper search.Finder.
func (s *Scanner) GrepSymbol(workspacePath, filePath, symbol string) ([]Match, error) {
	return s.grepSymbolViaWrapper(workspacePath, filePath, symbol, "")
}

// GrepSymbolLang is like GrepSymbol but forces a grepwrapper language profile.
func (s *Scanner) GrepSymbolLang(workspacePath, filePath, symbol, lang string) ([]Match, error) {
	return s.grepSymbolViaWrapper(workspacePath, filePath, symbol, lang)
}
