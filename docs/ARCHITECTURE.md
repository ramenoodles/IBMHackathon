# OnBober — System Architecture

## Overview

**OnBober** is an AI-assisted codebase onboarding platform. It ingests a source-code repository (from GitHub, a ZIP archive, or a local path), extracts control-flow graphs (CFGs) for individual functions, progressively reveals those graphs to developers, and enriches every node with AI-generated titles and on-demand explanations—tailored to the developer's experience level and language background.

```
                  ┌──────────────────────────────┐
                  │       Developer Browser       │
                  │   Vue 3 SPA (port 8080/80)    │
                  └─────────────┬────────────────┘
                                │ HTTP / SSE
                                ▼
             ┌────────────────────────────────────┐
             │         Nginx (production)          │
             │   /api/*  →  backend:8080           │
             │   /*      →  SPA index.html         │
             └───────────────┬────────────────────┘
                             │ private docker network (onbober)
                             ▼
          ┌──────────────────────────────────────────┐
          │          Go HTTP API  (port 8080)         │
          │                                          │
          │  ┌──────────┐  ┌──────────┐  ┌────────┐ │
          │  │ workspace│  │ analysis │  │  llm   │ │
          │  │ manager  │  │  (CFG)   │  │watsonx │ │
          │  └────┬─────┘  └────┬─────┘  └───┬────┘ │
          │       │             │             │      │
          │  ┌────▼──────────────▼──┐  ┌──────▼───┐  │
          │  │  source (file I/O)   │  │  search  │  │
          │  └──────────────────────┘  │(ripgrep) │  │
          │                            └──────────┘  │
          └──────────────────────────────────────────┘
```

---

## Repository Layout

```
IBMHackathon/
├── backend/                  Go HTTP API server
│   ├── cmd/
│   │   ├── api/              Server entry-point (main.go)
│   │   └── grepwrapper/      Standalone CLI tool (main.go)
│   ├── bridge/               Public integration façade
│   ├── internal/
│   │   ├── analysis/         CFG builder (flow.go, cfg.go, analysis.go)
│   │   ├── cli/              grepwrapper CLI implementation
│   │   ├── config/           Environment-variable configuration
│   │   ├── httpapi/          HTTP handler & routing
│   │   ├── llm/              Watsonx client & agent tools
│   │   ├── search/           Ripgrep wrapper (patterns, search)
│   │   ├── service/          Business logic (enrich, explain, lookup)
│   │   ├── source/           Safe file-system reader
│   │   └── workspace/        Workspace lifecycle manager
│   ├── Dockerfile
│   └── go.mod
├── frontend/                 Vue 3 + TypeScript SPA
│   ├── src/
│   │   ├── api/              API client (index.ts)
│   │   ├── components/       UI & workspace components
│   │   ├── composables/      Reusable logic hooks
│   │   ├── constants/        App-wide constants
│   │   ├── router/           Vue Router
│   │   ├── store/            Reactive user-context store
│   │   ├── types/            TypeScript type definitions
│   │   ├── utils/            Graph utilities
│   │   └── views/            Page-level components
│   ├── Dockerfile
│   └── vite.config.ts
├── docker-compose.yml
├── .env.example
└── docs/
    ├── ARCHITECTURE.md       ← this file
    ├── BACKEND.md
    ├── FRONTEND.md
    └── API.md
```

---

## Service Topology (Docker Compose)

| Service | Image Base | Port(s) | Notes |
|---------|-----------|---------|-------|
| `backend` | `golang:1.24-alpine` → `alpine:3.21` | 8080 (internal) | Exposes only to the `onbober` network |
| `frontend` | `node:22-alpine` → `nginx:alpine` | `8080:80` | Proxies `/api/*` to `backend:8080` |

Both services are attached to the `onbober` Docker bridge network. Docker's internal DNS resolves the name `backend` inside the nginx config, keeping host ports unexposed to the backend container.

### Development Mode

In development (without Docker):
- `cd backend && go run ./cmd/api` starts the Go server on `localhost:8080`
- `cd frontend && pnpm dev` starts Vite dev server (default port 5173) and proxies `/api/*` to `localhost:8080`

---

## Data Flow: Repository → Graph → Explanation

### 1. Workspace Creation

```
Browser  POST /api/workspaces
             │
     ┌───────▼────────┐
     │ workspace.Manager│
     │                │
     │  github URL →  │  git clone --depth 1
     │  zip upload →  │  extract to temp dir
     │  local path →  │  reference directly
     └───────┬────────┘
             │  returns workspace { id, name, source }
             ▼
     sessionStorage  (workspaceId stored in SPA)
```

### 2. Control-Flow Graph Generation

```
Browser  POST /api/workspaces/{id}/graphs
             │  { filePath, symbol }
             ▼
     analysis.Builder.Root()
             │
     1. source.Reader.ReadFile(filePath)
     2. analysis.extractFlow()          ← parse lines → flowSteps
     3. analysis.buildCFG()             ← flowSteps → nodes + edges
     4. analysis.resolveCalls()         ← ripgrep each call → mark expandable
             │
             ▼
     { rootId, nodes[], edges[], depth:1 }
```

### 3. Progressive Reveal & Enrichment

```
Graph arrives in browser
        │
  useFlowGraph.loadRoot()
        │
  entryOnlyRevealedIds()      ← show only entry node
        │
  User clicks "next step"
        │
  revealFromNode()             ← BFS reveal
        │
  enrichNodes()  [background] ← POST /graphs/enrich → Watsonx batch labels
        │
  User clicks call node
        │
  expandNode()                 ← POST /graphs/expand → callee CFG merged in
```

### 4. AI Explanation (per node)

```
User clicks "Explain this step"
        │
  api.explain()   POST /api/workspaces/{id}/explain
        │
  WatsonxClient.AskAgent()     ← agentic loop (max 8 steps)
        │
  LLM may call tools:
    read_file(path)
    read_context(path, line, ±N)
    search_symbol(name, lang)
        │
  Returns { explanation, trajectory[] }
```

---

## Key Design Decisions

### Bounded Graphs
CFGs are deliberately capped (root: 50 nodes, expansion: 30 nodes, depth: 8 levels). This prevents server overload on large files and keeps the UI scannable. Truncation is explicit—a "+N more steps" marker node is added, coloured differently so the user knows the graph is partial.

### Stateless Node IDs
Every node's ID encodes `depth:base64(file):base64(symbol):line:ordinal`. Expansion requests carry no server-side session state—the backend can reconstruct a callee's graph from the node ID alone.

### Progressive Reveal
The frontend never dumps the full graph onto the canvas at once. It starts with the entry node and reveals neighbors as the user steps through, maintaining a `SILENT_BUFFER_STEPS` lookahead for smooth pre-enrichment.

### AI + Heuristic Fallback
Enrichment tries Watsonx first. When the LLM returns a generic title (e.g., "Acquire lock"), it is silently replaced by a regex-based heuristic that names actual objects (e.g., "Lock m.mu"). When Watsonx is unavailable, heuristics cover the entire graph.

### Workspace Eviction
Cloned/extracted workspaces are stored in temporary directories and are evicted (background ticker) after `WORKSPACE_MAX_AGE_SECONDS` (default 30 minutes). This prevents disk exhaustion on shared or public deployments.

### Security Boundaries
- All file reads go through `source.Reader`, which validates every path against the workspace root via `filepath.Rel` + symlink resolution.
- ZIP extraction rejects any entry containing `..`.
- `ALLOW_LOCAL_SOURCE` is `false` by default and must never be enabled on a public-facing server.
- The backend CORS policy allows all origins (suitable for hackathon; restrict in production).

---

## Technology Choices

| Concern | Choice | Rationale |
|---------|--------|-----------|
| Backend language | Go 1.22 | Standard library HTTP + lean dependencies; single static binary |
| Symbol search | ripgrep | Fastest multi-language regex search; structured JSON output |
| AI provider | IBM Watsonx | Hackathon IBM integration; Granite models |
| Frontend framework | Vue 3 + Composition API | Reactive, ergonomic; excellent TypeScript support |
| Bundler | Vite 8 | Fast HMR; native ESM; first-class Vue plugin |
| CSS | Tailwind CSS 4 (Vite plugin) | Utility-first; no runtime overhead |
| Graph visualisation | Mermaid 11 | Renders flowchart from text DSL; easy to generate from data |
| Code editor | Monaco (CDN) | VS Code editor; line decoration; syntax highlighting |
| Pan/zoom | Panzoom 4 | Minimal, pointer-event-based; no canvas dependency |
| Container runtime | Docker + Compose | Reproducible deployment; multi-stage builds for minimal images |
