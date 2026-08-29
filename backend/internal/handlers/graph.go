package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

	params, ok := parseGraphNodeQuery(r)
	if !ok {
		http.Error(w, "workspace and nodeId are required", http.StatusBadRequest)
		return
	}

	detail, err := h.graphs.BuildNodeDetail(r.Context(), graph.BuildInput{
		WorkspacePath: params.Workspace,
		FilePath:      params.FilePath,
		Symbol:        params.Symbol,
		Experience:    params.Experience,
		Language:      params.Language,
	}, params.NodeID, params.Title, params.Line, params.Confidence, params.Code, params.Kind, params.Summary)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// GraphNodeStream streams a plain-text node explanation via SSE.
func (h *Handler) GraphNodeStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	params, ok := parseGraphNodeQuery(r)
	if !ok {
		http.Error(w, "workspace and nodeId are required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	input := graph.BuildInput{
		WorkspacePath: params.Workspace,
		FilePath:      params.FilePath,
		Symbol:        params.Symbol,
		Experience:    params.Experience,
		Language:      params.Language,
	}

	meta, _ := json.Marshal(graph.NodeDetail{
		ID:         params.NodeID,
		Title:      params.Title,
		Summary:    params.Summary,
		Confidence: params.Confidence,
		File:       params.FilePath,
		Line:       params.Line,
	})
	if err := writeSSE(w, "meta", string(meta)); err != nil {
		return
	}

	tokens, bundle, mock, err := h.graphs.StreamNodeDetailExplanation(r.Context(), input, params.NodeID, params.Code, params.Kind, params.Summary)
	if err != nil || mock {
		mockDetail := graph.MockNodeDetail(params.NodeID, params.Title, params.FilePath, params.Line)
		final := h.graphs.FinalizeStreamedDetail(mockDetail.Explanation, params.Code, bundle)
		final.ID = params.NodeID
		final.Title = params.Title
		final.Summary = params.Summary
		final.Confidence = params.Confidence
		final.File = params.FilePath
		final.Line = params.Line
		final.Mock = true
		for _, ch := range final.Explanation {
			payload := fmt.Sprintf(`{"content":%q,"mock":true}`, string(ch))
			if err := writeSSE(w, "token", payload); err != nil {
				return
			}
		}
		done, _ := json.Marshal(final)
		_ = writeSSE(w, "done", string(done))
		return
	}

	var full strings.Builder
	for token := range tokens {
		full.WriteString(token)
		payload := fmt.Sprintf(`{"content":%q}`, token)
		if err := writeSSE(w, "token", payload); err != nil {
			return
		}
	}

	final := h.graphs.FinalizeStreamedDetail(full.String(), params.Code, bundle)
	final.ID = params.NodeID
	final.Title = params.Title
	final.Summary = params.Summary
	final.Confidence = params.Confidence
	final.File = params.FilePath
	final.Line = params.Line
	done, _ := json.Marshal(final)
	_ = writeSSE(w, "done", string(done))
}

type graphNodeQuery struct {
	Workspace  string
	NodeID     string
	Symbol     string
	FilePath   string
	Title      string
	Code       string
	Kind       string
	Summary    string
	Experience string
	Language   string
	Line       int
	Confidence graph.Confidence
}

func parseGraphNodeQuery(r *http.Request) (graphNodeQuery, bool) {
	workspace := r.URL.Query().Get("workspace")
	nodeID := r.URL.Query().Get("nodeId")
	if workspace == "" || nodeID == "" {
		return graphNodeQuery{}, false
	}
	line, _ := strconv.Atoi(r.URL.Query().Get("line"))
	return graphNodeQuery{
		Workspace:  workspace,
		NodeID:     nodeID,
		Symbol:     r.URL.Query().Get("symbol"),
		FilePath:   r.URL.Query().Get("file"),
		Title:      r.URL.Query().Get("title"),
		Code:       r.URL.Query().Get("code"),
		Kind:       r.URL.Query().Get("kind"),
		Summary:    r.URL.Query().Get("summary"),
		Experience: r.URL.Query().Get("experience"),
		Language:   r.URL.Query().Get("language"),
		Line:       line,
		Confidence: graph.Confidence(r.URL.Query().Get("confidence")),
	}, true
}
