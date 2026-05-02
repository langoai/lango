package cockpit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/runledger"
)

type stubMissionControlRunLedgerReader struct {
	snapshots map[string]*runledger.RunSnapshot
	getErr    error
}

func (s stubMissionControlRunLedgerReader) ListRuns(context.Context, int) ([]runledger.RunSummary, error) {
	return nil, nil
}

func (s stubMissionControlRunLedgerReader) GetRunSnapshot(_ context.Context, runID string) (*runledger.RunSnapshot, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if snap, ok := s.snapshots[runID]; ok {
		return snap.DeepCopy(), nil
	}
	return nil, nil
}

func (s stubMissionControlRunLedgerReader) ListRunSummariesBySession(context.Context, string, int) ([]runledger.RunSummary, error) {
	return nil, nil
}

type stubMissionControlAgentRunReader struct {
	runs   map[string]*agentrt.AgentRun
	list   []*agentrt.AgentRun
	getErr error
}

func (s stubMissionControlAgentRunReader) Get(id string) (*agentrt.AgentRun, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if run, ok := s.runs[id]; ok {
		cp := *run
		return &cp, nil
	}
	return nil, nil
}

func (s stubMissionControlAgentRunReader) List() []*agentrt.AgentRun {
	if len(s.list) == 0 {
		return nil
	}
	out := make([]*agentrt.AgentRun, 0, len(s.list))
	for _, run := range s.list {
		cp := *run
		out = append(out, &cp)
	}
	return out
}

type stubMissionControlMissionReader struct {
	missions map[string][]*mission.Mission
	links    map[string][]*mission.ExecutionLink
	listErr  error
	linkErr  error
}

func (s stubMissionControlMissionReader) ListMissionsBySession(_ context.Context, sessionKey string, _ int) ([]*mission.Mission, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	rows := s.missions[sessionKey]
	out := make([]*mission.Mission, 0, len(rows))
	for _, row := range rows {
		cp := *row
		out = append(out, &cp)
	}
	return out, nil
}

func (s stubMissionControlMissionReader) ListExecutionLinks(_ context.Context, missionID string) ([]*mission.ExecutionLink, error) {
	if s.linkErr != nil {
		return nil, s.linkErr
	}
	rows := s.links[missionID]
	out := make([]*mission.ExecutionLink, 0, len(rows))
	for _, row := range rows {
		cp := *row
		out = append(out, &cp)
	}
	return out, nil
}

func TestMissionControlProjectorBackgroundTaskTitleDerivation(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{})
	now := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)

	snapshot := projector.Project([]background.TaskSnapshot{
		{
			ID:          "task-1",
			StatusText:  "running",
			Prompt:      "\n\nShip the mission control projector\nwith deterministic output\n",
			StartedAt:   now.Add(-3 * time.Minute),
			NextRetryAt: now.Add(-1 * time.Minute),
		},
	})

	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, "bg:task-1", snapshot.Missions[0].ID)
	assert.Equal(t, MissionKindActive, snapshot.Missions[0].Kind)
	assert.Equal(t, MissionStatusRunning, snapshot.Missions[0].Status)
	assert.Equal(t, "Ship the mission control projector", snapshot.Missions[0].Title)
	assert.Equal(t, "Monitor background execution", snapshot.Missions[0].NextAction)
	assert.Equal(t, now.Add(-1*time.Minute), snapshot.Missions[0].UpdatedAt)
}

func TestMissionControlProjectorApprovalDecisionDerivation(t *testing.T) {
	t.Parallel()

	registry := NewPendingApprovalRegistry()
	registry.Register(chat.ApprovalRequestMsg{
		Request: approval.ApprovalRequest{
			ID:        "req-7",
			ToolName:  "fs_write",
			Summary:   "Update mission-control copy to clarify degraded state.\nSecond line is ignored.",
			CreatedAt: time.Date(2026, 5, 3, 10, 5, 0, 0, time.UTC),
		},
		ViewModel: approval.ApprovalViewModel{
			RuleExplanation: "Filesystem writes require approval.",
			Risk: approval.RiskIndicator{
				Level: "critical",
				Label: "Modifies filesystem",
			},
		},
		Response: make(chan approval.ApprovalResponse, 1),
	})

	projector := NewMissionControlProjector(Deps{PendingApprovals: registry})
	snapshot := projector.Project(nil)

	require.NotNil(t, snapshot.Decision)
	assert.Equal(t, "req-7", snapshot.Decision.ID)
	assert.Equal(t, DecisionCategoryApproval, snapshot.Decision.Category)
	assert.Equal(t, "fs_write: Update mission-control copy to clarify degraded state.", snapshot.Decision.Title)
	assert.Equal(t, "Filesystem writes require approval.", snapshot.Decision.Reason)
	assert.Equal(t, "Update mission-control copy to clarify degraded state.\nSecond line is ignored.", snapshot.Decision.EffectText)
	assert.Equal(t, "critical", snapshot.Decision.RiskLevel)
	assert.Equal(t, "Modifies filesystem", snapshot.Decision.RiskLabel)
	assert.Equal(t, "Approve", snapshot.Decision.ApproveLabel)
	assert.Equal(t, "Deny", snapshot.Decision.DenyLabel)
	assert.Equal(t, "Allow for session", snapshot.Decision.AllowForSessionLabel)
}

func TestMissionControlProjectorLearningSuggestionDerivation(t *testing.T) {
	t.Parallel()

	buffer := NewLearningSuggestionBuffer(func() time.Time {
		return time.Date(2026, 5, 3, 10, 35, 0, 0, time.UTC)
	})
	buffer.Append(eventbus.LearningSuggestionEvent{
		SuggestionID: "older",
		ProposedRule: "Prefer deterministic mission ordering.",
		Rationale:    "Operators need stable ordering.",
		Timestamp:    time.Date(2026, 5, 3, 10, 10, 0, 0, time.UTC),
	})
	buffer.Append(eventbus.LearningSuggestionEvent{
		SuggestionID: "newer",
		ProposedRule: "Collapse mission overflow into a compact summary.",
		Rationale:    "The lane should stay scannable.",
		Timestamp:    time.Date(2026, 5, 3, 10, 30, 0, 0, time.UTC),
	})

	projector := NewMissionControlProjector(Deps{LearningBuffer: buffer})
	snapshot := projector.Project(nil)

	require.Len(t, snapshot.Missions, 2)
	assert.Equal(t, "learn:newer", snapshot.Missions[0].ID)
	assert.Equal(t, MissionKindProposed, snapshot.Missions[0].Kind)
	assert.Equal(t, MissionStatusPrepared, snapshot.Missions[0].Status)
	assert.Equal(t, "Apply learning rule: Collapse mission overflow into a compact summary.", snapshot.Missions[0].Title)
	assert.Equal(t, "proposed_learning", snapshot.Missions[0].SourceKind)
	assert.Equal(t, "newer", snapshot.Missions[0].SourceRef)
	assert.Equal(t, "learn:older", snapshot.Missions[1].ID)
	assert.Equal(t, "proposed_learning", snapshot.Missions[1].SourceKind)
	assert.Equal(t, "older", snapshot.Missions[1].SourceRef)
}

func TestMissionControlProjectorRunLedgerNextActionEnrichment(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	projector := NewMissionControlProjector(Deps{
		RunLedgerStore: stubMissionControlRunLedgerReader{
			snapshots: map[string]*runledger.RunSnapshot{
				"task-1": {
					RunID:     "task-1",
					Status:    runledger.RunStatusRunning,
					UpdatedAt: now.Add(-30 * time.Second),
					Steps: []runledger.Step{
						{
							StepID: "done",
							Goal:   "Collect runtime facts",
							Status: runledger.StepStatusCompleted,
						},
						{
							StepID:    "next",
							Goal:      "Render Mission Control lane",
							Status:    runledger.StepStatusPending,
							DependsOn: []string{"done"},
						},
					},
				},
			},
		},
	})

	snapshot := projector.Project([]background.TaskSnapshot{
		{
			ID:         "task-1",
			StatusText: "running",
			Prompt:     "Build the projector",
			StartedAt:  now.Add(-2 * time.Minute),
		},
	})

	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, "Next step: Render Mission Control lane", snapshot.Missions[0].NextAction)
	assert.Equal(t, now.Add(-30*time.Second), snapshot.Missions[0].UpdatedAt)
}

func TestMissionControlProjectorAgentRunBlockedEnrichment(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	projector := NewMissionControlProjector(Deps{
		AgentRunStore: stubMissionControlAgentRunReader{
			runs: map[string]*agentrt.AgentRun{
				"task-1": {
					ID:               "task-1",
					RequestedAgent:   "worker-c",
					Status:           agentrt.AgentRunRunning,
					RuntimeCondition: agentrt.AgentRunConditionBlockedWaitingApproval,
					BlockedReason:    "Awaiting operator approval for filesystem write",
				},
			},
		},
	})

	snapshot := projector.Project([]background.TaskSnapshot{
		{
			ID:         "task-1",
			StatusText: "running",
			Prompt:     "Apply patch safely",
			StartedAt:  now.Add(-2 * time.Minute),
		},
	})

	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, MissionStatusBlocked, snapshot.Missions[0].Status)
	assert.Equal(t, "worker-c", snapshot.Missions[0].OwnerAgent)
	assert.Equal(t, "Waiting for approval", snapshot.Missions[0].RuntimeHint)
	assert.Equal(t, "Awaiting operator approval for filesystem write", snapshot.Missions[0].BlockedReason)
	assert.Equal(t, "Resolve approval request", snapshot.Missions[0].NextAction)
}

func TestMissionControlDegradedNilReaders(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{})
	snapshot := projector.Project([]background.TaskSnapshot{{ID: "task-1", StatusText: "pending", Prompt: "Queue work"}})

	assert.True(t, snapshot.Degraded)
	assert.Contains(t, snapshot.Header.DegradedNote, "Mission store unavailable")
	assert.Contains(t, snapshot.Header.DegradedNote, "RunLedger unavailable")
	assert.Contains(t, snapshot.Header.DegradedNote, "Agent runtime unavailable")
	assert.Empty(t, snapshot.Missions[0].OwnerAgent)
	assert.Empty(t, snapshot.Missions[0].RuntimeHint)
}

func TestMissionControlDegradedReaderErrorsPreserveBaseMissionData(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{
		RunLedgerStore: stubMissionControlRunLedgerReader{getErr: context.DeadlineExceeded},
		AgentRunStore:  stubMissionControlAgentRunReader{getErr: context.Canceled},
	})

	snapshot := projector.Project([]background.TaskSnapshot{{
		ID:         "task-1",
		StatusText: "running",
		Prompt:     "Queue follow-up approval",
		StartedAt:  time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC),
	}})

	require.Len(t, snapshot.Missions, 1)
	assert.True(t, snapshot.Degraded)
	assert.Contains(t, snapshot.Header.DegradedNote, "Mission store unavailable")
	assert.Contains(t, snapshot.Header.DegradedNote, "RunLedger unavailable")
	assert.Contains(t, snapshot.Header.DegradedNote, "Agent runtime unavailable")
	assert.Equal(t, "bg:task-1", snapshot.Missions[0].ID)
	assert.Equal(t, "Queue follow-up approval", snapshot.Missions[0].Title)
	assert.Equal(t, MissionStatusRunning, snapshot.Missions[0].Status)
	assert.Empty(t, snapshot.Missions[0].OwnerAgent)
	assert.Empty(t, snapshot.Missions[0].RuntimeHint)
}

func TestMissionControlHeaderDerivation(t *testing.T) {
	t.Parallel()

	registry := NewPendingApprovalRegistry()
	registry.Register(chat.ApprovalRequestMsg{
		Request:  approval.ApprovalRequest{ID: "req-1"},
		Response: make(chan approval.ApprovalResponse, 1),
	})

	collector := observability.NewCollector()
	collector.RecordTokenUsage(observability.TokenUsage{
		SessionKey:   "sess-1",
		InputTokens:  120,
		OutputTokens: 30,
		TotalTokens:  150,
	})

	agentRun := &agentrt.AgentRun{
		ID:             "task-1",
		RequestedAgent: "worker-c",
		Status:         agentrt.AgentRunRunning,
		CreatedAt:      time.Date(2026, 5, 3, 9, 0, 0, 0, time.UTC),
	}

	projector := NewMissionControlProjector(Deps{
		Config: &config.Config{
			Agent: config.AgentConfig{
				Provider: "openai",
				Model:    "gpt-5",
			},
		},
		SessionKey:       "sess-1",
		MetricsCollector: collector,
		PendingApprovals: registry,
		AgentRunStore: stubMissionControlAgentRunReader{
			runs: map[string]*agentrt.AgentRun{"task-1": agentRun},
			list: []*agentrt.AgentRun{agentRun},
		},
		RunLedgerStore: stubMissionControlRunLedgerReader{},
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{
				"sess-1": {},
			},
		},
	})

	snapshot := projector.Project([]background.TaskSnapshot{
		{
			ID:         "task-1",
			StatusText: "running",
			Prompt:     "Build the projector",
			StartedAt:  time.Date(2026, 5, 3, 9, 5, 0, 0, time.UTC),
		},
	})

	assert.Equal(t, "worker-c active", snapshot.Header.ActiveAgentSummary)
	assert.Equal(t, "openai / gpt-5", snapshot.Header.ModelProviderSummary)
	assert.Equal(t, 1, snapshot.Header.PendingDecisionCount)
	assert.Empty(t, snapshot.Header.ContextSummary)
	assert.Equal(t, "150 tokens across 1 requests", snapshot.Header.MetricsSummary)
	assert.Empty(t, snapshot.Header.DegradedNote)
}

func TestMissionControlHeaderOmitsGlobalAgentRunSummaryWithoutProjectedMission(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{
		AgentRunStore: stubMissionControlAgentRunReader{
			list: []*agentrt.AgentRun{
				{
					ID:             "foreign-task",
					RequestedAgent: "worker-z",
					Status:         agentrt.AgentRunRunning,
					CreatedAt:      time.Date(2026, 5, 3, 9, 0, 0, 0, time.UTC),
				},
			},
		},
	})

	snapshot := projector.Project(nil)
	assert.Empty(t, snapshot.Header.ActiveAgentSummary)
}

func TestMissionControlOrdering(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{})
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)

	buffer := NewLearningSuggestionBuffer(func() time.Time { return now })
	buffer.Append(eventbus.LearningSuggestionEvent{
		SuggestionID: "learn-1",
		ProposedRule: "Add footer shortcuts.",
		Timestamp:    now.Add(-1 * time.Minute),
	})
	projector.learningBuffer = buffer

	snapshot := projector.Project([]background.TaskSnapshot{
		{ID: "done", StatusText: "done", Prompt: "Done mission", CompletedAt: now.Add(-2 * time.Minute)},
		{ID: "running", StatusText: "running", Prompt: "Running mission", StartedAt: now.Add(-4 * time.Minute)},
		{ID: "pending", StatusText: "pending", Prompt: "Pending mission", NextRetryAt: now.Add(-3 * time.Minute)},
	})

	require.Len(t, snapshot.Missions, 4)
	assert.Equal(t, "bg:running", snapshot.Missions[0].ID)
	assert.Equal(t, "bg:pending", snapshot.Missions[1].ID)
	assert.Equal(t, "bg:done", snapshot.Missions[2].ID)
	assert.Equal(t, "learn:learn-1", snapshot.Missions[3].ID)
}

func TestMissionControlDurableMissionsRenderFirst(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	durableID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{
				"sess-1": {
					{
						ID:         durableID,
						SessionKey: "sess-1",
						Title:      "Durable mission",
						Status:     mission.StatusActive,
						SourceKind: "user",
						UpdatedAt:  now,
						CreatedAt:  now.Add(-time.Minute),
					},
				},
			},
		},
	})

	snapshot := projector.Project([]background.TaskSnapshot{
		{ID: "task-1", StatusText: "running", Prompt: "Overlay mission", OriginSession: "sess-1", StartedAt: now.Add(-2 * time.Minute)},
	})

	require.Len(t, snapshot.Missions, 2)
	assert.Equal(t, durableID.String(), snapshot.Missions[0].ID)
	assert.Equal(t, MissionKindActive, snapshot.Missions[0].Kind)
	assert.Equal(t, "bg:task-1", snapshot.Missions[1].ID)
}

func TestMissionControlLinkedTaskEnrichesDurableMissionWithoutDuplicateOverlay(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	missionID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{
				"sess-1": {{
					ID:         missionID,
					SessionKey: "sess-1",
					Title:      "Durable mission",
					Status:     mission.StatusActive,
					SourceKind: "user",
					UpdatedAt:  now.Add(-time.Minute),
					CreatedAt:  now.Add(-2 * time.Minute),
				}},
			},
			links: map[string][]*mission.ExecutionLink{
				missionID.String(): {{
					MissionID:     missionID,
					ExecutionKind: mission.ExecutionKindTaskOSExecution,
					ExecutionRef:  "task-1",
				}},
			},
		},
		RunLedgerStore: stubMissionControlRunLedgerReader{
			snapshots: map[string]*runledger.RunSnapshot{
				"task-1": {
					RunID:     "task-1",
					Goal:      "Execute durable work",
					UpdatedAt: now,
					Steps: []runledger.Step{{
						StepID: "next",
						Goal:   "Render summary",
						Status: runledger.StepStatusPending,
					}},
				},
			},
		},
		AgentRunStore: stubMissionControlAgentRunReader{
			runs: map[string]*agentrt.AgentRun{
				"task-1": {
					ID:             "task-1",
					RequestedAgent: "worker-c",
				},
			},
		},
	})

	snapshot := projector.Project([]background.TaskSnapshot{
		{ID: "task-1", StatusText: "running", Prompt: "Overlay mission", OriginSession: "sess-1", StartedAt: now.Add(-3 * time.Minute)},
	})

	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, missionID.String(), snapshot.Missions[0].ID)
	assert.Equal(t, "worker-c", snapshot.Missions[0].OwnerAgent)
	assert.Equal(t, "Next step: Render summary", snapshot.Missions[0].NextAction)
}

func TestMissionControlUnmatchedRuntimeOverlayStillRendersWhenUnlinked(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{"sess-1": {}},
		},
	})

	snapshot := projector.Project([]background.TaskSnapshot{
		{ID: "task-1", StatusText: "running", Prompt: "Overlay mission", OriginSession: "sess-1"},
	})

	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, "bg:task-1", snapshot.Missions[0].ID)
}

func TestMissionControlOverlayFilteringUsesCurrentSession(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{"sess-1": {}},
		},
	})

	snapshot := projector.Project([]background.TaskSnapshot{
		{ID: "task-1", StatusText: "running", Prompt: "Current session", OriginSession: "sess-1"},
		{ID: "task-2", StatusText: "running", Prompt: "Foreign session", OriginSession: "sess-2"},
		{ID: "task-3", StatusText: "running", Prompt: "Legacy empty origin"},
	})

	require.Len(t, snapshot.Missions, 2)
	assert.Equal(t, "bg:task-1", snapshot.Missions[0].ID)
	assert.Equal(t, "bg:task-3", snapshot.Missions[1].ID)
}

func TestMissionControlDurableWaitingDecisionAndLiveApprovalRemainCoherent(t *testing.T) {
	t.Parallel()

	registry := NewPendingApprovalRegistry()
	registry.Register(chat.ApprovalRequestMsg{
		Request:  approval.ApprovalRequest{ID: "req-1", ToolName: "fs_write", Summary: "Write files"},
		Response: make(chan approval.ApprovalResponse, 1),
	})
	missionID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	summary := "Awaiting approval for filesystem write"
	kind := "approval"
	projector := NewMissionControlProjector(Deps{
		SessionKey:       "sess-1",
		PendingApprovals: registry,
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{
				"sess-1": {{
					ID:                     missionID,
					SessionKey:             "sess-1",
					Title:                  "Durable waiting mission",
					Status:                 mission.StatusWaitingDecision,
					SourceKind:             "user",
					CurrentDecisionKind:    &kind,
					CurrentDecisionSummary: &summary,
					UpdatedAt:              time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
					CreatedAt:              time.Date(2026, 5, 3, 11, 0, 0, 0, time.UTC),
				}},
			},
		},
	})

	snapshot := projector.Project(nil)

	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, MissionStatusPending, snapshot.Missions[0].Status)
	assert.Equal(t, "Resolve pending decision", snapshot.Missions[0].NextAction)
	assert.Equal(t, "Waiting for decision", snapshot.Missions[0].RuntimeHint)
	require.NotNil(t, snapshot.Decision)
	assert.Equal(t, "req-1", snapshot.Decision.ID)
}

func TestMissionControlMissionReaderUnavailableDegradesTruthfully(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{SessionKey: "sess-1"})
	snapshot := projector.Project([]background.TaskSnapshot{
		{ID: "task-1", StatusText: "running", Prompt: "Overlay mission", OriginSession: "sess-1"},
	})

	assert.True(t, snapshot.Degraded)
	assert.Contains(t, snapshot.Header.DegradedNote, "Mission store unavailable")
	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, "bg:task-1", snapshot.Missions[0].ID)
}

func TestMissionControlMissionReaderErrorsDegradeTruthfully(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{
		SessionKey:    "sess-1",
		MissionReader: stubMissionControlMissionReader{listErr: errors.New("db unavailable")},
	})
	snapshot := projector.Project([]background.TaskSnapshot{
		{ID: "task-1", StatusText: "running", Prompt: "Overlay mission", OriginSession: "sess-1"},
	})

	assert.True(t, snapshot.Degraded)
	assert.Contains(t, snapshot.Header.DegradedNote, "Mission store unavailable")
	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, "bg:task-1", snapshot.Missions[0].ID)
}

func TestMissionControlOverflow(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{})
	projector.missionLimit = 2
	projector.activityLimit = 2

	now := time.Date(2026, 5, 3, 13, 0, 0, 0, time.UTC)
	activity := NewMissionActivityBuffer()
	activity.Append(MissionActivityItem{Summary: "First", Timestamp: now.Add(-3 * time.Minute)})
	activity.Append(MissionActivityItem{Summary: "Second", Timestamp: now.Add(-2 * time.Minute)})
	activity.Append(MissionActivityItem{Summary: "Third", Timestamp: now.Add(-1 * time.Minute)})
	projector.activityBuffer = activity

	snapshot := projector.Project([]background.TaskSnapshot{
		{ID: "one", StatusText: "running", Prompt: "One", StartedAt: now.Add(-1 * time.Minute)},
		{ID: "two", StatusText: "pending", Prompt: "Two", StartedAt: now.Add(-2 * time.Minute)},
		{ID: "three", StatusText: "done", Prompt: "Three", CompletedAt: now.Add(-3 * time.Minute)},
	})

	require.Len(t, snapshot.Missions, 2)
	assert.Equal(t, 1, snapshot.HiddenMissionCount)
	assert.Equal(t, "1 more mission", snapshot.MissionOverflowSummary)
	require.Len(t, snapshot.Activities, 2)
	assert.Equal(t, 1, snapshot.HiddenActivityCount)
	assert.Equal(t, "1 more activity item", snapshot.ActivityOverflowSummary)
	assert.Equal(t, "Third", snapshot.Activities[0].Summary)
	assert.Equal(t, "Second", snapshot.Activities[1].Summary)
}
