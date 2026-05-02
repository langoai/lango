package agentrt

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/runledger"
)

type failingRunLedgerStore struct {
	runledger.RunLedgerStore
	appendErr error
}

func (s *failingRunLedgerStore) AppendJournalEvent(ctx context.Context, event runledger.JournalEvent) error {
	if s.appendErr != nil {
		return s.appendErr
	}
	return s.RunLedgerStore.AppendJournalEvent(ctx, event)
}

func TestRunLedgerMirrorStore_ApprovalBlockedAppendsJournalEvent(t *testing.T) {
	ctx := context.Background()
	base := NewInMemoryAgentRunStore()
	ledger := runledger.NewMemoryStore()

	require.NoError(t, base.Create(&AgentRun{
		ID:     "arun-1",
		Status: AgentRunRunning,
	}))

	store := NewRunLedgerMirrorStore(base, ledger, nil)
	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "arun-1",
		Type:    runledger.EventRunCreated,
		Payload: json.RawMessage(`{"goal":"teammate"}`),
	}))

	err := store.UpdateProjection("arun-1", RunProjectionPatch{
		ApplyRuntimeCondition: true,
		ApplyBlockedReason:    true,
		ApplyGrantRequestID:   true,
		RuntimeCondition:      AgentRunConditionBlockedWaitingApproval,
		BlockedReason:         "dangerous tool requires approval",
		GrantRequestID:        "grant-arun-1-exec",
	})
	require.NoError(t, err)

	events, err := ledger.GetJournalEvents(ctx, "arun-1")
	require.NoError(t, err)
	assert.Equal(t, runledger.EventTeammateApprovalBlocked, events[len(events)-1].Type)

	snap, _, err := ledger.GetCachedSnapshot(ctx, "arun-1")
	require.NoError(t, err)
	require.NotNil(t, snap)
	assert.Equal(t, "blocked_waiting_approval", snap.TeammateRuntimeCondition)
	assert.Equal(t, "dangerous tool requires approval", snap.TeammateBlockedReason)
	assert.Equal(t, "grant-arun-1-exec", snap.TeammateGrantRequestID)
}

func TestRunLedgerMirrorStore_ApprovalUnblockedAppendsJournalEvent(t *testing.T) {
	ctx := context.Background()
	base := NewInMemoryAgentRunStore()
	ledger := runledger.NewMemoryStore()

	require.NoError(t, base.Create(&AgentRun{
		ID:               "arun-2",
		Status:           AgentRunRunning,
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "dangerous tool requires approval",
		GrantRequestID:   "grant-arun-2-exec",
	}))

	store := NewRunLedgerMirrorStore(base, ledger, nil)
	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "arun-2",
		Type:    runledger.EventRunCreated,
		Payload: json.RawMessage(`{"goal":"teammate"}`),
	}))

	err := store.UpdateProjection("arun-2", RunProjectionPatch{
		ApplyRuntimeCondition: true,
		ApplyBlockedReason:    true,
		ApplyGrantRequestID:   true,
		RuntimeCondition:      AgentRunConditionNone,
		BlockedReason:         "",
		GrantRequestID:        "",
	})
	require.NoError(t, err)

	events, err := ledger.GetJournalEvents(ctx, "arun-2")
	require.NoError(t, err)
	assert.Equal(t, runledger.EventTeammateApprovalUnblocked, events[len(events)-1].Type)

	snap, _, err := ledger.GetCachedSnapshot(ctx, "arun-2")
	require.NoError(t, err)
	require.NotNil(t, snap)
	assert.Empty(t, snap.TeammateRuntimeCondition)
	assert.Empty(t, snap.TeammateBlockedReason)
	assert.Empty(t, snap.TeammateGrantRequestID)
}

func TestRunLedgerMirrorStore_ApprovalBlockedReplaceAppendsJournalEventAndRefreshesSnapshot(t *testing.T) {
	ctx := context.Background()
	base := NewInMemoryAgentRunStore()
	ledger := runledger.NewMemoryStore()

	require.NoError(t, base.Create(&AgentRun{
		ID:               "arun-3",
		Status:           AgentRunRunning,
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "initial reason",
		GrantRequestID:   "grant-arun-3-old",
	}))

	store := NewRunLedgerMirrorStore(base, ledger, nil)
	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "arun-3",
		Type:    runledger.EventRunCreated,
		Payload: json.RawMessage(`{"goal":"teammate"}`),
	}))
	_, err := ledger.GetRunSnapshot(ctx, "arun-3")
	require.NoError(t, err)

	err = store.UpdateProjection("arun-3", RunProjectionPatch{
		ApplyRuntimeCondition: true,
		ApplyBlockedReason:    true,
		ApplyGrantRequestID:   true,
		RuntimeCondition:      AgentRunConditionBlockedWaitingApproval,
		BlockedReason:         "replacement reason",
		GrantRequestID:        "grant-arun-3-new",
	})
	require.NoError(t, err)

	events, err := ledger.GetJournalEvents(ctx, "arun-3")
	require.NoError(t, err)
	assert.Equal(t, runledger.EventTeammateApprovalBlocked, events[len(events)-1].Type)

	snap, _, err := ledger.GetCachedSnapshot(ctx, "arun-3")
	require.NoError(t, err)
	require.NotNil(t, snap)
	assert.Equal(t, "blocked_waiting_approval", snap.TeammateRuntimeCondition)
	assert.Equal(t, "replacement reason", snap.TeammateBlockedReason)
	assert.Equal(t, "grant-arun-3-new", snap.TeammateGrantRequestID)
}

func TestRunLedgerMirrorStore_MirrorFailureDoesNotFailProjection(t *testing.T) {
	base := NewInMemoryAgentRunStore()
	ledger := &failingRunLedgerStore{
		RunLedgerStore: runledger.NewMemoryStore(),
		appendErr:      errors.New("boom"),
	}

	require.NoError(t, base.Create(&AgentRun{
		ID:     "arun-4",
		Status: AgentRunRunning,
	}))

	store := NewRunLedgerMirrorStore(base, ledger, nil)
	err := store.UpdateProjection("arun-4", RunProjectionPatch{
		ApplyRuntimeCondition: true,
		ApplyBlockedReason:    true,
		ApplyGrantRequestID:   true,
		RuntimeCondition:      AgentRunConditionBlockedWaitingApproval,
		BlockedReason:         "dangerous tool requires approval",
		GrantRequestID:        "grant-arun-4-exec",
	})
	require.NoError(t, err)

	run, err := base.Get("arun-4")
	require.NoError(t, err)
	assert.Equal(t, AgentRunConditionBlockedWaitingApproval, run.RuntimeCondition)
	assert.Equal(t, "dangerous tool requires approval", run.BlockedReason)
	assert.Equal(t, "grant-arun-4-exec", run.GrantRequestID)
}

func TestRunLedgerMirrorStore_ClearToClearDoesNotAppendJournalEvent(t *testing.T) {
	ctx := context.Background()
	base := NewInMemoryAgentRunStore()
	ledger := runledger.NewMemoryStore()

	require.NoError(t, base.Create(&AgentRun{
		ID:     "arun-5",
		Status: AgentRunRunning,
	}))

	store := NewRunLedgerMirrorStore(base, ledger, nil)
	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "arun-5",
		Type:    runledger.EventRunCreated,
		Payload: json.RawMessage(`{"goal":"teammate"}`),
	}))

	err := store.UpdateProjection("arun-5", RunProjectionPatch{
		ApplyRuntimeCondition: true,
		ApplyBlockedReason:    true,
		ApplyGrantRequestID:   true,
		RuntimeCondition:      AgentRunConditionNone,
		BlockedReason:         "",
		GrantRequestID:        "",
	})
	require.NoError(t, err)

	events, err := ledger.GetJournalEvents(ctx, "arun-5")
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, runledger.EventRunCreated, events[0].Type)
}

func TestRunLedgerMirrorStore_BlockedToBlockedSameMetadataDoesNotAppendJournalEvent(t *testing.T) {
	ctx := context.Background()
	base := NewInMemoryAgentRunStore()
	ledger := runledger.NewMemoryStore()

	require.NoError(t, base.Create(&AgentRun{
		ID:               "arun-6",
		Status:           AgentRunRunning,
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "same reason",
		GrantRequestID:   "grant-arun-6",
	}))

	store := NewRunLedgerMirrorStore(base, ledger, nil)
	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "arun-6",
		Type:    runledger.EventRunCreated,
		Payload: json.RawMessage(`{"goal":"teammate"}`),
	}))

	err := store.UpdateProjection("arun-6", RunProjectionPatch{
		ApplyRuntimeCondition: true,
		ApplyBlockedReason:    true,
		ApplyGrantRequestID:   true,
		RuntimeCondition:      AgentRunConditionBlockedWaitingApproval,
		BlockedReason:         "same reason",
		GrantRequestID:        "grant-arun-6",
	})
	require.NoError(t, err)

	events, err := ledger.GetJournalEvents(ctx, "arun-6")
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, runledger.EventRunCreated, events[0].Type)
}
