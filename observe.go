/*
Options for the 'observe' command are arranged so that global options
may have lowercase single-letter aliases, and options for a subcommand
may have uppercase single-letter aliases. Thus, "-o" always means
"set output file" but "-O" could mean different things for different
sub-commands. If you want to avoid confusion, use the long options.
*/
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

var FlagProfile = pflag.StringP("profile", "P", "default", "The name of a section in the ~/.config/observe.yaml config file. Make empty to read no profile. Can also be specified in environment OBSERVE_PROFILE")
var FlagCustomerId = pflag.StringP("customerid", "C", "", "The numeric ID of your Observe tenant.")
var FlagClusterStr = pflag.StringP("cluster", "U", "", "The domain of your Observe tenant cluster. Can include :port if needed.")
var FlagAuthtokenStr = pflag.StringP("authtoken", "A", "", "The bearer token for Observe authorization. May be two-part with a space separator. Does not include 'Bearer' word.")
var FlagOutput = pflag.StringP("output", "O", "", "The output file name for data output. If empty or '-', output goes to stdout.")
var FlagQuiet = pflag.BoolP("quiet", "Q", false, "Don't output info logs.")
var FlagDebug = pflag.BoolP("debug", "D", false, "Output extra debug logs.")
var FlagLog = pflag.BoolP("timestamp", "T", false, "Timestamp output to make better log files.")
var FlagHelp = pflag.BoolP("help", "h", false, "Print help.")
var FlagShowConfig = pflag.BoolP("show-config", "", false, "Print configuration before running command.")
var FlagConfigFile = pflag.String("config", "", "Read configuration from given file rather than ~/config/observe.yaml. Can also be specified in environment OBSERVE_CONFIG.")
var FlagWorkspace = pflag.String("workspace", "", "Default workspace to assume for objects if none is specified.")

var flagsParsed = false

// ParseFlags will parse the global flags if they haven't already been parsed.
// Great to call from main(), and can also be called in tests (although whoever
// calls it first, will cause parsing.) ParseFlags() will also verify some
// sanity properties about flag names/shortcuts.
func ParseFlags() {
	if !flagsParsed {
		// Sanity check global flags
		pflag.VisitAll(func(f *pflag.Flag) {
			// help is extra special
			if f.Shorthand != "" && f.Shorthand != "h" && strings.ToUpper(f.Shorthand) != f.Shorthand {
				panic(fmt.Sprintf("Global flag %q shorthand %q must be uppercase!", f.Name, f.Shorthand))
			}
		})
		IterateCommands(func(cmd *Command) {
			if cmd.Flags != nil {
				cmd.Flags.VisitAll(func(f *pflag.Flag) {
					if pflag.CommandLine.Lookup(f.Name) != nil {
						panic(fmt.Sprintf("Command %q flag %q clashes with global flag of same name!", cmd.Name, f.Name))
					}
				})
			}
		})
		flagsParsed = true
		pflag.Lookup("help").NoOptDefVal = "true"
		pflag.Lookup("quiet").NoOptDefVal = "true"
		pflag.Lookup("debug").NoOptDefVal = "true"
		pflag.Lookup("timestamp").NoOptDefVal = "true"
		pflag.SetInterspersed(false)
		pflag.Parse()
		envProfile := os.Getenv("OBSERVE_PROFILE")
		if !pflag.Lookup("profile").Changed && envProfile != "" {
			*FlagProfile = envProfile
		}
		configFile := os.Getenv("OBSERVE_CONFIG")
		if !pflag.Lookup("config").Changed && configFile != "" {
			*FlagConfigFile = configFile
		}
	}
}

func InitConfigFromFileAndFlags(cfg *Config, op *DefaultOutput) {
	if *FlagProfile != "" {
		RunRecoverWithTag("read config", op, func(Output) error {
			return ReadConfig(cfg, GetConfigFilePath(), *FlagProfile, false)
		})
	}
	if *FlagCustomerId != "" {
		cfg.CustomerIdStr = *FlagCustomerId
	}
	if *FlagClusterStr != "" {
		cfg.ClusterStr = *FlagClusterStr
	}
	if *FlagAuthtokenStr != "" {
		cfg.AuthtokenStr = *FlagAuthtokenStr
	}
	if pflag.Lookup("quiet").Changed || *FlagQuiet {
		cfg.Quiet = *FlagQuiet
	}
	if pflag.Lookup("debug").Changed || *FlagDebug {
		cfg.Debug = *FlagDebug
	}
	if pflag.Lookup("workspace").Changed || *FlagWorkspace != "" {
		cfg.Workspace = *FlagWorkspace
	}
	*op = DefaultOutput{EnableDebug: cfg.Debug, DisableInfo: cfg.Quiet, DataOutput: os.Stdout}
}

func SendOutputToFile(path string, op *DefaultOutput) func() {
	var out *os.File
	var err error
	RunRecoverWithTag("output file", op, func(Output) error {
		out, err = os.Create(path + ".tmp")
		if err != nil {
			return err
		}
		op.DataOutput = out
		return nil
	})
	return func() {
		out.Close()
		os.Remove(path)
		RunRecoverWithTag("finish output", op, func(Output) error { return os.Rename(path+".tmp", path) })
	}
}

// RunCommandWithConfig is convenient to use from unit tests.
// You can provide the appropriate config and output strings, as
// well as the command line arguments (starting with the sub-command
// name) and a HTTP client to use when network access is needed.
// (Typically your command will in turn pass that to DoRequest())
// Because this uses RunRecoverWithTag, it will call output.Exit()
// if there is an error; if you use CaptureOutput, this will turn
// into a panic.
func RunCommandWithConfig(cfg *Config, op Output, args []string, hc *http.Client) {
	if len(args) > 0 && (args[0] == "-" || args[0] == "--") {
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "\nobserve: a command is required\n\n")
		help()
	}
	cmd := FindCommand(args[0])
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "\nobserve: there is no command named %q\n\n", args[0])
		help()
	}
	var errors []string
	if !cmd.Unauthenticated {
		if cfg.CustomerIdStr == "" {
			errors = append(errors, "customerid")
		}
		if cfg.ClusterStr == "" {
			errors = append(errors, "cluster")
		}
		if cfg.AuthtokenStr == "" {
			// the login command doesn't need an authtoken
			errors = append(errors, "authtoken")
		}
	}
	if len(errors) > 0 {
		os.Stderr.WriteString("\nobserve: missing required configuration:\n")
		os.Stderr.WriteString(strings.Join(errors, ", "))
		os.Stderr.WriteString("\n")
		help()
	}
	RunRecoverWithTag(cmd.Name, op, func(o Output) error {
		if cmd.Flags != nil {
			if err := cmd.Flags.Parse(args[1:]); err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s\n", cmd.Name, err)
				fmt.Fprintf(os.Stderr, "%s", WrapPrefix(cmd.Help, "   ", 75))
				if !strings.Contains(err.Error(), "pflag: help requested") {
					cmd.Flags.PrintDefaults()
				}
				os.Exit(1)
			}
			args = append([]string{args[0]}, cmd.Flags.Args()...)
		}
		return cmd.Func(cfg, o, args, hc)
	})
}
