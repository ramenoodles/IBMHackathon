package llm

import (
	"fmt"
	"strings"
)

// BuildPrompt constructs system and user messages tailored to the developer profile.
func BuildPrompt(input PromptInput) (system string, user string) {
	system = buildSystemPrompt(input.UserContext)
	user = buildUserPrompt(input)
	return system, user
}

// buildSystemPrompt returns experience- and language-aware system instructions.
func buildSystemPrompt(ctx UserContext) string {
	base := `You are OnBober, an expert codebase onboarding assistant for complex C/C++ repositories.

Your response MUST include:
1. A clear explanation using ## headings
2. Verification badges: [VERIFIED] for directly supported claims, [INFERRED] for reasonable guesses
3. A fenced mermaid code block (flowchart TD or sequenceDiagram) showing call flow or architecture

Keep diagrams focused (max 8 nodes). Use valid Mermaid syntax only.`

	switch strings.ToLower(ctx.ExperienceLevel) {
	case "junior":
		base += "\n\nAudience: Junior developer. Use plain language, analogies, and step-by-step explanations. Avoid unexplained jargon."
	case "senior":
		base += "\n\nAudience: Senior engineer. Be concise. Focus on invariants, edge cases, and call chains."
	default:
		base += "\n\nAudience: Mid-level developer. Balance clarity with technical depth."
	}

	switch ctx.PrimaryLanguage {
	case "Python":
		base += " Relate concepts to Python where helpful."
	case "Go":
		base += " Relate concepts to Go where helpful."
	case "Rust":
		base += " Relate memory/ownership concepts to Rust where helpful."
	}

	return base
}

// buildUserPrompt formats scan results and source snippet for the user message.
func buildUserPrompt(input PromptInput) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Analyze the symbol `%s`", input.Symbol)
	if input.FilePath != "" {
		fmt.Fprintf(&b, " in file `%s`", input.FilePath)
	}
	b.WriteString(".\n\n")

	if len(input.Matches) > 0 {
		b.WriteString("## Ripgrep matches\n")
		for _, m := range input.Matches {
			fmt.Fprintf(&b, "- %s:%d: %s\n", m.File, m.Line, m.Content)
		}
		b.WriteString("\n")
	}

	if input.Snippet != "" {
		b.WriteString("## Source snippet\n```\n")
		b.WriteString(input.Snippet)
		b.WriteString("\n```\n")
	}

	return b.String()
}
