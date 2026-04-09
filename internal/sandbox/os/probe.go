package os

import "runtime"

// PlatformCapabilities describes what OS-level sandbox primitives are available.
type PlatformCapabilities struct {
	// HasSeatbelt indicates macOS sandbox-exec is available.
	HasSeatbelt bool
	// SeatbeltReason explains the Seatbelt probe result (e.g., "sandbox-exec found", "not on darwin").
	SeatbeltReason string

	// HasLandlock indicates Linux Landlock LSM is available (kernel 5.13+).
	// Detected by calling landlock_create_ruleset(NULL, 0, LANDLOCK_CREATE_RULESET_VERSION).
	HasLandlock bool
	// LandlockABI is the detected Landlock ABI version (0 = unavailable, 1+ = supported).
	LandlockABI int
	// LandlockReason explains the Landlock probe result (e.g., "Landlock ABI 3",
	// "Landlock not supported by this kernel (requires 5.13+)").
	LandlockReason string

	// HasSeccomp indicates that the kernel exposes the seccomp prctl interface
	// (PR_GET_SECCOMP succeeds). This does NOT prove that BPF filters can be
	// installed — it is a generic presence signal only. The SeccompReason field
	// carries the qualified description.
	HasSeccomp bool
	// SeccompReason explains the seccomp probe result, including the caveat
	// that PR_GET_SECCOMP success only proves interface presence, not filter
	// capability.
	SeccompReason string

	// Platform is the runtime GOOS value.
	Platform string

	// KernelVersion is the OS kernel version string.
	KernelVersion string
}

// Probe detects the available OS-level sandbox capabilities.
func Probe() PlatformCapabilities {
	caps := PlatformCapabilities{
		Platform: runtime.GOOS,
	}
	probePlatform(&caps)
	return caps
}

// Summary returns a human-readable summary of the detected capabilities.
func (c PlatformCapabilities) Summary() string {
	switch {
	case c.HasSeatbelt:
		return "seatbelt (macOS)"
	case c.HasLandlock && c.HasSeccomp:
		return "landlock+seccomp (Linux)"
	case c.HasLandlock:
		return "landlock (Linux, no seccomp)"
	case c.HasSeccomp:
		return "seccomp (Linux, no landlock)"
	case c.Platform == "linux":
		return "linux (no Landlock or seccomp interface detected)"
	default:
		return "none"
	}
}
