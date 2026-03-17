//go:build integration

package team

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/agentpool"
)

func integrationLogger() *zap.SugaredLogger { return zap.NewNop().Sugar() }

// setupIntegrationEnv creates a full coordinator with agent pool, event bus, and BoltDB store.
func setupIntegrationEnv(t *testing.T) (*Coordinator, *eventbus.Bus, *bolt.DB) {
	t.Helper()

	bus := eventbus.New()
	log := integrationLogger()

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

	dbPath := filepath.Join(t.TempDir(), "integration-test.db")
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
		os.Remove(dbPath)
	})

	store, err := NewBoltStore(db, log)
	require.NoError(t, err)

	invokeFn := func(_ context.Context, peerID, toolName string, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"tool": toolName, "from": peerID}, nil
	}

	sel := agentpool.NewSelector(pool, agentpool.DefaultWeights())
	coord := NewCoordinator(CoordinatorConfig{
		Pool:     pool,
		Selector: sel,
		InvokeFn: invokeFn,
		Bus:      bus,
		Store:    store,
		Logger:   log,
	})

	return coord, bus, db
}

// TestIntegration_FormDelegateDisband tests the full team lifecycle.
func TestIntegration_FormDelegateDisband(t *testing.T) {
	coord, bus, _ := setupIntegrationEnv(t)
	ctx := context.Background()

	// Track events.
	var formed, disbanded atomic.Int32
	eventbus.SubscribeTyped(bus, func(_ eventbus.TeamFormedEvent) { formed.Add(1) })
	eventbus.SubscribeTyped(bus, func(_ eventbus.TeamDisbandedEvent) { disbanded.Add(1) })

	// 1. Form team.
	tm, err := coord.FormTeam(ctx, FormTeamRequest{
		TeamID:      "integration-team-1",
		Name:        "test-team",
		Goal:        "integration test",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 2,
	})
	require.NoError(t, err)
	assert.Equal(t, StatusActive, tm.Status)
	assert.GreaterOrEqual(t, tm.MemberCount(), 2)

	// 2. Delegate task.
	results, err := coord.DelegateTask(ctx, "integration-team-1", "web_search", map[string]interface{}{"q": "test"})
	require.NoError(t, err)
	assert.NotEmpty(t, results)
	for _, r := range results {
		assert.NoError(t, r.Err)
		assert.NotNil(t, r.Result)
	}

	// 3. Collect results.
	resolved, err := coord.CollectResults("integration-team-1", "web_search", results)
	require.NoError(t, err)
	assert.NotNil(t, resolved)

	// 4. Disband team.
	err = coord.DisbandTeam("integration-team-1")
	require.NoError(t, err)

	assert.Equal(t, int32(1), formed.Load())
	assert.Equal(t, int32(1), disbanded.Load())

	// Team should no longer exist.
	_, err = coord.GetTeam("integration-team-1")
	assert.ErrorIs(t, err, ErrTeamNotFound)
}

// TestIntegration_BudgetExhaustion_GracefulShutdown tests shutdown when budget is exhausted.
func TestIntegration_BudgetExhaustion_GracefulShutdown(t *testing.T) {
	coord, bus, _ := setupIntegrationEnv(t)
	ctx := context.Background()

	var shutdownEvents []eventbus.TeamGracefulShutdownEvent
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamGracefulShutdownEvent) {
		shutdownEvents = append(shutdownEvents, ev)
	})

	// Form team with budget.
	tm, err := coord.FormTeam(ctx, FormTeamRequest{
		TeamID:      "budget-team",
		Name:        "budget-test",
		Goal:        "test budget shutdown",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 1,
	})
	require.NoError(t, err)
	tm.Budget = 1.0

	// Simulate some spending.
	require.NoError(t, tm.AddSpend(0.8))

	// Trigger graceful shutdown (simulating budget exhaustion).
	err = coord.GracefulShutdown(ctx, "budget-team", "budget exhausted")
	require.NoError(t, err)

	// Verify shutdown event was published.
	require.Len(t, shutdownEvents, 1)
	assert.Equal(t, "budget-team", shutdownEvents[0].TeamID)
	assert.Equal(t, "budget exhausted", shutdownEvents[0].Reason)

	// Team should be disbanded.
	_, err = coord.GetTeam("budget-team")
	assert.ErrorIs(t, err, ErrTeamNotFound)
}

// TestIntegration_MemberTimeout_HealthEvent tests health monitor detecting unhealthy members.
func TestIntegration_MemberTimeout_HealthEvent(t *testing.T) {
	coord, bus, _ := setupIntegrationEnv(t)
	ctx := context.Background()

	// Create invokeFn that fails for worker-2 (simulating timeout).
	failingInvoke := func(_ context.Context, peerID, toolName string, _ map[string]interface{}) (map[string]interface{}, error) {
		if peerID == "peer-w2" {
			return nil, errors.New("connection timeout")
		}
		return map[string]interface{}{"ok": true}, nil
	}

	// Form team.
	_, err := coord.FormTeam(ctx, FormTeamRequest{
		TeamID:      "health-team",
		Name:        "health-test",
		Goal:        "test health monitoring",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 2,
	})
	require.NoError(t, err)

	// Create health monitor with short interval and low threshold.
	var unhealthyEvents []eventbus.TeamMemberUnhealthyEvent
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamMemberUnhealthyEvent) {
		unhealthyEvents = append(unhealthyEvents, ev)
	})

	monitor := NewHealthMonitor(HealthMonitorConfig{
		Coordinator: coord,
		Bus:         bus,
		Logger:      integrationLogger(),
		Interval:    50 * time.Millisecond,
		MaxMissed:   2,
		InvokeFn:    failingInvoke,
	})

	require.NoError(t, monitor.Start(ctx, nil))
	defer monitor.Stop(ctx)

	// Wait for enough health checks to trigger unhealthy event.
	time.Sleep(200 * time.Millisecond)

	// Worker-2 should be unhealthy (missed >= 2 pings).
	assert.NotEmpty(t, unhealthyEvents, "expected unhealthy events for worker-2")
	found := false
	for _, ev := range unhealthyEvents {
		if ev.MemberDID == "did:worker2" {
			found = true
			assert.GreaterOrEqual(t, ev.MissedPings, 2)
		}
	}
	assert.True(t, found, "expected unhealthy event for did:worker2")
}

// TestIntegration_KickMember tests kicking a member from a team.
func TestIntegration_KickMember(t *testing.T) {
	coord, bus, _ := setupIntegrationEnv(t)
	ctx := context.Background()

	var leftEvents []eventbus.TeamMemberLeftEvent
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamMemberLeftEvent) {
		leftEvents = append(leftEvents, ev)
	})

	// Form team.
	tm, err := coord.FormTeam(ctx, FormTeamRequest{
		TeamID:      "kick-team",
		Name:        "kick-test",
		Goal:        "test member kick",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 2,
	})
	require.NoError(t, err)
	initialCount := tm.MemberCount()

	// Kick worker-2.
	err = coord.KickMember(ctx, "kick-team", "did:worker2", "low reputation")
	require.NoError(t, err)

	// Verify member was removed.
	tm, err = coord.GetTeam("kick-team")
	require.NoError(t, err)
	assert.Equal(t, initialCount-1, tm.MemberCount())
	assert.Nil(t, tm.GetMember("did:worker2"))

	// Verify left event was published.
	require.NotEmpty(t, leftEvents)
	var found bool
	for _, ev := range leftEvents {
		if ev.MemberDID == "did:worker2" && ev.Reason == "low reputation" {
			found = true
		}
	}
	assert.True(t, found, "expected TeamMemberLeftEvent for did:worker2 with reason 'low reputation'")
}

// TestIntegration_TeamsForMember tests finding teams for a given member DID.
func TestIntegration_TeamsForMember(t *testing.T) {
	coord, _, _ := setupIntegrationEnv(t)
	ctx := context.Background()

	// Form two teams with the same workers.
	_, err := coord.FormTeam(ctx, FormTeamRequest{
		TeamID: "multi-1", Name: "multi-1", Goal: "test 1",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 2,
	})
	require.NoError(t, err)

	_, err = coord.FormTeam(ctx, FormTeamRequest{
		TeamID: "multi-2", Name: "multi-2", Goal: "test 2",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 2,
	})
	require.NoError(t, err)

	// Worker should be in both teams.
	teams := coord.TeamsForMember("did:worker1")
	assert.GreaterOrEqual(t, len(teams), 1, "worker1 should be in at least 1 team")
}

// TestIntegration_Persistence_AcrossRestart tests team persistence across simulated restart.
func TestIntegration_Persistence_AcrossRestart(t *testing.T) {
	coord, _, db := setupIntegrationEnv(t)
	ctx := context.Background()

	// Form team.
	tm, err := coord.FormTeam(ctx, FormTeamRequest{
		TeamID:      "persist-team",
		Name:        "persist-test",
		Goal:        "test persistence",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 2,
	})
	require.NoError(t, err)
	tm.Budget = 50.0
	require.NoError(t, tm.AddSpend(10.0))

	// Persist updated state.
	store, err := NewBoltStore(db, integrationLogger())
	require.NoError(t, err)
	require.NoError(t, store.Save(tm))

	// Simulate restart: create new coordinator with same DB.
	newPool := agentpool.New(integrationLogger())
	require.NoError(t, newPool.Add(&agentpool.Agent{
		DID: "did:leader", Name: "leader", PeerID: "peer-leader",
		Capabilities: []string{"coordinate"}, Status: agentpool.StatusHealthy,
	}))

	newCoord := NewCoordinator(CoordinatorConfig{
		Pool:   newPool,
		Store:  store,
		Logger: integrationLogger(),
		Bus:    eventbus.New(),
		InvokeFn: func(_ context.Context, _, _ string, _ map[string]interface{}) (map[string]interface{}, error) {
			return nil, nil
		},
		Selector: agentpool.NewSelector(newPool, agentpool.DefaultWeights()),
	})

	// Load persisted teams.
	require.NoError(t, newCoord.LoadPersistedTeams())

	// Verify team was restored.
	restored, err := newCoord.GetTeam("persist-team")
	require.NoError(t, err)
	assert.Equal(t, "persist-test", restored.Name)
	assert.Equal(t, "test persistence", restored.Goal)
	assert.Equal(t, StatusActive, restored.Status)
	assert.GreaterOrEqual(t, restored.MemberCount(), 2)
}

// TestIntegration_ShuttingDown_BlocksTasks tests that shutting down blocks new task delegation.
func TestIntegration_ShuttingDown_BlocksTasks(t *testing.T) {
	coord, _, _ := setupIntegrationEnv(t)
	ctx := context.Background()

	// Form team.
	tm, err := coord.FormTeam(ctx, FormTeamRequest{
		TeamID:      "shutdown-block-team",
		Name:        "shutdown-block",
		Goal:        "test shutdown blocks tasks",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 1,
	})
	require.NoError(t, err)

	// Set status to shutting down manually.
	tm.mu.Lock()
	tm.Status = StatusShuttingDown
	tm.mu.Unlock()

	// Attempt to delegate should fail.
	_, err = coord.DelegateTask(ctx, "shutdown-block-team", "web_search", nil)
	assert.ErrorIs(t, err, ErrTeamShuttingDown)
}
