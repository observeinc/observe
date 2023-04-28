package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCmdListWorkspace(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"currentUser":{"workspaces":[{"id":"41042069","name":"The Stuff","timezone":"PDT"}]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.op, []string{"list", "workspace"}, fix.hc)
	if !strings.Contains(fix.op.DebugBuf.String(), "Authorization=Bearer 12345 legit-authtoken") {
		t.Error("unexpected debug output:", fix.op.DebugBuf.String())
	}
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	if diff := fix.op.InfoBuf.String(); diff != "" {
		t.Error("unexpected info output:", diff)
	}
	if diff := cmp.Diff(fix.op.OutputBuf.String(), `id       name     `+"\n"+`41042069 The Stuff`+"\n"); diff != "" {
		t.Error("unexpected data output:", diff)
	}
}
