//go:build !darwin && !linux

package os

func newPlatformIsolator() OSIsolator {
	return &noopIsolator{}
}

func probePlatform(caps *PlatformCapabilities) {
	// No OS-level sandbox available on this platform.
}
