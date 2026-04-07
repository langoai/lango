//go:build linux

package os

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// BwrapIsolator wraps exec.Cmd with bubblewrap (bwrap) on Linux to provide
// process-level isolation: filesystem bind mounts (read-only root + writable
// workspace), network unshare, and PID/IPC/UTS/cgroup namespaces.
//
// The absolute bwrap path is captured at probe time (NewBwrapIsolator) and
// reused at Apply time so that PATH or working-directory changes between
// probe and execution cannot redirect to a different binary.
type BwrapIsolator struct {
	available    bool
	reason       string
	version      string
	resolvedPath string
}

// Compile-time interface compliance check.
var _ OSIsolator = (*BwrapIsolator)(nil)

// NewBwrapIsolator probes the system for bwrap and returns an isolator that
// either reports availability with a captured absolute path + version, or is
// unavailable with a human-readable reason.
func NewBwrapIsolator() OSIsolator {
	path, err := exec.LookPath("bwrap")
	if err != nil {
		return &BwrapIsolator{
			reason: "bwrap binary not found in PATH (install bubblewrap package)",
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return &BwrapIsolator{
			reason: fmt.Sprintf("resolve absolute bwrap path: %v", err),
		}
	}
	out, err := exec.Command(abs, "--version").Output()
	if err != nil {
		return &BwrapIsolator{
			reason: fmt.Sprintf("bwrap --version failed: %v", err),
		}
	}
	return &BwrapIsolator{
		available:    true,
		version:      strings.TrimSpace(string(out)),
		resolvedPath: abs,
	}
}

// Apply rewrites cmd to run inside bwrap with arguments compiled from policy.
// The original argv becomes the child program after the bwrap "--" separator.
func (b *BwrapIsolator) Apply(_ context.Context, cmd *exec.Cmd, policy Policy) error {
	if !b.available {
		return ErrIsolatorUnavailable
	}
	args, err := compileBwrapArgs(policy)
	if err != nil {
		return fmt.Errorf("compile bwrap args: %w", err)
	}

	originalArgs := cmd.Args
	// Use the absolute path captured at probe time, NOT the bare "bwrap" string.
	// This guarantees probe-time and exec-time refer to the same binary even if
	// PATH or cwd changes between Apply() and cmd.Run().
	cmd.Path = b.resolvedPath
	cmd.Args = append(append([]string{b.resolvedPath}, args...), append([]string{"--"}, originalArgs...)...)
	return nil
}

func (b *BwrapIsolator) Available() bool { return b.available }
func (b *BwrapIsolator) Name() string    { return "bwrap" }

func (b *BwrapIsolator) Reason() string {
	if b.available {
		return ""
	}
	return b.reason
}

// Version returns the captured `bwrap --version` string, or "" when the
// isolator is unavailable. CLI surfaces (e.g. `lango sandbox test`) can
// type-assert to a versioner interface to display this.
func (b *BwrapIsolator) Version() string { return b.version }
