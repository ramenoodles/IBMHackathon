package service

import (
	"context"
	"errors"
	"testing"
)

type fakeEnrichLLM struct {
	answer string
	err    error
	prompt string
}

func (f *fakeEnrichLLM) Ask(_ context.Context, _, _ string) (string, error) {
	return "", errors.New("not implemented")
}

func (f *fakeEnrichLLM) EnrichBatch(_ context.Context, prompt string) (string, error) {
	f.prompt = prompt
	return f.answer, f.err
}

func TestContextualHeuristicLockUnlock(t *testing.T) {
	title, summary, ok := contextualHeuristicLabel("close", "call", "m.mu.Lock()", "m.mu.Lock()")
	if !ok || title != "Lock m.mu" || summary == "" {
		t.Fatalf("Lock heuristic = %q, %q, %v", title, summary, ok)
	}

	title, summary, ok = contextualHeuristicLabel("close", "call", "m.mu.Unlock()", "m.mu.Unlock()")
	if !ok || title != "Unlock m.mu" {
		t.Fatalf("Unlock heuristic = %q, %q, %v", title, summary, ok)
	}
}

func TestContextualHeuristicBroadcastAndClosed(t *testing.T) {
	title, _, ok := contextualHeuristicLabel("close", "call", "m.cond.Broadcast()", "m.cond.Broadcast()")
	if !ok || title != "Wake all on m.cond" {
		t.Fatalf("Broadcast heuristic = %q, %v", title, ok)
	}

	title, _, ok = contextualHeuristicLabel("close", "branch", "if m.closed {", "if m.closed")
	if !ok || title != "m.closed already true?" {
		t.Fatalf("closed check heuristic = %q, %v", title, ok)
	}
}

func TestContextualHeuristicReturn(t *testing.T) {
	title, _, ok := contextualHeuristicLabel("close", "return", "return", "return")
	if !ok || title != "Return from close" {
		t.Fatalf("return heuristic = %q, %v", title, ok)
	}
}

func TestBuildEnrichPromptMidAudience(t *testing.T) {
	prompt := buildEnrichPrompt(EnrichRequest{
		Symbol: "close",
		UserContext: EnrichUserContext{
			ExperienceLevel: "mid",
			PrimaryLanguage: "go",
		},
	}, []EnrichNodeInput{{ID: "a", Code: "m.mu.Lock()", Kind: "call"}})
	if !contains(prompt, "mid-level developer") {
		t.Fatalf("prompt missing mid audience: %q", prompt)
	}
}

func TestEnrichUsesContextualHeuristicsWithoutLLM(t *testing.T) {
	service := newTestService(t, nil, false)
	result, err := service.Enrich(context.Background(), EnrichRequest{
		Symbol: "close",
		Nodes: []EnrichNodeInput{
			{ID: "a", Code: "m.mu.Lock()", Kind: "call", Label: "m.mu.Lock()"},
			{ID: "b", Code: "m.mu.Unlock()", Kind: "call", Label: "m.mu.Unlock()"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Patches) != 2 {
		t.Fatalf("patches = %#v", result.Patches)
	}
	if result.Patches[0].Title != "Lock m.mu" {
		t.Fatalf("first patch = %#v", result.Patches[0])
	}
	if result.Patches[0].LabelSource != "heuristic" {
		t.Fatalf("labelSource = %q, want heuristic", result.Patches[0].LabelSource)
	}
}

func TestEnrichSendsAllNodesToLLMWhenAvailable(t *testing.T) {
	client := &fakeEnrichLLM{
		answer: `{"patches":[
			{"id":"a","title":"Lock broker mutex (m.mu)","summary":"Serializes broker shutdown so only one goroutine mutates m.closed."},
			{"id":"b","title":"Broker already shut down","summary":"Skips duplicate close work when m.closed is already set."}
		]}`,
	}
	service := newTestService(t, client, false)
	result, err := service.Enrich(context.Background(), EnrichRequest{
		Symbol:   "close",
		FilePath: "broker.go",
		Nodes: []EnrichNodeInput{
			{ID: "a", Code: "m.mu.Lock()", Kind: "call", Label: "m.mu.Lock()"},
			{ID: "b", Code: "if m.closed {", Kind: "branch", Label: "if m.closed"},
		},
		UserContext: EnrichUserContext{ExperienceLevel: "junior", PrimaryLanguage: "go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Patches) != 2 {
		t.Fatalf("patches = %#v", result.Patches)
	}
	if result.Patches[0].Title != "Lock broker mutex (m.mu)" {
		t.Fatalf("first patch = %#v", result.Patches[0])
	}
	if result.Patches[0].LabelSource != "ai" || result.Patches[1].LabelSource != "ai" {
		t.Fatalf("labelSource = %#v", result.Patches)
	}
	if !contains(client.prompt, "Forbidden examples") || !contains(client.prompt, "m.mu.Lock()") {
		t.Fatalf("prompt missing specificity guidance: %q", client.prompt)
	}
}

func TestEnrichReplacesGenericLLMLabels(t *testing.T) {
	client := &fakeEnrichLLM{
		answer: `{"patches":[{"id":"a","title":"Acquire lock","summary":"Takes a mutex so only one goroutine can change shared state at a time."}]}`,
	}
	service := newTestService(t, client, false)
	result, err := service.Enrich(context.Background(), EnrichRequest{
		Symbol: "close",
		Nodes: []EnrichNodeInput{
			{ID: "a", Code: "m.mu.Lock()", Kind: "call", Label: "m.mu.Lock()"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Patches) != 1 || result.Patches[0].Title != "Lock m.mu" {
		t.Fatalf("expected contextual replacement, got %#v", result.Patches[0])
	}
	if result.Patches[0].LabelSource != "heuristic" {
		t.Fatalf("labelSource = %q, want heuristic", result.Patches[0].LabelSource)
	}
}

func TestEnrichCallsLLMForCalleeNodes(t *testing.T) {
	client := &fakeEnrichLLM{
		answer: `{"patches":[{"id":"x","title":"Shut down broker","summary":"Marks the broker as closed so new requests are rejected."}]}`,
	}
	service := newTestService(t, client, false)
	result, err := service.Enrich(context.Background(), EnrichRequest{
		Symbol:   "close",
		FilePath: "broker.go",
		Nodes: []EnrichNodeInput{
			{ID: "x", Code: "b.shutdown()", Kind: "call", Line: 10, Label: "shutdown()"},
		},
		UserContext: EnrichUserContext{ExperienceLevel: "junior", PrimaryLanguage: "go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Patches) != 1 || result.Patches[0].Title != "Shut down broker" {
		t.Fatalf("patches = %#v", result.Patches)
	}
}

func TestParseEnrichResponseStripsMarkdownFence(t *testing.T) {
	raw := "```json\n{\"patches\":[{\"id\":\"a\",\"title\":\"Test\",\"summary\":\"One sentence.\"}]}\n```"
	patches, err := parseEnrichResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(patches) != 1 || patches[0].ID != "a" {
		t.Fatalf("patches = %#v", patches)
	}
}

func TestParseEnrichResponseTrailingObject(t *testing.T) {
	// LLM occasionally returns two concatenated top-level objects; only the
	// first should be parsed and the rest silently ignored.
	raw := `{"patches":[{"id":"a","title":"Test","summary":"One sentence."}]},{"patches":[]}`
	patches, err := parseEnrichResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(patches) != 1 || patches[0].ID != "a" {
		t.Fatalf("patches = %#v", patches)
	}
}

func TestEnrichSaramaCloseLikeFlowWithoutLLM(t *testing.T) {
	service := newTestService(t, nil, false)
	nodes := []EnrichNodeInput{
		{ID: "1", Code: "m.mu.Lock()", Kind: "call", Label: "m.mu.Lock()"},
		{ID: "2", Code: "if m.closed {", Kind: "branch", Label: "if m.closed"},
		{ID: "3", Code: "return", Kind: "return", Label: "return"},
		{ID: "4", Code: "m.mu.Unlock()", Kind: "call", Label: "m.mu.Unlock()"},
		{ID: "5", Code: "m.cond.Broadcast()", Kind: "call", Label: "m.cond.Broadcast()"},
	}
	result, err := service.Enrich(context.Background(), EnrichRequest{
		Symbol:   "close",
		FilePath: "broker.go",
		Nodes:    nodes,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Patches) != 5 {
		t.Fatalf("expected 5 contextual patches, got %#v", result.Patches)
	}
	expected := map[string]string{
		"1": "Lock m.mu",
		"2": "m.closed already true?",
		"3": "Return from close",
		"4": "Unlock m.mu",
		"5": "Wake all on m.cond",
	}
	for _, patch := range result.Patches {
		if expected[patch.ID] != patch.Title {
			t.Fatalf("patch %s title = %q, want %q", patch.ID, patch.Title, expected[patch.ID])
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
