package main

import (
	"io/fs"
	"strings"
	"testing"
)

func TestCmdUploadPrompt(t *testing.T) {
	fix := startFixture(t,
		testRequest{`/v1/document/\?name=my\+document\.md`, 404, `{"ok":false,"message":"no such document"}`},
		testRequest{`/v1/document/\?name=my\+document\.md\&usage=prompt`, 200, `{"ok":true,"data":{"meta":{"id":"o::1234:document:80000000022"},"config":{"name":"my document.md","usage":"prompt"},"state":{"createdBy":"1","createdDate":"2023-04-20T16:20:00Z","updatedBy":"1","updatedDate":"2023-04-20T16:20:00Z","url":"/v1/document/download/o::1234:document:80000000022","mimetype":"text/markdown","size":"42"}}}`},
	)
	paniced := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				paniced = true
			}
		}()
		// local file not found
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"upload", "prompt", "my document.md"}, fix.hc)
	}()
	if !paniced {
		t.Fatal("expected fixture to panic; it didn't")
	}
	if !strings.Contains(fix.op.ErrorBuf.String(), "my document.md") {
		t.Error("expected complaint about missing file:", fix.op.ErrorBuf.String())
	}
	if err := fix.fs.WriteFile("my document.md", []byte("# my document\n\ncontains so much goodness!\n"), fs.FileMode(0666)); err != nil {
		t.Fatal("couldn't set up file:", err)
	}
	fix.op = NewCaptureOutput()
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"upload", "prompt", "my document.md"}, fix.hc)
	if strings.Contains(fix.op.ErrorBuf.String(), "my document.md") {
		t.Error("expected no complaint about missing file:", fix.op.ErrorBuf.String())
	}
	if !strings.Contains(fix.op.OutputBuf.String(), "id: o::1234:document:80000000022\n") {
		t.Error("return value didn't contain id:", fix.op.OutputBuf.String())
	}
}
