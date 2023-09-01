package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCmdQueryEmpty(t *testing.T) {
	fix := startFixture(t)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"query"}, fix.hc)
	})
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"query", "-i", "40000062"}, fix.hc)
	})
	flagQueryInputs = nil // reset flags after parse
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"query", "-q", "filter true"}, fix.hc)
	})
	flagQueryText = "" // reset flags after parse
	fix.Assert()
}

func TestCmdQueryText(t *testing.T) {
	fix := startFixture(t,
		testRequest{`/v1/meta/export/query\?.*`, 200, "timestamp,log\n2023-04-20T16:20:00Z,\"this, is a message!\"\n"},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"query", "-i", "40000062", "-q", ""}, fix.hc)
	flagQueryInputs = nil // reset flags after parse
	flagQueryText = ""    // reset flags after parse
	fix.Assert()
	if diff := cmp.Diff("\n"+fix.op.OutputBuf.String(), `
| timestamp            | log                 |
----------------------------------------------
| 2023-04-20T16:20:00Z | this, is a message! |
`); diff != "" {
		t.Error("unexpected data output:", diff)
	}
}
