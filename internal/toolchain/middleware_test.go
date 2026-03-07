package toolchain

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/tools/browser"
)

func makeTool(name string, handler agent.ToolHandler) *agent.Tool {
	return &agent.Tool{
		Name:    name,
		Handler: handler,
	}
}

func TestChain_NoMiddleware(t *testing.T) {
	t.Parallel()

	tool := makeTool("test", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		return "ok", nil
	})
	result := Chain(tool)
	assert.Same(t, tool, result)
}

func TestChain_OrderOuterToInner(t *testing.T) {
	t.Parallel()

	var order []string

	mw := func(label string) Middleware {
		return func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
			return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				order = append(order, label+":before")
				result, err := next(ctx, params)
				order = append(order, label+":after")
				return result, err
			}
		}
	}

	tool := makeTool("test", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		order = append(order, "handler")
		return "ok", nil
	})

	wrapped := Chain(tool, mw("A"), mw("B"), mw("C"))
	_, _ = wrapped.Handler(context.Background(), nil)

	want := []string{"A:before", "B:before", "C:before", "handler", "C:after", "B:after", "A:after"}
	assert.Equal(t, want, order)
}

func TestChain_PreservesToolMetadata(t *testing.T) {
	t.Parallel()

	tool := &agent.Tool{
		Name:        "my_tool",
		Description: "desc",
		SafetyLevel: agent.SafetyLevelDangerous,
		Parameters:  map[string]interface{}{"key": "val"},
		Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	}

	noop := func(_ *agent.Tool, next agent.ToolHandler) agent.ToolHandler { return next }
	result := Chain(tool, noop)

	assert.Equal(t, tool.Name, result.Name)
	assert.Equal(t, tool.Description, result.Description)
	assert.Equal(t, tool.SafetyLevel, result.SafetyLevel)
}

func TestChainAll_WrapsAllTools(t *testing.T) {
	t.Parallel()

	var calls int
	counter := func(_ *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
		return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			calls++
			return next(ctx, params)
		}
	}

	tools := []*agent.Tool{
		makeTool("a", func(_ context.Context, _ map[string]interface{}) (interface{}, error) { return nil, nil }),
		makeTool("b", func(_ context.Context, _ map[string]interface{}) (interface{}, error) { return nil, nil }),
		makeTool("c", func(_ context.Context, _ map[string]interface{}) (interface{}, error) { return nil, nil }),
	}

	wrapped := ChainAll(tools, counter)
	for _, w := range wrapped {
		_, _ = w.Handler(context.Background(), nil)
	}

	assert.Equal(t, 3, calls)
}

func TestChainAll_NoMiddleware(t *testing.T) {
	t.Parallel()

	tools := []*agent.Tool{
		makeTool("a", nil),
		makeTool("b", nil),
	}
	result := ChainAll(tools)
	require.Len(t, result, len(tools))
	for i, r := range result {
		assert.Same(t, tools[i], r)
	}
}

func TestConditionalMiddleware_BrowserRecoverySkipsNonBrowser(t *testing.T) {
	t.Parallel()

	var called bool
	// Simulate WithBrowserRecovery's conditional logic: only applies to browser_ tools.
	conditional := func(tool *agent.Tool, next agent.ToolHandler) agent.ToolHandler {
		if tool.Name != "browser_navigate" {
			return next
		}
		return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			called = true
			return next(ctx, params)
		}
	}

	// Non-browser tool: middleware should be skipped.
	tool := makeTool("exec", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		return "ok", nil
	})
	wrapped := Chain(tool, conditional)
	_, _ = wrapped.Handler(context.Background(), nil)
	assert.False(t, called, "conditional middleware should not have been called for non-browser tool")

	// Browser tool: middleware should be called.
	browserTool := makeTool("browser_navigate", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		return "ok", nil
	})
	wrapped = Chain(browserTool, conditional)
	_, _ = wrapped.Handler(context.Background(), nil)
	assert.True(t, called, "conditional middleware should have been called for browser tool")
}

func TestMiddleware_ShortCircuit(t *testing.T) {
	t.Parallel()

	denied := errors.New("denied")
	blocker := func(_ *agent.Tool, _ agent.ToolHandler) agent.ToolHandler {
		return func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			return nil, denied
		}
	}

	var innerCalled bool
	tool := makeTool("test", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		innerCalled = true
		return "ok", nil
	})

	wrapped := Chain(tool, blocker)
	_, err := wrapped.Handler(context.Background(), nil)
	assert.ErrorIs(t, err, denied)
	assert.False(t, innerCalled, "inner handler should not have been called when middleware short-circuits")
}

func TestNeedsApproval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give     string
		tool     *agent.Tool
		ic       config.InterceptorConfig
		wantNeed bool
	}{
		{
			give:     "exempt tool bypasses approval",
			tool:     &agent.Tool{Name: "fs_read", SafetyLevel: agent.SafetyLevelDangerous},
			ic:       config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyAll, ExemptTools: []string{"fs_read"}},
			wantNeed: false,
		},
		{
			give:     "sensitive tool always requires approval",
			tool:     &agent.Tool{Name: "custom", SafetyLevel: agent.SafetyLevelSafe},
			ic:       config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyNone, SensitiveTools: []string{"custom"}},
			wantNeed: true,
		},
		{
			give:     "policy all requires all tools",
			tool:     &agent.Tool{Name: "safe_tool", SafetyLevel: agent.SafetyLevelSafe},
			ic:       config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyAll},
			wantNeed: true,
		},
		{
			give:     "policy dangerous only dangerous tools",
			tool:     &agent.Tool{Name: "exec", SafetyLevel: agent.SafetyLevelDangerous},
			ic:       config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyDangerous},
			wantNeed: true,
		},
		{
			give:     "policy dangerous skips safe tools",
			tool:     &agent.Tool{Name: "fs_read", SafetyLevel: agent.SafetyLevelSafe},
			ic:       config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyDangerous},
			wantNeed: false,
		},
		{
			give:     "policy configured only sensitive tools",
			tool:     &agent.Tool{Name: "exec", SafetyLevel: agent.SafetyLevelDangerous},
			ic:       config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyConfigured},
			wantNeed: false,
		},
		{
			give:     "policy none disables all",
			tool:     &agent.Tool{Name: "exec", SafetyLevel: agent.SafetyLevelDangerous},
			ic:       config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyNone},
			wantNeed: false,
		},
		{
			give:     "unknown policy fails safe",
			tool:     &agent.Tool{Name: "exec", SafetyLevel: agent.SafetyLevelSafe},
			ic:       config.InterceptorConfig{ApprovalPolicy: "unknown"},
			wantNeed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got := NeedsApproval(tt.tool, tt.ic)
			assert.Equal(t, tt.wantNeed, got)
		})
	}
}

func TestBuildApprovalSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		toolName   string
		params     map[string]interface{}
		wantPrefix string
	}{
		{
			give:       "exec tool",
			toolName:   "exec",
			params:     map[string]interface{}{"command": "ls -la"},
			wantPrefix: "Execute: ls -la",
		},
		{
			give:       "fs_write tool",
			toolName:   "fs_write",
			params:     map[string]interface{}{"path": "/tmp/test.txt", "content": "hello"},
			wantPrefix: "Write to /tmp/test.txt (5 bytes)",
		},
		{
			give:       "unknown tool fallback",
			toolName:   "custom_tool",
			params:     map[string]interface{}{},
			wantPrefix: "Tool: custom_tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got := BuildApprovalSummary(tt.toolName, tt.params)
			assert.Equal(t, tt.wantPrefix, got)
		})
	}
}

// --- WithLearning middleware tests ---

type mockObserver struct {
	calls []observerCall
}

type observerCall struct {
	sessionKey string
	toolName   string
	params     map[string]interface{}
	result     interface{}
	err        error
}

func (m *mockObserver) OnToolResult(_ context.Context, sessionKey, toolName string, params map[string]interface{}, result interface{}, err error) {
	m.calls = append(m.calls, observerCall{
		sessionKey: sessionKey,
		toolName:   toolName,
		params:     params,
		result:     result,
		err:        err,
	})
}

func TestWithLearning_ObservesToolResult(t *testing.T) {
	t.Parallel()

	obs := &mockObserver{}
	mw := WithLearning(obs)

	tool := makeTool("my_tool", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		return "result-value", nil
	})

	wrapped := Chain(tool, mw)
	params := map[string]interface{}{"key": "val"}
	result, err := wrapped.Handler(context.Background(), params)

	require.NoError(t, err)
	assert.Equal(t, "result-value", result)
	require.Len(t, obs.calls, 1)
	assert.Equal(t, "my_tool", obs.calls[0].toolName)
	assert.Equal(t, "result-value", obs.calls[0].result)
	assert.NoError(t, obs.calls[0].err)
}

func TestWithLearning_ObservesError(t *testing.T) {
	t.Parallel()

	obs := &mockObserver{}
	mw := WithLearning(obs)
	wantErr := errors.New("tool failed")

	tool := makeTool("fail_tool", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		return nil, wantErr
	})

	wrapped := Chain(tool, mw)
	_, err := wrapped.Handler(context.Background(), nil)

	assert.ErrorIs(t, err, wantErr)
	require.Len(t, obs.calls, 1)
	assert.Equal(t, wantErr, obs.calls[0].err)
}

// --- WithApproval middleware tests ---

type mockApprovalProvider struct {
	response approval.ApprovalResponse
	err      error
	received *approval.ApprovalRequest
}

func (m *mockApprovalProvider) RequestApproval(_ context.Context, req approval.ApprovalRequest) (approval.ApprovalResponse, error) {
	m.received = &req
	return m.response, m.err
}

func (m *mockApprovalProvider) CanHandle(_ string) bool { return true }

func TestWithApproval_DeniedExecution(t *testing.T) {
	t.Parallel()

	ap := &mockApprovalProvider{response: approval.ApprovalResponse{Approved: false}}
	ic := config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyAll}

	tool := &agent.Tool{
		Name:        "exec",
		SafetyLevel: agent.SafetyLevelDangerous,
		Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			t.Error("handler should not be called when denied")
			return nil, nil
		},
	}

	mw := WithApproval(ic, ap, nil, nil)
	wrapped := Chain(tool, mw)
	_, err := wrapped.Handler(context.Background(), nil)

	require.Error(t, err)
	require.NotNil(t, ap.received, "approval provider was not consulted")
}

func TestWithApproval_ApprovedExecution(t *testing.T) {
	t.Parallel()

	ap := &mockApprovalProvider{response: approval.ApprovalResponse{Approved: true}}
	ic := config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyAll}

	var called bool
	tool := &agent.Tool{
		Name:        "exec",
		SafetyLevel: agent.SafetyLevelDangerous,
		Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			called = true
			return "ok", nil
		},
	}

	mw := WithApproval(ic, ap, nil, nil)
	wrapped := Chain(tool, mw)
	result, err := wrapped.Handler(context.Background(), nil)

	require.NoError(t, err)
	assert.True(t, called, "handler was not called after approval")
	assert.Equal(t, "ok", result)
}

func TestWithApproval_GrantStoreAutoApproves(t *testing.T) {
	t.Parallel()

	ap := &mockApprovalProvider{response: approval.ApprovalResponse{Approved: false}}
	gs := approval.NewGrantStore()
	gs.Grant("", "exec") // pre-grant for empty session key
	ic := config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyAll}

	var called bool
	tool := &agent.Tool{
		Name:        "exec",
		SafetyLevel: agent.SafetyLevelDangerous,
		Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			called = true
			return "ok", nil
		},
	}

	mw := WithApproval(ic, ap, gs, nil)
	wrapped := Chain(tool, mw)
	_, err := wrapped.Handler(context.Background(), nil)

	require.NoError(t, err)
	assert.True(t, called, "handler should be auto-approved via grant store")
	assert.Nil(t, ap.received, "approval provider should not have been consulted (grant store bypass)")
}

func TestWithApproval_AlwaysAllowRecordsGrant(t *testing.T) {
	t.Parallel()

	ap := &mockApprovalProvider{response: approval.ApprovalResponse{Approved: true, AlwaysAllow: true}}
	gs := approval.NewGrantStore()
	ic := config.InterceptorConfig{ApprovalPolicy: config.ApprovalPolicyAll}

	tool := &agent.Tool{
		Name:        "exec",
		SafetyLevel: agent.SafetyLevelDangerous,
		Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}

	mw := WithApproval(ic, ap, gs, nil)
	wrapped := Chain(tool, mw)
	_, _ = wrapped.Handler(context.Background(), nil)

	assert.True(t, gs.IsGranted("", "exec"), "grant should have been recorded for always-allow response")
}

func TestWithApproval_ExemptToolSkipsApproval(t *testing.T) {
	t.Parallel()

	ap := &mockApprovalProvider{response: approval.ApprovalResponse{Approved: false}}
	ic := config.InterceptorConfig{
		ApprovalPolicy: config.ApprovalPolicyAll,
		ExemptTools:    []string{"fs_read"},
	}

	var called bool
	tool := &agent.Tool{
		Name:        "fs_read",
		SafetyLevel: agent.SafetyLevelSafe,
		Handler: func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
			called = true
			return "ok", nil
		},
	}

	mw := WithApproval(ic, ap, nil, nil)
	wrapped := Chain(tool, mw)
	_, err := wrapped.Handler(context.Background(), nil)

	require.NoError(t, err)
	assert.True(t, called, "exempt tool should bypass approval")
}

// --- WithBrowserRecovery middleware tests ---

func TestWithBrowserRecovery_PanicRecovery(t *testing.T) {
	t.Parallel()

	mw := WithBrowserRecovery(nil) // nil SessionManager — Close will not be called on first attempt

	attempts := 0
	tool := makeTool("browser_navigate", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		attempts++
		if attempts == 1 {
			panic("rod crashed")
		}
		return "recovered", nil
	})

	wrapped := Chain(tool, mw)
	result, err := wrapped.Handler(context.Background(), nil)

	if err != nil {
		// Expected: the panic was caught and wrapped.
		assert.ErrorIs(t, err, browser.ErrBrowserPanic)
	} else {
		// If somehow recovery + retry worked, check result.
		assert.Equal(t, "recovered", result)
	}
	assert.GreaterOrEqual(t, attempts, 1, "handler should have been called at least once")
}

func TestWithBrowserRecovery_ErrorRetryOnce(t *testing.T) {
	t.Parallel()

	mw := WithBrowserRecovery(nil)

	tool := makeTool("browser_navigate", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		return nil, fmt.Errorf("connection lost: %w", browser.ErrBrowserPanic)
	})

	wrapped := Chain(tool, mw)
	_, err := wrapped.Handler(context.Background(), nil)

	require.Error(t, err)
}

func TestWithBrowserRecovery_NonBrowserToolPassthrough(t *testing.T) {
	t.Parallel()

	mw := WithBrowserRecovery(nil)

	var called bool
	tool := makeTool("exec", func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	})

	wrapped := Chain(tool, mw)
	result, err := wrapped.Handler(context.Background(), nil)

	require.NoError(t, err)
	assert.True(t, called, "handler was not called")
	assert.Equal(t, "ok", result)
}

// --- BuildApprovalSummary extended tests ---

func TestBuildApprovalSummary_Extended(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give     string
		toolName string
		params   map[string]interface{}
		want     string
	}{
		{
			give:     "fs_edit tool",
			toolName: "fs_edit",
			params:   map[string]interface{}{"path": "/tmp/main.go"},
			want:     "Edit file: /tmp/main.go",
		},
		{
			give:     "fs_delete tool",
			toolName: "fs_delete",
			params:   map[string]interface{}{"path": "/tmp/old.log"},
			want:     "Delete: /tmp/old.log",
		},
		{
			give:     "browser_navigate tool",
			toolName: "browser_navigate",
			params:   map[string]interface{}{"url": "https://example.com"},
			want:     "Navigate to: https://example.com",
		},
		{
			give:     "browser_action with selector",
			toolName: "browser_action",
			params:   map[string]interface{}{"action": "click", "selector": "#submit-btn"},
			want:     "Browser click on: #submit-btn",
		},
		{
			give:     "browser_action without selector",
			toolName: "browser_action",
			params:   map[string]interface{}{"action": "screenshot"},
			want:     "Browser action: screenshot",
		},
		{
			give:     "secrets_store tool",
			toolName: "secrets_store",
			params:   map[string]interface{}{"name": "api_key"},
			want:     "Store secret: api_key",
		},
		{
			give:     "secrets_get tool",
			toolName: "secrets_get",
			params:   map[string]interface{}{"name": "api_key"},
			want:     "Retrieve secret: api_key",
		},
		{
			give:     "secrets_delete tool",
			toolName: "secrets_delete",
			params:   map[string]interface{}{"name": "old_key"},
			want:     "Delete secret: old_key",
		},
		{
			give:     "crypto_encrypt tool",
			toolName: "crypto_encrypt",
			params:   map[string]interface{}{},
			want:     "Encrypt data",
		},
		{
			give:     "crypto_decrypt tool",
			toolName: "crypto_decrypt",
			params:   map[string]interface{}{},
			want:     "Decrypt ciphertext",
		},
		{
			give:     "crypto_sign tool",
			toolName: "crypto_sign",
			params:   map[string]interface{}{},
			want:     "Generate digital signature",
		},
		{
			give:     "payment_send tool",
			toolName: "payment_send",
			params:   map[string]interface{}{"amount": "1.5", "to": "0xABC123", "purpose": "test"},
			want:     "Send 1.5 USDC to 0xABC123 (test)",
		},
		{
			give:     "payment_create_wallet tool",
			toolName: "payment_create_wallet",
			params:   map[string]interface{}{},
			want:     "Create new blockchain wallet",
		},
		{
			give:     "cron_add tool",
			toolName: "cron_add",
			params:   map[string]interface{}{"name": "daily-backup", "schedule_type": "cron", "schedule": "0 0 * * *"},
			want:     "Create cron job: daily-backup (cron=0 0 * * *)",
		},
		{
			give:     "cron_remove tool",
			toolName: "cron_remove",
			params:   map[string]interface{}{"id": "job-123"},
			want:     "Remove cron job: job-123",
		},
		{
			give:     "bg_submit tool",
			toolName: "bg_submit",
			params:   map[string]interface{}{"prompt": "analyze the data"},
			want:     "Submit background task: analyze the data",
		},
		{
			give:     "workflow_run with file",
			toolName: "workflow_run",
			params:   map[string]interface{}{"file_path": "pipelines/deploy.yaml"},
			want:     "Run workflow: pipelines/deploy.yaml",
		},
		{
			give:     "workflow_run inline",
			toolName: "workflow_run",
			params:   map[string]interface{}{},
			want:     "Run inline workflow",
		},
		{
			give:     "workflow_cancel tool",
			toolName: "workflow_cancel",
			params:   map[string]interface{}{"run_id": "run-456"},
			want:     "Cancel workflow: run-456",
		},
		{
			give:     "p2p_pay tool",
			toolName: "p2p_pay",
			params:   map[string]interface{}{"amount": "0.5", "peer_did": "did:example:peer1", "memo": "thanks"},
			want:     "Pay 0.5 USDC to peer did:example:peer... (thanks)",
		},
		{
			give:     "p2p_pay no memo",
			toolName: "p2p_pay",
			params:   map[string]interface{}{"amount": "1.0", "peer_did": "did:example:x"},
			want:     "Pay 1.0 USDC to peer did:example:x (P2P payment)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got := BuildApprovalSummary(tt.toolName, tt.params)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give   string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is a ..."},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d/%s", tt.maxLen, tt.give), func(t *testing.T) {
			t.Parallel()

			got := Truncate(tt.give, tt.maxLen)
			assert.Equal(t, tt.want, got)
		})
	}
}
