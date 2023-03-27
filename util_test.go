package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWrapPrefix(t *testing.T) {
	for i, tc := range []struct {
		in  string
		out string
	}{
		{
			in:  "asldfj afdlk, jgha fdkll gjhfda lkjghfd. lkjgfh lkaf jdhg adl kdjhfgs dkjgh ds!",
			out: "----asldfj afdlk, \n----jgha fdkll \n----gjhfda lkjghfd.\n---- lkjgfh lkaf \n----jdhg adl \n----kdjhfgs dkjgh \n----ds!\n",
		},
		{
			in:  "snarf\nblarg\n",
			out: "----snarf\n\n----blarg\n\n",
		},
	} {
		q := WrapPrefix(tc.in, "----", 15)
		if diff := cmp.Diff(q, tc.out); diff != "" {
			t.Errorf("unexpected %d:\n%s\n", i, diff)
		}
	}
}

func TestHostname(t *testing.T) {
	hn := GetHostname()
	t.Log(hn)
	if hn == "unknown-host" {
		t.Fatal("why does hostname not work?")
	}
}

func TestTableQuote(t *testing.T) {
	for i, tc := range []struct {
		input  []byte
		output string
	}{
		{[]byte("\n"), `\n`},
		{[]byte(`\`), `\\`},
		{[]byte(`\n`), `\\n`},
		{[]byte(`"`), `"`},
		{nil, ""},
		{[]byte(""), ""},
		{[]byte("hello"), "hello"},
		{[]byte("hello, world"), "hello, world"},
	} {
		if out := TableQuote(tc.input); out != tc.output {
			t.Errorf("test case %d: %q != %q", i, out, tc.output)
		}
	}
}
