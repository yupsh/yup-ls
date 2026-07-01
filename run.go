package main

import (
	"context"
	"fmt"
	"io"

	command "github.com/gloo-foo/cmd-ls"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

const name = "ls"

const (
	flagAll       = "all"
	flagRecursive = "recursive"
	flagLong      = "long"
)

// usageText is the command's multi-line usage synopsis, shown in --help.
// cli/v3 indents the whole block by 3 spaces, so these lines are flush-left to
// stay aligned in the rendered output.
const usageText = `ls [OPTIONS] [FILE]

List information about the FILE (the current directory by default).`

// init replaces urfave/cli's default --version/-v flag with a --version-only
// flag, freeing the single-letter -v for command flags while still exposing
// the injected build version.
func init() {
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Usage: "print version information and exit"}
}

// run builds and executes the ls CLI against the injected version, I/O, and
// filesystem, returning the process exit code. ls does not read stdin; it is
// injected for a uniform, testable wiring shape.
func run(version string, args []string, _ io.Reader, stdout, stderr io.Writer, fs afero.Fs) int {
	cmd := newCommand(version, stdout, fs)
	cmd.Writer = stdout
	cmd.ErrWriter = stderr
	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, name+": %v\n", err)
		return 1
	}
	return 0
}

func newCommand(version string, stdout io.Writer, fs afero.Fs) *cli.Command {
	return &cli.Command{
		Name:            name,
		Version:         version,
		Usage:           "list directory contents",
		UsageText:       usageText,
		HideHelpCommand: true,
		// Keep exit handling in run() rather than letting urfave/cli call
		// os.Exit, so the exit code stays testable.
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: flagAll, Aliases: []string{"a"}, Usage: "do not ignore entries starting with ."},
			&cli.BoolFlag{Name: flagRecursive, Aliases: []string{"R"}, Usage: "list subdirectories recursively"},
			&cli.BoolFlag{Name: flagLong, Aliases: []string{"l"}, Usage: "use a long listing format"},
		},
		Action: action(stdout, fs),
	}
}

func action(stdout io.Writer, fs afero.Fs) cli.ActionFunc {
	return func(_ context.Context, c *cli.Command) error {
		_, err := gloo.Run(command.Ls(path(c), options(c, fs)...), gloo.ByteWriteTo(stdout))
		return err
	}
}

func path(c *cli.Command) string {
	if c.NArg() == 0 {
		return "."
	}
	return c.Args().Get(0)
}

func options(c *cli.Command, fs afero.Fs) []any {
	opts := []any{command.LsFs{Fs: fs}}
	if c.Bool(flagAll) {
		opts = append(opts, command.LsAll)
	}
	if c.Bool(flagRecursive) {
		opts = append(opts, command.LsRecursive)
	}
	if c.Bool(flagLong) {
		opts = append(opts, command.LsLongFormat)
	}
	return opts
}
