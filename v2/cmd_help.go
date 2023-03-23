package main

import "net/http"

func init() {
	RegisterCommand(&Command{
		Name:            "help",
		Help:            "Prints more verbose help about particular commands, or the main program.",
		Func:            cmdHelp,
		Unauthenticated: true,
	})
}

func cmdHelp(cfg *Config, op Output, args []string, hc *http.Client) error {
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
		if args[1] != "" && args[1][0] == '-' {
			return NewObserveError(nil, "the help command takes no options")
		}
		return NewObserveError(nil, "there exists no command named %q", args[1])
	}
	op.Write(cmd.Docs)
	return nil
}
