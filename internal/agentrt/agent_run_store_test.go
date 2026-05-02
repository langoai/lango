package agentrt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentRunStore_CopyIncludesProjectionFields(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:               "copy-projection",
		Status:           AgentRunRunning,
		AllowedTools:     []string{"tool_a", "tool_b"},
		SpawnReason:      "delegated_for_approval",
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "awaiting human approval",
		GrantRequestID:   "grant-123",
		WaitingOnRunID:   "run-parent",
		RecoveryState:    "replay_pending",
	}))

	got, err := store.Get("copy-projection")
	require.NoError(t, err)

	assert.Equal(t, "delegated_for_approval", got.SpawnReason)
	assert.Equal(t, AgentRunConditionBlockedWaitingApproval, got.RuntimeCondition)
	assert.Equal(t, "awaiting human approval", got.BlockedReason)
	assert.Equal(t, "grant-123", got.GrantRequestID)
	assert.Equal(t, "run-parent", got.WaitingOnRunID)
	assert.Equal(t, "replay_pending", got.RecoveryState)
	assert.Equal(t, []string{"tool_a", "tool_b"}, got.AllowedTools)

	got.AllowedTools[0] = "mutated"

	fresh, err := store.Get("copy-projection")
	require.NoError(t, err)
	assert.Equal(t, []string{"tool_a", "tool_b"}, fresh.AllowedTools)
}

func TestAgentRunStore_CreateSnapshotsCallerOwnedRun(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	original := &AgentRun{
		ID:             "create-snapshot",
		Status:         AgentRunRunning,
		GrantRequestID: "grant-original",
		GrantAttempt:   1,
		GrantState:     "pending",
		AllowedTools:   []string{"fs_read"},
	}

	require.NoError(t, store.Create(original))

	original.GrantRequestID = "grant-mutated"
	original.GrantAttempt = 99
	original.GrantState = "mutated"
	original.AllowedTools[0] = "exec"

	got, err := store.Get("create-snapshot")
	require.NoError(t, err)
	assert.Equal(t, "grant-original", got.GrantRequestID)
	assert.Equal(t, 1, got.GrantAttempt)
	assert.Equal(t, "pending", got.GrantState)
	assert.Equal(t, []string{"fs_read"}, got.AllowedTools)
}

func TestAgentRunStore_UpdateProjectionSetsProjectedStateOnNonTerminalRun(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:     "projection-update",
		Status: AgentRunRunning,
	}))

	err := store.UpdateProjection("projection-update", RunProjectionPatch{
		ApplyRuntimeCondition: true,
		ApplyBlockedReason:    true,
		ApplyGrantRequestID:   true,
		ApplyWaitingOnRunID:   true,
		ApplyRecoveryState:    true,
		RuntimeCondition:      AgentRunConditionBlockedWaitingApproval,
		BlockedReason:         "awaiting approval",
		GrantRequestID:        "grant-456",
		WaitingOnRunID:        "run-789",
		RecoveryState:         "recovery-started",
	})
	require.NoError(t, err)

	got, err := store.Get("projection-update")
	require.NoError(t, err)
	assert.Equal(t, AgentRunConditionBlockedWaitingApproval, got.RuntimeCondition)
	assert.Equal(t, "awaiting approval", got.BlockedReason)
	assert.Equal(t, "grant-456", got.GrantRequestID)
	assert.Equal(t, "run-789", got.WaitingOnRunID)
	assert.Equal(t, "recovery-started", got.RecoveryState)
}

func TestInMemoryAgentRunStore_UpdateProjection_StoresGrantAttemptMetadata(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:     "run-1",
		Status: AgentRunRunning,
	}))

	err := store.UpdateProjection("run-1", RunProjectionPatch{
		ApplyGrantAttempt: true,
		ApplyGrantState:   true,
		GrantAttempt:      2,
		GrantState:        "pending",
	})
	require.NoError(t, err)

	run, err := store.Get("run-1")
	require.NoError(t, err)
	assert.Equal(t, 2, run.GrantAttempt)
	assert.Equal(t, "pending", run.GrantState)
}

func TestInMemoryAgentRunStore_UpdateProjection_ClearsGrantAttemptMetadata(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:           "run-2",
		Status:       AgentRunRunning,
		GrantAttempt: 3,
		GrantState:   "pending",
	}))

	err := store.UpdateProjection("run-2", RunProjectionPatch{
		ApplyGrantAttempt: true,
		ApplyGrantState:   true,
		GrantAttempt:      0,
		GrantState:        "",
	})
	require.NoError(t, err)

	run, err := store.Get("run-2")
	require.NoError(t, err)
	assert.Equal(t, 0, run.GrantAttempt)
	assert.Empty(t, run.GrantState)
}

func TestAgentRunStore_UpdateProjectionAddsAllowedToolExactlyOnce(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:           "projection-tool",
		Status:       AgentRunSpawned,
		AllowedTools: []string{"tool_a"},
	}))

	err := store.UpdateProjection("projection-tool", RunProjectionPatch{
		AddAllowedTool: "tool_b",
	})
	require.NoError(t, err)

	err = store.UpdateProjection("projection-tool", RunProjectionPatch{
		AddAllowedTool: "tool_b",
	})
	require.NoError(t, err)

	got, err := store.Get("projection-tool")
	require.NoError(t, err)
	assert.Equal(t, []string{"tool_a", "tool_b"}, got.AllowedTools)
}

func TestAgentRunStore_UpdateProjectionRejectsTerminalRun(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:     "projection-terminal",
		Status: AgentRunCompleted,
	}))

	err := store.UpdateProjection("projection-terminal", RunProjectionPatch{
		ApplyRuntimeCondition: true,
		RuntimeCondition:      AgentRunConditionBlockedWaitingApproval,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already")
}

func TestAgentRunStore_UpdateProjectionPreservesFieldsNotBeingUpdated(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:               "projection-partial",
		Status:           AgentRunRunning,
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "awaiting approval",
		GrantRequestID:   "grant-existing",
		WaitingOnRunID:   "run-existing",
		RecoveryState:    "recovery-existing",
	}))

	err := store.UpdateProjection("projection-partial", RunProjectionPatch{
		ApplyBlockedReason: true,
		BlockedReason:      "updated-reason",
		AddAllowedTool:     "exec",
	})
	require.NoError(t, err)

	got, err := store.Get("projection-partial")
	require.NoError(t, err)
	assert.Equal(t, AgentRunConditionBlockedWaitingApproval, got.RuntimeCondition)
	assert.Equal(t, "updated-reason", got.BlockedReason)
	assert.Equal(t, "grant-existing", got.GrantRequestID)
	assert.Equal(t, "run-existing", got.WaitingOnRunID)
	assert.Equal(t, "recovery-existing", got.RecoveryState)
	assert.Equal(t, []string{"exec"}, got.AllowedTools)
}

func TestAgentRunStore_UpdateProjectionNotFound(t *testing.T) {
	store := NewInMemoryAgentRunStore()

	err := store.UpdateProjection("missing-run", RunProjectionPatch{
		ApplyBlockedReason: true,
		BlockedReason:      "irrelevant",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
