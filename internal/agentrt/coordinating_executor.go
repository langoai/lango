package agentrt

import (
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"
	adksession "google.golang.org/adk/session"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/turnrunner"
)

func logger() *zap.SugaredLogger { return logging.SubsystemSugar("agentrt") }

// Compile-time interface check.
var _ turnrunner.Executor = (*CoordinatingExecutor)(nil)

// CoordinatingExecutor wraps a turnrunner.Executor to apply
// DelegationGuard, BudgetPolicy, and RecoveryPolicy.
// It is a policy/observation wrapper, not a new execution engine.
// Routing authority remains with the root orchestrator LLM.
type CoordinatingExecutor struct {
	inner    turnrunner.Executor
	guard    *DelegationGuard
	budget   *BudgetPolicy
	recovery *RecoveryPolicy
	bus      *eventbus.Bus
}

// NewCoordinatingExecutor creates a coordinating executor wrapping the inner executor.
func NewCoordinatingExecutor(
	inner turnrunner.Executor,
	guard *DelegationGuard,
	budget *BudgetPolicy,
	recovery *RecoveryPolicy,
	bus *eventbus.Bus,
) *CoordinatingExecutor {
	return &CoordinatingExecutor{
		inner:    inner,
		guard:    guard,
		budget:   budget,
		recovery: recovery,
		bus:      bus,
	}
}

// runState holds per-invocation mutable state. Created fresh for each
// RunStreamingDetailed call so concurrent turns never share state.
type runState struct {
	mu                   sync.Mutex
	lastDelegationTarget string
}

func (s *runState) setTarget(target string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastDelegationTarget = target
}

func (s *runState) target() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastDelegationTarget
}

// RunStreamingDetailed implements turnrunner.Executor.
func (c *CoordinatingExecutor) RunStreamingDetailed(
	ctx context.Context,
	sessionID, input string,
	onChunk adk.ChunkCallback,
	opts ...adk.RunOption,
) (adk.RunReport, error) {
	state := &runState{}       // per-run, not shared across concurrent turns
	budget := c.budget.Clone() // per-run mirrored counters

	return c.runWithRecovery(ctx, sessionID, input, onChunk, state, budget, 0, opts...)
}

func (c *CoordinatingExecutor) runWithRecovery(
	ctx context.Context,
	sessionID, input string,
	onChunk adk.ChunkCallback,
	state *runState,
	budget *BudgetPolicy,
	retryCount int,
	opts ...adk.RunOption,
) (adk.RunReport, error) {
	// ChainOnEvent preserves any existing onEvent handler (e.g., TurnRunner's
	// trace recorder) and appends our policy hook after it. WithOnEvent is a
	// simple setter — last value wins — so ChainOnEvent is required here.
	policyHook := adk.ChainOnEvent(func(event *adksession.Event) {
		c.observeEvent(event, sessionID, state, budget)
	})
	mergedOpts := append(append([]adk.RunOption{}, opts...), policyHook)
	hooks := adk.ResolveRunHooks(mergedOpts...)

	// Clear target before each attempt so retries that succeed without
	// a fresh delegation don't misattribute to the previous target.
	state.setTarget("")

	report, err := c.inner.RunStreamingDetailed(ctx, sessionID, input, onChunk, mergedOpts...)

	// Record circuit breaker outcome for the delegation target observed
	// during THIS attempt (not a stale value from a previous attempt).
	if t := state.target(); t != "" {
		c.guard.RecordOutcome(t, err == nil)
	}

	if err == nil {
		return report, nil
	}

	if c.recovery == nil {
		return report, err
	}

	failure := RecoveryContext{
		Error:         err,
		AgentName:     state.target(),
		PartialResult: report.Response,
		RetryCount:    retryCount,
		SessionID:     sessionID,
	}
	action := c.recovery.Decide(ctx, &failure)

	// Diagnostic: log error classification for root-cause analysis.
	var agentErr *adk.AgentError
	if errors.As(err, &agentErr) {
		logger().Warnw("recovery triggered",
			"session", sessionID,
			"agent", failure.AgentName,
			"action", action.String(),
			"error_code", agentErr.Code,
			"cause_class", agentErr.CauseClass,
			"retry", retryCount)
	} else {
		logger().Warnw("recovery triggered",
			"session", sessionID,
			"agent", failure.AgentName,
			"action", action.String(),
			"error", err.Error(),
			"retry", retryCount)
	}

	if hooks.OnRecovery != nil {
		hooks.OnRecovery(adk.RecoveryInfo{
			Action:    action.String(),
			AgentName: failure.AgentName,
			Error:     err.Error(),
		})
	}

	if c.bus != nil {
		c.bus.Publish(RecoveryEvent{
			Action:    action,
			AgentName: failure.AgentName,
			Error:     err.Error(),
			SessionID: sessionID,
		})
	}

	switch action {
	case RecoveryRetry:
		logger().Infow("recovery: retry same input",
			"session", sessionID,
			"retry", retryCount+1)
		return c.runWithRecovery(ctx, sessionID, input, onChunk, state, budget, retryCount+1, opts...)

	case RecoveryRetryWithHint:
		hintedInput := AddRerouteHint(input, failure)
		logger().Infow("recovery: retry with reroute hint",
			"session", sessionID,
			"retry", retryCount+1)
		return c.runWithRecovery(ctx, sessionID, hintedInput, onChunk, state, budget, retryCount+1, opts...)

	case RecoveryDirectAnswer:
		logger().Infow("recovery: using partial result",
			"session", sessionID,
			"partial_len", len(report.Response))
		return report, nil

	default: // RecoveryEscalate, RecoveryNone
		return report, err
	}
}

// observeEvent is the policy hook called for each ADK event.
func (c *CoordinatingExecutor) observeEvent(event *adksession.Event, sessionID string, state *runState, budget *BudgetPolicy) {
	if event == nil {
		return
	}

	// Delegation tracking.
	if event.Actions.TransferToAgent != "" {
		target := event.Actions.TransferToAgent
		if target != "" && target != "lango-orchestrator" {
			state.setTarget(target)
		}
		isOpen := c.guard.IsOpen(target)

		budget.RecordDelegation(target)

		if c.bus != nil {
			c.bus.Publish(DelegationObservedEvent{
				From:      event.Author,
				To:        target,
				IsOpen:    isOpen,
				SessionID: sessionID,
			})
		}

		if isOpen {
			logger().Warnw("delegation to circuit-open agent",
				"from", event.Author,
				"to", target,
				"session", sessionID)
		}
		return
	}

	// Turn counting: only function-call events that are not delegations
	// (matches inner budget semantics from agent.go:350).
	if event.Content != nil && hasFunctionCallParts(event) {
		budget.RecordTurn()
	}
}

// hasFunctionCallParts checks if an ADK event contains function call parts.
func hasFunctionCallParts(event *adksession.Event) bool {
	if event.Content == nil {
		return false
	}
	for _, part := range event.Content.Parts {
		if part.FunctionCall != nil {
			return true
		}
	}
	return false
}
