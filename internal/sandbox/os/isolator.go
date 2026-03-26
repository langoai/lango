package os

import (
	"context"
	"os/exec"
)

// OSIsolator applies OS-level security restrictions to a subprocess command.
// It modifies exec.Cmd in place before the caller runs it.
// On macOS: wraps with sandbox-exec and a generated Seatbelt profile.
// On Linux: injects self-restriction args for Landlock + seccomp.
type OSIsolator interface {
	// Apply configures the given exec.Cmd to run under OS-level isolation.
	// The command may be wrapped (e.g., sandbox-exec on macOS).
	// Apply does not start the process.
	Apply(ctx context.Context, cmd *exec.Cmd, policy Policy) error

	// Available reports whether this isolator's OS primitives are functional.
	Available() bool

	// Name returns the isolator name (e.g., "seatbelt", "landlock+seccomp").
	Name() string
}

// noopIsolator is returned when no OS-level sandbox is available.
type noopIsolator struct{}

func (n *noopIsolator) Apply(_ context.Context, _ *exec.Cmd, _ Policy) error {
	return ErrIsolatorUnavailable
}

func (n *noopIsolator) Available() bool { return false }
func (n *noopIsolator) Name() string    { return "noop" }

// NewOSIsolator returns the best available OS isolator for the current platform.
// On macOS: SeatbeltIsolator. On Linux: CompositeIsolator(Landlock + seccomp).
// On unsupported platforms: noopIsolator (Available() returns false).
func NewOSIsolator() OSIsolator {
	return newPlatformIsolator()
}
