# grepWrapper

`grepwrapper` is a small Go command-line wrapper around
[ripgrep](https://github.com/BurntSushi/ripgrep). Given a function name, it
prints likely implementation locations in a codebase.

This is the first building block for a larger code-onboarding tool. The search
package is separate from the CLI so later stages can reuse the results to find
call sites, inspect parameters, and generate explanations.

## Requirements

- Go 1.22 or newer
- `rg` (ripgrep) available on `PATH`

## Build

```sh
go build -o grepwrapper ./cmd/grepwrapper
```

## Use

Search the current codebase and infer the supported source language from file
extensions:

```sh
./grepwrapper ParseConfig
```

Search another codebase and restrict matches to Go declarations:

```sh
./grepwrapper -root ../another-project -lang go ParseConfig
```

Example output:

```text
internal/config/parser.go:42: func ParseConfig(path string) (*Config, error) {
```

Run `./grepwrapper -h` for all options. Supported language filters are `auto`,
`go`, `python`, `javascript`, `typescript`, `rust`, `java`, `c`, `cpp`, and
`csharp`.

## Current scope

The tool uses declaration-shaped regular expressions, not a language parser.
That keeps the first module fast and dependency-free, but complex or multiline
declarations can be missed and C-family patterns can occasionally produce false
positives. A later module can rank or verify these candidates with a syntax
tree.

## Test

```sh
go test ./...
```
