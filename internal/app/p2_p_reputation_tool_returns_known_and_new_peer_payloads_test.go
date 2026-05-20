package app

import (
	"context"
	"database/sql"
	"math/big"
	"reflect"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/economy/negotiation"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/paygate"
	p2pproto "github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/p2p/reputation"
	"github.com/langoai/lango/internal/testutil"
)

func TestP2PReputationToolReturnsKnownAndNewPeerPayloads(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repStore := reputation.NewStore(testutil.TestEntClient(t), testLog())
	knownPeer := "did:lango:initP2PEarlyExitBranchesAvoidNetworkState2-known"
	require.NoError(t, repStore.RecordSuccess(ctx, knownPeer))
	require.NoError(t, repStore.RecordTimeout(ctx, knownPeer))

	tool := findP2PTool(t, buildP2PTools(&p2pComponents{reputation: repStore}), "p2p_reputation")

	got, err := tool.Handler(ctx, map[string]interface{}{"peer_did": knownPeer})
	require.NoError(t, err)
	known := p2PToolsMetadataAndMissingDependencyBranchesP2PPayload(t, got)
	assert.Equal(t, knownPeer, known["peerDID"])
	assert.Equal(t, 1, known["successfulExchanges"])
	assert.Equal(t, 0, known["failedExchanges"])
	assert.Equal(t, 1, known["timeoutCount"])
	assert.NotEmpty(t, known["firstSeen"])
	assert.NotEmpty(t, known["lastInteraction"])

	got, err = tool.Handler(ctx, map[string]interface{}{"peer_did": "did:lango:initP2PEarlyExitBranchesAvoidNetworkState2-new"})
	require.NoError(t, err)
	newPeer := p2PToolsMetadataAndMissingDependencyBranchesP2PPayload(t, got)
	assert.Equal(t, "did:lango:initP2PEarlyExitBranchesAvoidNetworkState2-new", newPeer["peerDID"])
	assert.Equal(t, 0.0, newPeer["score"])
	assert.Equal(t, true, newPeer["isTrusted"])
	assert.Equal(t, "new peer — no reputation record", newPeer["message"])
}

func TestEconomyWiringBudgetAlertAndInvalidBudgetConfig(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	var alerts []eventbus.BudgetAlertEvent
	eventbus.SubscribeTyped(bus, func(event eventbus.BudgetAlertEvent) {
		alerts = append(alerts, event)
	})

	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Budget.DefaultMax = "1.00"
	cfg.Economy.Budget.AlertThresholds = []float64{0.5}

	components := initEconomy(cfg, nil, nil, bus)
	require.NotNil(t, components)
	require.NotNil(t, components.budgetEngine)
	_, err := components.budgetEngine.Allocate("initP2PEarlyExitBranchesAvoidNetworkState2-task", nil)
	require.NoError(t, err)
	require.NoError(t, components.budgetEngine.Record("initP2PEarlyExitBranchesAvoidNetworkState2-task", budgetSpend(500_000)))

	require.Len(t, alerts, 1)
	assert.Equal(t, "initP2PEarlyExitBranchesAvoidNetworkState2-task", alerts[0].TaskID)
	assert.Equal(t, 0.5, alerts[0].Threshold)

	invalid := config.DefaultConfig()
	invalid.Economy.Enabled = true
	invalid.Economy.Budget.DefaultMax = "not-usdc"

	invalidComponents := initEconomy(invalid, nil, nil, nil)
	require.NotNil(t, invalidComponents)
	assert.Nil(t, invalidComponents.budgetEngine)
	assert.NotNil(t, invalidComponents.riskEngine)
}

func TestSelectSettlerUnknownOnChainModeFallsBackToCustodian(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Economy.Escrow.OnChain.Enabled = true
	cfg.Economy.Escrow.OnChain.Mode = "unknown"

	settler := selectSettler(cfg, &paymentComponents{chainID: 8453}, nil)

	_, ok := settler.(*escrow.USDCSettler)
	assert.True(t, ok)
}

func TestInitEconomyPricingResidualBranches(t *testing.T) {
	t.Parallel()

	invalid := config.DefaultConfig()
	invalid.Economy.Enabled = true
	invalid.Economy.Pricing.Enabled = true
	invalid.Economy.Pricing.MinPrice = "not-usdc"

	invalidComponents := initEconomy(invalid, nil, nil, nil)
	require.NotNil(t, invalidComponents)
	assert.Nil(t, invalidComponents.pricingEngine)
	assert.NotNil(t, invalidComponents.budgetEngine)

	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Pricing.Enabled = true
	p2pc := &p2pComponents{payGate: paygate.New(paygate.Config{PricingFn: func(string) (string, bool) {
		return "stale", true
	}})}

	components := initEconomy(cfg, p2pc, nil, nil)
	require.NotNil(t, components)
	require.NotNil(t, components.pricingEngine)
	require.NotNil(t, p2pc.pricingFn)

	components.pricingEngine.SetBasePrice("economy_priced_tool", big.NewInt(123456))
	price, free := p2pc.pricingFn("economy_priced_tool")
	assert.False(t, free)
	assert.Equal(t, "0.123456", price)

	payGateResult, err := p2pc.payGate.Check("did:lango:economy-pricing-peer", "economy_priced_tool", nil)
	require.NoError(t, err)
	assert.Equal(t, paygate.StatusPaymentRequired, payGateResult.Status)
	require.NotNil(t, payGateResult.PriceQuote)
	assert.Equal(t, "0.123456", payGateResult.PriceQuote.Price)

	price, free = p2pc.pricingFn("economy_free_tool")
	assert.True(t, free)
	assert.Empty(t, price)
}

func TestInitEconomyPublishesExpiredNegotiationEvent(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Negotiate.Enabled = true

	bus := eventbus.New()
	failed := make(chan eventbus.NegotiationFailedEvent, 1)
	eventbus.SubscribeTyped(bus, func(evt eventbus.NegotiationFailedEvent) {
		failed <- evt
	})

	components := initEconomy(cfg, nil, nil, bus)
	require.NotNil(t, components)
	require.NotNil(t, components.negotiationEngine)

	session, err := components.negotiationEngine.Propose(
		context.Background(),
		"did:lango:initEconomyExpired-initiator",
		"did:lango:initEconomyExpired-responder",
		negotiation.Terms{ToolName: "expired_tool", Price: big.NewInt(1)},
	)
	require.NoError(t, err)
	session.ExpiresAt = time.Now().Add(-time.Second)

	assert.Equal(t, []string{session.ID}, components.negotiationEngine.CheckExpiry())

	select {
	case evt := <-failed:
		assert.Equal(t, session.ID, evt.SessionID)
		assert.Equal(t, "expired", evt.Reason)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for expired negotiation event")
	}
}

func TestInitEconomyEscrowCustomConfigAndP2PResolver(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Escrow.Enabled = true
	cfg.Economy.Escrow.DefaultTimeout = 2 * time.Hour
	cfg.Economy.Escrow.MaxMilestones = 1
	cfg.Economy.Escrow.DisputeWindow = 30 * time.Minute

	components := initEconomy(cfg, &p2pComponents{}, nil, eventbus.New())
	require.NotNil(t, components)
	require.NotNil(t, components.escrowEngine)
	assert.NotNil(t, components.sentinelEngine)
	_, ok := components.escrowSettler.(escrow.NoopSettler)
	assert.True(t, ok)

	_, err := components.escrowEngine.Create(context.Background(), escrow.CreateRequest{
		BuyerDID:  "did:lango:initEconomyEscrowCustomConfig-buyer",
		SellerDID: "did:lango:initEconomyEscrowCustomConfig-seller",
		TaskID:    "initEconomyEscrowCustomConfig-task",
		Reason:    "custom config coverage",
		Milestones: []escrow.MilestoneRequest{
			{Description: "one", Amount: big.NewInt(1)},
			{Description: "two", Amount: big.NewInt(1)},
		},
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "too many milestones")
}

func TestInitEconomyWiresP2PNegotiatorWithoutIdentity(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Negotiate.Enabled = true
	handler := p2pproto.NewHandler(p2pproto.HandlerConfig{})
	p2pc := &p2pComponents{handler: handler}

	components := initEconomy(cfg, p2pc, nil, nil)
	require.NotNil(t, components)
	require.NotNil(t, components.negotiationEngine)

	negotiator := reflect.ValueOf(handler).Elem().FieldByName("negotiator")
	require.True(t, negotiator.IsValid())
	assert.False(t, negotiator.IsNil())

	got, err := handleNegotiateProtocol(context.Background(), components.negotiationEngine, "", "did:lango:p2p-negotiator-peer", p2pproto.NegotiatePayload{
		Action:   string(negotiation.ActionPropose),
		ToolName: "p2p_negotiator_tool",
		Price:    "1",
	})
	require.NoError(t, err)
	assert.Equal(t, string(negotiation.PhaseProposed), got["phase"])
}

func TestKnowledgeFTS5NilAndMissingSourceTableBranches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	assert.False(t, initFTS5(ctx, nil, nil))

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	_, err = db.ExecContext(ctx, `CREATE TABLE knowledge_fts (source_id TEXT)`)
	require.NoError(t, err)
	err = bulkIndexKnowledge(ctx, db, nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, "query knowledge for FTS5 index")

	_, err = db.ExecContext(ctx, `CREATE TABLE learning_fts (source_id TEXT)`)
	require.NoError(t, err)
	err = bulkIndexLearnings(ctx, db, nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, "query learnings for FTS5 index")
}

func budgetSpend(amount int64) budget.SpendEntry {
	return budget.SpendEntry{Amount: big.NewInt(amount)}
}
