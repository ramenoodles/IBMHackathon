# bridge

Public integration layer for the **grepwrapper** Go module.

## Why this package exists

Jack's implementation lives under `grepwrapper/internal/search` and `grepwrapper/internal/source`. Go's `internal` visibility rules mean **other modules** (like OnBober's `backend`) cannot import those packages directly.

`bridge` re-exports the minimum surface needed by external consumers:

- `Finder` — language-aware symbol declaration search (wraps ripgrep)
- `Reader` — path-safe reads and line context around a match

## Usage (from another module in go.work)

```go
import "grepwrapper/bridge"

finder := bridge.NewFinder("") // "" → use "rg" on PATH
matches, err := finder.Find(ctx, bridge.Query{
    Name:     "start_kernel",
    Root:     "/path/to/linux",
    Language: "c",
    Limit:    20,
})
```

## OnBober

OnBober calls this package from [`backend/internal/scanner/grepwrapper_adapter.go`](../../backend/internal/scanner/grepwrapper_adapter.go). Prefer extending `bridge` over importing `internal/` from outside this module.

## Ownership

- **Jack** owns `internal/search` and `internal/source` behavior.
- **bridge** changes should be small forwarding wrappers so Jack can merge them back into his standalone repo.
- See [docs/GREPWRAPPER_SYNC.md](../../docs/GREPWRAPPER_SYNC.md).
