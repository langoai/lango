package status

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/postadjudicationstatus"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/toolcatalog"
)

type fakeDeadLetterBridge struct {
	page         deadLetterListPage
	detail       postadjudicationstatus.TransactionStatus
	listErr      error
	detailErr    error
	retryErr     error
	listCalls    int
	detailCalls  int
	retryCalls   int
	lastListOpts deadLetterListOptions
	lastDetailID string
	lastRetryID  string
}

func (f *fakeDeadLetterBridge) List(_ context.Context, opts deadLetterListOptions) (deadLetterListPage, error) {
	f.listCalls++
	f.lastListOpts = opts
	if f.listErr != nil {
		return deadLetterListPage{}, f.listErr
	}
	return f.page, nil
}

func (f *fakeDeadLetterBridge) Detail(_ context.Context, transactionReceiptID string) (postadjudicationstatus.TransactionStatus, error) {
	f.detailCalls++
	f.lastDetailID = transactionReceiptID
	if f.detailErr != nil {
		return postadjudicationstatus.TransactionStatus{}, f.detailErr
	}
	return f.detail, nil
}

func (f *fakeDeadLetterBridge) Retry(_ context.Context, transactionReceiptID string) error {
	f.retryCalls++
	f.lastRetryID = transactionReceiptID
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

func TestDeadLettersCmd_TableAndFilters(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: deadLetterListPage{
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
	cmd := newDeadLettersCmd(func() (deadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "--query", "tx-1", "--adjudication", "release")
	require.NoError(t, err)
	assert.Contains(t, out, "Dead-Letter Backlog")
	assert.Contains(t, out, "tx-1")
	assert.Contains(t, out, "worker exhausted")
	assert.Equal(t, 1, bridge.listCalls)
	assert.Equal(t, deadLetterListOptions{Query: "tx-1", Adjudication: "release"}, bridge.lastListOpts)
}

func TestDeadLettersCmd_ForwardsSubtypeAndFamilyFilters(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: deadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{{TransactionReceiptID: "tx-1"}},
			Count:   1,
			Total:   1,
		},
	}
	cmd := newDeadLettersCmd(func() (deadLetterBridge, func(), error) {
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
	assert.Equal(t, deadLetterListOptions{
		LatestStatusSubtype:       "manual-retry-requested",
		LatestStatusSubtypeFamily: "manual-retry",
	}, bridge.lastListOpts)
}

func TestDeadLettersCmd_ForwardsActorAndTimeFilters(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: deadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{{TransactionReceiptID: "tx-1"}},
			Count:   1,
			Total:   1,
		},
	}
	cmd := newDeadLettersCmd(func() (deadLetterBridge, func(), error) {
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
	assert.Equal(t, deadLetterListOptions{
		ManualReplayActor:  "operator-1",
		DeadLetteredAfter:  "2026-04-25T09:00:00Z",
		DeadLetteredBefore: "2026-04-25T18:00:00Z",
	}, bridge.lastListOpts)
}

func TestDeadLettersCmd_ForwardsReasonAndDispatchFilters(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: deadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{{TransactionReceiptID: "tx-1"}},
			Count:   1,
			Total:   1,
		},
	}
	cmd := newDeadLettersCmd(func() (deadLetterBridge, func(), error) {
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
	assert.Equal(t, deadLetterListOptions{
		DeadLetterReasonQuery:   "worker exhausted",
		LatestDispatchReference: "dispatch-7",
	}, bridge.lastListOpts)
}

func TestDeadLettersCmd_RejectsInvalidSubtype(t *testing.T) {
	loaderCalls := 0
	cmd := newDeadLettersCmd(func() (deadLetterBridge, func(), error) {
		loaderCalls++
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "--latest-status-subtype", "unknown")
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid --latest-status-subtype")
	assert.Equal(t, 0, loaderCalls)
}

func TestDeadLettersCmd_RejectsInvalidSubtypeFamily(t *testing.T) {
	loaderCalls := 0
	cmd := newDeadLettersCmd(func() (deadLetterBridge, func(), error) {
		loaderCalls++
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "--latest-status-subtype-family", "terminal")
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid --latest-status-subtype-family")
	assert.Equal(t, 0, loaderCalls)
}

func TestDeadLettersCmd_RejectsInvalidDeadLetteredAfter(t *testing.T) {
	loaderCalls := 0
	cmd := newDeadLettersCmd(func() (deadLetterBridge, func(), error) {
		loaderCalls++
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "--dead-lettered-after", "not-a-time")
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid --dead-lettered-after")
	assert.Equal(t, 0, loaderCalls)
}

func TestDeadLettersCmd_RejectsInvalidDeadLetteredBefore(t *testing.T) {
	loaderCalls := 0
	cmd := newDeadLettersCmd(func() (deadLetterBridge, func(), error) {
		loaderCalls++
		return &fakeDeadLetterBridge{}, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "--dead-lettered-before", "not-a-time")
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid --dead-lettered-before")
	assert.Equal(t, 0, loaderCalls)
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

	bridge := &toolCatalogDeadLetterBridge{catalog: catalog}
	_, err := bridge.List(context.Background(), deadLetterListOptions{
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

	bridge := &toolCatalogDeadLetterBridge{catalog: catalog}
	_, err := bridge.List(context.Background(), deadLetterListOptions{
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

	bridge := &toolCatalogDeadLetterBridge{catalog: catalog}
	_, err := bridge.List(context.Background(), deadLetterListOptions{
		DeadLetterReasonQuery:   "worker exhausted",
		LatestDispatchReference: "dispatch-7",
	})
	require.NoError(t, err)
	require.NotNil(t, gotParams)
	assert.Equal(t, "worker exhausted", gotParams["dead_letter_reason_query"])
	assert.Equal(t, "dispatch-7", gotParams["latest_dispatch_reference"])
}

func TestDeadLettersCmd_JSON(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		page: deadLetterListPage{
			Entries: []postadjudicationstatus.DeadLetterBacklogEntry{{TransactionReceiptID: "tx-1"}},
			Count:   1,
			Total:   1,
		},
	}
	cmd := newDeadLettersCmd(func() (deadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "\"entries\"")
	assert.Contains(t, out, "\"transaction_receipt_id\": \"tx-1\"")
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
	cmd := newDeadLetterCmd(func() (deadLetterBridge, func(), error) {
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
				TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
			},
		},
	}
	cmd := newDeadLetterCmd(func() (deadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "tx-1", "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "\"canonical_snapshot\"")
	assert.Contains(t, out, "\"transaction_receipt_id\": \"tx-1\"")
}

func TestDeadLetterCmd_PropagatesBridgeErrors(t *testing.T) {
	cmd := newDeadLetterCmd(func() (deadLetterBridge, func(), error) {
		return &fakeDeadLetterBridge{detailErr: errors.New("boom")}, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "tx-1")
	require.Error(t, err)
	assert.ErrorContains(t, err, "boom")
}

func TestDeadLetterRetryCmd_SucceedsWithYes(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		detail: postadjudicationstatus.TransactionStatus{
			CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
				TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
			},
			CanRetry: true,
		},
	}
	cmd := newDeadLetterCmd(func() (deadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "retry", "tx-1", "--yes")
	require.NoError(t, err)
	assert.Contains(t, out, "Retry requested")
	assert.Equal(t, 1, bridge.detailCalls)
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
	cmd := newDeadLetterCmd(func() (deadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	_, err := executeCommand(t, cmd, "retry", "tx-1", "--yes")
	require.Error(t, err)
	assert.ErrorContains(t, err, "is not retryable")
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
	cmd := newDeadLetterCmd(func() (deadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommandWithInput(t, cmd, "n\n", "retry", "tx-1")
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
	cmd := newDeadLetterCmd(func() (deadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommandWithInput(t, cmd, "y\n", "retry", "tx-1")
	require.NoError(t, err)
	assert.Contains(t, out, "Retry dead-lettered execution")
	assert.Contains(t, out, "Retry requested")
	assert.Equal(t, 1, bridge.retryCalls)
	assert.Equal(t, "tx-1", bridge.lastRetryID)
}
