package graph

import (
	"strconv"
	"sync"
	"time"
)

type enrichEntry struct {
	title   string
	summary string
	expires time.Time
}

// EnrichCache stores LLM-generated summaries keyed by workspace|file|line.
type EnrichCache struct {
	mu      sync.RWMutex
	entries map[string]enrichEntry
	ttl     time.Duration
}

// NewEnrichCache creates a summary cache.
func NewEnrichCache(ttl time.Duration) *EnrichCache {
	return &EnrichCache{entries: make(map[string]enrichEntry), ttl: ttl}
}

// Get returns cached title and summary for a node.
func (c *EnrichCache) Get(key string) (title, summary string, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, found := c.entries[key]
	if !found || time.Now().After(e.expires) {
		return "", "", false
	}
	return e.title, e.summary, true
}

// Set stores title and summary for a node.
func (c *EnrichCache) Set(key, title, summary string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = enrichEntry{title: title, summary: summary, expires: time.Now().Add(c.ttl)}
}

// EnrichKey builds a cache key for a node summary.
func EnrichKey(workspace, file, nodeID string, line int) string {
	return workspace + "|" + file + "|" + nodeID + "|" + strconv.Itoa(line)
}

// EnrichNodeInput is a node sent for LLM summary enrichment.
type EnrichNodeInput struct {
	ID   string `json:"id"`
	Line int    `json:"line"`
	Code string `json:"code"`
	Kind string `json:"kind"`
}

// EnrichPatch is a display update for one node.
type EnrichPatch struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Summary        string   `json:"summary"`
	RelatedSymbols []string `json:"relatedSymbols,omitempty"`
}

// EnrichResult is the response from graph enrichment.
type EnrichResult struct {
	Patches []EnrichPatch `json:"patches"`
	Mock    bool          `json:"mock,omitempty"`
}
