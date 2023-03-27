package main

import (
	"embed"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/pflag"
)

// remember to trim space from this when using it
var (
	GitCommit      string = "(devel)"
	GitModified    string = ""
	ReleaseVersion string = "(devel)"
)

var helpText = `
Observe command line tool
` + strings.TrimSpace(ReleaseVersion) + `
` + strings.TrimSpace(GitCommit) + strings.TrimSpace(GitModified) + `

Usage:
  observe [configuration] command [arguments]

Example:
  observe --customerid "1234567890" --cluster "observeinc.com" login "myname@example.com" --read-password --save

Reads configuration from ~/.config/observe.yaml, and command line.

`

func init() {
	bi, _ := debug.ReadBuildInfo()
	for _, s := range bi.Settings {
		if s.Key == "vcs.revision" {
			GitCommit = s.Value
		}
		if s.Key == "vcs.modified" {
			if s.Value == "true" {
				GitModified = "-modified"
			}
		}
	}
	ReleaseVersion = bi.Main.Version
}

func help() {
	os.Stderr.WriteString(helpText())
	os.Stderr.WriteString("Configuration options:\n\n")
	pflag.PrintDefaults()
	PrintCommands(os.Stderr)
	os.Stderr.WriteString("\nUse 'observe help observe' for more help and 'observe help objects' for object types.\n")
	os.Exit(1)
}

func PrintCommands(out io.Writer) {
	fmt.Fprintf(out, "\nCommands:\n\n")
	maxl := 0
	IterateCommands(func(cmd *Command) {
		if len(cmd.Name) > maxl {
			maxl = len(cmd.Name)
		}
	})
	IterateCommands(func(cmd *Command) {
		fmt.Fprintf(out, "  observe %-[1]*[2]s  %[3]s\n", maxl, cmd.Name, cmd.Help)
	})
	fmt.Fprintf(out, "\nUse 'observe help <command>' for more command help.\n")
}

//go:embed *.md
//go:embed docs/*.md
var docFS embed.FS

func ReadDocFile(name string) ([]byte, error) {
	ret, err := docFS.ReadFile(name + ".md")
	if err != nil {
		ret, err = docFS.ReadFile("docs/" + name + ".md")
	}
	return ret, err
}

func helpFile(op Output, name string) error {
	readme, err := ReadDocFile(name)
	if err != nil {
		return NewObserveError(err, "missing documentation for %q", name)
	}
	op.Write(readme)
	return nil
}

func helpObjects(op Output) error {
	fmt.Fprintf(op, "object types (use 'observe help <objecttype>' for specifics):\n\n")
	maxlen := 0
	for _, ot := range GetObjectTypes() {
		n := len(ot.TypeName())
		if n > maxlen {
			maxlen = n
		}
	}
	for _, ot := range GetObjectTypes() {
		fmt.Fprintf(op, "  %-[1]*[2]s  %[3]s\n", maxlen, ot.TypeName(), ot.Help())
	}
	fmt.Fprintf(op, "\n")
	return nil
}