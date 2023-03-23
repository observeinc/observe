package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"golang.org/x/term"
)

func WrapPrefix(str string, leading string, width int) string {
	var ret string
	last := 0
	breaknow := false
	for i := 0; i < len(str); i++ {
		if i == width || breaknow {
			if last == 0 || last < width-12 {
				last = i
			}
			ret = ret + leading + str[:last] + "\n"
			str = str[last:]
			i = 0
			breaknow = false
		}
		ch := str[i]
		switch {
		case ch >= 'a' && ch <= 'z':
		case ch >= 'A' && ch <= 'Z':
		case ch >= '0' && ch <= '9':
		case ch == '_':
		case ch == '\n':
			breaknow = true
		default:
			last = i + 1
		}
	}
	if len(str) > 0 {
		ret = ret + leading + str + "\n"
	}
	return ret
}

func GetConfigFilePath() string {
	return path.Join(os.Getenv("HOME"), ".config/observe.yaml")
}

func ReadPasswordFromTerminal(prompt string) ([]byte, error) {
	var pwdata []byte
	var err error
	if term.IsTerminal(int(syscall.Stdin)) {
		os.Stderr.WriteString(prompt)
		pwdata, err = term.ReadPassword(int(syscall.Stdin))
		fmt.Fprintf(os.Stderr, "\n")
	} else {
		pwdata, err = ioutil.ReadAll(os.Stdin)
		// I can't quite Trim, because someone might have a password that begins or ends with a space.
		for len(pwdata) > 0 && (pwdata[len(pwdata)-1] == '\n' || pwdata[len(pwdata)-1] == '\r') {
			pwdata = pwdata[:len(pwdata)-1]
		}
	}
	return pwdata, err
}

func GetHostname() string {
	hn, _ := os.Hostname() // should just work
	if hn == "" {
		hn = os.Getenv("HOSTNAME") // Unix fallback
	}
	if hn == "" {
		hn = os.Getenv("COMPUTERNAME") // Windows fallback
	}
	if hn == "" {
		hn = "unknown-host"
	}
	return hn
}

func CountFlags(fs *pflag.FlagSet, flags ...string) int {
	n := 0
	for _, f := range flags {
		if fs.Lookup(f).Changed {
			n++
		}
	}
	return n
}

func LoadQueryTextFromFile(filepath string) (string, error) {
	buf, err := os.ReadFile(filepath)
	return string(buf), err
}

// The given input strings should be a date-time in YYYY-MM-DD HH:MM:SS format,
// or some variation thereof. Also, epoch values of seconds, milliseconds,
// nanoseconds are allowed, as long as they are sufficiently positive to be
// disabiguated.
func ParseTime(tm string) (time.Time, error) {
	// If I have an epoch time, from UNIX, from JavaScript, or from OPAL, I can
	// punch it in here as a number.
	if i64, err := strconv.ParseInt(tm, 10, 64); err == nil {
		if i64 >= 1000000000 && i64 < 4999999999 { // seconds
			return time.Unix(i64, 0), nil
		} else if i64 >= 1000000000000 && i64 < 9999999999999 { // milliseconds
			return time.Unix(i64/1000, i64%1000), nil
		} else if i64 >= 1000000000000000000 { // nanoseconds
			return time.Unix(0, i64), nil
		}
	}
	// The German, Canadian, British, and US ways of writing dates
	// are too ambiguous and contradictory, so we *only* support
	// ISO-style YYYY-MM-DD 24-hour formats.
	for _, timeformat := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05.999",
		"2006-01-02 15:04:05",
	} {
		if t, err := time.Parse(timeformat, tm); err == nil {
			return t, nil
		}
	}
	if tm == "now" {
		return time.Now(), nil
	}
	return time.Time{}, NewObserveError(nil, "input should be a date/time in YYYY-MM-DD HH:MM:SS format")
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
