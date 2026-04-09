//go:build !darwin && !linux

package os

func newPlatformIsolator() OSIsolator {
	return &noopIsolator{reason: "unsupported platform"}
}

func probePlatform(caps *PlatformCapabilities) {
	caps.SeatbeltReason = "not on darwin"
	caps.LandlockReason = "not on Linux"
	caps.SeccompReason = "not on Linux"
}
