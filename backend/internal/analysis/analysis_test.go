package analysis

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootUsesRequestedFileAndFunctionBody(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "a.py", `def selected():
    wrong = 1
    return wrong
`)
	writeTestFile(t, root, "b.py", `def selected():
    right = 2
    return right

def unrelated():
    should_not_appear = 3
`)

	builder := newTestBuilder(t, root)
	graph, err := builder.Root(context.Background(), "b.py", "selected")
	if err != nil {
		t.Fatal(err)
	}
	if graph.Depth != 1 || graph.Symbol != "selected" || len(graph.Nodes) != 3 {
		t.Fatalf("Root() = depth %d, symbol %q, %d nodes", graph.Depth, graph.Symbol, len(graph.Nodes))
	}
	for _, node := range graph.Nodes {
		if node.File != "b.py" {
			t.Fatalf("Root() selected node from %q", node.File)
		}
		if strings.Contains(node.Code, "should_not_appear") || strings.Contains(node.Code, "wrong") {
			t.Fatalf("Root() included code outside selected function: %q", node.Code)
		}
	}
}

func TestRootIsBoundedAndMarksOmittedSteps(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "busy.py", `def busy():
    one = 1
    two = 2
    three = 3
    four = 4
    five = 5
    six = 6
    seven = 7
    eight = 8
    nine = 9
    ten = 10
    return ten
`)

	graph, err := newTestBuilder(t, root).Root(context.Background(), "busy.py", "busy")
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Nodes) != MaxRootNodes {
		t.Fatalf("Root() returned %d nodes, want %d", len(graph.Nodes), MaxRootNodes)
	}
	last := graph.Nodes[len(graph.Nodes)-1]
	if last.Kind != "branch" || !strings.Contains(last.Label, "more steps") {
		t.Fatalf("last node = %#v, want truncation marker", last)
	}
	assertInternalEdges(t, graph)
}

func TestExpandReturnsMergeableBoundedCalleeGraph(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "flow.py", `def child():
    one = 1
    two = 2
    three = 3
    four = 4
    return four

def parent():
    value = child()
    return value
`)

	builder := newTestBuilder(t, root)
	graph, err := builder.Root(context.Background(), "flow.py", "parent")
	if err != nil {
		t.Fatal(err)
	}
	var call Node
	for _, node := range graph.Nodes {
		if node.Kind == "call" {
			call = node
			break
		}
	}
	if call.ID == "" || !call.Collapsed || !call.Expandable || call.ChildCount != 5 {
		t.Fatalf("resolved call = %#v", call)
	}

	fragment, err := builder.Expand(context.Background(), call.ID, 3)
	if err != nil {
		t.Fatal(err)
	}
	if fragment.RootID != call.ID {
		t.Fatalf("Expand().RootID = %q, want %q", fragment.RootID, call.ID)
	}
	if fragment.Depth != 2 || fragment.Symbol != "child" || len(fragment.Nodes) != 3 {
		t.Fatalf("Expand() = depth %d, symbol %q, %d nodes", fragment.Depth, fragment.Symbol, len(fragment.Nodes))
	}
	if len(fragment.Edges) == 0 || fragment.Edges[0].From != call.ID || fragment.Edges[0].To != fragment.Nodes[0].ID {
		t.Fatalf("Expand() missing parent-to-callee edge: %#v", fragment.Edges)
	}
	for _, node := range fragment.Nodes {
		if node.ID == call.ID {
			t.Fatal("Expand() duplicated the existing parent node")
		}
	}
}

func TestRootBuildsConditionalEdges(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "guard.py", `def guard(value):
    if value < 0:
        return 0
    result = value
    return result
`)

	graph, err := newTestBuilder(t, root).Root(context.Background(), "guard.py", "guard")
	if err != nil {
		t.Fatal(err)
	}
	labels := make(map[string]bool)
	for _, edge := range graph.Edges {
		labels[edge.Label] = true
	}
	if !labels["start"] || !labels["true"] || !labels["false"] {
		t.Fatalf("Root() edge labels = %#v", labels)
	}
	assertInternalEdges(t, graph)
}

func TestNestedExpansionCanResumeWithNewBuilder(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "nested.py", `def leaf():
    result = 1
    return result

def child():
    value = leaf()
    return value

def parent():
    value = child()
    return value
`)

	builder := newTestBuilder(t, root)
	graph, err := builder.Root(context.Background(), "nested.py", "parent")
	if err != nil {
		t.Fatal(err)
	}
	childCall := firstCallNode(t, graph.Nodes)
	childFragment, err := builder.Expand(context.Background(), childCall.ID, MaxExpandNodes)
	if err != nil {
		t.Fatal(err)
	}
	leafCall := firstCallNode(t, childFragment.Nodes)

	newBuilder := newTestBuilder(t, root)
	leafFragment, err := newBuilder.Expand(context.Background(), leafCall.ID, MaxExpandNodes)
	if err != nil {
		t.Fatal(err)
	}
	if leafFragment.Depth != 3 || leafFragment.RootID != leafCall.ID || leafFragment.Symbol != "leaf" {
		t.Fatalf("nested Expand() = %#v", leafFragment)
	}
}

func TestGoFlowUsesMatchingBraceBoundary(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "flow.go", `package sample

func selected(value int) int {
	labels := map[string]string{"close": "}"}
	if value < 0 {
		return 0
	}
	return value
}

func unrelated() int {
	return 99
}
`)

	graph, err := newTestBuilder(t, root).Root(context.Background(), "flow.go", "selected")
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range graph.Nodes {
		if strings.Contains(node.Code, "99") || strings.Contains(node.Code, "unrelated") {
			t.Fatalf("Go flow crossed function boundary: %#v", node)
		}
	}
	assertInternalEdges(t, graph)
}

func TestExtractSymbolsSupportsModernJavaScriptAndDeduplicates(t *testing.T) {
	content := `export async function load() {}
const save = async (value) => { return value }
class First {
  run() {}
}
class Second {
  run() {}
}
`
	symbols := ExtractSymbols(content, "service.ts")
	got := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		got = append(got, symbol.Name)
		if symbol.Kind != "function" || symbol.Signature == "" {
			t.Fatalf("symbol metadata = %#v", symbol)
		}
	}
	want := []string{"load", "save", "run"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("ExtractSymbols() = %v, want %v", got, want)
	}
}

func newTestBuilder(t *testing.T, root string) *Builder {
	t.Helper()
	builder, err := New(root, "rg")
	if err != nil {
		t.Fatal(err)
	}
	return builder
}

func writeTestFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertInternalEdges(t *testing.T, graph Graph) {
	t.Helper()
	ids := make(map[string]bool, len(graph.Nodes))
	for _, node := range graph.Nodes {
		ids[node.ID] = true
	}
	for _, edge := range graph.Edges {
		if !ids[edge.From] || !ids[edge.To] {
			t.Fatalf("edge references omitted node: %#v", edge)
		}
	}
}

func firstCallNode(t *testing.T, nodes []Node) Node {
	t.Helper()
	for _, node := range nodes {
		if node.Kind == "call" {
			return node
		}
	}
	t.Fatal("graph has no call node")
	return Node{}
}
