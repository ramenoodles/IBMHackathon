package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const maxFileSize = 512 * 1024

var binaryExts = map[string]bool{
	".o": true, ".a": true, ".so": true, ".dll": true, ".exe": true,
	".png": true, ".jpg": true, ".gif": true, ".zip": true,
}

// ReadFile reads a file within workspace with size limits and returns detected language.
func (s *Scanner) ReadFile(workspace, relPath string) (string, string, error) {
	full, err := SafeJoin(workspace, relPath)
	if err != nil {
		return "", "", err
	}

	ext := strings.ToLower(filepath.Ext(full))
	if binaryExts[ext] {
		return "", "", fmt.Errorf("binary file not readable")
	}

	info, err := os.Stat(full)
	if err != nil {
		return "", "", fmt.Errorf("stat file: %w", err)
	}
	if info.IsDir() {
		return "", "", fmt.Errorf("path is a directory")
	}
	if info.Size() > maxFileSize {
		return "", "", fmt.Errorf("file exceeds size limit (%d bytes)", maxFileSize)
	}

	data, err := os.ReadFile(full)
	if err != nil {
		return "", "", fmt.Errorf("read file: %w", err)
	}

	return string(data), languageFromExt(ext), nil
}

// languageFromExt maps file extensions to language identifiers.
func languageFromExt(ext string) string {
	switch ext {
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc":
		return "cpp"
	case ".go":
		return "go"
	case ".rs":
		return "rust"
	case ".py":
		return "python"
	case ".sh":
		return "bash"
	default:
		return "text"
	}
}
