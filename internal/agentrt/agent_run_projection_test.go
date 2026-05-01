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

	ctx := withPendingAgentRunID(context.Background(), "run-42")

	id, err := proj.PrepareTask(ctx, "do something", background.Origin{
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

func TestAgentRunProjection_PrepareTaskMultiplePendingContextsAreDeterministic(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	proj := NewAgentRunProjection(store)

	ctxA := withPendingAgentRunID(context.Background(), "run-a")
	ctxB := withPendingAgentRunID(context.Background(), "run-b")

	id1, err := proj.PrepareTask(ctxA, "p1", background.Origin{})
	require.NoError(t, err)
	assert.Equal(t, "run-a", id1)

	id2, err := proj.PrepareTask(ctxB, "p2", background.Origin{})
	require.NoError(t, err)
	assert.Equal(t, "run-b", id2)
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

	// 2. Context carries the canonical pending ID for deterministic unification.
	ctx := withPendingAgentRunID(context.Background(), "lifecycle-1")

	// 3. PrepareTask returns the unified ID.
	id, err := proj.PrepareTask(ctx, "run analysis", background.Origin{
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

func TestBackgroundProjection_PrepareTaskUsesSameIDForAgentRunAndRunLedger(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	agentProjection := NewAgentRunProjection(store)
	runLedgerProjection := &recordingBackgroundWriteThrough{}
	projection := NewBackgroundProjection(agentProjection, runLedgerProjection)

	ctx := withPendingAgentRunID(context.Background(), "agent-run-1")

	id, err := projection.PrepareTask(ctx, "do something", background.Origin{
		Channel: "agent_control",
		Session: "sess-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "agent-run-1", id)
	assert.Equal(t, "agent-run-1", runLedgerProjection.preparedID)
	assert.Equal(t, "do something", runLedgerProjection.prompt)
	assert.Equal(t, "sess-1", runLedgerProjection.origin.Session)
}

type recordingBackgroundWriteThrough struct {
	preparedID string
	prompt     string
	origin     background.Origin
	syncSnap   background.TaskSnapshot
}

func (r *recordingBackgroundWriteThrough) PrepareTask(
	_ context.Context,
	_ string,
	_ background.Origin,
) (string, error) {
	return "fallback-run", nil
}

func (r *recordingBackgroundWriteThrough) PrepareTaskWithID(
	_ context.Context,
	prompt string,
	origin background.Origin,
	runID string,
) error {
	r.preparedID = runID
	r.prompt = prompt
	r.origin = origin
	return nil
}

func (r *recordingBackgroundWriteThrough) SyncTask(_ context.Context, snap background.TaskSnapshot) error {
	r.syncSnap = snap
	return nil
}

var _ interface {
	PrepareTask(context.Context, string, background.Origin) (string, error)
	PrepareTaskWithID(context.Context, string, background.Origin, string) error
	SyncTask(context.Context, background.TaskSnapshot) error
} = (*recordingBackgroundWriteThrough)(nil)
