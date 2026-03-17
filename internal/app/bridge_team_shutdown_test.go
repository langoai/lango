package app

import (
	"context"
	"math/big"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/agentpool"
	"github.com/langoai/lango/internal/p2p/team"
)

func testLog() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

func setupTestCoordinator(t *testing.T, bus *eventbus.Bus) *team.Coordinator {
	t.Helper()

	pool := agentpool.New(testLog())
	_ = pool.Add(&agentpool.Agent{
		DID: "did:leader", Name: "leader", PeerID: "peer-leader",
		Capabilities: []string{"coordinate"}, Status: agentpool.StatusHealthy,
	})
	_ = pool.Add(&agentpool.Agent{
		DID: "did:worker1", Name: "worker-1", PeerID: "peer-w1",
		Capabilities: []string{"search"}, Status: agentpool.StatusHealthy,
	})

	invokeFn := func(_ context.Context, peerID, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"ok": true}, nil
	}

	sel := agentpool.NewSelector(pool, agentpool.DefaultWeights())
	return team.NewCoordinator(team.CoordinatorConfig{
		Pool: pool, Selector: sel, InvokeFn: invokeFn, Bus: bus, Logger: testLog(),
	})
}

func TestTeamShutdownBridge_BudgetExhausted(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	coord := setupTestCoordinator(t, bus)

	_, err := coord.FormTeam(context.Background(), team.FormTeamRequest{
		TeamID: "t-exhaust", Name: "exhaust-team", Goal: "test",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 1,
	})
	require.NoError(t, err)

	var mu sync.Mutex
	var shutdownEvents []eventbus.TeamGracefulShutdownEvent
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamGracefulShutdownEvent) {
		mu.Lock()
		defer mu.Unlock()
		shutdownEvents = append(shutdownEvents, ev)
	})

	initTeamShutdownBridge(bus, coord, testLog())

	// Simulate budget exhausted.
	bus.Publish(eventbus.BudgetExhaustedEvent{
		TaskID:     "t-exhaust",
		TotalSpent: big.NewInt(1_000_000),
	})

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, shutdownEvents, 1)
	assert.Equal(t, "t-exhaust", shutdownEvents[0].TeamID)
	assert.Equal(t, "budget exhausted", shutdownEvents[0].Reason)

	// Team should be disbanded.
	_, err = coord.GetTeam("t-exhaust")
	assert.ErrorIs(t, err, team.ErrTeamNotFound)
}

func TestTeamShutdownBridge_BudgetWarning(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	coord := setupTestCoordinator(t, bus)

	tm, err := coord.FormTeam(context.Background(), team.FormTeamRequest{
		TeamID: "t-warn", Name: "warn-team", Goal: "test",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 1,
	})
	require.NoError(t, err)
	// Set budget on the team for warning event.
	tm.Budget = 100.0

	var mu sync.Mutex
	var warnings []eventbus.TeamBudgetWarningEvent
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamBudgetWarningEvent) {
		mu.Lock()
		defer mu.Unlock()
		warnings = append(warnings, ev)
	})

	initTeamShutdownBridge(bus, coord, testLog())

	// 80% threshold should trigger warning.
	bus.Publish(eventbus.BudgetAlertEvent{
		TaskID:    "t-warn",
		Threshold: 0.8,
	})

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, warnings, 1)
	assert.Equal(t, "t-warn", warnings[0].TeamID)
	assert.Equal(t, 0.8, warnings[0].Threshold)
	assert.Equal(t, 100.0, warnings[0].Budget)
}

func TestTeamShutdownBridge_LowThresholdIgnored(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	coord := setupTestCoordinator(t, bus)

	_, err := coord.FormTeam(context.Background(), team.FormTeamRequest{
		TeamID: "t-low", Name: "low-team", Goal: "test",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 1,
	})
	require.NoError(t, err)

	var mu sync.Mutex
	var warnings []eventbus.TeamBudgetWarningEvent
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamBudgetWarningEvent) {
		mu.Lock()
		defer mu.Unlock()
		warnings = append(warnings, ev)
	})

	initTeamShutdownBridge(bus, coord, testLog())

	// 50% threshold should NOT trigger warning.
	bus.Publish(eventbus.BudgetAlertEvent{
		TaskID:    "t-low",
		Threshold: 0.5,
	})

	mu.Lock()
	defer mu.Unlock()
	assert.Empty(t, warnings, "50% threshold should not trigger warning")
}

func TestDelegateTask_RejectsShuttingDown(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	coord := setupTestCoordinator(t, bus)

	tm, err := coord.FormTeam(context.Background(), team.FormTeamRequest{
		TeamID: "t-shutting", Name: "shutting-team", Goal: "test",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 1,
	})
	require.NoError(t, err)

	// Manually set shutting down.
	tm.Status = team.StatusShuttingDown

	_, err = coord.DelegateTask(context.Background(), "t-shutting", "search", nil)
	assert.ErrorIs(t, err, team.ErrTeamShuttingDown)
}
