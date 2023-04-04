package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCmdGetWorkspace(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"workspace":{"id":"41042069","name":"The Stuff","timezone":"PDT"}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.op, []string{"get", "workspace", "41042069"}, fix.hc)
	if !strings.Contains(fix.op.DebugBuf.String(), "Authorization=Bearer 12345 legit-authtoken") {
		t.Error("unexpected debug output:", fix.op.DebugBuf.String())
	}
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	if diff := fix.op.InfoBuf.String(); diff != "" {
		t.Error("unexpected info output:", diff)
	}
	if diff := cmp.Diff(fix.op.OutputBuf.String(), `object:
  type: "workspace"
  id: 41042069
  config:
    name: "The Stuff"
    timezone: "PDT"
`); diff != "" {
		t.Error("unexpected data output:", diff)
	}
}
