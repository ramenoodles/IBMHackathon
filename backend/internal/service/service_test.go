package service

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	llmclient "github.com/ramenoodles/IBMHackathon/backend/internal/llm"
)

type fakeLLM struct {
	answer   string
	err      error
	question string
	source   string
}

func (fake *fakeLLM) Ask(_ context.Context, question, source string) (string, error) {
	fake.question, fake.source = question, source
	return fake.answer, fake.err
}

type fakeAgent struct {
	fakeLLM
	trajectory []llmclient.TrajectoryEvent
}

func (fake *fakeAgent) AskAgent(_ context.Context, question, source string) (llmclient.AgentResponse, error) {
	fake.question, fake.source = question, source
	return llmclient.AgentResponse{Answer: fake.answer, Trajectory: fake.trajectory}, fake.err
}

func TestExplainRequiresLLM(t *testing.T) {
	service := newTestService(t, nil, false)
	_, err := service.Explain(context.Background(), ExplainRequest{Name: "Parse"})
	if !errors.Is(err, ErrLLMUnavailable) {
		t.Fatalf("Explain() error = %v, want ErrLLMUnavailable", err)
	}
}

func TestExplainUsesPlainClient(t *testing.T) {
	requireRipgrep(t)
	root := testRepository(t)
	client := &fakeLLM{answer: "plain answer"}
	service := newTestService(t, client, false, root)

	result, err := service.Explain(context.Background(), ExplainRequest{Name: "Parse", Question: "How?"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Answer != "plain answer" || client.question != "How?" || client.source == "" {
		t.Fatalf("Explain() result/client = %#v, %#v", result, client)
	}
	if result.Trajectory != nil {
		t.Fatalf("trajectory = %#v, want nil", result.Trajectory)
	}
}

func TestExplainUsesAgentWhenTrajectoryIncluded(t *testing.T) {
	requireRipgrep(t)
	root := testRepository(t)
	client := &fakeAgent{
		fakeLLM:    fakeLLM{answer: "agent answer"},
		trajectory: []llmclient.TrajectoryEvent{{Type: "message", Role: "assistant", Content: "trace"}},
	}
	service := newTestService(t, client, true, root)

	result, err := service.Explain(context.Background(), ExplainRequest{Name: "Parse"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Answer != "agent answer" || len(result.Trajectory) != 1 {
		t.Fatalf("Explain() = %#v", result)
	}
}

func TestExplainReportsMissingSymbolAndClientErrors(t *testing.T) {
	requireRipgrep(t)
	root := testRepository(t)
	service := newTestService(t, &fakeLLM{answer: "unused"}, false, root)
	_, err := service.Explain(context.Background(), ExplainRequest{Name: "Missing"})
	if !errors.Is(err, ErrSymbolNotFound) {
		t.Fatalf("missing symbol error = %v", err)
	}

	client := &fakeLLM{err: errors.New("backend failed")}
	service = newTestService(t, client, false, root)
	_, err = service.Explain(context.Background(), ExplainRequest{Name: "Parse"})
	if err == nil || !errors.Is(err, client.err) {
		t.Fatalf("client error = %v", err)
	}
}

func newTestService(t *testing.T, client llmclient.Client, includeTrajectory bool, roots ...string) *Service {
	t.Helper()
	root := t.TempDir()
	if len(roots) > 0 {
		root = roots[0]
	}
	value, err := New(root, "rg", client, includeTrajectory)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func testRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "sample.go")
	if err := os.WriteFile(path, []byte("package sample\n\nfunc Parse() string {\n\treturn \"ok\"\n}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func requireRipgrep(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep is not installed")
	}
}
