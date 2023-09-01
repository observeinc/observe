package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

type testFixture struct {
	t        *testing.T
	cfg      *Config
	op       *CaptureOutput
	hc       httpClient
	fs       fileSystem
	requests []testRequest
	rix      int
	caller   string
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type fakeFs struct {
	dir *map[string][]byte
}

// This fake file system is a thin shim to allow insulation of unit
// tests that read/write files as side effects; it is *not* an exact
// replica of the POSIX or Windows or Mac file system semantics. For
// example, there aren't even any directories!
func NewFakeFs() fileSystem {
	m := map[string][]byte{}
	return fakeFs{
		dir: &m,
	}
}

func (f fakeFs) Stat(path string) (fs.FileInfo, error) {
	_, ok := (*f.dir)[path]
	if !ok {
		return nil, fmt.Errorf("stat %s: no such file or directory", path)
	}
	return nil, nil
}

func (f fakeFs) ReadFile(path string) ([]byte, error) {
	file, ok := (*f.dir)[path]
	if !ok {
		return nil, fmt.Errorf("open %s: no such file or directory", path)
	}
	return file, nil
}

func (f fakeFs) WriteFile(path string, b []byte, perm fs.FileMode) error {
	(*f.dir)[path] = b
	return nil
}

func (f fakeFs) Remove(path string) error {
	if _, has := (*f.dir)[path]; !has {
		return fmt.Errorf("unlink %s: no such file or directory", path)
	}
	delete((*f.dir), path)
	return nil
}

func (f fakeFs) Rename(oldPath, newPath string) error {
	if data, has := (*f.dir)[oldPath]; !has {
		return fmt.Errorf("rename %s: no such file or directory", oldPath)
	} else {
		delete((*f.dir), oldPath)
		// This may nuke something previous -- can't be helped!
		// This is POSIX semantics, but Windows would fail it.
		(*f.dir)[newPath] = data
	}
	return nil
}

func (f fakeFs) MkdirAll(path string, perm fs.FileMode) error {
	// pretend it worked
	return nil
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

func getCaller(n int) string {
	_, f, l, _ := runtime.Caller(n + 1)
	return fmt.Sprintf("%s:%d", f, l)
}

func startFixture(t *testing.T, reqs ...testRequest) *testFixture {
	t.Helper()
	tf := &testFixture{t: t, requests: reqs, caller: getCaller(1)}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tf.rix >= len(tf.requests) {
			fmt.Fprintf(os.Stderr, "%s: too many HTTP request made: %d: %s\n", tf.caller, tf.rix, r.URL)
			t.Fail()
			w.WriteHeader(500)
		} else if rex := regexp.MustCompile(`^` + tf.requests[tf.rix].path + `$`); !rex.MatchString(r.URL.String()) {
			fmt.Fprintf(os.Stderr, "%s: call %d: unexpected URL; got (quoted): %s ; want (regex): %s\n", tf.caller, tf.rix, regexp.QuoteMeta(r.URL.String()), tf.requests[tf.rix].path)
			t.Fail()
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
		SiteStr:       strings.Split(srv.URL, "//")[1],
		AuthtokenStr:  "legit-authtoken",
	}
	tf.op = NewCaptureOutput()
	tf.hc = &http.Client{}
	tf.fs = NewFakeFs()
	return tf
}
