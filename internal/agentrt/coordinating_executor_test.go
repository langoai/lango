package agentrt

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adksession "google.golang.org/adk/session"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
)

// mockExecutor implements turnrunner.Executor for testing.
type mockExecutor struct {
	report adk.RunReport
	err    error
	calls  int
}

func (m *mockExecutor) RunStreamingDetailed(
	_ context.Context,
	_, _ string,
	onChunk adk.ChunkCallback,
	opts ...adk.RunOption,
) (adk.RunReport, error) {
	m.calls++

	// Apply event hooks (simulate ADK behavior).
	var ro struct{ onEvent func(*adksession.Event) }
	for _, opt := range opts {
		// RunOption is func(*runOptions), which we can't inspect directly.
		// For integration tests, we verify the wrapper behavior instead.
		_ = opt
	}
	_ = ro

	if onChunk != nil && m.report.Response != "" {
		onChunk(m.report.Response)
	}
	return m.report, m.err
}

func TestCoordinatingExecutor_PassthroughOnSuccess(t *testing.T) {
	inner := &mockExecutor{
		report: adk.RunReport{Response: "hello world"},
	}

	ce := NewCoordinatingExecutor(
		inner,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil),
		nil,
	)

	report, err := ce.RunStreamingDetailed(context.Background(), "sess-1", "input", nil)
	require.NoError(t, err)
	assert.Equal(t, "hello world", report.Response)
	assert.Equal(t, 1, inner.calls)
}

func TestCoordinatingExecutor_RecoveryRetryOnTransient(t *testing.T) {
	callCount := 0
	inner := &mockExecutor{}

	// First call fails with transient error, second succeeds.
	origRunStreamingDetailed := inner.RunStreamingDetailed
	_ = origRunStreamingDetailed

	// Use a custom mock that changes behavior on second call.
	type multiCallExecutor struct {
		reports []adk.RunReport
		errs    []error
	}
	mce := &multiCallExecutor{
		reports: []adk.RunReport{
			{Response: ""},
			{Response: "recovered"},
		},
		errs: []error{
			&adk.AgentError{Code: adk.ErrModelError, CauseClass: "provider_transient", Message: "503"},
			nil,
		},
	}

	var executor mockCallExecutor = func(ctx context.Context, sid, input string, onChunk adk.ChunkCallback, opts ...adk.RunOption) (adk.RunReport, error) {
		idx := callCount
		callCount++
		if idx >= len(mce.reports) {
			idx = len(mce.reports) - 1
		}
		return mce.reports[idx], mce.errs[idx]
	}

	ce := NewCoordinatingExecutor(
		executor,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil),
		nil,
	)

	report, err := ce.RunStreamingDetailed(context.Background(), "sess-1", "input", nil)
	require.NoError(t, err)
	assert.Equal(t, "recovered", report.Response)
	assert.Equal(t, 2, callCount)
}

func TestCoordinatingExecutor_RecoveryEscalateOnTimeout(t *testing.T) {
	timeoutErr := &adk.AgentError{Code: adk.ErrTimeout, Message: "deadline exceeded"}
	executor := mockCallExecutor(func(_ context.Context, _, _ string, _ adk.ChunkCallback, _ ...adk.RunOption) (adk.RunReport, error) {
		return adk.RunReport{}, timeoutErr
	})

	ce := NewCoordinatingExecutor(
		executor,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil),
		nil,
	)

	_, err := ce.RunStreamingDetailed(context.Background(), "sess-1", "input", nil)
	require.Error(t, err)
	var agentErr *adk.AgentError
	require.True(t, errors.As(err, &agentErr))
	assert.Equal(t, adk.ErrTimeout, agentErr.Code)
}

func TestCoordinatingExecutor_EventBusPublish(t *testing.T) {
	bus := eventbus.New()
	var recoveryEvents []RecoveryEvent
	eventbus.SubscribeTyped(bus, func(e RecoveryEvent) {
		recoveryEvents = append(recoveryEvents, e)
	})

	toolChurnErr := &adk.AgentError{Code: adk.ErrToolChurn, Message: "stuck"}
	callCount := 0
	executor := mockCallExecutor(func(_ context.Context, _, _ string, _ adk.ChunkCallback, _ ...adk.RunOption) (adk.RunReport, error) {
		callCount++
		if callCount <= 2 {
			return adk.RunReport{}, toolChurnErr
		}
		return adk.RunReport{Response: "ok"}, nil
	})

	ce := NewCoordinatingExecutor(
		executor,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 3}, nil),
		bus,
	)

	report, err := ce.RunStreamingDetailed(context.Background(), "sess-1", "input", nil)
	require.NoError(t, err)
	assert.Equal(t, "ok", report.Response)
	assert.Equal(t, 3, callCount)
	assert.Len(t, recoveryEvents, 2) // two recovery events published
}

func TestCoordinatingExecutor_ToolErrorAfterSpecialistUsesRerouteHint(t *testing.T) {
	bus := eventbus.New()
	var recoveryEvents []RecoveryEvent
	eventbus.SubscribeTyped(bus, func(e RecoveryEvent) {
		recoveryEvents = append(recoveryEvents, e)
	})

	var (
		inputs     []string
		recoveries []adk.RecoveryInfo
		callCount  int
	)
	executor := mockCallExecutor(func(_ context.Context, _, input string, _ adk.ChunkCallback, opts ...adk.RunOption) (adk.RunReport, error) {
		inputs = append(inputs, input)
		hooks := adk.ResolveRunHooks(opts...)
		if callCount == 0 {
			callCount++
			require.NotNil(t, hooks.OnEvent)
			hooks.OnEvent(&adksession.Event{
				Author: "lango-orchestrator",
				Actions: adksession.EventActions{
					TransferToAgent: "vault",
				},
			})
			return adk.RunReport{}, &adk.AgentError{Code: adk.ErrToolError, CauseClass: "unknown_tool_error", Message: "tool failed"}
		}
		callCount++
		return adk.RunReport{Response: "rerouted"}, nil
	})

	ce := NewCoordinatingExecutor(
		executor,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil),
		bus,
	)

	report, err := ce.RunStreamingDetailed(
		context.Background(),
		"sess-1",
		"check wallet balance",
		nil,
		adk.WithOnRecovery(func(info adk.RecoveryInfo) {
			recoveries = append(recoveries, info)
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, "rerouted", report.Response)
	require.Len(t, inputs, 2)
	assert.NotEqual(t, inputs[0], inputs[1], "second attempt should use reroute hint input")
	assert.Contains(t, inputs[1], "vault")
	require.Len(t, recoveryEvents, 1)
	assert.Equal(t, "vault", recoveryEvents[0].AgentName)
	assert.Equal(t, RecoveryRetryWithHint, recoveryEvents[0].Action)
	require.Len(t, recoveries, 1)
	assert.Equal(t, "vault", recoveries[0].AgentName)
	assert.Equal(t, "retry_with_hint", recoveries[0].Action)
}

func TestCoordinatingExecutor_ToolErrorBeforeSpecialistRetriesSameInput(t *testing.T) {
	var (
		inputs    []string
		callCount int
	)
	executor := mockCallExecutor(func(_ context.Context, _, input string, _ adk.ChunkCallback, _ ...adk.RunOption) (adk.RunReport, error) {
		inputs = append(inputs, input)
		if callCount == 0 {
			callCount++
			return adk.RunReport{}, &adk.AgentError{Code: adk.ErrToolError, CauseClass: "unknown_tool_error", Message: "tool failed"}
		}
		callCount++
		return adk.RunReport{Response: "retried"}, nil
	})

	ce := NewCoordinatingExecutor(
		executor,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil),
		nil,
	)

	report, err := ce.RunStreamingDetailed(context.Background(), "sess-1", "check wallet balance", nil)
	require.NoError(t, err)
	assert.Equal(t, "retried", report.Response)
	require.Len(t, inputs, 2)
	assert.Equal(t, inputs[0], inputs[1], "pre-specialist retries should keep same input")
}

func TestCoordinatingExecutor_PreservesExistingOnEventHook(t *testing.T) {
	var existingHookCalls int32

	executor := mockCallExecutor(func(_ context.Context, _, _ string, _ adk.ChunkCallback, opts ...adk.RunOption) (adk.RunReport, error) {
		hooks := adk.ResolveRunHooks(opts...)
		require.NotNil(t, hooks.OnEvent)
		hooks.OnEvent(&adksession.Event{
			Author: "lango-orchestrator",
			Actions: adksession.EventActions{
				TransferToAgent: "operator",
			},
		})
		return adk.RunReport{Response: "ok"}, nil
	})

	ce := NewCoordinatingExecutor(
		executor,
		NewDelegationGuard(config.CircuitBreakerCfg{FailureThreshold: 3}, nil),
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		NewRecoveryPolicy(config.RecoveryCfg{MaxRetries: 2}, nil),
		nil,
	)

	_, err := ce.RunStreamingDetailed(
		context.Background(),
		"sess-1",
		"input",
		nil,
		adk.WithOnEvent(func(_ *adksession.Event) {
			atomic.AddInt32(&existingHookCalls, 1)
		}),
	)
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&existingHookCalls))
}

func TestCoordinatingExecutor_RecordOutcomeUsesSpecialistNotReturnTarget(t *testing.T) {
	executor := mockCallExecutor(func(_ context.Context, _, _ string, _ adk.ChunkCallback, opts ...adk.RunOption) (adk.RunReport, error) {
		hooks := adk.ResolveRunHooks(opts...)
		require.NotNil(t, hooks.OnEvent)

		hooks.OnEvent(&adksession.Event{
			Author: "lango-orchestrator",
			Actions: adksession.EventActions{
				TransferToAgent: "operator",
			},
		})
		hooks.OnEvent(&adksession.Event{
			Author: "operator",
			Actions: adksession.EventActions{
				TransferToAgent: "lango-orchestrator",
			},
		})

		return adk.RunReport{}, &adk.AgentError{Code: adk.ErrToolChurn, Message: "loop"}
	})

	guard := NewDelegationGuard(config.CircuitBreakerCfg{
		FailureThreshold: 1,
		ResetTimeout:     30 * time.Second,
	}, nil)

	ce := NewCoordinatingExecutor(
		executor,
		guard,
		NewBudgetPolicy(config.BudgetCfg{ToolCallLimit: 50, DelegationLimit: 15, AlertThreshold: 0.8}),
		nil,
		nil,
	)

	_, err := ce.RunStreamingDetailed(context.Background(), "sess-1", "input", nil)
	require.Error(t, err)
	assert.Equal(t, CircuitOpen, guard.State("operator"))
	assert.Equal(t, CircuitClosed, guard.State("lango-orchestrator"))
}

// mockCallExecutor is a function-based turnrunner.Executor for testing.
type mockCallExecutor func(ctx context.Context, sessionID, input string, onChunk adk.ChunkCallback, opts ...adk.RunOption) (adk.RunReport, error)

func (f mockCallExecutor) RunStreamingDetailed(ctx context.Context, sessionID, input string, onChunk adk.ChunkCallback, opts ...adk.RunOption) (adk.RunReport, error) {
	return f(ctx, sessionID, input, onChunk, opts...)
}
