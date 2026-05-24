package cliexit

import (
	"errors"
	"fmt"
)

// Error carries a process exit code from a command package to the binary
// entrypoint without letting internal CLI packages terminate the process.
type Error struct {
	Code   int
	Err    error
	silent bool
}

// New returns an error carrying code and a user-facing cause.
func New(code int, err error) error {
	return &Error{Code: code, Err: err}
}

// NewSilent returns an exit-code error that should not be printed by the root
// entrypoint. Use it only after the command has already written user-facing
// output, such as a cancellation message.
func NewSilent(code int) error {
	return &Error{Code: code, silent: true}
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit code %d", e.Code)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Code extracts the carried process exit code from err.
func Code(err error) (int, bool) {
	var exitErr *Error
	if !errors.As(err, &exitErr) || exitErr == nil {
		return 0, false
	}
	return exitErr.Code, true
}

// Silent reports whether the root entrypoint should suppress stderr output for err.
func Silent(err error) bool {
	var exitErr *Error
	if !errors.As(err, &exitErr) || exitErr == nil {
		return false
	}
	return exitErr.silent
}
