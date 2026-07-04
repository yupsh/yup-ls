// Command yup-ls is the CLI wrapper around github.com/gloo-foo/cmd-ls.
package main

import (
	clix "github.com/gloo-foo/cli"
	command "github.com/gloo-foo/cmd-ls"
	urf "github.com/urfave/cli/v3"
)

// version is the build version. It defaults to "dev" for local builds and is
// overridden at release time via the linker: -ldflags "-X main.version=<v>".
var version = "dev"

const (
	name          = "ls"
	flagAll       = "all"
	flagRecursive = "recursive"
	flagLong      = "long"
)

// synopsis is the multi-line --help usage block; urfave/cli indents it three
// spaces, so the lines stay flush-left.
const synopsis = `ls [OPTIONS] [FILE]

List information about the FILE (the current directory by default).`

// spec declares the ls wrapper. ls is a source command: it produces its listing
// directly, so build returns it as the whole pipeline (a nil filter).
var spec = clix.Spec{
	Name:     name,
	Summary:  "list directory contents",
	Synopsis: synopsis,
	Build:    build,
	Flags: []urf.Flag{
		&urf.BoolFlag{Name: flagAll, Aliases: []string{"a"}, Usage: "do not ignore entries starting with ."},
		&urf.BoolFlag{Name: flagRecursive, Aliases: []string{"R"}, Usage: "list subdirectories recursively"},
		&urf.BoolFlag{Name: flagLong, Aliases: []string{"l"}, Usage: "use a long listing format"},
	},
}

// build maps the invocation to ls's pipeline: the directory operand and flags
// produce the listing source, with no filter.
func build(inv clix.Invocation) (clix.Source, clix.Command, error) {
	return command.Ls(clix.File(path(inv.Args)), options(inv)...), nil, nil
}

// path is the single directory operand, defaulting to the current directory.
func path(c *urf.Command) string {
	if c.NArg() == 0 {
		return "."
	}
	return c.Args().Get(0)
}

// options folds the parsed flags and the injected filesystem into ls's option
// values.
func options(inv clix.Invocation) []any {
	opts := []any{command.LsFs{Fs: inv.Fs}}
	if inv.Args.Bool(flagAll) {
		opts = append(opts, command.LsAll)
	}
	if inv.Args.Bool(flagRecursive) {
		opts = append(opts, command.LsRecursive)
	}
	if inv.Args.Bool(flagLong) {
		opts = append(opts, command.LsLongFormat)
	}
	return opts
}

// runMain is an indirection seam so main's wiring is testable without spawning
// the process; a test swaps it and restores it.
var runMain = clix.Main

func main() { runMain(spec, version) }
