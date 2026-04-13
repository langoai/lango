package app

import (
	"context"
	"strconv"
	"sync"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/turnrunner"
)

// sessionBudgetState tracks cumulative budget counters for a single session.
type sessionBudgetState struct {
	mu              sync.Mutex
	cumulativeTurns int64
	cumulativeDeleg int64
}

// budgetRestoringExecutor wraps an executor and tracks session-local budget
// state. On the first call per session key, it restores cumulative counters
// from Session.Metadata. After each run, it reads LastRunStatsForSession()
// from the inner CoordinatingExecutor (if present) and adds the delta to the
// session-local cumulative counters.
//
// Scope: budget state (turns, delegations) is restored to runtime.
// Token totals (cumulative_input_tokens, cumulative_output_tokens) are
// persisted to metadata for display but NOT restored to the metrics
// collector at resume time. Token continuity is deferred to phase 2.
type budgetRestoringExecutor struct {
	inner        turnrunner.Executor
	store        session.Store
	sessionState sync.Map // sessionKey -> *sessionBudgetState
	restored     sync.Map // sessionKey -> bool
}

// wrapWithBudgetRestore creates a budgetRestoringExecutor that tracks
// session-local budget state from session metadata on first use per session.
func wrapWithBudgetRestore(
	inner turnrunner.Executor,
	store session.Store,
) turnrunner.Executor {
	return &budgetRestoringExecutor{
		inner: inner,
		store: store,
	}
}

func (e *budgetRestoringExecutor) RunStreamingDetailed(
	ctx context.Context,
	sessionID, input string,
	onChunk adk.ChunkCallback,
	opts ...adk.RunOption,
) (adk.RunReport, error) {
	if _, loaded := e.restored.LoadOrStore(sessionID, true); !loaded {
		e.restoreBaseline(sessionID)
	}

	report, err := e.inner.RunStreamingDetailed(ctx, sessionID, input, onChunk, opts...)

	// After the run, accumulate the per-run clone's counters into session state.
	// LastRunStatsForSession is keyed by sessionID, so concurrent turns for
	// different sessions never overwrite each other's stats.
	if ce, ok := e.coordinatingExecutor(); ok {
		if stats, ok := ce.LastRunStatsForSession(sessionID); ok {
			state := e.getOrCreateState(sessionID)
			state.mu.Lock()
			state.cumulativeTurns += int64(stats.Turns)
			state.cumulativeDeleg += int64(stats.Delegations)
			state.mu.Unlock()
		}
	}

	return report, err
}

// restoreBaseline reads persisted cumulative counters from session metadata
// and seeds the session-local state.
func (e *budgetRestoringExecutor) restoreBaseline(sessionID string) {
	sess, err := e.store.Get(sessionID)
	if err != nil || sess == nil || len(sess.Metadata) == 0 {
		return
	}
	state := e.getOrCreateState(sessionID)
	state.mu.Lock()
	defer state.mu.Unlock()
	if v, parseErr := strconv.ParseInt(sess.Metadata["usage:budget_turns"], 10, 64); parseErr == nil {
		state.cumulativeTurns = v
	}
	if v, parseErr := strconv.ParseInt(sess.Metadata["usage:budget_delegations"], 10, 64); parseErr == nil {
		state.cumulativeDeleg = v
	}
}

// getOrCreateState returns the sessionBudgetState for a session key, creating
// one if it does not exist.
func (e *budgetRestoringExecutor) getOrCreateState(sessionID string) *sessionBudgetState {
	if v, ok := e.sessionState.Load(sessionID); ok {
		return v.(*sessionBudgetState)
	}
	state := &sessionBudgetState{}
	actual, _ := e.sessionState.LoadOrStore(sessionID, state)
	return actual.(*sessionBudgetState)
}

// coordinatingExecutor unwraps the inner executor to find a
// *agentrt.CoordinatingExecutor, if present.
func (e *budgetRestoringExecutor) coordinatingExecutor() (*agentrt.CoordinatingExecutor, bool) {
	ce, ok := e.inner.(*agentrt.CoordinatingExecutor)
	return ce, ok
}

// wireSessionUsage registers an OnTurnComplete callback that persists
// session-local budget counters and cumulative token usage into
// Session.Metadata after each turn.
func wireSessionUsage(
	runner *turnrunner.Runner,
	budgetExec *budgetRestoringExecutor,
	store session.Store,
	collector *observability.MetricsCollector,
) {
	if runner == nil || store == nil {
		return
	}
	runner.OnTurnComplete(func(sessionKey string) {
		sess, err := store.Get(sessionKey)
		if err != nil || sess == nil {
			return
		}
		if sess.Metadata == nil {
			sess.Metadata = make(map[string]string)
		}

		// Persist session-local budget counters.
		if budgetExec != nil {
			if state, ok := budgetExec.sessionState.Load(sessionKey); ok {
				s := state.(*sessionBudgetState)
				s.mu.Lock()
				sess.Metadata["usage:budget_turns"] = strconv.FormatInt(s.cumulativeTurns, 10)
				sess.Metadata["usage:budget_delegations"] = strconv.FormatInt(s.cumulativeDeleg, 10)
				s.mu.Unlock()
			}
		}

		// Persist cumulative token usage from metrics collector.
		if collector != nil {
			if sm := collector.SessionMetrics(sessionKey); sm != nil {
				sess.Metadata["usage:cumulative_input_tokens"] = strconv.FormatInt(sm.InputTokens, 10)
				sess.Metadata["usage:cumulative_output_tokens"] = strconv.FormatInt(sm.OutputTokens, 10)
			}
		}

		if err := store.Update(sess); err != nil {
			logger().Warnw("persist session usage metadata", "session", sessionKey, "error", err)
		}
	})
}
