# OnBober — HTTP API Reference

Base URL (local development): `http://localhost:8080`

All request and response bodies use `application/json` unless otherwise noted. Errors follow a consistent envelope:

```json
{ "error": { "code": "bad_request", "message": "filePath is required" } }
```

Common error codes: `bad_request`, `not_found`, `internal`, `llm_unavailable`.

---

## Health

### `GET /api/health`

Returns server status and whether the Watsonx LLM is configured.

**Response `200`**

```json
{
  "status": "ok",
  "watsonx": true
}
```

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | Always `"ok"` |
| `watsonx` | boolean | `true` when `WATSONX_API_KEY`, `WATSONX_PROJECT_ID`, and `WATSONX_MODEL` are all set |

---

## Workspaces

### `POST /api/workspaces`

Create a new workspace from a GitHub URL, a ZIP upload, or a local filesystem path (if enabled).

#### Option A — GitHub clone

```json
{
  "source": "github",
  "url": "owner/repo"
}
```

`url` accepts:
- Shorthand: `owner/repo`
- HTTPS: `https://github.com/owner/repo`
- SSH: `git@github.com:owner/repo.git`

The backend normalises all forms to HTTPS, does a pre-flight size check against the GitHub API, and then runs `git clone --depth 1 --single-branch --no-tags`.

#### Option B — ZIP upload

```
Content-Type: multipart/form-data
file: <binary zip data>
```

The ZIP is extracted into a temporary directory. Entries containing `..` are rejected. Total extracted size and file count are validated against server limits.

#### Option C — Local path (requires `ALLOW_LOCAL_SOURCE=true`)

```json
{
  "source": "local",
  "path": "/absolute/path/to/repo"
}
```

**Response `200`**

```json
{
  "id": "ws-4a9f2c8d1e7b",
  "name": "sarama",
  "source": "github"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Opaque workspace ID (`ws-<12 hex bytes>`) |
| `name` | string | Human-readable display name (repo name or filename) |
| `source` | string | `"github"` \| `"zip"` \| `"local"` |

---

### `GET /api/workspaces/{id}/tree`

List the contents of a directory within the workspace.

**Query parameters**

| Parameter | Required | Description |
|-----------|----------|-------------|
| `path` | No | Workspace-relative directory path. Defaults to the workspace root. |

**Response `200`**

```json
{
  "entries": [
    { "name": "broker.go",  "path": "broker.go",  "isDir": false },
    { "name": "consumer",   "path": "consumer",   "isDir": true  }
  ]
}
```

---

### `GET /api/workspaces/{id}/file`

Read a single file from the workspace. The file must be smaller than `MAX_FILE_BYTES` (default 2 MiB).

**Query parameters**

| Parameter | Required | Description |
|-----------|----------|-------------|
| `path` | Yes | Workspace-relative file path |

**Response `200`**

```json
{
  "path":     "broker.go",
  "content":  "package sarama\n\nimport ...",
  "language": "go"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | Echo of the requested path |
| `content` | string | Raw file contents |
| `language` | string | Detected language name (e.g. `"go"`, `"python"`, `"typescript"`) |

---

### `GET /api/workspaces/{id}/symbols`

Extract all top-level function/method declarations from a source file.

**Query parameters**

| Parameter | Required | Description |
|-----------|----------|-------------|
| `path` | Yes | Workspace-relative file path |

**Response `200`**

```json
{
  "symbols": [
    {
      "name":      "Close",
      "line":      142,
      "kind":      "function",
      "signature": "func (b *Broker) Close() error {"
    }
  ],
  "count": 47
}
```

---

## Flow Graphs

### `POST /api/workspaces/{id}/graphs`

Build the initial control-flow graph (CFG) for a named function.

**Request**

```json
{
  "filePath": "broker.go",
  "symbol":   "Close"
}
```

**Response `200`**

```json
{
  "rootId": "flow:1:YnJva2VyLmdv:Q2xvc2U=:142:0",
  "nodes": [
    {
      "id":           "flow:1:YnJva2VyLmdv:Q2xvc2U=:142:0",
      "label":        "Close",
      "title":        "",
      "summary":      "",
      "kind":         "entry",
      "confidence":   "verified",
      "labelSource":  "scan",
      "file":         "broker.go",
      "line":         142,
      "code":         "func (b *Broker) Close() error {",
      "calleeSymbol": "",
      "calleeFile":   "",
      "expandable":   false,
      "childCount":   0,
      "collapsed":    false
    }
  ],
  "edges": [
    { "from": "flow:1:...:0", "to": "flow:1:...:1", "label": "start" }
  ],
  "depth":  1,
  "symbol": "Close"
}
```

**Node fields**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Encodes `depth:base64(file):base64(symbol):line:ordinal` |
| `label` | string | Scan-time display text (raw callee name or code excerpt) |
| `title` | string | Short enriched title (populated after `/graphs/enrich`) |
| `summary` | string | One-sentence enriched summary |
| `kind` | string | `entry` \| `call` \| `branch` \| `loop` \| `return` \| `raise` \| `assign` |
| `confidence` | string | `verified` (scan-sourced) \| `inferred` (AI or truncated) |
| `labelSource` | string | `scan` \| `heuristic` \| `ai` |
| `file` | string | Workspace-relative file path |
| `line` | int | 1-based line number |
| `code` | string | Full source line |
| `calleeSymbol` | string | Function name for call nodes |
| `calleeFile` | string | Resolved callee file for call nodes |
| `expandable` | boolean | `true` if the callee CFG can be fetched via `/expand` |
| `childCount` | int | Approximate step count in callee body |
| `collapsed` | boolean | UI hint: callee is in compact preview mode |

**Edge labels**

| Label | Meaning |
|-------|---------|
| `start` | Entry → first step |
| `then` | Sequential execution |
| `true` | If-condition true path |
| `false` | If-condition false / skip path |
| `body` | Else / elif block |
| `each` | Loop body |
| `done` | After loop |
| `calls` | Expansion edge to callee root |

**Limits:** Root graphs are capped at **50 nodes**. When the function body exceeds this, a `"+N more steps"` marker node is appended.

---

### `POST /api/workspaces/{id}/graphs/expand`

Expand a call node to inline its callee's control-flow graph.

**Request**

```json
{
  "nodeId":       "flow:1:YnJva2VyLmdv:Q2xvc2U=:155:3",
  "filePath":     "broker.go",
  "symbol":       "Close",
  "parentPath":   "broker.go",
  "expandLimit":  30,
  "calleeFile":   "request.go",
  "calleeSymbol": "encode"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `nodeId` | Yes | ID of the call node to expand |
| `filePath` | Yes | File containing the parent function |
| `symbol` | Yes | Parent function name |
| `expandLimit` | No | Max nodes to return (default 30) |
| `calleeFile` | No | If known, the callee's file path |
| `calleeSymbol` | No | If known, the callee's symbol name |

**Response `200`**

Same shape as `/graphs`, plus an additional `calls` edge from `nodeId` to the new root.

```json
{
  "rootId": "flow:1:YnJva2VyLmdv:Q2xvc2U=:155:3",
  "nodes": [ ... ],
  "edges": [
    { "from": "flow:1:...:3", "to": "flow:2:...:0", "label": "calls" },
    ...
  ],
  "depth":  2,
  "symbol": "encode"
}
```

**Limits:** Expansion graphs are capped at **30 nodes**. Maximum call depth is **8 levels**.

---

### `POST /api/workspaces/{id}/graphs/enrich`

Enrich a set of graph nodes with AI-generated titles and summaries. When Watsonx is unavailable, regex-based heuristics are used instead.

**Request**

```json
{
  "filePath": "broker.go",
  "symbol":   "Close",
  "nodes": [
    {
      "id":    "flow:1:...:1",
      "line":  155,
      "code":  "b.mu.Lock()",
      "kind":  "call",
      "label": "b.mu.Lock"
    }
  ],
  "userContext": {
    "primaryLanguage":     "go",
    "experienceLevel":     "junior",
    "languageComparisons": true
  }
}
```

**`userContext` fields**

| Field | Values | Effect |
|-------|--------|--------|
| `primaryLanguage` | Language key (e.g. `"go"`) | Not currently used server-side but reserved |
| `experienceLevel` | `"junior"` \| `"mid"` \| `"senior"` | Controls label verbosity and jargon |
| `languageComparisons` | boolean | Adds cross-language analogies to summaries |

**Response `200`**

```json
{
  "patches": [
    {
      "id":          "flow:1:...:1",
      "title":       "Lock b.mu",
      "summary":     "Takes b.mu so only one goroutine runs Close at a time.",
      "labelSource": "heuristic"
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Matches the input node ID |
| `title` | string | Short label (≤8 words) specific to the actual objects in the code |
| `summary` | string | One sentence explaining why this step matters |
| `labelSource` | string | `"ai"` (Watsonx-generated) \| `"heuristic"` (regex-generated) |

---

## Code Explanation

### `POST /api/workspaces/{id}/explain`

Answer a question about a specific step in the code. Requires Watsonx to be configured; returns `llm_unavailable` otherwise.

The LLM runs an agentic loop (max 8 iterations) and may call the following tools:
- `read_file(path)` — reads an entire file from the workspace
- `read_context(path, line, before, after)` — reads lines around a target
- `search_symbol(name, language, limit)` — searches for function declarations

**Request**

```json
{
  "name":                "Close",
  "question":            "Why is the mutex locked here?",
  "language":            "go",
  "file":                "broker.go",
  "line":                155,
  "code":                "b.mu.Lock()",
  "kind":                "call",
  "title":               "Lock b.mu",
  "experience":          "junior",
  "familiarLanguages":   "Python",
  "languageComparisons": true
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Name of the parent function |
| `language` | Yes | Language key for ripgrep search |
| `file` | No | Workspace-relative path of the node |
| `line` | No | 1-based line number |
| `code` | No | Exact source snippet at the node (preferred; skips ripgrep lookup if provided) |
| `kind` | No | `entry` \| `call` \| `branch` \| `return` \| `raise` |
| `title` | No | Enriched label (provides context to the model) |
| `question` | No | Explicit question. If omitted, the backend auto-generates one from `kind`, `title`, and `code`. |
| `experience` | No | `"junior"` \| `"mid"` \| `"senior"` — adjusts explanation depth |
| `familiarLanguages` | No | Comma-separated display names, e.g. `"Python, Rust"` |
| `languageComparisons` | No | Append a brief analogy to familiar languages |

**Response `200`**

```json
{
  "path":       "broker.go",
  "line":       155,
  "start_line": 142,
  "end_line":   190,
  "explanation": "The mutex `b.mu` is locked here to ensure that only one goroutine executes `Close` at a time. Without this lock, two concurrent callers could both read `b.closed == false` and proceed to close the underlying connection twice, leading to a double-close error.\n\n*Python analogy:* This is similar to using a `threading.Lock()` as a context manager with `with lock:`.",
  "trajectory": [
    {
      "type":      "message",
      "role":      "user",
      "content":   "Source code:\n```\nb.mu.Lock()\n```\n\nQuestion: ..."
    },
    {
      "type":      "tool_call",
      "role":      "assistant",
      "name":      "read_context",
      "id":        "call_abc123",
      "arguments": "{\"path\":\"broker.go\",\"line\":155,\"before\":5,\"after\":10}"
    },
    {
      "type":      "tool_result",
      "role":      "tool",
      "name":      "read_context",
      "id":        "call_abc123",
      "content":   "File: broker.go (lines 150–165)\n\n..."
    },
    {
      "type":      "message",
      "role":      "assistant",
      "content":   "The mutex `b.mu` is locked here..."
    }
  ]
}
```

`trajectory` is only present when the server is started with `INCLUDE_TRAJECTORY=true`.

**TrajectoryEvent fields**

| Field | Description |
|-------|-------------|
| `type` | `"message"` \| `"tool_call"` \| `"tool_result"` |
| `role` | `"user"` \| `"assistant"` \| `"tool"` |
| `name` | Tool name (only for `tool_call` / `tool_result`) |
| `id` | Tool call correlation ID |
| `arguments` | JSON-encoded tool arguments (only for `tool_call`) |
| `content` | Message content or tool result text |

---

## Notes

### Workspace Lifetime

Workspaces created from GitHub clones or ZIP uploads are automatically evicted after `WORKSPACE_MAX_AGE_SECONDS` (default 30 minutes). Any request to an expired workspace returns `404 not_found`.

### Path Safety

All `path` parameters are validated server-side to prevent directory traversal. Paths containing `..` components or that resolve outside the workspace root are rejected with `bad_request`.

### Rate Limits

There are no built-in rate limits. For public deployments, place a reverse proxy (nginx, Caddy, Cloudflare) in front of the backend.

### CORS

The backend sends `Access-Control-Allow-Origin: *` on all responses. For production, restrict this to your frontend origin.
