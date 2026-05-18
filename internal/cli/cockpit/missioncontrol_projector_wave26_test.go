package cockpit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/proposal"
	"github.com/langoai/lango/internal/runledger"
)

func TestWave26MissionControlProjectorStatusHelpersCoverMappings(t *testing.T) {
	t.Parallel()

	taskTests := []struct {
		giveStatus         string
		wantMissionStatus  MissionStatus
		wantNextAction     string
		wantProposalStatus string
		wantStatusPriority int
		wantDurableStatus  string
	}{
		{
			giveStatus:         " pending ",
			wantMissionStatus:  MissionStatusPending,
			wantNextAction:     "Wait for execution slot",
			wantProposalStatus: "preparing",
			wantStatusPriority: 3,
			wantDurableStatus:  "waiting_decision",
		},
		{
			giveStatus:         "RUNNING",
			wantMissionStatus:  MissionStatusRunning,
			wantNextAction:     "Monitor background execution",
			wantStatusPriority: 0,
			wantDurableStatus:  "active",
		},
		{
			giveStatus:         "done",
			wantMissionStatus:  MissionStatusDone,
			wantNextAction:     "Review completed output",
			wantStatusPriority: 4,
			wantDurableStatus:  "done",
		},
		{
			giveStatus:         "failed",
			wantMissionStatus:  MissionStatusFailed,
			wantNextAction:     "Inspect failure and retry",
			wantStatusPriority: 5,
		},
		{
			giveStatus:         "cancelled",
			wantMissionStatus:  MissionStatusCancelled,
			wantNextAction:     "Restart if still needed",
			wantStatusPriority: 6,
			wantDurableStatus:  "cancelled",
		},
		{
			giveStatus:         "prepared",
			wantMissionStatus:  MissionStatusUnknown,
			wantNextAction:     "Inspect task state",
			wantStatusPriority: 7,
			wantDurableStatus:  "",
		},
	}

	for _, tt := range taskTests {
		t.Run(tt.giveStatus, func(t *testing.T) {
			gotStatus := missionStatusFromTask(tt.giveStatus)
			assert.Equal(t, tt.wantMissionStatus, gotStatus)
			assert.Equal(t, tt.wantNextAction, nextActionForTask(tt.giveStatus))
			assert.Equal(t, tt.wantProposalStatus, proposalStatusString(gotStatus))
			assert.Equal(t, tt.wantStatusPriority, missionStatusPriority(gotStatus))
			assert.Equal(t, tt.wantDurableStatus, durableMissionStatusString(gotStatus))
		})
	}

	proposalTests := []struct {
		giveStatus proposal.ProposalStatus
		wantStatus MissionStatus
	}{
		{giveStatus: proposal.ProposalStatusPrepared, wantStatus: MissionStatusPrepared},
		{giveStatus: proposal.ProposalStatusPreparing, wantStatus: MissionStatusPending},
		{giveStatus: proposal.ProposalStatusSuggested, wantStatus: MissionStatusPending},
		{giveStatus: proposal.ProposalStatusAccepted, wantStatus: MissionStatusUnknown},
	}
	for _, tt := range proposalTests {
		t.Run(string(tt.giveStatus), func(t *testing.T) {
			assert.Equal(t, tt.wantStatus, missionStatusFromProposal(tt.giveStatus))
		})
	}

	assert.Equal(t, 1, missionStatusPriority(MissionStatusBlocked))
	assert.Equal(t, 2, missionStatusPriority(MissionStatusPrepared))
	assert.Equal(t, "prepared", proposalStatusString(MissionStatusPrepared))
	assert.Equal(t, "blocked", durableMissionStatusString(MissionStatusBlocked))
}

func TestWave26MissionControlProjectorTimeAndSummaryHelpersCoverEdges(t *testing.T) {
	t.Parallel()

	wantTime := time.Date(2026, 5, 18, 9, 10, 11, 0, time.UTC)
	assert.Equal(t, wantTime, parseRFC3339OrZero("2026-05-18T09:10:11Z"))
	assert.True(t, parseRFC3339OrZero("   ").IsZero())
	assert.True(t, parseRFC3339OrZero("not-rfc3339").IsZero())

	assert.Equal(t, "1 more mission", pluralSummary(1, "mission", "missions"))
	assert.Equal(t, "2 more missions", pluralSummary(2, "mission", "missions"))
	assert.Equal(t, "0 more missions", pluralSummary(0, "mission", "missions"))
}

func TestWave26MissionControlProjectorCurrentRunStepStatusBranches(t *testing.T) {
	t.Parallel()

	assert.Empty(t, currentRunStepStatus(nil))

	tests := []struct {
		name     string
		giveSnap *runledger.RunSnapshot
		want     string
	}{
		{
			name: "current step wins",
			giveSnap: &runledger.RunSnapshot{
				CurrentStepID: " current ",
				Steps: []runledger.Step{
					{StepID: "current", Status: runledger.StepStatusInProgress},
					{StepID: "review", Status: runledger.StepStatusVerifyPending},
				},
			},
			want: "in_progress",
		},
		{
			name: "verify pending fallback",
			giveSnap: &runledger.RunSnapshot{
				CurrentStepID: "missing",
				Steps: []runledger.Step{
					{StepID: "done", Status: runledger.StepStatusCompleted},
					{StepID: "review", Status: runledger.StepStatusVerifyPending},
				},
			},
			want: "verify_pending",
		},
		{
			name: "no relevant step",
			giveSnap: &runledger.RunSnapshot{
				Steps: []runledger.Step{{StepID: "done", Status: runledger.StepStatusCompleted}},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, currentRunStepStatus(tt.giveSnap))
		})
	}
}

func TestWave26MissionControlProjectorRuntimeHintAndBlockedAgentBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		condition  agentrt.AgentRunCondition
		blocked    string
		wantHint   string
		wantReason string
		wantAction string
		wantBlock  bool
	}{
		{
			name:       "approval",
			condition:  agentrt.AgentRunConditionBlockedWaitingApproval,
			wantHint:   "Waiting for approval",
			wantReason: "Waiting for approval",
			wantAction: "Resolve approval request",
			wantBlock:  true,
		},
		{
			name:       "message with reason",
			condition:  agentrt.AgentRunConditionBlockedWaitingMessage,
			blocked:    "Need operator reply",
			wantHint:   "Waiting for message",
			wantReason: "Need operator reply",
			wantAction: "Respond to required message",
			wantBlock:  true,
		},
		{
			name:      "teammate",
			condition: agentrt.AgentRunConditionWaitingOnTeammate,
			wantHint:  "Waiting on teammate",
		},
		{
			name:      "resuming",
			condition: agentrt.AgentRunConditionResuming,
			wantHint:  "Resuming",
		},
		{
			name:      "orphaned",
			condition: agentrt.AgentRunConditionOrphaned,
			wantHint:  "Orphaned",
		},
		{
			name:      "recovering",
			condition: agentrt.AgentRunConditionRecovering,
			wantHint:  "Recovering",
		},
		{
			name:     "fallback reason",
			blocked:  "  custom runtime note  ",
			wantHint: "custom runtime note",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantHint, runtimeHintForAgentRun(tt.condition, tt.blocked))
			gotReason, gotAction, gotBlock := blockedStateForAgentRun(tt.condition, tt.blocked)
			assert.Equal(t, tt.wantReason, gotReason)
			assert.Equal(t, tt.wantAction, gotAction)
			assert.Equal(t, tt.wantBlock, gotBlock)
		})
	}
}

func TestWave26MissionControlProjectorRunLedgerEnrichmentBranches(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2026, 5, 18, 9, 0, 0, 0, time.UTC)
	missionView := MissionView{
		ID:          "bg:task-1",
		Status:      MissionStatusRunning,
		NextAction:  "Monitor background execution",
		RuntimeHint: "existing runtime",
		UpdatedAt:   baseTime,
	}

	enrichMissionFromRunLedger(&missionView, &runledger.RunSnapshot{
		Goal:                     "  Build\x1b[31m projector\ncoverage  ",
		CurrentBlocker:           "  approval required  ",
		TeammateRuntimeCondition: "should not replace existing",
		UpdatedAt:                baseTime.Add(time.Minute),
		Steps: []runledger.Step{{
			StepID: "next",
			Goal:   "This step is blocked",
			Status: runledger.StepStatusPending,
		}},
	})

	assert.Equal(t, MissionStatusBlocked, missionView.Status)
	assert.Equal(t, "Build projector coverage", missionView.Detail)
	assert.Equal(t, "approval required", missionView.BlockedReason)
	assert.Equal(t, "existing runtime", missionView.RuntimeHint)
	assert.Equal(t, "Resolve blocker", missionView.NextAction)
	assert.Equal(t, baseTime.Add(time.Minute), missionView.UpdatedAt)

	durableView := MissionView{
		ID:         "mission-1",
		NextAction: "Resolve pending decision",
		UpdatedAt:  baseTime,
	}
	enrichDurableMissionFromRunLedger(&durableView, &runledger.RunSnapshot{
		Goal:                     "Durable goal",
		TeammateBlockedReason:    "teammate blocked",
		TeammateRuntimeCondition: "waiting on specialist",
		UpdatedAt:                baseTime.Add(2 * time.Minute),
		Steps: []runledger.Step{{
			StepID: "next",
			Goal:   "Should not override decision",
			Status: runledger.StepStatusPending,
		}},
	})

	assert.Equal(t, "Durable goal", durableView.Detail)
	assert.Equal(t, "teammate blocked", durableView.BlockedReason)
	assert.Equal(t, "waiting on specialist", durableView.RuntimeHint)
	assert.Equal(t, "Resolve pending decision", durableView.NextAction)
	assert.Equal(t, baseTime.Add(2*time.Minute), durableView.UpdatedAt)
}

func TestWave26MissionControlProjectorTaskAndAgentEnrichmentBranches(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2026, 5, 18, 9, 0, 0, 0, time.UTC)
	missionView := MissionView{ID: "mission-1", UpdatedAt: baseTime}
	enrichDurableMissionFromTask(&missionView, background.TaskSnapshot{
		Prompt:      "\n\nFirst useful line\nignored",
		StartedAt:   baseTime.Add(-time.Minute),
		CompletedAt: baseTime.Add(time.Minute),
		NextRetryAt: baseTime.Add(2 * time.Minute),
	})

	assert.Equal(t, "First useful line", missionView.Detail)
	assert.Equal(t, baseTime.Add(2*time.Minute), missionView.UpdatedAt)

	durableView := MissionView{
		ID:            "mission-2",
		BlockedReason: "existing blocker",
	}
	enrichDurableMissionFromAgentRun(&durableView, &agentrt.AgentRun{
		RequestedAgent:   "  reviewer\x1b[31m\nagent  ",
		RuntimeCondition: agentrt.AgentRunConditionBlockedWaitingMessage,
		BlockedReason:    "new blocker",
	})

	assert.Equal(t, "reviewer agent", durableView.OwnerAgent)
	assert.Equal(t, "Waiting for message", durableView.RuntimeHint)
	assert.Equal(t, "existing blocker", durableView.BlockedReason)

	overlayView := MissionView{
		ID:         "bg:task-2",
		Status:     MissionStatusRunning,
		NextAction: "Monitor background execution",
	}
	enrichMissionFromAgentRun(&overlayView, &agentrt.AgentRun{
		RequestedAgent:   "operator",
		RuntimeCondition: agentrt.AgentRunConditionBlockedWaitingMessage,
	})

	assert.Equal(t, "operator", overlayView.OwnerAgent)
	assert.Equal(t, MissionStatusBlocked, overlayView.Status)
	assert.Equal(t, "Waiting for message", overlayView.RuntimeHint)
	assert.Equal(t, "Waiting for message", overlayView.BlockedReason)
	assert.Equal(t, "Respond to required message", overlayView.NextAction)
}

func TestWave26MissionControlProjectorProposalProjectionBranches(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		ProposalReader: stubMissionControlProposalReader{
			items: map[string][]proposal.Proposal{
				"sess-1": {
					{
						ProposalID: "preparing",
						SessionKey: "sess-1",
						Status:     proposal.ProposalStatusPreparing,
						Title:      "  ",
						Summary:    "Use summary when no prepared brief exists.",
						Source: proposal.ProposalSource{
							Kind: "learning",
							Ref:  "suggestion-1",
						},
						UpdatedAt: time.Date(2026, 5, 18, 9, 2, 0, 0, time.UTC),
					},
					{
						ProposalID: "accepted",
						SessionKey: "sess-1",
						Status:     proposal.ProposalStatusAccepted,
						Title:      "Accepted terminal proposal",
						Source: proposal.ProposalSource{
							Kind: "manual",
							Ref:  "operator",
						},
						UpdatedAt: time.Date(2026, 5, 18, 9, 1, 0, 0, time.UTC),
					},
				},
			},
		},
	})

	snapshot := projector.Project(nil)

	require.Len(t, snapshot.Missions, 2)
	assert.Equal(t, "preparing", snapshot.Missions[0].ID)
	assert.Equal(t, MissionKindProposed, snapshot.Missions[0].Kind)
	assert.Equal(t, MissionStatusPending, snapshot.Missions[0].Status)
	assert.Equal(t, "Inspect proposal", snapshot.Missions[0].Title)
	assert.Equal(t, "Use summary when no prepared brief exists.", snapshot.Missions[0].Detail)
	assert.Equal(t, "Finish proposal preparation", snapshot.Missions[0].NextAction)
	assert.Equal(t, "learning", snapshot.Missions[0].RuntimeHint)
	assert.Equal(t, "suggestion-1", snapshot.Missions[0].SourceRef)

	assert.Equal(t, "accepted", snapshot.Missions[1].ID)
	assert.Equal(t, MissionStatusUnknown, snapshot.Missions[1].Status)
	assert.Equal(t, "Accepted terminal proposal", snapshot.Missions[1].Title)
	assert.Equal(t, "Review proposal", snapshot.Missions[1].NextAction)
	assert.Equal(t, "manual", snapshot.Missions[1].RuntimeHint)
}

func TestWave26MissionControlProjectorDurableRowStatusAndDetailBranches(t *testing.T) {
	t.Parallel()

	description := "  Durable description  "
	decisionSummary := "Decision summary fallback"
	sourceRef := "  source-ref  "
	blockedReason := "  waiting for credentials  "

	tests := []struct {
		name       string
		giveRow    *mission.Mission
		wantStatus MissionStatus
		wantAction string
		wantDetail string
		wantHint   string
	}{
		{
			name: "prepared",
			giveRow: &mission.Mission{
				Status:      mission.StatusPrepared,
				Title:       "Prepared",
				Description: &description,
			},
			wantStatus: MissionStatusPrepared,
			wantAction: "Start mission",
			wantDetail: "Durable description",
		},
		{
			name: "blocked",
			giveRow: &mission.Mission{
				Status:               mission.StatusBlocked,
				Title:                "Blocked",
				CurrentBlockedReason: &blockedReason,
			},
			wantStatus: MissionStatusBlocked,
			wantAction: "Resolve blocker",
		},
		{
			name: "waiting decision",
			giveRow: &mission.Mission{
				Status:                 mission.StatusWaitingDecision,
				Title:                  "Waiting",
				CurrentDecisionSummary: &decisionSummary,
			},
			wantStatus: MissionStatusPending,
			wantAction: "Resolve pending decision",
			wantDetail: "Decision summary fallback",
			wantHint:   "Waiting for decision",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.giveRow.SourceKind = "user"
			tt.giveRow.SourceRef = &sourceRef
			view := missionViewFromDurableRow(tt.giveRow)
			assert.Equal(t, tt.wantStatus, view.Status)
			assert.Equal(t, tt.wantAction, view.NextAction)
			assert.Equal(t, tt.wantDetail, view.Detail)
			assert.Equal(t, tt.wantHint, view.RuntimeHint)
			assert.Equal(t, "source-ref", view.SourceRef)
			assert.Equal(t, "user", view.SourceKind)
		})
	}
}
