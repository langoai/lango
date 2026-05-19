package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/settlementexecution"
)

func TestWave45BuildMetaToolsWithEscrowEngineAddsDerivedEscrowTools(t *testing.T) {
	t.Parallel()

	engine := escrow.NewEngine(
		escrow.NewMemoryStore(),
		escrow.NoopSettler{},
		escrow.DefaultEngineConfig(),
	)

	tools := buildMetaToolsWithEscrow(
		nil,
		nil,
		nil,
		config.SkillConfig{},
		config.DefaultConfig(),
		receipts.NewStore(),
		engine,
	)

	for _, name := range []string{
		"execute_escrow_recommendation",
		"hold_escrow_for_dispute",
		"release_escrow_settlement",
		"refund_escrow_settlement",
	} {
		require.NotNil(t, findTool(tools, name), "expected %s to be registered", name)
	}
}

func TestWave45BuildMetaToolsWithSettlementRuntimeDerivesPartialRuntime(t *testing.T) {
	t.Parallel()

	runtime := &wave45SettlementRuntime{}
	tools := buildMetaToolsWithRuntimes(
		nil,
		nil,
		nil,
		config.SkillConfig{},
		nil,
		receipts.NewStore(),
		nil,
		runtime,
		nil,
		nil,
		nil,
		nil,
	)

	require.NotNil(t, findTool(tools, "execute_settlement"))
	require.NotNil(t, findTool(tools, "execute_partial_settlement"))
}

func TestWave45PathInsideDirAcceptsRootAndRejectsSiblingEscapes(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	assert.True(t, pathInsideDir(root, root))
	assert.True(t, pathInsideDir(filepath.Join(root, "skill", "SKILL.md"), root))
	assert.False(t, pathInsideDir(filepath.Join(root, "..", "outside.md"), root))
	assert.False(t, pathInsideDir(root+"-sibling", root))
}

func TestWave45NetworkModuleInitPaymentDisabledKeepsWorkspaceAndP2PDisabled(t *testing.T) {
	t.Parallel()

	cfg := wave45ModuleConfig(t)
	cfg.P2P.Enabled = true
	cfg.P2P.Workspace.Enabled = true
	cfg.SmartAccount.Enabled = true
	module := &networkModule{cfg: cfg, bus: eventbus.New(), app: &App{ctx: context.Background()}}

	result, err := module.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{ReceiptStore: receipts.NewStore()},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Empty(t, result.Components)
	assert.Nil(t, result.Values[appinit.ProvidesPayment])
	assert.Nil(t, result.Values[appinit.ProvidesP2P])
	assert.Nil(t, result.Values[appinit.ProvidesWorkspace])
	assert.Contains(t, requireCatalogEntry(t, result.CatalogEntries, "p2p").Description, "payment required")
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "workspace").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "smartaccount").Enabled)
}

func TestWave45InitSecurityUnsupportedProviderIsReported(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Security.Signer.Provider = "wave45-kms"

	provider, keys, secrets, err := initSecurity(cfg, &stubSessionStore{}, nil)

	require.Error(t, err)
	assert.ErrorContains(t, err, `unsupported security provider "wave45-kms"`)
	assert.Nil(t, provider)
	assert.Nil(t, keys)
	assert.Nil(t, secrets)
}

type wave45SettlementRuntime struct{}

func (r *wave45SettlementRuntime) ExecuteSettlement(
	context.Context,
	settlementexecution.DirectPaymentRequest,
) (settlementexecution.DirectPaymentResult, error) {
	return settlementexecution.DirectPaymentResult{Reference: "wave45-settlement"}, nil
}

func wave45ModuleConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = ""
	cfg.Providers = nil
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "session.db")
	cfg.Skill.Enabled = false
	cfg.Tools.Browser.Enabled = false
	cfg.Payment.Enabled = false
	cfg.P2P.Enabled = false
	cfg.P2P.Workspace.Enabled = false
	cfg.Economy.Enabled = false
	cfg.SmartAccount.Enabled = false
	cfg.Cron.Enabled = false
	cfg.Background.Enabled = false
	cfg.Workflow.Enabled = false
	cfg.Knowledge.Enabled = false
	cfg.Graph.Enabled = false
	cfg.ObservationalMemory.Enabled = false
	cfg.AgentMemory.Enabled = false
	cfg.Librarian.Enabled = false
	cfg.Ontology.Enabled = false
	cfg.MCP.Enabled = false
	cfg.Observability.Enabled = false
	return cfg
}
