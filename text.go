package main

import (
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"
)

// The ColumnFormatter parses input data (written to Write()) as CSC data,
// and then when closed, formats it to output in an ASCII table format.
type ColumnFormatter struct {
	Output          io.Writer // where to write to on Close()
	ColWidth        int       // maximum length of any individual column
	ExtendedFormat  bool      // print one column per line, rather than tabular
	LiteralStrings  bool      // print strings literally, without backslash quoting
	OmitLineDrawing bool      // just separate with spaces

	columnNames []string
	columnData  [][]string
	dashStore   []byte
}

type CSVParsingColumnFormatter struct {
	ColumnFormatter
	curValue     []byte   // being parsed
	curRow       []string // being parsed
	inQuote      bool
	inQuoteQuote bool
}

var newline = []byte{'\n'}
var plusminus = []byte{' ', '+', '-'}
var spacepipe = []byte{' ', '|', ' '}
var left = []byte{'|', ' '}
var right = []byte{' ', '|', '\n'}
var space = []byte{' '}

// This implements a state machine that decodes RFC CSV files.
func (c *CSVParsingColumnFormatter) Write(buf []byte) (int, error) {
	curValue := c.curValue
	if curValue == nil {
		curValue = make([]byte, 0, 16384)
	}
	curRow := c.curRow
	if curRow == nil {
		curRow = make([]string, 0, 256)
	}
	inQuote := c.inQuote
	inQuoteQuote := c.inQuoteQuote

	for i, n := 0, len(buf); i != n; i++ {
		if inQuote {
			if buf[i] == '"' {
				if inQuoteQuote {
					curValue = append(curValue, '"')
					inQuoteQuote = false
					continue // stay in quote
				} else {
					inQuoteQuote = true
					continue // stay in quote
				}
			} else if inQuoteQuote {
				inQuoteQuote = false
				inQuote = false
				// fall through to deal with this byte as unquoted
			} else {
				curValue = append(curValue, buf[i])
				continue // stay in quote
			}
		}
		switch buf[i] {
		case '\n', '\r': // handle CR, LF, and CRLF the same
			c.endLine(curValue, curRow)
			curRow = nil
			curValue = curValue[:0]
		case ',':
			curRow = append(curRow, c.makeColString(curValue))
			curValue = curValue[:0]
		case '"':
			inQuote = true
		default:
			curValue = append(curValue, buf[i])
		}
	}

	c.curValue = curValue
	c.curRow = curRow
	c.inQuote = inQuote
	c.inQuoteQuote = inQuoteQuote

	return len(buf), nil
}

func (c *CSVParsingColumnFormatter) endLine(curValue []byte, curRow []string) {
	if len(curValue) > 0 || len(curRow) > 0 {
		curRow = append(curRow, c.makeColString(curValue))
		if c.columnNames == nil {
			c.columnNames = curRow
		} else {
			c.columnData = append(c.columnData, curRow)
		}
	}
}

// Pay attention to ColWidth and UnquoteStrings, returning a string that is
// quoted if needed, and no wider than the limit if needed. Strings that get
// truncated, have a Unicode Ellipsis rune appended as the last character.
func (c *CSVParsingColumnFormatter) makeColString(buf []byte) string {
	if c.ColWidth <= 0 && c.LiteralStrings {
		// no toucha the data, but still, copy it, because we'll re-use it
		return string(buf)
	}
	var str bytes.Buffer
	nrunes := 0
	offset := 0
	needquote := false
	for offset < len(buf) {
		r, sz := utf8.DecodeRune(buf[offset:])
		if sz <= 0 {
			str.WriteRune(rune(buf[offset])) // can't think of anything better to do !?
			offset++
		} else {
			str.WriteRune(r)
			offset += sz
		}
		if r < 32 || r == '\\' {
			needquote = true
		}
		nrunes++
		// Stop when adding one extra rune, which will tell our decoder that we're past the end
		// and thus need to add ellipses.
		if c.ColWidth > 0 && nrunes == c.ColWidth+1 {
			break
		}
	}
	var ret string
	if needquote && !c.LiteralStrings {
		ret = TableQuote(str.Bytes())
	} else {
		ret = str.String()
	}
	if c.ColWidth > 0 && utf8.RuneCountInString(ret) > c.ColWidth {
		rs := []rune(ret)
		rs[c.ColWidth-1] = '…'
		ret = string(rs[:c.ColWidth])
	}
	return ret
}

func (c *CSVParsingColumnFormatter) Close() error {
	c.endLine(c.curValue, c.curRow)
	return c.ColumnFormatter.Close()
}

// Flush the parser, and then format to output
func (c *ColumnFormatter) Close() error {
	// if it didn't end with a newline, flush it anyway
	if c.columnNames == nil {
		return nil
	}
	// now, figure out length of columns
	colLens := make([]int, len(c.columnNames))
	colLens = maxColSize(colLens, c.columnNames)
	// print header
	if c.ExtendedFormat {
		colw := 0
		for _, w := range colLens {
			if w > colw {
				colw = w
			}
		}
		if colw < 9 {
			colw = 9
		}
		for _, row := range c.columnData {
			colLens = maxColSize(colLens, row)
		}
		roww := 0
		for _, w := range colLens {
			if w > roww {
				roww = w
			}
		}
		colwf := fmt.Sprintf("%%-%ds", colw)
		for i, row := range c.columnData {
			c.printExtended(colwf, roww, i, row)
		}
	} else {
		for _, row := range c.columnData {
			colLens = maxColSize(colLens, row)
		}
		fmts := make([]string, len(colLens))
		for i, j := range colLens {
			fmts[i] = fmt.Sprintf("%%-%ds", j)
		}
		c.printRow(fmts, c.columnNames)
		if !c.OmitLineDrawing {
			c.printDashes(colLens)
		}
		// then, print each row
		for _, row := range c.columnData {
			c.printRow(fmts, row)
		}
	}
	return nil
}

func maxColSize(colLens []int, row []string) []int {
	for i := range row {
		l := len(row[i])
		if i >= len(colLens) {
			colLens = append(colLens, l)
		} else {
			if l > colLens[i] {
				colLens[i] = l
			}
		}
	}
	return colLens
}

func (c *ColumnFormatter) SetColumnNames(header []string) {
	c.columnNames = header
}

func (c *ColumnFormatter) AddRow(row []string) {
	if c.ColWidth > 0 {
		trunc := false
		for _, s := range row {
			if len(s) > c.ColWidth {
				trunc = true
				break
			}
		}
		if trunc {
			tmp := make([]string, len(row))
			for i, s := range row {
				if len(s) > c.ColWidth {
					tmp[i] = s[:c.ColWidth-1] + "…"
				} else {
					tmp[i] = s
				}
			}
			row = tmp
		}
	}
	c.columnData = append(c.columnData, row)
}

func (c *ColumnFormatter) dashes(n int) []byte {
	for len(c.dashStore) < n {
		if len(c.dashStore) < 16 {
			c.dashStore = append(c.dashStore, '-', '-', '-', '-', '-', '-', '-', '-')
		} else if len(c.dashStore) < 32000 {
			c.dashStore = append(c.dashStore, c.dashStore...)
		} else {
			// don't waste TOO much extra space when that one value gets really large
			c.dashStore = append(c.dashStore, c.dashStore[:len(c.dashStore)/4]...)
		}
	}
	return c.dashStore[:n]
}

func (c *ColumnFormatter) printExtended(colwf string, roww int, rownum int, row []string) {
	if !c.OmitLineDrawing {
		fmt.Fprintf(c.Output, colwf, fmt.Sprintf("row %d", rownum))
		c.Output.Write(plusminus)
		c.Output.Write(c.dashes(roww))
	}
	c.Output.Write(newline)
	for i, v := range row {
		if i < len(c.columnNames) {
			fmt.Fprintf(c.Output, colwf, c.columnNames[i])
		} else {
			fmt.Fprintf(c.Output, colwf, "")
		}
		if !c.OmitLineDrawing {
			c.Output.Write(spacepipe)
		} else {
			c.Output.Write(space)
		}
		c.Output.Write([]byte(v))
		c.Output.Write(newline)
	}
}

func (c *ColumnFormatter) printRow(fmts []string, row []string) {
	for i := range row {
		if !c.OmitLineDrawing {
			if i == 0 {
				c.Output.Write(left)
			} else {
				c.Output.Write(spacepipe)
			}
		} else if i != 0 {
			c.Output.Write(space)
		}
		fmt.Fprintf(c.Output, fmts[i], row[i])
	}
	if !c.OmitLineDrawing {
		c.Output.Write(right)
	} else {
		c.Output.Write(newline)
	}
}

func (c *ColumnFormatter) printDashes(colLens []int) {
	num := 1
	for _, n := range colLens {
		num += n + 3
	}
	c.Output.Write(c.dashes(num))
	c.Output.Write(newline)
}

func TableQuote(buf []byte) string {
	var out bytes.Buffer
	start := 0
	for ix, ch := range buf {
		var quoted []byte
		switch ch {
		case '\n':
			quoted = []byte{'\\', 'n'}
		case '\r':
			quoted = []byte{'\\', 'r'}
		case '\t':
			quoted = []byte{'\\', 't'}
		case '\\':
			quoted = []byte{'\\', '\\'}
		default:
			if ch < 32 {
				quoted = []byte(fmt.Sprintf("\\x%02x", ch))
			}
		}
		if quoted != nil {
			if start < ix {
				out.Write([]byte(buf[start:ix]))
			}
			out.Write(quoted)
			start = ix + 1
		}
	}
	if start < len(buf) {
		out.Write([]byte(buf[start:]))
	}
	return string(out.Bytes())
}
