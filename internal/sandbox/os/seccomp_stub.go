//go:build !linux

package os

import (
	"context"
	"os/exec"
)

type seccompIsolator struct{}

// NewSeccompIsolator returns a stub on non-Linux.
func NewSeccompIsolator() OSIsolator { return &seccompIsolator{} }

func (s *seccompIsolator) Apply(_ context.Context, _ *exec.Cmd, _ Policy) error {
	return ErrIsolatorUnavailable
}

func (s *seccompIsolator) Available() bool { return false }
func (s *seccompIsolator) Name() string    { return "seccomp" }
func (s *seccompIsolator) Reason() string  { return "not on Linux" }
