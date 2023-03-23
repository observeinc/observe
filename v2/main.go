package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/pflag"
)

func main() {
	ParseFlags()
	if *FlagHelp {
		help()
	}
	var cfg Config
	var op DefaultOutput
	InitConfigFromFileAndFlags(&cfg, &op)
	if *FlagOutput != "" && *FlagOutput != "-" {
		defer SendOutputToFile(*FlagOutput, &op)()
	}
	if *FlagShowConfig {
		m := json.NewEncoder(op)
		m.SetIndent("", "  ")
		m.SetEscapeHTML(false)
		err := m.Encode(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			op.Exit(1)
		}
	}
	RunCommandWithConfig(&cfg, op, pflag.Args(), http.DefaultClient)
}
