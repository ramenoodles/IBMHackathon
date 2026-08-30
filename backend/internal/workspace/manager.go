package workspace

import (
	"archive/zip"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ramenoodles/IBMHackathon/backend/internal/config"
)

var githubShorthandPattern = regexp.MustCompile(`^[\w.-]+/[\w.-]+$`)

type Workspace struct {
	ID, Name, Source string
	Root             string
	Temporary        bool
	CreatedAt        time.Time
}

// Limits bounds how much untrusted input a workspace creation can consume.
// Zero values fall back to the package defaults.
type Limits struct {
	MaxRepoBytes     int64
	MaxZipFiles      int
	CloneTimeout     time.Duration
	AllowLocalSource bool
	WorkspaceMaxAge  time.Duration
}

type Manager struct {
	mu       sync.RWMutex
	items    map[string]Workspace
	tempRoot string
	limits   Limits
	stop     chan struct{}
}

func NewManager(limits Limits) (*Manager, error) {
	root, err := os.MkdirTemp("", "grepwrapper-workspaces-")
	if err != nil {
		return nil, err
	}
	if limits.MaxRepoBytes <= 0 {
		limits.MaxRepoBytes = int64(config.DefaultMaxRepoBytes)
	}
	if limits.MaxZipFiles <= 0 {
		limits.MaxZipFiles = config.DefaultMaxZipFiles
	}
	if limits.CloneTimeout <= 0 {
		limits.CloneTimeout = config.DefaultCloneTimeout
	}
	if limits.WorkspaceMaxAge <= 0 {
		limits.WorkspaceMaxAge = config.DefaultWorkspaceMaxAge
	}
	m := &Manager{items: map[string]Workspace{}, tempRoot: root, limits: limits, stop: make(chan struct{})}
	go m.evictLoop()
	return m, nil
}
func (m *Manager) Close() error {
	close(m.stop)
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, w := range m.items {
		if w.Temporary {
			_ = os.RemoveAll(w.Root)
		}
	}
	return os.RemoveAll(m.tempRoot)
}

// evictLoop periodically removes workspaces older than WorkspaceMaxAge so the
// VPS does not accumulate clones/zips from every visitor indefinitely.
func (m *Manager) evictLoop() {
	ticker := time.NewTicker(m.limits.WorkspaceMaxAge / 2)
	defer ticker.Stop()
	for {
		select {
		case <-m.stop:
			return
		case <-ticker.C:
			m.evict()
		}
	}
}

func (m *Manager) evict() {
	cutoff := time.Now().Add(-m.limits.WorkspaceMaxAge)
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, w := range m.items {
		if w.CreatedAt.Before(cutoff) {
			if w.Temporary {
				_ = os.RemoveAll(w.Root)
			}
			delete(m.items, id)
		}
	}
}
func (m *Manager) Get(id string) (Workspace, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.items[id]
	return w, ok
}
func (m *Manager) add(root, name, source string, temporary bool) (Workspace, error) {
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return Workspace{}, fmt.Errorf("workspace is not a directory")
	}
	id := randomID()
	w := Workspace{ID: id, Name: name, Source: source, Root: root, Temporary: temporary, CreatedAt: time.Now()}
	m.mu.Lock()
	m.items[id] = w
	m.mu.Unlock()
	return w, nil
}
func randomID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("ws-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("ws-%x", b)
}
func (m *Manager) Local(path string) (Workspace, error) {
	if !m.limits.AllowLocalSource {
		return Workspace{}, fmt.Errorf("local source is disabled on this server")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return Workspace{}, err
	}
	return m.add(path, filepath.Base(path), "local", false)
}
func normalizeGitHubURL(raw string) (string, error) {
	url := strings.TrimSpace(raw)
	if url == "" {
		return "", fmt.Errorf("repository URL is required")
	}
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, ".git")

	lower := strings.ToLower(url)
	switch {
	case githubShorthandPattern.MatchString(url):
		url = "https://github.com/" + url
	case strings.HasPrefix(lower, "https://github.com/"), strings.HasPrefix(lower, "http://github.com/"):
		// already a full URL
	case strings.HasPrefix(lower, "github.com/"):
		url = "https://" + url
	case strings.HasPrefix(lower, "www.github.com/"):
		url = "https://" + url[len("www."):]
	default:
		return "", fmt.Errorf("invalid GitHub repository: use owner/repo or a github.com URL")
	}

	if !strings.HasPrefix(strings.ToLower(url), "https://github.com/") &&
		!strings.HasPrefix(strings.ToLower(url), "http://github.com/") {
		return "", fmt.Errorf("invalid GitHub repository: use owner/repo or a github.com URL")
	}
	return url, nil
}

func githubRepoName(cloneURL string) string {
	trimmed := strings.TrimSuffix(strings.TrimSpace(cloneURL), "/")
	trimmed = strings.TrimSuffix(trimmed, ".git")
	if i := strings.LastIndex(trimmed, "/"); i >= 0 && i < len(trimmed)-1 {
		return trimmed[i+1:]
	}
	return filepath.Base(trimmed)
}

func (m *Manager) GitHub(ctx context.Context, url string) (Workspace, error) {
	cloneURL, err := normalizeGitHubURL(url)
	if err != nil {
		return Workspace{}, err
	}

	// Best-effort pre-flight check against the GitHub API so a huge
	// repository is rejected before any data is downloaded. Ignore failures
	// (rate limits, network, private repos) — the post-clone walk still guards.
	if m.limits.MaxRepoBytes > 0 {
		if size, ok := githubReportedSize(cloneURL); ok && size > m.limits.MaxRepoBytes {
			return Workspace{}, fmt.Errorf("repository exceeds max size (%d bytes)", m.limits.MaxRepoBytes)
		}
	}

	dir, err := os.MkdirTemp(m.tempRoot, "git-")
	if err != nil {
		return Workspace{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, m.limits.CloneTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--single-branch", "--no-tags", cloneURL, dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = os.RemoveAll(dir)
		detail := strings.TrimSpace(string(out))
		if detail == "" {
			detail = err.Error()
		}
		return Workspace{}, fmt.Errorf("clone repository: %s", detail)
	}
	if m.limits.MaxRepoBytes > 0 {
		over, err := dirSizeExceeds(dir, m.limits.MaxRepoBytes)
		if err != nil {
			_ = os.RemoveAll(dir)
			return Workspace{}, fmt.Errorf("check repository size: %s", err)
		}
		if over {
			_ = os.RemoveAll(dir)
			return Workspace{}, fmt.Errorf("repository exceeds max size (%d bytes)", m.limits.MaxRepoBytes)
		}
	}
	return m.add(dir, githubRepoName(cloneURL), "github", true)
}
func (m *Manager) Zip(file io.Reader, name string, size int64) (Workspace, error) {
	maxBytes := m.limits.MaxRepoBytes
	maxFiles := m.limits.MaxZipFiles
	if maxBytes > 0 && size > maxBytes {
		return Workspace{}, fmt.Errorf("zip upload exceeds %d bytes", maxBytes)
	}
	dir, err := os.MkdirTemp(m.tempRoot, "zip-")
	if err != nil {
		return Workspace{}, err
	}
	fail := func(err error) (Workspace, error) {
		_ = os.RemoveAll(dir)
		return Workspace{}, err
	}
	archive, err := os.CreateTemp(m.tempRoot, "workspace-*.zip")
	if err != nil {
		return Workspace{}, err
	}
	defer os.Remove(archive.Name())
	if _, err = io.Copy(archive, io.LimitReader(file, maxBytes+1)); err != nil {
		return Workspace{}, err
	}
	if err = archive.Close(); err != nil {
		return Workspace{}, err
	}
	zr, err := zip.OpenReader(archive.Name())
	if err != nil {
		return Workspace{}, fmt.Errorf("invalid zip: %w", err)
	}
	defer zr.Close()
	var total int64
	var entries int
	for _, f := range zr.File {
		entries++
		if maxFiles > 0 && entries > maxFiles {
			return fail(fmt.Errorf("zip contains too many entries (limit %d)", maxFiles))
		}
		target := filepath.Join(dir, filepath.Clean(f.Name))
		rel, err := filepath.Rel(dir, target)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fail(fmt.Errorf("zip entry escapes workspace: %s", f.Name))
		}
		if f.FileInfo().IsDir() {
			if err = os.MkdirAll(target, 0755); err != nil {
				return fail(err)
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fail(err)
		}
		in, err := f.Open()
		if err != nil {
			return fail(err)
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			_ = in.Close()
			return fail(err)
		}
		limit := maxBytes - total
		if limit < 0 {
			limit = 0
		}
		n, copyErr := io.Copy(out, io.LimitReader(in, limit+1))
		_ = in.Close()
		_ = out.Close()
		if copyErr != nil {
			return fail(copyErr)
		}
		total += n
		if maxBytes > 0 && total > maxBytes {
			return fail(fmt.Errorf("zip expands beyond %d bytes", maxBytes))
		}
	}
	return m.add(dir, strings.TrimSuffix(name, filepath.Ext(name)), "zip", true)
}

// githubReportedSize returns the byte size a public repository reports via the
// GitHub API. The second return value is false when the size could not be
// determined (rate limit, private repo, network error, ...).
func githubReportedSize(cloneURL string) (int64, bool) {
	lower := strings.ToLower(cloneURL)
	rest := strings.TrimPrefix(lower, "https://github.com/")
	rest = strings.TrimPrefix(rest, "http://github.com/")
	parts := strings.Split(rest, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return 0, false
	}
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/"+parts[0]+"/"+parts[1], nil)
	if err != nil {
		return 0, false
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return 0, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, false
	}
	var body struct {
		Size int64 `json:"size"` // reported in KiB
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, false
	}
	return body.Size * 1024, true
}

// dirSizeExceeds walks root (excluding the .git directory) summing regular
// file sizes. It returns true as soon as the running total passes max so a
// very large repository doesn't need a full walk.
func dirSizeExceeds(root string, max int64) (bool, error) {
	var total int64
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		total += info.Size()
		if total > max {
			return fs.SkipAll
		}
		return nil
	})
	if errors.Is(err, fs.SkipAll) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return total > max, nil
}
