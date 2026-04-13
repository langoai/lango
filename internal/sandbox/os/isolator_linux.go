//go:build linux

package os

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/unix"
)

func newPlatformIsolator() OSIsolator {
	return &noopIsolator{
		reason: "Linux native (Landlock+seccomp) backend not yet implemented; use backend=bwrap",
	}
}

func probePlatform(caps *PlatformCapabilities) {
	available, abi, reason := probeLandlockKernel()
	caps.HasLandlock = available
	caps.LandlockABI = abi
	caps.LandlockReason = reason

	seccompAvailable, seccompReason := probeSeccompKernel()
	caps.HasSeccomp = seccompAvailable
	caps.SeccompReason = seccompReason

	caps.KernelVersion = linuxKernelVersion()
}

// probeLandlockKernel calls landlock_create_ruleset(NULL, 0, LANDLOCK_CREATE_RULESET_VERSION)
// to detect kernel Landlock LSM support and ABI version. Requires kernel 5.13+.
//
// The returned ABI version is the value the kernel reports for the
// LANDLOCK_CREATE_RULESET_VERSION query (>0 on success). ENOSYS means the
// running kernel was not built with CONFIG_SECURITY_LANDLOCK.
func probeLandlockKernel() (available bool, abiVersion int, reason string) {
	r1, _, errno := unix.Syscall(
		unix.SYS_LANDLOCK_CREATE_RULESET,
		0, // ruleset_attr = NULL
		0, // size = 0
		uintptr(unix.LANDLOCK_CREATE_RULESET_VERSION),
	)
	if errno == 0 {
		abi := int(r1)
		if abi > 0 {
			return true, abi, fmt.Sprintf("Landlock ABI %d", abi)
		}
		return false, 0, "Landlock probe returned non-positive ABI version"
	}
	if errors.Is(errno, unix.ENOSYS) {
		return false, 0, "Landlock not supported by this kernel (requires 5.13+)"
	}
	return false, 0, "Landlock probe failed: " + errno.Error()
}

// probeSeccompKernel calls prctl(PR_GET_SECCOMP) to detect that the kernel
// exposes the seccomp prctl interface. NOTE: success here only proves that
// the current process can read its own seccomp mode — it does NOT prove that
// BPF filters are installable. The reason field reflects this caveat, and we
// best-effort augment it with /proc/self/status:Seccomp.
func probeSeccompKernel() (available bool, reason string) {
	mode, err := unix.PrctlRetInt(unix.PR_GET_SECCOMP, 0, 0, 0, 0)
	if err == nil {
		base := fmt.Sprintf(
			"seccomp interface present (PR_GET_SECCOMP=%d); BPF filter capability not directly verified",
			mode,
		)
		if procMode := readProcSelfSeccompMode(); procMode != "" {
			return true, base + " (/proc/self/status Seccomp=" + procMode + ")"
		}
		return true, base
	}
	if errors.Is(err, unix.ENOSYS) || errors.Is(err, unix.EINVAL) {
		return false, "seccomp not supported by this kernel (PR_GET_SECCOMP unavailable)"
	}
	return false, "seccomp probe failed: " + err.Error()
}

// readProcSelfSeccompMode returns the value of the "Seccomp:" field from
// /proc/self/status, or "" if the file is unreadable or the field is missing.
// Presence of this field implies CONFIG_SECCOMP=y in the running kernel.
func readProcSelfSeccompMode() string {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return ""
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if rest, ok := strings.CutPrefix(line, "Seccomp:"); ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

func linuxKernelVersion() string {
	out, err := exec.Command("uname", "-r").Output()
	if err != nil || len(out) == 0 {
		return "unknown"
	}
	return string(out[:len(out)-1]) // trim trailing newline
}
