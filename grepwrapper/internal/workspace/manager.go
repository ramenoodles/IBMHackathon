package workspace

import (
	"archive/zip"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Workspace struct {
	ID, Name, Source string
	Root             string
	Temporary        bool
	CreatedAt        time.Time
}
type Manager struct {
	mu       sync.RWMutex
	items    map[string]Workspace
	tempRoot string
}

func NewManager() (*Manager, error) {
	root, err := os.MkdirTemp("", "grepwrapper-workspaces-")
	if err != nil {
		return nil, err
	}
	return &Manager{items: map[string]Workspace{}, tempRoot: root}, nil
}
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, w := range m.items {
		if w.Temporary {
			_ = os.RemoveAll(w.Root)
		}
	}
	return os.RemoveAll(m.tempRoot)
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
	path, err := filepath.Abs(path)
	if err != nil {
		return Workspace{}, err
	}
	return m.add(path, filepath.Base(path), "local", false)
}
func (m *Manager) GitHub(ctx context.Context, url string) (Workspace, error) {
	if strings.TrimSpace(url) == "" {
		return Workspace{}, fmt.Errorf("repository URL is required")
	}
	dir, err := os.MkdirTemp(m.tempRoot, "git-")
	if err != nil {
		return Workspace{}, err
	}
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", url, dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = os.RemoveAll(dir)
		return Workspace{}, fmt.Errorf("clone repository: %s", strings.TrimSpace(string(out)))
	}
	return m.add(dir, filepath.Base(strings.TrimSuffix(url, "/")), "github", true)
}
func (m *Manager) Zip(file io.Reader, name string, size int64) (Workspace, error) {
	if size > 200<<20 {
		return Workspace{}, fmt.Errorf("zip upload exceeds 200 MB")
	}
	dir, err := os.MkdirTemp(m.tempRoot, "zip-")
	if err != nil {
		return Workspace{}, err
	}
	archive, err := os.CreateTemp("", "workspace-*.zip")
	if err != nil {
		return Workspace{}, err
	}
	defer os.Remove(archive.Name())
	if _, err = io.Copy(archive, io.LimitReader(file, 200<<20+1)); err != nil {
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
	for _, f := range zr.File {
		target := filepath.Join(dir, filepath.Clean(f.Name))
		rel, err := filepath.Rel(dir, target)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return Workspace{}, fmt.Errorf("zip entry escapes workspace: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err = os.MkdirAll(target, 0755); err != nil {
				return Workspace{}, err
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return Workspace{}, err
		}
		in, err := f.Open()
		if err != nil {
			return Workspace{}, err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err == nil {
			_, err = io.Copy(out, in)
		}
		_ = in.Close()
		_ = out.Close()
		if err != nil {
			return Workspace{}, err
		}
	}
	return m.add(dir, strings.TrimSuffix(name, filepath.Ext(name)), "zip", true)
}
