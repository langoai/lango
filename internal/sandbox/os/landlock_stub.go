//go:build !linux

package os

import (
	"context"
	"os/exec"
)

type landlockIsolator struct{}

// NewLandlockIsolator returns a stub on non-Linux.
func NewLandlockIsolator() OSIsolator { return &landlockIsolator{} }

func (l *landlockIsolator) Apply(_ context.Context, _ *exec.Cmd, _ Policy) error {
	return ErrIsolatorUnavailable
}

func (l *landlockIsolator) Available() bool { return false }
func (l *landlockIsolator) Name() string    { return "landlock" }
