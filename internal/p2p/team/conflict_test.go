package team

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveConflict_EmptyResults(t *testing.T) {
	t.Parallel()

	strategies := []ConflictStrategy{
		StrategyTrustWeighted,
		StrategyMajorityVote,
		StrategyLeaderDecides,
		StrategyFailOnConflict,
	}

	for _, s := range strategies {
		t.Run(string(s), func(t *testing.T) {
			t.Parallel()

			got, err := ResolveConflict(s, nil)
			assert.Nil(t, got)
			assert.ErrorIs(t, err, ErrConflict)
		})
	}
}

func TestResolveConflict_TrustWeighted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		results   []TaskResultSummary
		wantDID   string
		wantErr   bool
		wantErrIs error
	}{
		{
			give: "picks fastest successful agent",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:slow", Success: true, Result: "ok", DurationMs: 500},
				{TaskID: "t1", AgentDID: "did:fast", Success: true, Result: "ok", DurationMs: 100},
				{TaskID: "t1", AgentDID: "did:mid", Success: true, Result: "ok", DurationMs: 300},
			},
			wantDID: "did:fast",
		},
		{
			give: "skips failed results",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:fail", Success: false, DurationMs: 10},
				{TaskID: "t1", AgentDID: "did:ok", Success: true, Result: "ok", DurationMs: 200},
			},
			wantDID: "did:ok",
		},
		{
			give: "single successful result",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:only", Success: true, Result: "ok", DurationMs: 50},
			},
			wantDID: "did:only",
		},
		{
			give: "all failed",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:f1", Success: false, Error: "timeout"},
				{TaskID: "t1", AgentDID: "did:f2", Success: false, Error: "crash"},
			},
			wantErr:   true,
			wantErrIs: ErrConflict,
		},
		{
			give: "equal duration picks first encountered",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:a", Success: true, Result: "ok", DurationMs: 100},
				{TaskID: "t1", AgentDID: "did:b", Success: true, Result: "ok", DurationMs: 100},
			},
			wantDID: "did:a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got, err := ResolveConflict(StrategyTrustWeighted, tt.results)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErrIs)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantDID, got.AgentDID)
		})
	}
}

func TestResolveConflict_MajorityVote(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		results   []TaskResultSummary
		wantDID   string
		wantErr   bool
		wantErrIs error
	}{
		{
			give: "returns first successful result",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:a", Success: true, Result: "answer-a", DurationMs: 300},
				{TaskID: "t1", AgentDID: "did:b", Success: true, Result: "answer-b", DurationMs: 100},
			},
			wantDID: "did:a",
		},
		{
			give: "skips failures to find first success",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:fail", Success: false, Error: "timeout"},
				{TaskID: "t1", AgentDID: "did:ok", Success: true, Result: "ok"},
			},
			wantDID: "did:ok",
		},
		{
			give: "all failed",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:f1", Success: false, Error: "err1"},
				{TaskID: "t1", AgentDID: "did:f2", Success: false, Error: "err2"},
			},
			wantErr:   true,
			wantErrIs: ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got, err := ResolveConflict(StrategyMajorityVote, tt.results)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErrIs)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantDID, got.AgentDID)
		})
	}
}

func TestResolveConflict_LeaderDecides(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		results   []TaskResultSummary
		wantDID   string
		wantErr   bool
		wantErrIs error
	}{
		{
			give: "returns first successful result for leader review",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:a", Success: true, Result: "result-a"},
				{TaskID: "t1", AgentDID: "did:b", Success: true, Result: "result-b"},
			},
			wantDID: "did:a",
		},
		{
			give: "skips failures",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:fail", Success: false},
				{TaskID: "t1", AgentDID: "did:ok", Success: true, Result: "ok"},
			},
			wantDID: "did:ok",
		},
		{
			give: "all failed",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:f1", Success: false},
			},
			wantErr:   true,
			wantErrIs: ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got, err := ResolveConflict(StrategyLeaderDecides, tt.results)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErrIs)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantDID, got.AgentDID)
		})
	}
}

func TestResolveConflict_FailOnConflict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		results   []TaskResultSummary
		wantDID   string
		wantErr   bool
		wantErrIs error
	}{
		{
			give: "single successful result passes",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:only", Success: true, Result: "answer"},
			},
			wantDID: "did:only",
		},
		{
			give: "multiple agreeing results pass",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:a", Success: true, Result: "same-answer"},
				{TaskID: "t1", AgentDID: "did:b", Success: true, Result: "same-answer"},
				{TaskID: "t1", AgentDID: "did:c", Success: true, Result: "same-answer"},
			},
			wantDID: "did:a",
		},
		{
			give: "conflicting results return error",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:a", Success: true, Result: "answer-a"},
				{TaskID: "t1", AgentDID: "did:b", Success: true, Result: "answer-b"},
			},
			wantErr:   true,
			wantErrIs: ErrConflict,
		},
		{
			give: "ignores failed results when checking agreement",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:fail", Success: false, Result: "different"},
				{TaskID: "t1", AgentDID: "did:ok", Success: true, Result: "answer"},
			},
			wantDID: "did:ok",
		},
		{
			give: "all failed",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:f1", Success: false},
				{TaskID: "t1", AgentDID: "did:f2", Success: false},
			},
			wantErr:   true,
			wantErrIs: ErrConflict,
		},
		{
			give: "mixed success and failure with agreement passes",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:fail", Success: false, Error: "timeout"},
				{TaskID: "t1", AgentDID: "did:a", Success: true, Result: "same"},
				{TaskID: "t1", AgentDID: "did:b", Success: true, Result: "same"},
			},
			wantDID: "did:a",
		},
		{
			give: "mixed success and failure with disagreement fails",
			results: []TaskResultSummary{
				{TaskID: "t1", AgentDID: "did:fail", Success: false, Error: "timeout"},
				{TaskID: "t1", AgentDID: "did:a", Success: true, Result: "answer-a"},
				{TaskID: "t1", AgentDID: "did:b", Success: true, Result: "answer-b"},
			},
			wantErr:   true,
			wantErrIs: ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got, err := ResolveConflict(StrategyFailOnConflict, tt.results)
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErrIs)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantDID, got.AgentDID)
		})
	}
}

func TestResolveConflict_UnknownStrategyFallsBackToMajorityVote(t *testing.T) {
	t.Parallel()

	results := []TaskResultSummary{
		{TaskID: "t1", AgentDID: "did:fail", Success: false, Error: "err"},
		{TaskID: "t1", AgentDID: "did:ok", Success: true, Result: "answer"},
	}

	got, err := ResolveConflict("unknown_strategy", results)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "did:ok", got.AgentDID)
	assert.Equal(t, "answer", got.Result)
}

func TestResolveConflict_UnknownStrategyEmptyResults(t *testing.T) {
	t.Parallel()

	got, err := ResolveConflict("nonexistent", nil)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, ErrConflict)
}

func TestResolveConflict_TrustWeighted_PrefersFastestOverSlower(t *testing.T) {
	t.Parallel()

	// Verify that among mixed durations, the fastest successful agent wins
	// even if it appears last in the slice.
	results := []TaskResultSummary{
		{TaskID: "t1", AgentDID: "did:slow", Success: true, Result: "ok", DurationMs: 1000},
		{TaskID: "t1", AgentDID: "did:medium", Success: true, Result: "ok", DurationMs: 500},
		{TaskID: "t1", AgentDID: "did:fastest", Success: true, Result: "ok", DurationMs: 50},
	}

	got, err := ResolveConflict(StrategyTrustWeighted, results)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "did:fastest", got.AgentDID)
	assert.Equal(t, int64(50), got.DurationMs)
}

func TestResolveConflict_FailOnConflict_ThreeWayDisagreement(t *testing.T) {
	t.Parallel()

	results := []TaskResultSummary{
		{TaskID: "t1", AgentDID: "did:a", Success: true, Result: "alpha"},
		{TaskID: "t1", AgentDID: "did:b", Success: true, Result: "beta"},
		{TaskID: "t1", AgentDID: "did:c", Success: true, Result: "gamma"},
	}

	got, err := ResolveConflict(StrategyFailOnConflict, results)
	assert.Nil(t, got)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConflict)
	assert.Contains(t, err.Error(), "conflicting results from 3 agents")
}
