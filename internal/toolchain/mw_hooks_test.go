package toolchain

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/session"
)

func TestWithHooks_NormalFlow(t *testing.T) {
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

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Error("handler was not called")
	}
	if result != "result-value" {
		t.Errorf("result = %v, want %q", result, "result-value")
	}
	if !preHook.called {
		t.Error("pre-hook was not called")
	}
	if !postHook.called {
		t.Error("post-hook was not called")
	}
	if postHook.gotResult != "result-value" {
		t.Errorf("post-hook gotResult = %v, want %q", postHook.gotResult, "result-value")
	}
}

func TestWithHooks_PreHookBlocks(t *testing.T) {
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

	if err == nil {
		t.Fatal("expected error when blocked")
	}
	if !strings.Contains(err.Error(), "rate limit exceeded") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "rate limit exceeded")
	}
	if handlerCalled {
		t.Error("handler should not be called when blocked")
	}
}

func TestWithHooks_PreHookModifiesParams(t *testing.T) {
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

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, ok := receivedParams["key"].(string); !ok || v != "modified-value" {
		t.Errorf("handler received params[key] = %v, want %q", receivedParams["key"], "modified-value")
	}
}

func TestWithHooks_PostHookErrorDoesNotAffectResult(t *testing.T) {
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
	if err != nil {
		t.Fatalf("unexpected error: %v (post-hook errors should be logged, not returned)", err)
	}
	if result != "tool-result" {
		t.Errorf("result = %v, want %q", result, "tool-result")
	}
}

func TestWithHooks_PreHookError(t *testing.T) {
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

	if err == nil {
		t.Fatal("expected error from pre-hook failure")
	}
	if handlerCalled {
		t.Error("handler should not be called when pre-hook errors")
	}
}

func TestWithHooks_ContextPropagation(t *testing.T) {
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

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedCtx.ToolName != "my_tool" {
		t.Errorf("ToolName = %q, want %q", capturedCtx.ToolName, "my_tool")
	}
	if capturedCtx.AgentName != "researcher" {
		t.Errorf("AgentName = %q, want %q", capturedCtx.AgentName, "researcher")
	}
	if capturedCtx.SessionKey != "session-abc" {
		t.Errorf("SessionKey = %q, want %q", capturedCtx.SessionKey, "session-abc")
	}
}

func TestWithHooks_CompatibleWithChainAll(t *testing.T) {
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
	if len(wrapped) != 2 {
		t.Fatalf("len = %d, want 2", len(wrapped))
	}

	for i, w := range wrapped {
		result, err := w.Handler(context.Background(), nil)
		if err != nil {
			t.Errorf("tool[%d] error: %v", i, err)
		}
		if result != tools[i].Name {
			t.Errorf("tool[%d] result = %v, want %q", i, result, tools[i].Name)
		}
	}
}

func TestWithHooks_ToolErrorPassedToPostHook(t *testing.T) {
	postHook := &stubPostHook{name: "observer", priority: 1}
	reg := NewHookRegistry()
	reg.RegisterPost(postHook)

	toolErr := errors.New("tool failure")
	tool := makeTool("failing_tool", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		return nil, toolErr
	})

	wrapped := Chain(tool, WithHooks(reg))
	_, err := wrapped.Handler(context.Background(), nil)

	if !errors.Is(err, toolErr) {
		t.Errorf("err = %v, want %v", err, toolErr)
	}
	if postHook.gotErr != toolErr {
		t.Errorf("post-hook gotErr = %v, want %v", postHook.gotErr, toolErr)
	}
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
