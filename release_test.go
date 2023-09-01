package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

//go:embed *.go
var myCode embed.FS

// The observe command line tool is maintained inside the observe monorepo, but
// that means it might accidentally pick up imports from said repo and thus not
// be separately releasable. This test attempts to call out such errors, without
// going to the trouble to parse and load the entire source code using the go
// compiler to introspect it.
func TestNoExternalObserveImport(t *testing.T) {
	dir, err := myCode.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	if len(dir) < 10 {
		t.Fatal("expected more files:", len(dir))
	}
	checked := 0
	for _, de := range dir {
		if !strings.HasSuffix(de.Name(), ".go") {
			continue
		}
		file, err := myCode.Open(de.Name())
		if err != nil {
			t.Fatal(err)
		}
		fi, _ := de.Info()
		size := fi.Size()
		data := make([]byte, size)
		_, err = file.Read(data)
		if err != nil {
			t.Fatal(err)
		}
		lines := bytes.Split(data, []byte{'\n'})
		t.Run(de.Name(), checkFilename(de.Name(), lines))
		checked++
	}
	t.Logf("checked %d files", checked)
}

// A direct import, or a list import on a newline, will trigger this.  There
// are certain packages we might accidentally depend on, perhaps for testing,
// which we really don't want to pull in.
var badLineMatch = regexp.MustCompile(`^(import)?\s*"(observe/|github.com/gorilla/|github.com/99designs/)`)

func checkFilename(name string, lines [][]byte) func(t *testing.T) {
	return func(t *testing.T) {
		for i, l := range lines {
			if badLineMatch.Match(l) {
				fmt.Fprintf(os.Stderr, "%s:%d: Unallowed observe/ import: %s\n", name, i+1, l)
				t.Fail()
			}
		}
	}
}
