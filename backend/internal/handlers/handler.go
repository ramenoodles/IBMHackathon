// Package handlers provides HTTP handlers for the OnBober REST and SSE API.
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ibmhackathon/onbober/internal/graph"
	"github.com/ibmhackathon/onbober/internal/llm"
	"github.com/ibmhackathon/onbober/internal/scanner"
	"github.com/ibmhackathon/onbober/internal/workspace"
)

// Handler groups dependencies for HTTP endpoint handlers.
type Handler struct {
	scanner    *scanner.Scanner
	llm        *llm.OllamaClient
	workspaces *workspace.Manager
	graphs     *graph.Builder
}

// NewHandler constructs a Handler with scanner, LLM, and workspace clients.
func NewHandler(llmClient *llm.OllamaClient, workspaces *workspace.Manager) *Handler {
	sc := scanner.New()
	return &Handler{
		scanner:    sc,
		llm:        llmClient,
		workspaces: workspaces,
		graphs:     graph.NewBuilder(sc, llmClient),
	}
}

// Health responds with a simple status payload for readiness checks.
func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// writeJSON marshals v as JSON and writes it to the response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
