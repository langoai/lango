package team

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/eventbus"
)

func setupHealthMonitor(t *testing.T, invokeFn InvokeFunc) (*HealthMonitor, *Coordinator, *eventbus.Bus) {
	t.Helper()

	coord, _, bus := setupCoordinatorWithBus(t)

	hm := NewHealthMonitor(HealthMonitorConfig{
		Coordinator: coord,
		Bus:         bus,
		Logger:      testLogger(),
		Interval:    50 * time.Millisecond, // fast for tests
		MaxMissed:   3,
		InvokeFn:    invokeFn,
	})

	return hm, coord, bus
}

func TestHealthMonitor_HealthyMember(t *testing.T) {
	t.Parallel()

	invokeFn := func(_ context.Context, peerID, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"status": "ok"}, nil
	}

	hm, coord, bus := setupHealthMonitor(t, invokeFn)

	_, err := coord.FormTeam(context.Background(), FormTeamRequest{
		TeamID:      "t-healthy",
		Name:        "health-team",
		Goal:        "test health",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 1,
	})
	require.NoError(t, err)

	var mu sync.Mutex
	var unhealthy []eventbus.TeamMemberUnhealthyEvent
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamMemberUnhealthyEvent) {
		mu.Lock()
		defer mu.Unlock()
		unhealthy = append(unhealthy, ev)
	})

	// Start, let a few cycles run, then stop.
	require.NoError(t, hm.Start(context.Background(), &sync.WaitGroup{}))
	time.Sleep(200 * time.Millisecond)
	require.NoError(t, hm.Stop(context.Background()))

	mu.Lock()
	defer mu.Unlock()
	assert.Empty(t, unhealthy, "healthy member should not trigger unhealthy event")
}

func TestHealthMonitor_MissedPingsIncrement(t *testing.T) {
	t.Parallel()

	invokeFn := func(_ context.Context, peerID, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
		return nil, errors.New("timeout")
	}

	hm, coord, bus := setupHealthMonitor(t, invokeFn)

	_, err := coord.FormTeam(context.Background(), FormTeamRequest{
		TeamID:      "t-miss",
		Name:        "miss-team",
		Goal:        "test missed pings",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 1,
	})
	require.NoError(t, err)

	var mu sync.Mutex
	var unhealthy []eventbus.TeamMemberUnhealthyEvent
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamMemberUnhealthyEvent) {
		mu.Lock()
		defer mu.Unlock()
		unhealthy = append(unhealthy, ev)
	})

	require.NoError(t, hm.Start(context.Background(), &sync.WaitGroup{}))
	// Wait enough cycles for maxMissed (3) to be reached: 3 * 50ms + margin.
	time.Sleep(300 * time.Millisecond)
	require.NoError(t, hm.Stop(context.Background()))

	mu.Lock()
	defer mu.Unlock()
	assert.NotEmpty(t, unhealthy, "should publish unhealthy event after max missed pings")
	assert.GreaterOrEqual(t, unhealthy[0].MissedPings, 3)
}

func TestHealthMonitor_TaskCompletionResetsCounter(t *testing.T) {
	t.Parallel()

	var callCount int
	var mu2 sync.Mutex

	invokeFn := func(_ context.Context, peerID, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
		mu2.Lock()
		callCount++
		c := callCount
		mu2.Unlock()

		// Fail first 2, then succeed.
		if c <= 2 {
			return nil, errors.New("timeout")
		}
		return map[string]interface{}{"status": "ok"}, nil
	}

	hm, coord, bus := setupHealthMonitor(t, invokeFn)

	_, err := coord.FormTeam(context.Background(), FormTeamRequest{
		TeamID:      "t-reset",
		Name:        "reset-team",
		Goal:        "test counter reset",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 1,
	})
	require.NoError(t, err)

	var mu sync.Mutex
	var unhealthy []eventbus.TeamMemberUnhealthyEvent
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamMemberUnhealthyEvent) {
		mu.Lock()
		defer mu.Unlock()
		unhealthy = append(unhealthy, ev)
	})

	require.NoError(t, hm.Start(context.Background(), &sync.WaitGroup{}))

	// Let 2 failed pings happen, then simulate task completion to reset counters.
	time.Sleep(150 * time.Millisecond)
	bus.Publish(eventbus.TeamTaskCompletedEvent{
		TeamID:     "t-reset",
		ToolName:   "test",
		Successful: 1,
	})

	// Let more cycles run.
	time.Sleep(200 * time.Millisecond)
	require.NoError(t, hm.Stop(context.Background()))

	mu.Lock()
	defer mu.Unlock()
	// Counter was reset by task completion before reaching maxMissed=3.
	assert.Empty(t, unhealthy, "task completion should reset miss counter")
}

func TestHealthMonitor_NameAndLifecycle(t *testing.T) {
	t.Parallel()

	hm := NewHealthMonitor(HealthMonitorConfig{
		Coordinator: &Coordinator{teams: make(map[string]*Team)},
		Bus:         eventbus.New(),
		Logger:      testLogger(),
		Interval:    time.Hour, // long interval so no checks run
		MaxMissed:   3,
	})

	assert.Equal(t, "team-health-monitor", hm.Name())
	require.NoError(t, hm.Start(context.Background(), &sync.WaitGroup{}))
	require.NoError(t, hm.Stop(context.Background()))
}

func TestHealthMonitor_DefaultValues(t *testing.T) {
	t.Parallel()

	hm := NewHealthMonitor(HealthMonitorConfig{
		Coordinator: &Coordinator{teams: make(map[string]*Team)},
		Bus:         eventbus.New(),
		Logger:      testLogger(),
	})

	assert.Equal(t, 30*time.Second, hm.interval)
	assert.Equal(t, 3, hm.maxMissed)
}

func TestHealthMonitor_GitStateTracking(t *testing.T) {
	t.Parallel()

	hm := NewHealthMonitor(HealthMonitorConfig{
		Coordinator: &Coordinator{teams: make(map[string]*Team)},
		Bus:         eventbus.New(),
		Logger:      testLogger(),
		Interval:    time.Hour,
		MaxMissed:   3,
	})

	// Manually update git state.
	hm.updateGitState("ws-1", "did:agent-1", "aaa")
	hm.updateGitState("ws-1", "did:agent-2", "aaa")
	hm.updateGitState("ws-1", "did:agent-3", "bbb")

	divergent := hm.DetectDivergence("ws-1")
	assert.Len(t, divergent, 1)
	assert.Equal(t, "did:agent-3", divergent[0].MemberDID)
	assert.Equal(t, "bbb", divergent[0].MemberHead)
	assert.Equal(t, "aaa", divergent[0].MajorityHead)
}

func TestHealthMonitor_DetectDivergence_NoMembers(t *testing.T) {
	t.Parallel()

	hm := NewHealthMonitor(HealthMonitorConfig{
		Coordinator: &Coordinator{teams: make(map[string]*Team)},
		Bus:         eventbus.New(),
		Logger:      testLogger(),
		Interval:    time.Hour,
		MaxMissed:   3,
	})

	divergent := hm.DetectDivergence("ws-nonexistent")
	assert.Nil(t, divergent)
}

func TestHealthMonitor_DetectDivergence_AllSame(t *testing.T) {
	t.Parallel()

	hm := NewHealthMonitor(HealthMonitorConfig{
		Coordinator: &Coordinator{teams: make(map[string]*Team)},
		Bus:         eventbus.New(),
		Logger:      testLogger(),
		Interval:    time.Hour,
		MaxMissed:   3,
	})

	hm.updateGitState("ws-1", "did:a", "same-hash")
	hm.updateGitState("ws-1", "did:b", "same-hash")

	divergent := hm.DetectDivergence("ws-1")
	assert.Empty(t, divergent)
}

func TestHealthMonitor_UpdateGitState_EmptyHash(t *testing.T) {
	t.Parallel()

	hm := NewHealthMonitor(HealthMonitorConfig{
		Coordinator: &Coordinator{teams: make(map[string]*Team)},
		Bus:         eventbus.New(),
		Logger:      testLogger(),
		Interval:    time.Hour,
		MaxMissed:   3,
	})

	// Empty hash should be a no-op.
	hm.updateGitState("ws-1", "did:a", "")
	divergent := hm.DetectDivergence("ws-1")
	assert.Nil(t, divergent)
}
