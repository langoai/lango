package agentrt

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/logging"
	"github.com/langoai/lango/internal/toolchain"
)

type delayedAllowedToolStore struct {
	base                     *InMemoryAgentRunStore
	mu                       sync.Mutex
	getCalls                 int
	blockedProjectionAttempt bool
}

func (s *delayedAllowedToolStore) Create(run *AgentRun) error {
	return s.base.Create(run)
}

func (s *delayedAllowedToolStore) Get(id string) (*AgentRun, error) {
	s.mu.Lock()
	s.getCalls++
	callNumber := s.getCalls
	s.mu.Unlock()

	if id == "arun-1" && callNumber == 2 {
		if err := s.base.UpdateProjection(id, RunProjectionPatch{
			AddAllowedTool: "fs_write",
		}); err != nil {
			return nil, err
		}
	}

	return s.base.Get(id)
}

func (s *delayedAllowedToolStore) List() []*AgentRun {
	return s.base.List()
}

func (s *delayedAllowedToolStore) UpdateStatus(id string, status AgentRunStatus, result, errMsg string) error {
	return s.base.UpdateStatus(id, status, result, errMsg)
}

func (s *delayedAllowedToolStore) UpdateProjection(id string, patch RunProjectionPatch) error {
	if patch.ApplyRuntimeCondition && patch.RuntimeCondition == AgentRunConditionBlockedWaitingApproval {
		s.mu.Lock()
		s.blockedProjectionAttempt = true
		s.mu.Unlock()
	}
	return s.base.UpdateProjection(id, patch)
}

func (s *delayedAllowedToolStore) Cancel(id string) error {
	return s.base.Cancel(id)
}

type blockedProjectionGrantStore struct {
	base                *InMemoryAgentRunStore
	targetID            string
	mu                  sync.Mutex
	onBlockedProjection func() error
	onPostBlockedGet    func() error
	triggered           bool
	blockedWriteCount   int
	postGetTriggered    bool
}

func (s *blockedProjectionGrantStore) Create(run *AgentRun) error {
	return s.base.Create(run)
}

func (s *blockedProjectionGrantStore) Get(id string) (*AgentRun, error) {
	targetID := s.targetID
	if targetID == "" {
		targetID = "arun-2"
	}

	s.mu.Lock()
	shouldTrigger := id == targetID && s.blockedWriteCount > 0 && !s.postGetTriggered
	if shouldTrigger {
		s.postGetTriggered = true
	}
	hook := s.onPostBlockedGet
	s.mu.Unlock()

	if shouldTrigger && hook != nil {
		if err := hook(); err != nil {
			return nil, err
		}
	}

	return s.base.Get(id)
}

func (s *blockedProjectionGrantStore) List() []*AgentRun {
	return s.base.List()
}

func (s *blockedProjectionGrantStore) UpdateStatus(id string, status AgentRunStatus, result, errMsg string) error {
	return s.base.UpdateStatus(id, status, result, errMsg)
}

func (s *blockedProjectionGrantStore) UpdateProjection(id string, patch RunProjectionPatch) error {
	targetID := s.targetID
	if targetID == "" {
		targetID = "arun-2"
	}

	s.mu.Lock()
	shouldTrigger := id == targetID &&
		patch.ApplyRuntimeCondition &&
		patch.RuntimeCondition == AgentRunConditionBlockedWaitingApproval &&
		!s.triggered
	if shouldTrigger {
		s.triggered = true
	}
	hook := s.onBlockedProjection
	s.mu.Unlock()

	if shouldTrigger && hook != nil {
		if err := hook(); err != nil {
			return err
		}
	}

	if err := s.base.UpdateProjection(id, patch); err != nil {
		return err
	}
	if patch.ApplyRuntimeCondition && patch.RuntimeCondition == AgentRunConditionBlockedWaitingApproval {
		s.mu.Lock()
		s.blockedWriteCount++
		s.mu.Unlock()
	}

	return nil
}

func (s *blockedProjectionGrantStore) Cancel(id string) error {
	return s.base.Cancel(id)
}

type failingAgentRunStore struct {
	err error
}

func (s *failingAgentRunStore) Create(run *AgentRun) error {
	return nil
}

func (s *failingAgentRunStore) Get(id string) (*AgentRun, error) {
	return nil, s.err
}

func (s *failingAgentRunStore) List() []*AgentRun {
	return nil
}

func (s *failingAgentRunStore) UpdateStatus(id string, status AgentRunStatus, result, errMsg string) error {
	return nil
}

func (s *failingAgentRunStore) UpdateProjection(id string, patch RunProjectionPatch) error {
	return nil
}

func (s *failingAgentRunStore) Cancel(id string) error {
	return nil
}

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
	assert.Equal(t, 1, run.GrantAttempt)
	assert.Equal(t, "pending", run.GrantState)
}

func TestCapabilityRuntime_BlockedToolSinkLogsHandleErrors(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, logging.Init(logging.LogConfig{
		Level:  "error",
		Format: "json",
		Writer: &buf,
	}))

	rt := NewCapabilityRuntime(&failingAgentRunStore{err: fmt.Errorf("boom")}, &CapabilityPolicy{}, nil)
	sink := rt.BlockedToolSinkForRun("arun-log")
	sink(toolchain.BlockedToolCall{
		ToolName:    "fs_write",
		BlockReason: dynamicAllowedToolsBlockReason,
	})

	output := buf.String()
	assert.Contains(t, output, "blocked tool call handling failed")
	assert.Contains(t, output, "arun-log")
	assert.Contains(t, output, "fs_write")
	assert.True(t, strings.Contains(output, "boom"), "expected logged error, got %q", output)
}

func TestCapabilityRuntime_OutsideScopeDenialDoesNotRequestApproval(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:               "arun-plan",
		RequestedAgent:   "planner",
		Status:           AgentRunRunning,
		AllowedTools:     []string{"agent_wait"},
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "dangerous tool requires approval",
		GrantRequestID:   "grant-arun-plan-exec",
		GrantAttempt:     2,
		GrantState:       "pending",
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
	assert.Equal(t, 0, run.GrantAttempt)
	assert.Equal(t, "denied", run.GrantState)
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
		GrantAttempt:     1,
		GrantState:       "pending",
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
	assert.Equal(t, 0, run.GrantAttempt)
	assert.Equal(t, "granted", run.GrantState)
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
	assert.Equal(t, 1, updated.GrantAttempt)
	assert.Equal(t, "pending", updated.GrantState)
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

func TestCapabilityRuntime_RepeatedBlockedRequestIncrementsAttemptWithoutRotatingGrantRequestID(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:             "arun-repeat",
		RequestedAgent: "operator",
		Status:         AgentRunRunning,
		AllowedTools:   []string{"fs_read"},
	}))

	rt := NewCapabilityRuntime(store, &CapabilityPolicy{}, func(string) agent.SafetyLevel {
		return agent.SafetyLevelDangerous
	})

	call := toolchain.BlockedToolCall{
		ToolName:    "exec",
		BlockReason: dynamicAllowedToolsBlockReason,
	}

	require.NoError(t, rt.HandleBlockedToolCall("arun-repeat", call))
	first, err := store.Get("arun-repeat")
	require.NoError(t, err)

	require.NoError(t, rt.HandleBlockedToolCall("arun-repeat", call))
	second, err := store.Get("arun-repeat")
	require.NoError(t, err)

	assert.Equal(t, first.GrantRequestID, second.GrantRequestID)
	assert.Equal(t, 1, first.GrantAttempt)
	assert.Equal(t, 2, second.GrantAttempt)
	assert.Equal(t, "pending", second.GrantState)
}

func TestHandleBlockedToolCall_AllowedToolExpansionSkipsBlockedProjection(t *testing.T) {
	store := &delayedAllowedToolStore{base: NewInMemoryAgentRunStore()}
	require.NoError(t, store.Create(&AgentRun{
		ID:             "arun-1",
		RequestedAgent: "operator",
		Status:         AgentRunRunning,
		AllowedTools:   []string{"fs_read"},
	}))

	rt := NewCapabilityRuntime(store, &CapabilityPolicy{}, nil)
	require.NoError(t, rt.HandleBlockedToolCall("arun-1", toolchain.BlockedToolCall{
		ToolName:    "fs_write",
		BlockReason: dynamicAllowedToolsBlockReason,
	}))

	run, err := store.Get("arun-1")
	require.NoError(t, err)
	assert.Equal(t, AgentRunConditionNone, run.RuntimeCondition)
	assert.False(t, store.blockedProjectionAttempt)
}

func TestHandleBlockedToolCall_PostWriteReconciliationClearsBlockedProjectionAfterGrantWinsWriteWindow(t *testing.T) {
	policy := &CapabilityPolicy{}
	store := &blockedProjectionGrantStore{base: NewInMemoryAgentRunStore()}
	require.NoError(t, store.Create(&AgentRun{
		ID:             "arun-2",
		RequestedAgent: "operator",
		Status:         AgentRunRunning,
		AllowedTools:   []string{"fs_read"},
	}))

	rt := NewCapabilityRuntime(store, policy, nil)
	store.onBlockedProjection = func() error {
		rt.mu.Lock()
		policy.Grant("arun-2", "fs_write")
		rt.mu.Unlock()
		return store.base.UpdateProjection("arun-2", RunProjectionPatch{
			AddAllowedTool: "fs_write",
		})
	}

	err := rt.HandleBlockedToolCall("arun-2", toolchain.BlockedToolCall{
		ToolName:    "fs_write",
		BlockReason: dynamicAllowedToolsBlockReason,
	})
	require.NoError(t, err)

	run, err := store.Get("arun-2")
	require.NoError(t, err)
	assert.Equal(t, AgentRunConditionNone, run.RuntimeCondition)
	assert.Empty(t, run.BlockedReason)
	assert.Empty(t, run.GrantRequestID)
	assert.Equal(t, 0, run.GrantAttempt)
	assert.Empty(t, run.GrantState)
	assert.Contains(t, run.AllowedTools, "fs_write")
}

func TestHandleBlockedToolCall_PostWriteReconciliationPreservesNewerBlockedProjection(t *testing.T) {
	policy := &CapabilityPolicy{}
	store := &blockedProjectionGrantStore{base: NewInMemoryAgentRunStore(), targetID: "arun-3"}
	require.NoError(t, store.Create(&AgentRun{
		ID:             "arun-3",
		RequestedAgent: "operator",
		Status:         AgentRunRunning,
		AllowedTools:   []string{"fs_read"},
	}))

	rt := NewCapabilityRuntime(store, policy, nil)
	store.onPostBlockedGet = func() error {
		rt.mu.Lock()
		policy.Grant("arun-3", "fs_write")
		rt.mu.Unlock()
		if err := store.base.UpdateProjection("arun-3", RunProjectionPatch{
			AddAllowedTool: "fs_write",
		}); err != nil {
			return err
		}
		return store.base.UpdateProjection("arun-3", RunProjectionPatch{
			ApplyRuntimeCondition: true,
			ApplyBlockedReason:    true,
			ApplyGrantRequestID:   true,
			ApplyGrantAttempt:     true,
			ApplyGrantState:       true,
			RuntimeCondition:      AgentRunConditionBlockedWaitingApproval,
			BlockedReason:         "newer blocked projection",
			GrantRequestID:        "grant-arun-3-exec",
			GrantAttempt:          3,
			GrantState:            "pending",
		})
	}

	err := rt.HandleBlockedToolCall("arun-3", toolchain.BlockedToolCall{
		ToolName:    "fs_write",
		BlockReason: dynamicAllowedToolsBlockReason,
	})
	require.NoError(t, err)

	run, err := store.Get("arun-3")
	require.NoError(t, err)
	assert.Equal(t, AgentRunConditionBlockedWaitingApproval, run.RuntimeCondition)
	assert.Equal(t, "grant-arun-3-exec", run.GrantRequestID)
	assert.Equal(t, 3, run.GrantAttempt)
	assert.Equal(t, "pending", run.GrantState)
	assert.Equal(t, "newer blocked projection", run.BlockedReason)
	assert.Contains(t, run.AllowedTools, "fs_write")
}

func TestCapabilityRuntime_ReblockAfterDenyRestartsAttemptAtOneWithStableGrantRequestID(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:             "arun-reblock",
		RequestedAgent: "operator",
		Status:         AgentRunRunning,
		AllowedTools:   []string{"fs_read"},
	}))

	rt := NewCapabilityRuntime(store, &CapabilityPolicy{}, func(string) agent.SafetyLevel {
		return agent.SafetyLevelDangerous
	})

	call := toolchain.BlockedToolCall{
		ToolName:    "exec",
		BlockReason: dynamicAllowedToolsBlockReason,
	}

	require.NoError(t, rt.HandleBlockedToolCall("arun-reblock", call))
	first, err := store.Get("arun-reblock")
	require.NoError(t, err)

	require.NoError(t, store.UpdateProjection("arun-reblock", RunProjectionPatch{
		ApplyRuntimeCondition: true,
		ApplyBlockedReason:    true,
		ApplyGrantRequestID:   true,
		ApplyGrantAttempt:     true,
		ApplyGrantState:       true,
		RuntimeCondition:      AgentRunConditionNone,
		BlockedReason:         "denied once",
		GrantRequestID:        "",
		GrantAttempt:          0,
		GrantState:            "denied",
	}))

	require.NoError(t, rt.HandleBlockedToolCall("arun-reblock", call))
	second, err := store.Get("arun-reblock")
	require.NoError(t, err)

	assert.Equal(t, first.GrantRequestID, second.GrantRequestID)
	assert.Equal(t, 1, second.GrantAttempt)
	assert.Equal(t, "pending", second.GrantState)
}
