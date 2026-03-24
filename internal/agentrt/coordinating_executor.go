package agentrt

import (
	"context"

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

// RunStreamingDetailed implements turnrunner.Executor.
func (c *CoordinatingExecutor) RunStreamingDetailed(
	ctx context.Context,
	sessionID, input string,
	onChunk adk.ChunkCallback,
	opts ...adk.RunOption,
) (adk.RunReport, error) {
	c.budget.Reset()

	return c.runWithRecovery(ctx, sessionID, input, onChunk, 0, opts...)
}

func (c *CoordinatingExecutor) runWithRecovery(
	ctx context.Context,
	sessionID, input string,
	onChunk adk.ChunkCallback,
	retryCount int,
	opts ...adk.RunOption,
) (adk.RunReport, error) {
	// Compose a policy event hook into opts.
	// Delegation events are observed via ADK event hook (not onChunk).
	policyHook := adk.WithOnEvent(func(event *adksession.Event) {
		c.observeEvent(event, sessionID)
	})
	mergedOpts := append(append([]adk.RunOption{}, opts...), policyHook)

	report, err := c.inner.RunStreamingDetailed(ctx, sessionID, input, onChunk, mergedOpts...)
	if err == nil {
		return report, nil
	}

	if c.recovery == nil {
		return report, err
	}

	failure := RecoveryContext{
		Error:         err,
		PartialResult: report.Response,
		RetryCount:    retryCount,
		SessionID:     sessionID,
	}
	action := c.recovery.Decide(ctx, failure)

	if c.bus != nil {
		c.bus.Publish(RecoveryEvent{
			Action:    action,
			Error:     err.Error(),
			SessionID: sessionID,
		})
	}

	switch action {
	case RecoveryRetry:
		logger().Infow("recovery: retry same input",
			"session", sessionID,
			"retry", retryCount+1)
		return c.runWithRecovery(ctx, sessionID, input, onChunk, retryCount+1, opts...)

	case RecoveryRetryWithHint:
		hintedInput := AddRerouteHint(input, failure)
		logger().Infow("recovery: retry with reroute hint",
			"session", sessionID,
			"retry", retryCount+1)
		return c.runWithRecovery(ctx, sessionID, hintedInput, onChunk, retryCount+1, opts...)

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
func (c *CoordinatingExecutor) observeEvent(event *adksession.Event, sessionID string) {
	if event == nil {
		return
	}

	// Delegation tracking.
	if event.Actions.TransferToAgent != "" {
		target := event.Actions.TransferToAgent
		isOpen := c.guard.IsOpen(target)

		c.budget.RecordDelegation(target)

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
		c.budget.RecordTurn()
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
