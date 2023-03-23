package main

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMakeColString(t *testing.T) {
	cf := ColumnFormatter{
		ColWidth:       4,
		UnquoteStrings: false,
	}
	for i, tc := range []struct {
		input  []byte
		output string
	}{
		{[]byte("hello"), "hel…"},
		{[]byte(""), ""},
		{[]byte("………"), "………"},
		{[]byte("………x"), "………x"},
		{[]byte("………xx"), "…………"},
	} {
		if out := cf.makeColString(tc.input); out != tc.output {
			t.Errorf("test case %d: %q != %q", i, out, tc.output)
		}
	}
}

var csvData = []byte(`a,bb,ccc
123,23,3
three,two,one
\,",","
"
`)

func TestColumnFormatter1(t *testing.T) {
	cf := ColumnFormatter{Output: &bytes.Buffer{}}
	cf.Write(csvData)
	cf.Close()
	str := cf.Output.(*bytes.Buffer).String()
	if diff := cmp.Diff(str, `| a     | bb  | ccc |
---------------------
| 123   | 23  | 3   |
| three | two | one |
| \\    | ,   | \n  |
`); diff != "" {
		t.Fatalf("unexpected difference:\n%s", diff)
	}
}

var csvData2 = []byte(`field1,"field number two",3
one,two,three
"one,""","""two""",3
,\,
some longer line that has more characters,makes up this value,"with some
quotes involved"
`)

func TestColumnFormatter2(t *testing.T) {
	cf := ColumnFormatter{Output: &bytes.Buffer{}}
	cf.Write(csvData2)
	cf.Close()
	str := cf.Output.(*bytes.Buffer).String()
	if diff := cmp.Diff(str, `| field1                                    | field number two    | 3                          |
------------------------------------------------------------------------------------------------
| one                                       | two                 | three                      |
| one,"                                     | "two"               | 3                          |
|                                           | \\                  |                            |
| some longer line that has more characters | makes up this value | with some\nquotes involved |
`); diff != "" {
		t.Fatalf("unexpected difference:\n%s", diff)
	}
}
