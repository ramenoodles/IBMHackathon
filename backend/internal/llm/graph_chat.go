package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
func (c *OllamaClient) GenerateNodeDetail(ctx context.Context, ctx2 UserContext, symbol, payload string) (string, error) {
	system := NodeDetailSystemPrompt + `
Return ONLY JSON for NodeDetail: {id,title,summary,explanation,confidence,file,line,relatedSymbols}
Put the plain-text section-tagged explanation in the explanation field.`
	user := fmt.Sprintf("Explain a step in function `%s`.\n%s", symbol, payload)
	return c.ChatComplete(ctx, system, user)
}

// StreamNodeDetailExplanation streams a plain-text node explanation.
func (c *OllamaClient) StreamNodeDetailExplanation(ctx context.Context, userCtx UserContext, symbol, payload string) (<-chan string, bool, error) {
	system := NodeDetailSystemPrompt
	switch userCtx.ExperienceLevel {
	case "senior":
		system += "\nAudience: senior engineer — be terse and precise."
	case "junior":
		system += "\nAudience: junior developer — be plain and welcoming."
	}
	user := fmt.Sprintf("Function: %s\n%s", symbol, payload)
	return c.streamChatMessages(ctx, system, user)
}

func (c *OllamaClient) streamChatMessages(ctx context.Context, system, user string) (<-chan string, bool, error) {
	body, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Stream: true,
	})
	if err != nil {
		return nil, true, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, true, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, true, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, true, fmt.Errorf("ollama status %d", resp.StatusCode)
	}

	tokens := make(chan string)
	go func() {
		defer close(tokens)
		defer resp.Body.Close()
		reader := bufio.NewReader(resp.Body)
		for {
			line, readErr := reader.ReadString('\n')
			if readErr != nil {
				if readErr != io.EOF {
					return
				}
				if line == "" {
					return
				}
			}
			line = strings.TrimSpace(line)
			if line == "" {
				if readErr == io.EOF {
					return
				}
				continue
			}

			var chunk chatStreamLine
			if err := json.Unmarshal([]byte(line), &chunk); err != nil {
				continue
			}
			if chunk.Message.Content != "" {
				tokens <- chunk.Message.Content
			}
			if chunk.Done {
				return
			}
			if readErr == io.EOF {
				return
			}
		}
	}()

	return tokens, false, nil
}
