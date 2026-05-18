package turntrace

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/langoai/lango/internal/ent/enttest"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestEntStoreNilReceiverReturnsZeroValues(t *testing.T) {
	ctx := context.Background()
	var store *EntStore

	require.NoError(t, store.CreateTrace(ctx, Trace{TraceID: "trace"}))
	require.NoError(t, store.AppendEvent(ctx, Event{TraceID: "trace", Seq: 1}))
	require.NoError(t, store.FinishTrace(ctx, "trace", OutcomeSuccess, "", "", "", "", time.Now()))

	failures, err := store.RecentFailures(ctx, 1)
	require.NoError(t, err)
	require.Nil(t, failures)

	leaks, err := store.IsolationLeakCount(ctx, []string{"agent"})
	require.NoError(t, err)
	require.Zero(t, leaks)

	events, err := store.EventsForTrace(ctx, "trace")
	require.NoError(t, err)
	require.Nil(t, events)

	traces, err := store.TracesForSession(ctx, "session")
	require.NoError(t, err)
	require.Nil(t, traces)

	require.NoError(t, store.PurgeTraces(ctx, []string{"trace"}))

	count, err := store.TraceCount(ctx)
	require.NoError(t, err)
	require.Zero(t, count)

	old, err := store.OldTraces(ctx, time.Now(), true, 1)
	require.NoError(t, err)
	require.Nil(t, old)

	recent, err := store.RecentByOutcome(ctx, OutcomeSuccess, time.Now().Add(-time.Hour), 1)
	require.NoError(t, err)
	require.Nil(t, recent)
}

func TestEntStoreTraceLifecycleQueriesAndPurge(t *testing.T) {
	ctx := context.Background()
	store := newTurnTraceTestStore(t)

	start := time.Unix(100, 0).UTC()
	trace1End := start.Add(5 * time.Second)
	trace2Start := start.Add(time.Minute)
	trace3Start := start.Add(2 * time.Minute)

	require.NoError(t, store.CreateTrace(ctx, Trace{
		TraceID:     "trace-success",
		SessionKey:  "session-1",
		Entrypoint:  "cli",
		Outcome:     OutcomeRunning,
		CauseDetail: "   ",
		StartedAt:   start,
	}))
	require.NoError(t, store.AppendEvent(ctx, Event{
		TraceID:          "trace-success",
		Seq:              2,
		EventType:        EventToolResult,
		ToolName:         "web_search",
		PayloadJSON:      `{"ok":true}`,
		PayloadTruncated: true,
		CreatedAt:        start.Add(2 * time.Second),
	}))
	require.NoError(t, store.AppendEvent(ctx, Event{
		TraceID:       "trace-success",
		Seq:           1,
		EventType:     EventToolCall,
		AgentName:     "agent-a",
		ToolName:      "web_search",
		CallSignature: "web_search(query)",
		CreatedAt:     start.Add(time.Second),
	}))
	require.NoError(t, store.FinishTrace(
		ctx,
		"trace-success",
		OutcomeSuccess,
		"completed",
		"",
		"",
		"",
		trace1End,
	))

	require.NoError(t, store.CreateTrace(ctx, Trace{
		TraceID:    "trace-timeout",
		SessionKey: "session-1",
		Entrypoint: "gateway",
		Outcome:    OutcomeTimeout,
		ErrorCode:  "deadline_exceeded",
		CauseClass: "timeout",
		Summary:    "timed out",
		StartedAt:  trace2Start,
	}))
	require.NoError(t, store.CreateTrace(ctx, Trace{
		TraceID:    "trace-running",
		SessionKey: "session-2",
		Entrypoint: "cli",
		Outcome:    OutcomeRunning,
		StartedAt:  trace3Start,
	}))

	events, err := store.EventsForTrace(ctx, "trace-success")
	require.NoError(t, err)
	require.Len(t, events, 2)
	require.Equal(t, int64(1), events[0].Seq)
	require.Equal(t, EventToolCall, events[0].EventType)
	require.Equal(t, "agent-a", events[0].AgentName)
	require.Equal(t, "web_search(query)", events[0].CallSignature)
	require.Equal(t, int64(2), events[1].Seq)
	require.True(t, events[1].PayloadTruncated)

	sessionTraces, err := store.TracesForSession(ctx, "session-1")
	require.NoError(t, err)
	require.Len(t, sessionTraces, 2)
	require.Equal(t, "trace-success", sessionTraces[0].TraceID)
	require.Equal(t, "trace-timeout", sessionTraces[1].TraceID)
	require.Equal(t, OutcomeSuccess, sessionTraces[0].Outcome)
	require.NotNil(t, sessionTraces[0].EndedAt)
	require.True(t, sessionTraces[0].EndedAt.Equal(trace1End))
	require.Equal(t, "completed", sessionTraces[0].Summary)

	failures, err := store.RecentFailures(ctx, 0)
	require.NoError(t, err)
	require.Len(t, failures, 1)
	require.Equal(t, "trace-timeout", failures[0].TraceID)
	require.Equal(t, OutcomeTimeout, failures[0].Outcome)

	recentSuccess, err := store.RecentByOutcome(ctx, OutcomeSuccess, start.Add(-time.Second), 0)
	require.NoError(t, err)
	require.Len(t, recentSuccess, 1)
	require.Equal(t, "trace-success", recentSuccess[0].TraceID)

	oldSuccess, err := store.OldTraces(ctx, trace2Start, true, 0)
	require.NoError(t, err)
	require.Equal(t, []string{"trace-success"}, oldSuccess)

	allOld, err := store.OldTraces(ctx, trace3Start, false, 1)
	require.NoError(t, err)
	require.Equal(t, []string{"trace-success"}, allOld)

	count, err := store.TraceCount(ctx)
	require.NoError(t, err)
	require.Equal(t, 3, count)

	require.NoError(t, store.PurgeTraces(ctx, []string{"trace-success"}))
	count, err = store.TraceCount(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, count)

	events, err = store.EventsForTrace(ctx, "trace-success")
	require.NoError(t, err)
	require.Empty(t, events)
}

func TestEntStoreIsolationLeakCountCountsMatchingAuthors(t *testing.T) {
	ctx := context.Background()
	store := newTurnTraceTestStore(t)

	require.NoError(t, store.client.Message.Create().
		SetRole("assistant").
		SetContent("hidden").
		SetAuthor("isolated-agent").
		SetTimestamp(time.Unix(10, 0)).
		Exec(ctx))
	require.NoError(t, store.client.Message.Create().
		SetRole("assistant").
		SetContent("visible").
		SetAuthor("orchestrator").
		SetTimestamp(time.Unix(20, 0)).
		Exec(ctx))

	count, err := store.IsolationLeakCount(ctx, []string{"isolated-agent", "another-agent"})
	require.NoError(t, err)
	require.Equal(t, 1, count)

	count, err = store.IsolationLeakCount(ctx, nil)
	require.NoError(t, err)
	require.Zero(t, count)
}

func TestEntStoreClosedBackendReturnsErrors(t *testing.T) {
	ctx := context.Background()
	store := newTurnTraceTestStore(t)
	require.NoError(t, store.client.Close())

	err := store.CreateTrace(ctx, Trace{TraceID: "closed", SessionKey: "session", Outcome: OutcomeRunning})
	require.Error(t, err)
	require.Contains(t, err.Error(), `create trace "closed"`)

	err = store.AppendEvent(ctx, Event{TraceID: "closed", Seq: 1, EventType: EventText})
	require.Error(t, err)
	require.Contains(t, err.Error(), `append trace event "closed"/1`)

	err = store.FinishTrace(ctx, "closed", OutcomeSuccess, "done", "", "", "", time.Now())
	require.Error(t, err)
	require.Contains(t, err.Error(), `finish trace "closed"`)

	_, err = store.RecentFailures(ctx, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "query recent failures")

	_, err = store.EventsForTrace(ctx, "closed")
	require.Error(t, err)
	require.Contains(t, err.Error(), `query events for trace "closed"`)

	_, err = store.TracesForSession(ctx, "session")
	require.Error(t, err)
	require.Contains(t, err.Error(), `query traces for session "session"`)

	err = store.PurgeTraces(ctx, []string{"closed"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "purge trace events")

	_, err = store.TraceCount(ctx)
	require.Error(t, err)

	_, err = store.OldTraces(ctx, time.Now(), false, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "query old traces")

	_, err = store.RecentByOutcome(ctx, OutcomeSuccess, time.Now().Add(-time.Hour), 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), `query recent by outcome "success"`)

	_, err = store.IsolationLeakCount(ctx, []string{"agent"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "count isolation leaks")
}

func newTurnTraceTestStore(t *testing.T) *EntStore {
	t.Helper()

	dsnName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	client := enttest.Open(t, "sqlite3", "file:"+dsnName+"?mode=memory&_fk=1")
	t.Cleanup(func() {
		_ = client.Close()
	})
	return NewEntStore(client)
}
