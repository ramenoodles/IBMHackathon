package service

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	methodSyncPattern  = regexp.MustCompile(`([\w]+)\.([\w]+)\.(Lock|Unlock|Broadcast|Signal|Wait)\s*\(\s*\)`)
	fieldCheckPattern  = regexp.MustCompile(`(?i)if\s+([\w]+)\.(\w+)`)
	simpleCallPattern  = regexp.MustCompile(`([\w]+)\s*\(\s*\)`)
)

// contextualHeuristicLabel is a fallback when Watsonx is unavailable.
// It extracts object/field names from the code so labels stay specific.
func contextualHeuristicLabel(symbol, kind, code, label string) (title, summary string, ok bool) {
	clean := strings.TrimSpace(code)
	parent := strings.TrimSpace(symbol)
	if parent == "" {
		parent = "function"
	}

	if m := methodSyncPattern.FindStringSubmatch(clean); len(m) == 4 {
		recv, field, method := m[1], m[2], m[3]
		switch method {
		case "Lock":
			return fmt.Sprintf("Lock %s.%s", recv, field),
				fmt.Sprintf("Takes %s.%s so only one goroutine runs %s at a time.", recv, field, parent),
				true
		case "Unlock":
			return fmt.Sprintf("Unlock %s.%s", recv, field),
				fmt.Sprintf("Releases %s.%s so other goroutines can enter %s.", recv, field, parent),
				true
		case "Broadcast":
			return fmt.Sprintf("Wake all on %s.%s", recv, field),
				fmt.Sprintf("Broadcasts on %s.%s to unblock every waiter during %s.", recv, field, parent),
				true
		case "Signal":
			return fmt.Sprintf("Wake one on %s.%s", recv, field),
				fmt.Sprintf("Signals %s.%s to unblock one waiter during %s.", recv, field, parent),
				true
		case "Wait":
			return fmt.Sprintf("Wait on %s.%s", recv, field),
				fmt.Sprintf("Blocks on %s.%s until another goroutine signals during %s.", recv, field, parent),
				true
		}
	}

	if m := fieldCheckPattern.FindStringSubmatch(clean); len(m) == 3 {
		recv, field := m[1], m[2]
		if strings.EqualFold(field, "closed") {
			return fmt.Sprintf("%s.%s already true?", recv, field),
				fmt.Sprintf("Skips %s when %s.%s is already set.", parent, recv, field),
				true
		}
		return fmt.Sprintf("Check %s.%s", recv, field),
			fmt.Sprintf("Branches %s based on the current value of %s.%s.", parent, recv, field),
			true
	}

	switch kind {
	case "return", "raise":
		if kind == "raise" {
			return fmt.Sprintf("Raise from %s", parent),
				fmt.Sprintf("Stops %s and returns an error to the caller.", parent),
				true
		}
		return fmt.Sprintf("Return from %s", parent),
			fmt.Sprintf("Ends %s here — remaining steps on this path are skipped.", parent),
			true
	case "entry":
		if clean != "" {
			return fmt.Sprintf("Enter %s", parent),
				fmt.Sprintf("Execution of %s begins at this step.", parent),
				true
		}
		return fmt.Sprintf("Enter %s", parent),
			fmt.Sprintf("Execution of %s begins at this step.", parent),
			true
	case "call":
		callLabel := strings.TrimSpace(label)
		if callLabel == "" && simpleCallPattern.MatchString(clean) {
			if m := simpleCallPattern.FindStringSubmatch(clean); len(m) == 2 {
				callLabel = m[1] + "()"
			}
		}
		if callLabel != "" {
			name := strings.TrimSuffix(callLabel, "()")
			return fmt.Sprintf("Call %s", name),
				fmt.Sprintf("Invokes %s as part of %s.", name, parent),
				true
		}
	}

	return "", "", false
}
