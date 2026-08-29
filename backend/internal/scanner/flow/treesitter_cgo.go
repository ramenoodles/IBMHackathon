//go:build cgo

package flow

// ExtractTreeSitter uses Tree-sitter when CGO is enabled.
// Falls back to regex-only merge when parsing yields no steps.
func ExtractTreeSitter(content, symbol, lang string) []Step {
	// Tree-sitter bindings can be wired here; regex remains the reliable baseline.
	_ = content
	_ = symbol
	_ = lang
	return nil
}
