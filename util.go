package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"golang.org/x/term"
)

var ErrUnrecognizedTime error = ObserveError{Msg: "the time format is not recognized"}
var ErrInvalidDuration error = ObserveError{Msg: "the duration value is not recognized"}
var ErrSnapMustBePositive error = ObserveError{Msg: "the snap duration must be greater than 0"}

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

// Boo hiss reads the value directly -- fixup this by making it part of Config
// and passing Config everywhere it's needed.
func GetConfigFilePath() string {
	if *FlagConfigFile != "" {
		return *FlagConfigFile
	}
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

var timePiecesRegexAlt = regexp.MustCompile(`^(now)? *([+-][0-9]*[smhd])? *(@[0-9]*[smhd])?$`)
var timePiecesRegex = regexp.MustCompile(`^([^-+@][^@]*)?(@[0-9]*[smhd])? *([+-][0-9]*[smhd])?$`)

// The given input strings should be a date-time in YYYY-MM-DD HH:MM:SS format,
// or some variation thereof. Also, epoch values of seconds, milliseconds,
// nanoseconds are allowed, as long as they are sufficiently positive to be
// disabiguated. Finally, relative times are also supported, and will be
// interpreted relative to the 'now' argument.
func ParseTime(tm string, now time.Time) (time.Time, error) {

	tm = strings.TrimSpace(tm)
	if tm == "" {
		return now, nil
	}

	pieces := timePiecesRegexAlt.FindStringSubmatch(tm)
	if len(pieces) == 4 && pieces[0] != "" {
		//	the 'now-3h@1h' form flips the delta and snap
		pieces[2], pieces[3] = pieces[3], pieces[2]
	} else {
		pieces = timePiecesRegex.FindStringSubmatch(tm)
		if len(pieces) != 4 || pieces[0] == "" {
			return time.Time{}, ErrUnrecognizedTime
		}
	}
	// time
	abstime := now
	var err error
	if pieces[1] != "" {
		abstime, err = ReadAbsoluteTime(pieces[1], now)
		if err != nil {
			return time.Time{}, err
		}
	}
	// snap
	if len(pieces[2]) > 0 {
		delta, err := ReadDuration(pieces[2][1:])
		if err != nil {
			return time.Time{}, NewObserveError(err, "time snap")
		}
		if delta <= 0 {
			return time.Time{}, ErrSnapMustBePositive
		}
		abstime = abstime.Truncate(delta)
	}
	// delta
	if len(pieces[3]) > 0 {
		delta, err := ReadDuration(pieces[3])
		if err != nil {
			return time.Time{}, NewObserveError(err, "time offset")
		}
		abstime = abstime.Add(delta)
	}

	return abstime, nil
}

func ReadAbsoluteTime(tm string, rel time.Time) (time.Time, error) {
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
		"2006-01-02 15:04:05.999999999 -07:00",
		"2006-01-02 15:04:05.999 -07:00",
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05.999",
		"2006-01-02 15:04:05",
	} {
		if t, err := time.Parse(timeformat, tm); err == nil {
			return t, nil
		}
	}
	if tm == "now" {
		return rel, nil
	}
	return time.Time{}, ErrUnrecognizedTime
}

func ReadDuration(tm string) (time.Duration, error) {
	if tm == "" {
		return 0, ErrInvalidDuration
	}
	tmlen := len(tm)
	unit := tm[tmlen-1]
	tm = tm[:tmlen-1]
	var i64 int64
	var err error
	switch tm {
	case "-":
		i64 = -1
	case "+", "":
		i64 = 1
	default:
		i64, err = strconv.ParseInt(tm, 10, 64)
		if err != nil {
			return 0, ErrInvalidDuration
		}
	}
	switch unit {
	case 's':
		return time.Duration(i64) * time.Second, nil
	case 'm':
		return time.Duration(i64) * time.Minute, nil
	case 'h':
		return time.Duration(i64) * time.Hour, nil
	case 'd':
		return time.Duration(i64) * 24 * time.Hour, nil
	}
	return 0, ErrInvalidDuration
}

func maybe[T any](t *T) any {
	if t == nil {
		return nil
	}
	return *t
}

func must[T any](t T, e error) T {
	if e != nil {
		panic(fmt.Errorf("unexpected error: %w", e))
	}
	return t
}
