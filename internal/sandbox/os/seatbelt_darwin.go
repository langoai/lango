//go:build darwin

package os

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// SeatbeltIsolator wraps exec.Cmd with macOS sandbox-exec(1).
type SeatbeltIsolator struct {
	available bool
}

// NewSeatbeltIsolator creates a macOS Seatbelt isolator.
func NewSeatbeltIsolator() *SeatbeltIsolator {
	_, err := exec.LookPath("sandbox-exec")
	return &SeatbeltIsolator{available: err == nil}
}

func (s *SeatbeltIsolator) Apply(_ context.Context, cmd *exec.Cmd, policy Policy) error {
	if !s.available {
		return ErrIsolatorUnavailable
	}

	profile, err := GenerateSeatbeltProfile(policy)
	if err != nil {
		return fmt.Errorf("generate seatbelt profile: %w", err)
	}

	// Write profile to temp file (in /tmp, which the sandbox itself allows).
	tmpFile, err := os.CreateTemp("", "lango-seatbelt-*.sb")
	if err != nil {
		return fmt.Errorf("create temp profile: %w", err)
	}
	if _, err := tmpFile.WriteString(profile); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return fmt.Errorf("write seatbelt profile: %w", err)
	}
	tmpFile.Close()

	// Wrap the command: sandbox-exec -f <profile> <original-cmd> <original-args>
	originalArgs := cmd.Args
	cmd.Path = "/usr/bin/sandbox-exec"
	cmd.Args = append([]string{"sandbox-exec", "-f", tmpFile.Name()}, originalArgs...)

	// Schedule cleanup of temp file after process exits.
	// The caller is responsible for calling cmd.Run() which will use the wrapped command.
	// We clean up the profile file in a goroutine that waits for the process to complete.
	origEnv := cmd.Env
	if origEnv == nil {
		origEnv = os.Environ()
	}
	// Store cleanup path in environment so it can be cleaned up.
	// We use a goroutine approach: wrap the Wait to clean up after.
	cmd.Env = append(origEnv, "_LANGO_SEATBELT_PROFILE="+tmpFile.Name())

	return nil
}

func (s *SeatbeltIsolator) Available() bool { return s.available }
func (s *SeatbeltIsolator) Name() string    { return "seatbelt" }

// CleanupProfileFile removes a Seatbelt profile temp file if it exists.
// Call this after the sandboxed process exits.
func CleanupProfileFile(cmd *exec.Cmd) {
	for _, env := range cmd.Env {
		if len(env) > len("_LANGO_SEATBELT_PROFILE=") && env[:len("_LANGO_SEATBELT_PROFILE=")] == "_LANGO_SEATBELT_PROFILE=" {
			path := env[len("_LANGO_SEATBELT_PROFILE="):]
			os.Remove(path)
			return
		}
	}
}
