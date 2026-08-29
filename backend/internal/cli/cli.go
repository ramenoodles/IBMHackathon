// Package cli contains the command-line interface for grepwrapper.
package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/ramenoodles/IBMHackathon/backend/internal/search"
	"github.com/ramenoodles/IBMHackathon/backend/internal/source"
)

const usage = `Usage:
  grepwrapper [options] FUNCTION_NAME

Find likely implementation locations for a function by asking ripgrep to search
for declaration patterns. Paths are relative to the selected codebase root.

Options:`

// Run executes the command and returns a process exit code.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("grepwrapper", flag.ContinueOnError)
	flags.SetOutput(stderr)

	root := flags.String("root", ".", "codebase directory to search")
	language := flags.String("lang", "auto", "language: auto, go, python, javascript, typescript, rust, java, c, cpp, or csharp")
	limit := flags.Int("max", 20, "maximum number of results to print")
	rgBinary := flags.String("rg", "rg", "path to the ripgrep executable")
	sourceMode := flags.String(
		"source",
		"match",
		"source output: match, context, or file",
	)

	before := flags.Int(
		"before",
		5,
		"number of lines before a match in context mode",
	)

	after := flags.Int(
		"after",
		20,
		"number of lines after a match in context mode",
	)

	flags.Usage = func() {
		fmt.Fprintln(stderr, usage)
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		return 2
	}

	switch *sourceMode {
	case "match", "context", "file":
	default:
		fmt.Fprintln(stderr, "grepwrapper: -source must be match, context, or file")
		return 2
	}

	if *before < 0 {
		fmt.Fprintln(stderr, "grepwrapper: -before must not be negative")
		return 2
	}

	if *after < 0 {
		fmt.Fprintln(stderr, "grepwrapper: -after must not be negative")
		return 2
	}

	if flags.NArg() != 1 {
		flags.Usage()
		return 2
	}
	if *limit < 1 {
		fmt.Fprintln(stderr, "grepwrapper: -max must be at least 1")
		return 2
	}

	name := flags.Arg(0)
	finder := search.NewFinder(*rgBinary)
	matches, err := finder.Find(ctx, search.Query{
		Name:     name,
		Root:     *root,
		Language: *language,
		Limit:    *limit,
	})
	if err != nil {
		fmt.Fprintf(stderr, "grepwrapper: %v\n", err)
		return 2
	}
	if len(matches) == 0 {
		fmt.Fprintf(stderr, "No likely implementation found for %q under %s.\n", name, *root)
		return 1
	}

	if *sourceMode == "match" {
		for _, match := range matches {
			fmt.Fprintf(
				stdout,
				"%s:%d: %s\n",
				match.Path,
				match.Line,
				strings.TrimSpace(match.Text),
			)
		}
		return 0
	}

	reader, err := source.NewReader(*root)
	if err != nil {
		fmt.Fprintf(stderr, "grepwrapper: %v\n", err)
		return 2
	}

	switch *sourceMode {
	case "context":
		for _, match := range matches {
			snippet, err := reader.ReadContext(
				match.Path,
				match.Line,
				*before,
				*after,
			)
			if err != nil {
				fmt.Fprintf(
					stderr,
					"grepwrapper: read %s: %v\n",
					match.Path,
					err,
				)
				return 2
			}

			fmt.Fprintf(
				stdout,
				"=== %s:%d-%d ===\n",
				snippet.Path,
				snippet.StartLine,
				snippet.EndLine,
			)

			fmt.Fprintln(stdout, snippet.Content)
		}

	case "file":
		seen := make(map[string]struct{})

		for _, match := range matches {
			// Avoid printing the same file several times when multiple
			// declarations matched inside it
			if _, exists := seen[match.Path]; exists {
				continue
			}
			seen[match.Path] = struct{}{}

			content, err := reader.ReadFile(match.Path)
			if err != nil {
				fmt.Fprintf(
					stderr,
					"grepwrapper: read %s: %v\n",
					match.Path,
					err,
				)
				return 2
			}

			fmt.Fprintf(stdout, "=== %s ===\n", match.Path)
			fmt.Fprint(stdout, content)

			if !strings.HasSuffix(content, "\n") {
				fmt.Fprintln(stdout)
			}
		}
	}

	return 0
}
