package main

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/spf13/pflag"
)

var GitCommit string = "(development version)"

var helpText = `
Observe command line tool; ` + GitCommit + `

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
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})
	fmt.Fprintf(out, "\nCommands:\n\n")
	for _, c := range commands {
		fmt.Fprintf(out, "%s\n\n", c.Name)
		if c.Flags != nil {
			c.Flags.PrintDefaults()
		}
		fmt.Fprintf(out, "\n%s\n\n", WrapPrefix(c.Help, "    ", 74))
	}
}
