//go:build !linux

package os

import (
	"context"
	"os/exec"
)

// bwrapNonLinuxStub is the bwrap isolator placeholder for non-Linux platforms.
// It satisfies OSIsolator so PlatformBackendCandidates can register a "bwrap"
// slot on macOS without leaking Linux-only build dependencies.
type bwrapNonLinuxStub struct{}

// Compile-time interface compliance check.
var _ OSIsolator = (*bwrapNonLinuxStub)(nil)

// NewBwrapIsolator returns a non-Linux stub. The bwrap binary only exists on
// Linux; on darwin/windows the slot stays unavailable with a clear reason so
// `lango sandbox status` shows it as a recognised backend that simply does
// not apply to this platform.
func NewBwrapIsolator() OSIsolator { return &bwrapNonLinuxStub{} }

func (b *bwrapNonLinuxStub) Apply(_ context.Context, _ *exec.Cmd, _ Policy) error {
	return ErrIsolatorUnavailable
}

func (b *bwrapNonLinuxStub) Available() bool { return false }
func (b *bwrapNonLinuxStub) Name() string    { return "bwrap" }
func (b *bwrapNonLinuxStub) Reason() string  { return "bwrap is Linux-only" }
