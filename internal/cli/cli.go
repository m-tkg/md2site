package cli

import (
	"flag"
	"fmt"
	"strings"

	"github.com/m-tkg/md2site/internal/build"
	"github.com/m-tkg/md2site/internal/server"
)

const usage = `Usage:
  md2site build <input-dir> [-o <output-dir>] [--exclude <glob>]... [--title <name>] [--force]
  md2site serve <input-dir> [--port <port>] [--exclude <glob>]... [--title <name>]

Commands:
  build   Generate a static site from markdown files under <input-dir>.
  serve   Build into a temporary directory and serve it over HTTP.

Flags:
  -o string        Output directory (default "./public").
  --exclude glob   Additional exclude pattern, repeatable.
  --title string   Site title. Defaults to the root README's h1, then the directory name.
  --force          Overwrite a non-empty output directory that lacks the .md2site marker.
  --port int       Port for serve (default 8080).
`

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

// parseMixed parses flags allowing a single positional argument to appear
// anywhere among the flags (stdlib flag stops at the first non-flag arg).
func parseMixed(fs *flag.FlagSet, args []string) (string, error) {
	positional := ""
	for {
		if err := fs.Parse(args); err != nil {
			return "", err
		}
		if fs.NArg() == 0 {
			return positional, nil
		}
		if positional != "" {
			return "", fmt.Errorf("unexpected extra argument: %q", fs.Arg(0))
		}
		positional = fs.Arg(0)
		args = fs.Args()[1:]
	}
}

func Run(args []string) error {
	if len(args) == 0 {
		fmt.Print(usage)
		return fmt.Errorf("no command given")
	}
	switch args[0] {
	case "build":
		return runBuild(args[1:])
	case "serve":
		return runServe(args[1:])
	case "-h", "--help", "help":
		fmt.Print(usage)
		return nil
	default:
		fmt.Print(usage)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	out := fs.String("o", "./public", "output directory")
	title := fs.String("title", "", "site title")
	force := fs.Bool("force", false, "overwrite non-empty output directory without marker")
	var excludes stringList
	fs.Var(&excludes, "exclude", "additional exclude glob (repeatable)")
	input, err := parseMixed(fs, args)
	if err != nil {
		return err
	}
	if input == "" {
		return fmt.Errorf("build: input directory required")
	}
	return build.Run(build.Config{
		InputDir:  input,
		OutputDir: *out,
		Title:     *title,
		Excludes:  excludes,
		Force:     *force,
	})
}

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	port := fs.Int("port", 8080, "port to listen on")
	title := fs.String("title", "", "site title")
	var excludes stringList
	fs.Var(&excludes, "exclude", "additional exclude glob (repeatable)")
	input, err := parseMixed(fs, args)
	if err != nil {
		return err
	}
	if input == "" {
		return fmt.Errorf("serve: input directory required")
	}
	return server.BuildAndServe(build.Config{
		InputDir: input,
		Title:    *title,
		Excludes: excludes,
	}, *port)
}
