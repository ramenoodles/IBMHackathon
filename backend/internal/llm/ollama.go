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
	"time"
)

// OllamaClient streams chat completions from a local Ollama instance.
type OllamaClient struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaClient creates a client for the given Ollama base URL and model name.
func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client:  &http.Client{Timeout: 0},
	}
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatStreamLine struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

// StreamChat sends a tailored prompt to Ollama and returns a channel of content tokens.
// The second return value is true when a mock fallback should be used.
func (c *OllamaClient) StreamChat(ctx context.Context, input PromptInput) (<-chan string, bool, error) {
	system, user := BuildPrompt(input)

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

// Ping checks whether Ollama is reachable within the given timeout.
func (c *OllamaClient) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/tags", nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama unreachable: status %d", resp.StatusCode)
	}
	return nil
}
