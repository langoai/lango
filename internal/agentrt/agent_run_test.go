package agentrt

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentRunStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		give     AgentRunStatus
		wantTerm bool
	}{
		{give: AgentRunSpawned, wantTerm: false},
		{give: AgentRunRunning, wantTerm: false},
		{give: AgentRunCompleted, wantTerm: true},
		{give: AgentRunFailed, wantTerm: true},
		{give: AgentRunCancelled, wantTerm: true},
	}

	for _, tt := range tests {
		t.Run(string(tt.give), func(t *testing.T) {
			assert.Equal(t, tt.wantTerm, tt.give.isTerminal())
		})
	}
}

func TestInMemoryAgentRunStore_CreateAndGet(t *testing.T) {
	store := NewInMemoryAgentRunStore()

	run := &AgentRun{
		ID:             "run-1",
		ParentID:       "session-0",
		RequestedAgent: "operator",
		Instruction:    "check status",
		Status:         AgentRunSpawned,
		AllowedTools:   []string{"tool_a", "tool_b"},
		CreatedAt:      time.Now(),
	}

	require.NoError(t, store.Create(run))

	got, err := store.Get("run-1")
	require.NoError(t, err)
	assert.Equal(t, "run-1", got.ID)
	assert.Equal(t, "session-0", got.ParentID)
	assert.Equal(t, "operator", got.RequestedAgent)
	assert.Equal(t, "check status", got.Instruction)
	assert.Equal(t, AgentRunSpawned, got.Status)
	assert.Equal(t, []string{"tool_a", "tool_b"}, got.AllowedTools)
	assert.Nil(t, got.CancelFn, "CancelFn must not leak through Get")
}

func TestInMemoryAgentRunStore_CreateDuplicateID(t *testing.T) {
	store := NewInMemoryAgentRunStore()

	run := &AgentRun{ID: "dup-1", Status: AgentRunSpawned, CreatedAt: time.Now()}
	require.NoError(t, store.Create(run))

	err := store.Create(&AgentRun{ID: "dup-1", Status: AgentRunSpawned})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestInMemoryAgentRunStore_CreateNil(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	err := store.Create(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil run")
}

func TestInMemoryAgentRunStore_GetNotFound(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	_, err := store.Get("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestInMemoryAgentRunStore_GetReturnsCopy(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	run := &AgentRun{
		ID:           "copy-1",
		Status:       AgentRunRunning,
		AllowedTools: []string{"tool_x"},
		CreatedAt:    time.Now(),
	}
	require.NoError(t, store.Create(run))

	got, err := store.Get("copy-1")
	require.NoError(t, err)

	// Mutate the returned copy.
	got.Status = AgentRunCompleted
	got.AllowedTools[0] = "mutated"

	// Verify the internal state is unaffected.
	original, err := store.Get("copy-1")
	require.NoError(t, err)
	assert.Equal(t, AgentRunRunning, original.Status)
	assert.Equal(t, []string{"tool_x"}, original.AllowedTools)
}

func TestInMemoryAgentRunStore_List(t *testing.T) {
	store := NewInMemoryAgentRunStore()

	// Empty store returns empty slice.
	assert.Empty(t, store.List())

	require.NoError(t, store.Create(&AgentRun{ID: "a", Status: AgentRunSpawned}))
	require.NoError(t, store.Create(&AgentRun{ID: "b", Status: AgentRunRunning}))
	require.NoError(t, store.Create(&AgentRun{ID: "c", Status: AgentRunCompleted}))

	runs := store.List()
	assert.Len(t, runs, 3)

	ids := make(map[string]bool, 3)
	for _, r := range runs {
		ids[r.ID] = true
		assert.Nil(t, r.CancelFn, "CancelFn must not leak through List")
	}
	assert.True(t, ids["a"])
	assert.True(t, ids["b"])
	assert.True(t, ids["c"])
}

func TestInMemoryAgentRunStore_ListReturnsCopies(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{ID: "list-cp", Status: AgentRunRunning}))

	runs := store.List()
	require.Len(t, runs, 1)
	runs[0].Status = AgentRunFailed

	got, err := store.Get("list-cp")
	require.NoError(t, err)
	assert.Equal(t, AgentRunRunning, got.Status)
}

func TestInMemoryAgentRunStore_UpdateStatus(t *testing.T) {
	tests := []struct {
		give       string
		giveInit   AgentRunStatus
		giveUpdate AgentRunStatus
		giveResult string
		giveErr    string
		wantErr    bool
	}{
		{
			give:       "spawned to running",
			giveInit:   AgentRunSpawned,
			giveUpdate: AgentRunRunning,
		},
		{
			give:       "running to completed",
			giveInit:   AgentRunRunning,
			giveUpdate: AgentRunCompleted,
			giveResult: "done",
		},
		{
			give:       "running to failed",
			giveInit:   AgentRunRunning,
			giveUpdate: AgentRunFailed,
			giveErr:    "timeout",
		},
		{
			give:       "spawned to cancelled",
			giveInit:   AgentRunSpawned,
			giveUpdate: AgentRunCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			store := NewInMemoryAgentRunStore()
			require.NoError(t, store.Create(&AgentRun{
				ID:     "u-" + tt.give,
				Status: tt.giveInit,
			}))

			err := store.UpdateStatus("u-"+tt.give, tt.giveUpdate, tt.giveResult, tt.giveErr)
			require.NoError(t, err)

			got, err := store.Get("u-" + tt.give)
			require.NoError(t, err)
			assert.Equal(t, tt.giveUpdate, got.Status)
			assert.Equal(t, tt.giveResult, got.Result)
			assert.Equal(t, tt.giveErr, got.Error)

			if tt.giveUpdate.isTerminal() {
				assert.False(t, got.CompletedAt.IsZero(), "CompletedAt should be set for terminal status")
			}
		})
	}
}

func TestInMemoryAgentRunStore_UpdateStatusNotFound(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	err := store.UpdateStatus("ghost", AgentRunRunning, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestInMemoryAgentRunStore_UpdateStatusTerminalGuard(t *testing.T) {
	tests := []struct {
		give AgentRunStatus
	}{
		{give: AgentRunCompleted},
		{give: AgentRunFailed},
		{give: AgentRunCancelled},
	}

	for _, tt := range tests {
		t.Run(string(tt.give), func(t *testing.T) {
			store := NewInMemoryAgentRunStore()
			require.NoError(t, store.Create(&AgentRun{
				ID:     "term-" + string(tt.give),
				Status: tt.give,
			}))

			err := store.UpdateStatus("term-"+string(tt.give), AgentRunRunning, "", "")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "already")
		})
	}
}

func TestInMemoryAgentRunStore_Cancel(t *testing.T) {
	store := NewInMemoryAgentRunStore()

	var cancelled int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	run := &AgentRun{
		ID:       "cancel-1",
		Status:   AgentRunRunning,
		CancelFn: func() { atomic.AddInt32(&cancelled, 1); cancel() },
	}
	require.NoError(t, store.Create(run))

	err := store.Cancel("cancel-1")
	require.NoError(t, err)

	// Verify CancelFn was called.
	assert.Equal(t, int32(1), atomic.LoadInt32(&cancelled))

	// Verify context was actually cancelled.
	assert.Error(t, ctx.Err())

	// Verify status.
	got, err := store.Get("cancel-1")
	require.NoError(t, err)
	assert.Equal(t, AgentRunCancelled, got.Status)
	assert.False(t, got.CompletedAt.IsZero())
}

func TestInMemoryAgentRunStore_CancelNilCancelFn(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:     "cancel-nil",
		Status: AgentRunSpawned,
	}))

	err := store.Cancel("cancel-nil")
	require.NoError(t, err)

	got, err := store.Get("cancel-nil")
	require.NoError(t, err)
	assert.Equal(t, AgentRunCancelled, got.Status)
}

func TestInMemoryAgentRunStore_CancelNotFound(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	err := store.Cancel("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestInMemoryAgentRunStore_CancelTerminalGuard(t *testing.T) {
	tests := []struct {
		give AgentRunStatus
	}{
		{give: AgentRunCompleted},
		{give: AgentRunFailed},
		{give: AgentRunCancelled},
	}

	for _, tt := range tests {
		t.Run(string(tt.give), func(t *testing.T) {
			store := NewInMemoryAgentRunStore()
			require.NoError(t, store.Create(&AgentRun{
				ID:     "cterm-" + string(tt.give),
				Status: tt.give,
			}))

			err := store.Cancel("cterm-" + string(tt.give))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "already")
		})
	}
}
