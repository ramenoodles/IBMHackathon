// Package cli contains the command-line interface for grepwrapper.
package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"grepwrapper/internal/search"
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
	flags.Usage = func() {
		fmt.Fprintln(stderr, usage)
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
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

	for _, match := range matches {
		fmt.Fprintf(stdout, "%s:%d: %s\n", match.Path, match.Line, strings.TrimSpace(match.Text))
	}
	return 0
}
