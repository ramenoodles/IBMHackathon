package graph

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ibmhackathon/onbober/internal/llm"
	"github.com/ibmhackathon/onbober/internal/scanner"
	"github.com/ibmhackathon/onbober/internal/scanner/flow"
)

func TestBuildRootScanOnly(t *testing.T) {
	root := t.TempDir()
	filePath := "setup.py"
	full := filepath.Join(root, filePath)
	content := `class Tests(unittest.TestCase):
    def setUp(self):
        self.engine = create_engine(TEST_DATABASE_URL)
        self.session = Session(bind=self.engine)
`
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	sc := scanner.New()
	b := NewBuilder(sc, llm.NewOllamaClient("http://127.0.0.1:1", "test"))
	g, err := b.BuildRoot(context.Background(), BuildInput{
		WorkspacePath: root,
		FilePath:      filePath,
		Symbol:        "setUp",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Nodes) < 2 {
		t.Fatalf("expected scan graph nodes, got %d", len(g.Nodes))
	}
	if g.RootID == "" {
		t.Fatal("expected root id")
	}
}

func TestBuildCFGGraphFromFixture(t *testing.T) {
	path := filepath.Join("..", "scanner", "flow", "testdata", "setup.py")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Skip("fixture missing")
	}
	steps := flow.ExtractFlow(string(b), "setup.py", "setUp", "python")
	g := BuildCFGGraph("setUp", "setup.py", steps)
	if len(g.Edges) < 1 {
		t.Fatal("expected edges in linear setup graph")
	}
}
