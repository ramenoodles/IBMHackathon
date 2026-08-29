package graph

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ParseFlowGraph parses and validates LLM JSON output into a FlowGraph.
func ParseFlowGraph(raw string, symbol, filePath string) (FlowGraph, error) {
	cleaned := extractJSON(raw)
	var g FlowGraph
	if err := json.Unmarshal([]byte(cleaned), &g); err != nil {
		return FlowGraph{}, err
	}
	if g.Symbol == "" {
		g.Symbol = symbol
	}
	enforceLimits(&g, MaxRootNodes)
	if g.RootID == "" && len(g.Nodes) > 0 {
		g.RootID = g.Nodes[0].ID
	}
	for i := range g.Nodes {
		if g.Nodes[i].File == "" {
			g.Nodes[i].File = filePath
		}
	}
	return g, nil
}

// ParseNodeDetail parses LLM JSON into NodeDetail.
func ParseNodeDetail(raw string) (NodeDetail, error) {
	cleaned := extractJSON(raw)
	var d NodeDetail
	if err := json.Unmarshal([]byte(cleaned), &d); err != nil {
		return NodeDetail{}, err
	}
	return d, nil
}

// ValidateNodeDetail rejects obviously bad LLM output.
func ValidateNodeDetail(d NodeDetail) bool {
	if strings.TrimSpace(d.Explanation) == "" {
		return false
	}
	if looksLikeCodeLabel(d.Title) || looksLikeCodeLabel(d.Summary) {
		return false
	}
	return true
}

func looksLikeCodeLabel(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if strings.ContainsAny(s, "{};") {
		return true
	}
	return strings.HasPrefix(s, "return ")
}

var (
	stepNumberRe    = regexp.MustCompile(`(?i)\bstep\s+\d+`)
	nodeIDRe        = regexp.MustCompile(`(?i)\bnode\s+l\d+`)
	incomingStepRe  = regexp.MustCompile(`(?i)incoming data from step`)
	promptEchoRe    = regexp.MustCompile(`(?i)(cite only|use flow neighbors|step code line above|what this step does\s*[—-]|role in this function\s*[—-]|broader context\s*[—-])`)
	verifiedBlockRe = regexp.MustCompile(`(?is)\[VERIFIED\]\s*([^[]*)`)
	inferredBlockRe = regexp.MustCompile(`(?is)\[INFERRED\]\s*([^[]*)`)
)

// ValidateDeepDiveText rejects common hallucination and prompt-echo patterns.
func ValidateDeepDiveText(text string) (bool, string) {
	if strings.TrimSpace(text) == "" {
		return false, "empty explanation"
	}
	lower := strings.ToLower(text)
	if stepNumberRe.MatchString(lower) {
		return false, "references step numbers"
	}
	if nodeIDRe.MatchString(lower) {
		return false, "references internal node ids"
	}
	if incomingStepRe.MatchString(lower) {
		return false, "phantom incoming step reference"
	}
	if promptEchoRe.MatchString(lower) {
		return false, "echoes prompt instructions"
	}
	return true, ""
}

// SanitizeDeepDiveText strips echoed prompt boilerplate from model output.
func SanitizeDeepDiveText(text string) string {
	verified, inferred := SplitVerifiedInferred(text)
	if verified == "" && inferred == "" {
		return stripSectionBoilerplate(text)
	}
	parts := strings.Split(verified, "\n\n")
	for i, p := range parts {
		parts[i] = stripSectionBoilerplate(p)
	}
	verified = strings.TrimSpace(strings.Join(filterNonemptyStrings(parts), "\n\n"))
	inferred = stripSectionBoilerplate(inferred)
	if verified == "" && inferred == "" {
		return strings.TrimSpace(text)
	}
	var blocks []string
	if verified != "" {
		blocks = append(blocks, "[VERIFIED] "+verified)
	}
	if inferred != "" {
		blocks = append(blocks, "[INFERRED] "+inferred)
	}
	return strings.Join(blocks, "\n")
}

func stripSectionBoilerplate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	lower := strings.ToLower(s)
	prefixes := []string{
		"what this step does — ",
		"what this step does - ",
		"role in this function — ",
		"role in this function - ",
		"broader context — ",
		"broader context - ",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			s = strings.TrimSpace(s[len(p):])
			lower = strings.ToLower(s)
		}
	}
	for _, phrase := range []string{
		"cite only the step code line above",
		"cite only the provided step code",
		"use flow neighbors only",
		"only if evidence or domain terms support it",
	} {
		s = replaceFold(s, phrase, "")
	}
	return strings.TrimSpace(s)
}

func replaceFold(s, old, new string) string {
	lower := strings.ToLower(s)
	oldLower := strings.ToLower(old)
	var out strings.Builder
	for i := 0; i < len(s); {
		if strings.HasPrefix(lower[i:], oldLower) {
			out.WriteString(new)
			i += len(old)
			continue
		}
		out.WriteByte(s[i])
		i++
	}
	return out.String()
}

func filterNonemptyStrings(parts []string) []string {
	var out []string
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// SplitVerifiedInferred parses [VERIFIED] and [INFERRED] sections from LLM output.
func SplitVerifiedInferred(text string) (verified, inferred string) {
	var verifiedParts []string
	for _, m := range verifiedBlockRe.FindAllStringSubmatch(text, -1) {
		if s := strings.TrimSpace(m[1]); s != "" {
			verifiedParts = append(verifiedParts, s)
		}
	}
	if m := inferredBlockRe.FindStringSubmatch(text); len(m) > 1 {
		inferred = strings.TrimSpace(m[1])
	}
	return strings.TrimSpace(strings.Join(verifiedParts, "\n\n")), inferred
}

func extractJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if idx := strings.Index(raw, "{"); idx >= 0 {
		raw = raw[idx:]
	}
	if idx := strings.LastIndex(raw, "}"); idx >= 0 {
		raw = raw[:idx+1]
	}
	return raw
}

func enforceLimits(g *FlowGraph, maxNodes int) {
	if len(g.Nodes) <= maxNodes {
		return
	}
	hidden := len(g.Nodes) - maxNodes + 1
	last := g.Nodes[maxNodes-1]
	last.Collapsed = true
	last.ChildCount = hidden
	last.Expandable = true
	last.Kind = "branch"
	last.Label = fmt.Sprintf("+%d branches", hidden)
	g.Nodes = g.Nodes[:maxNodes]
	allowed := map[string]bool{}
	for _, n := range g.Nodes {
		allowed[n.ID] = true
	}
	var edges []FlowEdge
	for _, e := range g.Edges {
		if allowed[e.From] && allowed[e.To] {
			edges = append(edges, e)
		}
	}
	g.Edges = edges
}
