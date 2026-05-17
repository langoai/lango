package status

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/postadjudicationstatus"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/langoai/lango/internal/types"
)

type fakeDeadLetterBridge struct {
	page          DeadLetterListPage
	detail        postadjudicationstatus.TransactionStatus
	detailSeq     []postadjudicationstatus.TransactionStatus
	listErr       error
	detailErr     error
	detailErrSeq  []error
	retryErr      error
	listCalls     int
	detailCalls   int
	retryCalls    int
	lastListOpts  DeadLetterListOptions
	lastDetailID  string
	lastRetryID   string
	lastPrincipal string
}

type commandJSONErrorPayload struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

func (f *fakeDeadLetterBridge) List(_ context.Context, opts DeadLetterListOptions) (DeadLetterListPage, error) {
	f.listCalls++
	f.lastListOpts = opts
	if f.listErr != nil {
		return DeadLetterListPage{}, f.listErr
	}
	return f.page, nil
}

func (f *fakeDeadLetterBridge) Detail(_ context.Context, transactionReceiptID string) (postadjudicationstatus.TransactionStatus, error) {
	f.detailCalls++
	f.lastDetailID = transactionReceiptID
	if idx := f.detailCalls - 1; idx < len(f.detailErrSeq) && f.detailErrSeq[idx] != nil {
		return postadjudicationstatus.TransactionStatus{}, f.detailErrSeq[idx]
	}
	if idx := f.detailCalls - 1; idx < len(f.detailSeq) {
		return f.detailSeq[idx], nil
	}
	if f.detailErr != nil {
		return postadjudicationstatus.TransactionStatus{}, f.detailErr
	}
	return f.detail, nil
}

func (f *fakeDeadLetterBridge) Retry(ctx context.Context, transactionReceiptID string) error {
	f.retryCalls++
	f.lastRetryID = transactionReceiptID
	f.lastPrincipal = ctxkeys.PrincipalFromContext(ctx)
	if f.retryErr != nil {
		return f.retryErr
	}
	return nil
}

func executeCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func executeCommandWithInput(t *testing.T, cmd *cobra.Command, input string, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(strings.NewReader(input))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func configForStatusServer(t *testing.T, rawURL string) *config.Config {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)
	host, portText, err := net.SplitHostPort(parsed.Host)
	require.NoError(t, err)
	port, err := strconv.Atoi(portText)
	require.NoError(t, err)

	cfg := config.DefaultConfig()
	cfg.Server.Host = host
	cfg.Server.Port = port
	return cfg
}

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
		Provenance:          config.ProvenanceConfig{Enabled: true},
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

func TestRenderDashboard_SanitizesDisplayText(t *testing.T) {
	info := StatusInfo{
		Version:  "1.0.0",
		Profile:  "de\x1b[31mfault\n",
		Gateway:  "http://local\x1b[31mhost\n:18789",
		Provider: "open\x1b[31mai\n",
		Model:    "gpt\x1b[31m-5\n",
		Features: []FeatureInfo{
			{Name: "Know\x1b[31mledge\n", Enabled: true, Detail: "emb\x1b[31medded\n"},
		},
		Channels: []string{"tele\x1b[31mgram\nops"},
	}

	output := renderDashboard(info)
	assert.Contains(t, output, "openai (gpt-5)")
	assert.Contains(t, output, "Knowledge (embedded)")
	assert.Contains(t, output, "telegram ops")
	assert.NotContains(t, output, "\x1b")
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

func TestCollectStatus_SanitizesCollectedModelText(t *testing.T) {
	cfg := &config.Config{
		Server:         config.ServerConfig{Host: "local\x1b[31mhost\n", Port: 8080},
		Agent:          config.AgentConfig{Provider: "open\x1b[31mai\n", Model: "gpt\x1b[31m-5\n"},
		ContextProfile: config.ContextProfileName("bal\x1b[31manced\n"),
		Embedding:      config.EmbeddingConfig{Provider: "emb\x1b[31med\n"},
		Channels: config.ChannelsConfig{
			Telegram: config.TelegramConfig{Enabled: true},
		},
	}

	info := collectStatus(cfg, "pro\x1b[31md\n", "http://localhost:1")
	assert.Equal(t, "prod", info.Profile)
	assert.Equal(t, "balanced", info.ContextProfile)
	assert.Equal(t, "http://localhost:8080", info.Gateway)
	assert.Equal(t, "openai", info.Provider)
	assert.Equal(t, "gpt-5", info.Model)
	assert.Equal(t, []string{"telegram"}, info.Channels)
	assert.Contains(t, info.Features[1].Detail, "embed")
	assert.NotContains(t, info.Provider, "\x1b")
}

func TestCollectStatus_SanitizesLiveServerInfoFeatures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := struct {
			Features []types.FeatureStatus `json:"features"`
		}{
			Features: []types.FeatureStatus{{
				Name:       "Embed\x1b[31mding\n",
				Enabled:    false,
				Healthy:    false,
				Reason:     "no\x1b[31m provider\nconfigured",
				Suggestion: "set\x1b[31m provider\nfirst",
			}},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	info := collectStatus(config.DefaultConfig(), "default", server.URL)

	require.NotNil(t, info.ServerInfo)
	require.Len(t, info.ServerInfo.Features, 1)
	assert.Equal(t, "Embedding", info.ServerInfo.Features[0].Name)
	assert.Equal(t, "no provider configured", info.ServerInfo.Features[0].Reason)
	assert.Equal(t, "set provider first", info.ServerInfo.Features[0].Suggestion)
	assert.NotContains(t, info.ServerInfo.Features[0].Name, "\x1b")
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

func TestNewStatusCmd_WiresDeadLetterSummaryCommand(t *testing.T) {
	cmd := NewStatusCmd(func() (*bootstrap.Result, error) {
		return nil, errors.New("should not bootstrap for wiring test")
	}, func() (DeadLetterBridge, func(), error) {
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	names := make([]string, 0, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
	}
	assert.Contains(t, names, "dead-letter-summary")
}

func TestNewStatusCmd_TableWritesToCommandOutput(t *testing.T) {
	cmd := NewStatusCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:      config.DefaultConfig(),
			ProfileName: "test-profile",
		}, nil
	}, func() (DeadLetterBridge, func(), error) {
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	out, err := executeCommand(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, out, "Lango Status")
	assert.Contains(t, out, "test-profile")
}

func TestNewStatusCmd_JSONSanitizesVersion(t *testing.T) {
	tui.SetVersionInfo("1.2.\x1b[31m3\n", "2026-01-01")
	defer tui.SetVersionInfo("dev", "unknown")

	cfg := config.DefaultConfig()
	cfg.Server.Host = "127.0.0.1"
	cfg.Server.Port = 1
	cmd := NewStatusCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:      cfg,
			ProfileName: "json-profile",
		}, nil
	}, func() (DeadLetterBridge, func(), error) {
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "--output", "json")

	require.NoError(t, err)
	var got StatusInfo
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Equal(t, "1.2.3", got.Version)
	assert.NotContains(t, out, "\u001b")
	assert.NotContains(t, out, "\\u001b")
}

func TestNewStatusCmd_UsesConfiguredGatewayForProbeWhenAddrOmitted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/health", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"features": []any{}}))
	}))
	defer server.Close()

	cfg := configForStatusServer(t, server.URL)
	cmd := NewStatusCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{
			Config:      cfg,
			ProfileName: "configured-profile",
		}, nil
	}, func() (DeadLetterBridge, func(), error) {
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "--output", "json")

	require.NoError(t, err)
	var got StatusInfo
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.True(t, got.ServerUp)
	assert.Equal(t, server.URL, got.Gateway)
}

func TestNewStatusCmd_InvalidOutputRejectsBeforeBootstrap(t *testing.T) {
	bootCalls := 0
	cmd := NewStatusCmd(func() (*bootstrap.Result, error) {
		bootCalls++
		return nil, errors.New("boot loader must not run for invalid output")
	}, func() (DeadLetterBridge, func(), error) {
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "--output", "yaml")

	require.Error(t, err)
	assert.ErrorContains(t, err, `unknown output format "yaml"`)
	assert.Equal(t, 0, bootCalls)
}

func TestNewStatusCmd_DeadLetterCommandsUseInjectedLoader(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "summary", args: []string{"dead-letter-summary", "--output", "json"}},
		{name: "list", args: []string{"dead-letters", "--output", "json"}},
		{name: "detail", args: []string{"dead-letter", "tx-1", "--output", "json"}},
		{name: "retry", args: []string{"dead-letter", "retry", "tx-1", "--yes", "--output", "json"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := &fakeDeadLetterBridge{
				page: DeadLetterListPage{
					Entries: []postadjudicationstatus.DeadLetterBacklogEntry{{TransactionReceiptID: "tx-1"}},
					Count:   1,
					Total:   1,
				},
				detail: postadjudicationstatus.TransactionStatus{
					CanRetry: true,
				},
			}
			loadCalls := 0
			cleanupCalls := 0
			cmd := NewStatusCmd(func() (*bootstrap.Result, error) {
				return nil, errors.New("boot loader must not be used by dead-letter subcommands")
			}, func() (DeadLetterBridge, func(), error) {
				loadCalls++
				return bridge, func() { cleanupCalls++ }, nil
			})

			_, err := executeCommand(t, cmd, tt.args...)

			require.NoError(t, err)
			assert.Equal(t, 1, loadCalls)
			assert.Equal(t, 1, cleanupCalls)
		})
	}
}

func TestNewStatusCmd_DeadLettersInvalidOutputRejectsBeforeBridgeLoad(t *testing.T) {
	loadCalls := 0
	cmd := NewStatusCmd(func() (*bootstrap.Result, error) {
		return nil, errors.New("boot loader must not be used by dead-letter subcommands")
	}, func() (DeadLetterBridge, func(), error) {
		loadCalls++
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "dead-letters", "--output", "yaml")

	require.Error(t, err)
	assert.ErrorContains(t, err, `unknown output format "yaml"`)
	assert.Equal(t, 0, loadCalls)
}

func TestNewStatusCmd_DeadLetterLoaderNilCleanupIsNoop(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: DeadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{{TransactionReceiptID: "tx-1"}},
			Count:   1,
			Total:   1,
		},
	}
	cmd := NewStatusCmd(func() (*bootstrap.Result, error) {
		return nil, errors.New("boot loader must not be used by dead-letter subcommands")
	}, func() (DeadLetterBridge, func(), error) {
		return bridge, nil, nil
	})

	_, err := executeCommand(t, cmd, "dead-letter-summary")

	require.NoError(t, err)
	assert.Equal(t, 1, bridge.listCalls)
}

func TestNewStatusCmd_DeadLetterLoaderNilBridgeCleansUp(t *testing.T) {
	cleanupCalls := 0
	cmd := NewStatusCmd(func() (*bootstrap.Result, error) {
		return nil, errors.New("boot loader must not be used by dead-letter subcommands")
	}, func() (DeadLetterBridge, func(), error) {
		return nil, func() { cleanupCalls++ }, nil
	})

	_, err := executeCommand(t, cmd, "dead-letter-summary")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDeadLetterStatusToolsUnavailable)
	assert.Equal(t, 1, cleanupCalls)
}

func TestNewStatusCmd_SanitizesNonJSONErrorsWhilePreservingCause(t *testing.T) {
	boom := errors.New("bo\x1b[31mom\nfailure")
	cmd := NewStatusCmd(func() (*bootstrap.Result, error) {
		return nil, boom
	}, func() (DeadLetterBridge, func(), error) {
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	_, err := executeCommand(t, cmd)

	require.Error(t, err)
	assert.ErrorContains(t, err, "bootstrap: boom failure")
	assert.NotContains(t, err.Error(), "\x1b")
	assert.NotContains(t, err.Error(), "\n")
	assert.ErrorIs(t, err, boom)
}

func TestDeadLetterSummaryCmd_Table(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: DeadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{
				{TransactionReceiptID: "tx-1", Adjudication: "release", CanRetry: true, LatestStatusSubtypeFamily: "retry", LatestDeadLetterReason: "retry attempts exhausted", LatestManualReplayActor: "operator:alice"},
				{TransactionReceiptID: "tx-2", Adjudication: "refund", CanRetry: false, LatestStatusSubtypeFamily: "manual-retry", LatestDeadLetterReason: "policy gate denied replay", LatestManualReplayActor: "system:auto-retry"},
				{TransactionReceiptID: "tx-3", Adjudication: "release", CanRetry: true, LatestStatusSubtypeFamily: "dead-letter", LatestDeadLetterReason: "worker exhausted", LatestManualReplayActor: "service:bridge"},
				{TransactionReceiptID: "tx-4", Adjudication: "refund", CanRetry: false, LatestStatusSubtypeFamily: "dead-letter", LatestDeadLetterReason: "unknown reason", LatestManualReplayActor: "alice"},
			},
			Count: 4,
			Total: 4,
		},
	}
	cmd := newDeadLetterSummaryCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "Dead-Letter Summary")
	assert.Contains(t, out, "Total")
	assert.Contains(t, out, "4")
	assert.Contains(t, out, "Retryable")
	assert.Contains(t, out, "2")
	assert.Contains(t, out, "By Adjudication")
	assert.Contains(t, out, "release")
	assert.Contains(t, out, "refund")
	assert.Contains(t, out, "By Latest Family")
	assert.Contains(t, out, "retry")
	assert.Contains(t, out, "manual-retry")
	assert.Contains(t, out, "dead-letter")
	assert.Contains(t, out, "By reason family")
	assert.Contains(t, out, "retry-exhausted")
	assert.Contains(t, out, "policy-blocked")
	assert.Contains(t, out, "background-failed")
	assert.Contains(t, out, "By actor family")
	assert.Contains(t, out, postadjudicationstatus.ManualReplayActorFamilyOperator)
	assert.Contains(t, out, postadjudicationstatus.ManualReplayActorFamilySystem)
	assert.Contains(t, out, postadjudicationstatus.ManualReplayActorFamilyService)
	assert.Contains(t, out, postadjudicationstatus.ManualReplayActorFamilyUnknown)
	assert.Contains(t, out, "Top Latest Dead-Letter Reasons")
	assert.Contains(t, out, "retry attempts exhausted")
	assert.Contains(t, out, "policy gate denied replay")
	assert.Equal(t, 1, bridge.listCalls)
	assert.Equal(t, DeadLetterListOptions{}, bridge.lastListOpts)
}

func TestDeadLetterSummaryCmd_JSON(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: DeadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{
				{TransactionReceiptID: "tx-1", Adjudication: "release", CanRetry: true, LatestStatusSubtypeFamily: "retry", LatestDeadLetterReason: "worker exhausted", LatestManualReplayActor: "operator:bob", LatestDispatchReference: "dispatch-1"},
				{TransactionReceiptID: "tx-2", Adjudication: "refund", CanRetry: false, LatestStatusSubtypeFamily: "manual-retry", LatestDeadLetterReason: "insufficient evidence", LatestManualReplayActor: "system:auto-retry", LatestDispatchReference: "dispatch-2"},
				{TransactionReceiptID: "tx-3", Adjudication: "release", CanRetry: true, LatestStatusSubtypeFamily: "dead-letter", LatestDeadLetterReason: "worker exhausted", LatestManualReplayActor: "service:bridge", LatestDispatchReference: "dispatch-1"},
				{TransactionReceiptID: "tx-4", Adjudication: "refund", CanRetry: false, LatestStatusSubtypeFamily: "dead-letter", LatestDeadLetterReason: "policy blocked", LatestManualReplayActor: "alice", LatestDispatchReference: "dispatch-3"},
			},
			Count: 4,
			Total: 4,
		},
	}
	cmd := newDeadLetterSummaryCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "--output", "json")
	require.NoError(t, err)

	var got deadLetterSummaryResult
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Equal(t, 4, got.TotalDeadLetters)
	assert.Equal(t, 2, got.RetryableCount)
	assert.Equal(t, []deadLetterSummaryBucket{
		{Label: "release", Count: 2},
		{Label: "refund", Count: 2},
	}, got.ByAdjudication)
	assert.Equal(t, []deadLetterSummaryBucket{
		{Label: "retry", Count: 1},
		{Label: "manual-retry", Count: 1},
		{Label: "dead-letter", Count: 2},
	}, got.ByLatestFamily)
	assert.Equal(t, []deadLetterSummaryBucket{
		{Label: postadjudicationstatus.DeadLetterReasonFamilyPolicyBlocked, Count: 1},
		{Label: postadjudicationstatus.DeadLetterReasonFamilyBackgroundFailed, Count: 2},
		{Label: postadjudicationstatus.DeadLetterReasonFamilyUnknown, Count: 1},
	}, got.ByReasonFamily)
	assert.Equal(t, []deadLetterSummaryBucket{
		{Label: postadjudicationstatus.ManualReplayActorFamilyOperator, Count: 1},
		{Label: postadjudicationstatus.ManualReplayActorFamilySystem, Count: 1},
		{Label: postadjudicationstatus.ManualReplayActorFamilyService, Count: 1},
		{Label: postadjudicationstatus.ManualReplayActorFamilyUnknown, Count: 1},
	}, got.ByActorFamily)
	assert.Equal(t, []deadLetterReasonSummaryItem{
		{Reason: "worker exhausted", Count: 2},
		{Reason: "insufficient evidence", Count: 1},
		{Reason: "policy blocked", Count: 1},
	}, got.TopLatestDeadLetterReasons)
	assert.Equal(t, []deadLetterActorSummaryItem{
		{Actor: "alice", Count: 1},
		{Actor: "operator:bob", Count: 1},
		{Actor: "service:bridge", Count: 1},
		{Actor: "system:auto-retry", Count: 1},
	}, got.TopLatestManualReplayActors)
	assert.Equal(t, []deadLetterDispatchSummaryItem{
		{DispatchReference: "dispatch-1", Count: 2},
		{DispatchReference: "dispatch-2", Count: 1},
		{DispatchReference: "dispatch-3", Count: 1},
	}, got.TopLatestDispatchReferences)

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(out), &raw))
	assert.Contains(t, raw, "by_actor_family")
	assert.Contains(t, raw, "by_reason_family")
	assert.Contains(t, raw, "top_latest_dead_letter_reasons")
}

func TestDeadLetterSummaryCmd_JSONSanitizesSummaryLabels(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: DeadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{
				{
					TransactionReceiptID:      "tx-1",
					Adjudication:              "rele\x1b[31mase\n",
					CanRetry:                  true,
					LatestStatusSubtypeFamily: "dead-\x1b[31mletter\n",
					LatestDeadLetterReason:    "worker\x1b[31m exhausted\n",
					LatestManualReplayActor:   "operator:\x1b[31mbob\n",
					LatestDispatchReference:   "dispatch-\x1b[31m1\n",
					LatestDeadLetteredAt:      "2026-04-26T11:30:00Z",
				},
			},
			Count: 1,
			Total: 1,
		},
	}
	cmd := newDeadLetterSummaryCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "--output", "json")
	require.NoError(t, err)

	var got deadLetterSummaryResult
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	require.Equal(t, []deadLetterSummaryBucket{{Label: "release", Count: 1}}, got.ByAdjudication)
	require.Equal(t, []deadLetterSummaryBucket{{Label: "dead-letter", Count: 1}}, got.ByLatestFamily)
	require.Equal(t, []deadLetterReasonSummaryItem{{Reason: "worker exhausted", Count: 1}}, got.TopLatestDeadLetterReasons)
	require.Equal(t, []deadLetterActorSummaryItem{{Actor: "operator:bob", Count: 1}}, got.TopLatestManualReplayActors)
	require.Equal(t, []deadLetterDispatchSummaryItem{{DispatchReference: "dispatch-1", Count: 1}}, got.TopLatestDispatchReferences)

	raw := out
	assert.NotContains(t, raw, "\u001b")
	assert.NotContains(t, raw, "\\u001b")
	assert.NotContains(t, raw, "\\n")
}

func TestAggregateDeadLetterSummary_ByReasonFamily(t *testing.T) {
	page := DeadLetterListPage{
		Entries: []postadjudicationstatus.DeadLetterBacklogEntry{
			{LatestDeadLetterReason: "retry attempts exhausted after 5 attempts"},
			{LatestDeadLetterReason: "policy gate denied replay"},
			{LatestDeadLetterReason: "invalid transaction receipt evidence"},
			{LatestDeadLetterReason: "background dispatch worker failed"},
			{LatestDeadLetterReason: "unexpected storage condition"},
			{LatestDeadLetterReason: "POLICY BLOCKED BY GATE"},
		},
	}

	got := aggregateDeadLetterSummary(page)

	assert.Equal(t, []deadLetterSummaryBucket{
		{Label: postadjudicationstatus.DeadLetterReasonFamilyRetryExhausted, Count: 1},
		{Label: postadjudicationstatus.DeadLetterReasonFamilyPolicyBlocked, Count: 2},
		{Label: postadjudicationstatus.DeadLetterReasonFamilyReceiptInvalid, Count: 1},
		{Label: postadjudicationstatus.DeadLetterReasonFamilyBackgroundFailed, Count: 1},
		{Label: postadjudicationstatus.DeadLetterReasonFamilyUnknown, Count: 1},
	}, got.ByReasonFamily)
	assert.Equal(t, []deadLetterReasonSummaryItem{
		{Reason: "POLICY BLOCKED BY GATE", Count: 1},
		{Reason: "background dispatch worker failed", Count: 1},
		{Reason: "invalid transaction receipt evidence", Count: 1},
		{Reason: "policy gate denied replay", Count: 1},
		{Reason: "retry attempts exhausted after 5 attempts", Count: 1},
	}, got.TopLatestDeadLetterReasons)
}

func TestAggregateDeadLetterSummary_ByActorFamily(t *testing.T) {
	page := DeadLetterListPage{
		Entries: []postadjudicationstatus.DeadLetterBacklogEntry{
			{LatestManualReplayActor: "operator:alice"},
			{LatestManualReplayActor: "system:auto-retry"},
			{LatestManualReplayActor: "service:bridge"},
			{LatestManualReplayActor: "alice"},
			{LatestManualReplayActor: "OPERATOR:BOB"},
			{LatestManualReplayActor: "runtime:worker"},
			{LatestManualReplayActor: "integration:webhook"},
			{LatestManualReplayActor: "  "},
		},
	}

	got := aggregateDeadLetterSummary(page)

	assert.Equal(t, []deadLetterSummaryBucket{
		{Label: postadjudicationstatus.ManualReplayActorFamilyOperator, Count: 2},
		{Label: postadjudicationstatus.ManualReplayActorFamilySystem, Count: 2},
		{Label: postadjudicationstatus.ManualReplayActorFamilyService, Count: 2},
		{Label: postadjudicationstatus.ManualReplayActorFamilyUnknown, Count: 2},
	}, got.ByActorFamily)
}

func TestAggregateDeadLetterSummary_TopLatestDeadLetterReasons(t *testing.T) {
	page := DeadLetterListPage{
		Entries: []postadjudicationstatus.DeadLetterBacklogEntry{
			{LatestDeadLetterReason: "worker exhausted"},
			{LatestDeadLetterReason: "worker exhausted"},
			{LatestDeadLetterReason: "insufficient evidence"},
			{LatestDeadLetterReason: "missing signature"},
			{LatestDeadLetterReason: "dispatch timeout"},
			{LatestDeadLetterReason: "duplicate evidence"},
			{LatestDeadLetterReason: "already settled"},
			{LatestDeadLetterReason: "duplicate evidence"},
			{LatestDeadLetterReason: "missing signature"},
			{LatestDeadLetterReason: "dispatch timeout"},
			{LatestDeadLetterReason: "  "},
		},
	}

	got := aggregateDeadLetterSummary(page)

	assert.Equal(t, []deadLetterReasonSummaryItem{
		{Reason: "dispatch timeout", Count: 2},
		{Reason: "duplicate evidence", Count: 2},
		{Reason: "missing signature", Count: 2},
		{Reason: "worker exhausted", Count: 2},
		{Reason: "already settled", Count: 1},
	}, got.TopLatestDeadLetterReasons)
}

func TestDeadLetterSummaryCmd_TableIncludesTopLatestManualReplayActors(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: DeadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{
				{TransactionReceiptID: "tx-1", Adjudication: "release", CanRetry: true, LatestStatusSubtypeFamily: "retry", LatestDeadLetterReason: "worker exhausted", LatestManualReplayActor: "operator:bob", LatestDispatchReference: "dispatch-1"},
				{TransactionReceiptID: "tx-2", Adjudication: "refund", CanRetry: false, LatestStatusSubtypeFamily: "manual-retry", LatestDeadLetterReason: "insufficient evidence", LatestManualReplayActor: "operator:alice", LatestDispatchReference: "dispatch-2"},
				{TransactionReceiptID: "tx-3", Adjudication: "release", CanRetry: true, LatestStatusSubtypeFamily: "dead-letter", LatestDeadLetterReason: "worker exhausted", LatestManualReplayActor: "operator:bob", LatestDispatchReference: "dispatch-1"},
			},
			Count: 3,
			Total: 3,
		},
	}
	cmd := newDeadLetterSummaryCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "Top Latest Manual Replay Actors")
	assert.Contains(t, out, "operator:bob")
	assert.Contains(t, out, "operator:alice")
}

func TestDeadLetterSummaryCmd_TableIncludesTopLatestDispatchReferences(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: DeadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{
				{TransactionReceiptID: "tx-1", LatestDispatchReference: "dispatch-1"},
				{TransactionReceiptID: "tx-2", LatestDispatchReference: "dispatch-2"},
				{TransactionReceiptID: "tx-3", LatestDispatchReference: "dispatch-1"},
			},
			Count: 3,
			Total: 3,
		},
	}
	cmd := newDeadLetterSummaryCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "Top Latest Dispatch References")
	assert.Contains(t, out, "dispatch-1")
	assert.Contains(t, out, "dispatch-2")
}

func TestAggregateDeadLetterSummary_TopLatestManualReplayActors(t *testing.T) {
	page := DeadLetterListPage{
		Entries: []postadjudicationstatus.DeadLetterBacklogEntry{
			{LatestManualReplayActor: "operator:bob"},
			{LatestManualReplayActor: "operator:bob"},
			{LatestManualReplayActor: "operator:alice"},
			{LatestManualReplayActor: "operator:alice"},
			{LatestManualReplayActor: "operator:carol"},
			{LatestManualReplayActor: "operator:dave"},
			{LatestManualReplayActor: "operator:erin"},
			{LatestManualReplayActor: "operator:frank"},
			{LatestManualReplayActor: "operator:carol"},
			{LatestManualReplayActor: "  "},
		},
	}

	got := aggregateDeadLetterSummary(page)

	assert.Equal(t, []deadLetterActorSummaryItem{
		{Actor: "operator:alice", Count: 2},
		{Actor: "operator:bob", Count: 2},
		{Actor: "operator:carol", Count: 2},
		{Actor: "operator:dave", Count: 1},
		{Actor: "operator:erin", Count: 1},
	}, got.TopLatestManualReplayActors)
}

func TestAggregateDeadLetterSummary_TopLatestDispatchReferences(t *testing.T) {
	page := DeadLetterListPage{
		Entries: []postadjudicationstatus.DeadLetterBacklogEntry{
			{LatestDispatchReference: "dispatch-b"},
			{LatestDispatchReference: "dispatch-b"},
			{LatestDispatchReference: "dispatch-a"},
			{LatestDispatchReference: "dispatch-a"},
			{LatestDispatchReference: "dispatch-c"},
			{LatestDispatchReference: "dispatch-d"},
			{LatestDispatchReference: "dispatch-e"},
			{LatestDispatchReference: "dispatch-f"},
			{LatestDispatchReference: "dispatch-c"},
			{LatestDispatchReference: "  "},
		},
	}

	got := aggregateDeadLetterSummary(page)

	assert.Equal(t, []deadLetterDispatchSummaryItem{
		{DispatchReference: "dispatch-a", Count: 2},
		{DispatchReference: "dispatch-b", Count: 2},
		{DispatchReference: "dispatch-c", Count: 2},
		{DispatchReference: "dispatch-d", Count: 1},
		{DispatchReference: "dispatch-e", Count: 1},
	}, got.TopLatestDispatchReferences)
}

func TestAggregateDeadLetterSummary_WithDispatchFamiliesAndTrend(t *testing.T) {
	now := time.Date(2026, 4, 26, 12, 0, 0, 0, time.UTC)
	page := DeadLetterListPage{
		Entries: []postadjudicationstatus.DeadLetterBacklogEntry{
			{
				LatestDispatchReference: "dispatch-final-1",
				LatestDeadLetterReason:  "worker exhausted",
				LatestManualReplayActor: "operator:alice",
				LatestDeadLetteredAt:    "2026-04-26T11:30:00Z",
			},
			{
				LatestDispatchReference: "queue/retry/2",
				LatestDeadLetterReason:  "worker exhausted",
				LatestManualReplayActor: "operator:bob",
				LatestDeadLetteredAt:    "2026-04-26T05:45:00Z",
			},
			{
				LatestDispatchReference: "dispatch-final-2",
				LatestDeadLetterReason:  "policy blocked",
				LatestManualReplayActor: "service:bridge",
				LatestDeadLetteredAt:    "2026-04-25T23:30:00Z",
			},
			{
				LatestDispatchReference: "bridge:dead-letter:1",
				LatestDeadLetterReason:  "receipt invalid",
				LatestManualReplayActor: "system:auto-retry",
				LatestDeadLetteredAt:    "2026-04-24T23:30:00Z",
			},
		},
		Count: 4,
		Total: 4,
	}

	got := aggregateDeadLetterSummaryWithOptions(page, deadLetterSummaryOptions{
		TopN:        2,
		TrendWindow: 24 * time.Hour,
		TrendBucket: 12 * time.Hour,
		Now:         now,
	})

	assert.Equal(t, 2, got.TopLimit)
	assert.Equal(t, []deadLetterSummaryBucket{
		{Label: postadjudicationstatus.DispatchReferenceFamilyDispatch, Count: 2},
		{Label: postadjudicationstatus.DispatchReferenceFamilyQueue, Count: 1},
		{Label: postadjudicationstatus.DispatchReferenceFamilyBridge, Count: 1},
	}, got.ByDispatchFamily)
	assert.Equal(t, []deadLetterDispatchSummaryItem{
		{DispatchReference: "bridge:dead-letter:1", Count: 1},
		{DispatchReference: "dispatch-final-1", Count: 1},
	}, got.TopLatestDispatchReferences)
	assert.Equal(t, deadLetterTrendWindow{
		Window:        "24h0m0s",
		Bucket:        "12h0m0s",
		WindowedCount: 3,
		Buckets: []deadLetterTrendBucket{
			{Label: "2026-04-25T12:00:00Z -> 2026-04-26T00:00:00Z", Count: 1},
			{Label: "2026-04-26T00:00:00Z -> 2026-04-26T12:00:00Z", Count: 2},
		},
	}, got.RecentDeadLetterTrend)
}

func TestDeadLettersCmd_TableAndFilters(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: DeadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{
				{
					TransactionReceiptID:   "tx-1",
					LatestDeadLetterReason: "worker exhausted",
					Adjudication:           "release",
					LatestRetryAttempt:     3,
					CanRetry:               true,
				},
			},
			Count:  1,
			Total:  1,
			Offset: 0,
			Limit:  0,
		},
	}
	cmd := newDeadLettersCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "--query", "tx-1", "--adjudication", "release")
	require.NoError(t, err)
	assert.Contains(t, out, "Dead-Letter Backlog")
	assert.Contains(t, out, "tx-1")
	assert.Contains(t, out, "worker exhausted")
	assert.Equal(t, 1, bridge.listCalls)
	assert.Equal(t, DeadLetterListOptions{Query: "tx-1", Adjudication: "release"}, bridge.lastListOpts)
}

func TestDeadLettersCmd_ForwardsSubtypeAndFamilyFilters(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: DeadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{{TransactionReceiptID: "tx-1"}},
			Count:   1,
			Total:   1,
		},
	}
	cmd := newDeadLettersCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	_, err := executeCommand(
		t,
		cmd,
		"--latest-status-subtype", "manual-retry-requested",
		"--latest-status-subtype-family", "manual-retry",
	)
	require.NoError(t, err)
	assert.Equal(t, 1, bridge.listCalls)
	assert.Equal(t, DeadLetterListOptions{
		LatestStatusSubtype:       "manual-retry-requested",
		LatestStatusSubtypeFamily: "manual-retry",
	}, bridge.lastListOpts)
}

func TestDeadLettersCmd_ForwardsAnyMatchFamily(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: DeadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{{TransactionReceiptID: "tx-1"}},
			Count:   1,
			Total:   1,
		},
	}
	cmd := newDeadLettersCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "--any-match-family", "manual-retry")
	require.NoError(t, err)
	assert.Equal(t, 1, bridge.listCalls)
	assert.Equal(t, DeadLetterListOptions{AnyMatchFamily: "manual-retry"}, bridge.lastListOpts)
}

func TestDeadLettersCmd_ForwardsActorAndTimeFilters(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: DeadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{{TransactionReceiptID: "tx-1"}},
			Count:   1,
			Total:   1,
		},
	}
	cmd := newDeadLettersCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	_, err := executeCommand(
		t,
		cmd,
		"--manual-replay-actor", "operator-1",
		"--dead-lettered-after", "2026-04-25T09:00:00Z",
		"--dead-lettered-before", "2026-04-25T18:00:00Z",
	)
	require.NoError(t, err)
	assert.Equal(t, 1, bridge.listCalls)
	assert.Equal(t, DeadLetterListOptions{
		ManualReplayActor:  "operator-1",
		DeadLetteredAfter:  "2026-04-25T09:00:00Z",
		DeadLetteredBefore: "2026-04-25T18:00:00Z",
	}, bridge.lastListOpts)
}

func TestDeadLettersCmd_ForwardsReasonAndDispatchFilters(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: DeadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{{TransactionReceiptID: "tx-1"}},
			Count:   1,
			Total:   1,
		},
	}
	cmd := newDeadLettersCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	_, err := executeCommand(
		t,
		cmd,
		"--dead-letter-reason-query", "worker exhausted",
		"--latest-dispatch-reference", "dispatch-7",
	)
	require.NoError(t, err)
	assert.Equal(t, 1, bridge.listCalls)
	assert.Equal(t, DeadLetterListOptions{
		DeadLetterReasonQuery:   "worker exhausted",
		LatestDispatchReference: "dispatch-7",
	}, bridge.lastListOpts)
}

func TestDeadLettersCmd_ForwardsPagination(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: DeadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{{TransactionReceiptID: "tx-1"}},
			Count:   1,
			Total:   10,
			Offset:  4,
			Limit:   2,
		},
	}
	cmd := newDeadLettersCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "--offset", "4", "--limit", "2")
	require.NoError(t, err)
	assert.Equal(t, DeadLetterListOptions{
		Offset: 4,
		Limit:  2,
	}, bridge.lastListOpts)
}

func TestDeadLettersCmd_RejectsInvalidSubtype(t *testing.T) {
	loaderCalls := 0
	cmd := newDeadLettersCmd(func() (DeadLetterBridge, func(), error) {
		loaderCalls++
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "--latest-status-subtype", "unk\x1b[31mnown\n")
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid --latest-status-subtype")
	assert.ErrorContains(t, err, "unknown")
	assert.NotContains(t, err.Error(), "\x1b")
	assert.NotContains(t, err.Error(), "\n")
	assert.Equal(t, 0, loaderCalls)
}

func TestDeadLettersCmd_RejectsInvalidSubtypeFamily(t *testing.T) {
	loaderCalls := 0
	cmd := newDeadLettersCmd(func() (DeadLetterBridge, func(), error) {
		loaderCalls++
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "--latest-status-subtype-family", "term\x1b[31minal\n")
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid --latest-status-subtype-family")
	assert.ErrorContains(t, err, "terminal")
	assert.NotContains(t, err.Error(), "\x1b")
	assert.NotContains(t, err.Error(), "\n")
	assert.Equal(t, 0, loaderCalls)
}

func TestDeadLettersCmd_RejectsInvalidAnyMatchFamily(t *testing.T) {
	loaderCalls := 0
	cmd := newDeadLettersCmd(func() (DeadLetterBridge, func(), error) {
		loaderCalls++
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "--any-match-family", "term\x1b[31minal\n")
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid --any-match-family")
	assert.ErrorContains(t, err, "terminal")
	assert.NotContains(t, err.Error(), "\x1b")
	assert.NotContains(t, err.Error(), "\n")
	assert.Equal(t, 0, loaderCalls)
}

func TestDeadLettersCmd_RejectsInvalidDeadLetteredAfter(t *testing.T) {
	loaderCalls := 0
	cmd := newDeadLettersCmd(func() (DeadLetterBridge, func(), error) {
		loaderCalls++
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "--dead-lettered-after", "not-\x1b[31ma-time\n")
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid --dead-lettered-after")
	assert.ErrorContains(t, err, "not-a-time")
	assert.NotContains(t, err.Error(), "\x1b")
	assert.NotContains(t, err.Error(), "\n")
	assert.Equal(t, 0, loaderCalls)
}

func TestDeadLettersCmd_RejectsInvalidDeadLetteredBefore(t *testing.T) {
	loaderCalls := 0
	cmd := newDeadLettersCmd(func() (DeadLetterBridge, func(), error) {
		loaderCalls++
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "--dead-lettered-before", "not-\x1b[31ma-time\n")
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid --dead-lettered-before")
	assert.ErrorContains(t, err, "not-a-time")
	assert.NotContains(t, err.Error(), "\x1b")
	assert.NotContains(t, err.Error(), "\n")
	assert.Equal(t, 0, loaderCalls)
}

func TestToolCatalogDeadLetterBridge_ReadyRequiresReadToolsOnly(t *testing.T) {
	catalog := toolcatalog.New()
	catalog.Register("status", []*agent.Tool{
		{Name: "list_dead_lettered_post_adjudication_executions"},
		{Name: "get_post_adjudication_execution_status"},
	})

	bridge := NewToolCatalogDeadLetterBridge(catalog)

	assert.True(t, bridge.Ready())
}

func TestToolCatalogDeadLetterBridge_ForwardsSubtypeAndFamilyFilters(t *testing.T) {
	catalog := toolcatalog.New()
	var gotParams map[string]interface{}
	catalog.Register("status", []*agent.Tool{
		{
			Name: "list_dead_lettered_post_adjudication_executions",
			Handler: func(_ context.Context, params map[string]interface{}) (interface{}, error) {
				gotParams = params
				return map[string]interface{}{
					"entries": []map[string]interface{}{},
					"count":   0,
					"total":   0,
					"offset":  0,
					"limit":   0,
				}, nil
			},
		},
	})

	bridge := NewToolCatalogDeadLetterBridge(catalog)
	_, err := bridge.List(context.Background(), DeadLetterListOptions{
		LatestStatusSubtype:       "retry-scheduled",
		LatestStatusSubtypeFamily: "retry",
	})
	require.NoError(t, err)
	require.NotNil(t, gotParams)
	assert.Equal(t, "retry-scheduled", gotParams["latest_status_subtype"])
	assert.Equal(t, "retry", gotParams["latest_status_subtype_family"])
}

func TestToolCatalogDeadLetterBridge_ForwardsActorAndTimeFilters(t *testing.T) {
	catalog := toolcatalog.New()
	var gotParams map[string]interface{}
	catalog.Register("status", []*agent.Tool{
		{
			Name: "list_dead_lettered_post_adjudication_executions",
			Handler: func(_ context.Context, params map[string]interface{}) (interface{}, error) {
				gotParams = params
				return map[string]interface{}{
					"entries": []map[string]interface{}{},
					"count":   0,
					"total":   0,
					"offset":  0,
					"limit":   0,
				}, nil
			},
		},
	})

	bridge := NewToolCatalogDeadLetterBridge(catalog)
	_, err := bridge.List(context.Background(), DeadLetterListOptions{
		ManualReplayActor:  "operator-1",
		DeadLetteredAfter:  "2026-04-25T09:00:00Z",
		DeadLetteredBefore: "2026-04-25T18:00:00Z",
	})
	require.NoError(t, err)
	require.NotNil(t, gotParams)
	assert.Equal(t, "operator-1", gotParams["manual_replay_actor"])
	assert.Equal(t, "2026-04-25T09:00:00Z", gotParams["dead_lettered_after"])
	assert.Equal(t, "2026-04-25T18:00:00Z", gotParams["dead_lettered_before"])
}

func TestToolCatalogDeadLetterBridge_ForwardsAnyMatchFamily(t *testing.T) {
	catalog := toolcatalog.New()
	var gotParams map[string]interface{}
	catalog.Register("status", []*agent.Tool{
		{
			Name: "list_dead_lettered_post_adjudication_executions",
			Handler: func(_ context.Context, params map[string]interface{}) (interface{}, error) {
				gotParams = params
				return map[string]interface{}{
					"entries": []map[string]interface{}{},
					"count":   0,
					"total":   0,
					"offset":  0,
					"limit":   0,
				}, nil
			},
		},
	})

	bridge := NewToolCatalogDeadLetterBridge(catalog)
	_, err := bridge.List(context.Background(), DeadLetterListOptions{
		AnyMatchFamily: "manual-retry",
	})
	require.NoError(t, err)
	require.NotNil(t, gotParams)
	assert.Equal(t, "manual-retry", gotParams["any_match_family"])
}

func TestToolCatalogDeadLetterBridge_ForwardsReasonAndDispatchFilters(t *testing.T) {
	catalog := toolcatalog.New()
	var gotParams map[string]interface{}
	catalog.Register("status", []*agent.Tool{
		{
			Name: "list_dead_lettered_post_adjudication_executions",
			Handler: func(_ context.Context, params map[string]interface{}) (interface{}, error) {
				gotParams = params
				return map[string]interface{}{
					"entries": []map[string]interface{}{},
					"count":   0,
					"total":   0,
					"offset":  0,
					"limit":   0,
				}, nil
			},
		},
	})

	bridge := NewToolCatalogDeadLetterBridge(catalog)
	_, err := bridge.List(context.Background(), DeadLetterListOptions{
		DeadLetterReasonQuery:   "worker exhausted",
		LatestDispatchReference: "dispatch-7",
	})
	require.NoError(t, err)
	require.NotNil(t, gotParams)
	assert.Equal(t, "worker exhausted", gotParams["dead_letter_reason_query"])
	assert.Equal(t, "dispatch-7", gotParams["latest_dispatch_reference"])
}

func TestToolCatalogDeadLetterBridge_ForwardsPagination(t *testing.T) {
	catalog := toolcatalog.New()
	var gotParams map[string]interface{}
	catalog.Register("status", []*agent.Tool{
		{
			Name: "list_dead_lettered_post_adjudication_executions",
			Handler: func(_ context.Context, params map[string]interface{}) (interface{}, error) {
				gotParams = params
				return map[string]interface{}{
					"entries": []map[string]interface{}{},
					"count":   0,
					"total":   0,
					"offset":  4,
					"limit":   2,
				}, nil
			},
		},
	})

	bridge := NewToolCatalogDeadLetterBridge(catalog)
	_, err := bridge.List(context.Background(), DeadLetterListOptions{
		Offset: 4,
		Limit:  2,
	})
	require.NoError(t, err)
	require.NotNil(t, gotParams)
	assert.Equal(t, 4, gotParams["offset"])
	assert.Equal(t, 2, gotParams["limit"])
}

func TestDeadLettersCmd_JSON(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: DeadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{{
				TransactionReceiptID:      "tx-\x1b[31m1\n",
				SubmissionReceiptID:       "sub-\x1b[31m1\n",
				LatestDeadLetterReason:    "worker\x1b[31m exhausted\n",
				LatestManualReplayActor:   "operator:\x1b[31mbob\n",
				LatestStatusSubtype:       "dead-\x1b[31mlettered\n",
				LatestStatusSubtypeFamily: "dead-\x1b[31mletter\n",
				LatestDispatchReference:   "dispatch-\x1b[31m1\n",
				AnyMatchFamilies:          []string{"retry\x1b[31m\n", "manual-retry"},
			}},
			Count: 1,
			Total: 1,
		},
	}
	cmd := newDeadLettersCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "--output", "json")
	require.NoError(t, err)

	var got DeadLetterListPage
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	require.Len(t, got.Entries, 1)
	assert.Equal(t, "tx-1", got.Entries[0].TransactionReceiptID)
	assert.Equal(t, "sub-1", got.Entries[0].SubmissionReceiptID)
	assert.Equal(t, "worker exhausted", got.Entries[0].LatestDeadLetterReason)
	assert.Equal(t, "operator:bob", got.Entries[0].LatestManualReplayActor)
	assert.Equal(t, "dead-lettered", got.Entries[0].LatestStatusSubtype)
	assert.Equal(t, "dead-letter", got.Entries[0].LatestStatusSubtypeFamily)
	assert.Equal(t, "dispatch-1", got.Entries[0].LatestDispatchReference)
	assert.Equal(t, []string{"retry", "manual-retry"}, got.Entries[0].AnyMatchFamilies)
	assert.NotContains(t, out, "\u001b")
	assert.NotContains(t, out, "\\u001b")
}

func TestDeadLettersCmd_JSONError(t *testing.T) {
	bridge := &fakeDeadLetterBridge{listErr: errors.New("backend unavailable")}
	cmd := newDeadLettersCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "--output", "json")
	require.NoError(t, err)

	var payload commandJSONErrorPayload
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Equal(t, "error", payload.Result)
	assert.Contains(t, payload.Error, "backend unavailable")
}

func TestDeadLetterCmd_Table(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		detail: postadjudicationstatus.TransactionStatus{
			CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
				TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
				SubmissionReceipt:  receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"},
			},
			RetryDeadLetterSummary: postadjudicationstatus.RetryDeadLetterSummary{
				LatestDeadLetterReason:  "worker exhausted",
				LatestRetryAttempt:      3,
				LatestDispatchReference: "dispatch-1",
			},
			LatestBackgroundTask: &postadjudicationstatus.BackgroundTaskBridge{
				TaskID:       "task-1",
				Status:       "retrying",
				AttemptCount: 2,
				NextRetryAt:  "2026-04-25T12:00:00Z",
			},
			IsDeadLettered: true,
			CanRetry:       true,
			Adjudication:   "release",
		},
	}
	cmd := newDeadLetterCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "tx-1")
	require.NoError(t, err)
	assert.Contains(t, out, "Dead-Letter Detail")
	assert.Contains(t, out, "tx-1")
	assert.Contains(t, out, "task-1")
	assert.Equal(t, 1, bridge.detailCalls)
	assert.Equal(t, "tx-1", bridge.lastDetailID)
}

func TestDeadLetterCmd_JSON(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		detail: postadjudicationstatus.TransactionStatus{
			CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
				TransactionReceipt: receipts.TransactionReceipt{
					TransactionReceiptID:        "tx-\x1b[31m1\n",
					TransactionID:               "txn-\x1b[31m1\n",
					Counterparty:                "did:\x1b[31mexample:bob\n",
					SettlementProgressionReason: "need\x1b[31m review\n",
					CanonicalDecision:           "rele\x1b[31mase\n",
					EscrowExecutionInput:        &receipts.EscrowExecutionInput{Reason: "fund\x1b[31m escrow\n", TaskID: "task-\x1b[31m7\n"},
				},
				SubmissionReceipt: receipts.SubmissionReceipt{
					SubmissionReceiptID: "sub-\x1b[31m1\n",
					ArtifactLabel:       "arti\x1b[31mfact\n",
				},
				SubmissionEvents: []receipts.ReceiptEvent{{
					SubmissionReceiptID: "sub-\x1b[31m1\n",
					Source:              "queue\x1b[31m\n",
					Subtype:             "retry-\x1b[31mscheduled\n",
					Reason:              "worker\x1b[31m exhausted\n",
				}},
			},
			RetryDeadLetterSummary: postadjudicationstatus.RetryDeadLetterSummary{
				LatestDeadLetterReason:    "policy\x1b[31m blocked\n",
				LatestManualReplayActor:   "operator:\x1b[31mbob\n",
				LatestDispatchReference:   "dispatch-\x1b[31m9\n",
				LatestStatusSubtype:       "dead-\x1b[31mlettered\n",
				LatestStatusSubtypeFamily: "dead-\x1b[31mletter\n",
				AnyMatchFamilies:          []string{"retry\x1b[31m\n", "manual-retry"},
			},
			LatestBackgroundTask: &postadjudicationstatus.BackgroundTaskBridge{
				TaskID:      "task-\x1b[31m1\n",
				Status:      "retry\x1b[31ming\n",
				NextRetryAt: "2026-04-25T12:00:00Z\x1b[31m\n",
			},
			Adjudication: "rele\x1b[31mase\n",
		},
	}
	cmd := newDeadLetterCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "tx-1", "--output", "json")
	require.NoError(t, err)

	var got postadjudicationstatus.TransactionStatus
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Equal(t, "tx-1", got.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID)
	assert.Equal(t, "txn-1", got.CanonicalSnapshot.TransactionReceipt.TransactionID)
	assert.Equal(t, "did:example:bob", got.CanonicalSnapshot.TransactionReceipt.Counterparty)
	assert.Equal(t, "need review", got.CanonicalSnapshot.TransactionReceipt.SettlementProgressionReason)
	assert.Equal(t, "release", got.CanonicalSnapshot.TransactionReceipt.CanonicalDecision)
	require.NotNil(t, got.CanonicalSnapshot.TransactionReceipt.EscrowExecutionInput)
	assert.Equal(t, "fund escrow", got.CanonicalSnapshot.TransactionReceipt.EscrowExecutionInput.Reason)
	assert.Equal(t, "task-7", got.CanonicalSnapshot.TransactionReceipt.EscrowExecutionInput.TaskID)
	assert.Equal(t, "sub-1", got.CanonicalSnapshot.SubmissionReceipt.SubmissionReceiptID)
	assert.Equal(t, "artifact", got.CanonicalSnapshot.SubmissionReceipt.ArtifactLabel)
	require.Len(t, got.CanonicalSnapshot.SubmissionEvents, 1)
	assert.Equal(t, "queue", got.CanonicalSnapshot.SubmissionEvents[0].Source)
	assert.Equal(t, "retry-scheduled", got.CanonicalSnapshot.SubmissionEvents[0].Subtype)
	assert.Equal(t, "worker exhausted", got.CanonicalSnapshot.SubmissionEvents[0].Reason)
	assert.Equal(t, "policy blocked", got.RetryDeadLetterSummary.LatestDeadLetterReason)
	assert.Equal(t, "operator:bob", got.RetryDeadLetterSummary.LatestManualReplayActor)
	assert.Equal(t, "dispatch-9", got.RetryDeadLetterSummary.LatestDispatchReference)
	assert.Equal(t, "dead-lettered", got.RetryDeadLetterSummary.LatestStatusSubtype)
	assert.Equal(t, "dead-letter", got.RetryDeadLetterSummary.LatestStatusSubtypeFamily)
	assert.Equal(t, []string{"retry", "manual-retry"}, got.RetryDeadLetterSummary.AnyMatchFamilies)
	require.NotNil(t, got.LatestBackgroundTask)
	assert.Equal(t, "task-1", got.LatestBackgroundTask.TaskID)
	assert.Equal(t, "retrying", got.LatestBackgroundTask.Status)
	assert.Equal(t, "2026-04-25T12:00:00Z", got.LatestBackgroundTask.NextRetryAt)
	assert.Equal(t, "release", got.Adjudication)
	assert.NotContains(t, out, "\u001b")
	assert.NotContains(t, out, "\\u001b")
}

func TestDeadLetterCmd_PropagatesBridgeErrors(t *testing.T) {
	cmd := newDeadLetterCmd(func() (DeadLetterBridge, func(), error) {
		return &fakeDeadLetterBridge{detailErr: errors.New("boom")}, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "tx-1")
	require.Error(t, err)
	assert.ErrorContains(t, err, "boom")
}

func TestDeadLetterRetryCmd_SucceedsWithYes(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		detailSeq: []postadjudicationstatus.TransactionStatus{
			{
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
				},
				RetryDeadLetterSummary: postadjudicationstatus.RetryDeadLetterSummary{
					LatestStatusSubtype:       "dead-lettered",
					LatestStatusSubtypeFamily: "dead-letter",
					LatestDeadLetterReason:    "worker exhausted",
					LatestRetryAttempt:        3,
				},
				CanRetry:       true,
				IsDeadLettered: true,
			},
			{
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
				},
				RetryDeadLetterSummary: postadjudicationstatus.RetryDeadLetterSummary{
					LatestStatusSubtype:       "retry-scheduled",
					LatestStatusSubtypeFamily: "retry",
					LatestDeadLetterReason:    "worker exhausted",
					LatestRetryAttempt:        4,
					LatestDispatchReference:   "dispatch-2",
				},
				LatestBackgroundTask: &postadjudicationstatus.BackgroundTaskBridge{
					TaskID:       "task-2",
					Status:       "queued",
					AttemptCount: 1,
				},
				CanRetry:       true,
				IsDeadLettered: true,
			},
		},
	}
	cmd := newDeadLetterCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "retry", "tx-1", "--yes")
	require.NoError(t, err)
	assert.Contains(t, out, "Retry request accepted")
	assert.Contains(t, out, "Follow-up")
	assert.Contains(t, out, "retry-scheduled")
	assert.Contains(t, out, "task-2")
	assert.Equal(t, 2, bridge.detailCalls)
	assert.Equal(t, 1, bridge.retryCalls)
	assert.Equal(t, "tx-1", bridge.lastRetryID)
}

func TestDeadLetterRetryCmd_RejectsWhenCannotRetry(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		detail: postadjudicationstatus.TransactionStatus{
			CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
				TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
			},
			CanRetry: false,
		},
	}
	cmd := newDeadLetterCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "retry", "tx-\x1b[31m1\n", "--yes")
	require.Error(t, err)
	assert.ErrorContains(t, err, "retry precheck rejected")
	assert.ErrorContains(t, err, "can_retry=false")
	assert.ErrorContains(t, err, "tx-1")
	assert.NotContains(t, err.Error(), "\x1b")
	assert.NotContains(t, err.Error(), "\n")
	assert.Equal(t, 1, bridge.detailCalls)
	assert.Equal(t, 0, bridge.retryCalls)
}

func TestDeadLetterRetryCmd_RequiresConfirmationByDefault(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		detail: postadjudicationstatus.TransactionStatus{
			CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
				TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
			},
			CanRetry: true,
		},
	}
	cmd := newDeadLetterCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommandWithInput(t, cmd, "n\n", "retry", "tx-\x1b[31m1\n")
	require.NoError(t, err)
	assert.Contains(t, out, "Retry dead-lettered execution")
	assert.Contains(t, out, "tx-1")
	assert.Contains(t, out, "aborted")
	assert.NotContains(t, out, "\x1b")
	assert.Equal(t, 1, bridge.detailCalls)
	assert.Equal(t, 0, bridge.retryCalls)
}

func TestDeadLetterRetryCmd_EOFAbortsWithoutRetry(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		detail: postadjudicationstatus.TransactionStatus{
			CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
				TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
			},
			CanRetry: true,
		},
	}
	cmd := newDeadLetterCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommandWithInput(t, cmd, "", "retry", "tx-1")
	require.NoError(t, err)
	assert.Contains(t, out, "Retry dead-lettered execution")
	assert.Contains(t, out, "aborted")
	assert.Equal(t, 1, bridge.detailCalls)
	assert.Equal(t, 0, bridge.retryCalls)
}

func TestDeadLetterRetryCmd_InvokesRetryAfterConfirm(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		detail: postadjudicationstatus.TransactionStatus{
			CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
				TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
			},
			CanRetry: true,
		},
	}
	cmd := newDeadLetterCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommandWithInput(t, cmd, "y\n", "retry", "tx-1")
	require.NoError(t, err)
	assert.Contains(t, out, "Retry dead-lettered execution")
	assert.Contains(t, out, "Retry request accepted")
	assert.Equal(t, 1, bridge.retryCalls)
	assert.Equal(t, "tx-1", bridge.lastRetryID)
}

func TestDeadLetterRetryCmd_ForwardsExplicitActor(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		detail: postadjudicationstatus.TransactionStatus{
			CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
				TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
			},
			CanRetry: true,
		},
	}
	cmd := newDeadLetterCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "retry", "tx-1", "--yes", "--actor", "operator:ci")
	require.NoError(t, err)
	assert.Equal(t, "operator:ci", bridge.lastPrincipal)
}

func TestDeadLetterRetryCmd_ReportsInvocationFailureSeparately(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		detail: postadjudicationstatus.TransactionStatus{
			CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
				TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
			},
			CanRetry: true,
		},
		retryErr: errors.New("queue unavailable"),
	}
	cmd := newDeadLetterCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "retry", "tx-\x1b[31m1\n", "--yes")
	require.Error(t, err)
	assert.ErrorContains(t, err, "retry request failed")
	assert.ErrorContains(t, err, "queue unavailable")
	assert.ErrorContains(t, err, "tx-1")
	assert.NotContains(t, err.Error(), "\x1b")
	assert.NotContains(t, err.Error(), "\n")
	assert.Equal(t, 1, bridge.detailCalls)
	assert.Equal(t, 1, bridge.retryCalls)
}

func TestDeadLetterRetryCmd_JSONError(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		detail: postadjudicationstatus.TransactionStatus{
			CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
				TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
			},
			CanRetry: true,
		},
		retryErr: errors.New("queue \x1b[31munavailable\nretry request failed"),
	}
	cmd := newDeadLetterCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "retry", "tx-1", "--yes", "--output", "json")
	require.NoError(t, err)

	var payload commandJSONErrorPayload
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Equal(t, "error", payload.Result)
	assert.Contains(t, payload.Error, "retry request failed")
	assert.Contains(t, payload.Error, "queue unavailable")
	assert.NotContains(t, payload.Error, "\x1b")
	assert.NotContains(t, payload.Error, "\n")
}

func TestDeadLetterRetryCmd_JSONReportsAcceptedRequest(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		detailSeq: []postadjudicationstatus.TransactionStatus{
			{
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
				},
				RetryDeadLetterSummary: postadjudicationstatus.RetryDeadLetterSummary{
					LatestStatusSubtype:       "dead-lettered",
					LatestStatusSubtypeFamily: "dead-letter",
				},
				CanRetry:       true,
				IsDeadLettered: true,
			},
			{
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
				},
				RetryDeadLetterSummary: postadjudicationstatus.RetryDeadLetterSummary{
					LatestStatusSubtype:       "retry-\x1b[31mscheduled\n",
					LatestStatusSubtypeFamily: "re\x1b[31mtry\n",
					LatestRetryAttempt:        4,
					LatestDispatchReference:   "dispatch-\x1b[31m2\n",
				},
				LatestBackgroundTask: &postadjudicationstatus.BackgroundTaskBridge{
					TaskID:       "task-\x1b[31m2\n",
					Status:       "que\x1b[31mued\n",
					AttemptCount: 1,
				},
				CanRetry:       true,
				IsDeadLettered: true,
			},
		},
	}
	cmd := newDeadLetterCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "retry", "tx-1", "--yes", "--output", "json")
	require.NoError(t, err)

	var got deadLetterRetryResult
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Equal(t, "tx-1", got.TransactionReceiptID)
	assert.Equal(t, "accepted", got.Result)
	assert.Equal(t, "Retry request accepted for transaction tx-1.", got.Message)
	require.NotNil(t, got.FollowUp)
	assert.Equal(t, "retry-scheduled", got.FollowUp.LatestStatusSubtype)
	assert.Equal(t, "retry", got.FollowUp.LatestStatusSubtypeFamily)
	require.NotNil(t, got.FollowUp.BackgroundTask)
	assert.Equal(t, "task-2", got.FollowUp.BackgroundTask.TaskID)
	assert.Equal(t, "queued", got.FollowUp.BackgroundTask.Status)
	assert.Equal(t, "dispatch-2", got.FollowUp.LatestDispatchReference)
	assert.NotContains(t, got.Message, "\x1b")
	assert.Equal(t, 1, got.PollCount)
	assert.False(t, got.TimedOut)
}

func TestDeadLetterRetryCmd_WaitPollsUntilFollowUpChanges(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		detailSeq: []postadjudicationstatus.TransactionStatus{
			{
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
				},
				RetryDeadLetterSummary: postadjudicationstatus.RetryDeadLetterSummary{
					LatestStatusSubtype:       "dead-lettered",
					LatestStatusSubtypeFamily: "dead-letter",
					LatestRetryAttempt:        3,
				},
				CanRetry:       true,
				IsDeadLettered: true,
			},
			{
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
				},
				RetryDeadLetterSummary: postadjudicationstatus.RetryDeadLetterSummary{
					LatestStatusSubtype:       "dead-lettered",
					LatestStatusSubtypeFamily: "dead-letter",
					LatestRetryAttempt:        3,
				},
				CanRetry:       true,
				IsDeadLettered: true,
			},
			{
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
				},
				RetryDeadLetterSummary: postadjudicationstatus.RetryDeadLetterSummary{
					LatestStatusSubtype:       "retry-scheduled",
					LatestStatusSubtypeFamily: "retry",
					LatestRetryAttempt:        4,
				},
				LatestBackgroundTask: &postadjudicationstatus.BackgroundTaskBridge{
					TaskID:       "task-3",
					Status:       "running",
					AttemptCount: 1,
				},
				CanRetry:       true,
				IsDeadLettered: true,
			},
		},
	}
	cmd := newDeadLetterCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(
		t,
		cmd,
		"retry", "tx-1", "--yes",
		"--wait",
		"--wait-interval", "1ms",
		"--wait-timeout", "50ms",
	)
	require.NoError(t, err)
	assert.Contains(t, out, "Polling follow-up status")
	assert.Contains(t, out, "running")
	assert.Equal(t, 3, bridge.detailCalls)
}
