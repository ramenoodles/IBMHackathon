package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ibmhackathon/onbober/internal/graph"
)

// GraphRootRequest is the body for POST /api/graph/root.
type GraphRootRequest struct {
	WorkspacePath string      `json:"workspacePath"`
	FilePath      string      `json:"filePath"`
	Symbol        string      `json:"symbol"`
	UserContext   UserContext `json:"userContext"`
}

// GraphExpandRequest is the body for POST /api/graph/expand.
type GraphExpandRequest struct {
	WorkspacePath string      `json:"workspacePath"`
	FilePath      string      `json:"filePath"`
	Symbol        string      `json:"symbol"`
	NodeID        string      `json:"nodeId"`
	ParentPath    []string    `json:"parentPath"`
	ExpandLimit   int         `json:"expandLimit"`
	UserContext   UserContext `json:"userContext"`
}

// GraphRoot builds the initial execution-flow graph for a symbol.
func (h *Handler) GraphRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GraphRootRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.WorkspacePath == "" || req.Symbol == "" {
		http.Error(w, "workspacePath and symbol are required", http.StatusBadRequest)
		return
	}

	g, err := h.graphs.BuildRoot(r.Context(), graph.BuildInput{
		WorkspacePath: req.WorkspacePath,
		FilePath:      req.FilePath,
		Symbol:        req.Symbol,
		Experience:    req.UserContext.ExperienceLevel,
		Language:      req.UserContext.PrimaryLanguage,
		Depth:         1,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, g)
}

// GraphExpand lazily expands a collapsed branch node.
func (h *Handler) GraphExpand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GraphExpandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.WorkspacePath == "" || req.NodeID == "" {
		http.Error(w, "workspacePath and nodeId are required", http.StatusBadRequest)
		return
	}

	g, err := h.graphs.BuildExpand(r.Context(), graph.BuildInput{
		WorkspacePath: req.WorkspacePath,
		FilePath:      req.FilePath,
		Symbol:        req.Symbol,
		NodeID:        req.NodeID,
		ParentPath:    req.ParentPath,
		ExpandLimit:   req.ExpandLimit,
		Experience:    req.UserContext.ExperienceLevel,
		Language:      req.UserContext.PrimaryLanguage,
		Depth:         len(req.ParentPath) + 1,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, g)
}

// GraphNode returns detailed explanation for a single node.
func (h *Handler) GraphNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	workspace := r.URL.Query().Get("workspace")
	nodeID := r.URL.Query().Get("nodeId")
	symbol := r.URL.Query().Get("symbol")
	filePath := r.URL.Query().Get("file")
	title := r.URL.Query().Get("title")
	line, _ := strconv.Atoi(r.URL.Query().Get("line"))
	confidence := graph.Confidence(r.URL.Query().Get("confidence"))

	if workspace == "" || nodeID == "" {
		http.Error(w, "workspace and nodeId are required", http.StatusBadRequest)
		return
	}

	detail, err := h.graphs.BuildNodeDetail(context.Background(), graph.BuildInput{
		WorkspacePath: workspace,
		FilePath:      filePath,
		Symbol:        symbol,
		Experience:    r.URL.Query().Get("experience"),
		Language:      r.URL.Query().Get("language"),
	}, nodeID, title, line, confidence, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}
