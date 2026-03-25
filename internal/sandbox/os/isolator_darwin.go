//go:build darwin

package os

import "os/exec"

func newPlatformIsolator() OSIsolator {
	iso := NewSeatbeltIsolator()
	if iso.Available() {
		return iso
	}
	return &noopIsolator{}
}

func probePlatform(caps *PlatformCapabilities) {
	// Check if sandbox-exec exists.
	if _, err := exec.LookPath("sandbox-exec"); err == nil {
		caps.HasSeatbelt = true
	}
	caps.KernelVersion = darwinKernelVersion()
}

func darwinKernelVersion() string {
	out, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return "unknown"
	}
	return string(out[:len(out)-1]) // trim trailing newline
}
