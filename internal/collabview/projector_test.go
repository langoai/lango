package collabview

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMissionLinkedDelegationAttribution(t *testing.T) {
	t.Parallel()

	projector := NewProjector()
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)

	views := projector.Project(ProjectionInput{
		Missions: []MissionSource{
			{MissionID: "mission-1", ExecutionRefs: []string{"exec-1"}, UpdatedAt: now},
		},
		Delegations: []DelegationSource{
			{ExecutionRef: "exec-1", From: "planner", To: "researcher", Timestamp: now.Add(-time.Minute)},
		},
	})

	require.Len(t, views, 1)
	assert.Equal(t, CollaborationStateDelegating, views[0].CollaborationState)
	require.Len(t, views[0].HandoffEdges, 1)
	assert.Equal(t, "planner", views[0].HandoffEdges[0].From)
	assert.Equal(t, "researcher", views[0].HandoffEdges[0].To)
}

func TestNonAttributableSessionLevelDelegationIgnored(t *testing.T) {
	t.Parallel()

	projector := NewProjector()
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)

	views := projector.Project(ProjectionInput{
		Missions: []MissionSource{
			{MissionID: "mission-1", ExecutionRefs: []string{"exec-1"}, UpdatedAt: now},
		},
		Delegations: []DelegationSource{
			{ExecutionRef: "other-exec", From: "planner", To: "researcher", Timestamp: now},
			{ExecutionRef: "", From: "planner", To: "reviewer", Timestamp: now},
		},
	})

	require.Len(t, views, 1)
	assert.Equal(t, CollaborationStateSolo, views[0].CollaborationState)
	assert.Empty(t, views[0].HandoffEdges)
}

func TestParticipantExtraction(t *testing.T) {
	t.Parallel()

	projector := NewProjector()
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)

	views := projector.Project(ProjectionInput{
		Missions: []MissionSource{
			{MissionID: "mission-1", ExecutionRefs: []string{"exec-1"}, UpdatedAt: now},
		},
		AgentRuns: []AgentRunSource{
			{ExecutionRef: "exec-1", RequestedAgent: "researcher", UpdatedAt: now},
		},
		Delegations: []DelegationSource{
			{ExecutionRef: "exec-1", From: "planner", To: "researcher", Timestamp: now},
			{ExecutionRef: "exec-1", From: "researcher", To: "reviewer", Timestamp: now.Add(time.Minute)},
		},
	})

	require.Len(t, views, 1)
	participants := []string{
		views[0].Participants[0].Name,
		views[0].Participants[1].Name,
		views[0].Participants[2].Name,
	}
	assert.ElementsMatch(t, []string{"planner", "researcher", "reviewer"}, participants)
	assert.Equal(t, "reviewer", views[0].ActiveOwner)
}

func TestActiveOwnerFallsBackToLatestAttributedRunWhenNoHandoffExists(t *testing.T) {
	t.Parallel()

	projector := NewProjector()
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)

	views := projector.Project(ProjectionInput{
		Missions: []MissionSource{
			{MissionID: "mission-1", ExecutionRefs: []string{"exec-1", "exec-2"}, UpdatedAt: now},
		},
		AgentRuns: []AgentRunSource{
			{ExecutionRef: "exec-1", RequestedAgent: "researcher", UpdatedAt: now.Add(-2 * time.Minute)},
			{ExecutionRef: "exec-2", RequestedAgent: "reviewer", UpdatedAt: now.Add(-time.Minute)},
		},
	})

	require.Len(t, views, 1)
	assert.Equal(t, "reviewer", views[0].ActiveOwner)
}

func TestBlockedOnApprovalWaitingOnTeammateRecoveringStates(t *testing.T) {
	t.Parallel()

	projector := NewProjector()
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)

	views := projector.Project(ProjectionInput{
		Missions: []MissionSource{
			{MissionID: "approval", ExecutionRefs: []string{"exec-a"}, UpdatedAt: now},
			{MissionID: "teammate", ExecutionRefs: []string{"exec-b"}, UpdatedAt: now},
			{MissionID: "recovering", ExecutionRefs: []string{"exec-c"}, UpdatedAt: now},
		},
		AgentRuns: []AgentRunSource{
			{ExecutionRef: "exec-a", RequestedAgent: "researcher", RuntimeCondition: "blocked_waiting_approval", BlockedReason: "approval required", UpdatedAt: now},
			{ExecutionRef: "exec-b", RequestedAgent: "reviewer", RuntimeCondition: "waiting_on_teammate", UpdatedAt: now},
		},
		RecoverySignals: []RecoverySignalSource{
			{ExecutionRef: "exec-c", Action: "retry_with_hint", CauseClass: "timeout", Timestamp: now},
		},
	})

	require.Len(t, views, 3)
	byMission := viewsByMissionID(views)
	assert.Equal(t, CollaborationStateBlockedOnApproval, byMission["approval"].CollaborationState)
	assert.Equal(t, "approval required", byMission["approval"].BlockedOn)
	assert.Equal(t, CollaborationStateWaitingOnTeammate, byMission["teammate"].CollaborationState)
	assert.Equal(t, CollaborationStateRecovering, byMission["recovering"].CollaborationState)
	require.NotNil(t, byMission["recovering"].LastRecovery)
	assert.Equal(t, "retry_with_hint", byMission["recovering"].LastRecovery.Action)
}

func TestReviewStateSourceRules(t *testing.T) {
	t.Parallel()

	projector := NewProjector()
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)

	views := projector.Project(ProjectionInput{
		Missions: []MissionSource{
			{MissionID: "reviewing", ExecutionRefs: []string{"exec-a"}, UpdatedAt: now},
			{MissionID: "not-reviewing", ExecutionRefs: []string{"exec-b"}, UpdatedAt: now},
		},
		RunExecutions: []RunExecutionSource{
			{ExecutionRef: "exec-a", CurrentStepStatus: "verify_pending", UpdatedAt: now},
			{ExecutionRef: "exec-b", CurrentStepStatus: "in_progress", UpdatedAt: now},
		},
		Delegations: []DelegationSource{
			{ExecutionRef: "exec-b", From: "planner", To: "reviewer", Timestamp: now},
		},
	})

	require.Len(t, views, 2)
	byMission := viewsByMissionID(views)
	assert.Equal(t, CollaborationStateReviewing, byMission["reviewing"].CollaborationState)
	assert.Equal(t, CollaborationStateDelegating, byMission["not-reviewing"].CollaborationState)
}

func TestBudgetRecoveryOnlyWhenAttributionProvable(t *testing.T) {
	t.Parallel()

	projector := NewProjector()
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)

	views := projector.Project(ProjectionInput{
		Missions: []MissionSource{
			{MissionID: "mission-1", ExecutionRefs: []string{"exec-1"}, UpdatedAt: now},
		},
		BudgetSignals: []BudgetSignalSource{
			{ExecutionRef: "exec-1", Used: 12, Max: 15, Timestamp: now},
			{ExecutionRef: "other-exec", Used: 20, Max: 20, Timestamp: now},
		},
		RecoverySignals: []RecoverySignalSource{
			{ExecutionRef: "exec-1", Action: "retry_with_hint", CauseClass: "rate_limit", Timestamp: now},
			{ExecutionRef: "other-exec", Action: "fallback", CauseClass: "timeout", Timestamp: now},
		},
	})

	require.Len(t, views, 1)
	require.NotNil(t, views[0].BudgetSignal)
	assert.Equal(t, 12, views[0].BudgetSignal.Used)
	require.NotNil(t, views[0].LastRecovery)
	assert.Equal(t, "retry_with_hint", views[0].LastRecovery.Action)
}

func viewsByMissionID(views []CollaborationView) map[string]CollaborationView {
	out := make(map[string]CollaborationView, len(views))
	for _, view := range views {
		out[view.MissionID] = view
	}
	return out
}
