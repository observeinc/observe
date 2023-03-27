package main

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

type testFixture struct {
	t        *testing.T
	cfg      *Config
	op       *CaptureOutput
	hc       *http.Client
	requests []testRequest
	rix      int
}

type testRequest struct {
	path   string
	status int
	body   string
}

func (t *testFixture) Assert() {
	if t.rix != len(t.requests) {
		t.t.Error("not enough HTTP requests:", t.rix, "!=", len(t.requests))
	}
}

func mustPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		t.Helper()
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	f()
}

func startFixture(t *testing.T, reqs ...testRequest) *testFixture {
	t.Helper()
	tf := &testFixture{t: t, requests: reqs}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tf.rix >= len(tf.requests) {
			t.Error("too many HTTP requests:", tf.rix, r.URL)
			w.WriteHeader(500)
		} else if rex := regexp.MustCompile("^" + tf.requests[tf.rix].path + "$"); !rex.MatchString(r.URL.Path) {
			t.Error("unexpected URL path: got", r.URL.Path, "want", tf.requests[tf.rix].path)
			w.WriteHeader(404)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(tf.requests[tf.rix].status)
			w.Write([]byte(tf.requests[tf.rix].body))
			tf.rix++
		}
	}))
	t.Cleanup(srv.Close)
	tf.cfg = &Config{
		CustomerIdStr: "12345",
		ClusterStr:    strings.Split(srv.URL, "//")[1],
		AuthtokenStr:  "legit-authtoken",
	}
	tf.op = NewCaptureOutput()
	tf.hc = &http.Client{}
	return tf
}
