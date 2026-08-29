package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrLLMUnavailable = errors.New("LLM is not configured")
var ErrSymbolNotFound = errors.New("symbol not found")

type ExplainRequest struct {
	Name     string
	Language string
	Question string
	Before   int
	After    int
}

type ExplainResult struct {
	Path      string `json:"path"`
	Line      int    `json:"line"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Answer    string `json:"answer"`
}

func (s *Service) Explain(
	ctx context.Context,
	request ExplainRequest,
) (ExplainResult, error) {
	if s.llm == nil {
		return ExplainResult{}, ErrLLMUnavailable
	}

	if strings.TrimSpace(request.Question) == "" {
		request.Question = "Explain what this code does."
	}

	before := request.Before
	after := request.After

	if before <= 0 {
		before = 10
	}

	if after <= 0 {
		after = 80
	}

	results, err := s.Lookup(
		ctx,
		LookupRequest{
			Name:     request.Name,
			Language: request.Language,
			Limit:    1,
			Context:  true,
			Before:   before,
			After:    after,
		},
	)
	if err != nil {
		return ExplainResult{}, err
	}

	if len(results) == 0 {
		return ExplainResult{}, ErrSymbolNotFound
	}

	match := results[0]

	answer, err := s.llm.Ask(
		ctx,
		request.Question,
		match.Source,
	)
	if err != nil {
		return ExplainResult{},
			fmt.Errorf("analyze %s: %w", request.Name, err)
	}

	return ExplainResult{
		Path:      match.Path,
		Line:      match.Line,
		StartLine: match.StartLine,
		EndLine:   match.EndLine,
		Answer:    answer,
	}, nil
}
