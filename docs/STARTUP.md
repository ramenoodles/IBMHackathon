# OnBober — Startup Guide

This guide walks you through running OnBober locally from a fresh clone.

## What you are starting

OnBober has three moving parts:

| Service | Port | Purpose |
|---------|------|---------|
| **Ollama** | `11434` | Local LLM for graph labels, enrich summaries, and deep-dive explanations |
| **Backend (Go API)** | `8080` | Scanner, flow graphs, workspace management |
| **Frontend (Vite + Vue)** | `5173` | UI (proxies `/api` → backend in dev) |

You need all three for full AI features. If Ollama is offline, scan-based flowcharts still work; AI labels and deep dive fall back to demo/placeholder text.

---

## 1. Prerequisites

Install these before first run:

| Tool | Version | Check | Notes |
|------|---------|-------|-------|
| **Node.js** | 22+ | `node -v` | Includes npm |
| **Go** | 1.22+ | `go version` | Backend runtime |
| **ripgrep** | any | `rg --version` | Required for codebase search |
| **git** | any | `git --version` | Only needed for GitHub clone onboarding |
| **Ollama** | latest | `ollama --version` | [ollama.com](https://ollama.com) |

### Windows (PowerShell)

```powershell
# ripgrep (if not installed)
winget install BurntSushi.ripgrep.MSVC

# Ollama — download installer from https://ollama.com/download
```

### macOS

```bash
brew install go node ripgrep
brew install ollama
```

### Linux

```bash
# Debian/Ubuntu example
sudo apt install golang ripgrep git
curl -fsSL https://ollama.com/install.sh | sh
```

---

## 2. One-time setup

From the repo root (`IBMHackathon/`):

### Pull the default LLM model

```bash
ollama pull qwen2.5:7b
```

This matches the backend default (`OLLAMA_MODEL=qwen2.5:7b`). Use a different model by setting `OLLAMA_MODEL` when starting the backend (see below).

### Install frontend dependencies

```bash
cd frontend
npm install
cd ..
```

`npm install` also copies VS Code material icons into `public/icons/` via the postinstall script.

### Backend dependencies

Go modules download automatically on first `go run`. No separate install step.

---

## 3. Start the app (every session)

Open **three terminals** (or run Ollama as a background service).

### Terminal 1 — Ollama

Ollama usually runs as a system service after install. Confirm it is up:

```bash
ollama list
```

If the command fails, start the Ollama app (Windows/macOS) or:

```bash
ollama serve
```

### Terminal 2 — Backend API

```bash
cd backend
go run ./cmd/api
```

You should see:

```
OnBober API listening on :8080
```

**Optional environment variables** (PowerShell examples):

```powershell
$env:PORT = "8080"
$env:OLLAMA_URL = "http://localhost:11434"
$env:OLLAMA_MODEL = "qwen2.5:7b"
$env:ONBOBER_WORKSPACE_ROOT = "$env:TEMP\onbober-workspaces"
go run ./cmd/api
```

### Terminal 3 — Frontend dev server

```bash
cd frontend
npm run dev
```

Vite prints a local URL, typically:

```
http://localhost:5173
```

Open that URL in your browser.

---

## 4. Verify everything works

### Backend health

```bash
curl http://localhost:8080/api/health
```

Expected: `{"status":"ok"}`

### Ollama + model

```bash
ollama list
```

Confirm `qwen2.5:7b` (or your chosen model) appears in the list.

### End-to-end in the UI

1. Complete onboarding (language, experience, workspace).
2. **Local folder** — point at an existing repo on disk.
3. **GitHub** — paste a repo URL (requires git).
4. **Zip upload** — upload an archive.
5. Open a file → click a **symbol chip** → view the flow graph.
6. Click a step → **Explain this step** for a deep-dive explanation.

---

## 5. Environment variables reference

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Backend listen port |
| `OLLAMA_URL` | `http://localhost:11434` | Ollama API base URL |
| `OLLAMA_MODEL` | `qwen2.5:7b` | Model for enrich + deep dive |
| `ONBOBER_WORKSPACE_ROOT` | OS temp dir + `/onbober-workspaces` | Where cloned/uploaded repos are stored |

The frontend dev server proxies `/api/*` to `http://localhost:8080` — you do not need to configure a separate frontend API URL in development.

---

## 6. Production-style build (optional)

To build the frontend for static hosting (still needs the Go API running separately):

```bash
cd frontend
npm run build
npm run preview   # serves dist/ locally for smoke testing
```

The Go API is started the same way: `go run ./cmd/api` from `backend/`.

---

## 7. grepwrapper CLI (optional)

OnBober's symbol search uses Jack's [`grepwrapper`](../grepwrapper/) module (language-aware ripgrep). You can run the CLI directly for debugging:

```powershell
cd grepwrapper
go run ./cmd/grepwrapper -root C:\path\to\repo -lang c start_kernel
go run ./cmd/grepwrapper -root C:\path\to\repo -lang python -source context -before 3 -after 5 my_function
```

Supported languages: `auto`, `go`, `python`, `javascript`, `typescript`, `rust`, `java`, `c`, `cpp`, `csharp`.

See [GREPWRAPPER_SYNC.md](GREPWRAPPER_SYNC.md) for subtree sync with Jack's `grepWrapper` branch.

---

## 8. Running tests

```bash
# grepwrapper (Jack's module)
cd grepwrapper
go test ./...

# Backend (uses grepwrapper via go.work)
cd backend
go test ./...

# Frontend
cd frontend
npm run test
npm run build    # type-check + production build
```

---

## 9. Troubleshooting

### `rg` not found / scan errors

Install ripgrep and ensure `rg` is on your `PATH`. Restart the terminal after installing.

### Backend starts but AI labels say "Demo" / deep dive is placeholder

- Confirm Ollama is running: `ollama list`
- Confirm the model is pulled: `ollama pull qwen2.5:7b`
- Restart the backend after changing `OLLAMA_MODEL`

### Frontend shows API errors / network failed

- Backend must be running on port `8080`
- Start frontend with `npm run dev` (not `npm run preview` unless you have also configured API routing for preview)
- Check nothing else is bound to port `8080`

### Port already in use

```powershell
# Windows — find process on 8080
netstat -ano | findstr :8080
```

Change the backend port: `$env:PORT = "8081"` and update `frontend/vite.config.ts` proxy target to match.

### Slow first deep-dive response

The first request after starting Ollama loads the model into memory. Subsequent explanations are faster. A 7B model needs roughly 5–8 GB RAM/VRAM.

### Workspace path not found (local folder onboarding)

Use an absolute path to the repo root. On Windows, prefer forward slashes or escaped backslashes, e.g. `C:/Users/you/projects/my-repo`.

---

## 10. Quick reference — copy/paste

```bash
# One-time
ollama pull qwen2.5:7b
cd frontend && npm install && cd ..

# Every session (two terminals after Ollama is running)
cd backend && go run ./cmd/api
cd frontend && npm run dev
```

Then open **http://localhost:5173**.

---

## Related docs

- [README](../README.md) — project overview, API list, architecture notes
- [CODEBASE.md](CODEBASE.md) — full repo tour for new contributors
- [GREPWRAPPER.md](GREPWRAPPER.md) — symbol search architecture
- [GREPWRAPPER_SYNC.md](GREPWRAPPER_SYNC.md) — sync Jack's grepWrapper subtree
- IBM Bob handoff notes in README → `backend/internal/graph/builder.go`
