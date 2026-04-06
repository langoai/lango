//go:build darwin

package os

import "os/exec"

func newPlatformIsolator() OSIsolator {
	iso := NewSeatbeltIsolator()
	if iso.Available() {
		return iso
	}
	return &noopIsolator{reason: iso.Reason()}
}

func probePlatform(caps *PlatformCapabilities) {
	if _, err := exec.LookPath("sandbox-exec"); err == nil {
		caps.HasSeatbelt = true
		caps.SeatbeltReason = "sandbox-exec found"
	} else {
		caps.SeatbeltReason = "sandbox-exec not found in PATH"
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
