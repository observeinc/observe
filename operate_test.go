package main

import (
	"fmt"
	"testing"
)

func TestOperate(t *testing.T) {
	op := NewCaptureOutput()
	var q any
	func() {
		defer func() {
			q = recover()
		}()
		*FlagQuietExit = false
		RunRecoverWithTag("action", op, func(o Output) error {
			o.Info("an info\n")
			o.Debug("a debug\n")
			o.Error("an error\n")
			o.Write([]byte("a write"))
			return fmt.Errorf("some error")
		})
	}()
	// the Exit(1) gets caught, and forwarded
	if q.(int) != 1 {
		t.Fatal("expected error exit 1:", q)
	}
	if s := string(op.InfoBuf.Bytes()); s != "action: an info\n" {
		t.Errorf("unexpected Info: %q", s)
	}
	if s := string(op.DebugBuf.Bytes()); s != "starting action\naction: a debug\n" {
		t.Errorf("unexpected Debug: %q", s)
	}
	// note: exit doesn't turn into panic; it's funneled and caught
	if s := string(op.ErrorBuf.Bytes()); s != "action: an error\naction: some error\n" {
		t.Errorf("unexpected Error: %q", s)
	}
	if s := string(op.OutputBuf.Bytes()); s != "a write" {
		t.Errorf("unexpected Write: %q", s)
	}
}
