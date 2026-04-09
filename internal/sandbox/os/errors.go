// Package os provides OS-level kernel sandbox primitives for tool execution.
// On macOS, it uses Seatbelt (sandbox-exec). Linux isolation is planned but not yet enforced.
package os

import "errors"

var (
	// ErrIsolatorUnavailable indicates the requested OS sandbox is not available on this platform.
	ErrIsolatorUnavailable = errors.New("OS sandbox isolator unavailable")

	// ErrSandboxRequired indicates that sandbox.failClosed is true and no OS sandbox is available.
	ErrSandboxRequired = errors.New("sandbox required but OS isolator unavailable")

	// ErrInvalidPolicy indicates a policy configuration error.
	ErrInvalidPolicy = errors.New("invalid sandbox policy")
)
