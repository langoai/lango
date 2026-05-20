package app

import (
	"context"
	"database/sql"
	"math/big"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/eventbus"
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
