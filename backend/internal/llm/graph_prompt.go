package llm

import (
	"fmt"
	"strings"

	"github.com/ibmhackathon/onbober/internal/scanner"
)

const graphSystemPrompt = `You are OnBober, a codebase onboarding assistant that outputs execution-flow graphs as JSON.

Return ONLY a JSON object matching this schema (no markdown, no prose):
{
  "rootId": "string",
  "symbol": "string",
  "depth": 1,
  "nodes": [
    {
      "id": "unique_id",
      "label": "short label max 24 chars",
      "summary": "one line purpose",
      "kind": "entry|call|branch|return|side_effect",
      "confidence": "verified|inferred",
      "file": "optional/path",
      "line": 0,
      "code": "optional source line excerpt",
      "expandable": false,
      "childCount": 0,
      "collapsed": false
    }
  ],
  "edges": [{ "from": "id", "to": "id", "label": "optional" }]
}

Rules:
- Show execution path inside the function first, then up to 1 level of callees.
- Max 8 nodes for root graphs, max 6 for expansions.
- Mark nodes verified only when supported by provided scan data; otherwise inferred.
- For branch points with many paths, set collapsed=true and childCount=N instead of listing all branches.
- Labels must include line numbers when known (e.g. "L12 create_engine()").
- Summaries should quote or paraphrase the actual source line at that step.
- Edges must form one connected path from rootId — no orphan subgraphs.
- Put the source line excerpt in the "code" field when available.
- Use valid JSON only.`

// GraphBuildContext carries scan data for graph prompt construction.
type GraphBuildContext struct {
	Symbol      string
	FilePath    string
	Snippet     string
	Matches     []MatchRef
	Callees     []scanner.CalleeRef
	Branches    []scanner.BranchRef
	UserContext UserContext
	ParentNode  string
	ExpandLimit int
}

// BuildGraphSystemPrompt returns the system prompt for JSON graph generation.
func BuildGraphSystemPrompt(ctx UserContext) string {
	prompt := graphSystemPrompt
	switch strings.ToLower(ctx.ExperienceLevel) {
	case "junior":
		prompt += "\nAudience: junior developer. Use plain language in summaries."
	case "senior":
		prompt += "\nAudience: senior engineer. Be terse and precise in summaries."
	default:
		prompt += "\nAudience: mid-level developer."
	}
	return prompt
}

// BuildGraphUserPrompt formats the user message for graph generation.
func BuildGraphUserPrompt(input GraphBuildContext) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Build an execution flow graph for symbol `%s`", input.Symbol)
	if input.FilePath != "" {
		fmt.Fprintf(&b, " in `%s`", input.FilePath)
	}
	b.WriteString(".\nReturn ONLY valid JSON matching the FlowGraph schema.\n\n")

	if len(input.Callees) > 0 {
		b.WriteString("Verified callees from source (in call order when possible):\n")
		for _, c := range input.Callees {
			fmt.Fprintf(&b, "- %s (line %d)\n", c.Name, c.Line)
		}
		b.WriteString("\n")
	}
	if len(input.Branches) > 0 {
		b.WriteString("Branch points detected:\n")
		for _, br := range input.Branches {
			fmt.Fprintf(&b, "- line %d: %s\n", br.Line, br.Label)
		}
		b.WriteString("\n")
	}
	if len(input.Matches) > 0 {
		b.WriteString("Ripgrep matches:\n")
		for _, m := range input.Matches {
			fmt.Fprintf(&b, "- %s:%d %s\n", m.File, m.Line, m.Content)
		}
		b.WriteString("\n")
	}
	if input.Snippet != "" {
		b.WriteString("Source snippet:\n```\n")
		b.WriteString(input.Snippet)
		b.WriteString("\n```\n")
	}
	if input.ParentNode != "" {
		fmt.Fprintf(&b, "\nExpand children for collapsed node `%s`. Limit to %d nodes.\n", input.ParentNode, input.ExpandLimit)
	}
	return b.String()
}
