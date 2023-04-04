package main

import (
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
		Help:            "Prints documentation about particular commands, or the main program.",
		Func:            cmdHelp,
		Flags:           flagsHelp,
		Unauthenticated: true,
	})
}

func cmdHelp(cfg *Config, op Output, args []string, hc *http.Client) error {
	if flagHelpObjects {
	}
	if len(args) == 1 {
		help()
	}
	if len(args) != 2 {
		return NewObserveError(nil, "usage: observe help [command]")
	}
	if args[1] == "observe" {
		return helpFile(op, "README")
	}
	if args[1] == "objects" {
		return helpObjects(op)
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
