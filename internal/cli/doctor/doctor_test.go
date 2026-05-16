package doctor

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/doctor/checks"
	"github.com/langoai/lango/internal/config"
)

type stubDoctorCheck struct {
	result checks.Result
}

func (s stubDoctorCheck) Name() string { return s.result.Name }

func (s stubDoctorCheck) Run(_ context.Context, _ *config.Config) checks.Result {
	return s.result
}

func (s stubDoctorCheck) Fix(_ context.Context, _ *config.Config) checks.Result {
	return s.result
}

func executeDoctorCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestDoctorCommand_TableWritesToCommandOutput(t *testing.T) {
	origBootstrapRun := doctorBootstrapRun
	origAllChecks := doctorAllChecks
	t.Cleanup(func() {
		doctorBootstrapRun = origBootstrapRun
		doctorAllChecks = origAllChecks
	})

	doctorBootstrapRun = func(opts bootstrap.Options) (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: config.DefaultConfig()}, nil
	}
	doctorAllChecks = func() []checks.Check {
		return []checks.Check{
			stubDoctorCheck{result: checks.Result{
				Name:    "Stub Check",
				Status:  checks.StatusPass,
				Message: "doctor output routed",
			}},
		}
	}

	cmd := NewCommand()
	out, err := executeDoctorCommand(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, out, "Lango Doctor")
	assert.Contains(t, out, "Stub Check")
	assert.Contains(t, out, "doctor output routed")
}

func TestDoctorCommand_JSONWritesToCommandOutput(t *testing.T) {
	origBootstrapRun := doctorBootstrapRun
	origAllChecks := doctorAllChecks
	t.Cleanup(func() {
		doctorBootstrapRun = origBootstrapRun
		doctorAllChecks = origAllChecks
	})

	doctorBootstrapRun = func(opts bootstrap.Options) (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: config.DefaultConfig()}, nil
	}
	doctorAllChecks = func() []checks.Check {
		return []checks.Check{
			stubDoctorCheck{result: checks.Result{
				Name:    "Stub Check",
				Status:  checks.StatusPass,
				Message: "doctor json routed",
			}},
		}
	}

	cmd := NewCommand()
	out, err := executeDoctorCommand(t, cmd, "--output", "json")

	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	results, ok := decoded["results"].([]any)
	require.True(t, ok)
	require.Len(t, results, 1)
}

func TestDoctorCommand_InvalidOutputFailsBeforeBootstrap(t *testing.T) {
	origBootstrapRun := doctorBootstrapRun
	t.Cleanup(func() {
		doctorBootstrapRun = origBootstrapRun
	})

	doctorBootstrapRun = func(opts bootstrap.Options) (*bootstrap.Result, error) {
		t.Fatal("bootstrap should not run for invalid output")
		return nil, nil
	}

	cmd := NewCommand()
	out, err := executeDoctorCommand(t, cmd, "--output", "yaml")

	require.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestDoctorHelpReflectsCurrentCheckFamilies(t *testing.T) {
	cmd := NewCommand()
	out, err := executeDoctorCommand(t, cmd, "--help")

	require.NoError(t, err)
	assert.Contains(t, out, "Checks performed (27 total):")
	assert.Contains(t, out, "Core Configuration:")
	assert.Contains(t, out, "Tool Hooks & Agent Management:")
	assert.Contains(t, out, "Economy / Contract / Observability:")
	assert.Contains(t, out, "P2P Workspace:")
	assert.Contains(t, out, "Use --output json for machine-readable output.")
}
