package main

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/cockpit/pages"
	clistatus "github.com/langoai/lango/internal/cli/status"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/types"
)

func TestNewDeadLetterStatusLoaderWrapsBuildAppError(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	cfg.Session.DatabasePath = cfg.DataRoot
	loader := newDeadLetterStatusLoader(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, ProfileName: "newDeadLetterStatusLoaderWrapsBuildAppError0"}, nil
	})

	bridge, cleanup, err := loader()

	require.Error(t, err)
	assert.Nil(t, bridge)
	assert.Nil(t, cleanup)
	assert.Contains(t, err.Error(), "build app:")
}

func TestNewDeadLetterStatusLoaderReportsUnavailableToolsAfterAppBuild(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.DataRoot = t.TempDir()
	cfg.Session.DatabasePath = filepath.Join(cfg.DataRoot, "lango.db")
	cfg.Agent.Provider = "deadletter-status-test"
	cfg.Providers = map[string]config.ProviderConfig{
		"deadletter-status-test": {Type: types.ProviderOpenAI, APIKey: "test-key"},
	}
	loader := newDeadLetterStatusLoader(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: cfg, ProfileName: "deadletter-status-test"}, nil
	})

	bridge, cleanup, err := loader()

	require.Error(t, err)
	assert.ErrorIs(t, err, clistatus.ErrDeadLetterStatusToolsUnavailable)
	assert.Nil(t, bridge)
	assert.Nil(t, cleanup)
}

func TestCockpitDeadLetterListOptionsCopiesAllFilters(t *testing.T) {
	t.Parallel()

	got := cockpitDeadLetterListOptions(pages.DeadLetterListOptions{
		Query:                     "agent=worker-c",
		Adjudication:              "dead_lettered",
		LatestStatusSubtype:       "retry_failed",
		LatestStatusSubtypeFamily: "retry",
		AnyMatchFamily:            "dispatch",
		ManualReplayActor:         "operator",
		DeadLetteredAfter:         "2026-05-19T10:00:00Z",
		DeadLetteredBefore:        "2026-05-19T12:00:00Z",
		DeadLetterReasonQuery:     "timeout",
		LatestDispatchReference:   "dispatch-123",
	})

	assert.Equal(t, "agent=worker-c", got.Query)
	assert.Equal(t, "dead_lettered", got.Adjudication)
	assert.Equal(t, "retry_failed", got.LatestStatusSubtype)
	assert.Equal(t, "retry", got.LatestStatusSubtypeFamily)
	assert.Equal(t, "dispatch", got.AnyMatchFamily)
	assert.Equal(t, "operator", got.ManualReplayActor)
	assert.Equal(t, "2026-05-19T10:00:00Z", got.DeadLetteredAfter)
	assert.Equal(t, "2026-05-19T12:00:00Z", got.DeadLetteredBefore)
	assert.Equal(t, "timeout", got.DeadLetterReasonQuery)
	assert.Equal(t, "dispatch-123", got.LatestDispatchReference)
}

func TestNewDeadLetterStatusLoaderKeepsBootstrapErrorPath(t *testing.T) {
	t.Parallel()

	bootErr := errors.New("newDeadLetterStatusLoaderWrapsBuildAppError0 boot refused")
	loader := newDeadLetterStatusLoader(func() (*bootstrap.Result, error) {
		return nil, bootErr
	})

	bridge, cleanup, err := loader()

	require.ErrorIs(t, err, bootErr)
	assert.Nil(t, bridge)
	assert.Nil(t, cleanup)
	assert.Contains(t, err.Error(), "bootstrap: newDeadLetterStatusLoaderWrapsBuildAppError0 boot refused")
}
