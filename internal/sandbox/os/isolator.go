package os

import (
	"context"
	"os/exec"
)

// OSIsolator applies OS-level security restrictions to a subprocess command.
// It modifies exec.Cmd in place before the caller runs it.
// On macOS: wraps with sandbox-exec and a generated Seatbelt profile.
// On Linux: wraps with bubblewrap (bwrap) when the binary is installed
// (BwrapIsolator). The native Landlock+seccomp backend is not yet
// implemented; selecting backend=native or relying on the default noop
// path leaves the command unsandboxed.
type OSIsolator interface {
	// Apply configures the given exec.Cmd to run under OS-level isolation.
	// The command may be wrapped (e.g., sandbox-exec on macOS).
	// Apply does not start the process.
	Apply(ctx context.Context, cmd *exec.Cmd, policy Policy) error

	// Available reports whether this isolator's OS primitives are functional.
	Available() bool

	// Name returns the isolator name (e.g., "seatbelt", "noop").
	Name() string

	// Reason returns a human-readable explanation of why the isolator is
	// unavailable. Returns "" when Available() is true.
	Reason() string
}

// noopIsolator is returned when no OS-level sandbox is available.
type noopIsolator struct {
	reason string
}

func (n *noopIsolator) Apply(_ context.Context, _ *exec.Cmd, _ Policy) error {
	return ErrIsolatorUnavailable
}

func (n *noopIsolator) Available() bool { return false }
func (n *noopIsolator) Name() string    { return "noop" }

func (n *noopIsolator) Reason() string {
	if n.reason != "" {
		return n.reason
	}
	return "no OS isolator available"
}

// disabledIsolator is returned when sandbox is disabled by configuration.
type disabledIsolator struct{}

func (d *disabledIsolator) Apply(_ context.Context, _ *exec.Cmd, _ Policy) error {
	return ErrIsolatorUnavailable
}

func (d *disabledIsolator) Available() bool { return false }
func (d *disabledIsolator) Name() string    { return "disabled" }
func (d *disabledIsolator) Reason() string  { return "sandbox disabled by configuration" }

// NewOSIsolator returns the legacy "default" isolator for the current
// platform. It is kept for backwards compatibility with callers that have
// not yet migrated to the backend registry.
//
// On macOS: SeatbeltIsolator.
// On Linux: noopIsolator pointing to backend=bwrap (the native Landlock+
// seccomp backend is not yet implemented; bwrap is wired via the registry,
// not via this function).
// On unsupported platforms: noopIsolator.
//
// New code should use ParseBackendMode + SelectBackend(mode,
// PlatformBackendCandidates()) instead so that backend selection (auto,
// seatbelt, bwrap, native, none) is honored.
func NewOSIsolator() OSIsolator {
	return newPlatformIsolator()
}
