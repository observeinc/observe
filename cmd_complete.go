package main

import (
	"bytes"
	"net/http"
	"os"

	"github.com/posener/complete"
	"github.com/spf13/pflag"
)

var (
	flagsComplete       *pflag.FlagSet
	flagCompleteShell   string
	flagCompleteVerbose bool
)

func init() {
	flagsComplete = pflag.NewFlagSet("complete", pflag.ContinueOnError)
	flagsComplete.StringVar(&flagCompleteShell, "shell", "bash", "which shell to complete for (bash, fish, zsh)")
	flagsComplete.BoolVarP(&flagCompleteVerbose, "verbose", "v", false, "print errors")
	flagsComplete.Lookup("verbose").NoOptDefVal = "true"
	RegisterCommand(&Command{
		Name:            "complete",
		Help:            "Generates command-line completion for the given shell. Sets --quiet-exit.",
		Func:            cmdComplete,
		Flags:           flagsComplete,
		Unauthenticated: true,
		Unlisted:        true,
	})
}

var ErrCompleteUnknownShell = ObserveError{Msg: "the supported shells are 'bash', 'fish', and 'zsh'."}
var ErrCompleteFailed = ObserveError{Msg: "command line completion found no match"}

func cmdComplete(cfg *Config, op Output, args []string, hc *http.Client) error {
	switch flagCompleteShell {
	case "bash":
	case "fish":
	case "zsh":
	default:
		return ErrCompleteUnknownShell
	}
	*FlagQuietExit = true
	if !flagCompleteVerbose {
		op = &DefaultOutput{
			DisableInfo: true,
			DataOutput:  &bytes.Buffer{},
		}
	}

	cmds := complete.Commands{}
	IterateCommands(func(c *Command) {
		if c.Unlisted {
			return
		}
		cm := complete.Command{
			Flags: complete.Flags{},
		}
		if c.Flags != nil {
			c.Flags.VisitAll(completeFlagSetter(cm.Flags))
		}
		cmds[c.Name] = cm
	})
	flags := complete.Flags{}
	pflag.CommandLine.VisitAll(completeFlagSetter(flags))
	cmd := complete.New("observe", complete.Command{
		Sub:   cmds,
		Flags: flags,
	})
	cmd.Out = op

	// This library writes directly to stdout, so I can't really intercept or
	// unit test it.
	if !cmd.Complete() {
		return ErrCompleteFailed
	}
	os.Stdout.Write(op.(*DefaultOutput).DataOutput.(*bytes.Buffer).Bytes())
	return nil
}

func completeFlagSetter(flags complete.Flags) func(*pflag.Flag) {
	return func(pf *pflag.Flag) {
		pred := complete.PredictAnything
		flags["--"+pf.Name] = pred
	}
}
