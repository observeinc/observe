package main

import (
	"bytes"
	"unicode/utf8"
)

var ErrInvalidUtf8 = ObserveError{Msg: "the document is not valid utf-8 format"}
var ErrFileHasNoNewlines = ObserveError{Msg: "the document has no newlines"}
var ErrFileIsBinary = ObserveError{Msg: "the document contains binary data"}
var ErrFileTooLarge = ObserveError{Msg: "the document is too large"}

const MaxPromptFileSize = 100000

type docKindPrompt struct {
}

func init() {
	addDocumenKind(&docKindPrompt{})
}

func (d *docKindPrompt) Kind() string {
	return "prompt"
}

func isBOM(data []byte) bool {
	// 0xEF, 0xBB, 0xBF
	return len(data) >= 3 && data[0] == 0xef && data[1] == 0xbb && data[2] == 0xbf
}

func sniffMarkdown(data []byte) bool {
	return bytes.HasPrefix(data, []byte{'#', ' '}) ||
		bytes.Contains(data, []byte{'\n', '#', ' '}) ||
		bytes.Contains(data, []byte{'\n', '#', '#', ' '}) ||
		bytes.Contains(data, []byte{'\n', '`', '`', '`'}) ||
		bytes.Contains(data, []byte{'\n', '~', '~', '~'})
}

// The point of this function is to prevent obvious failure cases, not to be
// normative or authoritative. Thus, err slightly on the side of
// permissiveness.
func (d *docKindPrompt) SniffMimetype(data []byte) (string, error) {
	if len(data) > MaxPromptFileSize {
		return "", ErrFileTooLarge
	}
	var gotNl, gotBad, gotZero int
	if !utf8.Valid(data) {
		return "", ErrInvalidUtf8
	}
	if isBOM(data) {
		data = data[3:]
	}
	for _, c := range data {
		if c < 32 {
			switch c {
			case 10:
				gotNl++
			case 8, 9, 12, 13:
			case 0:
				gotZero++
			default:
				gotBad++
			}
		}
	}
	// we allow a very small number of "bad" low ASCII characters that aren't
	// printable
	if gotBad > 1+len(data)/1000 || gotZero > 0 {
		return "", ErrFileIsBinary
	}
	if gotNl == 0 {
		return "", ErrFileHasNoNewlines
	}
	if sniffMarkdown(data) {
		return "text/markdown", nil
	}
	return "text/plain", nil
}
