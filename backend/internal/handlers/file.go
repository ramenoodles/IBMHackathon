package handlers

import (
	"net/http"

	"github.com/ibmhackathon/onbober/internal/scanner"
)

// File returns the contents of a file within a workspace.
func (h *Handler) File(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	workspace := r.URL.Query().Get("workspace")
	relPath := r.URL.Query().Get("path")
	if workspace == "" || relPath == "" {
		http.Error(w, "workspace and path query parameters required", http.StatusBadRequest)
		return
	}

	content, lang, err := h.scanner.ReadFile(workspace, relPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"content":  content,
		"language": lang,
	})
}

// FileSymbols returns traceable symbols for a file within a workspace.
func (h *Handler) FileSymbols(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	workspace := r.URL.Query().Get("workspace")
	relPath := r.URL.Query().Get("path")
	if workspace == "" || relPath == "" {
		http.Error(w, "workspace and path query parameters required", http.StatusBadRequest)
		return
	}

	content, _, err := h.scanner.ReadFile(workspace, relPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	symbols := scanner.ExtractFileSymbols(content, relPath)
	writeJSON(w, http.StatusOK, map[string]any{
		"symbols": symbols,
		"count":   len(symbols),
	})
}
