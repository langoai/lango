package status

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestCollectStatus_DefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	info := collectStatus(cfg, "default", "http://localhost:1") // unreachable port

	assert.Equal(t, "default", info.Profile)
	assert.False(t, info.ServerUp, "server should not be running against unreachable port")
	assert.Nil(t, info.ServerInfo)
	assert.Empty(t, info.Channels, "default config should have no channels enabled")
	assert.NotEmpty(t, info.Features, "features list should not be empty")
	assert.Contains(t, info.Gateway, "localhost")
}

func TestCollectFeatures_AllEnabled(t *testing.T) {
	cfg := &config.Config{
		Knowledge:           config.KnowledgeConfig{Enabled: true},
		Embedding:           config.EmbeddingConfig{Provider: "openai"},
		Graph:               config.GraphConfig{Enabled: true},
		ObservationalMemory: config.ObservationalMemoryConfig{Enabled: true},
		Librarian:           config.LibrarianConfig{Enabled: true},
		Agent:               config.AgentConfig{MultiAgent: true},
		Cron:                config.CronConfig{Enabled: true},
		Background:          config.BackgroundConfig{Enabled: true},
		Workflow:            config.WorkflowConfig{Enabled: true},
		MCP:                 config.MCPConfig{Enabled: true, Servers: map[string]config.MCPServerConfig{"s1": {}}},
		P2P:                 config.P2PConfig{Enabled: true},
		Payment:             config.PaymentConfig{Enabled: true},
		Economy:             config.EconomyConfig{Enabled: true},
		A2A:                 config.A2AConfig{Enabled: true},
		RunLedger:           config.RunLedgerConfig{Enabled: true, WorkspaceIsolation: true},
	}

	features := collectFeatures(cfg)
	for _, f := range features {
		assert.True(t, f.Enabled, "feature %q should be enabled", f.Name)
	}
}

func TestCollectFeatures_NoneEnabled(t *testing.T) {
	cfg := &config.Config{}

	features := collectFeatures(cfg)
	for _, f := range features {
		assert.False(t, f.Enabled, "feature %q should be disabled", f.Name)
	}
}

func TestMcpDetail(t *testing.T) {
	tests := []struct {
		give *config.Config
		want string
	}{
		{
			give: &config.Config{MCP: config.MCPConfig{Enabled: false}},
			want: "",
		},
		{
			give: &config.Config{MCP: config.MCPConfig{Enabled: true}},
			want: "no servers",
		},
		{
			give: &config.Config{MCP: config.MCPConfig{
				Enabled: true,
				Servers: map[string]config.MCPServerConfig{"a": {}, "b": {}},
			}},
			want: "2 server(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, mcpDetail(tt.give))
		})
	}
}

func TestRenderDashboard_ContainsExpectedSections(t *testing.T) {
	info := StatusInfo{
		Version:  "1.0.0",
		Profile:  "test",
		Gateway:  "http://localhost:18789",
		Provider: "anthropic",
		Model:    "claude-3",
		Features: []FeatureInfo{
			{"Knowledge", true, ""},
			{"Graph", false, ""},
		},
		Channels: []string{"telegram"},
	}

	output := renderDashboard(info)
	assert.Contains(t, output, "System")
	assert.Contains(t, output, "Features")
	assert.Contains(t, output, "Channels")
	assert.Contains(t, output, "1.0.0")
	assert.Contains(t, output, "test")
	assert.Contains(t, output, "Knowledge")
	assert.Contains(t, output, "Graph")
	assert.Contains(t, output, "telegram")
}

func TestRenderDashboard_NoChannels(t *testing.T) {
	info := StatusInfo{
		Version:  "dev",
		Profile:  "default",
		Gateway:  "http://localhost:18789",
		Features: []FeatureInfo{{"Knowledge", false, ""}},
	}

	output := renderDashboard(info)
	assert.Contains(t, output, "System")
	assert.Contains(t, output, "Features")
	assert.NotContains(t, output, "Channels")
}

func TestRenderDashboard_ServerRunning(t *testing.T) {
	info := StatusInfo{
		Version:  "dev",
		Profile:  "default",
		ServerUp: true,
		Gateway:  "http://localhost:18789",
		Features: []FeatureInfo{},
	}

	output := renderDashboard(info)
	assert.Contains(t, output, "running")
}

func TestRenderDashboard_ServerNotRunning(t *testing.T) {
	info := StatusInfo{
		Version:  "dev",
		Profile:  "default",
		ServerUp: false,
		Gateway:  "http://localhost:18789",
		Features: []FeatureInfo{},
	}

	output := renderDashboard(info)
	assert.Contains(t, output, "not running")
}

func TestStatusInfo_JSON(t *testing.T) {
	info := StatusInfo{
		Version:  "1.2.3",
		Profile:  "prod",
		ServerUp: true,
		Gateway:  "http://localhost:18789",
		Provider: "openai",
		Model:    "gpt-4",
		Features: []FeatureInfo{
			{"Knowledge", true, ""},
			{"MCP", true, "2 server(s)"},
		},
		Channels:   []string{"telegram", "discord"},
		ServerInfo: &LiveInfo{Healthy: true},
	}

	data, err := json.Marshal(info)
	require.NoError(t, err)

	var decoded StatusInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, info.Version, decoded.Version)
	assert.Equal(t, info.Profile, decoded.Profile)
	assert.Equal(t, info.ServerUp, decoded.ServerUp)
	assert.Equal(t, info.Provider, decoded.Provider)
	assert.Equal(t, info.Model, decoded.Model)
	assert.Len(t, decoded.Features, 2)
	assert.Len(t, decoded.Channels, 2)
	assert.NotNil(t, decoded.ServerInfo)
	assert.True(t, decoded.ServerInfo.Healthy)
}

func TestCollectStatus_Channels(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Host: "0.0.0.0", Port: 8080},
		Channels: config.ChannelsConfig{
			Telegram: config.TelegramConfig{Enabled: true},
			Discord:  config.DiscordConfig{Enabled: true},
			Slack:    config.SlackConfig{Enabled: true},
		},
	}

	info := collectStatus(cfg, "multi", "http://localhost:1")
	assert.Equal(t, []string{"telegram", "discord", "slack"}, info.Channels)
	assert.Equal(t, "http://0.0.0.0:8080", info.Gateway)
}

func TestRenderDashboard_EmptyVersion(t *testing.T) {
	info := StatusInfo{
		Version:  "",
		Profile:  "default",
		Gateway:  "http://localhost:18789",
		Features: []FeatureInfo{},
	}

	output := renderDashboard(info)
	assert.True(t, strings.Contains(output, "dev"))
}
