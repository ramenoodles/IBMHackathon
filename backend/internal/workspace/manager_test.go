package workspace

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ramenoodles/IBMHackathon/backend/internal/config"
)

func TestNormalizeGitHubURL(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"torvalds/linux", "https://github.com/torvalds/linux"},
		{"IBM/sarama", "https://github.com/IBM/sarama"},
		{"https://github.com/torvalds/linux", "https://github.com/torvalds/linux"},
		{"https://github.com/torvalds/linux/", "https://github.com/torvalds/linux"},
		{"https://github.com/torvalds/linux.git", "https://github.com/torvalds/linux"},
		{"github.com/torvalds/linux", "https://github.com/torvalds/linux"},
	}

	for _, tc := range tests {
		got, err := normalizeGitHubURL(tc.in)
		if err != nil {
			t.Fatalf("normalizeGitHubURL(%q) error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("normalizeGitHubURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeGitHubURLRejectsInvalid(t *testing.T) {
	_, err := normalizeGitHubURL("not-a-repo")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestGitHubRepoName(t *testing.T) {
	if got := githubRepoName("https://github.com/torvalds/linux"); got != "linux" {
		t.Fatalf("githubRepoName = %q", got)
	}
}

func TestNewManagerAppliesLimits(t *testing.T) {
	m, err := NewManager(Limits{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Close() })
	if m.limits.MaxRepoBytes != int64(config.DefaultMaxRepoBytes) {
		t.Fatalf("MaxRepoBytes = %d", m.limits.MaxRepoBytes)
	}
	if m.limits.MaxZipFiles != config.DefaultMaxZipFiles {
		t.Fatalf("MaxZipFiles = %d", m.limits.MaxZipFiles)
	}
	if m.limits.CloneTimeout != config.DefaultCloneTimeout {
		t.Fatalf("CloneTimeout = %v", m.limits.CloneTimeout)
	}
}

func TestZipRejectsTooManyEntries(t *testing.T) {
	archive := buildTestZip(t, 3, 4)
	m, err := NewManager(Limits{MaxZipFiles: 2})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Close() })
	if _, err := m.Zip(bytes.NewReader(archive), "bundle.zip", int64(len(archive))); err == nil || !strings.Contains(err.Error(), "too many entries") {
		t.Fatalf("Zip() error = %v, want entry limit", err)
	}
}

func TestZipRejectsExpansionBeyondLimit(t *testing.T) {
	archive := buildTestZip(t, 3, 512)
	m, err := NewManager(Limits{MaxRepoBytes: 500})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Close() })
	if _, err := m.Zip(bytes.NewReader(archive), "bundle.zip", int64(len(archive))); err == nil || !strings.Contains(err.Error(), "expands beyond") {
		t.Fatalf("Zip() error = %v, want expansion limit", err)
	}
}

func TestZipRejectsOversizedUpload(t *testing.T) {
	archive := buildTestZip(t, 1, 4)
	m, err := NewManager(Limits{MaxRepoBytes: 8})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Close() })
	if _, err := m.Zip(bytes.NewReader(archive), "bundle.zip", 1024); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("Zip() error = %v, want upload size limit", err)
	}
}

func TestDirSizeExceeds(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.txt"), make([]byte, 50), 0o600); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "sub", "b.txt")
	if err := os.MkdirAll(filepath.Dir(sub), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sub, make([]byte, 60), 0o600); err != nil {
		t.Fatal(err)
	}
	gitObj := filepath.Join(root, ".git", "objects")
	if err := os.MkdirAll(gitObj, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitObj, "x"), make([]byte, 1<<20), 0o600); err != nil {
		t.Fatal(err)
	}

	over, err := dirSizeExceeds(root, 100)
	if err != nil {
		t.Fatal(err)
	}
	if !over {
		t.Fatal("dirSizeExceeds(root, 100) = false, want true (100 + .git should be excluded)")
	}
	if over, err := dirSizeExceeds(root, 1<<20); err != nil || over {
		t.Fatalf("dirSizeExceeds(root, large) = %v, %v; want false, nil (git dir ignored)", over, err)
	}
}

func buildTestZip(t *testing.T, files int, size int) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for i := 0; i < files; i++ {
		w, err := zw.Create("file" + string(rune('a'+i)) + ".txt")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(make([]byte, size)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
