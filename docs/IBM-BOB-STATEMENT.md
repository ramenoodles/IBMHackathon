# IBM Bob Usage Statement

**Team:** Segfault Sorcerers (IBM TechXchange 2026 Pre-conference Dev Day Hackathon)  
**Project:** OnBober — AI-assisted codebase onboarding  
**Demo:** https://onbober-web.fly.dev

---

## How IBM Bob was used to build OnBober

**OnBober** lets users load a repository, browse files, and explore control-flow graphs with progressive reveal, AI node labels, and on-demand explanations. We used **IBM Bob** as a multi-file, agentic coding assistant for tasks that needed exploration across the full stack—not for every small edit.

**Team members who used Bob:** Trent Brown and Jack Manning (documented Bob sessions in [`bob_sessions/`](../bob_sessions/)). Other teammates used complementary tools (e.g. Cursor) for faster iteration on UI polish and feature work.

### Where and how Bob was used

#### Jack Manning — documentation and security hardening

1. **Full repository documentation** — Bob was asked to document the entire codebase and **spawn subagents** for backend, frontend, and Docker/infra in parallel. Output:
   - [`ARCHITECTURE.md`](ARCHITECTURE.md)
   - [`BACKEND.md`](BACKEND.md)
   - [`FRONTEND.md`](FRONTEND.md)
   - [`API.md`](API.md)
   - Expanded [`README.md`](../README.md)

2. **Public-demo backend review** — Bob reviewed the Go API as if deployed on a VPS for ~100 casual users. It identified and implemented security fixes in `backend/internal/workspace/manager.go`, `backend/internal/httpapi/handler.go`, `backend/Dockerfile`, and `.env.example`, including:
   - Disabling `local` filesystem workspace source by default (`ALLOW_LOCAL_SOURCE`)
   - Workspace TTL eviction (`WORKSPACE_MAX_AGE_SECONDS`)
   - `MaxBytesReader` on all POST handlers
   - Non-root Docker user
   - Dead code cleanup in `httpapi`

#### Trent Brown — integration debugging and complex UI/backend features

1. **Watsonx integration debugging** — Bob traced 500 errors on `/api/workspaces/{id}/explain` to credential wiring in `backend/internal/llm/watsonx.go` and `backend/cmd/api/main.go`, fixing explicit `wx.WithWatsonxAPIKey` / `wx.WithWatsonxProjectID` passing and stale-workspace 404 handling in the frontend sidebar.

2. **Explain pipeline extension** — Bob extended the explain request to accept file, line, code, kind, title, and experience context so explanations are grounded in the actual node snippet (not only ripgrep lookup).

3. **Flow canvas UX** — Bob planned and implemented symbol pagination (8 per page, cached symbol graphs), zoom/center viewport fixes in `frontend/src/composables/useFlowPanZoom.ts`, and resolved build breaks (e.g. removed stale `FileFlowBrief` imports).

4. **AI onboarding labels** — Bob helped design context-aware node labeling (verified vs inferred, heuristic vs AI badges) so graphs stay syntactically accurate but more readable for newcomers.

### Why Bob vs other tools

We found **IBM Bob highly effective** for work that required **breadth, planning, and multi-step reasoning** across many files (security audits, full-doc generation, cross-stack bug hunts). We were **initially apprehensive about burning through Bob credits too quickly** because Bob sessions are **token-intensive** (large context, subagents, long transcripts). Our approach:

- Use Bob for **higher-tact tasks** that benefit from agentic exploration and flexible tool use
- Use lighter-weight assistants for **small, localized edits** (styling, copy, single-file fixes)
- Keep **session exports and screenshots** in `bob_sessions/` for auditability

---

## How the project uses IBM watsonx.ai

**We do not use watsonx Orchestrate.** The product integrates **watsonx.ai directly** via the Go SDK (`backend/internal/llm/watsonx.go`), model **`ibm/granite-4-h-small`**, configured with `WATSONX_API_KEY` and `WATSONX_PROJECT_ID`.

| Feature | Endpoint / code | watsonx role |
|--------|------------------|--------------|
| Node titles & summaries | `POST /graphs/enrich` → `backend/internal/service/enrich.go` | Batch Granite calls; falls back to regex heuristics if unavailable |
| Step explanations | `POST /explain` → `backend/internal/service/explain.go` | Agentic loop (up to 8 turns) with tools: `read_file`, `read_context`, `search_symbol` |
| Personalization | User context in enrich/explain | Experience level (junior/mid/senior) and optional language analogies |

Deployed demo: **https://onbober-web.fly.dev** (API on Fly.io with Watsonx secrets; `GET /api/health` reports `watsonx: true` when configured).

---

## Summary

IBM Bob accelerated **documentation, security review, and hard integration/debugging** that would have been slow to do manually under hackathon time pressure. **watsonx.ai is core to the product experience**—not just our dev toolchain—powering the onboarding-focused labels and explanations users see in the flow graph UI.
