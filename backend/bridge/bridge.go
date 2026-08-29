// Package bridge exposes grepwrapper search and source APIs for external modules
// It exposes the stable search primitives without leaking internal packages.
package bridge

import (
	"context"

	"grepwrapper/internal/search"
	"grepwrapper/internal/source"
)

// Query describes a symbol search (alias of search.Query).
type Query = search.Query

// Match is one ripgrep hit (alias of search.Match).
type Match = search.Match

// Snippet is a source context window (alias of source.Snippet).
type Snippet = source.Snippet

// Finder wraps ripgrep symbol search.
type Finder struct {
	inner *search.Finder
}

// NewFinder creates a Finder. Empty binary defaults to "rg".
func NewFinder(rgBinary string) *Finder {
	return &Finder{inner: search.NewFinder(rgBinary)}
}

// Find runs language-aware ripgrep for symbol declarations.
func (f *Finder) Find(ctx context.Context, query Query) ([]Match, error) {
	return f.inner.Find(ctx, query)
}

// Reader reads path-safe source under a repository root.
type Reader struct {
	inner *source.Reader
}

// NewReader creates a Reader for the given codebase root.
func NewReader(root string) (*Reader, error) {
	r, err := source.NewReader(root)
	if err != nil {
		return nil, err
	}
	return &Reader{inner: r}, nil
}

// ReadContext returns lines around a 1-based line number.
func (r *Reader) ReadContext(relativePath string, line, before, after int) (Snippet, error) {
	return r.inner.ReadContext(relativePath, line, before, after)
}
