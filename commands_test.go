package main

import "testing"

// Documentation files must have enough in them to be useful.
const MinRequiredDocSize = 350

func TestDocs(t *testing.T) {
	for _, cmd := range allCommands {
		if len(cmd.Docs) < MinRequiredDocSize && cmd.Name != "help" && !cmd.Unlisted {
			t.Errorf("Command %q has too little documentation (%d/%d). Write more!", cmd.Name, len(cmd.Docs), MinRequiredDocSize)
		}
	}
}
