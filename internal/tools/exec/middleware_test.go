package exec

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBus struct {
	published atomic.Int32
	lastEvent PolicyDecisionData
}

func (m *mockBus) Publish(event PolicyEvent) {
	m.published.Add(1)
	if e, ok := event.(PolicyDecisionData); ok {
		m.lastEvent = e
	}
}

func newTestPolicyEvaluatorWithBus(t *testing.T, bus EventPublisher) *PolicyEvaluator {
	t.Helper()
	guard := NewCommandGuard([]string{"~/.lango"})
	classifier := func(cmd string) (string, ReasonCode) {
		lower := strings.ToLower(strings.TrimSpace(cmd))
		if strings.HasPrefix(lower, "lango ") || lower == "lango" {
			return "blocked: lango CLI", ReasonLangoCLI
		}
		return "", ReasonNone
	}
	return NewPolicyEvaluator(guard, classifier, bus)
}

func TestWithPolicy_BlockPreventsNext(t *testing.T) {
	t.Parallel()

	pe := newTestPolicyEvaluatorWithBus(t, nil)
	mw := WithPolicy(pe)

	var nextCalled atomic.Int32
	execTool := &agent.Tool{Name: "exec"}
	next := func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		nextCalled.Add(1)
		return "executed", nil
	}

	handler := mw(execTool, next)
	result, err := handler(context.Background(), map[string]interface{}{"command": "kill 1234"})

	require.NoError(t, err)
	br, ok := result.(*BlockedResult)
	require.True(t, ok, "expected BlockedResult")
	assert.True(t, br.Blocked)
	assert.Contains(t, br.Message, "process management")
	assert.Equal(t, int32(0), nextCalled.Load(), "next should not be called for blocked commands")
}

func TestWithPolicy_BlockExecBg(t *testing.T) {
	t.Parallel()

	pe := newTestPolicyEvaluatorWithBus(t, nil)
	mw := WithPolicy(pe)

	var nextCalled atomic.Int32
	execBgTool := &agent.Tool{Name: "exec_bg"}
	next := func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		nextCalled.Add(1)
		return "started", nil
	}

	handler := mw(execBgTool, next)
	result, err := handler(context.Background(), map[string]interface{}{"command": `bash -c "lango security"`})

	require.NoError(t, err)
	br, ok := result.(*BlockedResult)
	require.True(t, ok, "expected BlockedResult")
	assert.True(t, br.Blocked)
	assert.Equal(t, int32(0), nextCalled.Load())
}

func TestWithPolicy_AllowCallsNext(t *testing.T) {
	t.Parallel()

	pe := newTestPolicyEvaluatorWithBus(t, nil)
	mw := WithPolicy(pe)

	var nextCalled atomic.Int32
	execTool := &agent.Tool{Name: "exec"}
	next := func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		nextCalled.Add(1)
		return "executed", nil
	}

	handler := mw(execTool, next)
	result, err := handler(context.Background(), map[string]interface{}{"command": "go build ./..."})

	require.NoError(t, err)
	assert.Equal(t, "executed", result)
	assert.Equal(t, int32(1), nextCalled.Load())
}

func TestWithPolicy_ObserveCallsNext(t *testing.T) {
	t.Parallel()

	bus := &mockBus{}
	pe := newTestPolicyEvaluatorWithBus(t, bus)
	mw := WithPolicy(pe)

	var nextCalled atomic.Int32
	execTool := &agent.Tool{Name: "exec"}
	next := func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		nextCalled.Add(1)
		return "executed", nil
	}

	handler := mw(execTool, next)
	result, err := handler(context.Background(), map[string]interface{}{"command": "echo $(whoami)"})

	require.NoError(t, err)
	assert.Equal(t, "executed", result)
	assert.Equal(t, int32(1), nextCalled.Load(), "next should be called for observe verdict")
	assert.Greater(t, bus.published.Load(), int32(0), "event should be published for observe")
}

func TestWithPolicy_PassthroughNonExecTools(t *testing.T) {
	t.Parallel()

	pe := newTestPolicyEvaluatorWithBus(t, nil)
	mw := WithPolicy(pe)

	for _, toolName := range []string{"exec_status", "exec_stop", "fs_read", "browser_navigate"} {
		t.Run(toolName, func(t *testing.T) {
			t.Parallel()
			var nextCalled atomic.Int32
			tool := &agent.Tool{Name: toolName}
			next := func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				nextCalled.Add(1)
				return "ok", nil
			}

			handler := mw(tool, next)
			_, err := handler(context.Background(), map[string]interface{}{"command": "kill 1234"})
			require.NoError(t, err)
			assert.Equal(t, int32(1), nextCalled.Load(), "non-exec tools should always pass through")
		})
	}
}

func TestWithPolicy_NilBusNoEvent(t *testing.T) {
	t.Parallel()

	pe := newTestPolicyEvaluatorWithBus(t, nil) // nil bus
	mw := WithPolicy(pe)

	execTool := &agent.Tool{Name: "exec"}
	next := func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		return "ok", nil
	}

	handler := mw(execTool, next)
	// This should not panic even with nil bus.
	_, err := handler(context.Background(), map[string]interface{}{"command": "kill 1234"})
	require.NoError(t, err)
}

func TestWithPolicy_BlockPublishesEvent(t *testing.T) {
	t.Parallel()

	bus := &mockBus{}
	pe := newTestPolicyEvaluatorWithBus(t, bus)
	mw := WithPolicy(pe)

	execTool := &agent.Tool{Name: "exec"}
	next := func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		return "executed", nil
	}

	handler := mw(execTool, next)
	_, err := handler(context.Background(), map[string]interface{}{"command": "kill 1234"})

	require.NoError(t, err)
	assert.Greater(t, bus.published.Load(), int32(0), "event should be published for block")
	assert.Equal(t, "block", bus.lastEvent.Verdict)
	assert.Equal(t, "kill_verb", bus.lastEvent.Reason)
}

func TestWithPolicy_EmptyCommandPassesThrough(t *testing.T) {
	t.Parallel()

	pe := newTestPolicyEvaluatorWithBus(t, nil)
	mw := WithPolicy(pe)

	var nextCalled atomic.Int32
	execTool := &agent.Tool{Name: "exec"}
	next := func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		nextCalled.Add(1)
		return "ok", nil
	}

	handler := mw(execTool, next)
	_, err := handler(context.Background(), map[string]interface{}{"command": ""})
	require.NoError(t, err)
	assert.Equal(t, int32(1), nextCalled.Load(), "empty command should pass through")
}
