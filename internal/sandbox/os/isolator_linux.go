//go:build linux

package os

import "os/exec"

func newPlatformIsolator() OSIsolator {
	return &noopIsolator{reason: "Linux isolation backend not yet implemented"}
}

func probePlatform(caps *PlatformCapabilities) {
	caps.HasLandlock, caps.LandlockABI = probeLandlockKernel()
	caps.LandlockReason = "probe not yet implemented"
	caps.HasSeccomp = probeSeccompKernel()
	caps.SeccompReason = "probe not yet implemented"
	caps.KernelVersion = linuxKernelVersion()
}

// probeLandlockKernel probes the kernel for Landlock LSM support.
// TODO: implement actual syscall probe via x/sys/unix.
func probeLandlockKernel() (available bool, abiVersion int) {
	return false, 0
}

// probeSeccompKernel probes the kernel for seccomp-bpf support.
// TODO: implement actual syscall probe via x/sys/unix.
func probeSeccompKernel() bool {
	return false
}

func linuxKernelVersion() string {
	out, err := exec.Command("uname", "-r").Output()
	if err != nil || len(out) == 0 {
		return "unknown"
	}
	return string(out[:len(out)-1]) // trim trailing newline
}
