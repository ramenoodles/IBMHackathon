# OnBober backend

Go HTTP API for scan-first flow graphs, workspace management, and Ollama-powered explanations.

**Module:** `github.com/ibmhackathon/onbober`  
**Entry:** `go run ./cmd/api` (listens on `:8080`)

Requires `go.work` at the repo root so `grepwrapper` resolves via `replace` in `go.mod`.

---

## Package map

```
cmd/api/                 HTTP server bootstrap
internal/
  handlers/              REST + SSE endpoints
  graph/                   Flow graphs, CFG, enrich, deep dive
  scanner/                 Filesystem, ripgrep adapter, callees
    flow/                  Hybrid step extraction (regex + tree-sitter)
  llm/                     Ollama client and prompts
  workspace/               Clone / upload / register repos
```

---

## Key endpoints

| Method | Path | Handler | Notes |
|--------|------|---------|-------|
| GET | `/api/health` | `Health` | Liveness |
| GET | `/api/tree` | `Tree` | File explorer |
| GET | `/api/file` | `File` | File contents |
| GET | `/api/file/symbols` | `FileSymbols` | Symbol chips for a file |
| POST | `/api/graph/root` | `GraphRoot` | Build CFG for a symbol |
| POST | `/api/graph/expand` | `GraphExpand` | Lazy branch expansion |
| POST | `/api/graph/enrich` | `GraphEnrich` | LLM summary patches |
| GET | `/api/graph/node` | `GraphNode` | Node detail (sync) |
| GET | `/api/graph/node/stream` | `GraphNodeStream` | Node detail (SSE) |
| POST | `/api/workspace/setup` | `WorkspaceSetup` | Local path or git clone |
| POST | `/api/workspace/upload` | `WorkspaceUpload` | Zip extract |

---

## Scanner package

[`internal/scanner/`](internal/scanner/) is the filesystem and search layer.

| File | Purpose |
|------|---------|
| `scanner.go` | `SafePath`, `SafeJoin` — path traversal guards |
| `tree.go`, `reader.go` | Directory listing, file reads |
| `grep.go` | `GrepSymbol`, `GrepSymbolLang` — public search API |
| `grepwrapper_adapter.go` | Delegates to `grepwrapper/bridge` |
| `grep_literal.go` | Fixed-string ripgrep for deep-dive evidence |
| `callees.go` | `ResolveCallee`, `LanguageFromPath` |
| `symbols.go` | Per-file symbol extraction |
| `flow/` | Deterministic control-flow steps |

See [docs/GREPWRAPPER.md](../docs/GREPWRAPPER.md) for how symbol search connects to Jack's module.

---

## Graph package

[`internal/graph/`](internal/graph/) builds and enriches execution-flow graphs.

| File | Purpose |
|------|---------|
| `builder.go` | `BuildRoot`, `BuildExpand`, node detail orchestration |
| `cfg.go` | Scan steps → nodes and edges |
| `callees.go` | Attach cross-file callee hints |
| `deep_dive_context.go` | Evidence bundle before LLM deep dive |
| `enrich_build.go` | Batch LLM summaries |
| `parse.go` | Validate LLM output, deep-dive text guards |

**Design rule:** graph **topology** is always scan-derived. LLM only enriches labels and explanations.

---

## Environment

| Variable | Default |
|----------|---------|
| `PORT` | `8080` |
| `OLLAMA_URL` | `http://localhost:11434` |
| `OLLAMA_MODEL` | `qwen2.5:7b` |
| `ONBOBER_WORKSPACE_ROOT` | OS temp + `/onbober-workspaces` |

---

## Tests

```powershell
go test ./...
```

Run from repo root with `go.work`, or from this directory after `go mod tidy`.
