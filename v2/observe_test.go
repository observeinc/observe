package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestFlagsConfig(t *testing.T) {
	ParseFlags()
	*FlagProfile = ""
	*FlagCustomerId = "101"
	*FlagClusterStr = "observe-sandbox.com:4444"
	*FlagAuthtokenStr = "some-authtoken-i-guess"
	*FlagQuiet = true
	*FlagDebug = true
	var op DefaultOutput
	var cfg Config
	InitConfigFromFileAndFlags(&cfg, &op)

	if diff := cmp.Diff(cfg, Config{
		CustomerIdStr: "101",
		ClusterStr:    "observe-sandbox.com:4444",
		AuthtokenStr:  "some-authtoken-i-guess",
		Quiet:         true,
		Debug:         true,
	}); diff != "" {
		t.Fatalf("unexpected difference:\n%s", diff)
	}

	if op.EnableDebug != true || op.DisableInfo != true {
		t.Fatalf("expected EnableDebug and DisableInfo")
	}
}

func TestSetupOutput(t *testing.T) {
	var op DefaultOutput
	var path = fmt.Sprintf("/tmp/test-setup-output-%d", time.Now().UnixNano())
	fn := SendOutputToFile(path, &op)
	if _, err := os.Stat(path + ".tmp"); err != nil {
		t.Fatalf("expected %s.tmp to exist: %s", path, err)
	}
	fn()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %s", path, err)
	}
	if _, err := os.Stat(path + ".tmp"); err == nil {
		t.Fatalf("expected %s.tmp to not exist", path)
	}
	os.Remove(path)
}
