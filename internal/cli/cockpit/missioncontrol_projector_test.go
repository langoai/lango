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
	apppkg "github.com/langoai/lango/internal/app"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/cron"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/librarian"
	"github.com/langoai/lango/internal/loopview"
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/postadjudicationstatus"
	"github.com/langoai/lango/internal/proposal"
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

type stubMissionControlProposalReader struct {
	items map[string][]proposal.Proposal
}

func (s stubMissionControlProposalReader) ListBySession(sessionKey string) []proposal.Proposal {
	items := s.items[sessionKey]
	out := make([]proposal.Proposal, len(items))
	copy(out, items)
	return out
}

func (s stubMissionControlProposalReader) ListLoopBySession(sessionKey string) []proposal.Proposal {
	return s.ListBySession(sessionKey)
}

func (s stubMissionControlProposalReader) GetByID(proposalID string) (proposal.Proposal, bool) {
	for _, items := range s.items {
		for _, item := range items {
			if item.ProposalID == proposalID {
				return item, true
			}
		}
	}
	return proposal.Proposal{}, false
}

type stubMissionControlLoopInquiryReader struct {
	items []librarian.Inquiry
	err   error
}

func (s stubMissionControlLoopInquiryReader) ListPendingInquiries(_ context.Context, sessionKey string, _ int) ([]librarian.Inquiry, error) {
	if s.err != nil {
		return nil, s.err
	}
	out := make([]librarian.Inquiry, 0, len(s.items))
	for _, item := range s.items {
		if item.SessionKey == sessionKey {
			out = append(out, item)
		}
	}
	return out, nil
}

type stubMissionControlLoopDeadReader struct {
	items []postadjudicationstatus.DeadLetterBacklogEntry
	err   error
}

func (s stubMissionControlLoopDeadReader) ListCurrentDeadLetters(context.Context) ([]postadjudicationstatus.DeadLetterBacklogEntry, error) {
	if s.err != nil {
		return nil, s.err
	}
	return append([]postadjudicationstatus.DeadLetterBacklogEntry(nil), s.items...), nil
}

type stubMissionControlLoopCronReader struct {
	items   []cron.Job
	history map[string][]cron.HistoryEntry
	err     error
}

func (s stubMissionControlLoopCronReader) List(context.Context) ([]cron.Job, error) {
	if s.err != nil {
		return nil, s.err
	}
	return append([]cron.Job(nil), s.items...), nil
}

func (s stubMissionControlLoopCronReader) ListHistory(_ context.Context, jobID string, limit int) ([]cron.HistoryEntry, error) {
	if s.err != nil {
		return nil, s.err
	}
	items := append([]cron.HistoryEntry(nil), s.history[jobID]...)
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

type stubMissionControlCollabMissionLinkReader struct {
	links map[string][]apppkg.CollaborationMissionExecutionLink
	err   error
}

func (s stubMissionControlCollabMissionLinkReader) ListMissionExecutionLinks(_ context.Context, missionID string) ([]apppkg.CollaborationMissionExecutionLink, error) {
	if s.err != nil {
		return nil, s.err
	}
	return append([]apppkg.CollaborationMissionExecutionLink(nil), s.links[missionID]...), nil
}

type stubMissionControlCollabAgentRunReader struct {
	runs []apppkg.CollaborationAgentRunView
}

func (s stubMissionControlCollabAgentRunReader) ListAgentRuns() []apppkg.CollaborationAgentRunView {
	return append([]apppkg.CollaborationAgentRunView(nil), s.runs...)
}

type stubMissionControlCollabDelegationReader struct {
	items []apppkg.CollaborationDelegationRecord
	err   error
}

func (s stubMissionControlCollabDelegationReader) ListDelegationsForSession(_ context.Context, sessionKey string) ([]apppkg.CollaborationDelegationRecord, error) {
	if s.err != nil {
		return nil, s.err
	}
	out := make([]apppkg.CollaborationDelegationRecord, 0, len(s.items))
	for _, item := range s.items {
		if item.SessionKey == sessionKey {
			out = append(out, item)
		}
	}
	return out, nil
}

type stubMissionControlCollabRuntimeReader struct {
	budget   map[string][]apppkg.CollaborationBudgetRecord
	recovery map[string][]apppkg.CollaborationRecoveryRecord
}

func (s stubMissionControlCollabRuntimeReader) ListBudgetSignals(missionID string) []apppkg.CollaborationBudgetRecord {
	return append([]apppkg.CollaborationBudgetRecord(nil), s.budget[missionID]...)
}

func (s stubMissionControlCollabRuntimeReader) ListRecoverySignals(missionID string) []apppkg.CollaborationRecoveryRecord {
	return append([]apppkg.CollaborationRecoveryRecord(nil), s.recovery[missionID]...)
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
	assert.Equal(t, MissionStatusPending, snapshot.Missions[0].Status)
	assert.Equal(t, "Apply learning rule: Collapse mission overflow into a compact summary.", snapshot.Missions[0].Title)
	assert.Equal(t, "proposed_learning", snapshot.Missions[0].SourceKind)
	assert.Equal(t, "newer", snapshot.Missions[0].SourceRef)
	assert.Equal(t, "learn:older", snapshot.Missions[1].ID)
	assert.Equal(t, "proposed_learning", snapshot.Missions[1].SourceKind)
	assert.Equal(t, "older", snapshot.Missions[1].SourceRef)
}

func TestMissionControlProposalRegistryPreparedProposalRendersFirstClass(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		ProposalReader: stubMissionControlProposalReader{
			items: map[string][]proposal.Proposal{
				"sess-1": {
					{
						ProposalID: "proposal-1",
						SessionKey: "sess-1",
						Source: proposal.ProposalSource{
							Kind: "proposed_learning",
							Ref:  "suggestion-1",
						},
						Title:  "Apply bounded retry guidance",
						Status: proposal.ProposalStatusPrepared,
						PreparedBrief: &proposal.PreparedBrief{
							SourceSummary:             "Learning suggestion: repeated timeout retries.",
							Reason:                    "Repeated timeout failures benefited from bounded retry.",
							SuggestedAcceptanceEffect: "Create a mission to apply bounded retry guidance.",
						},
						UpdatedAt: time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
						CreatedAt: time.Date(2026, 5, 3, 11, 0, 0, 0, time.UTC),
					},
				},
			},
		},
	})

	snapshot := projector.Project(nil)

	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, "proposal-1", snapshot.Missions[0].ID)
	assert.Equal(t, MissionKindProposed, snapshot.Missions[0].Kind)
	assert.Equal(t, MissionStatusPrepared, snapshot.Missions[0].Status)
	assert.Equal(t, "Apply bounded retry guidance", snapshot.Missions[0].Title)
	assert.Equal(t, "Learning suggestion: repeated timeout retries.", snapshot.Missions[0].Detail)
	assert.Equal(t, "Review prepared proposal", snapshot.Missions[0].NextAction)
	assert.Equal(t, "proposed_learning", snapshot.Missions[0].SourceKind)
	assert.Equal(t, "suggestion-1", snapshot.Missions[0].SourceRef)
	assert.Equal(t, "Repeated timeout failures benefited from bounded retry.", snapshot.Missions[0].RuntimeHint)
}

func TestMissionControlProposalRegistrySessionScoping(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		ProposalReader: stubMissionControlProposalReader{
			items: map[string][]proposal.Proposal{
				"sess-1": {{ProposalID: "proposal-1", SessionKey: "sess-1", Title: "Visible", Status: proposal.ProposalStatusPrepared}},
				"sess-2": {{ProposalID: "proposal-2", SessionKey: "sess-2", Title: "Hidden", Status: proposal.ProposalStatusPrepared}},
			},
		},
	})

	snapshot := projector.Project(nil)

	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, "proposal-1", snapshot.Missions[0].ID)
}

func TestMissionControlProposalRegistryUnavailableFallsBackHonestly(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 3, 10, 35, 0, 0, time.UTC)
	buffer := NewLearningSuggestionBuffer(func() time.Time { return now })
	buffer.Append(eventbus.LearningSuggestionEvent{
		SuggestionID: "fallback",
		ProposedRule: "Keep fallback honest.",
		Rationale:    "Proposal registry unavailable.",
		Timestamp:    now,
	})

	projector := NewMissionControlProjector(Deps{
		SessionKey:     "sess-1",
		LearningBuffer: buffer,
	})

	snapshot := projector.Project(nil)

	assert.True(t, snapshot.Degraded)
	assert.Contains(t, snapshot.Header.DegradedNote, "Proposal registry unavailable")
	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, "learn:fallback", snapshot.Missions[0].ID)
	assert.Equal(t, MissionStatusPending, snapshot.Missions[0].Status)
	assert.Equal(t, "Review raw suggestion", snapshot.Missions[0].NextAction)
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

func TestMissionControlMissionLinkErrorsDegradeWithoutClaimingStoreUnavailable(t *testing.T) {
	t.Parallel()

	missionID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
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
					UpdatedAt:  time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
					CreatedAt:  time.Date(2026, 5, 3, 11, 0, 0, 0, time.UTC),
				}},
			},
			linkErr: errors.New("link query failed"),
		},
	})

	snapshot := projector.Project(nil)

	assert.True(t, snapshot.Degraded)
	assert.Contains(t, snapshot.Header.DegradedNote, "Mission details unavailable")
	assert.NotContains(t, snapshot.Header.DegradedNote, "Mission store unavailable")
	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, missionID.String(), snapshot.Missions[0].ID)
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

func TestMissionControlLoopRowsRenderFromRealSources(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{
				"sess-1": {{
					ID:         uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
					SessionKey: "sess-1",
					Title:      "Wait for approval",
					Status:     mission.StatusWaitingDecision,
					SourceKind: "user",
					UpdatedAt:  now.Add(-5 * time.Minute),
					CreatedAt:  now.Add(-time.Hour),
				}},
			},
		},
		LoopInquiryReader: stubMissionControlLoopInquiryReader{
			items: []librarian.Inquiry{{
				ID:         uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
				SessionKey: "sess-1",
				Topic:      "Need user answer",
				Question:   "Which vendor should we use?",
				CreatedAt:  now.Add(-25 * time.Hour),
			}},
		},
		LoopCronReader: stubMissionControlLoopCronReader{
			items: []cron.Job{{
				ID:        "cron-1",
				Name:      "Nightly digest",
				Enabled:   true,
				NextRunAt: ptrTime(now.Add(2 * time.Hour)),
			}},
			history: map[string][]cron.HistoryEntry{
				"cron-1": {{
					JobID:     "cron-1",
					JobName:   "Nightly digest",
					Status:    "failed",
					StartedAt: now.Add(-time.Hour),
				}},
			},
		},
	})

	snapshot := projector.Project(nil)

	require.Len(t, snapshot.Loops, 3)
	assert.Equal(t, "mission:aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", snapshot.Loops[0].ID)
	assert.Equal(t, loopview.LoopKindMissionCluster, snapshot.Loops[0].Kind)
	assert.Equal(t, loopview.LoopKindInquiry, snapshot.Loops[1].Kind)
	assert.Equal(t, loopview.LoopKindScheduledAutomation, snapshot.Loops[2].Kind)
	assert.Equal(t, loopview.LoopStatusBlocked, snapshot.Loops[2].Status)
	assert.Equal(t, 3, snapshot.OpenLoopCount)
}

func TestMissionControlAgendaOrderingIsDeterministic(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{
				"sess-1": {
					{
						ID:         uuid.MustParse("11111111-1111-1111-1111-111111111111"),
						SessionKey: "sess-1",
						Title:      "Active mission",
						Status:     mission.StatusActive,
						SourceKind: "user",
						UpdatedAt:  now.Add(-3 * time.Minute),
						CreatedAt:  now.Add(-time.Hour),
					},
					{
						ID:         uuid.MustParse("22222222-2222-2222-2222-222222222222"),
						SessionKey: "sess-1",
						Title:      "Blocked mission",
						Status:     mission.StatusBlocked,
						SourceKind: "user",
						UpdatedAt:  now.Add(-2 * time.Minute),
						CreatedAt:  now.Add(-time.Hour),
					},
				},
			},
		},
		LoopCronReader: stubMissionControlLoopCronReader{
			items: []cron.Job{{
				ID:        "cron-1",
				Name:      "Nightly digest",
				Enabled:   true,
				NextRunAt: ptrTime(now.Add(2 * time.Hour)),
			}},
			history: map[string][]cron.HistoryEntry{
				"cron-1": {{
					JobID:     "cron-1",
					JobName:   "Nightly digest",
					Status:    "completed",
					StartedAt: now.Add(-time.Hour),
				}},
			},
		},
	})

	snapshot := projector.Project(nil)

	require.Len(t, snapshot.Loops, 3)
	assert.Equal(t, "mission:22222222-2222-2222-2222-222222222222", snapshot.Loops[0].ID)
	assert.Equal(t, "mission:11111111-1111-1111-1111-111111111111", snapshot.Loops[1].ID)
	assert.Equal(t, "cron:cron-1", snapshot.Loops[2].ID)
}

func TestMissionControlAbsentScheduledSourceDoesNotFabricateLoops(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{
		SessionKey:    "sess-1",
		MissionReader: stubMissionControlMissionReader{missions: map[string][]*mission.Mission{"sess-1": {}}},
	})

	snapshot := projector.Project(nil)

	for _, loop := range snapshot.Loops {
		assert.NotEqual(t, loopview.LoopKindScheduledAutomation, loop.Kind)
	}
}

func TestMissionControlLoopCronUsesRealLatestExecutionOutcome(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	projector := NewMissionControlProjector(Deps{
		SessionKey:    "sess-1",
		MissionReader: stubMissionControlMissionReader{missions: map[string][]*mission.Mission{"sess-1": {}}},
		LoopCronReader: stubMissionControlLoopCronReader{
			items: []cron.Job{{
				ID:        "cron-1",
				Name:      "Nightly digest",
				Enabled:   true,
				NextRunAt: ptrTime(now.Add(2 * time.Hour)),
			}},
			history: map[string][]cron.HistoryEntry{
				"cron-1": {{
					JobID:     "cron-1",
					JobName:   "Nightly digest",
					Status:    "failed",
					StartedAt: now.Add(-30 * time.Minute),
				}},
			},
		},
	})

	snapshot := projector.Project(nil)

	require.Len(t, snapshot.Loops, 1)
	assert.Equal(t, "cron:cron-1", snapshot.Loops[0].ID)
	assert.Equal(t, loopview.LoopStatusBlocked, snapshot.Loops[0].Status)
	assert.Equal(t, "Review failed cron run", snapshot.Loops[0].NextAction)
}

func TestMissionControlAcceptedProposalFollowUpLoopAppearsFromIntegratedReader(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	projector := NewMissionControlProjector(Deps{
		SessionKey:    "sess-1",
		MissionReader: stubMissionControlMissionReader{missions: map[string][]*mission.Mission{"sess-1": {}}},
		ProposalReader: stubMissionControlProposalReader{
			items: map[string][]proposal.Proposal{
				"sess-1": {{
					ProposalID: "proposal-accepted",
					SessionKey: "sess-1",
					Title:      "Accepted proposal follow-up",
					Status:     proposal.ProposalStatusAccepted,
					UpdatedAt:  now.Add(-11 * time.Minute),
				}},
			},
		},
	})
	projector.nowFn = func() time.Time { return now }

	snapshot := projector.Project(nil)

	require.Len(t, snapshot.Loops, 1)
	assert.Equal(t, "follow-up:proposal:proposal-accepted", snapshot.Loops[0].ID)
	assert.Equal(t, loopview.LoopKindFollowUp, snapshot.Loops[0].Kind)
	assert.Equal(t, loopview.LoopStatusActive, snapshot.Loops[0].Status)
}

func TestMissionControlRecentDoneMissionReviewFollowUpAppears(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{
				"sess-1": {{
					ID:          uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"),
					SessionKey:  "sess-1",
					Title:       "Done mission needing review",
					Status:      mission.StatusDone,
					SourceKind:  "user",
					UpdatedAt:   now.Add(-2 * time.Hour),
					CreatedAt:   now.Add(-24 * time.Hour),
					CompletedAt: ptrTime(now.Add(-2 * time.Hour)),
				}},
			},
		},
	})
	projector.nowFn = func() time.Time { return now }

	snapshot := projector.Project(nil)

	var found bool
	for _, loop := range snapshot.Loops {
		if loop.ID != "follow-up:mission:dddddddd-dddd-dddd-dddd-dddddddddddd" {
			continue
		}
		found = true
		assert.Equal(t, loopview.LoopStatusNeedsReview, loop.Status)
		assert.Equal(t, "Review completed mission", loop.NextAction)
	}
	assert.True(t, found, "expected recent done mission review follow-up loop")
}

func TestMissionControlCollaborationParticipantSummaryFromLinkedLocalSignals(t *testing.T) {
	t.Parallel()

	missionID := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{
				"sess-1": {{
					ID:         missionID,
					SessionKey: "sess-1",
					Title:      "Collaborative mission",
					Status:     mission.StatusActive,
					SourceKind: "user",
					UpdatedAt:  time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
					CreatedAt:  time.Date(2026, 5, 3, 11, 0, 0, 0, time.UTC),
				}},
			},
		},
		CollabMissionLinks: stubMissionControlCollabMissionLinkReader{
			links: map[string][]apppkg.CollaborationMissionExecutionLink{
				missionID.String(): {{ExecutionKind: "task_os_execution", ExecutionRef: "exec-1"}},
			},
		},
		CollabAgentRuns: stubMissionControlCollabAgentRunReader{
			runs: []apppkg.CollaborationAgentRunView{{ID: "exec-1", RequestedAgent: "researcher"}},
		},
	})

	snapshot := projector.Project(nil)

	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, "researcher", snapshot.Missions[0].Collaboration.ParticipantSummary)
}

func TestMissionControlCollaborationHandoffSummaryOnlyWhenAttributable(t *testing.T) {
	t.Parallel()

	missionID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{
				"sess-1": {{
					ID:         missionID,
					SessionKey: "sess-1",
					Title:      "Handoff mission",
					Status:     mission.StatusActive,
					SourceKind: "user",
					UpdatedAt:  time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
					CreatedAt:  time.Date(2026, 5, 3, 11, 0, 0, 0, time.UTC),
				}},
			},
		},
		CollabMissionLinks: stubMissionControlCollabMissionLinkReader{
			links: map[string][]apppkg.CollaborationMissionExecutionLink{
				missionID.String(): {{ExecutionKind: "task_os_execution", ExecutionRef: "exec-1"}},
			},
		},
		CollabDelegations: stubMissionControlCollabDelegationReader{
			items: []apppkg.CollaborationDelegationRecord{
				{SessionKey: "sess-1", ExecutionRef: "exec-1", From: "planner", To: "researcher", Timestamp: time.Date(2026, 5, 3, 12, 5, 0, 0, time.UTC)},
				{SessionKey: "sess-1", ExecutionRef: "other", From: "planner", To: "writer", Timestamp: time.Date(2026, 5, 3, 12, 6, 0, 0, time.UTC)},
			},
		},
	})

	snapshot := projector.Project(nil)
	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, "planner -> researcher", snapshot.Missions[0].Collaboration.HandoffSummary)
}

func TestMissionControlCollaborationStateHintsAndBudgetRecoveryAreAttributable(t *testing.T) {
	t.Parallel()

	missionID := uuid.MustParse("12121212-1212-1212-1212-121212121212")
	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{
				"sess-1": {{
					ID:         missionID,
					SessionKey: "sess-1",
					Title:      "Recovery mission",
					Status:     mission.StatusActive,
					SourceKind: "user",
					UpdatedAt:  time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
					CreatedAt:  time.Date(2026, 5, 3, 11, 0, 0, 0, time.UTC),
				}},
			},
		},
		CollabMissionLinks: stubMissionControlCollabMissionLinkReader{
			links: map[string][]apppkg.CollaborationMissionExecutionLink{
				missionID.String(): {{ExecutionKind: "task_os_execution", ExecutionRef: "exec-1"}},
			},
		},
		CollabAgentRuns: stubMissionControlCollabAgentRunReader{
			runs: []apppkg.CollaborationAgentRunView{{ID: "exec-1", RequestedAgent: "researcher", RuntimeCondition: "recovering"}},
		},
		CollabRuntime: stubMissionControlCollabRuntimeReader{
			budget: map[string][]apppkg.CollaborationBudgetRecord{
				missionID.String(): {{MissionID: missionID.String(), Used: 9, Max: 10, Timestamp: time.Date(2026, 5, 3, 12, 10, 0, 0, time.UTC)}},
			},
			recovery: map[string][]apppkg.CollaborationRecoveryRecord{
				missionID.String(): {{MissionID: missionID.String(), Action: "retry", CauseClass: "rate_limit", Timestamp: time.Date(2026, 5, 3, 12, 11, 0, 0, time.UTC)}},
			},
		},
	})

	snapshot := projector.Project(nil)
	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, "Recovering", snapshot.Missions[0].Collaboration.StateHint)
	assert.Contains(t, snapshot.Missions[0].Collaboration.BudgetHint, "9/10")
	assert.Contains(t, snapshot.Missions[0].Collaboration.RecoveryHint, "retry")
}

func TestMissionControlCollaborationReviewingFromLinkedRunExecution(t *testing.T) {
	t.Parallel()

	missionID := uuid.MustParse("56565656-5656-5656-5656-565656565656")
	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{
				"sess-1": {{
					ID:         missionID,
					SessionKey: "sess-1",
					Title:      "Review mission",
					Status:     mission.StatusActive,
					SourceKind: "user",
					UpdatedAt:  time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
					CreatedAt:  time.Date(2026, 5, 3, 11, 0, 0, 0, time.UTC),
				}},
			},
		},
		CollabMissionLinks: stubMissionControlCollabMissionLinkReader{
			links: map[string][]apppkg.CollaborationMissionExecutionLink{
				missionID.String(): {{ExecutionKind: "runledger_run", ExecutionRef: "run-1"}},
			},
		},
		RunLedgerStore: stubMissionControlRunLedgerReader{
			snapshots: map[string]*runledger.RunSnapshot{
				"run-1": {
					RunID:         "run-1",
					CurrentStepID: "step-1",
					Steps: []runledger.Step{{
						StepID: "step-1",
						Status: runledger.StepStatusVerifyPending,
					}},
					UpdatedAt: time.Date(2026, 5, 3, 12, 10, 0, 0, time.UTC),
				},
			},
		},
	})

	snapshot := projector.Project(nil)
	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, "Reviewing", snapshot.Missions[0].Collaboration.StateHint)
}

func TestMissionControlCollaborationActiveOwnerUsesLatestRealRunTimestamp(t *testing.T) {
	t.Parallel()

	missionID := uuid.MustParse("78787878-7878-7878-7878-787878787878")
	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{
				"sess-1": {{
					ID:         missionID,
					SessionKey: "sess-1",
					Title:      "Ownership mission",
					Status:     mission.StatusActive,
					SourceKind: "user",
					UpdatedAt:  time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
					CreatedAt:  time.Date(2026, 5, 3, 11, 0, 0, 0, time.UTC),
				}},
			},
		},
		CollabMissionLinks: stubMissionControlCollabMissionLinkReader{
			links: map[string][]apppkg.CollaborationMissionExecutionLink{
				missionID.String(): {
					{ExecutionKind: "task_os_execution", ExecutionRef: "exec-1"},
					{ExecutionKind: "task_os_execution", ExecutionRef: "exec-2"},
				},
			},
		},
		CollabAgentRuns: stubMissionControlCollabAgentRunReader{
			runs: []apppkg.CollaborationAgentRunView{
				{ID: "exec-2", RequestedAgent: "older-owner", UpdatedAt: time.Date(2026, 5, 3, 12, 1, 0, 0, time.UTC)},
				{ID: "exec-1", RequestedAgent: "newer-owner", UpdatedAt: time.Date(2026, 5, 3, 12, 5, 0, 0, time.UTC)},
			},
		},
	})

	snapshot := projector.Project(nil)
	require.Len(t, snapshot.Missions, 1)
	assert.Equal(t, "newer-owner", snapshot.Missions[0].Collaboration.ActiveOwner)
}

func TestMissionControlCollaborationNoExternalTeamOverstatement(t *testing.T) {
	t.Parallel()

	missionID := uuid.MustParse("34343434-3434-3434-3434-343434343434")
	projector := NewMissionControlProjector(Deps{
		SessionKey: "sess-1",
		MissionReader: stubMissionControlMissionReader{
			missions: map[string][]*mission.Mission{
				"sess-1": {{
					ID:         missionID,
					SessionKey: "sess-1",
					Title:      "Solo mission",
					Status:     mission.StatusActive,
					SourceKind: "user",
					UpdatedAt:  time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
					CreatedAt:  time.Date(2026, 5, 3, 11, 0, 0, 0, time.UTC),
				}},
			},
		},
		CollabMissionLinks: stubMissionControlCollabMissionLinkReader{
			links: map[string][]apppkg.CollaborationMissionExecutionLink{
				missionID.String(): {{ExecutionKind: "task_os_execution", ExecutionRef: "exec-1"}},
			},
		},
		CollabDelegations: stubMissionControlCollabDelegationReader{
			items: []apppkg.CollaborationDelegationRecord{
				{SessionKey: "sess-1", ExecutionRef: "foreign", From: "remote-peer", To: "external-team", Timestamp: time.Date(2026, 5, 3, 12, 5, 0, 0, time.UTC)},
			},
		},
	})

	snapshot := projector.Project(nil)
	require.Len(t, snapshot.Missions, 1)
	assert.Empty(t, snapshot.Missions[0].Collaboration.ParticipantSummary)
	assert.Empty(t, snapshot.Missions[0].Collaboration.HandoffSummary)
}

func TestMissionControlLoopSourceFailureDegradesTruthfully(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{
		SessionKey:        "sess-1",
		MissionReader:     stubMissionControlMissionReader{missions: map[string][]*mission.Mission{"sess-1": {}}},
		LoopInquiryReader: stubMissionControlLoopInquiryReader{err: errors.New("inquiry unavailable")},
		LoopCronReader:    stubMissionControlLoopCronReader{err: errors.New("cron unavailable")},
	})

	snapshot := projector.Project(nil)

	assert.True(t, snapshot.Degraded)
	assert.Contains(t, snapshot.Header.DegradedNote, "Inquiry loops unavailable")
	assert.Contains(t, snapshot.Header.DegradedNote, "Scheduled loops unavailable")
}

func ptrTime(t time.Time) *time.Time { return &t }
