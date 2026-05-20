package app

import (
	"context"
	"errors"
	"math/big"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/economy/negotiation"
	"github.com/langoai/lango/internal/eventbus"
	p2pproto "github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/toolcatalog"
)

func TestNetworkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled(t *testing.T) {
	t.Parallel()

	cfg := metaLearningToolsPersistSearchStatsAndCleanupDryRunModuleConfig(t)
	cfg.Payment.Enabled = false
	cfg.P2P.Enabled = true
	cfg.P2P.Workspace.Enabled = true
	cfg.Economy.Enabled = true
	cfg.Economy.Pricing.Enabled = true
	cfg.Economy.Negotiate.Enabled = true
	cfg.SmartAccount.Enabled = true

	result, err := (&networkModule{cfg: cfg, bus: eventbus.New()}).Init(
		context.Background(),
		staticResolver{appinit.ProvidesSupervisor: &foundationValues{}},
	)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Empty(t, result.Components)
	assert.Nil(t, result.Values[appinit.ProvidesPayment])
	assert.Nil(t, result.Values[appinit.ProvidesP2P])
	assert.NotNil(t, result.Values[appinit.ProvidesEconomy])
	assert.Nil(t, result.Values[appinit.ProvidesWorkspace])
	assert.Nil(t, result.Values[appinit.ProvidesSmartAccount])
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "payment").Enabled)
	assert.Contains(t, requireCatalogEntry(t, result.CatalogEntries, "p2p").Description, "payment required")
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "workspace").Enabled)
	assert.True(t, requireCatalogEntry(t, result.CatalogEntries, "economy").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "smartaccount").Enabled)
	assert.NotEmpty(t, result.Tools, "economy tools should still be registered without payment")
}

func TestIntelligenceModuleInitDisabledSubsystemsPublishesDisabledCatalog(t *testing.T) {
	t.Parallel()

	cfg := metaLearningToolsPersistSearchStatsAndCleanupDryRunModuleConfig(t)
	module := &intelligenceModule{cfg: cfg, bus: eventbus.New()}

	result, err := module.Init(context.Background(), staticResolver{
		appinit.ProvidesSupervisor: &foundationValues{
			Store: &stubSessionStore{},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	values, ok := result.Values[appinit.ProvidesKnowledge].(*intelligenceValues)
	require.True(t, ok)
	require.NotNil(t, values.FeatureStatuses)
	assert.Nil(t, values.KC)
	assert.Nil(t, values.MC)
	assert.Nil(t, values.GC)
	assert.Nil(t, values.LC)
	assert.Nil(t, values.AgentMemoryStore)
	assert.Empty(t, result.Components)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "meta").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "graph").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "memory").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "agent_memory").Enabled)
	assert.False(t, requireCatalogEntry(t, result.CatalogEntries, "librarian").Enabled)
}

func TestInitEconomyPublishesBudgetAndNegotiationEvents(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Budget.AlertThresholds = []float64{0.5}
	cfg.Economy.Pricing.Enabled = true
	cfg.Economy.Negotiate.Enabled = true

	bus := eventbus.New()
	var budgetAlerts []eventbus.BudgetAlertEvent
	eventOrder := make(chan string, 6)
	started := make(chan eventbus.NegotiationStartedEvent, 3)
	completed := make(chan eventbus.NegotiationCompletedEvent, 1)
	failed := make(chan eventbus.NegotiationFailedEvent, 2)
	eventbus.SubscribeTyped(bus, func(evt eventbus.BudgetAlertEvent) {
		budgetAlerts = append(budgetAlerts, evt)
	})
	eventbus.SubscribeTyped(bus, func(evt eventbus.NegotiationStartedEvent) {
		eventOrder <- "started:" + evt.SessionID
		started <- evt
	})
	eventbus.SubscribeTyped(bus, func(evt eventbus.NegotiationCompletedEvent) {
		eventOrder <- "completed:" + evt.SessionID
		completed <- evt
	})
	eventbus.SubscribeTyped(bus, func(evt eventbus.NegotiationFailedEvent) {
		eventOrder <- "failed:" + evt.SessionID + ":" + evt.Reason
		failed <- evt
	})

	components := initEconomy(cfg, nil, nil, bus)
	require.NotNil(t, components)
	require.NotNil(t, components.budgetEngine)
	require.NotNil(t, components.pricingEngine)
	require.NotNil(t, components.negotiationEngine)

	_, err := components.budgetEngine.Allocate("networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3-task", big.NewInt(100))
	require.NoError(t, err)
	require.NoError(t, components.budgetEngine.Record("networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3-task", budget.SpendEntry{
		Amount:  big.NewInt(50),
		PeerDID: "did:lango:networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3-budget-peer",
	}))
	require.Len(t, budgetAlerts, 1)
	assert.Equal(t, "networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3-task", budgetAlerts[0].TaskID)
	assert.Equal(t, 0.5, budgetAlerts[0].Threshold)

	components.pricingEngine.SetBasePrice("networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_tool", big.NewInt(100))
	session, err := components.negotiationEngine.Propose(
		context.Background(),
		"did:lango:networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3-initiator",
		"did:lango:networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3-responder",
		negotiationTerms("networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_tool", 100),
	)
	require.NoError(t, err)
	accepted, err := components.negotiationEngine.AutoRespond(context.Background(), session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, accepted.ID)

	rejected, err := components.negotiationEngine.Propose(
		context.Background(),
		"did:lango:networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3-reject-init",
		"did:lango:networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3-reject-responder",
		negotiationTerms("networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_tool", 1),
	)
	require.NoError(t, err)
	_, err = components.negotiationEngine.Reject(
		context.Background(),
		rejected.ID,
		"did:lango:networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3-reject-responder",
		"not worth it",
	)
	require.NoError(t, err)

	cancelled, err := components.negotiationEngine.Propose(
		context.Background(),
		"did:lango:networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3-cancel-init",
		"did:lango:networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3-cancel-responder",
		negotiationTerms("networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_tool", 1),
	)
	require.NoError(t, err)
	_, err = components.negotiationEngine.Cancel(
		context.Background(),
		cancelled.ID,
		"did:lango:networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3-cancel-init",
	)
	require.NoError(t, err)

	startedEvents := drainNetworkModuleInitEconomyWithoutPaymentKeepsNetworkDisabledStartedEvents(t, started, 3)
	assert.Len(t, startedEvents, 3)
	assert.Equal(t, session.ID, startedEvents[0].SessionID)
	assert.Equal(t, "networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_tool", startedEvents[0].ToolName)
	assert.Equal(t, rejected.ID, startedEvents[1].SessionID)
	assert.Equal(t, cancelled.ID, startedEvents[2].SessionID)

	completedEvent := receiveNetworkModuleInitEconomyWithoutPaymentKeepsNetworkDisabledEvent(t, completed)
	assert.Equal(t, session.ID, completedEvent.SessionID)
	assert.Equal(t, big.NewInt(100), completedEvent.AgreedPrice)

	rejectedFailure := receiveNetworkModuleInitEconomyWithoutPaymentKeepsNetworkDisabledEvent(t, failed)
	cancelledFailure := receiveNetworkModuleInitEconomyWithoutPaymentKeepsNetworkDisabledEvent(t, failed)
	failedReasons := []string{rejectedFailure.Reason, cancelledFailure.Reason}
	assert.ElementsMatch(t, []string{"rejected", "cancelled"}, failedReasons)
	assert.Equal(t, []string{
		"started:" + session.ID,
		"completed:" + session.ID,
		"started:" + rejected.ID,
		"failed:" + rejected.ID + ":rejected",
		"started:" + cancelled.ID,
		"failed:" + cancelled.ID + ":cancelled",
	}, drainNetworkModuleInitEconomyWithoutPaymentKeepsNetworkDisabledEventOrder(t, eventOrder, 6))
}

func TestWirePostAgentExecutesMappedToolsAndReportsErrors(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.ToolIsolation.Enabled = false
	cfg.P2P.MaxSafetyLevel = "dangerous"

	handler := p2pproto.NewHandler(p2pproto.HandlerConfig{})

	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{Name: "networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3", Description: "Network disabled fixture", Enabled: true})
	catalog.Register("networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3", []*agent.Tool{
		{Name: "networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_scalar", Description: "scalar", SafetyLevel: agent.SafetyLevelSafe},
		{Name: "networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_map", Description: "map", SafetyLevel: agent.SafetyLevelSafe},
		{Name: "networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_error", Description: "error", SafetyLevel: agent.SafetyLevelSafe},
	})
	application := &App{
		Config:      cfg,
		Gateway:     initGateway(cfg, nil, &stubSessionStore{}, nil),
		ToolCatalog: catalog,
	}

	wirePostAgent(
		application,
		staticResolver{appinit.ProvidesP2P: &p2pComponents{handler: handler}},
		[]*agent.Tool{
			{
				Name:        "networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_scalar",
				Description: "scalar",
				SafetyLevel: agent.SafetyLevelSafe,
				Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
					return "scalar-ok", nil
				},
			},
			{
				Name:        "networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_map",
				Description: "map",
				SafetyLevel: agent.SafetyLevelSafe,
				Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
					return map[string]interface{}{"mapped": true}, nil
				},
			},
			{
				Name:        "networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_error",
				Description: "error",
				SafetyLevel: agent.SafetyLevelSafe,
				Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
					return nil, errors.New("networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3 handler failed")
				},
			},
		},
		eventbus.New(),
		approval.NewCompositeProvider(),
		approval.NewGrantStore(),
		nil,
		nil,
	)

	executor := networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabledHandlerExecutor(t, handler)
	scalar, err := executor(context.Background(), "networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_scalar", nil)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"result": "scalar-ok"}, scalar)

	mapped, err := executor(context.Background(), "networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_map", nil)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"mapped": true}, mapped)

	missing, err := executor(context.Background(), "networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_missing", nil)
	assert.Nil(t, missing)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `tool "networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_missing" not found`)

	failed, err := executor(context.Background(), "networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled_error", nil)
	assert.Nil(t, failed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3 handler failed")
}

func negotiationTerms(toolName string, price int64) negotiation.Terms {
	return negotiation.Terms{
		ToolName: toolName,
		Price:    big.NewInt(price),
	}
}

func drainNetworkModuleInitEconomyWithoutPaymentKeepsNetworkDisabledStartedEvents(
	t *testing.T,
	events <-chan eventbus.NegotiationStartedEvent,
	count int,
) []eventbus.NegotiationStartedEvent {
	t.Helper()

	started := make([]eventbus.NegotiationStartedEvent, 0, count)
	for i := 0; i < count; i++ {
		started = append(started, receiveNetworkModuleInitEconomyWithoutPaymentKeepsNetworkDisabledEvent(t, events))
	}
	return started
}

func receiveNetworkModuleInitEconomyWithoutPaymentKeepsNetworkDisabledEvent[T any](t *testing.T, events <-chan T) T {
	t.Helper()

	select {
	case evt := <-events:
		return evt
	case <-time.After(time.Second):
		require.Fail(t, "timed out waiting for networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabled3 event")
		var zero T
		return zero
	}
}

func drainNetworkModuleInitEconomyWithoutPaymentKeepsNetworkDisabledEventOrder(t *testing.T, events <-chan string, count int) []string {
	t.Helper()

	order := make([]string, 0, count)
	for i := 0; i < count; i++ {
		order = append(order, receiveNetworkModuleInitEconomyWithoutPaymentKeepsNetworkDisabledEvent(t, events))
	}
	return order
}

func networkModuleInitEconomyWithoutPaymentKeepsNetworkDisabledHandlerExecutor(t *testing.T, handler *p2pproto.Handler) p2pproto.ToolExecutor {
	t.Helper()

	field := reflect.ValueOf(handler).Elem().FieldByName("executor")
	require.True(t, field.IsValid())
	require.False(t, field.IsNil())
	value := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
	executor, ok := value.(p2pproto.ToolExecutor)
	require.True(t, ok)
	return executor
}
