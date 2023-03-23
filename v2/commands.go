package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/pflag"
)

type Command struct {
	Name  string
	Help  string
	Flags *pflag.FlagSet
	//	The function will be given the config, the output printer, and remaining arguments,
	//	where the subcommand name is the first string.
	Func            func(cfg *Config, op Output, args []string, hc *http.Client) error
	Unauthenticated bool
	Docs            []byte
}

const MinRequiredDocSize = 750

var commands []*Command

// Typically you call RegisterCommand in an init() function in
// your particular command implementation file.
func RegisterCommand(cmd *Command) {
	// sanity check command
	if strings.ToLower(cmd.Name) != cmd.Name {
		panic(fmt.Sprintf("Commands should have lower-case names! %q is not that!", cmd.Name))
	}
	for _, c := range commands {
		if c.Name == cmd.Name {
			panic(fmt.Sprintf("There cannot be two commands named %q!", cmd.Name))
		}
	}
	// Sanity check flags
	if cmd.Flags != nil {
		cmd.Flags.VisitAll(func(f *pflag.Flag) {
			if f.Shorthand != "" && strings.ToUpper(f.Shorthand) != f.Shorthand {
				panic(fmt.Sprintf("Command %q flag %q shorthand %q must be uppercase!", cmd.Name, f.Name, f.Shorthand))
			}
		})
	}
	cmd.Docs, _ = ReadDocFile(cmd.Name)
	// Remember it for later lookup
	commands = append(commands, cmd)
}

func FindCommand(name string) *Command {
	for _, c := range commands {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// You shouldn't iterate commands during module initialization
// because not all commands may have been initialized yet.
func IterateCommands(fn func(*Command)) {
	for _, cmd := range commands {
		fn(cmd)
	}
}
