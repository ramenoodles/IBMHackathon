package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var skipDirs = map[string]bool{
	".git": true, "node_modules": true, ".svn": true, "vendor": true,
}

// ListDirAt returns entries under relDir within workspace, with paths relative to workspace root.
func (s *Scanner) ListDirAt(workspace, relDir string) ([]TreeEntry, error) {
	workspaceAbs, err := filepath.Abs(workspace)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace: %w", err)
	}

	target, err := SafeJoin(workspaceAbs, relDir)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(target)
	if err != nil {
		return nil, fmt.Errorf("directory not found: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory")
	}

	entries, err := os.ReadDir(target)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	relDir = filepath.Clean(relDir)
	if relDir == "." {
		relDir = ""
	}

	var result []TreeEntry
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() && skipDirs[name] {
			continue
		}

		var relPath string
		if relDir == "" {
			relPath = name
		} else {
			relPath = filepath.ToSlash(filepath.Join(relDir, name))
		}

		result = append(result, TreeEntry{
			Name:  name,
			Path:  relPath,
			IsDir: e.IsDir(),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	if len(result) > 500 {
		result = result[:500]
	}
	return result, nil
}

// ListDir returns a shallow listing of files and directories at the given absolute path.
// Deprecated: prefer ListDirAt for workspace-scoped navigation.
func (s *Scanner) ListDir(dirPath string) ([]TreeEntry, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	var result []TreeEntry
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() && skipDirs[name] {
			continue
		}
		result = append(result, TreeEntry{
			Name:  name,
			Path:  name,
			IsDir: e.IsDir(),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	if len(result) > 200 {
		result = result[:200]
	}
	return result, nil
}
