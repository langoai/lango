package os

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// BackendMode identifies a sandbox backend by type.
type BackendMode int

const (
	// BackendAuto selects the first available backend automatically.
	BackendAuto BackendMode = iota
	// BackendSeatbelt selects the macOS Seatbelt (sandbox-exec) backend.
	BackendSeatbelt
	// BackendBwrap selects the Linux bubblewrap (bwrap) backend.
	BackendBwrap
	// BackendNative selects the native kernel sandbox (landlock+seccomp) backend.
	BackendNative
	// BackendNone disables all sandbox backends explicitly.
	BackendNone
)

// String returns a human-readable label for the backend mode.
func (m BackendMode) String() string {
	switch m {
	case BackendAuto:
		return "auto"
	case BackendSeatbelt:
		return "seatbelt"
	case BackendBwrap:
		return "bwrap"
	case BackendNative:
		return "native"
	case BackendNone:
		return "none"
	default:
		return fmt.Sprintf("BackendMode(%d)", int(m))
	}
}

// BackendCandidate pairs a backend mode with its isolator implementation.
type BackendCandidate struct {
	Mode     BackendMode
	Isolator OSIsolator
}

// BackendInfo describes the status of a single sandbox backend.
type BackendInfo struct {
	Name      string
	Mode      BackendMode
	Available bool
	Reason    string
}

// ParseBackendMode converts a string to a BackendMode.
// An empty string maps to BackendAuto.
func ParseBackendMode(s string) (BackendMode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "auto":
		return BackendAuto, nil
	case "seatbelt":
		return BackendSeatbelt, nil
	case "bwrap":
		return BackendBwrap, nil
	case "native":
		return BackendNative, nil
	case "none":
		return BackendNone, nil
	default:
		return 0, fmt.Errorf("unknown sandbox backend: %q", s)
	}
}

// SelectBackend picks the appropriate isolator from candidates based on the requested mode.
//
// BackendAuto: returns the first available candidate. If none available, returns a noopIsolator.
// BackendNone: always returns a noopIsolator.
// Explicit mode (seatbelt/bwrap/native): returns the matching candidate's isolator as-is,
// even if unavailable. If the mode is not found in candidates, returns a noopIsolator.
func SelectBackend(mode BackendMode, candidates []BackendCandidate) (OSIsolator, BackendInfo) {
	switch mode {
	case BackendNone:
		iso := &noopIsolator{reason: "backend explicitly set to none"}
		return iso, infoFrom(iso, BackendNone)

	case BackendAuto:
		for _, c := range candidates {
			if c.Isolator.Available() {
				return c.Isolator, infoFrom(c.Isolator, c.Mode)
			}
		}
		iso := &noopIsolator{reason: aggregateUnavailableReasons(candidates)}
		return iso, infoFrom(iso, BackendAuto)

	default:
		for _, c := range candidates {
			if c.Mode == mode {
				return c.Isolator, infoFrom(c.Isolator, c.Mode)
			}
		}
		reason := fmt.Sprintf("backend %s not available on this platform", mode)
		iso := &noopIsolator{reason: reason}
		return iso, infoFrom(iso, mode)
	}
}

// ListBackends returns status information for each candidate backend.
func ListBackends(candidates []BackendCandidate) []BackendInfo {
	infos := make([]BackendInfo, 0, len(candidates))
	for _, c := range candidates {
		infos = append(infos, infoFrom(c.Isolator, c.Mode))
	}
	return infos
}

// aggregateUnavailableReasons concatenates each candidate's Reason() so that
// the auto-mode noop fallback retains actionable diagnostic information.
func aggregateUnavailableReasons(candidates []BackendCandidate) string {
	if len(candidates) == 0 {
		return "no backends configured for this platform"
	}
	parts := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if r := c.Isolator.Reason(); r != "" {
			parts = append(parts, fmt.Sprintf("%s: %s", c.Isolator.Name(), r))
		}
	}
	if len(parts) == 0 {
		return "no available backend"
	}
	return strings.Join(parts, "; ")
}

// infoFrom builds a BackendInfo from an isolator and its mode.
func infoFrom(iso OSIsolator, mode BackendMode) BackendInfo {
	return BackendInfo{
		Name:      iso.Name(),
		Mode:      mode,
		Available: iso.Available(),
		Reason:    iso.Reason(),
	}
}

// PlatformBackendCandidates returns the candidate list for the current platform.
// macOS: seatbelt, bwrap (Linux-only stub on darwin), native (stub).
// Linux: bwrap (real isolator if bubblewrap installed, otherwise unavailable), native (stub).
// Other: empty (will fallback to noop via SelectBackend).
func PlatformBackendCandidates() []BackendCandidate {
	switch runtime.GOOS {
	case "darwin":
		return []BackendCandidate{
			{Mode: BackendSeatbelt, Isolator: NewSeatbeltIsolator()},
			{Mode: BackendBwrap, Isolator: NewBwrapIsolator()},
			{Mode: BackendNative, Isolator: NewNativeStub()},
		}
	case "linux":
		return []BackendCandidate{
			{Mode: BackendBwrap, Isolator: NewBwrapIsolator()},
			{Mode: BackendNative, Isolator: NewNativeStub()},
		}
	default:
		return nil
	}
}

// --- Stub isolators for planned backends ---

// nativeStub is a placeholder isolator for the native kernel sandbox backend.
type nativeStub struct{}

// Compile-time interface compliance checks.
var _ OSIsolator = (*nativeStub)(nil)

// NewNativeStub returns a stub isolator for the native kernel sandbox backend.
func NewNativeStub() OSIsolator { return &nativeStub{} }

func (n *nativeStub) Apply(_ context.Context, _ *exec.Cmd, _ Policy) error {
	return ErrIsolatorUnavailable
}

func (n *nativeStub) Available() bool { return false }
func (n *nativeStub) Name() string    { return "native" }
func (n *nativeStub) Reason() string  { return "native backend not yet implemented" }
