package app

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildFoundationCatalogEntries_AssignsEnabledToolsByPrefix(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Tools.Browser.Enabled = true
	baseTools := []*agent.Tool{
		{Name: "exec_run"},
		{Name: "fs_read"},
		{Name: "web_fetch"},
		{Name: "browser_open"},
		{Name: "custom_tool"},
	}
	cryptoTools := []*agent.Tool{{Name: "crypto_sign"}}
	secretTools := []*agent.Tool{{Name: "secrets_get"}}

	entries := buildFoundationCatalogEntries(cfg, baseTools, cryptoTools, secretTools)

	execEntry := requireCatalogEntry(t, entries, "exec")
	assert.True(t, execEntry.Enabled)
	assert.Equal(t, []string{"exec_run"}, catalogEntryToolNames(execEntry))

	fsEntry := requireCatalogEntry(t, entries, "filesystem")
	assert.True(t, fsEntry.Enabled)
	assert.Equal(t, []string{"fs_read"}, catalogEntryToolNames(fsEntry))

	browserEntry := requireCatalogEntry(t, entries, "browser")
	assert.True(t, browserEntry.Enabled)
	assert.Equal(t, []string{"browser_open"}, catalogEntryToolNames(browserEntry))

	webEntry := requireCatalogEntry(t, entries, "web")
	assert.True(t, webEntry.Enabled)
	assert.Equal(t, []string{"web_fetch"}, catalogEntryToolNames(webEntry))

	assert.True(t, requireCatalogEntry(t, entries, "crypto").Enabled)
	assert.Equal(t, []string{"crypto_sign"}, catalogEntryToolNames(requireCatalogEntry(t, entries, "crypto")))
	assert.True(t, requireCatalogEntry(t, entries, "secrets").Enabled)
	assert.Equal(t, []string{"secrets_get"}, catalogEntryToolNames(requireCatalogEntry(t, entries, "secrets")))
}

func TestBuildFoundationCatalogEntries_DisabledOptionalCategoriesOmitTools(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Tools.Browser.Enabled = false
	entries := buildFoundationCatalogEntries(cfg, []*agent.Tool{{Name: "browser_open"}}, nil, nil)

	browserEntry := requireCatalogEntry(t, entries, "browser")
	assert.False(t, browserEntry.Enabled)
	assert.Empty(t, browserEntry.Tools)
	assert.Contains(t, browserEntry.Description, "disabled")

	cryptoEntry := requireCatalogEntry(t, entries, "crypto")
	assert.False(t, cryptoEntry.Enabled)
	assert.Empty(t, cryptoEntry.Tools)

	secretsEntry := requireCatalogEntry(t, entries, "secrets")
	assert.False(t, secretsEntry.Enabled)
	assert.Empty(t, secretsEntry.Tools)
}

func TestNetworkModuleInit_P2PWithoutPaymentRegistersDisabledNetworkEntries(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = false
	cfg.P2P.Enabled = true
	cfg.P2P.Workspace.Enabled = true
	mod := &networkModule{cfg: cfg, bus: eventbus.New()}

	result, err := mod.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Tools)
	assert.Empty(t, result.Components)
	assert.Nil(t, result.Values[appinit.ProvidesPayment])
	assert.Nil(t, result.Values[appinit.ProvidesP2P])

	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "payment").Enabled)
	p2pEntry := requireCatalogEntry(t, result.CatalogEntries, "p2p")
	assert.False(t, p2pEntry.Enabled)
	assert.Contains(t, p2pEntry.Description, "payment required")
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "workspace").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "contract").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "smartaccount").Enabled)
}

func TestExtensionModuleInit_DisabledSubsystemsReturnDisabledCatalogEntries(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.MCP.Enabled = false
	cfg.Observability.Enabled = false
	mod := &extensionModule{cfg: cfg, bus: eventbus.New()}

	result, err := mod.Init(context.Background(), staticResolver{})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Tools)
	assert.Empty(t, result.Components)
	assert.Nil(t, result.Values[appinit.ProvidesMCP])
	assert.Nil(t, result.Values[appinit.ProvidesObservability])

	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "mcp").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "observability").Enabled)
}

func TestAutomationModuleInit_BackgroundOnlyRegistersControlAndDisabledEntries(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Background.Enabled = true
	cfg.Cron.Enabled = false
	cfg.Workflow.Enabled = false
	app := &App{Config: cfg}
	mod := &automationModule{cfg: cfg, app: app, bus: eventbus.New()}

	result, err := mod.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{Store: &stubSessionStore{}},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotNil(t, result.Values[appinit.ProvidesAutomation])
	assert.NotEmpty(t, result.Tools)
	assert.NotEmpty(t, result.Components)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "background").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "agent_control").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "task_tracking").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "cron").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "workflow").Enabled)

	values, ok := result.Values[appinit.ProvidesAutomation].(*automationValues)
	require.True(t, ok)
	assert.NotNil(t, values.BackgroundManager)
	assert.Nil(t, values.CronScheduler)
	assert.Nil(t, values.WorkflowEngine)
	assert.NotNil(t, values.AgentRunStore)
}

func requireCatalogEntry(
	t *testing.T,
	entries []appinit.CatalogEntry,
	category string,
) appinit.CatalogEntry {
	t.Helper()

	for _, entry := range entries {
		if entry.Category == category {
			return entry
		}
	}
	require.Failf(t, "missing catalog entry", "category %q not found", category)
	return appinit.CatalogEntry{}
}

func catalogEntryToolNames(entry appinit.CatalogEntry) []string {
	names := make([]string, 0, len(entry.Tools))
	for _, tool := range entry.Tools {
		names = append(names, tool.Name)
	}
	return names
}
