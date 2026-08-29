# OnBober

An onboarding compass for complex codebases. Pick a symbol, map its execution flow, expand branches lazily, and get tailored explanations at each step.

## Project Structure

```
IBMHackathon/
├── frontend/     # Vue 3 graph-first UI
└── backend/      # Go API with ripgrep scanner and Ollama JSON graphs
```

## Prerequisites

- **Node.js** 22+ and **npm**
- **Go** 1.22+
- **ripgrep** (`rg`) on PATH
- **git** on PATH (for GitHub clone option)
- **Ollama** running locally with a model pulled (e.g. `ollama pull qwen2.5:7b`)

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Backend API port |
| `OLLAMA_URL` | `http://localhost:11434` | Ollama base URL |
| `OLLAMA_MODEL` | `qwen2.5:7b` | Model name for graph generation and deep dive |
| `ONBOBER_WORKSPACE_ROOT` | `%TEMP%/onbober-workspaces` | Storage for cloned/uploaded repos |

## Development

**→ Full setup and run instructions: [docs/STARTUP.md](docs/STARTUP.md)**

Quick start:

```bash
# One-time
ollama pull qwen2.5:7b
cd frontend && npm install && cd ..

# Terminal 1 — Backend
cd backend
go run ./cmd/api

# Terminal 2 — Frontend
cd frontend
npm run dev
```

Open http://localhost:5173

## Demo Flow

1. Complete onboarding (language, experience, workspace path / GitHub / zip)
2. Open a file from the explorer
3. Click a **symbol chip** in the code drawer
4. View the **execution flow graph** in the center panel
5. Click a **collapsed branch** (`+N`) to expand with a branch budget prompt
6. Click any step to see **node details** in the right drawer

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/tree?workspace=&dir=` | Directory listing |
| GET | `/api/file?workspace=&path=` | File contents |
| POST | `/api/graph/root` | Scan-built execution-flow graph for a symbol |
| POST | `/api/graph/expand` | Lazy-expand collapsed branch/callee nodes (scan-only) |
| POST | `/api/graph/enrich` | Async LLM summary patches for visible nodes |
| GET | `/api/graph/node` | Detailed explanation for a node (LLM) |
| POST | `/api/workspace/setup` | Register local path or clone GitHub repo |
| POST | `/api/workspace/upload` | Extract uploaded zip archive |
| POST | `/api/analyze` | *(deprecated)* Legacy SSE markdown analysis |

## IBM Bob Handoff

- Scan-first graph builder: [`backend/internal/graph/builder.go`](backend/internal/graph/builder.go)
- Hybrid flow extractor: [`backend/internal/scanner/flow/`](backend/internal/scanner/flow/)
- Bounded expansion limits: max 8 root nodes, 6 expand nodes, depth 4

## Notes

- If Ollama is offline, graphs still load instantly from source scans; summaries fall back to code lines.
- Branch expansion requires explicit user confirmation to prevent compute bombs on large codebases.
