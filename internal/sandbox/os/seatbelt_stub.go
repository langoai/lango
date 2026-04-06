//go:build !darwin

package os

import (
	"context"
	"os/exec"
)

// SeatbeltIsolator is unavailable on non-macOS platforms.
type SeatbeltIsolator struct{}

// NewSeatbeltIsolator returns a stub on non-macOS.
func NewSeatbeltIsolator() *SeatbeltIsolator { return &SeatbeltIsolator{} }

func (s *SeatbeltIsolator) Apply(_ context.Context, _ *exec.Cmd, _ Policy) error {
	return ErrIsolatorUnavailable
}

func (s *SeatbeltIsolator) Available() bool { return false }
func (s *SeatbeltIsolator) Name() string    { return "seatbelt" }
func (s *SeatbeltIsolator) Reason() string  { return "not on macOS" }

// CleanupProfileFile is a no-op on non-macOS.
func CleanupProfileFile(_ *exec.Cmd) {}
