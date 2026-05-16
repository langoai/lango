package approval

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func executeApprovalCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestNewApprovalCmd_Structure(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewApprovalCmd(func() (*config.Config, error) { return cfg, nil })

	require.NotNil(t, cmd)
	assert.Equal(t, "approval", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewApprovalCmd_Subcommands(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewApprovalCmd(func() (*config.Config, error) { return cfg, nil })

	subCmds := make(map[string]bool, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		subCmds[sub.Name()] = true
	}

	assert.True(t, subCmds["status"], "missing subcommand: status")
}

func TestStatusCmd_HappyPath(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Interceptor.Enabled = true
	cfg.Security.Interceptor.ApprovalPolicy = config.ApprovalPolicyDangerous
	cfg.Security.Interceptor.HeadlessAutoApprove = false
	cfg.Security.Interceptor.ApprovalTimeoutSec = 30
	cfg.Security.Interceptor.RedactPII = true
	cmd := NewApprovalCmd(func() (*config.Config, error) { return cfg, nil })

	out, err := executeApprovalCommand(t, cmd, "status")
	require.NoError(t, err)
	assert.Contains(t, out, "Approval Status")
	assert.Contains(t, out, "Interceptor Enabled:   true")
	assert.Contains(t, out, "Approval Policy:       dangerous")
	assert.Contains(t, out, "Redact PII:            true")
}

func TestStatusCmd_JSONOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Interceptor.Enabled = false
	cfg.Security.Interceptor.ApprovalPolicy = config.ApprovalPolicyDangerous
	cfg.Security.Interceptor.SensitiveTools = []string{"shell_exec", "file_write"}
	cfg.Security.Interceptor.ExemptTools = []string{"search"}
	cmd := NewApprovalCmd(func() (*config.Config, error) { return cfg, nil })

	out, err := executeApprovalCommand(t, cmd, "status", "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"interceptor_enabled": false`)
	assert.Contains(t, out, `"approval_policy": "dangerous"`)
	assert.Contains(t, out, `"shell_exec"`)
	assert.Contains(t, out, `"search"`)
}

func TestStatusCmd_ConfigError(t *testing.T) {
	cmd := NewApprovalCmd(func() (*config.Config, error) { return nil, assert.AnError })

	_, err := executeApprovalCommand(t, cmd, "status")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load config")
}

func TestStatusCmd_WithSensitiveTools(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Interceptor.Enabled = true
	cfg.Security.Interceptor.SensitiveTools = []string{"shell_exec", "file_write"}
	cmd := NewApprovalCmd(func() (*config.Config, error) { return cfg, nil })

	out, err := executeApprovalCommand(t, cmd, "status")
	require.NoError(t, err)
	assert.Contains(t, out, "Sensitive Tools (2)")
	assert.Contains(t, out, "shell_exec")
	assert.Contains(t, out, "file_write")
}

func TestStatusCmd_WithExemptTools(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Interceptor.Enabled = true
	cfg.Security.Interceptor.ExemptTools = []string{"search", "get_time"}
	cmd := NewApprovalCmd(func() (*config.Config, error) { return cfg, nil })

	out, err := executeApprovalCommand(t, cmd, "status")
	require.NoError(t, err)
	assert.Contains(t, out, "Exempt Tools (2)")
	assert.Contains(t, out, "search")
	assert.Contains(t, out, "get_time")
}

func TestStatusCmd_WithNotifyChannel(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Interceptor.Enabled = true
	cfg.Security.Interceptor.NotifyChannel = "discord"
	cmd := NewApprovalCmd(func() (*config.Config, error) { return cfg, nil })

	out, err := executeApprovalCommand(t, cmd, "status")
	require.NoError(t, err)
	assert.Contains(t, out, "Notify Channel:        discord")
}

func TestStatusCmd_TableWritesToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Interceptor.Enabled = true
	cmd := NewApprovalCmd(func() (*config.Config, error) { return cfg, nil })

	out, err := executeApprovalCommand(t, cmd, "status")

	require.NoError(t, err)
	assert.Contains(t, out, "Approval Status")
	assert.Contains(t, out, "Interceptor Enabled:")
}

func TestStatusCmd_JSONWritesToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Interceptor.Enabled = true
	cfg.Security.Interceptor.ExemptTools = []string{"search"}
	cmd := NewApprovalCmd(func() (*config.Config, error) { return cfg, nil })

	out, err := executeApprovalCommand(t, cmd, "status", "--output", "json")

	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Equal(t, true, payload["interceptor_enabled"])
}

func TestStatusCmd_InvalidOutputFailsBeforeConfigLoad(t *testing.T) {
	cmd := NewApprovalCmd(func() (*config.Config, error) {
		t.Fatal("config loader should not be called for invalid output")
		return nil, nil
	})

	_, err := executeApprovalCommand(t, cmd, "status", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}
