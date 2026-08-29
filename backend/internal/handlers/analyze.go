package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ibmhackathon/onbober/internal/llm"
	"github.com/ibmhackathon/onbober/internal/scanner"
)

// UserContext mirrors the frontend onboarding payload.
type UserContext struct {
	PrimaryLanguage string `json:"primaryLanguage"`
	ExperienceLevel string `json:"experienceLevel"`
	WorkspacePath   string `json:"workspacePath"`
}

// AnalyzeRequest is the JSON body for POST /api/analyze.
type AnalyzeRequest struct {
	WorkspacePath string      `json:"workspacePath"`
	FilePath      string      `json:"filePath"`
	Symbol        string      `json:"symbol"`
	UserContext   UserContext `json:"userContext"`
}

// Analyze scans the codebase and streams legacy markdown analysis via SSE.
// Deprecated: prefer POST /api/graph/root for the graph-first UI.
func (h *Handler) Analyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.WorkspacePath == "" {
		http.Error(w, "workspacePath is required", http.StatusBadRequest)
		return
	}

	if req.Symbol == "" && req.FilePath != "" {
		base := filepath.Base(req.FilePath)
		req.Symbol = strings.TrimSuffix(base, filepath.Ext(base))
	}
	if req.Symbol == "" {
		http.Error(w, "symbol or filePath is required", http.StatusBadRequest)
		return
	}

	safeWorkspace, err := scanner.SafePath(req.WorkspacePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	matches, err := h.scanner.GrepSymbol(safeWorkspace, req.FilePath, req.Symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	snippet := ""
	if req.FilePath != "" {
		content, _, readErr := h.scanner.ReadFile(safeWorkspace, req.FilePath)
		if readErr == nil {
			snippet = truncate(content, 4000)
		}
	}

	matchRefs := make([]llm.MatchRef, len(matches))
	for i, m := range matches {
		matchRefs[i] = llm.MatchRef{File: m.File, Line: m.Line, Content: m.Content}
	}

	promptInput := llm.PromptInput{
		Symbol:      req.Symbol,
		FilePath:    req.FilePath,
		Matches:     matchRefs,
		Snippet:     snippet,
		UserContext: llm.UserContext(req.UserContext),
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	tokens, useMock, err := h.llm.StreamChat(r.Context(), promptInput)
	if useMock || err != nil {
		_ = streamMockResponse(w)
		return
	}

	if err := streamTokens(w, tokens, false); err != nil {
		_ = writeSSE(w, "error", fmt.Sprintf(`{"message":%q}`, err.Error()))
	}
}

// truncate limits a string to maxLen runes for prompt inclusion.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... (truncated)"
}
