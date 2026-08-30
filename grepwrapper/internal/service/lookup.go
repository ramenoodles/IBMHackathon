package service

import (
	"context"
	"fmt"

	"grepwrapper/internal/search"
	"grepwrapper/internal/source"

	llmclient "grepwrapper/internal/llm"
)

type LookupRequest struct {
	Name     string
	Language string
	Limit    int
	Context  bool
	Before   int
	After    int
}

type LookupResult struct {
	Path      string `json:"path"`
	Line      int    `json:"line"`
	Text      string `json:"text"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
	Source    string `json:"source,omitempty"`
}

type Service struct {
	root              string         // which codebase we are allowed to inspect
	finder            *search.Finder // finds declarations using ripgrep
	reader            *source.Reader // reads the source code from the files finder locations
	llm               llmclient.Client
	includeTrajectory bool // whether the agent's trajectory should be included in the response
}

func New(
	root string,
	rgBinary string,
	client llmclient.Client,
	includeTrajectory bool,
) (*Service, error) {
	reader, err := source.NewReader(root)
	if err != nil {
		return nil, err
	}

	return &Service{
		root:              root,
		finder:            search.NewFinder(rgBinary),
		reader:            reader,
		llm:               client,
		includeTrajectory: includeTrajectory,
	}, nil
}

func (service *Service) Lookup(
	ctx context.Context,
	request LookupRequest,
) ([]LookupResult, error) {
	limit := request.Limit
	if limit <= 0 {
		limit = 20
	}

	matches, err := service.finder.Find(ctx, search.Query{
		Name:     request.Name,
		Root:     service.root,
		Language: request.Language,
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}

	results := make([]LookupResult, 0, len(matches))

	for _, match := range matches {
		result := LookupResult{
			Path: match.Path,
			Line: match.Line,
			Text: match.Text,
		}

		if request.Context {
			snippet, err := service.reader.ReadContext(
				match.Path,
				match.Line,
				request.Before,
				request.After,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"read context for %s:%d: %w",
					match.Path,
					match.Line,
					err,
				)
			}

			result.StartLine = snippet.StartLine
			result.EndLine = snippet.EndLine
			result.Source = snippet.Content
		}

		results = append(results, result)
	}

	return results, nil
}
