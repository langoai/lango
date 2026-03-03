package agentpool

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

func testLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

func TestPool_AddAndGet(t *testing.T) {
	p := New(testLogger())

	a := &Agent{DID: "did:test:1", Name: "agent-1", Capabilities: []string{"search"}}
	if err := p.Add(a); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	got := p.Get("did:test:1")
	if got == nil {
		t.Fatal("Get() returned nil")
	}
	if got.Name != "agent-1" {
		t.Errorf("Name = %q, want %q", got.Name, "agent-1")
	}
}

func TestPool_AddDuplicate(t *testing.T) {
	p := New(testLogger())

	a := &Agent{DID: "did:test:1", Name: "agent-1"}
	if err := p.Add(a); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	err := p.Add(a)
	if err != ErrAgentExists {
		t.Errorf("Add duplicate: got %v, want ErrAgentExists", err)
	}
}

func TestPool_Remove(t *testing.T) {
	p := New(testLogger())

	a := &Agent{DID: "did:test:1", Name: "agent-1"}
	_ = p.Add(a)

	p.Remove("did:test:1")
	if p.Get("did:test:1") != nil {
		t.Error("Get() after Remove should return nil")
	}
	if p.Size() != 0 {
		t.Errorf("Size() = %d, want 0", p.Size())
	}
}

func TestPool_FindByCapability(t *testing.T) {
	p := New(testLogger())

	_ = p.Add(&Agent{DID: "did:1", Name: "search-agent", Capabilities: []string{"search"}, Status: StatusHealthy})
	_ = p.Add(&Agent{DID: "did:2", Name: "code-agent", Capabilities: []string{"code"}, Status: StatusHealthy})
	_ = p.Add(&Agent{DID: "did:3", Name: "dead-agent", Capabilities: []string{"search"}, Status: StatusUnhealthy})

	results := p.FindByCapability("search")
	if len(results) != 1 {
		t.Fatalf("FindByCapability(search) = %d agents, want 1 (unhealthy excluded)", len(results))
	}
	if results[0].DID != "did:1" {
		t.Errorf("DID = %q, want %q", results[0].DID, "did:1")
	}
}

func TestPool_EvictStale(t *testing.T) {
	p := New(testLogger())

	old := &Agent{DID: "did:old", Name: "old", LastSeen: time.Now().Add(-2 * time.Hour)}
	fresh := &Agent{DID: "did:fresh", Name: "fresh", LastSeen: time.Now()}

	_ = p.Add(old)
	_ = p.Add(fresh)

	evicted := p.EvictStale(1 * time.Hour)
	if evicted != 1 {
		t.Errorf("EvictStale() = %d, want 1", evicted)
	}
	if p.Size() != 1 {
		t.Errorf("Size() = %d, want 1", p.Size())
	}
}

func TestPool_MarkHealthy(t *testing.T) {
	p := New(testLogger())
	_ = p.Add(&Agent{DID: "did:1", Status: StatusUnknown})

	p.MarkHealthy("did:1", 50*time.Millisecond)

	a := p.Get("did:1")
	if a.Status != StatusHealthy {
		t.Errorf("Status = %q, want %q", a.Status, StatusHealthy)
	}
	if a.Latency != 50*time.Millisecond {
		t.Errorf("Latency = %v, want 50ms", a.Latency)
	}
}

func TestPool_MarkUnhealthy(t *testing.T) {
	p := New(testLogger())
	_ = p.Add(&Agent{DID: "did:1", Status: StatusHealthy})

	// First two failures → degraded.
	p.MarkUnhealthy("did:1")
	if p.Get("did:1").Status != StatusDegraded {
		t.Errorf("after 1 failure: Status = %q, want %q", p.Get("did:1").Status, StatusDegraded)
	}
	p.MarkUnhealthy("did:1")

	// Third failure → unhealthy.
	p.MarkUnhealthy("did:1")
	if p.Get("did:1").Status != StatusUnhealthy {
		t.Errorf("after 3 failures: Status = %q, want %q", p.Get("did:1").Status, StatusUnhealthy)
	}
}

func TestSelector_Select(t *testing.T) {
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
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if best.DID != "did:1" {
		t.Errorf("Select() = %q, want %q (fast-trusted should win)", best.DID, "did:1")
	}
}

func TestSelector_SelectN(t *testing.T) {
	p := New(testLogger())
	_ = p.Add(&Agent{DID: "did:1", Name: "a", Capabilities: []string{"code"}, Status: StatusHealthy, TrustScore: 0.9})
	_ = p.Add(&Agent{DID: "did:2", Name: "b", Capabilities: []string{"code"}, Status: StatusHealthy, TrustScore: 0.5})
	_ = p.Add(&Agent{DID: "did:3", Name: "c", Capabilities: []string{"code"}, Status: StatusHealthy, TrustScore: 0.7})

	sel := NewSelector(p, DefaultWeights())
	top2, err := sel.SelectN("code", 2)
	if err != nil {
		t.Fatalf("SelectN() error = %v", err)
	}
	if len(top2) != 2 {
		t.Fatalf("SelectN() returned %d agents, want 2", len(top2))
	}
	// Highest trust first.
	if top2[0].DID != "did:1" {
		t.Errorf("top2[0].DID = %q, want %q", top2[0].DID, "did:1")
	}
}

func TestSelector_NoAgents(t *testing.T) {
	p := New(testLogger())
	sel := NewSelector(p, DefaultWeights())

	_, err := sel.Select("nonexistent")
	if err != ErrNoAgents {
		t.Errorf("Select() error = %v, want ErrNoAgents", err)
	}
}

func TestPool_UpdatePerformance(t *testing.T) {
	p := New(testLogger())
	_ = p.Add(&Agent{DID: "did:1", Status: StatusHealthy})

	p.UpdatePerformance("did:1", 100.0, true)
	p.UpdatePerformance("did:1", 200.0, false)

	a := p.Get("did:1")
	if a.Performance.TotalCalls != 2 {
		t.Errorf("TotalCalls = %d, want 2", a.Performance.TotalCalls)
	}
	// Average of 100 and 200 = 150.
	if a.Performance.AvgLatencyMs < 149.0 || a.Performance.AvgLatencyMs > 151.0 {
		t.Errorf("AvgLatencyMs = %f, want ~150", a.Performance.AvgLatencyMs)
	}
	// 1 success out of 2 = 0.5.
	if a.Performance.SuccessRate < 0.49 || a.Performance.SuccessRate > 0.51 {
		t.Errorf("SuccessRate = %f, want ~0.5", a.Performance.SuccessRate)
	}
}

func TestSelector_ScoreWithCaps(t *testing.T) {
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

	if s1 <= s2 {
		t.Errorf("agent with both caps (%f) should score higher than agent with one cap (%f)", s1, s2)
	}
}

func TestSelector_SelectBest(t *testing.T) {
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
	if len(best) != 2 {
		t.Fatalf("SelectBest() returned %d, want 2", len(best))
	}
	if best[0].DID != "did:2" {
		t.Errorf("best[0].DID = %q, want %q", best[0].DID, "did:2")
	}
}

func TestHealthChecker(t *testing.T) {
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
	if a.Status != StatusHealthy {
		t.Errorf("after health check: Status = %q, want %q", a.Status, StatusHealthy)
	}
}

