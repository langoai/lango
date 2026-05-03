package app

import (
	"context"
	"testing"
	"time"

	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/testutil"
	"github.com/langoai/lango/internal/turntrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTurnTraceStore struct {
	tracesBySession map[string][]turntrace.Trace
	eventsByTrace   map[string][]turntrace.Event
}

func (f *fakeTurnTraceStore) CreateTrace(context.Context, turntrace.Trace) error { return nil }
func (f *fakeTurnTraceStore) AppendEvent(context.Context, turntrace.Event) error { return nil }
func (f *fakeTurnTraceStore) FinishTrace(context.Context, string, turntrace.Outcome, string, string, string, string, time.Time) error {
	return nil
}
func (f *fakeTurnTraceStore) RecentFailures(context.Context, int) ([]turntrace.Trace, error) {
	return nil, nil
}
func (f *fakeTurnTraceStore) IsolationLeakCount(context.Context, []string) (int, error) {
	return 0, nil
}
func (f *fakeTurnTraceStore) EventsForTrace(_ context.Context, traceID string) ([]turntrace.Event, error) {
	return append([]turntrace.Event(nil), f.eventsByTrace[traceID]...), nil
}
func (f *fakeTurnTraceStore) TracesForSession(_ context.Context, sessionKey string) ([]turntrace.Trace, error) {
	return append([]turntrace.Trace(nil), f.tracesBySession[sessionKey]...), nil
}
func (f *fakeTurnTraceStore) PurgeTraces(context.Context, []string) error { return nil }
func (f *fakeTurnTraceStore) TraceCount(context.Context) (int, error)     { return 0, nil }
func (f *fakeTurnTraceStore) OldTraces(context.Context, time.Time, bool, int) ([]string, error) {
	return nil, nil
}
func (f *fakeTurnTraceStore) RecentByOutcome(context.Context, turntrace.Outcome, time.Time, int) ([]turntrace.Trace, error) {
	return nil, nil
}

func TestCollaborationRuntimeBridge_StoresOnlyMissionAttributedRecords(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	bridge := newCollaborationRuntimeBridge(bus)
	bridge.BindSessionMission("sess-1", "mission-1")

	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	origNow := eventTime
	eventTime = func() time.Time { return now }
	t.Cleanup(func() { eventTime = origNow })

	bus.Publish(agentrt.BudgetAlertEvent{SessionID: "sess-1", Used: 12, Limit: 15})
	bus.Publish(agentrt.BudgetAlertEvent{SessionID: "sess-2", Used: 20, Limit: 20})
	bus.Publish(agentrt.BudgetAlertEvent{Used: 1, Limit: 2})

	bus.Publish(agentrt.RecoveryDecisionEvent{SessionKey: "sess-1", Action: "retry_with_hint", CauseClass: "timeout"})
	bus.Publish(agentrt.RecoveryDecisionEvent{SessionKey: "sess-2", Action: "fallback", CauseClass: "rate_limit"})
	bus.Publish(agentrt.RecoveryDecisionEvent{Action: "retry", CauseClass: "timeout"})

	budget := bridge.ListBudgetSignals("mission-1")
	recovery := bridge.ListRecoverySignals("mission-1")

	require.Len(t, budget, 1)
	assert.Equal(t, 12, budget[0].Used)
	require.Len(t, recovery, 1)
	assert.Equal(t, "retry_with_hint", recovery[0].Action)
	assert.Empty(t, bridge.ListBudgetSignals("mission-2"))
	assert.Empty(t, bridge.ListRecoverySignals("mission-2"))
}

func TestCollaborationDelegationReader_ParsesOnlyDelegationEvents(t *testing.T) {
	t.Parallel()

	store := &fakeTurnTraceStore{
		tracesBySession: map[string][]turntrace.Trace{
			"sess-1": {
				{TraceID: "trace-1", SessionKey: "sess-1"},
			},
		},
		eventsByTrace: map[string][]turntrace.Event{
			"trace-1": {
				{TraceID: "trace-1", EventType: turntrace.EventDelegation, AgentName: "planner", PayloadJSON: `{"to":"researcher"}`, CreatedAt: time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)},
				{TraceID: "trace-1", EventType: turntrace.EventText, AgentName: "planner", PayloadJSON: `{"text":"ignore"}`, CreatedAt: time.Date(2026, 5, 3, 12, 0, 1, 0, time.UTC)},
				{TraceID: "trace-1", EventType: turntrace.EventDelegationReturn, AgentName: "researcher", PayloadJSON: `{"to":"planner"}`, CreatedAt: time.Date(2026, 5, 3, 12, 0, 2, 0, time.UTC)},
			},
		},
	}

	reader := &collaborationDelegationReader{store: store}
	records, err := reader.ListDelegationsForSession(context.Background(), "sess-1")
	require.NoError(t, err)
	require.Len(t, records, 2)
	assert.Equal(t, "planner", records[0].From)
	assert.Equal(t, "researcher", records[0].To)
	assert.Equal(t, "researcher", records[1].From)
	assert.Equal(t, "planner", records[1].To)
}

func TestCollaborationMissionAndAgentRunReadersAreNarrow(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	missionStore := mission.NewEntStore(client)
	missionRow, err := missionStore.CreateMission(context.Background(), mission.CreateMissionInput{
		SessionKey: "sess-1",
		Title:      "Mission",
		SourceKind: "user",
	})
	require.NoError(t, err)
	err = missionStore.AppendExecutionLink(context.Background(), mission.AppendExecutionLinkInput{
		MissionID:     missionRow.ID.String(),
		ExecutionKind: mission.ExecutionKindTaskOSExecution,
		ExecutionRef:  "exec-1",
		LinkRole:      mission.LinkRolePrimary,
	})
	require.NoError(t, err)

	runStore := agentrt.NewInMemoryAgentRunStore()
	require.NoError(t, runStore.Create(&agentrt.AgentRun{
		ID:               "exec-1",
		RequestedAgent:   "researcher",
		RuntimeCondition: agentrt.AgentRunConditionWaitingOnTeammate,
		BlockedReason:    "waiting",
	}))

	linkReader := &collaborationMissionLinkReader{store: missionStore}
	runReader := &collaborationAgentRunReader{store: runStore}

	links, err := linkReader.ListMissionExecutionLinks(context.Background(), missionRow.ID.String())
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, "exec-1", links[0].ExecutionRef)

	runs := runReader.ListAgentRuns()
	require.Len(t, runs, 1)
	assert.Equal(t, "researcher", runs[0].RequestedAgent)
	assert.Equal(t, string(agentrt.AgentRunConditionWaitingOnTeammate), runs[0].RuntimeCondition)
}
