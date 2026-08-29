package graph

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

type cacheEntry struct {
	graph     FlowGraph
	expiresAt time.Time
}

// Cache stores recently built flow graphs in memory.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

// NewCache creates an in-memory graph cache with the given TTL.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

// Get returns a cached graph when present and not expired.
func (c *Cache) Get(key string) (FlowGraph, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return FlowGraph{}, false
	}
	return entry.graph, true
}

// Set stores a graph in the cache.
func (c *Cache) Set(key string, graph FlowGraph) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{graph: graph, expiresAt: time.Now().Add(c.ttl)}
}

// CacheKey builds a stable cache key from build parameters.
func CacheKey(workspace, symbol, file, nodeID string, parentPath []string, depth int) string {
	parts := []string{workspace, symbol, file, nodeID, strings.Join(parentPath, ">"), strconv.Itoa(depth)}
	return strings.Join(parts, "|")
}
