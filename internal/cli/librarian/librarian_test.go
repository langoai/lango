package librarian

import (
	"bytes"
	"testing"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/testutil"
)

func executeLibrarianCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestNewLibrarianCmd_Structure(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLibrarianCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	require.NotNil(t, cmd)
	assert.Equal(t, "librarian", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewLibrarianCmd_Subcommands(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLibrarianCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	expected := []string{"status", "inquiries"}
	subCmds := make(map[string]bool, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		subCmds[sub.Name()] = true
	}

	for _, name := range expected {
		assert.True(t, subCmds[name], "missing subcommand: %s", name)
	}
}

func TestStatusCmd_HappyPath(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Librarian.Enabled = true
	cfg.Librarian.ObservationThreshold = 5
	cfg.Librarian.InquiryCooldownTurns = 3
	cfg.Librarian.MaxPendingInquiries = 2
	cfg.Librarian.AutoSaveConfidence = "high"
	cmd := NewLibrarianCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	out, err := executeLibrarianCmd(t, cmd, "status")
	require.NoError(t, err)
	assert.Contains(t, out, "Librarian Status")
	assert.Contains(t, out, "Enabled:               true")
}

func TestStatusCmd_JSONOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Librarian.Enabled = false
	cfg.Librarian.Provider = "anthropic"
	cfg.Librarian.Model = "claude-4"
	cmd := NewLibrarianCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	out, err := executeLibrarianCmd(t, cmd, "status", "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"enabled": false`)
	assert.Contains(t, out, `"provider": "anthropic"`)
	assert.Contains(t, out, `"model": "claude-4"`)
}

func TestStatusCmd_InvalidOutputFailsBeforeConfigLoad(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLibrarianCmd(func() (*config.Config, error) {
		t.Fatal("config loader should not be called for invalid output")
		return nil, nil
	}, testutil.FakeBootLoader(t, cfg))

	_, err := executeLibrarianCmd(t, cmd, "status", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestStatusCmd_ConfigError(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLibrarianCmd(testutil.FailCfgLoader(assert.AnError), testutil.FakeBootLoader(t, cfg))

	_, err := executeLibrarianCmd(t, cmd, "status")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load config")
}

func TestInquiriesCmd_EmptyDB(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLibrarianCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	out, err := executeLibrarianCmd(t, cmd, "inquiries")
	require.NoError(t, err)
	assert.Contains(t, out, "No pending inquiries.")
}

func TestInquiriesCmd_JSONEmptyDB(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLibrarianCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	out, err := executeLibrarianCmd(t, cmd, "inquiries", "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "[]")
}

func TestInquiriesCmd_InvalidOutputFailsBeforeBootLoad(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLibrarianCmd(testutil.FakeCfgLoader(cfg), func() (*bootstrap.Result, error) {
		t.Fatal("boot loader should not be called for invalid output")
		return nil, nil
	})

	_, err := executeLibrarianCmd(t, cmd, "inquiries", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestInquiriesCmd_BootError(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLibrarianCmd(testutil.FakeCfgLoader(cfg), testutil.FailBootLoader(assert.AnError))

	_, err := executeLibrarianCmd(t, cmd, "inquiries")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bootstrap")
}

func TestInquiriesCmd_HasLimitFlag(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLibrarianCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	for _, sub := range cmd.Commands() {
		if sub.Name() == "inquiries" {
			f := sub.Flags().Lookup("limit")
			require.NotNil(t, f, "inquiries command should have --limit flag")
			assert.Equal(t, "20", f.DefValue)
			return
		}
	}
	t.Fatal("inquiries subcommand not found")
}
