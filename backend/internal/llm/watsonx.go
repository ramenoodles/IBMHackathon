package llm

import (
	"context"
	"fmt"

	wx "github.com/IBM/watsonx-go/pkg/models"

	"grepwrapper/internal/search"
	"grepwrapper/internal/source"
)

type WatsonxClient struct {
	client *wx.Client
	model  string
	tools  *repositoryTools
}

func NewWatsonxClient(model, root, rgBinary string) (*WatsonxClient, error) {
	client, err := wx.NewClient()
	if err != nil {
		return nil, fmt.Errorf("create watsonx client: %w", err)
	}

	reader, err := source.NewReader(root)
	if err != nil {
		return nil, fmt.Errorf("create source reader: %w", err)
	}

	return &WatsonxClient{
		client: client,
		model:  model,
		tools:  newRepositoryTools(reader, search.NewFinder(rgBinary), root),
	}, nil
}

func (w *WatsonxClient) Ask(
	ctx context.Context,
	question string,
	source string,
) (string, error) {
	response, err := w.AskAgent(ctx, question, source)
	return response.Answer, err
}

func (w *WatsonxClient) AskAgent(
	ctx context.Context,
	question string,
	source string,
) (AgentResponse, error) {
	messages := []wx.ChatMessage{
		wx.CreateSystemMessage(
			"You are a software engineering assistant. " +
				"Answer questions using the provided source code. " +
				"Be precise and concise.",
		),
		stringMessage(wx.RoleUser,
			fmt.Sprintf(
				"Question:\n%s\n\nSource code:\n%s",
				question,
				source,
			),
		),
	}

	return w.runAgent(ctx, messages)
}

func (w *WatsonxClient) Explain(
	ctx context.Context,
	source string,
) (string, error) {
	messages := []wx.ChatMessage{
		wx.CreateSystemMessage(
			"You are a software engineering assistant. Explain code precisely and concisely.",
		),
		stringMessage(wx.RoleUser,
			"Explain this implementation:\n\n"+source,
		),
	}

	response, err := w.runAgent(ctx, messages)
	return response.Answer, err
}

const maxAgentSteps = 8

func (w *WatsonxClient) runAgent(
	ctx context.Context,
	messages []wx.ChatMessage,
) (AgentResponse, error) {
	trajectory := make([]TrajectoryEvent, 0, len(messages))
	for _, message := range messages {
		trajectory = append(trajectory, messageEvent(message))
	}

	for step := 0; step < maxAgentSteps; step++ {
		response, err := w.client.Chat(
			w.model,
			messages,
			wx.WithChatTools(w.tools.definitions()...),
			wx.WithChatToolChoice("auto"),
			wx.WithChatTemperature(0.2),
			wx.WithChatMaxTokens(2048),
		)
		if err != nil {
			return AgentResponse{Trajectory: trajectory}, fmt.Errorf("watsonx chat: %w", err)
		}
		if len(response.Choices) == 0 || response.Choices[0].Message == nil {
			return AgentResponse{Trajectory: trajectory}, fmt.Errorf("watsonx returned no message")
		}

		message := response.Choices[0].Message
		if len(message.ToolCalls) == 0 {
			trajectory = append(trajectory, messageEvent(*message))
			return AgentResponse{
				Answer:     message.Content.GetText(),
				Trajectory: trajectory,
			}, nil
		}

		// Watsonx requires message content to be a JSON string when the
		// conversation is sent back after a tool call.
		assistant := *message
		assistant.Content = stringContent(message.Content.GetText())
		messages = append(messages, assistant)
		trajectory = append(trajectory, messageEvent(assistant))
		for _, call := range message.ToolCalls {
			trajectory = append(trajectory, TrajectoryEvent{
				Type:      "tool_call",
				Name:      call.Function.Name,
				ID:        call.ID,
				Arguments: call.Function.Arguments,
			})
			result, err := w.tools.execute(ctx, call)
			if err != nil {
				result = "tool error: " + err.Error()
			}
			messages = append(messages, toolMessage(call.ID, result))
			trajectory = append(trajectory, TrajectoryEvent{
				Type:    "tool_result",
				Name:    call.Function.Name,
				ID:      call.ID,
				Content: result,
			})
		}
	}

	return AgentResponse{Trajectory: trajectory}, fmt.Errorf(
		"watsonx agent exceeded %d tool steps",
		maxAgentSteps,
	)
}

func messageEvent(message wx.ChatMessage) TrajectoryEvent {
	return TrajectoryEvent{
		Type:    "message",
		Role:    message.Role,
		Content: message.Content.GetText(),
	}
}

func stringMessage(role, content string) wx.ChatMessage {
	return wx.CreateChatMessage(role, stringContent(content))
}

func stringContent(content string) wx.ChatMessageContentUnion {
	return wx.ChatMessageContentUnion{StringContent: &content}
}

func toolMessage(id, content string) wx.ChatMessage {
	return wx.ChatMessage{
		Role:       wx.RoleTool,
		Content:    stringContent(content),
		ToolCallID: &id,
	}
}
