package app

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/agentpool"
	"github.com/langoai/lango/internal/p2p/team"
)


func bridgeTestLog() *zap.SugaredLogger { return zap.NewNop().Sugar() }

// setupBridgeTestEnv creates an event bus, coordinator, escrow engine, and budget engine
// for integration testing. The agent pool contains a leader and 2 workers with "search"
// capability.
func setupBridgeTestEnv(t *testing.T) (
	*eventbus.Bus,
	*team.Coordinator,
	*escrow.Engine,
	*budget.Engine,
) {
	t.Helper()

	bus := eventbus.New()
	log := bridgeTestLog()

	// Agent pool with leader + 2 workers.
	pool := agentpool.New(log)
	require.NoError(t, pool.Add(&agentpool.Agent{
		DID: "did:leader", Name: "leader", PeerID: "peer-leader",
		Capabilities: []string{"coordinate"}, Status: agentpool.StatusHealthy, TrustScore: 0.95,
	}))
	require.NoError(t, pool.Add(&agentpool.Agent{
		DID: "did:worker1", Name: "worker-1", PeerID: "peer-w1",
		Capabilities: []string{"search"}, Status: agentpool.StatusHealthy, TrustScore: 0.8,
	}))
	require.NoError(t, pool.Add(&agentpool.Agent{
		DID: "did:worker2", Name: "worker-2", PeerID: "peer-w2",
		Capabilities: []string{"search"}, Status: agentpool.StatusHealthy, TrustScore: 0.7,
	}))

	invokeFn := func(_ context.Context, peerID, toolName string, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"tool": toolName, "from": peerID}, nil
	}

	sel := agentpool.NewSelector(pool, agentpool.DefaultWeights())
	coord := team.NewCoordinator(team.CoordinatorConfig{
		Pool:     pool,
		Selector: sel,
		InvokeFn: invokeFn,
		Bus:      bus,
		Logger:   log,
	})

	// Escrow engine with in-memory store and noop settler.
	escrowStore := escrow.NewMemoryStore()
	escrowCfg := escrow.DefaultEngineConfig()
	escrowCfg.AutoRelease = false
	escrowEngine := escrow.NewEngine(escrowStore, escrow.NoopSettler{}, escrowCfg)

	// Budget engine with in-memory store.
	budgetStore := budget.NewStore()
	hardLimit := true
	budgetCfg := config.BudgetConfig{
		DefaultMax:      "100.0",
		HardLimit:       &hardLimit,
		AlertThresholds: []float64{0.5, 0.8},
	}
	budgetEngine, err := budget.NewEngine(budgetStore, budgetCfg)
	require.NoError(t, err)

	return bus, coord, escrowEngine, budgetEngine
}

// formTeamWithBudget forms a team via the coordinator, sets its budget, and
// re-publishes TeamFormedEvent so that bridges react to the non-zero budget.
// The initial FormTeam publish sees Budget=0 and is skipped by bridges.
func formTeamWithBudget(
	t *testing.T,
	ctx context.Context,
	coord *team.Coordinator,
	bus *eventbus.Bus,
	teamID, name, goal string,
	budgetAmount float64,
	memberCount int,
) *team.Team {
	t.Helper()

	tm, err := coord.FormTeam(ctx, team.FormTeamRequest{
		TeamID:      teamID,
		Name:        name,
		Goal:        goal,
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: memberCount,
	})
	require.NoError(t, err)
	require.NotNil(t, tm)

	tm.Budget = budgetAmount

	// Re-publish so bridges see the non-zero budget.
	bus.Publish(eventbus.TeamFormedEvent{
		TeamID:    teamID,
		Name:      name,
		Goal:      goal,
		LeaderDID: "did:leader",
		Members:   tm.MemberCount(),
	})

	return tm
}

func TestBridge_TeamFormed_CreatesEscrowAndBudget(t *testing.T) {
	bus, coord, escrowEngine, budgetEngine := setupBridgeTestEnv(t)
	log := bridgeTestLog()

	wireTeamEscrowBridge(bus, escrowEngine, coord, log)
	wireTeamBudgetBridge(context.Background(), bus, budgetEngine, coord, log)

	formTeamWithBudget(t, context.Background(), coord, bus,
		"team-1", "test-team", "test goal", 10.0, 2)

	// Verify escrow was created.
	escrows := escrowEngine.List()
	require.Len(t, escrows, 1, "expected 1 escrow")
	assert.Equal(t, "team-1", escrows[0].TaskID)
	assert.Equal(t, "did:leader", escrows[0].BuyerDID)
	assert.Equal(t, escrow.StatusPending, escrows[0].Status)
	assert.Len(t, escrows[0].Milestones, 2, "one milestone per worker")

	// Verify budget was allocated.
	err := budgetEngine.Check("team-1", big.NewInt(1))
	assert.NoError(t, err, "budget should be allocated for team-1")
}

func TestBridge_TeamTaskCompleted_CompletesMilestoneAndRecordsBudget(t *testing.T) {
	bus, coord, escrowEngine, budgetEngine := setupBridgeTestEnv(t)
	log := bridgeTestLog()

	wireTeamEscrowBridge(bus, escrowEngine, coord, log)
	wireTeamBudgetBridge(context.Background(), bus, budgetEngine, coord, log)

	formTeamWithBudget(t, context.Background(), coord, bus,
		"team-2", "task-team", "complete tasks", 5.0, 2)

	// Fund and activate the escrow so milestones can be completed.
	escrows := escrowEngine.List()
	require.Len(t, escrows, 1)
	escrowID := escrows[0].ID

	_, err := escrowEngine.Fund(context.Background(), escrowID)
	require.NoError(t, err)
	_, err = escrowEngine.Activate(context.Background(), escrowID)
	require.NoError(t, err)

	// Publish task completed event.
	bus.Publish(eventbus.TeamTaskCompletedEvent{
		TeamID:     "team-2",
		ToolName:   "web_search",
		Successful: 2,
		Failed:     0,
		Duration:   100 * time.Millisecond,
	})

	// Verify milestone was completed (synchronous bus, so no sleep needed).
	entry, err := escrowEngine.Get(escrowID)
	require.NoError(t, err)
	assert.Equal(t, 1, entry.CompletedMilestones(), "should have 1 completed milestone")

	// Verify budget spend was recorded.
	// 2 successful invocations * 0.1 USDC (100_000 micro) = 200_000.
	err = budgetEngine.Check("team-2", big.NewInt(1))
	assert.NoError(t, err, "budget should still be available after recording spend")
}

func TestBridge_TeamDisbanded_RefundsIncompleteEscrow(t *testing.T) {
	bus, coord, escrowEngine, _ := setupBridgeTestEnv(t)
	log := bridgeTestLog()

	wireTeamEscrowBridge(bus, escrowEngine, coord, log)

	formTeamWithBudget(t, context.Background(), coord, bus,
		"team-3", "disband-team", "test disband", 2.0, 1)

	escrows := escrowEngine.List()
	require.Len(t, escrows, 1)
	escrowID := escrows[0].ID

	// Fund and activate.
	_, err := escrowEngine.Fund(context.Background(), escrowID)
	require.NoError(t, err)
	_, err = escrowEngine.Activate(context.Background(), escrowID)
	require.NoError(t, err)

	// Disband team (milestones not completed -> should dispute then refund).
	err = coord.DisbandTeam("team-3")
	require.NoError(t, err)

	// Escrow should be refunded.
	entry, err := escrowEngine.Get(escrowID)
	require.NoError(t, err)
	assert.Equal(t, escrow.StatusRefunded, entry.Status,
		"escrow should be refunded on incomplete disband")
}

func TestBridge_FullLifecycle_ReleasesOnAllMilestonesCompleted(t *testing.T) {
	bus, coord, escrowEngine, budgetEngine := setupBridgeTestEnv(t)
	log := bridgeTestLog()

	wireTeamEscrowBridge(bus, escrowEngine, coord, log)
	wireTeamBudgetBridge(context.Background(), bus, budgetEngine, coord, log)

	// 1. Form team with budget.
	formTeamWithBudget(t, context.Background(), coord, bus,
		"team-lifecycle", "lifecycle-team", "full lifecycle test", 1.0, 1)

	// 2. Fund and activate escrow.
	escrows := escrowEngine.List()
	require.Len(t, escrows, 1)
	escrowID := escrows[0].ID

	_, err := escrowEngine.Fund(context.Background(), escrowID)
	require.NoError(t, err)
	_, err = escrowEngine.Activate(context.Background(), escrowID)
	require.NoError(t, err)

	// 3. Complete all milestones via task completion events.
	entry, err := escrowEngine.Get(escrowID)
	require.NoError(t, err)
	for range entry.Milestones {
		bus.Publish(eventbus.TeamTaskCompletedEvent{
			TeamID:     "team-lifecycle",
			ToolName:   "web_search",
			Successful: 1,
			Failed:     0,
			Duration:   50 * time.Millisecond,
		})
	}

	// Verify all milestones completed.
	entry, err = escrowEngine.Get(escrowID)
	require.NoError(t, err)
	assert.True(t, entry.AllMilestonesCompleted(), "all milestones should be completed")

	// 4. Disband -> should release (all milestones done).
	err = coord.DisbandTeam("team-lifecycle")
	require.NoError(t, err)

	entry, err = escrowEngine.Get(escrowID)
	require.NoError(t, err)
	assert.Equal(t, escrow.StatusReleased, entry.Status,
		"escrow should be released after full lifecycle")
}

func TestBridge_NoBudget_SkipsBridges(t *testing.T) {
	bus, coord, escrowEngine, budgetEngine := setupBridgeTestEnv(t)
	log := bridgeTestLog()

	wireTeamEscrowBridge(bus, escrowEngine, coord, log)
	wireTeamBudgetBridge(context.Background(), bus, budgetEngine, coord, log)

	// Form team WITHOUT setting a budget (Budget stays at 0).
	_, err := coord.FormTeam(context.Background(), team.FormTeamRequest{
		TeamID:      "team-nobudget",
		Name:        "no-budget-team",
		Goal:        "test no budget",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 1,
	})
	require.NoError(t, err)

	// No escrow should be created.
	assert.Empty(t, escrowEngine.List(), "no escrow for zero-budget team")

	// No budget should be allocated.
	err = budgetEngine.Check("team-nobudget", big.NewInt(1))
	assert.Error(t, err, "budget check should fail for unallocated team")
}
