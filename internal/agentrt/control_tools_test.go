package agentrt

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/session"
)

// --- BuildControlTools ---

func TestBuildControlTools_ToolCount(t *testing.T) {
	cp := &AgentControlPlane{
		RunStore:   NewInMemoryAgentRunStore(),
		Projection: NewAgentRunProjection(NewInMemoryAgentRunStore()),
	}
	tools := BuildControlTools(cp)
	assert.Len(t, tools, 3)

	names := make(map[string]bool, 3)
	for _, tool := range tools {
		names[tool.Name] = true
	}
	assert.True(t, names["agent_spawn"])
	assert.True(t, names["agent_wait"])
	assert.True(t, names["agent_stop"])
}

// --- agent_spawn ---

func TestAgentSpawn_Basic(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	cp := &AgentControlPlane{
		RunStore:   store,
		Projection: NewAgentRunProjection(store),
	}
	tools := BuildControlTools(cp)
	spawnTool := findControlTool(t, tools, "agent_spawn")

	result, err := spawnTool.call(context.Background(), map[string]interface{}{
		"instruction": "analyze the data",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.NotEmpty(t, m["agent_id"])
	assert.Equal(t, "spawned", m["status"])
	assert.Equal(t, "", m["requested_agent"])

	run, err := store.Get(m["agent_id"].(string))
	require.NoError(t, err)
	assert.Equal(t, AgentRunSpawned, run.Status)
	assert.Equal(t, "analyze the data", run.Instruction)
	assert.Equal(t, 1, run.SpawnDepth)
}

func TestAgentSpawn_WithAgent(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	cp := &AgentControlPlane{
		RunStore:   store,
		Projection: NewAgentRunProjection(store),
	}
	tools := BuildControlTools(cp)
	spawnTool := findControlTool(t, tools, "agent_spawn")

	result, err := spawnTool.call(context.Background(), map[string]interface{}{
		"instruction": "fix the bug",
		"agent":       "debugger",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, "debugger", m["requested_agent"])

	run, err := store.Get(m["agent_id"].(string))
	require.NoError(t, err)
	assert.Contains(t, run.Instruction, "[System: This task is best handled by the 'debugger' specialist.]")
	assert.Contains(t, run.Instruction, "fix the bug")
	assert.Equal(t, "debugger", run.RequestedAgent)
}

func TestAgentSpawn_WithAllowedTools(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	cp := &AgentControlPlane{
		RunStore:   store,
		Projection: NewAgentRunProjection(store),
	}
	tools := BuildControlTools(cp)
	spawnTool := findControlTool(t, tools, "agent_spawn")

	result, err := spawnTool.call(context.Background(), map[string]interface{}{
		"instruction":   "restricted task",
		"allowed_tools": []interface{}{"fs_read", "web_search"},
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	run, err := store.Get(m["agent_id"].(string))
	require.NoError(t, err)
	assert.Equal(t, []string{"fs_read", "web_search"}, run.AllowedTools)
}

func TestAgentSpawn_WithAllowedToolsForBuiltinTeammate(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	cp := &AgentControlPlane{
		RunStore:   store,
		Projection: NewAgentRunProjection(store),
	}
	tools := BuildControlTools(cp)
	spawnTool := findControlTool(t, tools, "agent_spawn")

	result, err := spawnTool.call(context.Background(), map[string]interface{}{
		"instruction":   "operate on local files",
		"agent":         "operator",
		"allowed_tools": []interface{}{"fs_read", "exec"},
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, "operator", m["requested_agent"])

	run, err := store.Get(m["agent_id"].(string))
	require.NoError(t, err)
	assert.Equal(t, "operator", run.RequestedAgent)
	assert.Equal(t, []string{"fs_read", "exec"}, run.AllowedTools)
}

func TestAgentSpawn_WithSpawnReason(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	cp := &AgentControlPlane{
		RunStore:   store,
		Projection: NewAgentRunProjection(store),
	}
	tools := BuildControlTools(cp)
	spawnTool := findControlTool(t, tools, "agent_spawn")

	result, err := spawnTool.call(context.Background(), map[string]interface{}{
		"instruction":  "approval task",
		"spawn_reason": "delegated_for_approval",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, "delegated_for_approval", m["spawn_reason"])

	run, err := store.Get(m["agent_id"].(string))
	require.NoError(t, err)
	assert.Equal(t, "delegated_for_approval", run.SpawnReason)
}

func TestAgentSpawn_RejectsAllowedToolsOutsideRoleScope(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	cp := &AgentControlPlane{
		RunStore:   store,
		Projection: NewAgentRunProjection(store),
	}
	tools := BuildControlTools(cp)
	spawnTool := findControlTool(t, tools, "agent_spawn")

	result, err := spawnTool.call(context.Background(), map[string]interface{}{
		"instruction":   "plan work",
		"agent":         "planner",
		"allowed_tools": []interface{}{"exec"},
	})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, `tool "exec" outside role maximum scope for teammate type "planner"`, err.Error())
	assert.Empty(t, store.List())
}

func TestAgentSpawn_SpawnDepthPropagation(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	cp := &AgentControlPlane{
		RunStore:   store,
		Projection: NewAgentRunProjection(store),
	}
	tools := BuildControlTools(cp)
	spawnTool := findControlTool(t, tools, "agent_spawn")

	ctx := ctxkeys.WithSpawnDepth(context.Background(), 2)

	result, err := spawnTool.call(ctx, map[string]interface{}{
		"instruction": "deep task",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	run, err := store.Get(m["agent_id"].(string))
	require.NoError(t, err)
	assert.Equal(t, 3, run.SpawnDepth)
}

func TestAgentSpawn_SubmitsWithTeammateRuntimeContext(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	submitter := &recordingAgentRunSubmitter{}
	cp := &AgentControlPlane{
		RunStore:          store,
		Projection:        NewAgentRunProjection(store),
		Submitter:         submitter,
		CapabilityRuntime: NewCapabilityRuntime(store, &CapabilityPolicy{}, nil),
	}
	tools := BuildControlTools(cp)
	spawnTool := findControlTool(t, tools, "agent_spawn")

	ctx := session.WithSessionKey(context.Background(), "sess-parent")
	result, err := spawnTool.call(ctx, map[string]interface{}{
		"instruction":   "review the logs",
		"agent":         "operator",
		"allowed_tools": []interface{}{"fs_read"},
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.NotEmpty(t, m["agent_id"])
	assert.Equal(t, m["agent_id"], submitter.returnedID)
	assert.Equal(t, "sess-parent", submitter.origin.Session)
	assert.Equal(t, "agent_control", submitter.origin.Channel)
	assert.Contains(t, submitter.prompt, "review the logs")
	assert.Equal(t, "operator", ctxkeys.AgentNameFromContext(submitter.ctx))
	assert.Equal(t, []string{"fs_read"}, ctxkeys.DynamicAllowedToolsFromContext(submitter.ctx))
	assert.Equal(t, m["agent_id"], pendingAgentRunIDFromContext(submitter.ctx))
}

func TestAgentSpawn_SubmitterMismatchedReturnIDFails(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	submitter := &recordingAgentRunSubmitter{returnID: "wrong-id"}
	cp := &AgentControlPlane{
		RunStore:          store,
		Projection:        NewAgentRunProjection(store),
		Submitter:         submitter,
		CapabilityRuntime: NewCapabilityRuntime(store, &CapabilityPolicy{}, nil),
	}
	tools := BuildControlTools(cp)
	spawnTool := findControlTool(t, tools, "agent_spawn")

	result, err := spawnTool.call(context.Background(), map[string]interface{}{
		"instruction": "review the logs",
		"agent":       "operator",
	})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "mismatched task ID")

	runs := store.List()
	require.Len(t, runs, 1)
	assert.Equal(t, AgentRunFailed, runs[0].Status)
}

func TestAgentSpawn_MissingInstruction(t *testing.T) {
	cp := &AgentControlPlane{
		RunStore:   NewInMemoryAgentRunStore(),
		Projection: NewAgentRunProjection(NewInMemoryAgentRunStore()),
	}
	tools := BuildControlTools(cp)
	spawnTool := findControlTool(t, tools, "agent_spawn")

	_, err := spawnTool.call(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "instruction")
}

func TestAgentSpawn_NilProjection(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	cp := &AgentControlPlane{
		RunStore:   store,
		Projection: nil,
	}
	tools := BuildControlTools(cp)
	spawnTool := findControlTool(t, tools, "agent_spawn")

	result, err := spawnTool.call(context.Background(), map[string]interface{}{
		"instruction": "no projection",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.NotEmpty(t, m["agent_id"])
}

func TestAgentSpawn_SafetyLevel(t *testing.T) {
	cp := &AgentControlPlane{
		RunStore:   NewInMemoryAgentRunStore(),
		Projection: NewAgentRunProjection(NewInMemoryAgentRunStore()),
	}
	tools := BuildControlTools(cp)
	for _, tool := range tools {
		if tool.Name == "agent_spawn" {
			assert.Equal(t, agent.SafetyLevelModerate, tool.SafetyLevel)
			return
		}
	}
	t.Fatal("agent_spawn tool not found")
}

// --- agent_wait ---

func TestAgentWait_TerminalImmediate(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:     "wait-1",
		Status: AgentRunCompleted,
		Result: "done",
	}))

	cp := &AgentControlPlane{RunStore: store}
	tools := BuildControlTools(cp)
	waitTool := findControlTool(t, tools, "agent_wait")

	result, err := waitTool.call(context.Background(), map[string]interface{}{
		"agent_id": "wait-1",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, "wait-1", m["agent_id"])
	assert.Equal(t, "completed", m["status"])
	assert.Equal(t, "done", m["result"])
}

func TestAgentWait_TerminalFailed(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:     "wait-fail",
		Status: AgentRunFailed,
		Error:  "out of memory",
	}))

	cp := &AgentControlPlane{RunStore: store}
	tools := BuildControlTools(cp)
	waitTool := findControlTool(t, tools, "agent_wait")

	result, err := waitTool.call(context.Background(), map[string]interface{}{
		"agent_id": "wait-fail",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, "failed", m["status"])
	assert.Equal(t, "out of memory", m["error"])
}

func TestAgentWait_PollUntilComplete(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:     "wait-poll",
		Status: AgentRunRunning,
	}))

	cp := &AgentControlPlane{RunStore: store}
	tools := BuildControlTools(cp)
	waitTool := findControlTool(t, tools, "agent_wait")

	go func() {
		time.Sleep(600 * time.Millisecond)
		_ = store.UpdateStatus("wait-poll", AgentRunCompleted, "poll result", "")
	}()

	result, err := waitTool.call(context.Background(), map[string]interface{}{
		"agent_id": "wait-poll",
		"timeout":  float64(5),
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, "completed", m["status"])
	assert.Equal(t, "poll result", m["result"])
}

func TestAgentWait_Timeout(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:     "wait-timeout",
		Status: AgentRunRunning,
	}))

	cp := &AgentControlPlane{RunStore: store}
	tools := BuildControlTools(cp)
	waitTool := findControlTool(t, tools, "agent_wait")

	result, err := waitTool.call(context.Background(), map[string]interface{}{
		"agent_id": "wait-timeout",
		"timeout":  float64(1),
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, true, m["timeout"])
	assert.Equal(t, "running", m["status"])
}

func TestAgentWait_TimeoutIncludesBlockedProjectionWithoutCancelling(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:               "wait-blocked",
		Status:           AgentRunRunning,
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "capability request pending",
		GrantRequestID:   "grant-wait-blocked-fs_write",
	}))

	cp := &AgentControlPlane{RunStore: store}
	tools := BuildControlTools(cp)
	waitTool := findControlTool(t, tools, "agent_wait")

	result, err := waitTool.call(context.Background(), map[string]interface{}{
		"agent_id": "wait-blocked",
		"timeout":  float64(1),
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, true, m["timeout"])
	assert.Equal(t, "running", m["status"])
	assert.Equal(t, "blocked_waiting_approval", m["condition"])
	assert.Equal(t, "capability request pending", m["blocked_reason"])
	assert.Equal(t, "grant-wait-blocked-fs_write", m["grant_request_id"])

	run, err := store.Get("wait-blocked")
	require.NoError(t, err)
	assert.Equal(t, AgentRunRunning, run.Status)
}

func TestAgentWait_TerminalResponseOmitsStaleProjectionFields(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:               "wait-terminal-stale",
		Status:           AgentRunCompleted,
		Result:           "done",
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "stale block",
		GrantRequestID:   "grant-stale",
		WaitingOnRunID:   "run-stale",
		RecoveryState:    "resume_pending",
	}))

	cp := &AgentControlPlane{RunStore: store}
	tools := BuildControlTools(cp)
	waitTool := findControlTool(t, tools, "agent_wait")

	result, err := waitTool.call(context.Background(), map[string]interface{}{
		"agent_id": "wait-terminal-stale",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, "completed", m["status"])
	assert.Equal(t, "done", m["result"])
	assert.NotContains(t, m, "condition")
	assert.NotContains(t, m, "blocked_reason")
	assert.NotContains(t, m, "grant_request_id")
	assert.NotContains(t, m, "waiting_on_run_id")
	assert.NotContains(t, m, "recovery_state")
}

func TestAgentWait_ContextCancelled(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:     "wait-cancel",
		Status: AgentRunRunning,
	}))

	cp := &AgentControlPlane{RunStore: store}
	tools := BuildControlTools(cp)
	waitTool := findControlTool(t, tools, "agent_wait")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(300 * time.Millisecond)
		cancel()
	}()

	_, err := waitTool.call(ctx, map[string]interface{}{
		"agent_id": "wait-cancel",
		"timeout":  float64(30),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestAgentWait_NotFound(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	cp := &AgentControlPlane{RunStore: store}
	tools := BuildControlTools(cp)
	waitTool := findControlTool(t, tools, "agent_wait")

	_, err := waitTool.call(context.Background(), map[string]interface{}{
		"agent_id": "nonexistent",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAgentWait_MissingAgentID(t *testing.T) {
	cp := &AgentControlPlane{RunStore: NewInMemoryAgentRunStore()}
	tools := BuildControlTools(cp)
	waitTool := findControlTool(t, tools, "agent_wait")

	_, err := waitTool.call(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent_id")
}

func TestAgentWait_SafetyLevel(t *testing.T) {
	cp := &AgentControlPlane{RunStore: NewInMemoryAgentRunStore()}
	tools := BuildControlTools(cp)
	for _, tool := range tools {
		if tool.Name == "agent_wait" {
			assert.Equal(t, agent.SafetyLevelSafe, tool.SafetyLevel)
			return
		}
	}
	t.Fatal("agent_wait tool not found")
}

// --- agent_stop ---

func TestAgentStop_Basic(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:     "stop-1",
		Status: AgentRunRunning,
	}))

	cp := &AgentControlPlane{RunStore: store}
	tools := BuildControlTools(cp)
	stopTool := findControlTool(t, tools, "agent_stop")

	result, err := stopTool.call(context.Background(), map[string]interface{}{
		"agent_id": "stop-1",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, "stop-1", m["agent_id"])
	assert.Equal(t, "cancelled", m["status"])

	run, err := store.Get("stop-1")
	require.NoError(t, err)
	assert.Equal(t, AgentRunCancelled, run.Status)
}

func TestAgentStop_NotFound(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	cp := &AgentControlPlane{RunStore: store}
	tools := BuildControlTools(cp)
	stopTool := findControlTool(t, tools, "agent_stop")

	_, err := stopTool.call(context.Background(), map[string]interface{}{
		"agent_id": "nonexistent",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAgentStop_AlreadyTerminal(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:     "stop-term",
		Status: AgentRunCompleted,
	}))

	cp := &AgentControlPlane{RunStore: store}
	tools := BuildControlTools(cp)
	stopTool := findControlTool(t, tools, "agent_stop")

	_, err := stopTool.call(context.Background(), map[string]interface{}{
		"agent_id": "stop-term",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already")
}

func TestAgentStop_MissingAgentID(t *testing.T) {
	cp := &AgentControlPlane{RunStore: NewInMemoryAgentRunStore()}
	tools := BuildControlTools(cp)
	stopTool := findControlTool(t, tools, "agent_stop")

	_, err := stopTool.call(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent_id")
}

func TestAgentStop_SafetyLevel(t *testing.T) {
	cp := &AgentControlPlane{RunStore: NewInMemoryAgentRunStore()}
	tools := BuildControlTools(cp)
	for _, tool := range tools {
		if tool.Name == "agent_stop" {
			assert.Equal(t, agent.SafetyLevelSafe, tool.SafetyLevel)
			return
		}
	}
	t.Fatal("agent_stop tool not found")
}

// --- generateAgentRunID ---

func TestGenerateAgentRunID(t *testing.T) {
	id, err := generateAgentRunID()
	require.NoError(t, err)
	assert.Contains(t, id, "arun-")
	assert.Len(t, id, 21)

	id2, err := generateAgentRunID()
	require.NoError(t, err)
	assert.NotEqual(t, id, id2)
}

// --- Helpers ---

type recordingAgentRunSubmitter struct {
	ctx        context.Context
	prompt     string
	origin     background.Origin
	returnID   string
	returnedID string
	err        error
}

func (r *recordingAgentRunSubmitter) Submit(ctx context.Context, prompt string, origin background.Origin) (string, error) {
	r.ctx = ctx
	r.prompt = prompt
	r.origin = origin
	if r.err != nil {
		return "", r.err
	}
	if r.returnID != "" {
		r.returnedID = r.returnID
		return r.returnID, nil
	}
	r.returnedID = pendingAgentRunIDFromContext(ctx)
	return r.returnedID, nil
}

type controlToolHelper struct {
	tool *agent.Tool
}

func (h *controlToolHelper) call(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return h.tool.Handler(ctx, params)
}

func findControlTool(t *testing.T, tools []*agent.Tool, name string) *controlToolHelper {
	t.Helper()
	for _, tool := range tools {
		if tool.Name == name {
			return &controlToolHelper{tool}
		}
	}
	t.Fatalf("tool %q not found", name)
	return nil
}
