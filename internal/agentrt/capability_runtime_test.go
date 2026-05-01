package agentrt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/toolchain"
)

func TestCapabilityRuntime_BlockedDynamicAllowedToolsUpdatesProjection(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:             "arun-cap",
		RequestedAgent: "operator",
		Status:         AgentRunRunning,
		AllowedTools:   []string{"fs_read"},
	}))

	rt := NewCapabilityRuntime(store, &CapabilityPolicy{}, func(toolName string) agent.SafetyLevel {
		assert.Equal(t, "exec", toolName)
		return agent.SafetyLevelDangerous
	})

	err := rt.HandleBlockedToolCall("arun-cap", toolchain.BlockedToolCall{
		ToolName:    "exec",
		BlockReason: dynamicAllowedToolsBlockReason,
	})
	require.NoError(t, err)

	run, err := store.Get("arun-cap")
	require.NoError(t, err)
	assert.Equal(t, AgentRunRunning, run.Status)
	assert.Equal(t, AgentRunConditionBlockedWaitingApproval, run.RuntimeCondition)
	assert.Equal(t, "dangerous tool requires approval", run.BlockedReason)
	assert.Equal(t, "grant-arun-cap-exec", run.GrantRequestID)
}

func TestCapabilityRuntime_OutsideScopeDenialDoesNotRequestApproval(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:             "arun-plan",
		RequestedAgent: "planner",
		Status:         AgentRunRunning,
		AllowedTools:   []string{"agent_wait"},
	}))

	rt := NewCapabilityRuntime(store, &CapabilityPolicy{}, nil)

	err := rt.HandleBlockedToolCall("arun-plan", toolchain.BlockedToolCall{
		ToolName:    "exec",
		BlockReason: dynamicAllowedToolsBlockReason,
	})
	require.NoError(t, err)

	run, err := store.Get("arun-plan")
	require.NoError(t, err)
	assert.Equal(t, AgentRunConditionNone, run.RuntimeCondition)
	assert.Contains(t, run.BlockedReason, "outside role maximum scope")
	assert.Empty(t, run.GrantRequestID)
}

func TestCapabilityRuntime_ApplyGrantClearsBlockedProjection(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:               "arun-grant",
		RequestedAgent:   "operator",
		Status:           AgentRunRunning,
		AllowedTools:     []string{"fs_read"},
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "dangerous tool requires approval",
		GrantRequestID:   "grant-arun-grant-exec",
	}))

	policy := &CapabilityPolicy{}
	rt := NewCapabilityRuntime(store, policy, nil)

	err := rt.ApplyGrant("arun-grant", "exec")
	require.NoError(t, err)

	run, err := store.Get("arun-grant")
	require.NoError(t, err)
	assert.Equal(t, AgentRunConditionNone, run.RuntimeCondition)
	assert.Empty(t, run.BlockedReason)
	assert.Empty(t, run.GrantRequestID)
	assert.Contains(t, run.AllowedTools, "exec")
	require.True(t, policy.ActiveGrants["arun-grant"]["exec"])

	ctx := rt.ContextForRun(context.Background(), run)
	assert.Contains(t, ctxkeys.DynamicAllowedToolsFromContext(ctx), "exec")
}

func TestCapabilityRuntime_ContextForRunWiresBlockedHookToProjection(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:             "arun-hook",
		RequestedAgent: "operator",
		Status:         AgentRunRunning,
		AllowedTools:   []string{"fs_read"},
	}))

	rt := NewCapabilityRuntime(store, &CapabilityPolicy{}, nil)
	run, err := store.Get("arun-hook")
	require.NoError(t, err)

	registry := toolchain.NewHookRegistry()
	registry.RegisterPre(toolchain.NewAgentAccessControlHook(nil))

	tool := &agent.Tool{
		Name: "exec",
		Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
			t.Fatal("handler should not be called when blocked")
			return nil, nil
		},
	}

	wrapped := toolchain.Chain(tool, toolchain.WithHooks(registry))
	_, err = wrapped.Handler(rt.ContextForRun(context.Background(), run), map[string]interface{}{
		"command": "echo hello",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), dynamicAllowedToolsBlockReason)

	updated, err := store.Get("arun-hook")
	require.NoError(t, err)
	assert.Equal(t, AgentRunConditionBlockedWaitingApproval, updated.RuntimeCondition)
	assert.Equal(t, "dangerous tool requires approval", updated.BlockedReason)
	assert.Equal(t, "grant-arun-hook-exec", updated.GrantRequestID)
}

func TestCapabilityRuntime_SameContextSeesLiveGrantAfterApplyGrant(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:             "arun-live",
		RequestedAgent: "operator",
		Status:         AgentRunRunning,
		AllowedTools:   []string{"fs_read"},
	}))

	rt := NewCapabilityRuntime(store, &CapabilityPolicy{}, nil)
	run, err := store.Get("arun-live")
	require.NoError(t, err)

	registry := toolchain.NewHookRegistry()
	registry.RegisterPre(toolchain.NewAgentAccessControlHook(nil))

	tool := &agent.Tool{
		Name: "exec",
		Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}

	ctx := rt.ContextForRun(context.Background(), run)
	wrapped := toolchain.Chain(tool, toolchain.WithHooks(registry))

	_, err = wrapped.Handler(ctx, map[string]interface{}{"command": "echo hello"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), dynamicAllowedToolsBlockReason)

	require.NoError(t, rt.ApplyGrant("arun-live", "exec"))

	result, err := wrapped.Handler(ctx, map[string]interface{}{"command": "echo hello"})
	require.NoError(t, err)
	assert.Equal(t, "ok", result)
}
