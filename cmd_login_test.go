package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCmdLoginUserPassword(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/login", 200, `{"ok":true,"message":"success","access_key":"totally-legit-access-token"}`},
	)
	// this will panic if there's an error
	RunCommandWithConfig(fix.cfg, fix.op, []string{"login", "me@example.com", "hunter123"}, fix.hc)
	if !strings.Contains(fix.op.DebugBuf.String(), "User-Agent=observe") {
		t.Error("unexpected debug output:", fix.op.DebugBuf.String())
	}
	if diff := cmp.Diff(fix.op.ErrorBuf.String(), ``); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	if diff := cmp.Diff(fix.op.InfoBuf.String(), "login: saved authtoken to section \"default\" in config file \"/home/dev/.config/observe.yaml\"\n"); diff != "" {
		t.Error("unexpected info output:", diff)
	}
	if diff := cmp.Diff(fix.op.OutputBuf.String(), "totally-legit-access-token\n"); diff != "" {
		t.Error("unexpected data output:", diff)
	}
}

func TestCmdLoginDelegated(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/login/delegated", 200, `{"ok":true,"message":"success","url":"http://12345.observe-sandbox.com:4444/settings/account?serverToken=some-server-token","serverToken":"some-server-token"}`},
		testRequest{"/v1/login/delegated/some-server-token", 200, `{"ok":true,"settled":false}`},
		testRequest{"/v1/login/delegated/some-server-token", 200, `{"ok":true,"settled":true,"accessKey":"totally-legit-access-token"}`},
	)
	// this will panic if there's an error
	RunCommandWithConfig(fix.cfg, fix.op, []string{"login", "me@example.com", "--sso"}, fix.hc)
	flagLoginSSO = false // reset options after running
	fix.Assert()
	if !strings.Contains(fix.op.DebugBuf.String(), "/v1/login/delegated") {
		t.Error("unexpected debug output:", fix.op.DebugBuf.String())
	}
	if diff := cmp.Diff(fix.op.ErrorBuf.String(), ``); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	if diff := cmp.Diff(fix.op.InfoBuf.String(), "login: saved authtoken to section \"default\" in config file \"/home/dev/.config/observe.yaml\"\n"); diff != "" {
		t.Error("unexpected info output:", diff)
	}
	if diff := cmp.Diff(fix.op.OutputBuf.String(), "totally-legit-access-token\n"); diff != "" {
		t.Error("unexpected data output:", diff)
	}
}

func TestLoginMissingPassword(t *testing.T) {
	fix := startFixture(t)
	// this will panic if there's an error
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.op, []string{"login", "me@example.com"}, fix.hc)
	})
	fix.Assert()
	if !strings.Contains(fix.op.DebugBuf.String(), "starting login") {
		t.Error("unexpected debug output:", fix.op.DebugBuf.String())
	}
	if !strings.Contains(fix.op.InfoBuf.String(), "no password provided") {
		t.Error("unexpected info output:", fix.op.DebugBuf.String())
	}
	if !strings.Contains(fix.op.ErrorBuf.String(), "email address and password") {
		t.Error("unexpected error output:", fix.op.ErrorBuf.String())
	}
}
