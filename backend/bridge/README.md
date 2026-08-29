# bridge

Public integration layer for the **grepwrapper** Go module.

## Why this package exists

The implementation lives under `backend/internal/search` and `backend/internal/source`.

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


## Ownership

- **Jack** owns `internal/search` and `internal/source` behavior.
- **bridge** changes should be small forwarding wrappers so Jack can merge them back into his standalone repo.
- See [docs/GREPWRAPPER_SYNC.md](../../docs/GREPWRAPPER_SYNC.md).
