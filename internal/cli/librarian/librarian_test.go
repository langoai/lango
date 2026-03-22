package librarian

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/testutil"
)

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

	result := testutil.ExecCmdOK(t, cmd, "status")
	assert.Contains(t, result.Stdout, "Librarian Status")
	assert.Contains(t, result.Stdout, "Enabled:               true")
}

func TestStatusCmd_JSONOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Librarian.Enabled = false
	cfg.Librarian.Provider = "anthropic"
	cfg.Librarian.Model = "claude-4"
	cmd := NewLibrarianCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmdOK(t, cmd, "status", "--json")
	assert.Contains(t, result.Stdout, `"enabled": false`)
	assert.Contains(t, result.Stdout, `"provider": "anthropic"`)
	assert.Contains(t, result.Stdout, `"model": "claude-4"`)
}

func TestStatusCmd_ConfigError(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLibrarianCmd(testutil.FailCfgLoader(assert.AnError), testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmd(t, cmd, "status")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "load config")
}

func TestInquiriesCmd_EmptyDB(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLibrarianCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmdOK(t, cmd, "inquiries")
	assert.Contains(t, result.Stdout, "No pending inquiries.")
}

func TestInquiriesCmd_JSONEmptyDB(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLibrarianCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmdOK(t, cmd, "inquiries", "--json")
	assert.Contains(t, result.Stdout, "[]")
}

func TestInquiriesCmd_BootError(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLibrarianCmd(testutil.FakeCfgLoader(cfg), testutil.FailBootLoader(assert.AnError))

	result := testutil.ExecCmd(t, cmd, "inquiries")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "bootstrap")
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
