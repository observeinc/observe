package main

import (
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

func cmdHelp(fa FuncArgs) error {
	if flagHelpObjects {
	}
	if len(fa.args) == 1 {
		help()
	}
	if len(fa.args) != 2 {
		return NewObserveError(nil, "usage: observe help [command]")
	}
	if fa.args[1] == "observe" {
		return helpFile(fa.op, "README")
	}
	if fa.args[1] == "objects" {
		return helpObjects(fa.op)
	}
	cmd := FindCommand(fa.args[1])
	if cmd == nil {
		ot := GetObjectType(fa.args[1])
		if ot != nil {
			writeObjectTypeDocs(fa.op, ot)
			fa.op.Write(newline)
			return nil
		}
		if fa.args[1] != "" && fa.args[1][0] == '-' {
			return NewObserveError(nil, "the help command takes no options")
		}
		return NewObserveError(nil, "there exists no command named %q", fa.args[1])
	}
	fa.op.Write(cmd.Docs)
	return nil
}
