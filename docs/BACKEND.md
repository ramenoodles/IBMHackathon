# OnBober — Backend Reference

The backend is a single Go binary (`cmd/api`) that serves a JSON REST API. It has no runtime framework dependencies—routing is handled by Go 1.22's built-in `http.ServeMux`.

## Table of Contents
1. [Project Layout](#project-layout)
2. [Configuration](#configuration)
3. [HTTP Handler (`internal/httpapi`)](#http-handler)
4. [Workspace Manager (`internal/workspace`)](#workspace-manager)
5. [Analysis Package (`internal/analysis`)](#analysis-package)
6. [Service Layer (`internal/service`)](#service-layer)
7. [LLM Integration (`internal/llm`)](#llm-integration)
8. [Search (`internal/search`)](#search)
9. [Source Reader (`internal/source`)](#source-reader)
10. [Bridge Package (`bridge/`)](#bridge-package)
11. [Grepwrapper CLI (`cmd/grepwrapper`)](#grepwrapper-cli)
12. [Error Handling Conventions](#error-handling-conventions)
13. [Resource Limits & Security](#resource-limits--security)
14. [Testing](#testing)

---

## Project Layout

```
backend/
├── cmd/
│   ├── api/main.go           ← HTTP server entry-point
│   └── grepwrapper/main.go   ← Standalone CLI entry-point
├── bridge/
│   ├── bridge.go             ← Public façade over search + source
│   └── README.md
├── internal/
│   ├── analysis/
│   │   ├── analysis.go       ← Builder: Root(), Expand()
│   │   ├── cfg.go            ← buildCFG(), edge rules, limitGraph()
│   │   └── flow.go           ← extractFlow(), classifyLine(), ExtractSymbols()
│   ├── cli/
│   │   └── cli.go            ← grepwrapper flag parsing + execution
│   ├── config/
│   │   └── config.go         ← Environment variable loading
│   ├── httpapi/
│   │   └── handler.go        ← HTTP mux, all route handlers
│   ├── llm/
│   │   ├── client.go         ← Client / AgentClient / EnrichClient interfaces
│   │   ├── tools.go          ← Agent tools: read_file, read_context, search_symbol
│   │   └── watsonx.go        ← WatsonxClient implementation, agentic loop
│   ├── search/
│   │   ├── patterns.go       ← Language-specific regex patterns + globs
│   │   └── search.go         ← Finder: run ripgrep, parse JSON output
│   ├── service/
│   │   ├── enrich.go         ← Enrich flow nodes with AI labels
│   │   ├── enrich_heuristics.go ← Regex-based fallback labeling
│   │   ├── explain.go        ← Answer questions about code via Watsonx
│   │   └── lookup.go         ← Find + read symbol declarations
│   ├── source/
│   │   └── source.go         ← Path-safe file I/O within workspace root
│   └── workspace/
│       └── manager.go        ← Local / GitHub / Zip workspace creation & eviction
├── Dockerfile
├── go.mod
└── README.md
```

---

## Configuration

`internal/config/config.go` — loaded once at startup from environment variables (or a `.env` file via `godotenv` in development).

| Variable | Default | Type | Description |
|----------|---------|------|-------------|
| `HOST` | `127.0.0.1` | string | Server bind address. Docker Compose overrides to `0.0.0.0`. |
| `PORT` | `8080` | string | TCP port. |
| `RG_BINARY` | `rg` | string | Path to the `ripgrep` executable. |
| `WATSONX_API_KEY` | — | string | IBM Watsonx API key. Required for AI features. |
| `WATSONX_PROJECT_ID` | — | string | Watsonx project ID. Required for AI features. |
| `WATSONX_MODEL` | — | string | Model identifier, e.g. `ibm/granite-4-h-small`. |
| `INCLUDE_TRAJECTORY` | `false` | bool | Include agent tool-call trajectory in explain responses. |
| `MAX_BODY_BYTES` | `209715200` (200 MiB) | int64 | Maximum HTTP request body. |
| `MAX_FILE_BYTES` | `2097152` (2 MiB) | int64 | Maximum bytes read from a single file. |
| `MAX_REPO_BYTES` | `209715200` (200 MiB) | int64 | Maximum total workspace size. |
| `MAX_ZIP_FILES` | `50000` | int | Maximum entries in an uploaded ZIP. |
| `REQUEST_TIMEOUT_SECONDS` | `120` | duration | Server-side request deadline. |
| `CLONE_TIMEOUT_SECONDS` | `120` | duration | `git clone` deadline. |
| `WORKSPACE_MAX_AGE_SECONDS` | `1800` (30 min) | duration | Unused workspace eviction interval. Set `0` to disable. |
| `ALLOW_LOCAL_SOURCE` | `false` | bool | Allow "local" workspace type. **Never enable on public servers.** |

---

## HTTP Handler

`internal/httpapi/handler.go`

### Handler struct

```go
type Handler struct {
    Workspaces       *workspace.Manager
    RGBinary         string
    WatsonxModel     string
    WatsonxAPIKey    string
    WatsonxProjectID string
    WatsonxEnabled   bool   // true when all three Watsonx fields are set
    MaxBodyBytes     int64
    MaxFileBytes     int64
}
```

All routes are registered in `Handler.ServeHTTP` using `http.ServeMux`.

### Route Table

| Method | Pattern | Handler func | Notes |
|--------|---------|-------------|-------|
| `GET` | `/api/health` | `health()` | |
| `POST` | `/api/workspaces` | `create()` | |
| `GET` | `/api/workspaces/{id}/tree` | `tree()` | Query: `path` |
| `GET` | `/api/workspaces/{id}/file` | `file()` | Query: `path` |
| `GET` | `/api/workspaces/{id}/symbols` | `symbols()` | Query: `path` |
| `POST` | `/api/workspaces/{id}/graphs` | `graph()` | |
| `POST` | `/api/workspaces/{id}/graphs/expand` | `expand()` | |
| `POST` | `/api/workspaces/{id}/graphs/enrich` | `enrich()` | |
| `POST` | `/api/workspaces/{id}/explain` | `explain()` | |

A CORS middleware wraps every response with `Access-Control-Allow-Origin: *`.

### Error responses

All errors are returned as JSON:

```json
{ "error": { "code": "not_found", "message": "workspace not found" } }
```

Common codes: `bad_request`, `not_found`, `internal`, `llm_unavailable`.

---

## Workspace Manager

`internal/workspace/manager.go`

### Workspace struct

```go
type Workspace struct {
    ID, Name, Source string   // ID = "ws-<12 hex bytes>"
    Root             string   // absolute path to workspace directory
    Temporary        bool     // managed dirs are cleaned up on eviction
    CreatedAt        time.Time
}
```

### Creating workspaces

**Local** (`Manager.Local(path string)`):
- Requires `Limits.AllowLocalSource = true`.
- Returns a non-temporary workspace pointing to the provided absolute path.
- Intended for development use only.

**GitHub** (`Manager.GitHub(ctx, url string)`):
- Accepts `owner/repo`, HTTPS, or SSH formats; normalised to HTTPS.
- Pre-flight size check via the GitHub REST API (`GET /repos/{owner}/{repo}`).
- Clones with `git clone --depth 1 --single-branch --no-tags` into a temp dir.
- Post-flight walk sums all file sizes; rejects if total exceeds `MaxRepoBytes`.

**ZIP upload** (`Manager.Zip(file, name, size)`):
- Validates upload size before extracting.
- Rejects entries with `..` path components.
- Validates entry count and total extracted size against limits.

### Eviction

A background goroutine (started at construction) ticks every `WorkspaceMaxAge / 2` and removes all temporary workspaces older than `WorkspaceMaxAge`. Temp directories are deleted from disk with `os.RemoveAll`.

---

## Analysis Package

`internal/analysis/`

### Constants

| Constant | Value | Purpose |
|----------|-------|---------|
| `MaxRootNodes` | 50 | Nodes in an initial `Root()` graph |
| `MaxExpandNodes` | 30 | Nodes when expanding a call node |
| `MaxDepth` | 8 | Maximum call-nesting depth |

### Node struct (key fields)

```go
type Node struct {
    ID           string   // "flow:{depth}:{b64(file)}:{b64(symbol)}:{line}:{ordinal}"
    Label        string   // display text (raw callee name or code)
    Title        string   // enriched short title (set by service.Enrich)
    Summary      string   // enriched one-sentence summary
    Kind         string   // entry | call | branch | loop | return | raise | assign
    File         string   // workspace-relative path
    Line         int
    Code         string   // full source line
    CalleeSymbol string   // if kind == "call"
    CalleeFile   string   // resolved callee file
    Expandable   bool
    ChildCount   int
    Collapsed    bool     // UI hint (compacted callee preview)
    Confidence   string   // "verified" | "inferred"
}
```

### Edge struct

```go
type Edge struct {
    From, To string
    Label    string  // start | then | true | false | body | each | done | calls
}
```

### CFG Building

**`Builder.Root(ctx, file, symbol)`** — entry point for initial graph.

Calls `buildFunction(file, symbol, depth=1, limit=MaxRootNodes, addTruncationMarker=true)`:

1. **`extractFlow(content, file, symbol)`** (`flow.go`): Locates the function declaration (language-specific regex), extracts the body (indent-based for Python; brace-counting for all others), classifies each line into a `flowStep` with a `kind` and a `calleeSymbol`.
2. **`buildCFG(file, symbol, depth, steps, limit, truncationMarker)`** (`cfg.go`): Assigns node IDs, creates edges following CFG rules (see below), truncates to limit.
3. **`resolveCalls(ctx, graph, content)`**: For each call node, runs a ripgrep search for the callee; if found, extracts its own flow steps to determine `ChildCount` and sets `Expandable = true`.

**`Builder.Expand(ctx, nodeID, limit, calleeFile, calleeSymbol)`**:

1. Decodes depth + file from the node ID.
2. Validates depth < `MaxDepth`.
3. Calls `buildFunction(calleeFile, calleeSymbol, depth+1, limit, false)`.
4. Prepends an edge `{From: nodeID, To: callee.RootID, Label: "calls"}`.

### CFG Edge Rules (`cfg.go`)

| Step kind | Outgoing edges |
|-----------|---------------|
| `entry` | `start` → next |
| `call`, `assign` | `then` → next |
| `branch` (if/elif) | `true` → body block, `false` → next at same indent |
| `branch` (else) | `body` → body block |
| `loop` | `each` → loop body, `done` → after loop |
| `return`, `raise` | — (terminal) |

### Flow extraction (`flow.go`)

Supported languages: Go, Python, JavaScript, TypeScript, Rust, Java, C, C++, C#.

Key helpers:
- `LanguageFromPath(path)` — extension → language name.
- `ExtractSymbols(content, file)` — scans for all function declarations; returns `[]Symbol`.
- `primaryCallee(line)` — finds the rightmost non-noise function call in a line (ignores builtins like `len`, `append`, `print`).

---

## Service Layer

`internal/service/`

### `Service` struct

```go
type Service struct {
    root              string
    finder            *search.Finder
    reader            *source.Reader
    llm               llmclient.Client
    includeTrajectory bool
}
```

### Enrich (`enrich.go` + `enrich_heuristics.go`)

**`Service.Enrich(ctx, req EnrichRequest) (EnrichResult, error)`**

Input per node: `ID`, `Line`, `Code`, `Kind`, `Label`.
User context: `PrimaryLanguage`, `ExperienceLevel`, `LanguageComparisons`.

Algorithm:

1. Build a structured prompt (`buildEnrichPrompt`) including parent function name, experience-level instructions, forbidden generic examples, and all node JSON.
2. Call `llm.EnrichBatch(ctx, prompt)` → raw model response string.
3. Strip markdown fences; parse first `{...}` as JSON patches.
4. For each AI patch: if the title matches a known-generic phrase, replace with `contextualHeuristicLabel`.
5. For nodes the LLM skipped: apply `contextualHeuristicLabel` directly.
6. Return `EnrichPatch[]` with `LabelSource` = `"ai"` or `"heuristic"`.

**Heuristic patterns** (`enrich_heuristics.go`): regex-matches concrete patterns like `m.mu.Lock()`, `m.cond.Broadcast()`, `if m.closed {` and produces precise titles/summaries naming the actual object (e.g., "Lock m.mu").

### Explain (`explain.go`)

**`Service.Explain(ctx, req ExplainRequest) (ExplainResult, error)`**

Returns `ErrLLMUnavailable` if no LLM client is configured.

If `req.Code` is provided it is sent directly to the model (no ripgrep needed). Otherwise, `Lookup` resolves the symbol and reads context lines from disk.

Question building:
- Auto-generates a concise prompt from `kind`, `file`, `line`, `code`, `title`.
- Appends audience calibration: **junior** = define jargon; **senior** = concise + edge cases.
- Optionally appends a cross-language analogy request if `LanguageComparisons` is true.

Returns `ExplainResult` with `Answer` and, when `includeTrajectory` is true, the full `Trajectory` of agent tool calls.

### Lookup (`lookup.go`)

**`Service.Lookup(ctx, req LookupRequest) ([]LookupResult, error)`**

Runs a ripgrep search for `req.Name` in the workspace root, then optionally reads surrounding lines for each match. Used by `Explain` when no inline code is available.

---

## LLM Integration

`internal/llm/`

### Interfaces (`client.go`)

```go
// Plain question → answer
type Client interface {
    Ask(ctx context.Context, question, source string) (string, error)
}

// Agentic question → answer + tool-call trajectory
type AgentClient interface {
    AskAgent(ctx context.Context, question, source string) (AgentResponse, error)
}

// Bulk label generation for CFG nodes
type EnrichClient interface {
    EnrichBatch(ctx context.Context, prompt string) (string, error)
}
```

### Watsonx client (`watsonx.go`)

`WatsonxClient` implements all three interfaces.

**`AskAgent`**: Builds a system prompt ("You are a software engineering assistant…") and calls `runAgent`.

**`runAgent`** (agentic loop, max 8 iterations):
1. Call `model.Chat(ctx, messages, tools, auto_tool_choice)`.
2. If the model calls tools → execute each tool, append result messages, continue.
3. If no tools called → extract the final answer text and return.

LLM parameters: temperature `0.2`, max tokens `2048`.

**`EnrichBatch`**: Single Chat call (no tools, no loop). System: "You label execution-flow steps… Respond with strict JSON only." Returns raw model text for further parsing.

### Agent tools (`tools.go`)

| Tool name | Parameters | Description |
|-----------|-----------|-------------|
| `read_file` | `path` | Read entire file; truncated to 16 KB |
| `read_context` | `path`, `line`, `before`, `after` | Read lines around target |
| `search_symbol` | `name`, `language`, `limit` | Ripgrep symbol search |

Tool output is capped at `maxToolOutput = 16000` bytes to stay within context windows.

---

## Search

`internal/search/`

### Finder

```go
type Finder struct { binary string }
type Query  struct { Name, Root, Language string; Limit int }
type Match  struct { Path string; Line int; Text string }
```

**`Finder.Find(ctx, query)`** builds a ripgrep command:

```
rg --json --line-number --color=never --case-sensitive \
   --glob <ext-pattern> \
   --regexp <declaration-regex> \
   -- .
```

Run with `cwd = query.Root`. Output is newline-delimited JSON; only `"type":"match"` events are kept. Matches are sorted by path then line; capped at `query.Limit`.

Exit code `1` from ripgrep means no matches (not an error).

### Patterns (`patterns.go`)

Maps each language to one or more file globs and one or more declaration regexes. The token `{{name}}` is replaced with `regexp.QuoteMeta(symbolName)`.

| Language | Globs |
|----------|-------|
| Go | `*.go` |
| Python | `*.py` |
| JavaScript | `*.js`, `*.jsx`, `*.mjs`, `*.cjs` |
| TypeScript | `*.ts`, `*.tsx`, `*.mts`, `*.cts` |
| Rust | `*.rs` |
| Java | `*.java` |
| C | `*.c`, `*.h` |
| C++ | `*.cc`, `*.cpp`, `*.cxx`, `*.hh`, `*.hpp`, `*.hxx` |
| C# | `*.cs` |

Language `"auto"` applies all patterns and all globs simultaneously.

---

## Source Reader

`internal/source/source.go`

```go
type Reader struct { root string; maxFileBytes int64 }
```

**`ReadFile(relativePath)`**: validates path doesn't escape root (resolves symlinks, uses `filepath.Rel`), checks file size ≤ `maxFileBytes`, reads and returns content.

**`ReadContext(relativePath, line, before, after)`**: returns a `Snippet` with `StartLine`, `EndLine`, and `Content` (the requested line range, clamped to file bounds). `line` is 1-based.

**Path validation** (`resolve`): `filepath.Clean` → join with root → check relative → `filepath.EvalSymlinks` → re-check relative. Any path that escapes the root is rejected.

---

## Bridge Package

`bridge/bridge.go`

A thin public façade that re-exports `search.Finder` and `source.Reader` under cleaner constructors. Intended for use by the standalone `grepwrapper` module (see `bridge/README.md` for sync instructions).

```go
type Query   = search.Query    // alias
type Match   = search.Match    // alias
type Snippet = source.Snippet  // alias

func NewFinder(rgBinary string) *Finder          // "" → defaults to "rg"
func NewReader(root string) (*Reader, error)
```

---

## Grepwrapper CLI

`cmd/grepwrapper/main.go` + `internal/cli/cli.go`

A standalone command-line tool for searching function declarations.

```
grepwrapper [flags] FUNCTION_NAME
```

| Flag | Default | Description |
|------|---------|-------------|
| `-root` | `.` | Directory to search |
| `-lang` | `auto` | Language filter |
| `-max` | `20` | Max results |
| `-rg` | `rg` | Ripgrep binary path |
| `-source` | `match` | Output mode: `match` \| `context` \| `file` |
| `-before` | `5` | Context lines before match |
| `-after` | `20` | Context lines after match |

**Output modes:**
- `match` — `path:line:text` compact format
- `context` — code snippet with configurable padding
- `file` — complete source file for each match

---

## Error Handling Conventions

- HTTP handlers return structured `{"error":{"code":"...","message":"..."}}` JSON with an appropriate HTTP status.
- Internal packages return typed sentinel errors (e.g., `service.ErrLLMUnavailable`, `service.ErrSymbolNotFound`) which the handler maps to appropriate HTTP codes.
- Ripgrep exit code `1` (no matches) is normalised to an empty result, not an error.
- `exec.ErrNotFound` from ripgrep produces a human-readable message directing the user to install ripgrep.

---

## Resource Limits & Security

### Size guards
- Request body: `http.MaxBytesReader` at `MaxBodyBytes`.
- Per-file reads: checked before `os.ReadFile` in `source.Reader`.
- Total workspace: post-clone directory walk (`dirSizeExceeds`); ZIP extraction tracked incrementally.

### Path traversal prevention
- `source.Reader.resolve()` rejects any path with `..` after `filepath.Clean`, joins with root, calls `filepath.Rel` to verify containment, then re-validates after symlink resolution via `filepath.EvalSymlinks`.
- ZIP entry paths are checked for `..` components before extraction.

### Access control
- `ALLOW_LOCAL_SOURCE` is `false` by default; a `local` workspace creation attempt is rejected with `bad_request` when disabled.
- CORS allows all origins (acceptable for a hackathon deployment; restrict in production).

### Workspace isolation
- Each workspace has an opaque random ID (`ws-<12-byte hex>`).
- The `Root` path is never exposed to clients; only relative paths within the workspace are accepted from client requests.

---

## Testing

Tests live alongside the packages they test (`_test.go` files).

| File | What it tests |
|------|--------------|
| `internal/analysis/analysis_test.go` | CFG bounds (root ≤50 nodes, expand ≤30 nodes), error cases |
| `internal/analysis/bench_test.go` | Performance benchmarks for CFG building |
| `internal/cli/cli_test.go` | CLI flag parsing and output modes |
| `internal/httpapi/handler_test.go` | Graph API bounds, missing-field validation |
| `internal/llm/tools_test.go` | Agent tool execution and output truncation |
| `internal/llm/watsonx_test.go` | Watsonx client round-trips (with fake server) |
| `internal/search/patterns_test.go` | Pattern generation and symbol name validation |
| `internal/search/search_test.go` | Ripgrep invocation and JSON parsing |
| `internal/service/enrich_test.go` | AI + heuristic enrichment, prompt generation, fallback logic |
| `internal/service/service_test.go` | Explain service (LLM unavailable, plain, agent modes) |
| `internal/source/source_test.go` | Path traversal rejection, file reading, context windowing |
| `internal/workspace/manager_test.go` | Workspace creation, eviction, URL normalisation |

Run all tests:

```sh
cd backend
go test ./...
```
