package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ramenoodles/IBMHackathon/backend/internal/analysis"
	"github.com/ramenoodles/IBMHackathon/backend/internal/workspace"
)

func TestGraphRoutesReturnBoundedRootAndExpansion(t *testing.T) {
	root := t.TempDir()
	content := `def child():
    one = 1
    two = 2
    three = 3
    four = 4
    return four

def parent():
    value = child()
    return value
`
	if err := os.WriteFile(filepath.Join(root, "flow.py"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	manager, err := workspace.NewManager(workspace.Limits{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	ws, err := manager.Local(root)
	if err != nil {
		t.Fatal(err)
	}
	handler := New(manager, Options{RGBinary: "rg"}).Handler()

	rootResponse := performJSON(t, handler, http.MethodPost, "/api/workspaces/"+ws.ID+"/graphs", map[string]any{
		"filePath": "flow.py",
		"symbol":   "parent",
	})
	if rootResponse.Code != http.StatusOK {
		t.Fatalf("root status = %d, body = %s", rootResponse.Code, rootResponse.Body.String())
	}
	var graph analysis.Graph
	if err := json.Unmarshal(rootResponse.Body.Bytes(), &graph); err != nil {
		t.Fatal(err)
	}
	if len(graph.Nodes) == 0 || len(graph.Nodes) > analysis.MaxRootNodes {
		t.Fatalf("root returned %d nodes", len(graph.Nodes))
	}
	var callID string
	for _, node := range graph.Nodes {
		if node.Kind == "call" {
			callID = node.ID
			break
		}
	}
	if callID == "" {
		t.Fatal("root response has no call node")
	}

	expandResponse := performJSON(t, handler, http.MethodPost, "/api/workspaces/"+ws.ID+"/graphs/expand", map[string]any{
		"nodeId":      callID,
		"filePath":    "flow.py",
		"symbol":      "parent",
		"parentPath":  []string{callID},
		"expandLimit": 3,
	})
	if expandResponse.Code != http.StatusOK {
		t.Fatalf("expand status = %d, body = %s", expandResponse.Code, expandResponse.Body.String())
	}
	var fragment analysis.Graph
	if err := json.Unmarshal(expandResponse.Body.Bytes(), &fragment); err != nil {
		t.Fatal(err)
	}
	if fragment.RootID != callID || len(fragment.Nodes) != 3 {
		t.Fatalf("expand root = %q, nodes = %d", fragment.RootID, len(fragment.Nodes))
	}
}

func TestGraphRootRequiresFilePath(t *testing.T) {
	manager, err := workspace.NewManager(workspace.Limits{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close() })
	ws, err := manager.Local(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	response := performJSON(t, New(manager, Options{RGBinary: "rg"}).Handler(), http.MethodPost, "/api/workspaces/"+ws.ID+"/graphs", map[string]any{
		"symbol": "missingFile",
	})
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func performJSON(t *testing.T, handler http.Handler, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(method, target, bytes.NewReader(encoded))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
