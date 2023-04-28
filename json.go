package main

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONFormatter is a table formatter that prints nd-json lines; if extended
// format, records are splayed across one line per field and wrapped in a
// top-level array with commas (actual JSON); if normal, one line per record
// without commas (ND-JSON).
type JSONFormatter struct {
	Output         io.Writer
	ExtendedFormat bool
	headers        []string
	started        bool
	comma          bool
	enc            *json.Encoder
}

func (j *JSONFormatter) SetColumnNames(headers []string) {
	j.headers = headers
}

func (j *JSONFormatter) AddRow(row []string) {
	rl := len(row)
	if rl != len(j.headers) {
		panic(fmt.Sprintf("internal error: headers has %d fields, AddRow() has %d fields", len(j.headers), len(row)))
	}
	if !j.started {
		j.started = true
		j.start()
	}
	if j.ExtendedFormat {
		// I want careful control over formatting/field ordering of the extended format,
		// so manually generate it. Because it's just strings, and we already quoted them,
		// it's still syntactically safe.
		if j.comma {
			j.Output.Write(jsonExtendedComma)
		} else {
			j.comma = true
		}
		j.Output.Write(jsonExtendedRecordStart)
		for i, h := range j.headers {
			j.Output.Write(jsonExtendedRecordFieldStart)
			j.enc.Encode(h)
			j.Output.Write(jsonExtendedRecordValueStart)
			j.enc.Encode(row[i])
			if i+1 != rl {
				j.Output.Write(jsonExtendedRecordValueEnd)
			}
		}
		j.Output.Write(jsonExtendedRecordEnd)
	} else {
		// build a map, jam the data into it, let the encoder sort it out
		m := map[string]string{}
		for i, h := range j.headers {
			m[h] = row[i]
		}
		j.enc.Encode(m)
	}
}

func (j *JSONFormatter) Close() error {
	j.stop()
	return nil
}

var jsonExtendedStart = []byte{'[', '\n'}
var jsonExtendedRecordStart = []byte{'{', '\n'}
var jsonExtendedRecordFieldStart = []byte{' ', ' '}
var jsonExtendedRecordValueStart = []byte{':', ' '}
var jsonExtendedRecordValueEnd = []byte{',', '\n'}
var jsonExtendedRecordEnd = []byte{'\n', '}'}
var jsonExtendedComma = []byte{',', '\n'}
var jsonExtendedEnd = []byte{'\n', ']', '\n'}

// Strip newlined from the end of writing, to adapt the JSON encoder which puts
// them there after each thing it writes.
type writerWithoutNewline struct {
	w io.Writer
}

func (w writerWithoutNewline) Write(b []byte) (int, error) {
	l := len(b)
	if l > 0 && b[l-1] == '\n' {
		if l == 1 {
			return 1, nil
		}
		i, j := w.w.Write(b[:l-1])
		if j != nil {
			return i, j
		}
		return i + 1, nil
	}
	return w.w.Write(b)
}

func (j *JSONFormatter) start() {
	if j.ExtendedFormat {
		j.Output.Write(jsonExtendedStart)
		j.enc = json.NewEncoder(writerWithoutNewline{j.Output})
	} else {
		j.enc = json.NewEncoder(j.Output)
	}
}

func (j *JSONFormatter) stop() {
	if j.ExtendedFormat {
		j.Output.Write(jsonExtendedEnd)
	}
}
