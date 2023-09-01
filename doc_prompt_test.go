package main

import "testing"

func TestDocPrompt(t *testing.T) {
	prompt := &docKindPrompt{}
	for i, tc := range []struct {
		input  string
		result string
		err    error
	}{
		{"some text file\nwith just text\n", "text/plain", nil},
		{"# some markdown file\n\nwith markdown text.\n", "text/markdown", nil},
		{"some text file without newline should give an error", "", ErrFileHasNoNewlines},
		{"hello, world!\n\x00", "", ErrFileIsBinary},
		{"hello, \x01world!\x02\n", "", ErrFileIsBinary},
		{string([]byte{'h', 'e', 'l', 'l', 'o', 0xC0, ' ', 'w', 'o', 'r', 'l', 'd', '\n'}), "", ErrInvalidUtf8},
		// not going to test TooLarge
	} {
		if mt, err := prompt.SniffMimetype([]byte(tc.input)); mt != tc.result || err != tc.err {
			t.Errorf("case %d: unexpected compare: %s != %s or %v != %v", i, mt, tc.result, err, tc.err)
		}
	}
}
