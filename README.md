# OnBober

**OnBober** is an AI-assisted codebase onboarding platform. Point it at a GitHub repository, a ZIP archive, or a local directory, and it generates interactive control-flow graphs for every function in your codebase—then progressively reveals those graphs to you, node by node, with AI-generated titles, summaries, and on-demand explanations tailored to your experience level and language background.

---

## Features

- **Scan-first control-flow graphs** — No heavyweight analysis framework; a fast regex-based scanner extracts steps from any language and builds a bounded CFG instantly.
- **Progressive reveal** — Start from the entry node and step through the execution path. The graph grows as you explore, preventing information overload.
- **Expandable call nodes** — Click any function call to inline its callee's CFG up to 8 levels deep.
- **AI-generated labels** — IBM Watsonx (Granite) labels every node with a short title and one-sentence summary. Falls back to deterministic regex heuristics when Watsonx is unavailable.
- **On-demand explanations** — Ask "why does this step matter?" for any node. The LLM agent can browse the codebase itself (via `read_file`, `read_context`, `search_symbol` tools) before answering.
- **Experience tuning** — Junior, Mid, or Senior modes adjust label verbosity and explanation depth.
- **Language analogies** — Select your familiar languages and the AI adds cross-language analogies to explanations.
- **Multi-language support** — Go, Python, JavaScript, TypeScript, Rust, Java, C, C++, C#.

---

## Quick Start

### Requirements

- Go 1.22+
- Node.js + pnpm
- [ripgrep](https://github.com/BurntSushi/ripgrep) (`rg` on PATH)
- Git (for GitHub workspace sources)

### Development

```sh
# Backend
cd backend
go run ./cmd/api

# Frontend (separate terminal)
cd frontend
pnpm install
pnpm run dev
```

The frontend Vite dev server proxies `/api` to `http://localhost:8080`. Open `http://localhost:5173`.

### Docker Compose (production)

```sh
# Copy the example environment file and fill in your values
cp .env.example .env

# Build and start
docker compose up --build
```

Navigate to `http://localhost:8080`. The frontend nginx container proxies `/api/*` to the backend container on the internal `onbober` Docker network.

---

## Configuration

Copy `.env.example` to `.env` and adjust as needed.

| Variable | Default | Description |
|----------|---------|-------------|
| `WATSONX_API_KEY` | — | IBM Watsonx API key (required for AI features) |
| `WATSONX_PROJECT_ID` | — | Watsonx project ID |
| `WATSONX_MODEL` | `ibm/granite-4-h-small` | Model identifier |
| `HOST` | `127.0.0.1` | Server bind address (`0.0.0.0` in Docker) |
| `PORT` | `8080` | HTTP port |
| `RG_BINARY` | `rg` | Path to ripgrep executable |
| `INCLUDE_TRAJECTORY` | `false` | Include LLM agent tool-call trajectory in explain responses |
| `MAX_BODY_BYTES` | `209715200` | Max HTTP request body (200 MiB) |
| `MAX_FILE_BYTES` | `2097152` | Max single-file read (2 MiB) |
| `MAX_REPO_BYTES` | `209715200` | Max workspace total size (200 MiB) |
| `MAX_ZIP_FILES` | `50000` | Max entries in a ZIP upload |
| `REQUEST_TIMEOUT_SECONDS` | `120` | Request deadline |
| `CLONE_TIMEOUT_SECONDS` | `120` | Git clone deadline |
| `WORKSPACE_MAX_AGE_SECONDS` | `1800` | Unused workspace eviction (30 min). `0` disables. |
| `ALLOW_LOCAL_SOURCE` | `false` | Allow "local" workspace type. **Never enable on public servers.** |

---

## Documentation

| Document | Description |
|----------|-------------|
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | System design, service topology, key data-flow diagrams, design decisions |
| [`docs/BACKEND.md`](docs/BACKEND.md) | Full backend reference: all packages, types, algorithms, security model, tests |
| [`docs/FRONTEND.md`](docs/FRONTEND.md) | Full frontend reference: views, components, composables, state, styling |
| [`docs/API.md`](docs/API.md) | Complete HTTP API reference with request/response schemas |

---

## Project Structure

```
IBMHackathon/
├── backend/                  Go 1.22 HTTP API
│   ├── cmd/api/              Server entry-point
│   ├── cmd/grepwrapper/      Standalone CLI for symbol search
│   ├── bridge/               Public façade for grepwrapper integration
│   └── internal/
│       ├── analysis/         Control-flow graph builder
│       ├── config/           Environment configuration
│       ├── httpapi/          HTTP handler & routing
│       ├── llm/              Watsonx client & agentic loop
│       ├── search/           Ripgrep wrapper
│       ├── service/          Business logic (enrich, explain, lookup)
│       ├── source/           Path-safe file reader
│       └── workspace/        Workspace lifecycle (create, evict)
├── frontend/                 Vue 3 + TypeScript SPA
│   └── src/
│       ├── api/              API client
│       ├── components/       UI & workspace components
│       ├── composables/      Logic hooks
│       ├── store/            Reactive user-context store
│       ├── types/            TypeScript type definitions
│       ├── utils/            Graph utilities
│       └── views/            Page-level route components
├── docker-compose.yml
├── .env.example
└── docs/
```

---

## Architecture Overview

```
Browser
  │
  ▼  (port 8080)
Nginx ──── /api/* ──────────► Go API (port 8080, internal)
  │                              │
  └── /* ──► Vue SPA           ├── workspace.Manager   (git clone / zip / local)
                                ├── analysis.Builder    (CFG extraction)
                                ├── service.Enrich      (AI + heuristic labels)
                                ├── service.Explain     (Watsonx agent Q&A)
                                └── search.Finder       (ripgrep symbol search)
```

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for detailed data-flow diagrams and design rationale.

---

## Grepwrapper CLI

The `grepwrapper` binary is a standalone tool for searching function declarations across a codebase:

```sh
cd backend
go build ./cmd/grepwrapper

./grepwrapper -root /path/to/repo -lang go Close
./grepwrapper -root /path/to/repo -lang python -source context -after 30 my_function
```

| Flag | Default | Description |
|------|---------|-------------|
| `-root` | `.` | Directory to search |
| `-lang` | `auto` | `auto` \| `go` \| `python` \| `javascript` \| `typescript` \| `rust` \| `java` \| `c` \| `cpp` \| `csharp` |
| `-max` | `20` | Maximum results |
| `-rg` | `rg` | Ripgrep binary path |
| `-source` | `match` | Output mode: `match` \| `context` \| `file` |
| `-before` | `5` | Context lines before match (context mode) |
| `-after` | `20` | Context lines after match (context mode) |

---

## Running Tests

```sh
# Backend
cd backend
go test ./...

# Frontend
cd frontend
pnpm test
```

---

## Security Notes

- **`ALLOW_LOCAL_SOURCE=false`** must remain `false` on any internet-facing deployment. When `true`, the API allows callers to read arbitrary filesystem paths on the server.
- The source reader validates every file path against the workspace root and resolves symlinks before access. Directory traversal attempts are rejected.
- ZIP uploads are validated for path traversal (`..` components), entry count, and total extracted size before writing to disk.
- The CORS policy currently allows all origins. For production, restrict to your frontend domain.
- Workspace IDs are opaque random 12-byte hex strings; the underlying filesystem path is never exposed to clients.
