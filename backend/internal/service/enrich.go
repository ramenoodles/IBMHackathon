package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	llmclient "github.com/ramenoodles/IBMHackathon/backend/internal/llm"
)

type EnrichNodeInput struct {
	ID    string `json:"id"`
	Line  int    `json:"line"`
	Code  string `json:"code"`
	Kind  string `json:"kind"`
	Label string `json:"label,omitempty"`
}

type EnrichUserContext struct {
	PrimaryLanguage string `json:"primaryLanguage"`
	ExperienceLevel string `json:"experienceLevel"`
}

type EnrichRequest struct {
	FilePath    string
	Symbol      string
	Nodes       []EnrichNodeInput
	UserContext EnrichUserContext
}

type EnrichPatch struct {
	ID          string `json:"id"`
	Title       string `json:"title,omitempty"`
	Summary     string `json:"summary"`
	LabelSource string `json:"labelSource,omitempty"`
}

type EnrichResult struct {
	Patches []EnrichPatch `json:"patches"`
}

type enrichLLMResponse struct {
	Patches []EnrichPatch `json:"patches"`
}

func (s *Service) Enrich(ctx context.Context, req EnrichRequest) (EnrichResult, error) {
	if len(req.Nodes) == 0 {
		return EnrichResult{}, fmt.Errorf("nodes are required")
	}

	patches := make([]EnrichPatch, 0, len(req.Nodes))
	patched := make(map[string]struct{}, len(req.Nodes))

	tryLLM := func(nodes []EnrichNodeInput) error {
		if len(nodes) == 0 || s.llm == nil {
			return nil
		}
		enricher, ok := s.llm.(llmclient.EnrichClient)
		if !ok {
			return nil
		}

		raw, err := enricher.EnrichBatch(ctx, buildEnrichPrompt(req, nodes))
		if err != nil {
			return err
		}

		llmPatches, err := parseEnrichResponse(raw)
		if err != nil {
			return err
		}

		allowed := make(map[string]EnrichNodeInput, len(nodes))
		for _, node := range nodes {
			allowed[node.ID] = node
		}
		for _, patch := range llmPatches {
			node, ok := allowed[patch.ID]
			if !ok || strings.TrimSpace(patch.Summary) == "" {
				continue
			}
			title := strings.TrimSpace(patch.Title)
			source := "ai"
			if title == "" || isGenericLabel(title) {
				if fallbackTitle, fallbackSummary, ok := contextualHeuristicLabel(req.Symbol, node.Kind, node.Code, node.Label); ok {
					if title == "" || isGenericLabel(title) {
						title = fallbackTitle
						source = "heuristic"
					}
					if strings.TrimSpace(patch.Summary) == "" || isGenericSummary(patch.Summary) {
						patch.Summary = fallbackSummary
						source = "heuristic"
					}
				}
			}
			patches = append(patches, EnrichPatch{
				ID:          patch.ID,
				Title:       title,
				Summary:     patch.Summary,
				LabelSource: source,
			})
			patched[patch.ID] = struct{}{}
		}
		return nil
	}

	if err := tryLLM(req.Nodes); err != nil {
		return EnrichResult{}, fmt.Errorf("enrich batch: %w", err)
	}

	for _, node := range req.Nodes {
		if _, done := patched[node.ID]; done {
			continue
		}
		if title, summary, ok := contextualHeuristicLabel(req.Symbol, node.Kind, node.Code, node.Label); ok {
			patches = append(patches, EnrichPatch{
				ID:          node.ID,
				Title:       title,
				Summary:     summary,
				LabelSource: "heuristic",
			})
		}
	}

	return EnrichResult{Patches: patches}, nil
}

var genericTitles = []string{
	"acquire lock",
	"release lock",
	"already closed?",
	"exit early",
	"function starts",
	"wake waiting threads",
	"wake one waiter",
	"wait for signal",
	"raise error",
}

func isGenericLabel(title string) bool {
	normalized := strings.ToLower(strings.TrimSpace(title))
	for _, generic := range genericTitles {
		if normalized == generic {
			return true
		}
	}
	return false
}

func isGenericSummary(summary string) bool {
	lower := strings.ToLower(summary)
	return strings.Contains(lower, "only one goroutine can change shared state") ||
		strings.Contains(lower, "resource was shut down already") ||
		strings.Contains(lower, "ending this execution path")
}

func buildEnrichPrompt(req EnrichRequest, nodes []EnrichNodeInput) string {
	var b strings.Builder
	b.WriteString("You are labeling execution-flow steps for a developer onboarding to a codebase.\n")
	b.WriteString("Parent function: ")
	b.WriteString(req.Symbol)
	if req.FilePath != "" {
		b.WriteString(" in ")
		b.WriteString(req.FilePath)
	}
	b.WriteString("\n")
	if lang := strings.TrimSpace(req.UserContext.PrimaryLanguage); lang != "" {
		b.WriteString("Primary language: ")
		b.WriteString(lang)
		b.WriteByte('\n')
	}

	switch strings.ToLower(strings.TrimSpace(req.UserContext.ExperienceLevel)) {
	case "junior":
		b.WriteString("Audience: junior developer — use plain language and briefly define jargon.\n")
	case "senior":
		b.WriteString("Audience: senior developer — focus on intent and non-obvious behavior.\n")
	default:
		b.WriteString("Audience: developer new to this codebase.\n")
	}

	b.WriteString("\nFor each node, produce:\n")
	b.WriteString("- title: at most 8 words, plain English, names the SPECIFIC object/field/callee from the code\n")
	b.WriteString("- summary: one sentence explaining why this step matters inside ")
	b.WriteString(req.Symbol)
	b.WriteString("\n")
	b.WriteString("\nNEVER use vague titles without naming what is involved. Forbidden examples:\n")
	b.WriteString("- \"Acquire lock\", \"Release lock\", \"Already closed?\", \"Exit early\", \"Function starts\"\n")
	b.WriteString("Good examples for m.mu.Lock() in close(): \"Lock broker mutex (m.mu)\", \"Take m.mu before mutating state\"\n")
	b.WriteString("Good for if m.closed: \"Broker already marked closed\", \"Skip when m.closed is set\"\n")
	b.WriteString("Infer what receivers (m, b, self) represent from the parent function and field names.\n")
	b.WriteString("Do not add, remove, or reorder steps. Only reinterpret existing ones.\n")
	b.WriteString("Return ONLY valid JSON: {\"patches\":[{\"id\":\"...\",\"title\":\"...\",\"summary\":\"...\"}]}\n\n")
	b.WriteString("Nodes (id, kind, label, code):\n")

	type promptNode struct {
		ID    string `json:"id"`
		Kind  string `json:"kind"`
		Label string `json:"label,omitempty"`
		Code  string `json:"code"`
		Line  int    `json:"line,omitempty"`
	}
	out := make([]promptNode, 0, len(nodes))
	for _, node := range nodes {
		out = append(out, promptNode{
			ID:    node.ID,
			Kind:  node.Kind,
			Label: node.Label,
			Code:  node.Code,
			Line:  node.Line,
		})
	}
	enc, _ := json.Marshal(out)
	b.Write(enc)

	return b.String()
}

func parseEnrichResponse(raw string) ([]EnrichPatch, error) {
	raw = strings.TrimSpace(raw)
	if i := strings.Index(raw, "{"); i > 0 {
		raw = raw[i:]
	}

	var parsed enrichLLMResponse
	if err := json.NewDecoder(strings.NewReader(raw)).Decode(&parsed); err != nil {
		return nil, err
	}
	return parsed.Patches, nil
}
