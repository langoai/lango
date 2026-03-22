package testutil

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
)

// CLIResult captures the output of a CLI command execution.
type CLIResult struct {
	Stdout string
	Stderr string
	Err    error
}

// ExecCmd executes a cobra command with the given args and captures output.
// It intercepts os.Stdout via os.Pipe to catch direct fmt.Print* / json.Encoder
// writes that bypass cobra's OutOrStdout.
//
// NOTE: This function replaces os.Stdout globally and is NOT safe for use
// with t.Parallel(). Tests using ExecCmd must run sequentially.
func ExecCmd(t testing.TB, cmd *cobra.Command, args ...string) CLIResult {
	t.Helper()

	// Capture os.Stdout via an OS pipe.
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "create pipe for stdout capture")
	os.Stdout = w

	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	execErr := cmd.Execute()

	// Close the write end and restore os.Stdout before reading.
	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	r.Close()

	return CLIResult{
		Stdout: buf.String(),
		Stderr: stderr.String(),
		Err:    execErr,
	}
}

// ExecCmdOK executes a cobra command and asserts it succeeds.
func ExecCmdOK(t testing.TB, cmd *cobra.Command, args ...string) CLIResult {
	t.Helper()
	result := ExecCmd(t, cmd, args...)
	require.NoError(t, result.Err, "command should succeed; stderr=%s", result.Stderr)
	return result
}

// FakeCfgLoader returns a cfgLoader func that always returns the given config.
func FakeCfgLoader(cfg *config.Config) func() (*config.Config, error) {
	return func() (*config.Config, error) {
		return cfg, nil
	}
}

// FailCfgLoader returns a cfgLoader func that always returns the given error.
func FailCfgLoader(err error) func() (*config.Config, error) {
	return func() (*config.Config, error) {
		return nil, err
	}
}

// FakeBootLoader returns a bootLoader func that creates an in-memory Ent client.
// The client is closed when the test completes.
func FakeBootLoader(t testing.TB, cfg *config.Config) func() (*bootstrap.Result, error) {
	return func() (*bootstrap.Result, error) {
		client := TestEntClient(t)
		return &bootstrap.Result{
			Config:   cfg,
			DBClient: client,
		}, nil
	}
}

// FailBootLoader returns a bootLoader func that always returns the given error.
func FailBootLoader(err error) func() (*bootstrap.Result, error) {
	return func() (*bootstrap.Result, error) {
		return nil, err
	}
}

