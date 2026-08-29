# grepWrapper sync guide (for Jack)

OnBober imports Jack's [`grepWrapper`](https://github.com/ramenoodles/IBMHackathon/tree/grepWrapper) branch as a **git subtree** under `grepwrapper/`. That directory mirrors your repo root layout (`cmd/`, `internal/`, `go.mod`, etc.).

**Not imported:** `grepWrapper-agentic` (Watsonx agentic work stays on your branch).

## Monorepo layout

```
IBMHackathon/
├── backend/          # OnBober API — adapter in internal/scanner/grepwrapper_adapter.go
├── frontend/         # OnBober UI
└── grepwrapper/      # Your module (module grepwrapper)
    ├── bridge/       # Public API for OnBober (see below)
    ├── cmd/grepwrapper/
    ├── internal/search/
    └── ...
```

## Public bridge package

Go does not allow other modules to import `grepwrapper/internal/*`. OnBober uses [`grepwrapper/bridge`](grepwrapper/bridge/bridge.go) — a thin public wrapper around your `internal/search` and `internal/source` packages. **Prefer adding external APIs here** rather than having OnBober import `internal/` directly.

## Rules for OnBober contributors

- **Do not edit** `grepwrapper/internal/*` for OnBober features — keep changes in `backend/internal/scanner/` or discuss with Jack first.
- **OK to extend** `grepwrapper/bridge/` when OnBober needs new exported APIs.
- Jack's CLI and tests remain runnable from `grepwrapper/` unchanged.

## Jack → monorepo (pull your latest grepWrapper)

From repo root, on a branch that should receive your updates:

```powershell
cd IBMHackathon
git fetch origin
git subtree pull --prefix=grepwrapper origin grepWrapper --squash -m "Sync grepwrapper subtree from origin/grepWrapper"
```

Resolve conflicts only under `grepwrapper/` if needed. Run:

```powershell
cd grepwrapper && go test ./...
cd ../backend && go test ./...
```

## Monorepo → Jack (push OnBober-side grepwrapper changes back)

If we only changed `grepwrapper/` (e.g. `bridge/`):

1. Extract subtree commits, or
2. Copy changed files from `IBMHackathon/grepwrapper/` into your standalone repo and commit there, or
3. Add OnBober's fork as a remote and cherry-pick the subtree squash commit.

**Recommended:** Jack owns `grepWrapper`; OnBober pulls via `git subtree pull`. OnBober-specific glue stays in `backend/internal/scanner/grepwrapper_adapter.go`.

## Verify after sync

```powershell
cd grepwrapper
go test ./...
go run ./cmd/grepwrapper -root .. -lang go -h

cd ../backend
go test ./...
```

## Subtree reference

Initial import commit: squash merge from `origin/grepWrapper` (`a689e4c` area) onto `feat/integrate-grepwrapper`.
