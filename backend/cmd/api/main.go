// Package main is the entry point for the OnBober API server.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ibmhackathon/onbober/internal/handlers"
	"github.com/ibmhackathon/onbober/internal/llm"
	"github.com/ibmhackathon/onbober/internal/workspace"
)

// main bootstraps the HTTP server with API routes and graceful shutdown.
func main() {
	port := envOrDefault("PORT", "8080")
	ollamaURL := envOrDefault("OLLAMA_URL", "http://localhost:11434")
	ollamaModel := envOrDefault("OLLAMA_MODEL", "llama3.2")

	llmClient := llm.NewOllamaClient(ollamaURL, ollamaModel)

	workspaceRoot := envOrDefault("ONBOBER_WORKSPACE_ROOT", filepath.Join(os.TempDir(), "onbober-workspaces"))
	wsManager, err := workspace.NewManager(workspaceRoot)
	if err != nil {
		log.Fatalf("workspace manager: %v", err)
	}

	h := handlers.NewHandler(llmClient, wsManager)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", h.Health)
	mux.HandleFunc("/api/tree", h.Tree)
	mux.HandleFunc("/api/file", h.File)
	mux.HandleFunc("/api/analyze", h.Analyze)
	mux.HandleFunc("/api/graph/root", h.GraphRoot)
	mux.HandleFunc("/api/graph/expand", h.GraphExpand)
	mux.HandleFunc("/api/graph/enrich", h.GraphEnrich)
	mux.HandleFunc("/api/graph/node", h.GraphNode)
	mux.HandleFunc("/api/workspace/setup", h.WorkspaceSetup)
	mux.HandleFunc("/api/workspace/upload", h.WorkspaceUpload)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handlers.WithCORS(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("OnBober API listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

// envOrDefault returns the environment variable value or a fallback default.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
