package mission

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	entmission "github.com/langoai/lango/internal/ent/mission"
	entmissionstatehistory "github.com/langoai/lango/internal/ent/missionstatehistory"
)

func TestServiceStartMissionCreatesPreparedMission(t *testing.T) {
	t.Parallel()

	store, client := newTestStore(t)
	svc := NewService(store)
	ctx := context.Background()

	row, err := svc.StartMission(ctx, StartMissionInput{
		SessionKey:  "svc-start",
		Title:       "Ship Wave 2 durable missions",
		Description: "Create durable mission rows before turn dispatch.",
	})
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, StatusPrepared, row.Status)
	assert.Equal(t, "Ship Wave 2 durable missions", row.Title)
	assert.Equal(t, "user", row.SourceKind)
	require.NotNil(t, row.Description)
	assert.Equal(t, "Create durable mission rows before turn dispatch.", *row.Description)

	historyCount, err := client.MissionStateHistory.Query().
		Where(entmissionstatehistory.MissionID(row.ID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, historyCount)
}

func TestServiceAcceptProposalCreatesFirstDurableMissionRow(t *testing.T) {
	t.Parallel()

	store, client := newTestStore(t)
	svc := NewService(store)
	ctx := context.Background()

	row, err := svc.AcceptProposal(ctx, AcceptProposalInput{
		SessionKey:  "svc-proposal",
		SourceKind:  "proposed_learning",
		SourceRef:   "learn-42",
		Title:       "Apply learning rule: preserve mission binding",
		Description: "Use the accepted proposal as the first durable mission row.",
	})
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, StatusPrepared, row.Status)
	assert.Equal(t, "proposed_learning", row.SourceKind)
	require.NotNil(t, row.SourceRef)
	assert.Equal(t, "learn-42", *row.SourceRef)

	listed, err := store.ListMissionsBySession(ctx, "svc-proposal", 10)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, row.ID, listed[0].ID)

	historyCount, err := client.MissionStateHistory.Query().
		Where(entmissionstatehistory.MissionID(row.ID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, historyCount)
}

func TestServiceMarkWaitingDecisionStoresCoarseSummary(t *testing.T) {
	t.Parallel()

	store, _ := newTestStore(t)
	svc := NewService(store)
	ctx := context.Background()

	row, err := svc.StartMission(ctx, StartMissionInput{
		SessionKey:  "svc-waiting",
		Title:       "Patch mission approval flow",
		Description: "Start active so the mission can pause on approval.",
		StartActive: true,
	})
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, StatusActive, row.Status)

	waiting, err := svc.MarkWaitingDecision(ctx, WaitForDecisionInput{
		MissionID:       row.ID.String(),
		Reason:          "Tool approval requested",
		ActorKind:       "system",
		ActorRef:        "approval-middleware",
		DecisionKind:    "tool_approval",
		DecisionSummary: "Approve filesystem write for mission patch",
	})
	require.NoError(t, err)
	require.NotNil(t, waiting)
	assert.Equal(t, StatusWaitingDecision, waiting.Status)
	require.NotNil(t, waiting.CurrentDecisionKind)
	assert.Equal(t, "tool_approval", *waiting.CurrentDecisionKind)
	require.NotNil(t, waiting.CurrentDecisionSummary)
	assert.Equal(t, "Approve filesystem write for mission patch", *waiting.CurrentDecisionSummary)
	assert.Nil(t, waiting.CurrentBlockedReason)
}

func TestServiceAttachExecutionIsIdempotent(t *testing.T) {
	t.Parallel()

	store, _ := newTestStore(t)
	svc := NewService(store)
	ctx := context.Background()

	row, err := svc.StartMission(ctx, StartMissionInput{
		SessionKey: "svc-links",
		Title:      "Track mission execution linkage",
	})
	require.NoError(t, err)

	input := AttachExecutionInput{
		MissionID:     row.ID.String(),
		ExecutionKind: ExecutionKindTaskOSExecution,
		ExecutionRef:  "task-123",
		LinkRole:      LinkRolePrimary,
	}
	require.NoError(t, svc.AttachExecution(ctx, input))
	require.NoError(t, svc.AttachExecution(ctx, input))

	links, err := store.ListExecutionLinks(ctx, row.ID.String())
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, "task-123", links[0].ExecutionRef)
	assert.Equal(t, LinkRolePrimary, links[0].LinkRole)
}

func TestServiceInvalidTransitionPropagatesFromStore(t *testing.T) {
	t.Parallel()

	store, _ := newTestStore(t)
	svc := NewService(store)
	ctx := context.Background()

	row, err := svc.StartMission(ctx, StartMissionInput{
		SessionKey:  "svc-invalid",
		Title:       "Already active mission",
		StartActive: true,
	})
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, entmission.StatusActive, row.Status)

	updated, err := svc.MarkActive(ctx, ActivateMissionInput{
		MissionID: row.ID.String(),
		Reason:    "Attempt to re-activate",
		ActorKind: "system",
		ActorRef:  "test",
	})
	require.Error(t, err)
	assert.Nil(t, updated)
	assert.Contains(t, err.Error(), "invalid transition")
}
