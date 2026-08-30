package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ramenoodles/IBMHackathon/backend/internal/analysis"
	"github.com/ramenoodles/IBMHackathon/backend/internal/config"
	"github.com/ramenoodles/IBMHackathon/backend/internal/llm"
	"github.com/ramenoodles/IBMHackathon/backend/internal/service"
	"github.com/ramenoodles/IBMHackathon/backend/internal/source"
	"github.com/ramenoodles/IBMHackathon/backend/internal/workspace"
)

type Handler struct {
	Workspaces        *workspace.Manager
	RGBinary          string
	WatsonxModel      string
	WatsonxAPIKey     string
	WatsonxProjectID  string
	WatsonxEnabled    bool
	AllowLocalSource  bool
	MaxBodyBytes      int64
	MaxFileBytes      int64
}

// Options carries the tunables the API derives from server configuration.
type Options struct {
	RGBinary         string
	WatsonxModel     string
	WatsonxAPIKey    string
	WatsonxProjectID string
	WatsonxEnabled   bool
	AllowLocalSource bool
	MaxBodyBytes     int64
	MaxFileBytes     int64
}

func New(m *workspace.Manager, opts Options) *Handler {
	if opts.MaxBodyBytes <= 0 {
		opts.MaxBodyBytes = int64(config.DefaultMaxBodyBytes)
	}
	if opts.MaxFileBytes <= 0 {
		opts.MaxFileBytes = int64(source.DefaultMaxFileBytes)
	}
	return &Handler{
		Workspaces:       m,
		RGBinary:         opts.RGBinary,
		WatsonxModel:     opts.WatsonxModel,
		WatsonxAPIKey:    opts.WatsonxAPIKey,
		WatsonxProjectID: opts.WatsonxProjectID,
		WatsonxEnabled:   opts.WatsonxEnabled,
		AllowLocalSource: opts.AllowLocalSource,
		MaxBodyBytes:     opts.MaxBodyBytes,
		MaxFileBytes:     opts.MaxFileBytes,
	}
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
	mux.HandleFunc("POST /api/workspaces/{id}/graphs/enrich", h.enrich)
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
	write(w, 200, map[string]any{
		"status":           "ok",
		"watsonx":          h.WatsonxEnabled,
		"allowLocalSource": h.AllowLocalSource,
	})
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.MaxBodyBytes)
	ct := r.Header.Get("Content-Type")
	var ws workspace.Workspace
	var err error
	if strings.HasPrefix(ct, "multipart/") {
		if err = r.ParseMultipartForm(h.MaxBodyBytes); err == nil {
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
	sort.SliceStable(out, func(i, j int) bool {
		di, dj := out[i]["isDir"].(bool), out[j]["isDir"].(bool)
		if di != dj {
			return di
		}
		return out[i]["name"].(string) < out[j]["name"].(string)
	})
	write(w, 200, map[string]any{"entries": out})
}
func (h *Handler) file(w http.ResponseWriter, r *http.Request) {
	ws, ok := h.get(w, r)
	if !ok {
		return
	}
	p := r.URL.Query().Get("path")
	reader, e := source.NewReaderWithLimit(ws.Root, h.MaxFileBytes)
	if e != nil {
		fail(w, 500, e.Error())
		return
	}
	data, e := reader.ReadFile(p)
	if e != nil {
		fail(w, 404, e.Error())
		return
	}
	write(w, 200, map[string]any{"path": p, "content": data, "language": analysis.LanguageFromPath(p)})
}

func (h *Handler) symbols(w http.ResponseWriter, r *http.Request) {
	ws, ok := h.get(w, r)
	if !ok {
		return
	}
	p := r.URL.Query().Get("path")
	reader, e := source.NewReaderWithLimit(ws.Root, h.MaxFileBytes)
	if e != nil {
		fail(w, 500, e.Error())
		return
	}
	text, e := reader.ReadFile(p)
	if e != nil {
		fail(w, 404, e.Error())
		return
	}
	out := analysis.ExtractSymbols(text, p)
	write(w, 200, map[string]any{"symbols": out, "count": len(out)})
}
func (h *Handler) graph(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.MaxBodyBytes)
	ws, ok := h.get(w, r)
	if !ok {
		return
	}
	var in struct {
		FilePath string `json:"filePath"`
		Symbol   string `json:"symbol"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.FilePath == "" || in.Symbol == "" {
		fail(w, 400, "filePath and symbol are required")
		return
	}
	b, e := analysis.NewWithLimit(ws.Root, h.RGBinary, h.MaxFileBytes)
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
	r.Body = http.MaxBytesReader(w, r.Body, h.MaxBodyBytes)
	ws, ok := h.get(w, r)
	if !ok {
		return
	}
	var in struct {
		NodeID        string   `json:"nodeId"`
		FilePath      string   `json:"filePath"`
		Symbol        string   `json:"symbol"`
		ParentPath    []string `json:"parentPath"`
		ExpandLimit   int      `json:"expandLimit"`
		CalleeFile    string   `json:"calleeFile"`
		CalleeSymbol  string   `json:"calleeSymbol"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.NodeID == "" {
		fail(w, 400, "nodeId is required")
		return
	}
	b, e := analysis.NewWithLimit(ws.Root, h.RGBinary, h.MaxFileBytes)
	if e != nil {
		fail(w, 500, e.Error())
		return
	}
	g, e := b.Expand(r.Context(), in.NodeID, in.ExpandLimit, in.CalleeFile, in.CalleeSymbol)
	if e != nil {
		fail(w, 404, e.Error())
		return
	}
	write(w, 200, g)
}
func (h *Handler) enrich(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.MaxBodyBytes)
	ws, ok := h.get(w, r)
	if !ok {
		return
	}
	var in struct {
		FilePath    string `json:"filePath"`
		Symbol      string `json:"symbol"`
		Nodes       []service.EnrichNodeInput `json:"nodes"`
		UserContext service.EnrichUserContext `json:"userContext"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || len(in.Nodes) == 0 {
		fail(w, 400, "nodes are required")
		return
	}
	if in.Symbol == "" {
		fail(w, 400, "symbol is required")
		return
	}

	var client llm.Client
	if h.WatsonxEnabled {
		watsonx, e := llm.NewWatsonxClientWithLimit(h.WatsonxModel, ws.Root, h.RGBinary, h.WatsonxAPIKey, h.WatsonxProjectID, h.MaxFileBytes)
		if e != nil {
			fail(w, 500, e.Error())
			return
		}
		client = watsonx
	}

	s, e := service.NewWithLimit(ws.Root, h.RGBinary, client, false, h.MaxFileBytes)
	if e != nil {
		fail(w, 500, e.Error())
		return
	}
	res, e := s.Enrich(r.Context(), service.EnrichRequest{
		FilePath:    in.FilePath,
		Symbol:      in.Symbol,
		Nodes:       in.Nodes,
		UserContext: in.UserContext,
	})
	if e != nil {
		fail(w, 500, e.Error())
		return
	}
	write(w, 200, res)
}

func (h *Handler) explain(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.MaxBodyBytes)
	ws, ok := h.get(w, r)
	if !ok {
		return
	}
	if !h.WatsonxEnabled {
		fail(w, 503, "Watsonx provider is not configured")
		return
	}
	var in struct {
		Name       string `json:"name"`
		Question   string `json:"question"`
		Language   string `json:"language"`
		// Node context — enriches the prompt and lets us skip ripgrep when code is present.
		File                string `json:"file"`
		Line                int    `json:"line"`
		Code                string `json:"code"`
		Kind                string `json:"kind"`
		Title               string `json:"title"`
		Experience          string `json:"experience"`
		FamiliarLanguages   string `json:"familiarLanguages"`
		LanguageComparisons bool   `json:"languageComparisons"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		fail(w, 400, "invalid JSON")
		return
	}
	client, e := llm.NewWatsonxClientWithLimit(h.WatsonxModel, ws.Root, h.RGBinary, h.WatsonxAPIKey, h.WatsonxProjectID, h.MaxFileBytes)
	if e != nil {
		fail(w, 500, e.Error())
		return
	}
	s, e := service.NewWithLimit(ws.Root, h.RGBinary, client, false, h.MaxFileBytes)
	if e != nil {
		fail(w, 500, e.Error())
		return
	}
	res, e := s.Explain(r.Context(), service.ExplainRequest{
		Name:                in.Name,
		Question:            in.Question,
		Language:            in.Language,
		File:                in.File,
		Line:                in.Line,
		Code:                in.Code,
		Kind:                in.Kind,
		Title:               in.Title,
		Experience:          in.Experience,
		FamiliarLanguages:   in.FamiliarLanguages,
		LanguageComparisons: in.LanguageComparisons,
	})
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
