# grepWrapper

`grepwrapper` is a Go wrapper around [ripgrep](https://github.com/BurntSushi/ripgrep) for finding likely symbol implementations in a codebase.

It can return the matching declaration, surrounding source context, or the full matching file. The project is intended to become the retrieval layer for an AI-assisted code understanding tool using IBM watsonx.

## Requirements

* Go 1.22+
* `rg` available on `PATH`

## Run

```sh
go run ./cmd/grepwrapper [options] SYMBOL_NAME
```

Example:

```sh
go run ./cmd/grepwrapper \
  -root ../another-project \
  -lang go \
  ParseConfig
```

Output:

```text
internal/config/parser.go:42: func ParseConfig(path string) (*Config, error) {
```

## Source context

Return surrounding code:

```sh
go run ./cmd/grepwrapper \
  -root ~/Documents/UnboundDepths/src \
  -lang javascript \
  -source context \
  updateUI
```

Control the context size with:

```text
-before N
-after N
```

Return the full matching file with:

```text
-source file
```

## Supported languages

`auto`, `go`, `python`, `javascript`, `typescript`, `rust`, `java`, `c`, `cpp`, `csharp`

Run:

```sh
go run ./cmd/grepwrapper -h
```

for all options.

## Architecture

```text
CLI / HTTP API
      |
      v
   Service
    /    \
 Search  Source
   |
ripgrep
```

The search and source packages are kept separate so they can later be reused by the HTTP API and LLM integration.

## HTTP API

A REST API is being added for programmatic symbol lookup.

Run the server with Watsonx configured through a local `.env` file:

```sh
cp .env.example .env
go run ./cmd/grepwrapper-server -root .
```

The server loads `WATSONX_API_KEY`, `WATSONX_PROJECT_ID`, `WATSONX_MODEL`, `HOST`, and `PORT` from `.env`.
The `-model` flag overrides `WATSONX_MODEL`, and `-addr` overrides `HOST`/`PORT`, when supplied.

Planned endpoint:

```http
POST /v1/symbols/lookup
```

Example request:

```json
{
  "name": "updateUI",
  "language": "javascript",
  "context": true,
  "before": 5,
  "after": 20
}
```

The codebase root is configured by the server rather than supplied by API callers.

## Watsonx Agent

The Watsonx client can iteratively call repository tools before producing an answer:

* `search_symbol` finds likely declarations with ripgrep.
* `read_file` reads a file under the configured repository root.
* `read_context` reads lines around a source location.

Tool execution is limited to eight steps per request and tool output is capped to keep
conversation context bounded. The `-root` and `-rg` server options configure the agent's
repository and ripgrep executable.
Set `INCLUDE_TRAJECTORY=true` or start the server with `-include-trajectory` to include the agent trajectory in
explain responses. The `-include-trajectory` flag can be used as an explicit override; without either,
the trajectory is not returned.

## Current limitations

Search currently uses declaration-shaped regular expressions rather than a full language parser.

This keeps lookup fast, but complex or multiline declarations may be missed. A future syntax-aware extraction stage can be added for exact symbol boundaries.

## Test

```sh
go test ./...
```
