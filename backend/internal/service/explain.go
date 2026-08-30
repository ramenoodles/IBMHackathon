package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	llmclient "github.com/ramenoodles/IBMHackathon/backend/internal/llm"
)

var ErrLLMUnavailable = errors.New("LLM is not configured")
var ErrSymbolNotFound = errors.New("symbol not found")

type ExplainRequest struct {
	Name     string
	Language string
	Question string
	Before   int
	After    int
	// Node context forwarded from the UI.
	File       string
	Line       int
	Code       string // exact snippet shown in the graph node
	Kind       string // entry / call / branch / return / raise
	Title      string // display label (e.g. "strip()")
	Experience string // junior / mid / senior
}

type ExplainResult struct {
	Path       string                      `json:"path"`
	Line       int                         `json:"line"`
	StartLine  int                         `json:"start_line"`
	EndLine    int                         `json:"end_line"`
	Answer     string                      `json:"explanation"`
	Trajectory []llmclient.TrajectoryEvent `json:"trajectory,omitempty"`
}

// buildQuestion constructs a focused question for the LLM using all available
// node context so it never has to guess the filename, line number, or step kind.
func buildQuestion(req ExplainRequest) string {
	if strings.TrimSpace(req.Question) != "" {
		// Caller supplied an explicit question — honour it, but still inject context.
		return enrichQuestion(req.Question, req)
	}

	var b strings.Builder

	// Location
	if req.File != "" && req.Line > 0 {
		fmt.Fprintf(&b, "In `%s` at line %d", req.File, req.Line)
	} else if req.File != "" {
		fmt.Fprintf(&b, "In `%s`", req.File)
	}

	// Step kind + label
	if req.Kind != "" && req.Title != "" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "there is a `%s` step labelled \"%s\"", req.Kind, req.Title)
	} else if req.Title != "" {
		if b.Len() > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "there is a step labelled \"%s\"", req.Title)
	}

	if b.Len() > 0 {
		b.WriteString(".")
	}

	// Inline code snippet
	if req.Code != "" {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		fmt.Fprintf(&b, "The exact code at this step is:\n```\n%s\n```", strings.TrimSpace(req.Code))
	}

	// Experience calibration
	switch req.Experience {
	case "junior":
		b.WriteString("\n\nExplain this step clearly for a junior developer — define any non-obvious concepts.")
	case "senior":
		b.WriteString("\n\nExplain this step concisely for a senior developer — focus on intent and edge cases only.")
	default:
		b.WriteString("\n\nExplain what this specific step does, why it exists, and what happens if it fails or is skipped.")
	}

	return b.String()
}

// enrichQuestion prepends location/code context to a caller-supplied question.
func enrichQuestion(question string, req ExplainRequest) string {
	var prefix strings.Builder
	if req.File != "" && req.Line > 0 {
		fmt.Fprintf(&prefix, "Context: `%s` line %d", req.File, req.Line)
		if req.Kind != "" {
			fmt.Fprintf(&prefix, " (%s step)", req.Kind)
		}
		prefix.WriteString(".\n")
	}
	if req.Code != "" {
		fmt.Fprintf(&prefix, "Code at this step:\n```\n%s\n```\n\n", strings.TrimSpace(req.Code))
	}
	return prefix.String() + question
}

func (s *Service) Explain(
	ctx context.Context,
	request ExplainRequest,
) (ExplainResult, error) {
	if s.llm == nil {
		return ExplainResult{}, ErrLLMUnavailable
	}

	question := buildQuestion(request)

	// If the frontend supplied the exact code snippet for this node, use it
	// directly as the source — no ripgrep lookup needed, and the LLM gets the
	// precise lines the user is looking at rather than the whole function body.
	if strings.TrimSpace(request.Code) != "" {
		var answer string
		var trajectory []llmclient.TrajectoryEvent
		var err error

		if s.includeTrajectory {
			if agent, ok := s.llm.(llmclient.AgentClient); ok {
				response, agentErr := agent.AskAgent(ctx, question, request.Code)
				answer = response.Answer
				trajectory = response.Trajectory
				err = agentErr
			} else {
				answer, err = s.llm.Ask(ctx, question, request.Code)
			}
		} else {
			answer, err = s.llm.Ask(ctx, question, request.Code)
		}
		if err != nil {
			return ExplainResult{}, fmt.Errorf("analyze %s: %w", request.Name, err)
		}
		return ExplainResult{
			Path:       request.File,
			Line:       request.Line,
			StartLine:  request.Line,
			EndLine:    request.Line,
			Answer:     answer,
			Trajectory: trajectory,
		}, nil
	}

	// No inline code — fall back to ripgrep symbol lookup.
	before := request.Before
	after := request.After
	if before <= 0 {
		before = 5
	}
	if after <= 0 {
		after = 40
	}

	results, err := s.Lookup(ctx, LookupRequest{
		Name:     request.Name,
		Language: request.Language,
		Limit:    1,
		Context:  true,
		Before:   before,
		After:    after,
	})
	if err != nil {
		return ExplainResult{}, err
	}
	if len(results) == 0 {
		return ExplainResult{}, ErrSymbolNotFound
	}

	match := results[0]

	var answer string
	var trajectory []llmclient.TrajectoryEvent
	if s.includeTrajectory {
		if agent, ok := s.llm.(llmclient.AgentClient); ok {
			response, agentErr := agent.AskAgent(ctx, question, match.Source)
			answer = response.Answer
			trajectory = response.Trajectory
			err = agentErr
		} else {
			answer, err = s.llm.Ask(ctx, question, match.Source)
		}
	} else {
		answer, err = s.llm.Ask(ctx, question, match.Source)
	}
	if err != nil {
		return ExplainResult{}, fmt.Errorf("analyze %s: %w", request.Name, err)
	}

	return ExplainResult{
		Path:       match.Path,
		Line:       match.Line,
		StartLine:  match.StartLine,
		EndLine:    match.EndLine,
		Answer:     answer,
		Trajectory: trajectory,
	}, nil
}
