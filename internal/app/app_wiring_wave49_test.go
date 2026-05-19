package app

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/toolcatalog"
)

func TestWave49P2PWiringSafeBranchesAvoidRuntimeStartup(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = false
	cfg.P2P.ZKHandshake = false
	cfg.P2P.ZKAttestation = false
	cfg.P2P.KeyDir = filepath.Join(t.TempDir(), "disabled-keys")
	cfg.P2P.ListenAddrs = []string{"/ip4/127.0.0.1/tcp/0"}

	assert.Nil(t, initP2P(cfg, &wiringP2PWallet{}, nil, nil, nil, nil, nil, nil, ""))
	assert.NoDirExists(t, cfg.P2P.KeyDir)
	assert.Nil(t, initZKP(cfg))

	cfg.P2P.Enabled = true
	cfg.P2P.KeyDir = filepath.Join(t.TempDir(), "missing-wallet-keys")
	assert.Nil(t, initP2P(cfg, nil, nil, nil, nil, nil, nil, nil, ""))
	assert.NoDirExists(t, cfg.P2P.KeyDir)
}

func TestWave49WalletSignerPropagatesPublicKeyErrors(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("public key unavailable")
	signer := &walletHandshakeSigner{wp: &wiringP2PWallet{pubErr: wantErr}}

	gotDID, err := signer.DID(context.Background())
	require.ErrorIs(t, err, wantErr)
	assert.Empty(t, gotDID)
	assert.Equal(t, security.AlgorithmSecp256k1Keccak256, signer.Algorithm())
}

func TestWave49CatalogSourceAdapterFiltersModesAndDisabledCategories(t *testing.T) {
	t.Parallel()

	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{
		Name:        "alpha",
		Description: "Alpha tools",
		Enabled:     true,
	})
	catalog.RegisterCategory(toolcatalog.Category{
		Name:        "beta",
		Description: "Beta tools",
		ConfigKey:   "beta.enabled",
		Enabled:     false,
	})
	catalog.Register("alpha", []*agent.Tool{
		{Name: "alpha_visible", Description: "visible alpha"},
		{
			Name:        "alpha_deferred",
			Description: "deferred alpha",
			Capability:  agent.ToolCapability{Exposure: agent.ExposureDeferred},
		},
	})
	catalog.Register("beta", []*agent.Tool{{Name: "beta_visible", Description: "visible beta"}})

	cfg := config.DefaultConfig()
	cfg.Modes = map[string]config.SessionMode{
		"wave49": {
			Name:  "wave49",
			Tools: []string{"alpha_visible"},
		},
	}
	adapter := &catalogSourceAdapter{catalog: catalog, cfg: cfg}

	modeSection := adapter.BuildToolCatalogSection("wave49")
	assert.Contains(t, modeSection, "## Tools Available in `wave49` Mode")
	assert.Contains(t, modeSection, "alpha_visible")
	assert.NotContains(t, modeSection, "alpha_deferred")
	assert.NotContains(t, modeSection, "beta_visible")
	assert.NotContains(t, modeSection, "Disabled categories")

	defaultSection := adapter.BuildToolCatalogSection("")
	assert.Contains(t, defaultSection, "## Available Tool Categories")
	assert.Contains(t, defaultSection, "alpha_visible")
	assert.Contains(t, defaultSection, "beta_visible")
	assert.Contains(t, defaultSection, "Disabled categories (enable via config): beta (beta.enabled)")
	assert.Contains(t, defaultSection, "builtin_list")
}

func TestWave49ModeResolverAndAutomationPromptSafeFallbacks(t *testing.T) {
	t.Parallel()

	assert.Empty(t, (*modeResolverAdapter)(nil).LookupModeHint("wave49"))
	assert.Empty(t, (&modeResolverAdapter{}).LookupModeHint("wave49"))
	assert.Empty(t, (&modeResolverAdapter{cfg: config.DefaultConfig()}).LookupModeHint(""))

	cfg := config.DefaultConfig()
	cfg.Modes = map[string]config.SessionMode{
		"wave49": {
			Name:       "wave49",
			SystemHint: "Stay inside deterministic wiring tests.",
		},
	}
	assert.Equal(t, "Stay inside deterministic wiring tests.",
		(&modeResolverAdapter{cfg: cfg}).LookupModeHint("wave49"))
	assert.Empty(t, (&modeResolverAdapter{cfg: cfg}).LookupModeHint("missing"))

	cfg.Cron.Enabled = false
	cfg.Background.Enabled = true
	cfg.Workflow.Enabled = true
	rendered := buildAutomationPromptSection(cfg).Render()
	assert.Contains(t, rendered, "## Automation")
	assert.NotContains(t, rendered, "### Cron Scheduling")
	assert.Contains(t, rendered, "### Background Tasks")
	assert.Contains(t, rendered, "### Workflow Pipelines")
	assert.Contains(t, rendered, "NEVER use exec")
}

func TestWave49NetworkModuleDisabledCatalogAndValuesWithoutPeers(t *testing.T) {
	t.Parallel()

	cfg := wave49ModuleConfig(t)
	cfg.P2P.Enabled = false
	cfg.P2P.Workspace.Enabled = true
	cfg.Economy.Enabled = false
	module := &networkModule{cfg: cfg, bus: eventbus.New(), app: &App{ctx: context.Background()}}

	result, err := module.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{ReceiptStore: receipts.NewStore()},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Tools)
	assert.Empty(t, result.Components)
	assert.Nil(t, result.Values[appinit.ProvidesPayment])
	assert.Nil(t, result.Values[appinit.ProvidesP2P])
	assert.Nil(t, result.Values[appinit.ProvidesEconomy])
	assert.Nil(t, result.Values[appinit.ProvidesWorkspace])

	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "payment").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "contract").Enabled)
	p2pEntry := requireCatalogEntry(t, result.CatalogEntries, "p2p")
	assert.False(t, p2pEntry.Enabled)
	assert.NotContains(t, p2pEntry.Description, "payment required")
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "workspace").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "economy").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "smartaccount").Enabled)
}

func TestWave49NetworkModuleEnabledFlagFollowsFeatureConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*config.Config)
		want      bool
	}{
		{
			name: "all network features disabled",
			want: false,
		},
		{
			name: "payment enabled",
			configure: func(cfg *config.Config) {
				cfg.Payment.Enabled = true
			},
			want: true,
		},
		{
			name: "p2p enabled",
			configure: func(cfg *config.Config) {
				cfg.P2P.Enabled = true
			},
			want: true,
		},
		{
			name: "economy enabled",
			configure: func(cfg *config.Config) {
				cfg.Economy.Enabled = true
			},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := wave49ModuleConfig(t)
			if tt.configure != nil {
				tt.configure(cfg)
			}

			assert.Equal(t, tt.want, (&networkModule{cfg: cfg}).Enabled())
		})
	}
}

func TestWave49MissionAwareSubmitterCancelsSubmittedTaskWhenLinkingFails(t *testing.T) {
	t.Parallel()

	linkErr := errors.New("mission store unavailable")
	submitter := &wave49Submitter{taskID: "task-wave49"}
	wrapped := &missionAwareSubmitter{
		base:   submitter,
		linker: &wave49Linker{err: linkErr},
	}

	taskID, err := wrapped.Submit(context.Background(), "summarize", background.Origin{
		Channel: "telegram:wave49",
		Session: "session-wave49",
	})
	require.Error(t, err)
	assert.Empty(t, taskID)
	assert.ErrorContains(t, err, "attach spawned child execution to mission")
	assert.ErrorIs(t, err, linkErr)
	assert.Equal(t, []string{"task-wave49"}, submitter.canceled)
}

func wave49ModuleConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = ""
	cfg.Providers = nil
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "session.db")
	cfg.Security.Signer.Provider = ""
	cfg.Payment.Enabled = false
	cfg.P2P.Enabled = false
	cfg.P2P.Workspace.Enabled = false
	cfg.Economy.Enabled = false
	cfg.SmartAccount.Enabled = false
	cfg.Tools.Browser.Enabled = false
	cfg.MCP.Enabled = false
	return cfg
}

type wave49Submitter struct {
	taskID    string
	submitErr error
	cancelErr error
	canceled  []string
}

func (s *wave49Submitter) Submit(_ context.Context, _ string, _ background.Origin) (string, error) {
	if s.submitErr != nil {
		return "", s.submitErr
	}
	return s.taskID, nil
}

func (s *wave49Submitter) Cancel(taskID string) error {
	s.canceled = append(s.canceled, taskID)
	return s.cancelErr
}

type wave49Linker struct {
	err error
}

func (l *wave49Linker) LinkBackgroundTask(context.Context, string, background.Origin, string) error {
	return l.err
}
