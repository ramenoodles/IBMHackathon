package llm

import "context"

type Client interface {
	Ask(
		ctx context.Context,
		question string,
		source string,
	) (string, error)
}

// TrajectoryEvent is one observable step in an agent run.
type TrajectoryEvent struct {
	Type      string `json:"type"`
	Role      string `json:"role,omitempty"`
	Name      string `json:"name,omitempty"`
	ID        string `json:"id,omitempty"`
	Arguments string `json:"arguments,omitempty"`
	Content   string `json:"content,omitempty"`
}

type AgentResponse struct {
	Answer     string
	Trajectory []TrajectoryEvent
}

type AgentClient interface {
	AskAgent(
		ctx context.Context,
		question string,
		source string,
	) (AgentResponse, error)
}
