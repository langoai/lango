package team

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	t.Parallel()

	coord, _ := setupCoordinator(t)

	tm, err := coord.FormTeam(context.Background(), FormTeamRequest{
		TeamID:      "t1",
		Name:        "search-team",
		Goal:        "find information",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 2,
	})
	require.NoError(t, err)
	assert.Equal(t, StatusActive, tm.Status)

	// Should have leader + up to 2 workers.
	assert.GreaterOrEqual(t, tm.MemberCount(), 2)
}

func TestDelegateTask(t *testing.T) {
	t.Parallel()

	coord, _ := setupCoordinator(t)

	_, err := coord.FormTeam(context.Background(), FormTeamRequest{
		TeamID:      "t1",
		Name:        "search-team",
		Goal:        "find info",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 2,
	})
	require.NoError(t, err)

	results, err := coord.DelegateTask(context.Background(), "t1", "web_search", map[string]interface{}{"q": "test"})
	require.NoError(t, err)
	require.NotEmpty(t, results)

	for _, r := range results {
		assert.NoError(t, r.Err, "result from %s", r.MemberDID)
		assert.NotNil(t, r.Result, "result from %s", r.MemberDID)
		assert.NotZero(t, r.Duration, "result from %s", r.MemberDID)
	}
}

func TestCollectResults_MajorityResolver(t *testing.T) {
	t.Parallel()

	results := []TaskResult{
		{MemberDID: "did:1", Result: map[string]interface{}{"answer": "42"}, Duration: time.Millisecond},
		{MemberDID: "did:2", Err: errors.New("timeout"), Duration: 5 * time.Second},
		{MemberDID: "did:3", Result: map[string]interface{}{"answer": "42"}, Duration: 2 * time.Millisecond},
	}

	resolved, err := MajorityResolver(results)
	require.NoError(t, err)
	assert.Equal(t, "42", resolved["answer"])
}

func TestCollectResults_AllFailed(t *testing.T) {
	t.Parallel()

	results := []TaskResult{
		{MemberDID: "did:1", Err: errors.New("fail")},
		{MemberDID: "did:2", Err: errors.New("fail")},
	}

	_, err := MajorityResolver(results)
	assert.Error(t, err)
}

func TestFastestResolver(t *testing.T) {
	t.Parallel()

	results := []TaskResult{
		{MemberDID: "did:1", Result: map[string]interface{}{"v": 1}, Duration: 100 * time.Millisecond},
		{MemberDID: "did:2", Err: errors.New("timeout")},
	}

	resolved, err := FastestResolver(results)
	require.NoError(t, err)
	assert.Equal(t, 1, resolved["v"])
}

func TestDisbandTeam(t *testing.T) {
	t.Parallel()

	coord, _ := setupCoordinator(t)

	_, err := coord.FormTeam(context.Background(), FormTeamRequest{
		TeamID:      "t1",
		Name:        "temp-team",
		Goal:        "temporary",
		LeaderDID:   "did:leader",
		Capability:  "search",
		MemberCount: 1,
	})
	require.NoError(t, err)

	require.NoError(t, coord.DisbandTeam("t1"))

	_, err = coord.GetTeam("t1")
	assert.ErrorIs(t, err, ErrTeamNotFound)
}

func TestDisbandTeam_NotFound(t *testing.T) {
	t.Parallel()

	coord, _ := setupCoordinator(t)

	err := coord.DisbandTeam("nonexistent")
	assert.ErrorIs(t, err, ErrTeamNotFound)
}

func TestResolveConflict(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			result, err := ResolveConflict(tt.strategy, tt.results)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.True(t, result.Success)
		})
	}
}

func TestListTeams(t *testing.T) {
	t.Parallel()

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
	assert.Len(t, teams, 2)
}
