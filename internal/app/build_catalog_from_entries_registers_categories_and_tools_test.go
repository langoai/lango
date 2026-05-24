package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	p2pproto "github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/toolcatalog"
	"github.com/langoai/lango/internal/types"
)

func TestBuildCatalogFromEntriesRegistersCategoriesAndTools(t *testing.T) {
	t.Parallel()

	safeTool := &agent.Tool{
		Name:        "safe_lookup",
		Description: "Safe lookup",
		SafetyLevel: agent.SafetyLevelSafe,
	}
	dangerousTool := &agent.Tool{
		Name:        "dangerous_write",
		Description: "Dangerous write",
		SafetyLevel: agent.SafetyLevelDangerous,
	}
	catalog := buildCatalogFromEntries([]appinit.CatalogEntry{
		{
			Category:    "safe",
			Description: "Safe helpers",
			Enabled:     true,
			Tools:       []*agent.Tool{safeTool},
		},
		{
			Category:    "dangerous",
			Description: "Dangerous helpers",
			ConfigKey:   "tools.dangerous.enabled",
			Enabled:     false,
			Tools:       []*agent.Tool{dangerousTool},
		},
		{
			Category:    "empty",
			Description: "Empty category",
			Enabled:     true,
		},
	})

	assert.Equal(t, 2, catalog.ToolCount())
	assert.Equal(t, []string{"safe_lookup"}, catalog.ToolNamesForCategory("safe"))
	assert.Equal(t, []string{"dangerous_write"}, catalog.ToolNamesForCategory("dangerous"))
	assert.Empty(t, catalog.ToolNamesForCategory("empty"))

	level, found := catalog.GetToolSafetyLevel("dangerous_write")
	require.True(t, found)
	assert.Equal(t, agent.SafetyLevelDangerous, level)

	section := (&catalogSourceAdapter{catalog: catalog}).BuildToolCatalogSection("")
	assert.Contains(t, section, "safe_lookup")
	assert.Contains(t, section, "dangerous_write")
	assert.Contains(t, section, "Disabled categories (enable via config): dangerous (tools.dangerous.enabled)")
}

func TestPopulateAppFieldsPublishesFeatureStatusesToGatewayHealth(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	application := &App{
		Config:  cfg,
		Gateway: initGateway(cfg, nil, &stubSessionStore{}, nil),
	}
	statuses := NewStatusCollector()
	statuses.Add(&types.FeatureStatus{
		Name:    "Knowledge",
		Enabled: true,
		Healthy: true,
		Reason:  "buildCatalogFromEntriesRegistersCategoriesAndTools1 deterministic status",
	})

	populateAppFields(application, staticResolver{
		appinit.ProvidesKnowledge: &intelligenceValues{
			FeatureStatuses: statuses,
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	application.Gateway.Router().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	features, ok := body["features"].([]any)
	require.True(t, ok)
	require.Len(t, features, 1)
	feature, ok := features[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Knowledge", feature["name"])
	assert.Equal(t, true, feature["enabled"])
	assert.Equal(t, true, feature["healthy"])
	assert.Equal(t, "buildCatalogFromEntriesRegistersCategoriesAndTools1 deterministic status", feature["reason"])
	assert.Same(t, statuses, application.FeatureStatuses)
}

func TestWirePostAgentWiresP2PHandlerExecutorAndSafetyGate(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.ToolIsolation.Enabled = false
	cfg.P2P.MaxSafetyLevel = "safe"

	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{
		Name:        "buildCatalogFromEntriesRegistersCategoriesAndTools1",
		Description: "Catalog fixture tools",
		Enabled:     true,
	})
	catalog.Register("buildCatalogFromEntriesRegistersCategoriesAndTools1", []*agent.Tool{
		{Name: "buildCatalogFromEntriesRegistersCategoriesAndTools_safe", Description: "safe", SafetyLevel: agent.SafetyLevelSafe},
		{Name: "buildCatalogFromEntriesRegistersCategoriesAndTools_dangerous", Description: "dangerous", SafetyLevel: agent.SafetyLevelDangerous},
	})

	handler := p2pproto.NewHandler(p2pproto.HandlerConfig{})
	application := &App{
		Config:      cfg,
		Gateway:     initGateway(cfg, nil, &stubSessionStore{}, nil),
		ToolCatalog: catalog,
	}
	tools := []*agent.Tool{
		{
			Name:        "buildCatalogFromEntriesRegistersCategoriesAndTools_safe",
			Description: "safe",
			SafetyLevel: agent.SafetyLevelSafe,
			Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
				return "ok", nil
			},
		},
	}

	wirePostAgent(
		application,
		staticResolver{appinit.ProvidesP2P: &p2pComponents{handler: handler}},
		tools,
		eventbus.New(),
		approval.NewCompositeProvider(),
		approval.NewGrantStore(),
		nil,
		nil,
	)

	value := reflect.ValueOf(handler).Elem()
	assert.False(t, value.FieldByName("executor").IsNil())
	assert.True(t, value.FieldByName("sandboxExec").IsNil())
	assert.True(t, value.FieldByName("approvalFn").IsNil())
	assert.False(t, value.FieldByName("safetyChecker").IsNil())
	assert.Equal(t, int64(agent.SafetyLevelSafe), value.FieldByName("maxSafetyLevel").Int())
}

func TestModeResolverAndCatalogSourceHandleNilAndMissingInputs(t *testing.T) {
	t.Parallel()

	assert.Empty(t, (&catalogSourceAdapter{}).BuildToolCatalogSection(""))
	assert.Empty(t, (&catalogSourceAdapter{catalog: toolcatalog.New()}).BuildToolCatalogSection(""))
	assert.Empty(t, (&modeResolverAdapter{}).LookupModeHint("missing"))

	cfg := config.DefaultConfig()
	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{
		Name:        "general",
		Description: "General tools",
		Enabled:     true,
	})
	catalog.Register("general", []*agent.Tool{{
		Name:        "general_tool",
		Description: "general",
		SafetyLevel: agent.SafetyLevelSafe,
	}})
	section := (&catalogSourceAdapter{catalog: catalog, cfg: cfg}).BuildToolCatalogSection("missing-mode")

	assert.Contains(t, section, "## Tools Available in `missing-mode` Mode")
	assert.Contains(t, section, "general_tool")
	assert.Contains(t, section, "Only tools in this mode's allowlist")
}

func TestNetworkModuleInitRegistersDisabledEntriesWhenPaymentUnavailable(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = false
	cfg.P2P.Enabled = true
	cfg.P2P.Workspace.Enabled = true
	cfg.Economy.Enabled = true
	cfg.SmartAccount.Enabled = true

	result, err := (&networkModule{cfg: cfg, bus: eventbus.New()}).Init(
		context.Background(),
		staticResolver{appinit.ProvidesSupervisor: &foundationValues{}},
	)
	require.NoError(t, err)

	assert.NotEmpty(t, result.Tools)
	assert.Empty(t, result.Components)
	assert.Nil(t, result.Values[appinit.ProvidesPayment])
	assert.Nil(t, result.Values[appinit.ProvidesP2P])
	assert.NotNil(t, result.Values[appinit.ProvidesEconomy])
	assert.Nil(t, result.Values[appinit.ProvidesSmartAccount])
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "payment").Enabled)
	assert.Contains(t, requireCatalogEntry(t, result.CatalogEntries, "p2p").Description, "payment required")
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "economy").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "workspace").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "smartaccount").Enabled)
}

func TestInitP2PSkipsDisabledAndMissingWalletBeforeNetworkSetup(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	//nolint:staticcheck // intentional: regression test pins legacy KeyDir behavior on disabled/missing-wallet paths.
	cfg.P2P.KeyDir = filepath.Join(t.TempDir(), "p2p-keys")
	cfg.P2P.ListenAddrs = []string{"/ip4/127.0.0.1/tcp/0"}

	cfg.P2P.Enabled = false
	assert.Nil(t, initP2P(cfg, nil, nil, nil, nil, nil, nil, nil, ""))
	//nolint:staticcheck // intentional: see above.
	assert.NoDirExists(t, cfg.P2P.KeyDir)

	cfg.P2P.Enabled = true
	assert.Nil(t, initP2P(cfg, nil, nil, nil, nil, nil, nil, nil, ""))
	//nolint:staticcheck // intentional: see above.
	assert.NoDirExists(t, cfg.P2P.KeyDir)
}
