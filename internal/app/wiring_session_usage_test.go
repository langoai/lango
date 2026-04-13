package app

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/turnrunner"
)

// --- test helpers ---

// fakeExecutor records calls for verification.
type fakeExecutor struct {
	mu    sync.Mutex
	calls int
}

func (e *fakeExecutor) RunStreamingDetailed(
	_ context.Context,
	_, _ string,
	_ adk.ChunkCallback,
	_ ...adk.RunOption,
) (adk.RunReport, error) {
	e.mu.Lock()
	e.calls++
	e.mu.Unlock()
	return adk.RunReport{Response: "ok"}, nil
}

// sessionUsageStore is a session.Store stub that supports Get/Update with
// metadata tracking.
type sessionUsageStore struct {
	stubSessionStore
	mu       sync.Mutex
	sessions map[string]*session.Session
}

func newSessionUsageStore() *sessionUsageStore {
	return &sessionUsageStore{sessions: make(map[string]*session.Session)}
}

func (s *sessionUsageStore) Get(key string) (*session.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[key]
	if !ok {
		return nil, nil
	}
	// Return a copy to avoid test cross-talk.
	cp := *sess
	if sess.Metadata != nil {
		cp.Metadata = make(map[string]string, len(sess.Metadata))
		for k, v := range sess.Metadata {
			cp.Metadata[k] = v
		}
	}
	return &cp, nil
}

func (s *sessionUsageStore) Update(sess *session.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sess.Key] = sess
	return nil
}

func (s *sessionUsageStore) put(sess *session.Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sess.Key] = sess
}

// --- budgetRestoringExecutor tests ---

func TestBudgetRestoringExecutor_RestoresBaselineAndAccumulates(t *testing.T) {
	store := newSessionUsageStore()
	store.put(&session.Session{
		Key: "sess-1",
		Metadata: map[string]string{
			"usage:budget_turns":       "7",
			"usage:budget_delegations": "3",
		},
	})

	// Create a CoordinatingExecutor that wraps a fake.
	inner := &fakeExecutor{}
	budget := agentrt.NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   50,
		DelegationLimit: 15,
		AlertThreshold:  0.8,
	})
	coordExec := agentrt.NewCoordinatingExecutor(inner, nil, budget, nil, nil)
	wrapped := wrapWithBudgetRestore(coordExec, store)

	report, err := wrapped.RunStreamingDetailed(
		context.Background(), "sess-1", "hello", nil,
	)
	require.NoError(t, err)
	assert.Equal(t, "ok", report.Response)

	// After the run, session state should have restored baseline (7/3) + run delta (0/0).
	bre := wrapped.(*budgetRestoringExecutor)
	stateVal, ok := bre.sessionState.Load("sess-1")
	require.True(t, ok)
	state := stateVal.(*sessionBudgetState)
	state.mu.Lock()
	assert.Equal(t, int64(7), state.cumulativeTurns)
	assert.Equal(t, int64(3), state.cumulativeDeleg)
	state.mu.Unlock()
}

func TestBudgetRestoringExecutor_SkipsSubsequentRestores(t *testing.T) {
	store := newSessionUsageStore()
	store.put(&session.Session{
		Key: "sess-2",
		Metadata: map[string]string{
			"usage:budget_turns":       "5",
			"usage:budget_delegations": "2",
		},
	})

	inner := &fakeExecutor{}
	budget := agentrt.NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   50,
		DelegationLimit: 15,
		AlertThreshold:  0.8,
	})
	coordExec := agentrt.NewCoordinatingExecutor(inner, nil, budget, nil, nil)
	wrapped := wrapWithBudgetRestore(coordExec, store)

	// First call — restores baseline.
	_, _ = wrapped.RunStreamingDetailed(context.Background(), "sess-2", "a", nil)

	bre := wrapped.(*budgetRestoringExecutor)
	stateVal, _ := bre.sessionState.Load("sess-2")
	state := stateVal.(*sessionBudgetState)
	state.mu.Lock()
	assert.Equal(t, int64(5), state.cumulativeTurns)
	state.mu.Unlock()

	// Second call — should NOT re-restore (baseline stays, only adds run delta).
	_, _ = wrapped.RunStreamingDetailed(context.Background(), "sess-2", "b", nil)
	state.mu.Lock()
	// Still 5 + 0 (run delta) + 0 (second run delta) = 5
	assert.Equal(t, int64(5), state.cumulativeTurns)
	state.mu.Unlock()
	assert.Equal(t, 2, inner.calls)
}

func TestBudgetRestoringExecutor_NoSessionMetadata(t *testing.T) {
	store := newSessionUsageStore()
	// No session in store — Get returns nil.

	inner := &fakeExecutor{}
	budget := agentrt.NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   50,
		DelegationLimit: 15,
		AlertThreshold:  0.8,
	})
	coordExec := agentrt.NewCoordinatingExecutor(inner, nil, budget, nil, nil)
	wrapped := wrapWithBudgetRestore(coordExec, store)

	_, err := wrapped.RunStreamingDetailed(context.Background(), "missing", "x", nil)
	require.NoError(t, err)

	// Session state should exist but be zero.
	bre := wrapped.(*budgetRestoringExecutor)
	stateVal, ok := bre.sessionState.Load("missing")
	require.True(t, ok)
	state := stateVal.(*sessionBudgetState)
	state.mu.Lock()
	assert.Equal(t, int64(0), state.cumulativeTurns)
	assert.Equal(t, int64(0), state.cumulativeDeleg)
	state.mu.Unlock()
}

func TestBudgetRestoringExecutor_ClassicModeGracefulSkip(t *testing.T) {
	store := newSessionUsageStore()
	store.put(&session.Session{
		Key:      "sess-classic",
		Metadata: map[string]string{},
	})

	// Classic mode: inner is NOT a CoordinatingExecutor.
	inner := &fakeExecutor{}
	wrapped := wrapWithBudgetRestore(inner, store)

	_, err := wrapped.RunStreamingDetailed(context.Background(), "sess-classic", "x", nil)
	require.NoError(t, err)

	// No session state should be created (no CoordinatingExecutor to read stats from).
	bre := wrapped.(*budgetRestoringExecutor)
	_, ok := bre.sessionState.Load("sess-classic")
	assert.False(t, ok, "classic mode should not create session budget state")
}

func TestBudgetRestoringExecutor_NoSessionCrossContamination(t *testing.T) {
	store := newSessionUsageStore()
	store.put(&session.Session{
		Key: "sess-a",
		Metadata: map[string]string{
			"usage:budget_turns":       "10",
			"usage:budget_delegations": "4",
		},
	})
	store.put(&session.Session{
		Key: "sess-b",
		Metadata: map[string]string{
			"usage:budget_turns":       "20",
			"usage:budget_delegations": "8",
		},
	})

	inner := &fakeExecutor{}
	budget := agentrt.NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   50,
		DelegationLimit: 15,
		AlertThreshold:  0.8,
	})
	coordExec := agentrt.NewCoordinatingExecutor(inner, nil, budget, nil, nil)
	wrapped := wrapWithBudgetRestore(coordExec, store)

	_, _ = wrapped.RunStreamingDetailed(context.Background(), "sess-a", "a", nil)
	_, _ = wrapped.RunStreamingDetailed(context.Background(), "sess-b", "b", nil)

	bre := wrapped.(*budgetRestoringExecutor)

	stateA, _ := bre.sessionState.Load("sess-a")
	sA := stateA.(*sessionBudgetState)
	sA.mu.Lock()
	assert.Equal(t, int64(10), sA.cumulativeTurns)
	assert.Equal(t, int64(4), sA.cumulativeDeleg)
	sA.mu.Unlock()

	stateB, _ := bre.sessionState.Load("sess-b")
	sB := stateB.(*sessionBudgetState)
	sB.mu.Lock()
	assert.Equal(t, int64(20), sB.cumulativeTurns)
	assert.Equal(t, int64(8), sB.cumulativeDeleg)
	sB.mu.Unlock()
}

// --- wireSessionUsage tests ---

func TestWireSessionUsage_PersistsSessionLocalBudgetAndTokens(t *testing.T) {
	store := newSessionUsageStore()
	store.put(&session.Session{Key: "sess-3"})

	collector := observability.NewCollector()
	collector.RecordTokenUsage(observability.TokenUsage{
		SessionKey:   "sess-3",
		InputTokens:  1000,
		OutputTokens: 500,
	})

	// Create a budgetRestoringExecutor with pre-seeded session state.
	inner := &fakeExecutor{}
	bre := &budgetRestoringExecutor{
		inner: inner,
		store: store,
	}
	// Seed session state to simulate a run that accumulated counters.
	bre.sessionState.Store("sess-3", &sessionBudgetState{
		cumulativeTurns: 2,
		cumulativeDeleg: 1,
	})

	runner := turnrunner.New(turnrunner.Config{}, inner, store, nil)
	wireSessionUsage(runner, bre, store, collector)

	_, _ = runner.Run(context.Background(), turnrunner.Request{
		SessionKey: "sess-3",
		Input:      "test",
	})

	sess, err := store.Get("sess-3")
	require.NoError(t, err)
	require.NotNil(t, sess)
	assert.Equal(t, "2", sess.Metadata["usage:budget_turns"])
	assert.Equal(t, "1", sess.Metadata["usage:budget_delegations"])
	assert.Equal(t, "1000", sess.Metadata["usage:cumulative_input_tokens"])
	assert.Equal(t, "500", sess.Metadata["usage:cumulative_output_tokens"])
}

func TestWireSessionUsage_NilBudgetExecNoOp(t *testing.T) {
	store := newSessionUsageStore()
	store.put(&session.Session{Key: "sess-nil"})
	inner := &fakeExecutor{}
	runner := turnrunner.New(turnrunner.Config{}, inner, store, nil)

	// Should not panic with nil budgetExec.
	wireSessionUsage(runner, nil, store, nil)

	_, _ = runner.Run(context.Background(), turnrunner.Request{
		SessionKey: "sess-nil",
		Input:      "test",
	})

	// Session should exist but have no budget keys.
	sess, err := store.Get("sess-nil")
	require.NoError(t, err)
	if sess != nil && sess.Metadata != nil {
		_, hasTurns := sess.Metadata["usage:budget_turns"]
		assert.False(t, hasTurns, "no budget keys should be written with nil budgetExec")
	}
}
