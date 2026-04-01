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

func TestBudgetRestoringExecutor_RestoresOnFirstCall(t *testing.T) {
	store := newSessionUsageStore()
	store.put(&session.Session{
		Key: "sess-1",
		Metadata: map[string]string{
			"usage:budget_turns":       "7",
			"usage:budget_delegations": "3",
		},
	})

	budget := agentrt.NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   50,
		DelegationLimit: 15,
		AlertThreshold:  0.8,
	})
	inner := &fakeExecutor{}
	wrapped := wrapWithBudgetRestore(inner, budget, store)

	report, err := wrapped.RunStreamingDetailed(
		context.Background(), "sess-1", "hello", nil,
	)
	require.NoError(t, err)
	assert.Equal(t, "ok", report.Response)
	assert.Equal(t, 7, budget.TurnCount())
	assert.Equal(t, 3, budget.DelegationCount())
	assert.Equal(t, 1, inner.calls)
}

func TestBudgetRestoringExecutor_SkipsSubsequentCalls(t *testing.T) {
	store := newSessionUsageStore()
	store.put(&session.Session{
		Key: "sess-2",
		Metadata: map[string]string{
			"usage:budget_turns":       "5",
			"usage:budget_delegations": "2",
		},
	})

	budget := agentrt.NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   50,
		DelegationLimit: 15,
		AlertThreshold:  0.8,
	})
	inner := &fakeExecutor{}
	wrapped := wrapWithBudgetRestore(inner, budget, store)

	// First call — restores.
	_, _ = wrapped.RunStreamingDetailed(context.Background(), "sess-2", "a", nil)
	assert.Equal(t, 5, budget.TurnCount())

	// Record additional turns.
	budget.RecordTurn()
	assert.Equal(t, 6, budget.TurnCount())

	// Second call — should NOT re-restore (would reset to 5).
	_, _ = wrapped.RunStreamingDetailed(context.Background(), "sess-2", "b", nil)
	assert.Equal(t, 6, budget.TurnCount())
	assert.Equal(t, 2, inner.calls)
}

func TestBudgetRestoringExecutor_NoSessionMetadata(t *testing.T) {
	store := newSessionUsageStore()
	// No session in store — Get returns nil.

	budget := agentrt.NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   50,
		DelegationLimit: 15,
		AlertThreshold:  0.8,
	})
	inner := &fakeExecutor{}
	wrapped := wrapWithBudgetRestore(inner, budget, store)

	_, err := wrapped.RunStreamingDetailed(context.Background(), "missing", "x", nil)
	require.NoError(t, err)
	assert.Equal(t, 0, budget.TurnCount())
	assert.Equal(t, 0, budget.DelegationCount())
}

// --- wireSessionUsage tests ---

func TestWireSessionUsage_PersistsBudgetAndTokens(t *testing.T) {
	store := newSessionUsageStore()
	store.put(&session.Session{Key: "sess-3"})

	budget := agentrt.NewBudgetPolicy(config.BudgetCfg{
		ToolCallLimit:   50,
		DelegationLimit: 15,
		AlertThreshold:  0.8,
	})
	budget.RecordTurn()
	budget.RecordTurn()
	budget.RecordDelegation("agent-a")

	collector := observability.NewCollector()
	collector.RecordTokenUsage(observability.TokenUsage{
		SessionKey:   "sess-3",
		InputTokens:  1000,
		OutputTokens: 500,
	})

	inner := &fakeExecutor{}
	runner := turnrunner.New(turnrunner.Config{}, inner, store, nil)
	wireSessionUsage(runner, budget, store, collector)

	// Simulate turn complete by invoking the callback indirectly.
	// We fire the runner's callbacks via a real Run call — but that requires
	// a full executor roundtrip. Instead, we test the callback logic by
	// manually triggering OnTurnComplete's registered callbacks.
	// The runner stores callbacks; we can exercise them via a turn.
	// For simplicity, call the callback approach directly:
	// wireSessionUsage registered a callback. We'll simulate by calling Run.
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

func TestWireSessionUsage_NilBudgetNoOp(t *testing.T) {
	store := newSessionUsageStore()
	inner := &fakeExecutor{}
	runner := turnrunner.New(turnrunner.Config{}, inner, store, nil)

	// Should not panic with nil budget.
	wireSessionUsage(runner, nil, store, nil)
}
