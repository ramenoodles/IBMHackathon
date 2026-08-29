package flow

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	ifCondPattern  = regexp.MustCompile(`^\s*if\s+(.+?)\s*:\s*$`)
	elifCondPattern = regexp.MustCompile(`^\s*elif\s+(.+?)\s*:\s*$`)
	forLoopPattern = regexp.MustCompile(`^\s*for\s+(.+?)\s*:\s*$`)
	whilePattern   = regexp.MustCompile(`^\s*while\s+(.+?)\s*:\s*$`)
	returnPattern  = regexp.MustCompile(`^\s*return\s*(.*)$`)
)

// EnrichSteps adds human-readable labels and summaries for graph display.
func EnrichSteps(steps []Step) []Step {
	out := make([]Step, len(steps))
	for i, s := range steps {
		out[i] = s
		label, summary := humanize(s)
		if label != "" {
			out[i].Label = label
		}
		if summary != "" {
			out[i].Summary = summary
		}
	}
	return out
}

func humanize(s Step) (label, summary string) {
	code := strings.TrimSpace(s.Code)
	switch s.Kind {
	case "entry":
		return s.Label, "Function entry — execution starts here"
	case "branch":
		switch s.BranchKind {
		case "if", "elif":
			if m := ifCondPattern.FindStringSubmatch(code); m != nil {
				cond := shorten(m[1], 40)
				return fmt.Sprintf("if %s", cond), fmt.Sprintf("Branch when: %s", cond)
			}
			if m := elifCondPattern.FindStringSubmatch(code); m != nil {
				cond := shorten(m[1], 40)
				return fmt.Sprintf("elif %s", cond), fmt.Sprintf("Else-if: %s", cond)
			}
		case "else":
			return "else", "Fallback branch when prior conditions fail"
		}
		return shorten(code, 44), "Conditional branch"
	case "loop":
		if m := forLoopPattern.FindStringSubmatch(code); m != nil {
			iter := shorten(m[1], 36)
			return fmt.Sprintf("for %s", iter), fmt.Sprintf("Loop over: %s", iter)
		}
		if m := whilePattern.FindStringSubmatch(code); m != nil {
			cond := shorten(m[1], 36)
			return fmt.Sprintf("while %s", cond), fmt.Sprintf("Loop while: %s", cond)
		}
		return shorten(code, 44), "Loop"
	case "return", "raise":
		if m := returnPattern.FindStringSubmatch(code); m != nil {
			val := strings.TrimSpace(m[1])
			if val == "" || strings.HasPrefix(val, "{") {
				return "Return early", "Exit the function with a default or empty result"
			}
			return "return " + shorten(val, 24), "Exit with: " + shorten(val, 50)
		}
		return "Return", "Exit the function"
	case "call":
		name := strings.TrimSuffix(s.Label, "()")
		if strings.HasPrefix(name, "_") {
			name = name[1:]
		}
		pretty := strings.ReplaceAll(name, "_", " ")
		return s.Label, fmt.Sprintf("Calls %s", pretty)
	case "assign":
		return s.Label, "Assign: " + shorten(code, 60)
	default:
		return shorten(code, 44), shorten(code, 72)
	}
}

func shorten(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
