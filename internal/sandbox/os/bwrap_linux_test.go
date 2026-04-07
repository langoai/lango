//go:build linux

package os

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBwrapIsolator_HostDependent(t *testing.T) {
	iso := NewBwrapIsolator()
	require.Equal(t, "bwrap", iso.Name())

	// Availability depends on whether bwrap is installed in this environment.
	// We do not assert true/false directly — we only assert the contract.
	if iso.Available() {
		assert.Empty(t, iso.Reason(), "available isolator should have no reason")
		// When available, BwrapIsolator should expose a captured version string.
		bwrap, ok := iso.(*BwrapIsolator)
		require.True(t, ok, "Linux build should return *BwrapIsolator")
		assert.NotEmpty(t, bwrap.Version(), "available isolator should have captured a version")
		assert.True(t, filepath.IsAbs(bwrap.resolvedPath), "resolvedPath must be absolute")
	} else {
		assert.NotEmpty(t, iso.Reason(), "unavailable isolator must explain why")
	}
}

func TestBwrapIsolator_ApplyWrapsCommand(t *testing.T) {
	if _, err := exec.LookPath("bwrap"); err != nil {
		t.Skip("bwrap not installed; skipping Apply wrapping test")
	}

	iso := NewBwrapIsolator()
	require.True(t, iso.Available(), "bwrap was found via LookPath but isolator reports unavailable")

	work := t.TempDir()
	cmd := exec.Command("/bin/echo", "hello")
	originalArgs := append([]string{}, cmd.Args...)

	err := iso.Apply(context.Background(), cmd, DefaultToolPolicy(work))
	require.NoError(t, err)

	bwrap := iso.(*BwrapIsolator)
	assert.Equal(t, bwrap.resolvedPath, cmd.Path,
		"cmd.Path must be the absolute bwrap path captured at probe time")
	require.NotEmpty(t, cmd.Args)
	assert.Equal(t, bwrap.resolvedPath, cmd.Args[0],
		"cmd.Args[0] must match resolved bwrap path (not bare 'bwrap')")

	// The original command must appear after a "--" separator.
	sepIndex := -1
	for i, a := range cmd.Args {
		if a == "--" {
			sepIndex = i
			break
		}
	}
	require.NotEqual(t, -1, sepIndex, `cmd.Args must contain a "--" separator`)
	assert.Equal(t, originalArgs, cmd.Args[sepIndex+1:],
		"args after -- must be the original command")
}

func TestBwrapIsolator_ApplyUnavailableReturnsError(t *testing.T) {
	// Force-construct an unavailable BwrapIsolator without invoking the probe
	// so the test runs even when bwrap is installed.
	iso := &BwrapIsolator{available: false, reason: "test"}
	err := iso.Apply(context.Background(), &exec.Cmd{}, Policy{})
	assert.ErrorIs(t, err, ErrIsolatorUnavailable)
}

func TestBwrapIsolator_ReasonWhenAvailable(t *testing.T) {
	iso := &BwrapIsolator{available: true}
	assert.Empty(t, iso.Reason())
}
