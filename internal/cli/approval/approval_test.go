package approval

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/testutil"
)

func TestNewApprovalCmd_Structure(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewApprovalCmd(testutil.FakeCfgLoader(cfg))

	require.NotNil(t, cmd)
	assert.Equal(t, "approval", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewApprovalCmd_Subcommands(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewApprovalCmd(testutil.FakeCfgLoader(cfg))

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
	cmd := NewApprovalCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmdOK(t, cmd, "status")
	assert.Contains(t, result.Stdout, "Approval Status")
	assert.Contains(t, result.Stdout, "Interceptor Enabled:   true")
	assert.Contains(t, result.Stdout, "Approval Policy:       dangerous")
	assert.Contains(t, result.Stdout, "Redact PII:            true")
}

func TestStatusCmd_JSONOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Interceptor.Enabled = false
	cfg.Security.Interceptor.ApprovalPolicy = config.ApprovalPolicyDangerous
	cfg.Security.Interceptor.SensitiveTools = []string{"shell_exec", "file_write"}
	cfg.Security.Interceptor.ExemptTools = []string{"search"}
	cmd := NewApprovalCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmdOK(t, cmd, "status", "--json")
	assert.Contains(t, result.Stdout, `"interceptor_enabled": false`)
	assert.Contains(t, result.Stdout, `"approval_policy": "dangerous"`)
	assert.Contains(t, result.Stdout, `"shell_exec"`)
	assert.Contains(t, result.Stdout, `"search"`)
}

func TestStatusCmd_ConfigError(t *testing.T) {
	cmd := NewApprovalCmd(testutil.FailCfgLoader(assert.AnError))

	result := testutil.ExecCmd(t, cmd, "status")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "load config")
}

func TestStatusCmd_WithSensitiveTools(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Interceptor.Enabled = true
	cfg.Security.Interceptor.SensitiveTools = []string{"shell_exec", "file_write"}
	cmd := NewApprovalCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmdOK(t, cmd, "status")
	assert.Contains(t, result.Stdout, "Sensitive Tools (2)")
	assert.Contains(t, result.Stdout, "shell_exec")
	assert.Contains(t, result.Stdout, "file_write")
}

func TestStatusCmd_WithExemptTools(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Interceptor.Enabled = true
	cfg.Security.Interceptor.ExemptTools = []string{"search", "get_time"}
	cmd := NewApprovalCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmdOK(t, cmd, "status")
	assert.Contains(t, result.Stdout, "Exempt Tools (2)")
	assert.Contains(t, result.Stdout, "search")
	assert.Contains(t, result.Stdout, "get_time")
}

func TestStatusCmd_WithNotifyChannel(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Security.Interceptor.Enabled = true
	cfg.Security.Interceptor.NotifyChannel = "discord"
	cmd := NewApprovalCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmdOK(t, cmd, "status")
	assert.Contains(t, result.Stdout, "Notify Channel:        discord")
}
