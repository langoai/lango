package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/cockpit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/postadjudicationstatus"
	"github.com/langoai/lango/internal/toolcatalog"
)

func TestWave36RunCockpitBootstrapErrorStopsBeforeStartup(t *testing.T) {
	restoreWave19TUIProfile(t)
	origBootLoader := cockpitBootLoaderFn
	origLoggingInit := cockpitLoggingInitFn
	origBuilder := cockpitAppBuilderFn
	t.Cleanup(func() {
		cockpitBootLoaderFn = origBootLoader
		cockpitLoggingInitFn = origLoggingInit
		cockpitAppBuilderFn = origBuilder
	})

	cockpitBootLoaderFn = func() (*bootstrap.Result, error) {
		return nil, errors.New("wave36 boot refused")
	}
	cockpitLoggingInitFn = func(logging.LogConfig) error {
		t.Fatal("logging must not initialize after cockpit bootstrap failure")
		return nil
	}
	cockpitAppBuilderFn = func(*bootstrap.Result, app.AppOption) (*app.App, error) {
		t.Fatal("app builder must not run after cockpit bootstrap failure")
		return nil, nil
	}

	err := runCockpit("")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "bootstrap: wave36 boot refused")
}

func TestWave36PrepareTUIStartupAllowsMissingLogFileParent(t *testing.T) {
	restoreWave19TUIProfile(t)

	cfg := config.DefaultConfig()
	cfg.DataRoot = filepath.Join(t.TempDir(), "missing-parent")
	cfg.Logging.Level = "debug"
	cfg.Logging.Format = "json"

	var gotLogConfig logging.LogConfig
	syncCalled := false
	var notice bytes.Buffer
	cleanup, err := prepareTUIStartup(
		cfg,
		"wave36",
		"missing/chat.log",
		func(logCfg logging.LogConfig) error {
			gotLogConfig = logCfg
			return nil
		},
		func() error {
			syncCalled = true
			return nil
		},
		&notice,
		"  Initializing wave36...",
	)

	require.NoError(t, err)
	require.NotNil(t, cleanup)
	cleanup()

	wantPath := filepath.Join(cfg.DataRoot, "missing/chat.log")
	assert.Equal(t, logging.LogConfig{
		Level:      "debug",
		Format:     "json",
		OutputPath: wantPath,
	}, gotLogConfig)
	assert.True(t, syncCalled)
	assert.Contains(t, notice.String(), wantPath)
	assert.Contains(t, notice.String(), "Initializing wave36...")
}

func TestWave36PrepareTUIStartupCleanupClosesStdlibLogFileAndRestoresWriter(t *testing.T) {
	restoreWave19TUIProfile(t)
	t.Cleanup(func() { log.SetOutput(io.Discard) })

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	var restored bytes.Buffer
	log.SetOutput(&restored)

	cleanup, err := prepareTUIStartup(
		cfg,
		"wave36",
		"chat.log",
		func(logCfg logging.LogConfig) error {
			require.NoError(t, os.MkdirAll(filepath.Dir(logCfg.OutputPath), 0o700))
			return nil
		},
		func() error { return nil },
		io.Discard,
		"  Initializing wave36...",
	)
	require.NoError(t, err)
	require.NotNil(t, cleanup)

	require.NoError(t, log.Output(2, "before cleanup"))
	cleanup()

	require.NoError(t, log.Output(2, "after cleanup"))
	assert.Contains(t, restored.String(), "after cleanup")
}

func TestWave36ConfigCreateInvalidPresetStopsBeforeBootstrap(t *testing.T) {
	cmd := configCmd()
	cmd.SetArgs([]string{"create", "wave36", "--preset", "unknown"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown preset "unknown"`)
}

func TestWave36RegisterCockpitPagesUsesDeadLetterToolBridge(t *testing.T) {
	t.Parallel()

	catalog := toolcatalog.New()
	var listCalled bool
	var detailCalled bool
	catalog.Register("post-adjudication", []*agent.Tool{
		{
			Name: "list_dead_lettered_post_adjudication_executions",
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				listCalled = true
				return map[string]interface{}{
					"entries": []postadjudicationstatus.DeadLetterBacklogEntry{
						{
							TransactionReceiptID:            "tx-wave36",
							SubmissionReceiptID:             "sub-wave36",
							Adjudication:                    "release",
							IsDeadLettered:                  true,
							CanRetry:                        true,
							LatestDeadLetterReason:          "worker exhausted",
							LatestStatusSubtype:             "dead-lettered",
							LatestStatusSubtypeFamily:       "dead-letter",
							LatestDispatchReference:         "dispatch-wave36",
							LatestManualReplayActor:         "operator:wave36",
							LatestRetryAttempt:              2,
							LatestDeadLetteredAt:            "2026-05-19T01:02:03Z",
							TransactionGlobalDominantFamily: "dead-letter",
						},
					},
				}, nil
			},
		},
		{
			Name: "get_post_adjudication_execution_status",
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				detailCalled = true
				assert.Equal(t, "tx-wave36", params["transaction_receipt_id"])
				return postadjudicationstatus.TransactionStatus{
					Adjudication:   "release",
					IsDeadLettered: true,
					CanRetry:       true,
					RetryDeadLetterSummary: postadjudicationstatus.RetryDeadLetterSummary{
						HasDeadLetter:          true,
						LatestDeadLetterReason: "worker exhausted",
					},
				}, nil
			},
		},
		{
			Name: "retry_post_adjudication_execution",
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				return map[string]interface{}{"accepted": true}, nil
			},
		},
	})

	model := cockpit.New(cockpit.Deps{})
	chatModel := model.ChatModel()
	require.NotNil(t, chatModel)
	registerCockpitPages(
		model,
		&app.App{Store: stubCockpitSessionStore{}, ToolCatalog: catalog},
		config.DefaultConfig(),
		"wave36",
		nil,
		cockpit.Deps{},
		chatModel,
	)

	page := model.Pages()[cockpit.PageDeadLetters]
	require.NotNil(t, page)
	loadCmd := page.Activate()
	require.NotNil(t, loadCmd)

	updated, detailCmd := page.Update(loadCmd())
	require.NotNil(t, detailCmd)
	page = updated.(cockpit.Page)
	updated, followup := page.Update(detailCmd())
	require.Nil(t, followup)
	page = updated.(cockpit.Page)

	assert.True(t, listCalled)
	assert.True(t, detailCalled)
	assert.Contains(t, page.View(), "tx-wave36")
	assert.Contains(t, page.View(), "worker exhausted")
}
