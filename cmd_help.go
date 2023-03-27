package main

import (
	"fmt"
	"net/http"

	"github.com/spf13/pflag"
)

var (
	flagHelpObjects bool
	flagsHelp       *pflag.FlagSet
)

func init() {
	flagsHelp = pflag.NewFlagSet("help", pflag.ContinueOnError)
	flagsHelp.BoolVar(&flagHelpObjects, "objects", false, "show help on object types")
	flagsHelp.Lookup("objects").NoOptDefVal = "true"
	RegisterCommand(&Command{
		Name:            "help",
		Help:            "Prints more verbose help about particular commands, or the main program.",
		Func:            cmdHelp,
		Flags:           flagsHelp,
		Unauthenticated: true,
	})
}

func cmdHelp(cfg *Config, op Output, args []string, hc *http.Client) error {
	if flagHelpObjects {
		for _, ot := range GetObjectTypes() {
			writeObjectTypeDocs(op, ot)
		}
		fmt.Fprintf(op, "\n")
		return nil
	}
	if len(args) == 1 {
		readme, err := ReadDocFile("README")
		if err != nil {
			return NewObserveError(err, "missing documentation")
		}
		op.Write(readme)
		return nil
	}
	if len(args) != 2 {
		return NewObserveError(nil, "usage: help [command]")
	}
	cmd := FindCommand(args[1])
	if cmd == nil {
		ot := GetObjectType(args[1])
		if ot != nil {
			writeObjectTypeDocs(op, ot)
			op.Write(newline)
			return nil
		}
		if args[1] != "" && args[1][0] == '-' {
			return NewObserveError(nil, "the help command takes no options")
		}
		return NewObserveError(nil, "there exists no command named %q", args[1])
	}
	op.Write(cmd.Docs)
	return nil
}
