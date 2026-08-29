// Package workspace manages local, uploaded, and cloned codebases.
package workspace

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ibmhackathon/onbober/internal/scanner"
)

// Manager handles workspace registration from multiple sources.
type Manager struct {
	root string
}

// NewManager creates a workspace manager with the given storage root directory.
func NewManager(root string) (*Manager, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create workspace root: %w", err)
	}
	return &Manager{root: root}, nil
}

// RegisterLocal validates an existing directory on disk.
func (m *Manager) RegisterLocal(path string) (string, error) {
	return scanner.SafePath(path)
}

var githubURLPattern = regexp.MustCompile(`^(?:https?://)?(?:www\.)?github\.com/([\w.-]+)/([\w.-]+?)(?:\.git)?/?$`)
var githubShortPattern = regexp.MustCompile(`^([\w.-]+)/([\w.-]+)$`)

// CloneGitHub shallow-clones a public GitHub repository into the workspace root.
func (m *Manager) CloneGitHub(url string) (string, error) {
	owner, repo, err := parseGitHubURL(url)
	if err != nil {
		return "", err
	}

	dest := filepath.Join(m.root, fmt.Sprintf("%s-%s-%d", owner, repo, time.Now().Unix()))
	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)

	cmd := exec.Command("git", "clone", "--depth", "1", cloneURL, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git clone failed (is git installed?): %w", err)
	}

	return dest, nil
}

// parseGitHubURL extracts owner and repo from common GitHub URL formats.
func parseGitHubURL(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", fmt.Errorf("github url is required")
	}

	if matches := githubURLPattern.FindStringSubmatch(raw); len(matches) == 3 {
		return matches[1], strings.TrimSuffix(matches[2], ".git"), nil
	}
	if matches := githubShortPattern.FindStringSubmatch(raw); len(matches) == 3 {
		return matches[1], matches[2], nil
	}
	return "", "", fmt.Errorf("invalid github url: use https://github.com/owner/repo or owner/repo")
}

const maxZipSize = 200 * 1024 * 1024 // 200MB

// ExtractZip saves and extracts an uploaded zip archive into the workspace root.
func (m *Manager) ExtractZip(r io.Reader, filename string) (string, error) {
	safeName := sanitizeFilename(filename)
	if safeName == "" {
		safeName = "upload"
	}

	dest := filepath.Join(m.root, fmt.Sprintf("%s-%d", strings.TrimSuffix(safeName, ".zip"), time.Now().Unix()))
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", err
	}

	zipPath := filepath.Join(dest, safeName)
	out, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}

	written, err := io.Copy(out, io.LimitReader(r, maxZipSize+1))
	out.Close()
	if err != nil {
		os.RemoveAll(dest)
		return "", err
	}
	if written > maxZipSize {
		os.RemoveAll(dest)
		return "", fmt.Errorf("zip exceeds maximum size (%d MB)", maxZipSize/(1024*1024))
	}

	extractDir := filepath.Join(dest, "src")
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		os.RemoveAll(dest)
		return "", err
	}
	if err := unzip(zipPath, extractDir); err != nil {
		os.RemoveAll(dest)
		return "", err
	}
	_ = os.Remove(zipPath)

	// If zip contained a single top-level folder, use that as workspace root.
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		os.RemoveAll(dest)
		return "", err
	}
	if len(entries) == 1 && entries[0].IsDir() {
		return filepath.Join(extractDir, entries[0].Name()), nil
	}
	return extractDir, nil
}

// unzip extracts all files from zipPath into dest.
func unzip(zipPath, dest string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	for _, f := range reader.File {
		target, err := safeZipTarget(destAbs, f.Name)
		if err != nil {
			return err
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := extractZipFile(f, target); err != nil {
			return err
		}
	}
	return nil
}

// safeZipTarget prevents zip slip path traversal attacks.
func safeZipTarget(destAbs, name string) (string, error) {
	target := filepath.Join(destAbs, filepath.Clean(strings.ReplaceAll(name, "\\", "/")))
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(targetAbs, destAbs+string(os.PathSeparator)) && targetAbs != destAbs {
		return "", fmt.Errorf("illegal zip path: %s", name)
	}
	return targetAbs, nil
}

// extractZipFile writes a single zip entry to disk.
func extractZipFile(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

// sanitizeFilename strips unsafe characters from an upload filename.
func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
	return name
}

// PingGit checks whether git is available on PATH.
func PingGit() error {
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git not found on PATH")
	}
	return nil
}

// IsGitHubReachable performs a lightweight HEAD request to github.com.
func IsGitHubReachable() bool {
	resp, err := http.Head("https://github.com")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode < 500
}
