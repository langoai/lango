package agentrt

import (
	"context"
	"testing"
	"time"

	"github.com/langoai/lango/internal/background"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentRunProjection_InterfaceSatisfaction(t *testing.T) {
	// Compile-time check is in the production file; this confirms at test time.
	var _ background.Projection = (*AgentRunProjection)(nil)
}

func TestAgentRunProjection_PrepareTaskReturnsPendingID(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	proj := NewAgentRunProjection(store)

	proj.RegisterPending("run-42")

	id, err := proj.PrepareTask(context.Background(), "do something", background.Origin{
		Channel: "test",
		Session: "sess-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "run-42", id)
}

func TestAgentRunProjection_PrepareTaskNoPending(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	proj := NewAgentRunProjection(store)

	_, err := proj.PrepareTask(context.Background(), "prompt", background.Origin{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no pending agent run ID")
}

func TestAgentRunProjection_PrepareTaskConsumesOnce(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	proj := NewAgentRunProjection(store)

	proj.RegisterPending("run-once")

	id, err := proj.PrepareTask(context.Background(), "p1", background.Origin{})
	require.NoError(t, err)
	assert.Equal(t, "run-once", id)

	// Second call should fail — the pending ID was consumed.
	_, err = proj.PrepareTask(context.Background(), "p2", background.Origin{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no pending agent run ID")
}

func TestAgentRunProjection_PrepareTaskMultiplePending(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	proj := NewAgentRunProjection(store)

	proj.RegisterPending("run-a")
	proj.RegisterPending("run-b")

	// Both should be consumed (order is non-deterministic due to map iteration).
	id1, err := proj.PrepareTask(context.Background(), "p1", background.Origin{})
	require.NoError(t, err)

	id2, err := proj.PrepareTask(context.Background(), "p2", background.Origin{})
	require.NoError(t, err)

	assert.NotEqual(t, id1, id2)
	ids := map[string]bool{id1: true, id2: true}
	assert.True(t, ids["run-a"])
	assert.True(t, ids["run-b"])

	// Third call should fail.
	_, err = proj.PrepareTask(context.Background(), "p3", background.Origin{})
	require.Error(t, err)
}

func TestAgentRunProjection_SyncTaskStatusMapping(t *testing.T) {
	tests := []struct {
		give       string
		giveBgStat background.Status
		giveResult string
		giveErr    string
		wantStatus AgentRunStatus
		wantResult string
		wantErr    string
	}{
		{
			give:       "pending maps to spawned",
			giveBgStat: background.Pending,
			wantStatus: AgentRunSpawned,
		},
		{
			give:       "running maps to running",
			giveBgStat: background.Running,
			wantStatus: AgentRunRunning,
		},
		{
			give:       "done maps to completed",
			giveBgStat: background.Done,
			giveResult: "task output",
			wantStatus: AgentRunCompleted,
			wantResult: "task output",
		},
		{
			give:       "failed maps to failed",
			giveBgStat: background.Failed,
			giveErr:    "timeout exceeded",
			wantStatus: AgentRunFailed,
			wantErr:    "timeout exceeded",
		},
		{
			give:       "cancelled maps to cancelled",
			giveBgStat: background.Cancelled,
			wantStatus: AgentRunCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			store := NewInMemoryAgentRunStore()
			proj := NewAgentRunProjection(store)

			// Pre-create the agent run in Spawned status.
			runID := "sync-" + tt.give
			require.NoError(t, store.Create(&AgentRun{
				ID:        runID,
				Status:    AgentRunSpawned,
				CreatedAt: time.Now(),
			}))

			snap := background.TaskSnapshot{
				ID:     runID,
				Status: tt.giveBgStat,
				Result: tt.giveResult,
				Error:  tt.giveErr,
			}

			err := proj.SyncTask(context.Background(), snap)
			require.NoError(t, err)

			got, err := store.Get(runID)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, got.Status)
			assert.Equal(t, tt.wantResult, got.Result)
			assert.Equal(t, tt.wantErr, got.Error)
		})
	}
}

func TestAgentRunProjection_SyncTaskUnknownStatus(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	proj := NewAgentRunProjection(store)

	snap := background.TaskSnapshot{
		ID:     "bad-status",
		Status: background.Status(999),
	}

	err := proj.SyncTask(context.Background(), snap)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown status")
}

func TestAgentRunProjection_SyncTaskStoreError(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	proj := NewAgentRunProjection(store)

	// SyncTask on a non-existent run should propagate the store error.
	snap := background.TaskSnapshot{
		ID:     "nonexistent",
		Status: background.Running,
	}

	err := proj.SyncTask(context.Background(), snap)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAgentRunProjection_FullLifecycle(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	proj := NewAgentRunProjection(store)

	// 1. Create agent run in store (simulating D2 spawn).
	require.NoError(t, store.Create(&AgentRun{
		ID:          "lifecycle-1",
		ParentID:    "parent-sess",
		Instruction: "run analysis",
		Status:      AgentRunSpawned,
		CreatedAt:   time.Now(),
	}))

	// 2. Register pending for ID unification.
	proj.RegisterPending("lifecycle-1")

	// 3. PrepareTask returns the unified ID.
	id, err := proj.PrepareTask(context.Background(), "run analysis", background.Origin{
		Channel: "test",
		Session: "parent-sess",
	})
	require.NoError(t, err)
	assert.Equal(t, "lifecycle-1", id)

	// 4. SyncTask: Pending (bgManager created the task).
	require.NoError(t, proj.SyncTask(context.Background(), background.TaskSnapshot{
		ID:     id,
		Status: background.Pending,
	}))
	got, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, AgentRunSpawned, got.Status)

	// 5. SyncTask: Running.
	require.NoError(t, proj.SyncTask(context.Background(), background.TaskSnapshot{
		ID:     id,
		Status: background.Running,
	}))
	got, err = store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, AgentRunRunning, got.Status)

	// 6. SyncTask: Done with result.
	require.NoError(t, proj.SyncTask(context.Background(), background.TaskSnapshot{
		ID:     id,
		Status: background.Done,
		Result: "analysis complete",
	}))
	got, err = store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, AgentRunCompleted, got.Status)
	assert.Equal(t, "analysis complete", got.Result)
	assert.False(t, got.CompletedAt.IsZero())
}

func TestMapBgStatus(t *testing.T) {
	tests := []struct {
		give background.Status
		want AgentRunStatus
	}{
		{give: background.Pending, want: AgentRunSpawned},
		{give: background.Running, want: AgentRunRunning},
		{give: background.Done, want: AgentRunCompleted},
		{give: background.Failed, want: AgentRunFailed},
		{give: background.Cancelled, want: AgentRunCancelled},
	}

	for _, tt := range tests {
		t.Run(string(tt.want), func(t *testing.T) {
			got, err := mapBgStatus(tt.give)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
