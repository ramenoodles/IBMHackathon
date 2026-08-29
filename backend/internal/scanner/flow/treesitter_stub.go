//go:build !cgo

package flow

// ExtractTreeSitter is unavailable without CGO; returns nil so regex is used.
func ExtractTreeSitter(content, symbol, lang string) []Step {
	return nil
}
