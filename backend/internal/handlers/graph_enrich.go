package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ibmhackathon/onbober/internal/graph"
)

// GraphEnrichRequest is the body for POST /api/graph/enrich.
type GraphEnrichRequest struct {
	WorkspacePath string                  `json:"workspacePath"`
	FilePath      string                  `json:"filePath"`
	Symbol        string                  `json:"symbol"`
	Nodes         []graph.EnrichNodeInput `json:"nodes"`
	UserContext   UserContext             `json:"userContext"`
}

// GraphEnrich returns LLM summary patches for flow nodes.
func (h *Handler) GraphEnrich(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GraphEnrichRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.WorkspacePath == "" || req.Symbol == "" {
		http.Error(w, "workspacePath and symbol are required", http.StatusBadRequest)
		return
	}

	result := h.graphs.BuildEnrich(r.Context(), graph.BuildInput{
		WorkspacePath: req.WorkspacePath,
		FilePath:      req.FilePath,
		Symbol:        req.Symbol,
		Experience:    req.UserContext.ExperienceLevel,
		Language:      req.UserContext.PrimaryLanguage,
	}, req.Nodes)
	writeJSON(w, http.StatusOK, result)
}
