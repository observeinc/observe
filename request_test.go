package main

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHttpStatusError(t *testing.T) {
	op := NewCaptureOutput()
	url := "http://example.com/path/"
	for i, tc := range []struct {
		code int
		json string
		want string
	}{
		{404, `{"ok":false,"message":"no such document"}`, "no such document"},
		{500, `{"ok":false,"errors":["stack trace:","/file.go:123: error here"]}`, "file.go:123"},
		{400, `{"ok":false,"error":"invalid request"}`, "invalid request"},
		{503, `<!doctype HTML><h1>Server Not Available</h1>`, "HTTP error: 503"},
	} {
		hresp := &http.Response{
			StatusCode: tc.code,
			Status:     "xxxx",
			Body:       io.NopCloser(strings.NewReader(tc.json)),
		}
		e := HttpStatusError(op, url, hresp)
		if e == nil {
			t.Fatalf("case %d: expected error", i)
		}
		if !strings.Contains(e.Error(), tc.want) {
			t.Errorf("case %d: unexpected error: %s", i, e)
		}
		if !strings.Contains(op.DebugBuf.String(), tc.json) {
			t.Errorf("case %d: expected debug output: %s", i, op.DebugBuf.String())
		}
	}
}

func TestQueryGetList(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/things/", 200, `{"ok":true}`},
		testRequest{"/v1/things/", 200, `{"ok":true,"data":{"currentUser":{"workspaces":[{"id":"41042069","name":"The Stuff","timezone":"PDT"}]}}}`},
		testRequest{"/v1/things/", 200, `{"ok":true,"data":[{"id":"x"}]}`},
		testRequest{"/v1/things/", 200, `{"ok":true,"data":[{"xid":"x"}]}`},
	)
	var err error
	_, err = Query(fix.hc).Config(fix.cfg).Output(fix.op).Path("/v1/things/").PropMap(PropertyMap{"id": mkpath("xid")}).GetList()
	if err == nil || !errors.Is(err, ErrNotAnArray) {
		t.Fatal("expected ErrNotAnArray:", err)
	}
	_, err = Query(fix.hc).Config(fix.cfg).Output(fix.op).Path("/v1/things/").PropMap(PropertyMap{"id": mkpath("xid")}).GetList()
	if err == nil || !errors.Is(err, ErrNotAnArray) {
		t.Fatal("expected ErrNotAnArray:", err)
	}
	_, err = Query(fix.hc).Config(fix.cfg).Output(fix.op).Path("/v1/things/").PropMap(PropertyMap{"id": mkpath("xid")}).GetList()
	if err == nil || !strings.Contains(err.Error(), "item 0") {
		t.Fatal("expected item 0 error:", err)
	}
	var a array
	a, err = Query(fix.hc).Config(fix.cfg).Output(fix.op).Path("/v1/things/").PropMap(PropertyMap{"id": mkpath("xid")}).GetList()
	if err != nil || len(a) != 1 {
		t.Fatal("expected one item:", err, a)
	}
}

func TestQueryGet(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/things/", 200, `{"ok":true}`},
		testRequest{"/v1/things/", 200, `{"ok":true,"data":[{"id":"x"}]}`},
		testRequest{"/v1/things/", 200, `{"ok":true,"data":{"xid":"xx"}}`},
	)
	var err error
	_, err = Query(fix.hc).Config(fix.cfg).Output(fix.op).Path("/v1/things/").PropMap(PropertyMap{"id": mkpath("xid")}).Get()
	if err == nil || !errors.Is(err, ErrNotAnObject) {
		t.Fatal("expected ErrNotAnObject:", err)
	}
	_, err = Query(fix.hc).Config(fix.cfg).Output(fix.op).Path("/v1/things/").PropMap(PropertyMap{"id": mkpath("xid")}).Get()
	if err == nil || !errors.Is(err, ErrNotAnObject) {
		t.Fatal("expected ErrNotAnObject:", err)
	}
	var o object
	o, err = Query(fix.hc).Config(fix.cfg).Output(fix.op).Path("/v1/things/").PropMap(PropertyMap{"id": mkpath("xid")}).Get()
	if err != nil || o["id"].(string) != "xx" {
		t.Fatal("expected one item:", err, o)
	}
}
