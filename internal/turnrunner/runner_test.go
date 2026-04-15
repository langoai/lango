package turnrunner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/adk/model"
	adksession "google.golang.org/adk/session"
	"google.golang.org/genai"

	"github.com/langoai/lango/internal/adk"
	langosession "github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/turntrace"
)

type fixtureExecutor struct {
	events            []*adksession.Event
	recoveries        []adk.RecoveryInfo
	report            adk.RunReport
	err               error
	chunks            []string
	sleepBeforeEvents time.Duration
	sleepBeforeReturn time.Duration
}

func (e *fixtureExecutor) RunStreamingDetailed(
	_ context.Context,
	_, _ string,
	onChunk adk.ChunkCallback,
	opts ...adk.RunOption,
) (adk.RunReport, error) {
	hooks := adk.ResolveRunHooks(opts...)
	defer func() {
		if hooks.OnFinish != nil {
			hooks.OnFinish()
		}
	}()

	if e.sleepBeforeEvents > 0 {
		time.Sleep(e.sleepBeforeEvents)
	}
	for _, event := range e.events {
		if hooks.OnEvent != nil {
			hooks.OnEvent(event)
		}
	}
	for _, info := range e.recoveries {
		if hooks.OnRecovery != nil {
			hooks.OnRecovery(info)
		}
	}
	for _, chunk := range e.chunks {
		if onChunk != nil {
			onChunk(chunk)
		}
	}
	if e.sleepBeforeReturn > 0 {
		time.Sleep(e.sleepBeforeReturn)
	}
	return e.report, e.err
}

type memoryTraceStore struct {
	traces map[string]turntrace.Trace
	events map[string][]turntrace.Event
}

func newMemoryTraceStore() *memoryTraceStore {
	return &memoryTraceStore{
		traces: make(map[string]turntrace.Trace),
		events: make(map[string][]turntrace.Event),
	}
}

func (s *memoryTraceStore) CreateTrace(_ context.Context, trace turntrace.Trace) error {
	s.traces[trace.TraceID] = trace
	return nil
}

func (s *memoryTraceStore) AppendEvent(_ context.Context, event turntrace.Event) error {
	s.events[event.TraceID] = append(s.events[event.TraceID], event)
	return nil
}

func (s *memoryTraceStore) FinishTrace(
	_ context.Context,
	traceID string,
	outcome turntrace.Outcome,
	summary string,
	errorCode string,
	causeClass string,
	causeDetail string,
	endedAt time.Time,
) error {
	trace := s.traces[traceID]
	trace.Outcome = outcome
	trace.Summary = summary
	trace.ErrorCode = errorCode
	trace.CauseClass = causeClass
	trace.CauseDetail = causeDetail
	trace.EndedAt = &endedAt
	s.traces[traceID] = trace
	return nil
}

func (s *memoryTraceStore) RecentFailures(_ context.Context, _ int) ([]turntrace.Trace, error) {
	return nil, nil
}

func (s *memoryTraceStore) IsolationLeakCount(_ context.Context, _ []string) (int, error) {
	return 0, nil
}

func (s *memoryTraceStore) EventsForTrace(_ context.Context, traceID string) ([]turntrace.Event, error) {
	return s.events[traceID], nil
}

func (s *memoryTraceStore) TracesForSession(_ context.Context, _ string) ([]turntrace.Trace, error) {
	return nil, nil
}

func (s *memoryTraceStore) PurgeTraces(_ context.Context, _ []string) error {
	return nil
}

func (s *memoryTraceStore) TraceCount(_ context.Context) (int, error) {
	return len(s.traces), nil
}

func (s *memoryTraceStore) OldTraces(_ context.Context, _ time.Time, _ bool, _ int) ([]string, error) {
	return nil, nil
}

func (s *memoryTraceStore) RecentByOutcome(_ context.Context, _ turntrace.Outcome, _ time.Time, _ int) ([]turntrace.Trace, error) {
	return nil, nil
}

type stubSessionStore struct {
	annotated []string
}

func (s *stubSessionStore) Create(*langosession.Session) error               { return nil }
func (s *stubSessionStore) Get(string) (*langosession.Session, error)        { return nil, nil }
func (s *stubSessionStore) Update(*langosession.Session) error               { return nil }
func (s *stubSessionStore) Delete(string) error                              { return nil }
func (s *stubSessionStore) AppendMessage(string, langosession.Message) error { return nil }
func (s *stubSessionStore) Close() error                                     { return nil }
func (s *stubSessionStore) GetSalt(string) ([]byte, error)                   { return nil, nil }
func (s *stubSessionStore) SetSalt(string, []byte) error                                    { return nil }
func (s *stubSessionStore) ListSessions(context.Context) ([]langosession.SessionSummary, error) { return nil, nil }
func (s *stubSessionStore) AnnotateTimeout(key, _ string) error {
	s.annotated = append(s.annotated, key)
	return nil
}

type fixtureFile struct {
	Events []fixtureEvent `json:"events"`
}

type fixtureEvent struct {
	Author     string         `json:"author"`
	Kind       string         `json:"kind"`
	Tool       string         `json:"tool"`
	TransferTo string         `json:"transfer_to"`
	Text       string         `json:"text"`
	Args       map[string]any `json:"args"`
	Response   map[string]any `json:"response"`
}

func loadFixtureEvents(t *testing.T, name string) []*adksession.Event {
	t.Helper()
	path := filepath.Join("testdata", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var fixture fixtureFile
	require.NoError(t, json.Unmarshal(data, &fixture))

	events := make([]*adksession.Event, 0, len(fixture.Events))
	for _, item := range fixture.Events {
		evt := &adksession.Event{
			Timestamp: time.Now(),
			Author:    item.Author,
		}
		if item.TransferTo != "" {
			evt.Actions.TransferToAgent = item.TransferTo
		}
		switch item.Kind {
		case "tool_call":
			evt.LLMResponse = model.LLMResponse{
				Content: &genai.Content{
					Role: "model",
					Parts: []*genai.Part{{
						FunctionCall: &genai.FunctionCall{
							ID:   fmt.Sprintf("call-%s", item.Tool),
							Name: item.Tool,
							Args: item.Args,
						},
					}},
				},
			}
		case "tool_result":
			evt.LLMResponse = model.LLMResponse{
				Content: &genai.Content{
					Role: "function",
					Parts: []*genai.Part{{
						FunctionResponse: &genai.FunctionResponse{
							ID:       fmt.Sprintf("call-%s", item.Tool),
							Name:     item.Tool,
							Response: item.Response,
						},
					}},
				},
			}
		case "text":
			evt.LLMResponse = model.LLMResponse{
				Content: &genai.Content{
					Role:  "model",
					Parts: []*genai.Part{{Text: item.Text}},
				},
			}
		}
		events = append(events, evt)
	}
	return events
}

func TestRunner_LoopDetectedFromFixture(t *testing.T) {
	t.Parallel()

	events := loadFixtureEvents(t, "vault_balance_loop.json")
	traceStore := newMemoryTraceStore()
	executor := &fixtureExecutor{
		events: events,
		report: adk.RunReport{
			Diagnostics: adk.RunDiagnostics{ToolResultCount: 4},
		},
		err: &adk.AgentError{
			Code:            adk.ErrToolChurn,
			Message:         "agent error",
			Cause:           fmt.Errorf("call signature repeated"),
			CauseClass:      adk.CauseRepeatedCallSignature,
			CauseDetail:     "payment_balance {} repeated",
			OperatorSummary: "[E007] repeated_call_signature",
		},
	}
	runner := New(Config{
		HardCeiling: 30 * time.Second,
		TraceStore:  traceStore,
	}, executor, &stubSessionStore{}, nil)

	result, err := runner.Run(context.Background(), Request{
		SessionKey: "telegram:test",
		Input:      "check wallet balance",
		Entrypoint: "channel",
	})
	require.NoError(t, err)
	assert.Equal(t, turntrace.OutcomeLoopDetected, result.Outcome)
	assert.Equal(t, string(adk.ErrToolChurn), result.ErrorCode)
	assert.NotEmpty(t, result.TraceID)
	assert.Contains(t, result.ResponseText, "same tool repeatedly")
	assert.NotEmpty(t, traceStore.events[result.TraceID])
	assert.Equal(t, turntrace.OutcomeLoopDetected, traceStore.traces[result.TraceID].Outcome)
	foundTerminal := false
	for _, event := range traceStore.events[result.TraceID] {
		if event.EventType == "terminal_error" {
			foundTerminal = true
			assert.Contains(t, event.PayloadJSON, "repeated_call_signature")
		}
	}
	assert.True(t, foundTerminal, "expected terminal_error event")
}

func TestRunner_EmptyAfterToolUse(t *testing.T) {
	t.Parallel()

	executor := &fixtureExecutor{
		events: []*adksession.Event{
			{
				Timestamp: time.Now(),
				Author:    "vault",
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Role: "function",
						Parts: []*genai.Part{{
							FunctionResponse: &genai.FunctionResponse{
								ID:       "call-payment_balance",
								Name:     "payment_balance",
								Response: map[string]any{"balance": "0.00"},
							},
						}},
					},
				},
			},
		},
		report: adk.RunReport{
			Diagnostics: adk.RunDiagnostics{ToolResultCount: 1},
		},
	}
	runner := New(Config{HardCeiling: 30 * time.Second}, executor, &stubSessionStore{}, nil)

	result, err := runner.Run(context.Background(), Request{
		SessionKey: "telegram:test",
		Input:      "check wallet balance",
		Entrypoint: "channel",
	})
	require.NoError(t, err)
	assert.Equal(t, turntrace.OutcomeEmptyAfterToolUse, result.Outcome)
	assert.Equal(t, string(adk.ErrEmptyAfterToolUse), result.ErrorCode)
	assert.Contains(t, result.ResponseText, "completed tool actions")
	assert.Equal(t, adk.CauseEmptyAfterToolUse, result.CauseClass)
}

func TestRunner_TerminalErrorRecordedWhenNoEventsObserved(t *testing.T) {
	t.Parallel()

	traceStore := newMemoryTraceStore()
	executor := &fixtureExecutor{
		err: &adk.AgentError{
			Code:            adk.ErrToolError,
			Message:         "agent error",
			Cause:           fmt.Errorf("tool 'payment_balance' not found"),
			CauseClass:      adk.CauseToolNotFound,
			CauseDetail:     "tool 'payment_balance' not found",
			OperatorSummary: "[E003] tool_not_found",
		},
	}
	runner := New(Config{
		HardCeiling: 30 * time.Second,
		TraceStore:  traceStore,
	}, executor, &stubSessionStore{}, nil)

	result, err := runner.Run(context.Background(), Request{
		SessionKey: "telegram:test",
		Input:      "payment_balance",
		Entrypoint: "channel",
	})
	require.NoError(t, err)
	assert.Equal(t, turntrace.OutcomeToolError, result.Outcome)
	events := traceStore.events[result.TraceID]
	require.Len(t, events, 1)
	assert.Equal(t, "terminal_error", events[0].EventType)
	assert.Contains(t, events[0].PayloadJSON, "tool_not_found")
}

func TestRunner_TruncatedPayloadFlag(t *testing.T) {
	t.Parallel()

	traceStore := newMemoryTraceStore()
	recorder := newTraceRecorder(context.Background(), traceStore, "trace-1", 15)
	recorder.append(turntrace.Event{
		TraceID:     "trace-1",
		EventType:   "terminal_error",
		PayloadJSON: strings.Repeat("x", 1500),
	})

	require.Len(t, traceStore.events["trace-1"], 1)
	assert.True(t, traceStore.events["trace-1"][0].PayloadTruncated)
}

type deadlineRecordingTraceStore struct {
	createDeadline time.Time
	appendDeadline time.Time
	finishDeadline time.Time
}

func (s *deadlineRecordingTraceStore) CreateTrace(ctx context.Context, _ turntrace.Trace) error {
	s.createDeadline, _ = ctx.Deadline()
	return nil
}

func (s *deadlineRecordingTraceStore) AppendEvent(ctx context.Context, _ turntrace.Event) error {
	s.appendDeadline, _ = ctx.Deadline()
	return nil
}

func (s *deadlineRecordingTraceStore) FinishTrace(ctx context.Context, _ string, _ turntrace.Outcome, _ string, _ string, _ string, _ string, _ time.Time) error {
	s.finishDeadline, _ = ctx.Deadline()
	return nil
}

func (s *deadlineRecordingTraceStore) RecentFailures(context.Context, int) ([]turntrace.Trace, error) {
	return nil, nil
}

func (s *deadlineRecordingTraceStore) IsolationLeakCount(context.Context, []string) (int, error) {
	return 0, nil
}

func (s *deadlineRecordingTraceStore) EventsForTrace(context.Context, string) ([]turntrace.Event, error) {
	return nil, nil
}

func (s *deadlineRecordingTraceStore) TracesForSession(context.Context, string) ([]turntrace.Trace, error) {
	return nil, nil
}

func (s *deadlineRecordingTraceStore) PurgeTraces(context.Context, []string) error {
	return nil
}

func (s *deadlineRecordingTraceStore) TraceCount(context.Context) (int, error) {
	return 0, nil
}

func (s *deadlineRecordingTraceStore) OldTraces(context.Context, time.Time, bool, int) ([]string, error) {
	return nil, nil
}

func (s *deadlineRecordingTraceStore) RecentByOutcome(context.Context, turntrace.Outcome, time.Time, int) ([]turntrace.Trace, error) {
	return nil, nil
}

func TestRunner_TraceWritesUseFreshDeadlines(t *testing.T) {
	t.Parallel()

	traceStore := &deadlineRecordingTraceStore{}
	executor := &fixtureExecutor{
		events: []*adksession.Event{
			{
				Timestamp: time.Now(),
				Author:    "lango-orchestrator",
				LLMResponse: model.LLMResponse{
					Content: &genai.Content{
						Role:  "model",
						Parts: []*genai.Part{{Text: "hello"}},
					},
				},
			},
		},
		report:            adk.RunReport{Response: "done"},
		sleepBeforeEvents: 1100 * time.Millisecond,
		sleepBeforeReturn: 1100 * time.Millisecond,
	}
	runner := New(Config{
		HardCeiling: 10 * time.Second,
		TraceStore:  traceStore,
	}, executor, &stubSessionStore{}, nil)

	_, err := runner.Run(context.Background(), Request{
		SessionKey: "telegram:test",
		Input:      "hello",
		Entrypoint: "channel",
	})
	require.NoError(t, err)
	require.False(t, traceStore.createDeadline.IsZero())
	require.False(t, traceStore.appendDeadline.IsZero())
	require.False(t, traceStore.finishDeadline.IsZero())
	assert.True(t, traceStore.appendDeadline.After(traceStore.createDeadline.Add(500*time.Millisecond)))
	assert.True(t, traceStore.finishDeadline.After(traceStore.appendDeadline.Add(500*time.Millisecond)))
}

func TestRunner_RecoveryAttemptsRecorded(t *testing.T) {
	t.Parallel()

	traceStore := newMemoryTraceStore()
	executor := &fixtureExecutor{
		recoveries: []adk.RecoveryInfo{{
			Action:    "retry_with_hint",
			AgentName: "vault",
			Error:     "tool failed",
		}},
		report: adk.RunReport{Response: "done"},
	}
	runner := New(Config{
		HardCeiling: 10 * time.Second,
		TraceStore:  traceStore,
	}, executor, &stubSessionStore{}, nil)

	result, err := runner.Run(context.Background(), Request{
		SessionKey: "telegram:test",
		Input:      "hello",
		Entrypoint: "channel",
	})
	require.NoError(t, err)
	events := traceStore.events[result.TraceID]
	found := false
	for _, event := range events {
		if event.EventType != turntrace.EventRecoveryAttempt {
			continue
		}
		found = true
		assert.Equal(t, "vault", event.AgentName)
		assert.Contains(t, event.PayloadJSON, "retry_with_hint")
	}
	assert.True(t, found, "expected recovery_attempt event in trace")
}

// countingExecutor tracks call count and returns different results per call.
type countingExecutor struct {
	calls   int
	results []struct {
		report adk.RunReport
		err    error
		chunks []string
	}
}

func (e *countingExecutor) RunStreamingDetailed(
	_ context.Context,
	_, _ string,
	onChunk adk.ChunkCallback,
	opts ...adk.RunOption,
) (adk.RunReport, error) {
	hooks := adk.ResolveRunHooks(opts...)
	defer func() {
		if hooks.OnFinish != nil {
			hooks.OnFinish()
		}
	}()

	idx := e.calls
	if idx >= len(e.results) {
		idx = len(e.results) - 1
	}
	e.calls++
	r := e.results[idx]
	for _, chunk := range r.chunks {
		if onChunk != nil {
			onChunk(chunk)
		}
	}
	return r.report, r.err
}

func transientAgentError() error {
	return &adk.AgentError{
		Code:       adk.ErrModelError,
		Message:    "provider transient",
		CauseClass: adk.CauseProviderTransient,
	}
}

func rateLimitAgentError() error {
	return &adk.AgentError{
		Code:       adk.ErrModelError,
		Message:    "rate limited",
		CauseClass: adk.CauseProviderRateLimit,
	}
}

func authAgentError() error {
	return &adk.AgentError{
		Code:       adk.ErrModelError,
		Message:    "auth failed",
		CauseClass: adk.CauseProviderAuth,
	}
}

func TestRetryLoop_TransientErrorRetries(t *testing.T) {
	exec := &countingExecutor{
		results: []struct {
			report adk.RunReport
			err    error
			chunks []string
		}{
			{err: transientAgentError()},
			{report: adk.RunReport{Response: "ok"}, chunks: []string{"ok"}},
		},
	}
	runner := New(Config{HardCeiling: 30 * time.Second}, exec, nil, nil)
	result, err := runner.Run(context.Background(), Request{
		SessionKey: "test-retry",
		Input:      "hello",
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", result.ResponseText)
	assert.Equal(t, 2, exec.calls, "should have retried once")
}

func TestRetryLoop_NonRetryableExitsImmediately(t *testing.T) {
	exec := &countingExecutor{
		results: []struct {
			report adk.RunReport
			err    error
			chunks []string
		}{
			{err: authAgentError()},
		},
	}
	runner := New(Config{HardCeiling: 30 * time.Second}, exec, nil, nil)
	result, err := runner.Run(context.Background(), Request{
		SessionKey: "test-no-retry",
		Input:      "hello",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, exec.calls, "should NOT retry auth errors")
	assert.Equal(t, adk.CauseProviderAuth, result.CauseClass)
}

func TestRetryLoop_MaxAttemptsExhausted(t *testing.T) {
	exec := &countingExecutor{
		results: []struct {
			report adk.RunReport
			err    error
			chunks []string
		}{
			{err: rateLimitAgentError()},
			{err: rateLimitAgentError()},
			{err: rateLimitAgentError()},
		},
	}
	runner := New(Config{HardCeiling: 30 * time.Second}, exec, nil, nil)
	result, err := runner.Run(context.Background(), Request{
		SessionKey: "test-max-attempts",
		Input:      "hello",
	})
	require.NoError(t, err)
	assert.Equal(t, 3, exec.calls, "should exhaust all 3 attempts")
	assert.Equal(t, adk.CauseProviderRateLimit, result.CauseClass)
}

// modeCapturingExecutor records the mode name resolved from context when
// RunStreamingDetailed is invoked.
type modeCapturingExecutor struct {
	capturedMode string
}

func (e *modeCapturingExecutor) RunStreamingDetailed(
	ctx context.Context,
	_, _ string,
	_ adk.ChunkCallback,
	opts ...adk.RunOption,
) (adk.RunReport, error) {
	hooks := adk.ResolveRunHooks(opts...)
	defer func() {
		if hooks.OnFinish != nil {
			hooks.OnFinish()
		}
	}()
	e.capturedMode = langosession.ModeNameFromContext(ctx)
	return adk.RunReport{Response: "ok"}, nil
}

// stubModeStore satisfies the langosession.Store subset needed by the runner
// for the mode-propagation test (Get + AnnotateTimeout). Other methods are
// no-ops returning nil to keep the fake minimal.
type stubModeStore struct {
	session *langosession.Session
}

func (s *stubModeStore) Create(*langosession.Session) error { return nil }
func (s *stubModeStore) Get(key string) (*langosession.Session, error) {
	if s.session != nil && s.session.Key == key {
		return s.session, nil
	}
	return nil, fmt.Errorf("not found")
}
func (s *stubModeStore) Update(*langosession.Session) error            { return nil }
func (s *stubModeStore) Delete(string) error                           { return nil }
func (s *stubModeStore) AppendMessage(string, langosession.Message) error { return nil }
func (s *stubModeStore) AnnotateTimeout(string, string) error          { return nil }
func (s *stubModeStore) ListSessions(context.Context) ([]langosession.SessionSummary, error) {
	return nil, nil
}
func (s *stubModeStore) GetSalt(string) ([]byte, error)  { return nil, nil }
func (s *stubModeStore) SetSalt(string, []byte) error    { return nil }
func (s *stubModeStore) Close() error                    { return nil }

func TestRunner_PropagatesSessionModeToExecutor(t *testing.T) {
	exec := &modeCapturingExecutor{}

	sess := &langosession.Session{Key: "mode-prop-test"}
	sess.SetMode("research")
	store := &stubModeStore{session: sess}

	runner := New(Config{HardCeiling: 30 * time.Second}, exec, store, nil)
	_, err := runner.Run(context.Background(), Request{
		SessionKey: "mode-prop-test",
		Input:      "hello",
	})
	require.NoError(t, err)
	assert.Equal(t, "research", exec.capturedMode)
}

func TestStaleDetection_TimerResetOnChunk(t *testing.T) {
	// If chunks keep arriving, no stale detection should fire.
	exec := &fixtureExecutor{
		chunks: []string{"a", "b", "c"},
		report: adk.RunReport{Response: "abc"},
	}
	runner := New(Config{
		HardCeiling:  30 * time.Second,
		StaleTimeout: 100 * time.Millisecond,
	}, exec, nil, nil)
	result, err := runner.Run(context.Background(), Request{
		SessionKey: "test-stale-reset",
		Input:      "hello",
	})
	require.NoError(t, err)
	assert.Equal(t, turntrace.OutcomeSuccess, result.Outcome)
}

func TestStaleDetection_InactiveBeforeFirstChunk(t *testing.T) {
	// No chunks at all — stale timer should not fire (only activates after first chunk).
	exec := &fixtureExecutor{
		report: adk.RunReport{Response: "no-stream"},
	}
	runner := New(Config{
		HardCeiling:  30 * time.Second,
		StaleTimeout: 50 * time.Millisecond,
	}, exec, nil, nil)
	result, err := runner.Run(context.Background(), Request{
		SessionKey: "test-stale-inactive",
		Input:      "hello",
	})
	require.NoError(t, err)
	assert.Equal(t, turntrace.OutcomeSuccess, result.Outcome)
}

func TestRetryLoop_SuccessfulFirstAttempt(t *testing.T) {
	exec := &countingExecutor{
		results: []struct {
			report adk.RunReport
			err    error
			chunks []string
		}{
			{report: adk.RunReport{Response: "success"}, chunks: []string{"success"}},
		},
	}
	runner := New(Config{HardCeiling: 30 * time.Second}, exec, nil, nil)
	result, err := runner.Run(context.Background(), Request{
		SessionKey: "test-success",
		Input:      "hello",
	})
	require.NoError(t, err)
	assert.Equal(t, "success", result.ResponseText)
	assert.Equal(t, 1, exec.calls, "should not retry on success")
	assert.Equal(t, turntrace.OutcomeSuccess, result.Outcome)
}
