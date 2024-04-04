package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCmdRbacDot(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"user":{"id":"4","name":"Test User","email":"reader@observeinc.com","status":"UserStatusActive","role":"reader"}}}`},
		testRequest{"/v1/meta", 200, `{"data":{"rbacGroups":[
            {"id":"o::101:rbacgroup:8000001001","name":"reader","description":"for reading"},
            {"id":"o::101:rbacgroup:8000001002","name":"writer","description":"for writing"},
            {"id":"o::101:rbacgroup:8000001003","name":"admin","description":"for adminning"},
            {"id":"o::101:rbacgroup:8000001004","name":"oxygen breather","description":"for breathing oxygen"}
        ]}}`},
		testRequest{"/v1/meta", 200, `{"data":{"rbacGroupmembers":[
            {"id":"o::101:rbacgroupmember:8000001005","description":"mem1","groupid":"o::101:rbacgroup:8000001001","membergroupid":null,"memberuserid":"3"},
            {"id":"o::101:rbacgroupmember:8000001006","description":"mem2","groupid":"o::101:rbacgroup:8000001001","membergroupid":null,"memberuserid":"4"},
            {"id":"o::101:rbacgroupmember:8000001007","description":"mem3","groupid":"o::101:rbacgroup:8000001002","membergroupid":null,"memberuserid":"1"},
            {"id":"o::101:rbacgroupmember:8000001008","description":"mem4","groupid":"o::101:rbacgroup:8000001003","membergroupid":null,"memberuserid":"5"},
            {"id":"o::101:rbacgroupmember:8000001009","description":"mem5","groupid":"o::101:rbacgroup:8000001004","membergroupid":"o::101:rbacgroup:8000001001","memberuserid":null},
            {"id":"o::101:rbacgroupmember:8000001010","description":"mem6","groupid":"o::101:rbacgroup:8000001004","membergroupid":"o::101:rbacgroup:8000001002","memberuserid":null},
            {"id":"o::101:rbacgroupmember:8000001011","description":"mem7","groupid":"o::101:rbacgroup:8000001004","membergroupid":"o::101:rbacgroup:8000001003","memberuserid":null}
        ]}}`},
	)
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "panic: %v\n", r)
			t.Error("panic on exit")
			fmt.Fprintf(os.Stderr, "debug:\n%s\n", fix.op.DebugBuf.String())
			fmt.Fprintf(os.Stderr, "info:\n%s\n", fix.op.InfoBuf.String())
			fmt.Fprintf(os.Stderr, "error:\n%s\n", fix.op.ErrorBuf.String())
			fmt.Fprintf(os.Stderr, "output:\n%s\n", fix.op.OutputBuf.String())
		}
	}()
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"rbac-dot", "--user", "4"}, fix.hc)
	if !strings.Contains(fix.op.DebugBuf.String(), "Authorization=Bearer 12345 legit-authtoken") {
		t.Error("unexpected debug output:", fix.op.DebugBuf.String())
	}
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	if diff := fix.op.InfoBuf.String(); diff != "" {
		t.Error("unexpected info output:", diff)
	}
	if diff := cmp.Diff(fix.op.OutputBuf.String(), `digraph {
  node [shape=box];
  rankdir=LR;
  ranksep=1.5;
  "4" [label="Test User"];
  "4" -> "o::101:rbacgroup:8000001001";
  "o::101:rbacgroup:8000001001" [label="reader"];
  "o::101:rbacgroup:8000001001" -> "o::101:rbacgroup:8000001004";
  "o::101:rbacgroup:8000001004" [label="oxygen breather"];
}
`); diff != "" {
		t.Error("unexpected data output:", diff)
	}
}
