package llm

import (
	"context"
	"fmt"

	wx "github.com/IBM/watsonx-go/pkg/models"
)

type WatsonxClient struct {
	client *wx.Client
	model  string
}

func NewWatsonxClient(model string) (*WatsonxClient, error) {
	client, err := wx.NewClient()
	if err != nil {
		return nil, fmt.Errorf("create watsonx client: %w", err)
	}

	return &WatsonxClient{
		client: client,
		model:  model,
	}, nil
}

func (w *WatsonxClient) Ask(
	ctx context.Context,
	question string,
	source string,
) (string, error) {
	messages := []wx.ChatMessage{
		wx.CreateSystemMessage(
			"You are a software engineering assistant. " +
				"Answer questions using the provided source code. " +
				"Be precise and concise.",
		),
		wx.CreateUserMessage(
			fmt.Sprintf(
				"Question:\n%s\n\nSource code:\n%s",
				question,
				source,
			),
		),
	}

	response, err := w.client.Chat(
		w.model,
		messages,
		wx.WithChatTemperature(0.2),
		wx.WithChatMaxTokens(600),
	)
	if err != nil {
		return "", fmt.Errorf("watsonx chat: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("watsonx returned no choices")
	}

	return response.Choices[0].Message.Content.GetText(), nil
}

func (w *WatsonxClient) Explain(
	ctx context.Context,
	source string,
) (string, error) {
	messages := []wx.ChatMessage{
		wx.CreateSystemMessage(
			"You are a software engineering assistant. Explain code precisely and concisely.",
		),
		wx.CreateUserMessage(
			"Explain this implementation:\n\n" + source,
		),
	}

	response, err := w.client.Chat(
		w.model,
		messages,
		wx.WithChatTemperature(0.2),
		wx.WithChatMaxTokens(600),
	)
	if err != nil {
		return "", fmt.Errorf("watsonx chat: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("watsonx returned no choices")
	}

	return response.Choices[0].Message.Content.GetText(), nil
}
