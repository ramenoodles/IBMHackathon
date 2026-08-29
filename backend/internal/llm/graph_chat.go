package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type chatResponse struct {
	Message chatMessage `json:"message"`
}

// ChatComplete sends a non-streaming chat request and returns the full response text.
func (c *OllamaClient) ChatComplete(ctx context.Context, system, user string) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Stream: false,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama status %d: %s", resp.StatusCode, string(b))
	}

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Message.Content, nil
}

// GenerateFlowGraph requests a flow graph JSON from Ollama.
func (c *OllamaClient) GenerateFlowGraph(ctx context.Context, input GraphBuildContext) (string, error) {
	system := BuildGraphSystemPrompt(input.UserContext)
	user := BuildGraphUserPrompt(input)
	return c.ChatComplete(ctx, system, user)
}

// GenerateNodeDetail requests a node detail JSON from Ollama.
func (c *OllamaClient) GenerateNodeDetail(ctx context.Context, ctx2 UserContext, nodeID, symbol, snippet string) (string, error) {
	system := BuildGraphSystemPrompt(ctx2) + "\nReturn ONLY JSON for NodeDetail: {id,title,summary,explanation,confidence,file,line,relatedSymbols}"
	user := fmt.Sprintf("Explain node `%s` for symbol `%s`.\nSource:\n```\n%s\n```", nodeID, symbol, snippet)
	return c.ChatComplete(ctx, system, user)
}
