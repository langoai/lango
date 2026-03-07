package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_WithTempYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "lango.json")

	content := `{
		"server": { "port": 9999 },
		"agent": { "provider": "anthropic" },
		"logging": { "level": "debug", "format": "json" }
	}`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	cfg, err := Load(cfgPath)
	require.NoError(t, err)

	assert.Equal(t, 9999, cfg.Server.Port)
	assert.Equal(t, "anthropic", cfg.Agent.Provider)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
}

func TestLoad_DefaultsWhenNoFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "nonexistent.json")

	cfg, err := Load(cfgPath)
	// File not found with explicit path returns an error
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_InvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "lango.json")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`{invalid json`), 0644))

	cfg, err := Load(cfgPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_EnvOverrides(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "lango.json")

	envKey := "TEST_LOAD_ENV_KEY_OVERRIDE"
	os.Setenv(envKey, "resolved-api-key")
	defer os.Unsetenv(envKey)

	content := `{
		"providers": {
			"anthropic": { "type": "anthropic", "apiKey": "${` + envKey + `}" }
		},
		"agent": { "provider": "anthropic" },
		"logging": { "level": "info", "format": "console" }
	}`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	cfg, err := Load(cfgPath)
	require.NoError(t, err)

	assert.Equal(t, "resolved-api-key", cfg.Providers["anthropic"].APIKey)
}

func TestLoad_ValidationFailure(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "lango.json")

	content := `{
		"server": { "port": 0 },
		"logging": { "level": "info", "format": "console" }
	}`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	cfg, err := Load(cfgPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "invalid port")
}

func TestLoad_PartialConfig_UsesDefaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "lango.json")

	// Only override logging; everything else should use defaults
	content := `{
		"logging": { "level": "warn", "format": "json" }
	}`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0644))

	cfg, err := Load(cfgPath)
	require.NoError(t, err)

	// Overridden values
	assert.Equal(t, "warn", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)

	// Default values preserved
	assert.Equal(t, 18789, cfg.Server.Port)
	assert.Equal(t, "anthropic", cfg.Agent.Provider)
	assert.True(t, cfg.Security.Interceptor.Enabled)
}

func TestExpandEnvVars_MultipleVars(t *testing.T) {
	os.Setenv("EXPAND_A", "hello")
	os.Setenv("EXPAND_B", "world")
	defer os.Unsetenv("EXPAND_A")
	defer os.Unsetenv("EXPAND_B")

	result := ExpandEnvVars("${EXPAND_A} ${EXPAND_B}")
	assert.Equal(t, "hello world", result)
}

func TestExpandEnvVars_NoVars(t *testing.T) {
	t.Parallel()

	result := ExpandEnvVars("plain string no vars")
	assert.Equal(t, "plain string no vars", result)
}

func TestExpandEnvVars_EmptyString(t *testing.T) {
	t.Parallel()

	result := ExpandEnvVars("")
	assert.Empty(t, result)
}

func TestSubstituteEnvVars_Providers(t *testing.T) {
	os.Setenv("SUB_TEST_KEY", "my-secret")
	defer os.Unsetenv("SUB_TEST_KEY")

	cfg := &Config{
		Providers: map[string]ProviderConfig{
			"test": {APIKey: "${SUB_TEST_KEY}"},
		},
	}
	substituteEnvVars(cfg)

	assert.Equal(t, "my-secret", cfg.Providers["test"].APIKey)
}

func TestSubstituteEnvVars_Channels(t *testing.T) {
	os.Setenv("SUB_TG_TOKEN", "tg-token-123")
	os.Setenv("SUB_DISCORD_TOKEN", "dc-token-456")
	os.Setenv("SUB_SLACK_TOKEN", "sl-token-789")
	defer os.Unsetenv("SUB_TG_TOKEN")
	defer os.Unsetenv("SUB_DISCORD_TOKEN")
	defer os.Unsetenv("SUB_SLACK_TOKEN")

	cfg := &Config{}
	cfg.Channels.Telegram.BotToken = "${SUB_TG_TOKEN}"
	cfg.Channels.Discord.BotToken = "${SUB_DISCORD_TOKEN}"
	cfg.Channels.Slack.BotToken = "${SUB_SLACK_TOKEN}"
	substituteEnvVars(cfg)

	assert.Equal(t, "tg-token-123", cfg.Channels.Telegram.BotToken)
	assert.Equal(t, "dc-token-456", cfg.Channels.Discord.BotToken)
	assert.Equal(t, "sl-token-789", cfg.Channels.Slack.BotToken)
}

func TestSubstituteEnvVars_MCPServers(t *testing.T) {
	os.Setenv("SUB_MCP_KEY", "mcp-secret")
	defer os.Unsetenv("SUB_MCP_KEY")

	cfg := &Config{
		MCP: MCPConfig{
			Servers: map[string]MCPServerConfig{
				"test": {
					Env:     map[string]string{"API_KEY": "${SUB_MCP_KEY}"},
					Headers: map[string]string{"Authorization": "Bearer ${SUB_MCP_KEY}"},
				},
			},
		},
	}
	substituteEnvVars(cfg)

	assert.Equal(t, "mcp-secret", cfg.MCP.Servers["test"].Env["API_KEY"])
	assert.Equal(t, "Bearer mcp-secret", cfg.MCP.Servers["test"].Headers["Authorization"])
}

func TestSubstituteEnvVars_AuthProviders(t *testing.T) {
	os.Setenv("SUB_AUTH_ID", "my-client-id")
	os.Setenv("SUB_AUTH_SECRET", "my-client-secret")
	defer os.Unsetenv("SUB_AUTH_ID")
	defer os.Unsetenv("SUB_AUTH_SECRET")

	cfg := &Config{
		Auth: AuthConfig{
			Providers: map[string]OIDCProviderConfig{
				"google": {
					ClientID:     "${SUB_AUTH_ID}",
					ClientSecret: "${SUB_AUTH_SECRET}",
				},
			},
		},
	}
	substituteEnvVars(cfg)

	assert.Equal(t, "my-client-id", cfg.Auth.Providers["google"].ClientID)
	assert.Equal(t, "my-client-secret", cfg.Auth.Providers["google"].ClientSecret)
}

func TestSubstituteEnvVars_Payment(t *testing.T) {
	os.Setenv("SUB_RPC_URL", "https://rpc.example.com")
	defer os.Unsetenv("SUB_RPC_URL")

	cfg := &Config{}
	cfg.Payment.Network.RPCURL = "${SUB_RPC_URL}"
	substituteEnvVars(cfg)

	assert.Equal(t, "https://rpc.example.com", cfg.Payment.Network.RPCURL)
}

func TestSubstituteEnvVars_SessionDatabasePath(t *testing.T) {
	os.Setenv("SUB_DB_PATH", "/custom/db.sqlite")
	defer os.Unsetenv("SUB_DB_PATH")

	cfg := &Config{}
	cfg.Session.DatabasePath = "${SUB_DB_PATH}"
	substituteEnvVars(cfg)

	assert.Equal(t, "/custom/db.sqlite", cfg.Session.DatabasePath)
}

func TestSubstituteEnvVars_SlackAppTokenAndSigningSecret(t *testing.T) {
	os.Setenv("SUB_SLACK_APP", "xapp-token")
	os.Setenv("SUB_SLACK_SIGN", "signing-secret")
	defer os.Unsetenv("SUB_SLACK_APP")
	defer os.Unsetenv("SUB_SLACK_SIGN")

	cfg := &Config{}
	cfg.Channels.Slack.AppToken = "${SUB_SLACK_APP}"
	cfg.Channels.Slack.SigningSecret = "${SUB_SLACK_SIGN}"
	substituteEnvVars(cfg)

	assert.Equal(t, "xapp-token", cfg.Channels.Slack.AppToken)
	assert.Equal(t, "signing-secret", cfg.Channels.Slack.SigningSecret)
}
