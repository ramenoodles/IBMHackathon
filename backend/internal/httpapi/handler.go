package httpapi

import (
	"encoding/json"
	"fmt"
	"grepwrapper/internal/analysis"
	"grepwrapper/internal/llm"
	"grepwrapper/internal/service"
	"grepwrapper/internal/source"
	"grepwrapper/internal/workspace"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Handler struct {
	Workspaces     *workspace.Manager
	RGBinary       string
	WatsonxModel   string
	WatsonxEnabled bool
}

func New(m *workspace.Manager, rg, model string, enabled bool) *Handler {
	return &Handler{m, rg, model, enabled}
}
func (h *Handler) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", h.health)
	mux.HandleFunc("POST /api/workspaces", h.create)
	mux.HandleFunc("GET /api/workspaces/{id}/tree", h.tree)
	mux.HandleFunc("GET /api/workspaces/{id}/file", h.file)
	mux.HandleFunc("GET /api/workspaces/{id}/symbols", h.symbols)
	mux.HandleFunc("POST /api/workspaces/{id}/graphs", h.graph)
	mux.HandleFunc("POST /api/workspaces/{id}/graphs/expand", h.expand)
	mux.HandleFunc("POST /api/workspaces/{id}/explain", h.explain)
	return cors(mux)
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, msg string) {
	write(w, status, map[string]any{"error": map[string]string{"code": http.StatusText(status), "message": msg}})
}
func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	write(w, 200, map[string]any{"status": "ok", "watsonx": h.WatsonxEnabled})
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 200<<20)
	ct := r.Header.Get("Content-Type")
	var ws workspace.Workspace
	var err error
	if strings.HasPrefix(ct, "multipart/") {
		if err = r.ParseMultipartForm(200 << 20); err == nil {
			f, head, formErr := r.FormFile("file")
			if formErr != nil {
				err = formErr
			}
			if err == nil {
				defer f.Close()
				ws, err = h.Workspaces.Zip(f, head.Filename, head.Size)
			}
		}
	} else {
		var in struct {
			Source string `json:"source"`
			Path   string `json:"path"`
			URL    string `json:"url"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			fail(w, 400, "invalid JSON")
			return
		}
		switch in.Source {
		case "local":
			ws, err = h.Workspaces.Local(in.Path)
		case "github":
			ws, err = h.Workspaces.GitHub(r.Context(), in.URL)
		default:
			err = fmt.Errorf("source must be local or github")
		}
	}
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	write(w, 201, map[string]any{"id": ws.ID, "name": ws.Name, "source": ws.Source})
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) (workspace.Workspace, bool) {
	ws, ok := h.Workspaces.Get(r.PathValue("id"))
	if !ok {
		fail(w, 404, "workspace not found")
	}
	return ws, ok
}
func (h *Handler) tree(w http.ResponseWriter, r *http.Request) {
	ws, ok := h.get(w, r)
	if !ok {
		return
	}
	dir := r.URL.Query().Get("path")
	root := filepath.Join(ws.Root, filepath.Clean(dir))
	rel, e := filepath.Rel(ws.Root, root)
	if e != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		fail(w, 400, "invalid directory")
		return
	}
	ents, e := os.ReadDir(root)
	if e != nil {
		fail(w, 404, "directory not found")
		return
	}
	out := []map[string]any{}
	for _, x := range ents {
		p := filepath.ToSlash(filepath.Join(rel, x.Name()))
		if p == "." {
			p = x.Name()
		}
		out = append(out, map[string]any{"name": x.Name(), "path": p, "isDir": x.IsDir()})
	}
	write(w, 200, map[string]any{"entries": out})
}
func (h *Handler) file(w http.ResponseWriter, r *http.Request) {
	ws, ok := h.get(w, r)
	if !ok {
		return
	}
	p := r.URL.Query().Get("path")
	reader, e := source.NewReader(ws.Root)
	if e != nil {
		fail(w, 500, e.Error())
		return
	}
	data, e := reader.ReadFile(p)
	if e != nil {
		fail(w, 404, e.Error())
		return
	}
	write(w, 200, map[string]any{"path": p, "content": data, "language": filepath.Ext(p)})
}

var decl = regexp.MustCompile(`(?m)^\s*(?:func\s+(?:\([^)]*\)\s+)?|(?:async\s+)?def\s+|class\s+|(?:pub\s+)?fn\s+)([A-Za-z_]\w*)`)

func (h *Handler) symbols(w http.ResponseWriter, r *http.Request) {
	ws, ok := h.get(w, r)
	if !ok {
		return
	}
	p := r.URL.Query().Get("path")
	reader, e := source.NewReader(ws.Root)
	if e != nil {
		fail(w, 500, e.Error())
		return
	}
	text, e := reader.ReadFile(p)
	if e != nil {
		fail(w, 404, e.Error())
		return
	}
	type sym struct {
		Name string `json:"name"`
		Line int    `json:"line"`
	}
	out := []sym{}
	for _, m := range decl.FindAllStringSubmatchIndex(text, -1) {
		name := text[m[2]:m[3]]
		out = append(out, sym{name, 1 + strings.Count(text[:m[0]], "\n")})
	}
	write(w, 200, map[string]any{"symbols": out, "count": len(out)})
}
func (h *Handler) graph(w http.ResponseWriter, r *http.Request) {
	ws, ok := h.get(w, r)
	if !ok {
		return
	}
	var in struct {
		FilePath string `json:"filePath"`
		Symbol   string `json:"symbol"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.Symbol == "" {
		fail(w, 400, "filePath and symbol are required")
		return
	}
	b, e := analysis.New(ws.Root, h.RGBinary)
	if e != nil {
		fail(w, 500, e.Error())
		return
	}
	g, e := b.Root(r.Context(), in.FilePath, in.Symbol)
	if e != nil {
		fail(w, 404, e.Error())
		return
	}
	write(w, 200, g)
}
func (h *Handler) expand(w http.ResponseWriter, r *http.Request) {
	ws, ok := h.get(w, r)
	if !ok {
		return
	}
	var in struct {
		NodeID           string `json:"nodeId"`
		FilePath, Symbol string
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		fail(w, 400, "invalid JSON")
		return
	}
	b, e := analysis.New(ws.Root, h.RGBinary)
	if e != nil {
		fail(w, 500, e.Error())
		return
	}
	g, e := b.Root(r.Context(), in.FilePath, in.Symbol)
	if e != nil {
		fail(w, 404, e.Error())
		return
	}
	g, e = b.Expand(r.Context(), g, in.NodeID, 3)
	if e != nil {
		fail(w, 404, e.Error())
		return
	}
	write(w, 200, g)
}
func (h *Handler) explain(w http.ResponseWriter, r *http.Request) {
	ws, ok := h.get(w, r)
	if !ok {
		return
	}
	if !h.WatsonxEnabled {
		fail(w, 503, "Watsonx provider is not configured")
		return
	}
	var in struct{ Name, Question, Language string }
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		fail(w, 400, "invalid JSON")
		return
	}
	client, e := llm.NewWatsonxClient(h.WatsonxModel, ws.Root, h.RGBinary)
	if e != nil {
		fail(w, 500, e.Error())
		return
	}
	s, e := service.New(ws.Root, h.RGBinary, client, false)
	if e != nil {
		fail(w, 500, e.Error())
		return
	}
	res, e := s.Explain(r.Context(), service.ExplainRequest{Name: in.Name, Question: in.Question, Language: in.Language})
	if e != nil {
		fail(w, 500, e.Error())
		return
	}
	write(w, 200, res)
}
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
