package llm

import (
	"context"
	"fmt"
)

const enrichSystemPrompt = `You are OnBober, a codebase onboarding assistant. Return ONLY JSON:
{"patches":[{"id":"node_id","title":"3-6 word graph label","summary":"one plain-English sentence explaining why this step matters in the flow","relatedSymbols":[]}]}
Rules:
- title: short, no code, no quotes, no braces — meant for a flowchart box (e.g. "Normalize date string", "Guard empty input").
- summary: explain intent and context for a new developer onboarding to this function.
- Do NOT paste raw source code or JSON literals in title or summary.
- Do NOT invent steps or change structure.
- Keep title under 40 chars, summary under 140 chars.`

// GenerateEnrichSummaries requests batch display patches for flow nodes.
func (c *OllamaClient) GenerateEnrichSummaries(ctx context.Context, userCtx UserContext, symbol, nodesPayload string) (string, error) {
	system := enrichSystemPrompt
	switch userCtx.ExperienceLevel {
	case "senior":
		system += "\nAudience: senior engineer — be terse and precise."
	case "junior":
		system += "\nAudience: junior developer — be plain and welcoming."
	}
	user := fmt.Sprintf("Function: %s\nSteps to describe:\n%s\nReturn JSON patches only.", symbol, nodesPayload)
	return c.ChatComplete(ctx, system, user)
}
