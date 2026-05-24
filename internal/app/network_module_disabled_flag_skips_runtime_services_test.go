package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/receipts"
)

func TestNetworkModuleDisabledFlagSkipsRuntimeServices(t *testing.T) {
	t.Parallel()

	cfg := networkModuleDisabledFlagSkipsRuntimeServicesModuleConfig(t)
	cfg.P2P.Workspace.Enabled = true
	cfg.SmartAccount.Enabled = true
	module := &networkModule{cfg: cfg, bus: eventbus.New()}

	require.False(t, module.Enabled())
}

func TestNetworkModuleInitPaymentRequiredCatalogWithoutExternalServices(t *testing.T) {
	t.Parallel()

	cfg := networkModuleDisabledFlagSkipsRuntimeServicesModuleConfig(t)
	cfg.Payment.Enabled = true
	cfg.P2P.Enabled = true
	cfg.P2P.Workspace.Enabled = true
	cfg.Economy.Enabled = true
	module := &networkModule{cfg: cfg, bus: eventbus.New()}

	require.True(t, module.Enabled())

	result, err := module.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: networkModuleDisabledFlagSkipsRuntimeServicesFoundationValues(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Tools)
	assert.Empty(t, result.Components)
	assert.Nil(t, result.Values[appinit.ProvidesPayment])
	assert.Nil(t, result.Values[appinit.ProvidesP2P])
	assert.NotNil(t, result.Values[appinit.ProvidesEconomy])
	assert.Nil(t, result.Values[appinit.ProvidesContract])
	assert.Nil(t, result.Values[appinit.ProvidesSmartAccount])
	assert.Nil(t, result.Values[appinit.ProvidesWorkspace])

	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "payment").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "contract").Enabled)
	p2pEntry := requireCatalogEntry(t, result.CatalogEntries, "p2p")
	assert.False(t, p2pEntry.Enabled)
	assert.Contains(t, p2pEntry.Description, "payment required")
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "workspace").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "smartaccount").Enabled)
	economyEntry := requireCatalogEntry(t, result.CatalogEntries, "economy")
	assert.True(t, economyEntry.Enabled)
	assert.NotEmpty(t, economyEntry.Tools)
}

func TestIntelligenceModuleInitDisabledCatalogWithoutLifecycleComponents(t *testing.T) {
	t.Parallel()

	cfg := networkModuleDisabledFlagSkipsRuntimeServicesModuleConfig(t)
	module := &intelligenceModule{cfg: cfg, bus: eventbus.New()}

	require.True(t, module.Enabled())

	result, err := module.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: networkModuleDisabledFlagSkipsRuntimeServicesFoundationValues(),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Tools)
	assert.Empty(t, result.Components)
	assert.Nil(t, result.Values[appinit.ProvidesGraph])
	assert.Nil(t, result.Values[appinit.ProvidesMemory])
	assert.Nil(t, result.Values[appinit.ProvidesLibrarian])
	assert.Nil(t, result.Values[appinit.ProvidesSkills])

	values, ok := result.Values[appinit.ProvidesKnowledge].(*intelligenceValues)
	require.True(t, ok)
	require.NotNil(t, values)
	assert.Nil(t, values.KC)
	assert.Nil(t, values.MC)
	assert.Nil(t, values.GC)
	assert.Nil(t, values.LC)
	assert.Nil(t, values.AB)
	assert.Nil(t, values.Observer)
	assert.Nil(t, values.SkillRegistry)
	assert.Nil(t, values.AgentMemoryStore)
	assert.Nil(t, values.OntologyBridge)
	require.NotNil(t, values.FeatureStatuses)
	assert.Zero(t, values.FeatureStatuses.SilentDisabledCount())

	for _, category := range []string{
		"meta",
		"graph",
		"memory",
		"agent_memory",
		"librarian",
	} {
		entry := requireCatalogEntry(t, result.CatalogEntries, category)
		assert.False(t, entry.Enabled, "expected %s catalog entry to be disabled", category)
		assert.Empty(t, entry.Tools, "expected %s catalog entry to be lightweight", category)
	}
	assert.Nil(t, networkModuleDisabledFlagSkipsRuntimeServicesCatalogEntry(result.CatalogEntries, "ontology"))
}

func networkModuleDisabledFlagSkipsRuntimeServicesModuleConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = ""
	cfg.Providers = nil
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "session.db")
	cfg.Skill.Enabled = false
	cfg.Skill.SkillsDir = filepath.Join(t.TempDir(), "skills")
	cfg.Tools.Browser.Enabled = false
	cfg.Payment.Enabled = false
	cfg.P2P.Enabled = false
	cfg.P2P.Workspace.Enabled = false
	cfg.Economy.Enabled = false
	cfg.SmartAccount.Enabled = false
	cfg.Knowledge.Enabled = false
	cfg.Graph.Enabled = false
	cfg.ObservationalMemory.Enabled = false
	cfg.AgentMemory.Enabled = false
	cfg.Librarian.Enabled = false
	cfg.Ontology.Enabled = false
	return cfg
}

func networkModuleDisabledFlagSkipsRuntimeServicesFoundationValues() *foundationValues {
	return &foundationValues{
		Store:        &stubSessionStore{},
		ReceiptStore: receipts.NewStore(),
	}
}

func networkModuleDisabledFlagSkipsRuntimeServicesCatalogEntry(
	entries []appinit.CatalogEntry,
	category string,
) *appinit.CatalogEntry {
	for _, entry := range entries {
		if entry.Category == category {
			return &entry
		}
	}
	return nil
}
