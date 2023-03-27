package main

import "testing"

func TestDocs(t *testing.T) {
	for _, cmd := range gCommands {
		if len(cmd.Docs) < MinRequiredDocSize && cmd.Name != "help" && !cmd.Unlisted {
			t.Errorf("Command %q has too little documentation (%d/%d). Write more!", cmd.Name, len(cmd.Docs), MinRequiredDocSize)
		}
	}
}
