package mission

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/mission"
	"github.com/langoai/lango/internal/ent/missionexecutionlink"
	"github.com/langoai/lango/internal/ent/missionstatehistory"
	"github.com/langoai/lango/internal/testutil"
)

func TestStoreCreateGetListOrdering(t *testing.T) {
	t.Parallel()

	store, client := newTestStore(t)
	ctx := context.Background()

	first, err := store.CreateMission(ctx, CreateMissionInput{
		SessionKey:  "sess-1",
		Title:       "Prepare follow-up brief",
		Description: "Collect context before execution.",
		SourceKind:  "user",
		SourceRef:   "input-1",
	})
	require.NoError(t, err)

	second, err := store.CreateMission(ctx, CreateMissionInput{
		SessionKey: "sess-1",
		Title:      "Ship mission lifecycle",
		Status:     mission.StatusActive,
		SourceKind: "manual",
	})
	require.NoError(t, err)

	older := time.Date(2026, 5, 3, 9, 0, 0, 0, time.UTC)
	newer := older.Add(5 * time.Minute)
	_, err = client.Mission.UpdateOneID(first.ID).SetUpdatedAt(older).Save(ctx)
	require.NoError(t, err)
	_, err = client.Mission.UpdateOneID(second.ID).SetUpdatedAt(newer).Save(ctx)
	require.NoError(t, err)

	got, err := store.GetMission(ctx, first.ID.String())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, first.ID, got.ID)
	assert.Equal(t, "Prepare follow-up brief", got.Title)
	require.NotNil(t, got.Description)
	assert.Equal(t, "Collect context before execution.", *got.Description)
	require.NotNil(t, got.SourceRef)
	assert.Equal(t, "input-1", *got.SourceRef)

	listed, err := store.ListMissionsBySession(ctx, "sess-1", 10)
	require.NoError(t, err)
	require.Len(t, listed, 2)
	assert.Equal(t, second.ID, listed[0].ID)
	assert.Equal(t, first.ID, listed[1].ID)
}

func TestStoreTransitionAppendsPerMissionHistorySeq(t *testing.T) {
	t.Parallel()

	store, client := newTestStore(t)
	ctx := context.Background()

	created, err := store.CreateMission(ctx, CreateMissionInput{
		SessionKey: "sess-2",
		Title:      "Investigate blocker",
		SourceKind: "user",
	})
	require.NoError(t, err)

	transitioned, err := store.TransitionMission(ctx, TransitionMissionInput{
		MissionID:     created.ID.String(),
		ToStatus:      mission.StatusActive,
		Reason:        "Execution started",
		ActorKind:     "user",
		ActorRef:      "operator",
		ExecutionKind: string(missionexecutionlink.ExecutionKindRunledgerRun),
		ExecutionRef:  "run-1",
	})
	require.NoError(t, err)
	require.NotNil(t, transitioned)
	assert.Equal(t, mission.StatusActive, transitioned.Status)
	assert.Nil(t, transitioned.CompletedAt)

	transitioned, err = store.TransitionMission(ctx, TransitionMissionInput{
		MissionID:       created.ID.String(),
		ToStatus:        mission.StatusBlocked,
		Reason:          "Approval required",
		ActorKind:       "system",
		ActorRef:        "approval-router",
		ExecutionKind:   string(missionexecutionlink.ExecutionKindRunledgerRun),
		ExecutionRef:    "run-1",
		BlockedReason:   "Waiting for filesystem approval",
		DecisionKind:    "tool_approval",
		DecisionSummary: "Approve filesystem write for patch application",
	})
	require.NoError(t, err)
	require.NotNil(t, transitioned)
	assert.Equal(t, mission.StatusBlocked, transitioned.Status)
	require.NotNil(t, transitioned.CurrentBlockedReason)
	assert.Equal(t, "Waiting for filesystem approval", *transitioned.CurrentBlockedReason)

	rows, err := client.MissionStateHistory.Query().
		Where(missionstatehistory.MissionID(created.ID)).
		Order(missionstatehistory.BySeq()).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 2)

	assert.Equal(t, int64(1), rows[0].Seq)
	require.NotNil(t, rows[0].FromStatus)
	assert.Equal(t, missionstatehistory.FromStatusPrepared, *rows[0].FromStatus)
	assert.Equal(t, missionstatehistory.ToStatusActive, rows[0].ToStatus)

	assert.Equal(t, int64(2), rows[1].Seq)
	require.NotNil(t, rows[1].FromStatus)
	assert.Equal(t, missionstatehistory.FromStatusActive, *rows[1].FromStatus)
	assert.Equal(t, missionstatehistory.ToStatusBlocked, rows[1].ToStatus)
	require.NotNil(t, rows[1].DecisionKind)
	assert.Equal(t, "tool_approval", *rows[1].DecisionKind)

	latest, err := store.GetMission(ctx, created.ID.String())
	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, mission.StatusBlocked, latest.Status)
	require.NotNil(t, latest.CurrentBlockedReason)
	assert.Equal(t, "Waiting for filesystem approval", *latest.CurrentBlockedReason)
}

func TestStoreCreateMissionAppendsInitialHistoryWhenRequested(t *testing.T) {
	t.Parallel()

	store, client := newTestStore(t)
	ctx := context.Background()

	created, err := store.CreateMission(ctx, CreateMissionInput{
		SessionKey:       "sess-init-history",
		Title:            "Create first durable mission row",
		Description:      "Initial creation should also write seq=1 history.",
		Status:           mission.StatusPrepared,
		SourceKind:       "user",
		InitialReason:    "mission created",
		InitialActorKind: "user",
		InitialActorRef:  "operator",
		InitialPayload:   map[string]any{"path": "direct_start"},
	})
	require.NoError(t, err)
	require.NotNil(t, created)

	rows, err := client.MissionStateHistory.Query().
		Where(missionstatehistory.MissionID(created.ID)).
		Order(missionstatehistory.BySeq()).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 1)

	assert.Equal(t, int64(1), rows[0].Seq)
	assert.Nil(t, rows[0].FromStatus)
	assert.Equal(t, missionstatehistory.ToStatusPrepared, rows[0].ToStatus)
	require.NotNil(t, rows[0].Reason)
	assert.Equal(t, "mission created", *rows[0].Reason)
	assert.Equal(t, "user", rows[0].ActorKind)
	require.NotNil(t, rows[0].ActorRef)
	assert.Equal(t, "operator", *rows[0].ActorRef)
	require.NotNil(t, rows[0].Payload)
	assert.Equal(t, "direct_start", rows[0].Payload["path"])
}

func TestStoreAppendExecutionLinkRejectsDuplicateExecution(t *testing.T) {
	t.Parallel()

	store, _ := newTestStore(t)
	ctx := context.Background()

	first, err := store.CreateMission(ctx, CreateMissionInput{
		SessionKey: "sess-3",
		Title:      "Primary mission",
		SourceKind: "user",
	})
	require.NoError(t, err)

	second, err := store.CreateMission(ctx, CreateMissionInput{
		SessionKey: "sess-3",
		Title:      "Secondary mission",
		SourceKind: "user",
	})
	require.NoError(t, err)

	err = store.AppendExecutionLink(ctx, AppendExecutionLinkInput{
		MissionID:     first.ID.String(),
		ExecutionKind: missionexecutionlink.ExecutionKindTaskOsExecution,
		ExecutionRef:  "task-123",
		LinkRole:      missionexecutionlink.LinkRolePrimary,
	})
	require.NoError(t, err)

	err = store.AppendExecutionLink(ctx, AppendExecutionLinkInput{
		MissionID:     second.ID.String(),
		ExecutionKind: missionexecutionlink.ExecutionKindTaskOsExecution,
		ExecutionRef:  "task-123",
		LinkRole:      missionexecutionlink.LinkRoleRetry,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "execution link")
}

func TestStoreFindMissionByExecutionReference(t *testing.T) {
	t.Parallel()

	store, _ := newTestStore(t)
	ctx := context.Background()

	created, err := store.CreateMission(ctx, CreateMissionInput{
		SessionKey: "sess-4",
		Title:      "Link runtime execution",
		SourceKind: "manual",
	})
	require.NoError(t, err)

	err = store.AppendExecutionLink(ctx, AppendExecutionLinkInput{
		MissionID:     created.ID.String(),
		ExecutionKind: missionexecutionlink.ExecutionKindRunledgerRun,
		ExecutionRef:  "run-42",
		LinkRole:      missionexecutionlink.LinkRolePrimary,
	})
	require.NoError(t, err)

	link, err := store.FindExecutionLinkByExecution(ctx, missionexecutionlink.ExecutionKindRunledgerRun, "run-42")
	require.NoError(t, err)
	require.NotNil(t, link)
	assert.Equal(t, created.ID, link.MissionID)

	got, err := store.FindMissionByExecution(ctx, missionexecutionlink.ExecutionKindRunledgerRun, "run-42")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, created.ID, got.ID)

	links, err := store.ListExecutionLinks(ctx, created.ID.String())
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, "run-42", links[0].ExecutionRef)
}

func TestStoreTransitionRejectsInvalidTransition(t *testing.T) {
	t.Parallel()

	store, client := newTestStore(t)
	ctx := context.Background()

	created, err := store.CreateMission(ctx, CreateMissionInput{
		SessionKey: "sess-5",
		Title:      "Invalid transition guard",
		SourceKind: "user",
	})
	require.NoError(t, err)

	transitioned, err := store.TransitionMission(ctx, TransitionMissionInput{
		MissionID:   created.ID.String(),
		ToStatus:    mission.StatusDone,
		Reason:      "Skip directly to done",
		ActorKind:   "system",
		ActorRef:    "test",
		CompletedAt: nil,
	})
	require.Error(t, err)
	assert.Nil(t, transitioned)
	assert.Contains(t, err.Error(), "invalid transition")

	latest, err := store.GetMission(ctx, created.ID.String())
	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, mission.StatusPrepared, latest.Status)
	assert.Nil(t, latest.CompletedAt)

	historyCount, err := client.MissionStateHistory.Query().
		Where(missionstatehistory.MissionID(created.ID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, historyCount)
}

func TestStoreLatestStateInvariantsAndTerminalTimestamp(t *testing.T) {
	t.Parallel()

	store, _ := newTestStore(t)
	ctx := context.Background()

	impossibleCompletedAt := time.Date(2026, 5, 3, 11, 0, 0, 0, time.UTC)
	active, err := store.CreateMission(ctx, CreateMissionInput{
		SessionKey:             "sess-6",
		Title:                  "Normalize impossible latest state",
		Status:                 mission.StatusActive,
		SourceKind:             "manual",
		CurrentBlockedReason:   "Should be cleared",
		CurrentDecisionKind:    "tool_approval",
		CurrentDecisionSummary: "Should be cleared",
		CompletedAt:            &impossibleCompletedAt,
	})
	require.NoError(t, err)
	assert.Equal(t, mission.StatusActive, active.Status)
	assert.Nil(t, active.CurrentBlockedReason)
	assert.Nil(t, active.CurrentDecisionKind)
	assert.Nil(t, active.CurrentDecisionSummary)
	assert.Nil(t, active.CompletedAt)

	waiting, err := store.CreateMission(ctx, CreateMissionInput{
		SessionKey:             "sess-6",
		Title:                  "Cancel from waiting decision",
		Status:                 mission.StatusWaitingDecision,
		SourceKind:             "manual",
		CurrentBlockedReason:   "Should be cleared",
		CurrentDecisionKind:    "tool_approval",
		CurrentDecisionSummary: "Approve filesystem write",
	})
	require.NoError(t, err)
	require.NotNil(t, waiting.CurrentDecisionKind)
	require.NotNil(t, waiting.CurrentDecisionSummary)
	assert.Nil(t, waiting.CurrentBlockedReason)
	assert.Nil(t, waiting.CompletedAt)

	cancelledRow, err := store.TransitionMission(ctx, TransitionMissionInput{
		MissionID: waiting.ID.String(),
		ToStatus:  mission.StatusCancelled,
		Reason:    "User abandoned the work",
		ActorKind: "user",
		ActorRef:  "operator",
	})
	require.NoError(t, err)
	require.NotNil(t, cancelledRow)
	assert.Equal(t, mission.StatusCancelled, cancelledRow.Status)
	require.NotNil(t, cancelledRow.CompletedAt)

	cancelled, err := store.GetMission(ctx, waiting.ID.String())
	require.NoError(t, err)
	require.NotNil(t, cancelled)
	assert.Equal(t, mission.StatusCancelled, cancelled.Status)
	assert.Nil(t, cancelled.CurrentBlockedReason)
	assert.Nil(t, cancelled.CurrentDecisionKind)
	assert.Nil(t, cancelled.CurrentDecisionSummary)
	require.NotNil(t, cancelled.CompletedAt)
	assert.False(t, cancelled.CompletedAt.IsZero())

	explicitCompletedAt := time.Date(2026, 5, 3, 12, 30, 0, 0, time.UTC)
	cancelledAtCreate, err := store.CreateMission(ctx, CreateMissionInput{
		SessionKey:             "sess-6",
		Title:                  "Cancelled before execution",
		Status:                 mission.StatusCancelled,
		SourceKind:             "manual",
		CurrentDecisionKind:    "tool_approval",
		CurrentDecisionSummary: "Should be cleared",
		CompletedAt:            &explicitCompletedAt,
	})
	require.NoError(t, err)
	require.NotNil(t, cancelledAtCreate.CompletedAt)
	assert.Equal(t, explicitCompletedAt, *cancelledAtCreate.CompletedAt)
	assert.Nil(t, cancelledAtCreate.CurrentDecisionKind)
	assert.Nil(t, cancelledAtCreate.CurrentDecisionSummary)
}

func TestStoreTransitionAllowsWaitingDecisionSummaryRefresh(t *testing.T) {
	t.Parallel()

	store, client := newTestStore(t)
	ctx := context.Background()

	created, err := store.CreateMission(ctx, CreateMissionInput{
		SessionKey: "sess-decision-refresh",
		Title:      "Refresh waiting decision summary",
		SourceKind: "user",
	})
	require.NoError(t, err)

	waiting, err := store.TransitionMission(ctx, TransitionMissionInput{
		MissionID:       created.ID.String(),
		ToStatus:        mission.StatusWaitingDecision,
		Reason:          "First approval requested",
		ActorKind:       "system",
		ActorRef:        "approval",
		DecisionKind:    "tool_approval",
		DecisionSummary: "Approve filesystem write",
	})
	require.NoError(t, err)
	require.NotNil(t, waiting)
	assert.Equal(t, mission.StatusWaitingDecision, waiting.Status)

	waiting, err = store.TransitionMission(ctx, TransitionMissionInput{
		MissionID:       created.ID.String(),
		ToStatus:        mission.StatusWaitingDecision,
		Reason:          "Approval denied",
		ActorKind:       "system",
		ActorRef:        "approval",
		DecisionKind:    "tool_approval",
		DecisionSummary: "Filesystem write denied; choose a safer path",
	})
	require.NoError(t, err)
	require.NotNil(t, waiting)
	assert.Equal(t, mission.StatusWaitingDecision, waiting.Status)
	require.NotNil(t, waiting.CurrentDecisionSummary)
	assert.Equal(t, "Filesystem write denied; choose a safer path", *waiting.CurrentDecisionSummary)

	rows, err := client.MissionStateHistory.Query().
		Where(missionstatehistory.MissionID(created.ID)).
		Order(missionstatehistory.BySeq()).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, int64(2), rows[1].Seq)
	require.NotNil(t, rows[1].FromStatus)
	assert.Equal(t, missionstatehistory.FromStatusWaitingDecision, *rows[1].FromStatus)
	assert.Equal(t, missionstatehistory.ToStatusWaitingDecision, rows[1].ToStatus)
	require.NotNil(t, rows[1].DecisionSummary)
	assert.Equal(t, "Filesystem write denied; choose a safer path", *rows[1].DecisionSummary)
}

func newTestStore(t *testing.T) (*EntStore, *ent.Client) {
	t.Helper()

	client := testutil.TestEntClient(t)
	return NewEntStore(client), client
}
