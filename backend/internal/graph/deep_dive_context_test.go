package graph

import "testing"

func TestDeepDiveBundleFormatNoNodeIDs(t *testing.T) {
	bundle := DeepDiveBundle{
		Symbol:     "provider_name",
		StepCode:   `return "adzuna"`,
		StepKind:   "return",
		StepLine:   41,
		SymbolBody: "L41: return \"adzuna\"\n",
		FlowNeighbors: []string{
			"Comes after: Function entry — execution starts here",
		},
		DomainTerms: []string{"adzuna"},
		Evidence:    []string{"providers/adzuna.py:41: return \"adzuna\""},
	}
	formatted := bundle.Format()
	if formatted == "" {
		t.Fatal("expected formatted payload")
	}
	if containsAny(formatted, "L40", "node L", "from L", "Write your answer", "cite only") {
		t.Fatalf("formatted payload should be evidence-only: %s", formatted)
	}
	if !containsAny(formatted, "Evidence from codebase:") {
		t.Fatalf("expected evidence section: %s", formatted)
	}
}

func TestDeepDiveBundleFormatOmitsInferredHintWhenNoEvidence(t *testing.T) {
	bundle := DeepDiveBundle{Symbol: "normalize_url_cache_key", StepCode: `parsed = urlparse(url)`}
	formatted := bundle.Format()
	if !containsAny(formatted, "omit [INFERRED]") {
		t.Fatalf("expected inferred omission hint: %s", formatted)
	}
}

func TestValidateDeepDiveTextRejectsStep40(t *testing.T) {
	if ok, _ := ValidateDeepDiveText("incoming data from step 40 is ignored"); ok {
		t.Fatal("expected rejection")
	}
}

func TestValidateDeepDiveTextRejectsPromptEcho(t *testing.T) {
	echo := "[VERIFIED] What this step does — cite only the step code line above."
	if ok, _ := ValidateDeepDiveText(echo); ok {
		t.Fatal("expected prompt echo rejection")
	}
}

func TestSanitizeDeepDiveTextStripsBoilerplate(t *testing.T) {
	raw := "[VERIFIED] What this step does — Parses the URL with urlparse.\n[VERIFIED] Role in this function — Comes after function entry."
	clean := SanitizeDeepDiveText(raw)
	if containsAny(clean, "cite only", "Role in this function") {
		t.Fatalf("expected boilerplate stripped: %q", clean)
	}
	if !containsAny(clean, "Parses the URL") {
		t.Fatalf("expected content preserved: %q", clean)
	}
}

func TestSplitVerifiedInferred(t *testing.T) {
	text := "[VERIFIED] Runs return.\n[VERIFIED] Ends the function.\n[INFERRED] Based on codebase evidence: adzuna provider id."
	v, i := SplitVerifiedInferred(text)
	if v == "" || i == "" {
		t.Fatalf("got verified=%q inferred=%q", v, i)
	}
}

func containsAny(s string, parts ...string) bool {
	for _, p := range parts {
		if len(p) > 0 && stringContains(s, p) {
			return true
		}
	}
	return false
}

func stringContains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
