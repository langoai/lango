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

// budgetRestoringExecutor wraps an executor and lazily restores budget state
// from Session.Metadata on the first call per session key.
type budgetRestoringExecutor struct {
	inner    turnrunner.Executor
	budget   *agentrt.BudgetPolicy
	store    session.Store
	restored sync.Map // sessionKey → bool
}

// wrapWithBudgetRestore creates a budgetRestoringExecutor that restores budget
// state from session metadata on first use per session.
func wrapWithBudgetRestore(
	inner turnrunner.Executor,
	budget *agentrt.BudgetPolicy,
	store session.Store,
) turnrunner.Executor {
	return &budgetRestoringExecutor{
		inner:  inner,
		budget: budget,
		store:  store,
	}
}

func (e *budgetRestoringExecutor) RunStreamingDetailed(
	ctx context.Context,
	sessionID, input string,
	onChunk adk.ChunkCallback,
	opts ...adk.RunOption,
) (adk.RunReport, error) {
	if _, loaded := e.restored.LoadOrStore(sessionID, true); !loaded {
		e.restoreBudget(sessionID)
	}
	return e.inner.RunStreamingDetailed(ctx, sessionID, input, onChunk, opts...)
}

func (e *budgetRestoringExecutor) restoreBudget(sessionID string) {
	sess, err := e.store.Get(sessionID)
	if err != nil || sess == nil || len(sess.Metadata) == 0 {
		return
	}
	e.budget.Restore(sess.Metadata)
}

// wireSessionUsage registers an OnTurnComplete callback that persists budget
// counters and cumulative token usage into Session.Metadata after each turn.
func wireSessionUsage(
	runner *turnrunner.Runner,
	budget *agentrt.BudgetPolicy,
	store session.Store,
	collector *observability.MetricsCollector,
) {
	if runner == nil || budget == nil || store == nil {
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

		// Persist budget counters.
		for k, v := range budget.Serialize() {
			sess.Metadata[k] = v
		}

		// Persist cumulative token usage from metrics collector.
		if collector != nil {
			snap := collector.Snapshot()
			if sm, ok := snap.SessionBreakdown[sessionKey]; ok {
				sess.Metadata["usage:cumulative_input_tokens"] = strconv.FormatInt(sm.InputTokens, 10)
				sess.Metadata["usage:cumulative_output_tokens"] = strconv.FormatInt(sm.OutputTokens, 10)
			}
		}

		if err := store.Update(sess); err != nil {
			logger().Warnw("persist session usage metadata", "session", sessionKey, "error", err)
		}
	})
}
