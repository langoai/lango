package team

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/p2p/agentpool"
)

func testLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

func setupCoordinator(t *testing.T) (*Coordinator, *agentpool.Pool) {
	t.Helper()
	pool := agentpool.New(testLogger())

	_ = pool.Add(&agentpool.Agent{
		DID:          "did:leader",
		Name:         "leader",
		PeerID:       "peer-leader",
		Capabilities: []string{"coordinate"},
		Status:       agentpool.StatusHealthy,
		TrustScore:   0.95,
	})
	_ = pool.Add(&agentpool.Agent{
		DID:          "did:worker1",
		Name:         "worker-1",
		PeerID:       "peer-w1",
		Capabilities: []string{"search"},
		Status:       agentpool.StatusHealthy,
		TrustScore:   0.8,
	})
	_ = pool.Add(&agentpool.Agent{
		DID:          "did:worker2",
		Name:         "worker-2",
		PeerID:       "peer-w2",
		Capabilities: []string{"search"},
		Status:       agentpool.StatusHealthy,
		TrustScore:   0.7,
	})

	invokeFn := func(_ context.Context, peerID, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"tool": toolName, "from": peerID}, nil
	}

	sel := agentpool.NewSelector(pool, agentpool.DefaultWeights())
	coord := NewCoordinator(CoordinatorConfig{
		Pool:     pool,
		Selector: sel,
		InvokeFn: invokeFn,
		Logger:   testLogger(),
	})

	return coord, pool
}

func TestFormTeam(t *testing.T) {
	coord, _ := setupCoordinator(t)

	tm, err := coord.FormTeam(context.Background(), FormTeamRequest{
		TeamID:      "t1",
		Name:        "search-team",
		Goal:        "find information",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 2,
	})
	if err != nil {
		t.Fatalf("FormTeam() error = %v", err)
	}

	if tm.Status != StatusActive {
		t.Errorf("Status = %q, want %q", tm.Status, StatusActive)
	}

	// Should have leader + up to 2 workers.
	if tm.MemberCount() < 2 {
		t.Errorf("MemberCount() = %d, want >= 2", tm.MemberCount())
	}
}

func TestDelegateTask(t *testing.T) {
	coord, _ := setupCoordinator(t)

	_, err := coord.FormTeam(context.Background(), FormTeamRequest{
		TeamID:      "t1",
		Name:        "search-team",
		Goal:        "find info",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 2,
	})
	if err != nil {
		t.Fatalf("FormTeam() error = %v", err)
	}

	results, err := coord.DelegateTask(context.Background(), "t1", "web_search", map[string]interface{}{"q": "test"})
	if err != nil {
		t.Fatalf("DelegateTask() error = %v", err)
	}

	if len(results) == 0 {
		t.Fatal("DelegateTask() returned empty results")
	}

	for _, r := range results {
		if r.Err != nil {
			t.Errorf("result from %s has error: %v", r.MemberDID, r.Err)
		}
		if r.Result == nil {
			t.Errorf("result from %s is nil", r.MemberDID)
		}
		if r.Duration == 0 {
			t.Errorf("result from %s has zero duration", r.MemberDID)
		}
	}
}

func TestCollectResults_MajorityResolver(t *testing.T) {
	results := []TaskResult{
		{MemberDID: "did:1", Result: map[string]interface{}{"answer": "42"}, Duration: time.Millisecond},
		{MemberDID: "did:2", Err: errors.New("timeout"), Duration: 5 * time.Second},
		{MemberDID: "did:3", Result: map[string]interface{}{"answer": "42"}, Duration: 2 * time.Millisecond},
	}

	resolved, err := MajorityResolver(results)
	if err != nil {
		t.Fatalf("MajorityResolver() error = %v", err)
	}
	if resolved["answer"] != "42" {
		t.Errorf("answer = %v, want 42", resolved["answer"])
	}
}

func TestCollectResults_AllFailed(t *testing.T) {
	results := []TaskResult{
		{MemberDID: "did:1", Err: errors.New("fail")},
		{MemberDID: "did:2", Err: errors.New("fail")},
	}

	_, err := MajorityResolver(results)
	if err == nil {
		t.Error("MajorityResolver() should return error when all failed")
	}
}

func TestFastestResolver(t *testing.T) {
	results := []TaskResult{
		{MemberDID: "did:1", Result: map[string]interface{}{"v": 1}, Duration: 100 * time.Millisecond},
		{MemberDID: "did:2", Err: errors.New("timeout")},
	}

	resolved, err := FastestResolver(results)
	if err != nil {
		t.Fatalf("FastestResolver() error = %v", err)
	}
	if resolved["v"] != 1 {
		t.Errorf("v = %v, want 1", resolved["v"])
	}
}

func TestDisbandTeam(t *testing.T) {
	coord, _ := setupCoordinator(t)

	_, err := coord.FormTeam(context.Background(), FormTeamRequest{
		TeamID:      "t1",
		Name:        "temp-team",
		Goal:        "temporary",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 1,
	})
	if err != nil {
		t.Fatalf("FormTeam() error = %v", err)
	}

	if err := coord.DisbandTeam("t1"); err != nil {
		t.Fatalf("DisbandTeam() error = %v", err)
	}

	_, err = coord.GetTeam("t1")
	if err != ErrTeamNotFound {
		t.Errorf("GetTeam after disband: got %v, want ErrTeamNotFound", err)
	}
}

func TestDisbandTeam_NotFound(t *testing.T) {
	coord, _ := setupCoordinator(t)

	err := coord.DisbandTeam("nonexistent")
	if err != ErrTeamNotFound {
		t.Errorf("DisbandTeam nonexistent: got %v, want ErrTeamNotFound", err)
	}
}

func TestResolveConflict(t *testing.T) {
	tests := []struct {
		give     string
		strategy ConflictStrategy
		results  []TaskResultSummary
		wantErr  bool
	}{
		{
			give:     "trust weighted picks first successful",
			strategy: StrategyTrustWeighted,
			results: []TaskResultSummary{
				{AgentDID: "did:1", Success: true, Result: "a", DurationMs: 200},
				{AgentDID: "did:2", Success: true, Result: "b", DurationMs: 100},
			},
			wantErr: false,
		},
		{
			give:     "majority vote returns first success",
			strategy: StrategyMajorityVote,
			results: []TaskResultSummary{
				{AgentDID: "did:1", Success: false, Error: "timeout"},
				{AgentDID: "did:2", Success: true, Result: "ok"},
			},
			wantErr: false,
		},
		{
			give:     "fail on conflict with same results",
			strategy: StrategyFailOnConflict,
			results: []TaskResultSummary{
				{AgentDID: "did:1", Success: true, Result: "same"},
				{AgentDID: "did:2", Success: true, Result: "same"},
			},
			wantErr: false,
		},
		{
			give:     "fail on conflict with different results",
			strategy: StrategyFailOnConflict,
			results: []TaskResultSummary{
				{AgentDID: "did:1", Success: true, Result: "a"},
				{AgentDID: "did:2", Success: true, Result: "b"},
			},
			wantErr: true,
		},
		{
			give:     "empty results",
			strategy: StrategyMajorityVote,
			results:  nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result, err := ResolveConflict(tt.strategy, tt.results)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("result is nil")
			}
			if !result.Success {
				t.Error("result should be successful")
			}
		})
	}
}

func TestListTeams(t *testing.T) {
	coord, _ := setupCoordinator(t)

	_, _ = coord.FormTeam(context.Background(), FormTeamRequest{
		TeamID: "t1", Name: "team-1", Goal: "goal", LeaderDID: "did:leader",
		Capability: "search", MemberCount: 1,
	})
	_, _ = coord.FormTeam(context.Background(), FormTeamRequest{
		TeamID: "t2", Name: "team-2", Goal: "goal", LeaderDID: "did:leader",
		Capability: "search", MemberCount: 1,
	})

	teams := coord.ListTeams()
	if len(teams) != 2 {
		t.Errorf("ListTeams() returned %d teams, want 2", len(teams))
	}
}
