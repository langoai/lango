package toolchain

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/session"
)

func TestWithHooks_NormalFlow(t *testing.T) {
	t.Parallel()

	preHook := &stubPreHook{
		name:     "pre",
		priority: 1,
		result:   PreHookResult{Action: Continue},
	}
	postHook := &stubPostHook{
		name:     "post",
		priority: 1,
	}

	reg := NewHookRegistry()
	reg.RegisterPre(preHook)
	reg.RegisterPost(postHook)

	var handlerCalled bool
	tool := makeTool("my_tool", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		handlerCalled = true
		return "result-value", nil
	})

	wrapped := Chain(tool, WithHooks(reg))
	result, err := wrapped.Handler(context.Background(), map[string]interface{}{"k": "v"})

	require.NoError(t, err)
	assert.True(t, handlerCalled, "handler was not called")
	assert.Equal(t, "result-value", result)
	assert.True(t, preHook.called, "pre-hook was not called")
	assert.True(t, postHook.called, "post-hook was not called")
	assert.Equal(t, "result-value", postHook.gotResult)
}

func TestWithHooks_PreHookBlocks(t *testing.T) {
	t.Parallel()

	reg := NewHookRegistry()
	reg.RegisterPre(&stubPreHook{
		name:     "blocker",
		priority: 1,
		result:   PreHookResult{Action: Block, BlockReason: "rate limit exceeded"},
	})

	var handlerCalled bool
	tool := makeTool("my_tool", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		handlerCalled = true
		return "should-not-see", nil
	})

	wrapped := Chain(tool, WithHooks(reg))
	_, err := wrapped.Handler(context.Background(), nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded")
	assert.False(t, handlerCalled, "handler should not be called when blocked")
}

func TestWithHooks_PreHookModifiesParams(t *testing.T) {
	t.Parallel()

	modifiedParams := map[string]interface{}{"key": "modified-value"}
	reg := NewHookRegistry()
	reg.RegisterPre(&stubPreHook{
		name:     "modifier",
		priority: 1,
		result:   PreHookResult{Action: Modify, ModifiedParams: modifiedParams},
	})

	var receivedParams map[string]interface{}
	tool := makeTool("my_tool", func(_ context.Context, params map[string]interface{}) (interface{}, error) {
		receivedParams = params
		return "ok", nil
	})

	wrapped := Chain(tool, WithHooks(reg))
	_, err := wrapped.Handler(context.Background(), map[string]interface{}{"key": "original"})

	require.NoError(t, err)
	assert.Equal(t, "modified-value", receivedParams["key"])
}

func TestWithHooks_PostHookErrorDoesNotAffectResult(t *testing.T) {
	t.Parallel()

	reg := NewHookRegistry()
	reg.RegisterPost(&stubPostHook{
		name:     "failing-post",
		priority: 1,
		err:      errors.New("post hook failed"),
	})

	tool := makeTool("my_tool", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		return "tool-result", nil
	})

	wrapped := Chain(tool, WithHooks(reg))
	result, err := wrapped.Handler(context.Background(), nil)

	// Post-hook errors are logged, not propagated to caller.
	require.NoError(t, err)
	assert.Equal(t, "tool-result", result)
}

func TestWithHooks_PreHookError(t *testing.T) {
	t.Parallel()

	reg := NewHookRegistry()
	reg.RegisterPre(&stubPreHook{
		name:     "err-hook",
		priority: 1,
		err:      errors.New("pre hook failed"),
	})

	var handlerCalled bool
	tool := makeTool("my_tool", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		handlerCalled = true
		return nil, nil
	})

	wrapped := Chain(tool, WithHooks(reg))
	_, err := wrapped.Handler(context.Background(), nil)

	require.Error(t, err)
	assert.False(t, handlerCalled, "handler should not be called when pre-hook errors")
}

func TestWithHooks_ContextPropagation(t *testing.T) {
	t.Parallel()

	// Verify that agent name and session key are propagated to HookContext.
	var capturedCtx HookContext
	capturingHook := &captureHookCtxPreHook{captured: &capturedCtx}

	reg := NewHookRegistry()
	reg.RegisterPre(capturingHook)

	tool := makeTool("my_tool", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		return nil, nil
	})

	ctx := context.Background()
	ctx = WithAgentName(ctx, "researcher")
	ctx = session.WithSessionKey(ctx, "session-abc")

	wrapped := Chain(tool, WithHooks(reg))
	_, err := wrapped.Handler(ctx, map[string]interface{}{"p": "v"})

	require.NoError(t, err)
	assert.Equal(t, "my_tool", capturedCtx.ToolName)
	assert.Equal(t, "researcher", capturedCtx.AgentName)
	assert.Equal(t, "session-abc", capturedCtx.SessionKey)
}

func TestWithHooks_CompatibleWithChainAll(t *testing.T) {
	t.Parallel()

	reg := NewHookRegistry()
	reg.RegisterPre(&stubPreHook{
		name:     "noop",
		priority: 1,
		result:   PreHookResult{Action: Continue},
	})

	tools := []*agent.Tool{
		makeTool("a", func(_ context.Context, _ map[string]interface{}) (interface{}, error) { return "a", nil }),
		makeTool("b", func(_ context.Context, _ map[string]interface{}) (interface{}, error) { return "b", nil }),
	}

	wrapped := ChainAll(tools, WithHooks(reg))
	require.Len(t, wrapped, 2)

	for i, w := range wrapped {
		result, err := w.Handler(context.Background(), nil)
		require.NoError(t, err)
		assert.Equal(t, tools[i].Name, result)
	}
}

func TestWithHooks_ToolErrorPassedToPostHook(t *testing.T) {
	t.Parallel()

	postHook := &stubPostHook{name: "observer", priority: 1}
	reg := NewHookRegistry()
	reg.RegisterPost(postHook)

	toolErr := errors.New("tool failure")
	tool := makeTool("failing_tool", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		return nil, toolErr
	})

	wrapped := Chain(tool, WithHooks(reg))
	_, err := wrapped.Handler(context.Background(), nil)

	assert.ErrorIs(t, err, toolErr)
	assert.Equal(t, toolErr, postHook.gotErr)
}

func TestWithHooks_PreHookObserveContinuesExecution(t *testing.T) {
	t.Parallel()

	reg := NewHookRegistry()
	reg.RegisterPre(&stubPreHook{
		name:     "observer",
		priority: 1,
		result:   PreHookResult{Action: Observe, ObserveReason: "scripting language detected"},
	})

	var handlerCalled bool
	tool := makeTool("exec", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		handlerCalled = true
		return "tool-result", nil
	})

	wrapped := Chain(tool, WithHooks(reg))
	result, err := wrapped.Handler(context.Background(), map[string]interface{}{"command": "python -c 'print(1)'"})

	require.NoError(t, err)
	assert.True(t, handlerCalled, "handler should be called when action is Observe")
	assert.Equal(t, "tool-result", result)
}

// --- test helpers ---

type captureHookCtxPreHook struct {
	captured *HookContext
}

func (h *captureHookCtxPreHook) Name() string  { return "capture" }
func (h *captureHookCtxPreHook) Priority() int { return 1 }
func (h *captureHookCtxPreHook) Pre(ctx HookContext) (PreHookResult, error) {
	*h.captured = ctx
	return PreHookResult{Action: Continue}, nil
}
