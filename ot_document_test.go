package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestObjectTypeDocument(t *testing.T) {
	fix := startFixture(t,
		testRequest{`/v1/document/`, 200, `{"ok":true,"data":[]}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"list", "document", "doc"}, fix.hc)
	if !strings.Contains(fix.op.DebugBuf.String(), "Authorization=Bearer 12345 legit-authtoken") {
		t.Error("unexpected debug output:", fix.op.DebugBuf.String())
	}
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	if diff := fix.op.InfoBuf.String(); diff != "" {
		t.Error("unexpected info output:", diff)
	}
	if diff := cmp.Diff(fix.op.OutputBuf.String(), "id usage name\n"); diff != "" {
		t.Error("unexpected data output:", diff)
	}
}
