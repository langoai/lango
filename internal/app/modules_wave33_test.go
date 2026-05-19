package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/testutil"
)

func TestWave33AutomationModuleInitAllEnabledComposesToolsWithoutStartingComponents(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Cron.Enabled = true
	cfg.Background.Enabled = true
	cfg.Workflow.Enabled = true
	cfg.Workflow.StateDir = t.TempDir()
	application := &App{Config: cfg}
	module := &automationModule{cfg: cfg, app: application, bus: eventbus.New()}
	store := session.NewEntStoreWithClient(testutil.TestEntClient(t))

	result, err := module.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{
			Store:        store,
			ReceiptStore: receipts.NewStore(),
		},
		appinit.ProvidesRunLedger: &runLedgerValues{},
		appinit.ProvidesMission:   &missionValues{},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Components, 3)

	values, ok := result.Values[appinit.ProvidesAutomation].(*automationValues)
	require.True(t, ok)
	require.NotNil(t, values)
	assert.NotNil(t, values.CronScheduler)
	assert.NotNil(t, values.BackgroundManager)
	assert.NotNil(t, values.WorkflowEngine)
	assert.NotNil(t, values.AgentRunStore)

	for _, category := range []string{
		"cron",
		"background",
		"workflow",
		"agent_control",
		"task_tracking",
	} {
		entry := requireCatalogEntry(t, result.CatalogEntries, category)
		assert.True(t, entry.Enabled, "expected %s to be enabled", category)
		assert.NotEmpty(t, entry.Tools, "expected %s tools", category)
	}

	assert.NotNil(t, findTool(result.Tools, "cron_add"))
	assert.NotNil(t, findTool(result.Tools, "bg_submit"))
	assert.NotNil(t, findTool(result.Tools, "workflow_run"))
	assert.NotNil(t, findTool(result.Tools, "agent_spawn"))
	assert.NotNil(t, findTool(result.Tools, "task_create"))
}
