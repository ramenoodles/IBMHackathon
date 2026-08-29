package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ibmhackathon/onbober/internal/llm"
)

// BuildEnrich generates LLM display patches for visible nodes (non-blocking path).
func (b *Builder) BuildEnrich(ctx context.Context, input BuildInput, nodes []EnrichNodeInput) EnrichResult {
	if len(nodes) == 0 {
		return EnrichResult{}
	}
	if len(nodes) > MaxRootNodes {
		nodes = nodes[:MaxRootNodes]
	}

	var patches []EnrichPatch
	var pending []EnrichNodeInput

	for _, n := range nodes {
		key := EnrichKey(input.WorkspacePath, input.FilePath, n.Line)
		if title, summary, ok := b.enrich.Get(key); ok && summary != "" {
			patches = append(patches, EnrichPatch{ID: n.ID, Title: title, Summary: summary})
			continue
		}
		pending = append(pending, n)
	}

	if len(pending) == 0 {
		return EnrichResult{Patches: patches}
	}

	raw, err := b.llm.GenerateEnrichSummaries(ctx, llm.UserContext{
		PrimaryLanguage: input.Language,
		ExperienceLevel: input.Experience,
	}, input.Symbol, pendingPayload(pending))
	if err != nil {
		for _, n := range pending {
			title, summary := fallbackDisplay(n)
			patches = append(patches, EnrichPatch{ID: n.ID, Title: title, Summary: summary})
		}
		return EnrichResult{Patches: patches, Mock: true}
	}

	generated, err := parseEnrichPatches(raw)
	if err != nil {
		for _, n := range pending {
			title, summary := fallbackDisplay(n)
			patches = append(patches, EnrichPatch{ID: n.ID, Title: title, Summary: summary})
		}
		return EnrichResult{Patches: patches, Mock: true}
	}

	for _, p := range generated {
		for _, n := range pending {
			if n.ID == p.ID {
				key := EnrichKey(input.WorkspacePath, input.FilePath, n.Line)
				title := p.Title
				if title == "" {
					title, _ = fallbackDisplay(n)
				}
				b.enrich.Set(key, title, p.Summary)
			}
		}
		patches = append(patches, p)
	}
	return EnrichResult{Patches: patches}
}

func pendingPayload(nodes []EnrichNodeInput) string {
	var b strings.Builder
	for _, n := range nodes {
		fmt.Fprintf(&b, "- id=%s line=%d kind=%s code=%q\n", n.ID, n.Line, n.Kind, truncateStr(n.Code, 120))
	}
	return b.String()
}

func fallbackDisplay(n EnrichNodeInput) (title, summary string) {
	switch n.Kind {
	case "entry":
		return "Function entry", "Execution starts here"
	case "branch":
		return "Conditional check", truncateStr(n.Code, 72)
	case "return":
		return "Return result", "Exit the function with a value"
	case "loop":
		return "Loop", truncateStr(n.Code, 72)
	case "call":
		return "Call helper", truncateStr(n.Code, 72)
	default:
		return truncateStr(n.Code, 40), truncateStr(n.Code, 72)
	}
}

func parseEnrichPatches(raw string) ([]EnrichPatch, error) {
	cleaned := raw
	if idx := strings.Index(raw, "{"); idx >= 0 {
		cleaned = raw[idx:]
	}
	if idx := strings.LastIndex(cleaned, "}"); idx >= 0 {
		cleaned = cleaned[:idx+1]
	}
	var wrapper struct {
		Patches []EnrichPatch `json:"patches"`
	}
	if err := json.Unmarshal([]byte(cleaned), &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Patches, nil
}
