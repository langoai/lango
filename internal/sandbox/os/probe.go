package os

import "runtime"

// PlatformCapabilities describes what OS-level sandbox primitives are available.
type PlatformCapabilities struct {
	// HasSeatbelt indicates macOS sandbox-exec is available.
	HasSeatbelt bool
	// SeatbeltReason explains the Seatbelt probe result (e.g., "sandbox-exec found", "not on darwin").
	SeatbeltReason string

	// HasLandlock indicates Linux Landlock LSM is available.
	HasLandlock bool
	// LandlockABI is the detected Landlock ABI version (0 = unavailable or not probed, 1-4 = supported).
	LandlockABI int
	// LandlockReason explains the Landlock probe result (e.g., "ABI 3", "probe not yet implemented").
	LandlockReason string

	// HasSeccomp indicates Linux seccomp-bpf is available.
	HasSeccomp bool
	// SeccompReason explains the seccomp probe result (e.g., "available", "probe not yet implemented").
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
	case c.Platform == "linux" && c.LandlockReason != "" && !c.HasLandlock:
		return "unknown (Landlock/seccomp probe not yet implemented)"
	default:
		return "none"
	}
}
