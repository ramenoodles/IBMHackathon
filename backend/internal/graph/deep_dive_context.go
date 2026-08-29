package graph

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ibmhackathon/onbober/internal/scanner"
	"github.com/ibmhackathon/onbober/internal/scanner/flow"
)

// DeepDiveBundle is scan-first evidence for grounded node explanations.
type DeepDiveBundle struct {
	Symbol        string
	StepCode      string
	StepLine      int
	StepKind      string
	EnrichSummary string
	SymbolBody    string
	EnclosingName string
	EnclosingKind string
	Imports       string
	FlowNeighbors []string
	DomainTerms   []string
	Evidence      []string
}

func (b *Builder) buildDeepDiveBundle(input BuildInput, g FlowGraph, nodeID, code, kind, enrichSummary string) DeepDiveBundle {
	bundle := DeepDiveBundle{
		Symbol:        input.Symbol,
		EnrichSummary: enrichSummary,
		StepKind:      kind,
		StepCode:      code,
	}

	node, ok := FindNodeByID(g, nodeID)
	if ok {
		if bundle.StepCode == "" {
			bundle.StepCode = node.Code
		}
		if bundle.StepKind == "" {
			bundle.StepKind = node.Kind
		}
		bundle.StepLine = node.Line
		bundle.FlowNeighbors = humanFlowNeighbors(g, nodeID)
	}

	if input.FilePath != "" {
		content, lang, err := b.scanner.ReadFile(input.WorkspacePath, input.FilePath)
		if err == nil {
			bundle.Imports = extractImportsBlock(content, 30)
			if bundle.StepLine > 0 {
				bundle.EnclosingName, bundle.EnclosingKind = scanner.FindEnclosingSymbol(content, bundle.StepLine)
			}
			steps := flow.ExtractFlow(content, input.FilePath, input.Symbol, lang)
			bundle.SymbolBody = joinFlowStepsNearLine(steps, bundle.StepLine, 8)
			if bundle.StepCode == "" && bundle.StepLine > 0 {
				for _, s := range steps {
					if s.Line == bundle.StepLine {
						bundle.StepCode = s.Code
						break
					}
				}
			}
		}
	}

	baseName := strings.TrimSuffix(filepath.Base(input.FilePath), filepath.Ext(input.FilePath))
	terms := scanner.DistinctiveDomainTerms(
		baseName,
		bundle.EnclosingName,
		strings.Join(scanner.ExtractStringLiterals(bundle.StepCode), " "),
		strings.Join(scanner.ExtractStringLiterals(bundle.SymbolBody), " "),
	)
	bundle.DomainTerms = terms

	seenEvidence := map[string]bool{}
	for _, term := range terms {
		matches, _ := b.scanner.GrepLiteral(input.WorkspacePath, term, 4)
		for _, m := range matches {
			line := fmt.Sprintf("%s:%d: %s", m.File, m.Line, m.Content)
			if ctx, err := b.scanner.ReadMatchContext(input.WorkspacePath, m.File, m.Line, 1, 1); err == nil && strings.TrimSpace(ctx) != "" {
				flat := strings.ReplaceAll(strings.TrimSpace(ctx), "\n", " | ")
				if len(flat) > 120 {
					flat = flat[:117] + "..."
				}
				line = fmt.Sprintf("%s:%d: %s", m.File, m.Line, flat)
			}
			if seenEvidence[line] {
				continue
			}
			seenEvidence[line] = true
			bundle.Evidence = append(bundle.Evidence, line)
			if len(bundle.Evidence) >= 8 {
				break
			}
		}
		if len(bundle.Evidence) >= 8 {
			break
		}
	}

	return bundle
}

func humanFlowNeighbors(g FlowGraph, nodeID string) []string {
	var lines []string
	for _, e := range g.Edges {
		if e.To != nodeID {
			continue
		}
		parent, ok := FindNodeByID(g, e.From)
		if !ok {
			continue
		}
		lines = append(lines, fmt.Sprintf("Comes after: %s", humanStepLabel(parent)))
	}
	for _, e := range g.Edges {
		if e.From != nodeID {
			continue
		}
		child, ok := FindNodeByID(g, e.To)
		if !ok {
			continue
		}
		lines = append(lines, fmt.Sprintf("Leads to: %s", humanStepLabel(child)))
	}
	return lines
}

func humanStepLabel(n FlowNode) string {
	if s := strings.TrimSpace(n.Summary); s != "" {
		return s
	}
	if c := strings.TrimSpace(n.Code); c != "" {
		return truncateStr(c, 80)
	}
	label := strings.TrimSpace(n.Label)
	if label != "" {
		if idx := strings.Index(label, " "); idx > 0 {
			label = label[idx+1:]
		}
		return truncateStr(label, 80)
	}
	return n.Kind
}

func joinFlowStepsNearLine(steps []flow.Step, line, radius int) string {
	if line < 1 || len(steps) == 0 {
		return joinFlowSteps(steps)
	}
	var b strings.Builder
	for _, s := range steps {
		if s.Code == "" {
			continue
		}
		if s.Line >= line-radius && s.Line <= line+radius {
			fmt.Fprintf(&b, "L%d: %s\n", s.Line, s.Code)
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return joinFlowSteps(steps)
	}
	return truncateStr(out, 1200)
}

func joinFlowSteps(steps []flow.Step) string {
	var b strings.Builder
	for _, s := range steps {
		if s.Code != "" {
			fmt.Fprintf(&b, "L%d: %s\n", s.Line, s.Code)
		}
	}
	return truncateStr(b.String(), 2500)
}

func extractImportsBlock(content string, maxLines int) string {
	lines := strings.Split(content, "\n")
	var out []string
	for i, line := range lines {
		if i >= maxLines {
			break
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "from ") ||
			strings.HasPrefix(trimmed, "use ") || strings.HasPrefix(trimmed, "#include") {
			out = append(out, trimmed)
		}
	}
	return strings.Join(out, "\n")
}

func (b DeepDiveBundle) Format() string {
	var out strings.Builder
	fmt.Fprintf(&out, "Function: %s\n", b.Symbol)
	if b.EnclosingName != "" {
		fmt.Fprintf(&out, "Enclosing %s: %s\n", b.EnclosingKind, b.EnclosingName)
	}
	if b.StepKind != "" {
		fmt.Fprintf(&out, "Step kind: %s\n", b.StepKind)
	}
	if b.EnrichSummary != "" {
		fmt.Fprintf(&out, "Existing summary (do not contradict): %s\n", b.EnrichSummary)
	}
	if b.StepCode != "" {
		fmt.Fprintf(&out, "Step code (explain ONLY this line):\n```\n%s\n```\n", b.StepCode)
	}
	if b.SymbolBody != "" && b.SymbolBody != b.StepCode {
		fmt.Fprintf(&out, "Nearby symbol context (for orientation — do not describe other steps as if they are the current step):\n```\n%s\n```\n", b.SymbolBody)
	}
	if b.Imports != "" {
		fmt.Fprintf(&out, "Imports:\n%s\n", b.Imports)
	}
	if len(b.FlowNeighbors) > 0 {
		out.WriteString("Flow neighbors:\n")
		for _, line := range b.FlowNeighbors {
			fmt.Fprintf(&out, "- %s\n", line)
		}
	}
	if len(b.DomainTerms) > 0 {
		fmt.Fprintf(&out, "Domain terms: %s\n", strings.Join(b.DomainTerms, ", "))
	}
	if len(b.Evidence) > 0 {
		out.WriteString("Evidence from codebase:\n")
		for _, e := range b.Evidence {
			fmt.Fprintf(&out, "- %s\n", e)
		}
	} else {
		out.WriteString("Evidence from codebase: none — omit [INFERRED] section.\n")
	}
	return out.String()
}

func fallbackDeepDiveExplanation(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return "[VERIFIED] This step is part of the scanned execution flow."
	}
	return fmt.Sprintf("[VERIFIED] This step runs: %s", truncateStr(code, 240))
}

func applyDeepDiveSections(d *NodeDetail, text string, bundle DeepDiveBundle) {
	text = SanitizeDeepDiveText(text)
	verified, inferred := SplitVerifiedInferred(text)
	if verified == "" && inferred == "" {
		verified = strings.TrimSpace(text)
	}
	if len(bundle.Evidence) == 0 {
		inferred = ""
	}
	d.VerifiedExplanation = verified
	d.InferredExplanation = inferred
	d.Explanation = strings.TrimSpace(strings.Join(filterNonEmpty([]string{verified, inferred}), "\n\n"))
	d.Evidence = bundle.Evidence
}

func filterNonEmpty(parts []string) []string {
	var out []string
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, p)
		}
	}
	return out
}
