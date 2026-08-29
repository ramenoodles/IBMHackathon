package handlers

import (
	"encoding/json"
	"net/http"
)

// WorkspaceSetupRequest configures a workspace from a local path or GitHub URL.
type WorkspaceSetupRequest struct {
	Source string `json:"source"`
	Path   string `json:"path"`
	URL    string `json:"url"`
}

// WorkspaceSetup registers a workspace from a local directory or GitHub repository.
func (h *Handler) WorkspaceSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req WorkspaceSetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	var workspacePath string
	var err error

	switch req.Source {
	case "local":
		workspacePath, err = h.workspaces.RegisterLocal(req.Path)
	case "github":
		workspacePath, err = h.workspaces.CloneGitHub(req.URL)
	default:
		http.Error(w, "source must be 'local' or 'github'", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"workspacePath": workspacePath})
}

// WorkspaceUpload accepts a zip archive, extracts it, and returns the workspace path.
func (h *Handler) WorkspaceUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(210 << 20); err != nil {
		http.Error(w, "failed to parse upload", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file field required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	workspacePath, err := h.workspaces.ExtractZip(file, header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"workspacePath": workspacePath})
}
