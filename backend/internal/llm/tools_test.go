package llm

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ramenoodles/IBMHackathon/backend/internal/search"
	"github.com/ramenoodles/IBMHackathon/backend/internal/source"

	wx "github.com/IBM/watsonx-go/pkg/models"
)

func TestRepositoryToolsRegisterDefinitions(t *testing.T) {
	tools := testTools(t)
	definitions := tools.definitions()
	if len(definitions) != 3 {
		t.Fatalf("definitions() returned %d tools, want 3", len(definitions))
	}
	for _, definition := range definitions {
		if definition.Function.Name == "" || definition.Function.Description == nil || *definition.Function.Description == "" {
			t.Errorf("incomplete definition: %#v", definition)
		}
	}
}

func TestRepositoryToolsReadFileAndContext(t *testing.T) {
	tools := testTools(t)
	file, err := tools.readFile(context.Background(), rawArgs(`{"path":"sample.go"}`))
	if err != nil || !strings.Contains(file, "func Parse") {
		t.Fatalf("readFile() = %q, %v", file, err)
	}
	contextResult, err := tools.readContext(context.Background(), rawArgs(`{"path":"sample.go","line":3,"before":1,"after":1}`))
	if err != nil || !strings.Contains(contextResult, "Lines: 2-4") {
		t.Fatalf("readContext() = %q, %v", contextResult, err)
	}
}

func TestRepositoryToolsRejectInvalidAndUnknownCalls(t *testing.T) {
	tools := testTools(t)
	if _, err := tools.execute(context.Background(), wx.ChatToolCall{}); err == nil {
		t.Fatal("execute() accepted unknown tool")
	}
	if _, err := tools.readFile(context.Background(), rawArgs(`{"path":1}`)); err == nil {
		t.Fatal("readFile() accepted invalid path type")
	}
	if _, err := tools.execute(context.Background(), wx.ChatToolCall{Function: wx.ChatToolCallFunction{Name: "read_file", Arguments: "{"}}); err == nil {
		t.Fatal("execute() accepted invalid JSON arguments")
	}
}

func TestToolOutputIsLimited(t *testing.T) {
	value := limitOutput(strings.Repeat("x", maxToolOutput+10))
	if len(value) != maxToolOutput+len("\n[output truncated]") || !strings.HasSuffix(value, "[output truncated]") {
		t.Fatalf("limitOutput() length/suffix = %d/%q", len(value), value[maxToolOutput:])
	}
	if limitOutput("short") != "short" {
		t.Fatal("limitOutput() changed short output")
	}
}

func TestSearchSymbolReturnsJSONMatches(t *testing.T) {
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("ripgrep is not installed")
	}
	tools := testTools(t)
	result, err := tools.searchSymbol(context.Background(), rawArgs(`{"name":"Parse","language":"go","limit":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "sample.go") || !json.Valid([]byte(result)) {
		t.Fatalf("searchSymbol() = %q", result)
	}
}

func testTools(t *testing.T) *repositoryTools {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "sample.go"), []byte("package sample\n\nfunc Parse() string {\n\treturn \"ok\"\n}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	reader, err := source.NewReader(root)
	if err != nil {
		t.Fatal(err)
	}
	return newRepositoryTools(reader, search.NewFinder("rg"), root)
}

func rawArgs(value string) map[string]json.RawMessage {
	var args map[string]json.RawMessage
	if err := json.Unmarshal([]byte(value), &args); err != nil {
		panic(err)
	}
	return args
}
