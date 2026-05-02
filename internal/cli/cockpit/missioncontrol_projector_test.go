package cockpit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/runledger"
)

type stubMissionControlRunLedgerReader struct {
	snapshots map[string]*runledger.RunSnapshot
}

func (s stubMissionControlRunLedgerReader) ListRuns(context.Context, int) ([]runledger.RunSummary, error) {
	return nil, nil
}

func (s stubMissionControlRunLedgerReader) GetRunSnapshot(_ context.Context, runID string) (*runledger.RunSnapshot, error) {
	if snap, ok := s.snapshots[runID]; ok {
		return snap.DeepCopy(), nil
	}
	return nil, nil
}

func (s stubMissionControlRunLedgerReader) ListRunSummariesBySession(context.Context, string, int) ([]runledger.RunSummary, error) {
	return nil, nil
}

type stubMissionControlAgentRunReader struct {
	runs map[string]*agentrt.AgentRun
	list []*agentrt.AgentRun
}

func (s stubMissionControlAgentRunReader) Get(id string) (*agentrt.AgentRun, error) {
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
	assert.Equal(t, "learn:older", snapshot.Missions[1].ID)
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

func TestMissionControlDegradedNilReaders(t *testing.T) {
	t.Parallel()

	projector := NewMissionControlProjector(Deps{})
	snapshot := projector.Project([]background.TaskSnapshot{{ID: "task-1", StatusText: "pending", Prompt: "Queue work"}})

	assert.True(t, snapshot.Degraded)
	assert.Contains(t, snapshot.Header.DegradedNote, "RunLedger unavailable")
	assert.Contains(t, snapshot.Header.DegradedNote, "Agent runtime unavailable")
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
	})

	snapshot := projector.Project(nil)

	assert.Equal(t, "worker-c active", snapshot.Header.ActiveAgentSummary)
	assert.Equal(t, "openai / gpt-5", snapshot.Header.ModelProviderSummary)
	assert.Equal(t, 1, snapshot.Header.PendingDecisionCount)
	assert.Empty(t, snapshot.Header.ContextSummary)
	assert.Equal(t, "150 tokens across 1 requests", snapshot.Header.MetricsSummary)
	assert.Empty(t, snapshot.Header.DegradedNote)
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
