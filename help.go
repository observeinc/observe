package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

// remember to trim space from this when using it
var GitCommit string = "(development version)"

// remember to trim space from this when using it
var ReleaseName string

var helpText = `
Observe command line tool
` + strings.TrimSpace(ReleaseName) + `
` + strings.TrimSpace(GitCommit) + `

Usage:
  observe [configuration] command [arguments]

Example:
  observe --customerid "1234567890" --cluster "observeinc.com" login "myname@example.com" --read-password --save

Reads configuration from ~/.config/observe.yaml, and command line.
Example observe.yaml file:

profile:
  default:
    customerid: 1234567890
    cluster: observeinc.com
    authtoken: KLJADFSFDSA898987AFAFSA
    debug: true

`

func shorthelp() {
	fmt.Fprintf(os.Stderr, "usage: observe [configuration] command [arguments]\n")
	fmt.Fprintf(os.Stderr, "observe --help for more help\n")
	os.Exit(1)
}

func help() {
	os.Stderr.WriteString(helpText)
	os.Stderr.WriteString("Configuration options:\n\n")
	pflag.PrintDefaults()
	PrintCommands(os.Stderr)
	os.Exit(1)
}

func PrintCommands(out io.Writer) {
	fmt.Fprintf(out, "\nCommands:\n\n")
	IterateCommands(func(c *Command) {
		if c.Unlisted {
			return
		}
		fmt.Fprintf(out, "%s\n\n", c.Name)
		if c.Flags != nil {
			c.Flags.PrintDefaults()
		}
		fmt.Fprintf(out, "\n%s\n\n", WrapPrefix(c.Help, "    ", 74))
	})
}
