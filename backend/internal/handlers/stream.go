package handlers

import (
	"fmt"
	"net/http"
)

// writeSSE writes a Server-Sent Event frame to the response writer.
func writeSSE(w http.ResponseWriter, event, data string) error {
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return err
	}
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

// streamTokens writes each token as an SSE event with JSON payload {"content":"..."}.
func streamTokens(w http.ResponseWriter, tokens <-chan string, mock bool) error {
	for token := range tokens {
		payload := fmt.Sprintf(`{"content":%q`, token)
		if mock {
			payload += `,"mock":true`
		}
		payload += "}"
		if err := writeSSE(w, "token", payload); err != nil {
			return err
		}
	}
	return writeSSE(w, "done", `{}`)
}

const mockResponse = `## Overview

This is a **demo analysis** because Ollama is not reachable. [INFERRED]

The selected symbol participates in a typical kernel-style call chain.

## Call Flow

[VERIFIED] Entry point delegates to helper routines before returning.

` + "```mermaid\nflowchart TD\n    A[Caller] --> B[SelectedSymbol]\n    B --> C[HelperFunction]\n    C --> D[Return]\n```"

// streamMockResponse writes a canned demo response when Ollama is unavailable.
func streamMockResponse(w http.ResponseWriter) error {
	for _, ch := range mockResponse {
		payload := fmt.Sprintf(`{"content":%q,"mock":true}`, string(ch))
		if err := writeSSE(w, "token", payload); err != nil {
			return err
		}
	}
	return writeSSE(w, "done", `{}`)
}
