package agentrt

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
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

	// Verify the run was persisted.
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

func TestAgentSpawn_SpawnDepthPropagation(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	cp := &AgentControlPlane{
		RunStore:   store,
		Projection: NewAgentRunProjection(store),
	}
	tools := BuildControlTools(cp)
	spawnTool := findControlTool(t, tools, "agent_spawn")

	// Simulate a parent at depth 2.
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

	// Complete the run after a short delay.
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

	// Verify store state.
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
	// "arun-" (5 chars) + 16 hex chars = 21 total.
	assert.Len(t, id, 21)

	// IDs should be unique.
	id2, err := generateAgentRunID()
	require.NoError(t, err)
	assert.NotEqual(t, id, id2)
}

// --- Helpers ---

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
