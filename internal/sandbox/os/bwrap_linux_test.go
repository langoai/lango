//go:build linux

package os

import (
	"context"
	"os"
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
	// DefaultToolPolicy denies <work>/.git as a baseline; compileBwrapArgs
	// requires deny paths to exist as directories.
	require.NoError(t, os.Mkdir(filepath.Join(work, ".git"), 0o755))

	cmd := exec.Command("/bin/echo", "hello")
	originalArgs := append([]string{}, cmd.Args...)

	err := iso.Apply(context.Background(), cmd, DefaultToolPolicy(work, ""))
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

// TestNewBwrapIsolator_NetworkIsolationContract_HostDependent verifies the
// two-phase smoke probe contract: when base availability is true, Reason()
// must be empty regardless of network probe result, and the separate
// NetworkIsolationAvailable/NetworkIsolationReason methods must follow the
// spec — base-available + network-unavailable must surface a non-empty
// NetworkIsolationReason while keeping Available()==true.
func TestNewBwrapIsolator_NetworkIsolationContract_HostDependent(t *testing.T) {
	iso := NewBwrapIsolator()
	if !iso.Available() {
		t.Skipf("bwrap unavailable on this host: %s", iso.Reason())
	}
	bwrap, ok := iso.(*BwrapIsolator)
	require.True(t, ok, "Linux build should return *BwrapIsolator")

	assert.Empty(t, iso.Reason(),
		"Reason() must stay empty when Available()==true, even with partial network degradation")

	if bwrap.NetworkIsolationAvailable() {
		assert.Empty(t, bwrap.NetworkIsolationReason(),
			"NetworkIsolationReason must be empty when NetworkIsolationAvailable==true")
	} else {
		assert.NotEmpty(t, bwrap.NetworkIsolationReason(),
			"NetworkIsolationReason must explain degradation when NetworkIsolationAvailable==false")
	}
}

// TestBwrapIsolator_ApplyRejectsNetworkDenyWhenDowngraded force-constructs
// an isolator with base availability but the network probe marked failed,
// and verifies that Apply with a NetworkDeny policy returns an
// ErrIsolatorUnavailable-wrapped error WITHOUT mutating cmd.Args. The
// cmd.Args preservation is load-bearing: callers may retry or fall back to
// a different isolator, and a partially-mutated cmd would leave the
// caller in an inconsistent state.
func TestBwrapIsolator_ApplyRejectsNetworkDenyWhenDowngraded(t *testing.T) {
	iso := &BwrapIsolator{
		available:              true,
		resolvedPath:           "/usr/bin/bwrap", // synthetic — never executed
		networkIsolation:       false,
		networkIsolationReason: "simulated network probe failure",
	}
	cmd := exec.Command("/bin/echo", "hello")
	originalArgs := append([]string{}, cmd.Args...)
	originalPath := cmd.Path

	policy := Policy{
		Filesystem: FilesystemPolicy{ReadOnlyGlobal: true},
		Network:    NetworkDeny,
		Process:    ProcessPolicy{AllowFork: true},
	}
	err := iso.Apply(context.Background(), cmd, policy)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIsolatorUnavailable)
	assert.Contains(t, err.Error(), "simulated network probe failure")

	assert.Equal(t, originalPath, cmd.Path,
		"cmd.Path must NOT be mutated when Apply rejects a policy")
	assert.Equal(t, originalArgs, cmd.Args,
		"cmd.Args must NOT be mutated when Apply rejects a policy")
}

// TestBwrapIsolator_ApplyPermitsNetworkAllowWhenDowngraded verifies that a
// partial-degradation isolator (base available, network probe failed) still
// wraps NetworkAllow policies — MCPServerPolicy uses NetworkAllow, so a
// host that blocks --unshare-net must not stop MCP from running. We can't
// actually Run() the synthetic command, but we can verify Apply rewrites
// cmd.Path and cmd.Args without returning an error.
func TestBwrapIsolator_ApplyPermitsNetworkAllowWhenDowngraded(t *testing.T) {
	const syntheticBwrap = "/usr/bin/bwrap"
	iso := &BwrapIsolator{
		available:              true,
		resolvedPath:           syntheticBwrap,
		networkIsolation:       false,
		networkIsolationReason: "simulated network probe failure",
	}
	cmd := exec.Command("/bin/echo", "hello")
	originalArgs := append([]string{}, cmd.Args...)

	policy := Policy{
		Filesystem: FilesystemPolicy{ReadOnlyGlobal: true},
		Network:    NetworkAllow, // MCP-style policy, should pass the gate
		Process:    ProcessPolicy{AllowFork: true},
	}
	err := iso.Apply(context.Background(), cmd, policy)
	require.NoError(t, err,
		"Apply must succeed for NetworkAllow even when network isolation probe failed")

	assert.Equal(t, syntheticBwrap, cmd.Path,
		"cmd.Path must be rewritten to the resolved bwrap path")
	require.NotEmpty(t, cmd.Args)
	assert.Equal(t, syntheticBwrap, cmd.Args[0],
		"cmd.Args[0] must match the resolved bwrap path")

	// Original argv must appear after the "--" separator.
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

// TestSmokeProbeBwrapBase_HostDependent exercises the base smoke probe
// directly. It does not assert a specific outcome — on a normal Linux host
// with bwrap installed and userns enabled, the probe succeeds; on hardened
// hosts it may fail. Either way, the function must not panic and must return
// a non-nil error xor a nil error.
func TestSmokeProbeBwrapBase_HostDependent(t *testing.T) {
	abs, err := exec.LookPath("bwrap")
	if err != nil {
		t.Skip("bwrap not installed")
	}
	probeErr := smokeProbeBwrapBase(abs)
	t.Logf("base smoke probe result: %v", probeErr)
}

// TestSmokeProbeBwrapNetwork_HostDependent exercises the network smoke probe
// directly. Same observational semantics as the base probe test.
func TestSmokeProbeBwrapNetwork_HostDependent(t *testing.T) {
	abs, err := exec.LookPath("bwrap")
	if err != nil {
		t.Skip("bwrap not installed")
	}
	probeErr := smokeProbeBwrapNetwork(abs)
	t.Logf("network smoke probe result: %v", probeErr)
}
