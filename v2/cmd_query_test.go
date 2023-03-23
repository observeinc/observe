package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCmdQueryEmpty(t *testing.T) {
	fix := startFixture(t)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.op, []string{"query"}, fix.hc)
	})
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.op, []string{"query", "-I", "40000062"}, fix.hc)
	})
	flagInput = nil // reset flags after parse
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.op, []string{"query", "-Q", "filter true"}, fix.hc)
	})
	flagText = "" // reset flags after parse
	fix.Assert()
}

func TestCmdQueryText(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta/export/query", 200, "timestamp,log\n2023-04-20T16:20:00Z,\"this, is a message!\"\n"},
	)
	RunCommandWithConfig(fix.cfg, fix.op, []string{"query", "-I", "40000062", "-Q", ""}, fix.hc)
	flagInput = nil // reset flags after parse
	flagText = ""   // reset flags after parse
	fix.Assert()
	if diff := cmp.Diff("\n"+fix.op.OutputBuf.String(), `
| timestamp            | log                 |
----------------------------------------------
| 2023-04-20T16:20:00Z | this, is a message! |
`); diff != "" {
		t.Error("unexpected data output:", diff)
	}
}
