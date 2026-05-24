package mission

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	entmission "github.com/langoai/lango/internal/ent/mission"
	entmissionexecutionlink "github.com/langoai/lango/internal/ent/missionexecutionlink"
	entmissionstatehistory "github.com/langoai/lango/internal/ent/missionstatehistory"
)

func TestStoreUnavailableAndValidationBranches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var nilStore *EntStore

	_, err := nilStore.CreateMission(ctx, CreateMissionInput{})
	require.ErrorContains(t, err, "store unavailable")
	_, err = nilStore.GetMission(ctx, "mission")
	require.ErrorContains(t, err, "store unavailable")
	_, err = nilStore.ListMissionsBySession(ctx, "session", 1)
	require.ErrorContains(t, err, "store unavailable")
	_, err = nilStore.TransitionMission(ctx, TransitionMissionInput{})
	require.ErrorContains(t, err, "store unavailable")
	err = nilStore.AppendExecutionLink(ctx, AppendExecutionLinkInput{})
	require.ErrorContains(t, err, "store unavailable")
	_, err = nilStore.ListExecutionLinks(ctx, "mission")
	require.ErrorContains(t, err, "store unavailable")
	_, err = nilStore.FindExecutionLinkByExecution(ctx, ExecutionKindRunLedgerRun, "run")
	require.ErrorContains(t, err, "store unavailable")

	store, _ := newTestStore(t)
	_, err = store.CreateMission(ctx, CreateMissionInput{
		Title:      "missing session",
		SourceKind: "user",
	})
	require.ErrorContains(t, err, "session_key is required")
	_, err = store.CreateMission(ctx, CreateMissionInput{
		SessionKey: "session",
		SourceKind: "user",
	})
	require.ErrorContains(t, err, "title is required")
	_, err = store.CreateMission(ctx, CreateMissionInput{
		SessionKey: "session",
		Title:      "missing source",
	})
	require.ErrorContains(t, err, "source_kind is required")
	_, err = store.CreateMission(ctx, CreateMissionInput{
		SessionKey: "session",
		Title:      "invalid status",
		SourceKind: "user",
		Status:     entmission.Status("impossible"),
	})
	require.ErrorContains(t, err, "create mission")

	_, err = store.ListMissionsBySession(ctx, "   ", 1)
	require.ErrorContains(t, err, "session_key is required")
	_, err = store.GetMission(ctx, "not-a-uuid")
	require.ErrorContains(t, err, "parse mission id")
}

func TestStoreExecutionLinkDefaultsAndLookupMisses(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, _ := newTestStore(t)

	created, err := store.CreateMission(ctx, CreateMissionInput{
		SessionKey: "mission-link-branch-session",
		Title:      "Link default role",
		SourceKind: "user",
	})
	require.NoError(t, err)

	err = store.AppendExecutionLink(ctx, AppendExecutionLinkInput{
		MissionID:     created.ID.String(),
		ExecutionKind: ExecutionKindRunLedgerRun,
		ExecutionRef:  "  run-mission-link-branch  ",
	})
	require.NoError(t, err)

	links, err := store.ListExecutionLinks(ctx, created.ID.String())
	require.NoError(t, err)
	require.Len(t, links, 1)
	require.Equal(t, "run-mission-link-branch", links[0].ExecutionRef)
	require.Equal(t, LinkRolePrimary, links[0].LinkRole)

	got, err := store.FindMissionByExecution(ctx, ExecutionKindRunLedgerRun, "run-mission-link-branch")
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)

	missing, err := store.FindExecutionLinkByExecution(ctx, ExecutionKindRunLedgerRun, "missing")
	require.NoError(t, err)
	require.Nil(t, missing)
	missingMission, err := store.FindMissionByExecution(ctx, ExecutionKindRunLedgerRun, "missing")
	require.NoError(t, err)
	require.Nil(t, missingMission)

	err = store.AppendExecutionLink(ctx, AppendExecutionLinkInput{
		MissionID:     created.ID.String(),
		ExecutionKind: ExecutionKind("bad-kind"),
		ExecutionRef:  "run",
	})
	require.ErrorContains(t, err, "append execution link")
	err = store.AppendExecutionLink(ctx, AppendExecutionLinkInput{
		MissionID:     created.ID.String(),
		ExecutionKind: ExecutionKindRunLedgerRun,
	})
	require.ErrorContains(t, err, "execution_ref is required")
	err = store.AppendExecutionLink(ctx, AppendExecutionLinkInput{
		MissionID:     created.ID.String(),
		ExecutionKind: ExecutionKindRunLedgerRun,
		ExecutionRef:  "run",
		LinkRole:      LinkRole("bad-role"),
	})
	require.ErrorContains(t, err, "append execution link")
}

func TestMissionStateHelperBranches(t *testing.T) {
	t.Parallel()

	require.Equal(t, entmissionstatehistory.FromStatusPrepared, mustHistoryFromStatus(t, StatusPrepared))
	require.Equal(t, entmissionstatehistory.FromStatusCancelled, mustHistoryFromStatus(t, StatusCancelled))
	_, ok := toHistoryFromStatus(entmission.Status("unknown"))
	require.False(t, ok)

	require.Equal(t, entmissionstatehistory.ToStatusPrepared, toHistoryStatus(entmission.Status("unknown")))
	require.False(t, isAllowedMissionTransition(StatusDone, StatusActive))
	require.True(t, isAllowedMissionTransition(StatusWaitingDecision, StatusWaitingDecision))
	require.True(t, isAllowedMissionTransition(StatusBlocked, StatusCancelled))

	existingCompletedAt := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	latest := normalizeLatestState(
		StatusCancelled,
		" stale blocker ",
		"decision",
		"summary",
		nil,
		&existingCompletedAt,
		existingCompletedAt.Add(time.Hour),
	)
	require.Equal(t, StatusCancelled, latest.status)
	require.Nil(t, latest.blockedReason)
	require.Nil(t, latest.decisionKind)
	require.Nil(t, latest.decisionSummary)
	require.NotNil(t, latest.completedAt)
	require.Equal(t, existingCompletedAt, *latest.completedAt)

	require.NoError(t, entmissionexecutionlink.ExecutionKindValidator(ExecutionKindTaskOSExecution))
}

func mustHistoryFromStatus(t *testing.T, status Status) entmissionstatehistory.FromStatus {
	t.Helper()
	got, ok := toHistoryFromStatus(status)
	require.True(t, ok)
	return got
}
