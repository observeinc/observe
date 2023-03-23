package main

import "fmt"

// The rule for turning errors into ObserveError is:
//
//   - Whenever a system/library function is called, that returns an error,
//     it is immediately turned into an ObserveError with explaining arguments
//     in the message.
//   - Whenever a random string shows up as error (for example, a message in
//     a returned payload), it is immediately turned into an ObserveError.
//   - Functions in this module that return an error, always return ObserveError
//
// The context of "what am I doing" ("logging in" for example) is *not* provided
// in the logging-in code, but instead in whoever calls the logging-in code.
// This means that you do not need to further wrap errors returned from functions
// in this package, unless you need to provide additional context.
//
// We want to avoid error strings with redundant context like:
//
//	login: logging in: login failed: status 403: authorization denied
//
// Instead, we want:
//
//	login: status 403: authorization denied
func NewObserveError(inner error, msg string, args ...any) ObserveError {
	return ObserveError{Msg: fmt.Sprintf(msg, args...), Inner: inner}
}

type ObserveError struct {
	Msg   string
	Inner error
}

func (e ObserveError) Error() string {
	if e.Inner != nil {
		if e.Msg != "" {
			return e.Msg + ": " + e.Inner.Error()
		}
		return e.Inner.Error()
	}
	return e.Msg
}

func (e ObserveError) Unwrap() error {
	return e.Inner
}

var ErrNotImplemented = ObserveError{Msg: "not implemented"}
