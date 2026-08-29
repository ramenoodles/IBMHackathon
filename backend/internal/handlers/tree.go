package handlers

import (
	"net/http"

	"github.com/ibmhackathon/onbober/internal/scanner"
)

// Tree returns a shallow directory listing within a workspace.
// Query params: workspace (required), dir (optional relative subdirectory).
func (h *Handler) Tree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	workspace := r.URL.Query().Get("workspace")
	dir := r.URL.Query().Get("dir")

	if workspace != "" {
		safeWorkspace, err := scanner.SafePath(workspace)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		entries, err := h.scanner.ListDirAt(safeWorkspace, dir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"entries": entries})
		return
	}

	// Legacy: absolute path listing (paths are entry names only).
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "workspace query parameter required", http.StatusBadRequest)
		return
	}

	safePath, err := scanner.SafePath(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	entries, err := h.scanner.ListDir(safePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"entries": entries})
}
