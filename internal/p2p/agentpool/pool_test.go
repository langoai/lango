package agentpool

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

func TestPool_AddAndGet(t *testing.T) {
	t.Parallel()

	p := New(testLogger())

	a := &Agent{DID: "did:test:1", Name: "agent-1", Capabilities: []string{"search"}}
	require.NoError(t, p.Add(a))

	got := p.Get("did:test:1")
	require.NotNil(t, got)
	assert.Equal(t, "agent-1", got.Name)
}

func TestPool_AddDuplicate(t *testing.T) {
	t.Parallel()

	p := New(testLogger())

	a := &Agent{DID: "did:test:1", Name: "agent-1"}
	require.NoError(t, p.Add(a))

	err := p.Add(a)
	assert.ErrorIs(t, err, ErrAgentExists)
}

func TestPool_Remove(t *testing.T) {
	t.Parallel()

	p := New(testLogger())

	a := &Agent{DID: "did:test:1", Name: "agent-1"}
	_ = p.Add(a)

	p.Remove("did:test:1")
	assert.Nil(t, p.Get("did:test:1"))
	assert.Equal(t, 0, p.Size())
}

func TestPool_FindByCapability(t *testing.T) {
	t.Parallel()

	p := New(testLogger())

	_ = p.Add(&Agent{DID: "did:1", Name: "search-agent", Capabilities: []string{"search"}, Status: StatusHealthy})
	_ = p.Add(&Agent{DID: "did:2", Name: "code-agent", Capabilities: []string{"code"}, Status: StatusHealthy})
	_ = p.Add(&Agent{DID: "did:3", Name: "dead-agent", Capabilities: []string{"search"}, Status: StatusUnhealthy})

	results := p.FindByCapability("search")
	require.Len(t, results, 1, "unhealthy excluded")
	assert.Equal(t, "did:1", results[0].DID)
}

func TestPool_EvictStale(t *testing.T) {
	t.Parallel()

	p := New(testLogger())

	old := &Agent{DID: "did:old", Name: "old", LastSeen: time.Now().Add(-2 * time.Hour)}
	fresh := &Agent{DID: "did:fresh", Name: "fresh", LastSeen: time.Now()}

	_ = p.Add(old)
	_ = p.Add(fresh)

	evicted := p.EvictStale(1 * time.Hour)
	assert.Equal(t, 1, evicted)
	assert.Equal(t, 1, p.Size())
}

func TestPool_MarkHealthy(t *testing.T) {
	t.Parallel()

	p := New(testLogger())
	_ = p.Add(&Agent{DID: "did:1", Status: StatusUnknown})

	p.MarkHealthy("did:1", 50*time.Millisecond)

	a := p.Get("did:1")
	assert.Equal(t, StatusHealthy, a.Status)
	assert.Equal(t, 50*time.Millisecond, a.Latency)
}

func TestPool_MarkUnhealthy(t *testing.T) {
	t.Parallel()

	p := New(testLogger())
	_ = p.Add(&Agent{DID: "did:1", Status: StatusHealthy})

	// First two failures → degraded.
	p.MarkUnhealthy("did:1")
	assert.Equal(t, StatusDegraded, p.Get("did:1").Status)
	p.MarkUnhealthy("did:1")

	// Third failure → unhealthy.
	p.MarkUnhealthy("did:1")
	assert.Equal(t, StatusUnhealthy, p.Get("did:1").Status)
}

func TestSelector_Select(t *testing.T) {
	t.Parallel()

	p := New(testLogger())
	_ = p.Add(&Agent{
		DID: "did:1", Name: "fast-trusted", Capabilities: []string{"search"},
		Status: StatusHealthy, TrustScore: 0.9, Latency: 10 * time.Millisecond,
	})
	_ = p.Add(&Agent{
		DID: "did:2", Name: "slow-untrusted", Capabilities: []string{"search"},
		Status: StatusHealthy, TrustScore: 0.3, Latency: 5 * time.Second,
	})

	sel := NewSelector(p, DefaultWeights())
	best, err := sel.Select("search")
	require.NoError(t, err)
	assert.Equal(t, "did:1", best.DID, "fast-trusted should win")
}

func TestSelector_SelectN(t *testing.T) {
	t.Parallel()

	p := New(testLogger())
	_ = p.Add(&Agent{DID: "did:1", Name: "a", Capabilities: []string{"code"}, Status: StatusHealthy, TrustScore: 0.9})
	_ = p.Add(&Agent{DID: "did:2", Name: "b", Capabilities: []string{"code"}, Status: StatusHealthy, TrustScore: 0.5})
	_ = p.Add(&Agent{DID: "did:3", Name: "c", Capabilities: []string{"code"}, Status: StatusHealthy, TrustScore: 0.7})

	sel := NewSelector(p, DefaultWeights())
	top2, err := sel.SelectN("code", 2)
	require.NoError(t, err)
	require.Len(t, top2, 2)
	// Highest trust first.
	assert.Equal(t, "did:1", top2[0].DID)
}

func TestSelector_NoAgents(t *testing.T) {
	t.Parallel()

	p := New(testLogger())
	sel := NewSelector(p, DefaultWeights())

	_, err := sel.Select("nonexistent")
	assert.ErrorIs(t, err, ErrNoAgents)
}

func TestPool_UpdatePerformance(t *testing.T) {
	t.Parallel()

	p := New(testLogger())
	_ = p.Add(&Agent{DID: "did:1", Status: StatusHealthy})

	p.UpdatePerformance("did:1", 100.0, true)
	p.UpdatePerformance("did:1", 200.0, false)

	a := p.Get("did:1")
	assert.Equal(t, 2, a.Performance.TotalCalls)
	// Average of 100 and 200 = 150.
	assert.InDelta(t, 150.0, a.Performance.AvgLatencyMs, 1.0)
	// 1 success out of 2 = 0.5.
	assert.InDelta(t, 0.5, a.Performance.SuccessRate, 0.01)
}

func TestSelector_ScoreWithCaps(t *testing.T) {
	t.Parallel()

	p := New(testLogger())
	_ = p.Add(&Agent{
		DID: "did:1", Name: "multi", Capabilities: []string{"search", "code"},
		Status: StatusHealthy, TrustScore: 0.8, Performance: AgentPerformance{SuccessRate: 0.9},
	})
	_ = p.Add(&Agent{
		DID: "did:2", Name: "single", Capabilities: []string{"search"},
		Status: StatusHealthy, TrustScore: 0.8, Performance: AgentPerformance{SuccessRate: 0.9},
	})

	sel := NewSelector(p, DefaultWeights())
	s1 := sel.ScoreWithCaps(p.Get("did:1"), []string{"search", "code"})
	s2 := sel.ScoreWithCaps(p.Get("did:2"), []string{"search", "code"})

	assert.Greater(t, s1, s2, "agent with both caps should score higher")
}

func TestSelector_SelectBest(t *testing.T) {
	t.Parallel()

	p := New(testLogger())
	agents := []*Agent{
		{DID: "did:1", Capabilities: []string{"code"}, Status: StatusHealthy, TrustScore: 0.5, Performance: AgentPerformance{SuccessRate: 0.5}},
		{DID: "did:2", Capabilities: []string{"code"}, Status: StatusHealthy, TrustScore: 0.9, Performance: AgentPerformance{SuccessRate: 0.9}},
		{DID: "did:3", Capabilities: []string{"code"}, Status: StatusHealthy, TrustScore: 0.7, Performance: AgentPerformance{SuccessRate: 0.7}},
	}
	for _, a := range agents {
		_ = p.Add(a)
	}

	sel := NewSelector(p, DefaultWeights())
	best := sel.SelectBest(agents, []string{"code"}, 2)
	require.Len(t, best, 2)
	assert.Equal(t, "did:2", best[0].DID)
}

func TestHealthChecker(t *testing.T) {
	t.Parallel()

	p := New(testLogger())
	_ = p.Add(&Agent{DID: "did:1", Status: StatusUnknown})

	checkCalled := make(chan struct{}, 1)
	checkFn := func(_ context.Context, a *Agent) (time.Duration, error) {
		select {
		case checkCalled <- struct{}{}:
		default:
		}
		return 5 * time.Millisecond, nil
	}

	hc := NewHealthChecker(p, checkFn, 50*time.Millisecond, testLogger())
	var wg sync.WaitGroup
	hc.Start(&wg)

	select {
	case <-checkCalled:
	case <-time.After(2 * time.Second):
		t.Fatal("health check function was not called within timeout")
	}

	hc.Stop()
	wg.Wait()

	a := p.Get("did:1")
	assert.Equal(t, StatusHealthy, a.Status)
}
