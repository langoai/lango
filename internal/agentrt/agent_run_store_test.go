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
