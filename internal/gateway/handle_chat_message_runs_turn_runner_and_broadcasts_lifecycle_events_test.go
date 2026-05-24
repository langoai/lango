package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	adksession "google.golang.org/adk/session"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/turnrunner"
	"github.com/langoai/lango/internal/turntrace"
)

func TestHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEvents(t *testing.T) {
	t.Parallel()

	traceStore := newHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore()
	executor := &handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsExecutor{
		run: func(
			ctx context.Context,
			sessionID, input string,
			onChunk adk.ChunkCallback,
			hooks adk.RunHooks,
		) (adk.RunReport, error) {
			assert.Equal(t, "sess-runChatUsesProgramSeamAndCleansUpSession1", sessionID)
			assert.Equal(t, "hello gateway", input)
			assert.NotEmpty(t, ctx)

			onChunk("")
			onChunk("streamed chunk")
			for _, target := range []string{"agent-a", "agent-b", "agent-c", "agent-d"} {
				hooks.OnEvent(&adksession.Event{
					Author: "lango-orchestrator",
					Actions: adksession.EventActions{
						TransferToAgent: target,
					},
				})
			}
			return adk.RunReport{TraceID: "trace-runChatUsesProgramSeamAndCleansUpSession1", Response: "final answer"}, nil
		},
	}
	server := New(Config{}, &adk.Agent{}, nil, nil, nil)
	server.SetTurnRunner(turnrunner.New(
		turnrunner.Config{
			HardCeiling:         time.Second,
			TraceStore:          traceStore,
			DelegationBudgetMax: 5,
		},
		executor,
		nil,
		nil,
	))

	sendCh := make(chan []byte, 16)
	otherCh := make(chan []byte, 16)
	server.clientsMu.Lock()
	server.clients["ui-1"] = &Client{ID: "ui-1", Type: "ui", SessionKey: "sess-runChatUsesProgramSeamAndCleansUpSession1", Send: sendCh}
	server.clients["ui-2"] = &Client{ID: "ui-2", Type: "ui", SessionKey: "sess-other", Send: otherCh}
	server.clientsMu.Unlock()

	client := &Client{ID: "ui-1", Type: "ui", Server: server, SessionKey: "sess-runChatUsesProgramSeamAndCleansUpSession1"}
	result, err := server.handleChatMessage(
		client,
		json.RawMessage(`{"message":"hello gateway","sessionKey":"ignored"}`),
	)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"response": "final answer"}, result)

	events := []handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent{
		readHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent(t, sendCh),
		readHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent(t, sendCh),
		readHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent(t, sendCh),
		readHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent(t, sendCh),
		readHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent(t, sendCh),
		readHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent(t, sendCh),
		readHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent(t, sendCh),
	}
	assert.Equal(t, []string{
		"agent.thinking",
		"agent.chunk",
		"agent.delegation",
		"agent.delegation",
		"agent.delegation",
		"agent.delegation",
		"agent.budget_warning",
	}, handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsEventNames(events[:7]))

	done := readHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent(t, sendCh)
	assert.Equal(t, "agent.done", done.Event)
	assert.Equal(t, "trace-runChatUsesProgramSeamAndCleansUpSession1", done.Payload["traceId"])

	chunk := events[1]
	assert.Equal(t, "streamed chunk", chunk.Payload["chunk"])
	assert.Equal(t, "sess-runChatUsesProgramSeamAndCleansUpSession1", chunk.Payload["sessionKey"])

	delegation := events[2]
	assert.Equal(t, "lango-orchestrator", delegation.Payload["from"])
	assert.Equal(t, "agent-a", delegation.Payload["to"])

	budget := events[6]
	assert.Equal(t, "4", budget.Payload["used"])
	assert.Equal(t, "5", budget.Payload["max"])

	select {
	case msg := <-otherCh:
		t.Fatalf("other session should not receive scoped chat event: %s", msg)
	default:
	}
}

func TestHandleChatMessageBroadcastsStructuredFailure(t *testing.T) {
	t.Parallel()

	executor := &handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsExecutor{
		run: func(
			context.Context,
			string,
			string,
			adk.ChunkCallback,
			adk.RunHooks,
		) (adk.RunReport, error) {
			return adk.RunReport{TraceID: "trace-fail"}, &adk.AgentError{
				Code:            adk.ErrModelError,
				Message:         "provider auth failed",
				CauseClass:      adk.CauseProviderAuth,
				CauseDetail:     "bad key",
				OperatorSummary: "provider authentication failed",
			}
		},
	}
	server := New(Config{}, &adk.Agent{}, nil, nil, nil)
	server.SetTurnRunner(turnrunner.New(turnrunner.Config{HardCeiling: time.Second}, executor, nil, nil))

	sendCh := make(chan []byte, 4)
	server.clientsMu.Lock()
	server.clients["ui-1"] = &Client{ID: "ui-1", Type: "ui", SessionKey: "sess-fail", Send: sendCh}
	server.clientsMu.Unlock()

	client := &Client{ID: "ui-1", Type: "ui", Server: server, SessionKey: "sess-fail"}
	result, err := server.handleChatMessage(client, json.RawMessage(`{"message":"hello"}`))
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "[E002] Authentication failed")

	thinking := readHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent(t, sendCh)
	assert.Equal(t, "agent.thinking", thinking.Event)

	failure := readHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent(t, sendCh)
	assert.Equal(t, "agent.error", failure.Event)
	assert.Equal(t, "model_error", failure.Payload["type"])
	assert.Equal(t, "E002", failure.Payload["code"])
	assert.Equal(t, "provider_auth", failure.Payload["causeClass"])
	assert.Equal(t, "provider authentication failed", failure.Payload["summary"])
	assert.Equal(t, "trace-fail", failure.Payload["traceId"])
}

func TestHandleChatMessageRejectsInvalidPayloadsAndMissingRunner(t *testing.T) {
	t.Parallel()

	server := New(Config{}, &adk.Agent{}, nil, nil, nil)
	client := &Client{ID: "ui-1", Type: "ui", Server: server}

	_, err := server.handleChatMessage(client, json.RawMessage(`{`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid params")

	_, err = server.handleChatMessage(client, json.RawMessage(`{"message":""}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "message is required")

	sendCh := make(chan []byte, 2)
	server.clientsMu.Lock()
	server.clients["ui-1"] = &Client{ID: "ui-1", Type: "ui", Send: sendCh}
	server.clientsMu.Unlock()

	_, err = server.handleChatMessage(client, json.RawMessage(`{"message":"hello"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "turn runner is not initialized")

	event := readHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent(t, sendCh)
	assert.Equal(t, "agent.thinking", event.Event)
}

func TestWebSocketReadPumpReportsParseAndMissingMethodErrors(t *testing.T) {
	t.Parallel()

	server := New(Config{HTTPEnabled: true, WebSocketEnabled: true}, nil, nil, nil, nil)
	ts := httptest.NewServer(server.router)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(`{`)))
	parseErr := readHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsRPCResponse(t, conn)
	assert.Equal(t, "", parseErr.ID)
	require.NotNil(t, parseErr.Error)
	assert.Equal(t, -32700, parseErr.Error.Code)
	assert.Equal(t, "parse error", parseErr.Error.Message)

	require.NoError(t, conn.WriteJSON(RPCRequest{ID: "missing-1", Method: "missing.method"}))
	missingErr := readHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsRPCResponse(t, conn)
	assert.Equal(t, "missing-1", missingErr.ID)
	require.NotNil(t, missingErr.Error)
	assert.Equal(t, -32601, missingErr.Error.Code)
	assert.Equal(t, "method not found: missing.method", missingErr.Error.Message)
}

func TestServeInitializesHTTPServerAndReturnsListenerError(t *testing.T) {
	t.Parallel()

	server := New(Config{}, nil, nil, nil, nil)
	err := server.serve(handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsFailingListener{err: errors.New("accept failed")})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accept failed")
	assert.NotNil(t, server.httpServer)
}

func TestStartBackgroundServesUntilShutdown(t *testing.T) {
	t.Parallel()

	server := New(Config{Host: "127.0.0.1", Port: handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsFreeGatewayPort(t), HTTPEnabled: true}, nil, nil, nil, nil)
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	require.NoError(t, server.StartBackground(&wg, func(err error) {
		errCh <- err
	}))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, server.Shutdown(ctx))

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("StartBackground serve goroutine did not stop")
	}

	select {
	case err := <-errCh:
		t.Fatalf("graceful shutdown should not report background error: %v", err)
	default:
	}
}

func handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsFreeGatewayPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)
	return addr.Port
}

type handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsExecutor struct {
	run func(
		context.Context,
		string,
		string,
		adk.ChunkCallback,
		adk.RunHooks,
	) (adk.RunReport, error)
}

func (e *handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsExecutor) RunStreamingDetailed(
	ctx context.Context,
	sessionID, input string,
	onChunk adk.ChunkCallback,
	opts ...adk.RunOption,
) (adk.RunReport, error) {
	hooks := adk.ResolveRunHooks(opts...)
	defer func() {
		if hooks.OnFinish != nil {
			hooks.OnFinish()
		}
	}()
	if e.run == nil {
		return adk.RunReport{Response: "ok"}, nil
	}
	return e.run(ctx, sessionID, input, onChunk, hooks)
}

type handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent struct {
	Type    string         `json:"type"`
	Event   string         `json:"event"`
	Payload map[string]any `json:"payload"`
}

func readHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent(t *testing.T, ch <-chan []byte) handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent {
	t.Helper()

	select {
	case msg := <-ch:
		var event handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent
		require.NoError(t, json.Unmarshal(msg, &event))
		require.Equal(t, "event", event.Type)
		return event
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for gateway event")
		return handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent{}
	}
}

func handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsEventNames(events []handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsGatewayEvent) []string {
	names := make([]string, 0, len(events))
	for _, event := range events {
		names = append(names, event.Event)
	}
	return names
}

func readHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsRPCResponse(t *testing.T, conn *websocket.Conn) RPCResponse {
	t.Helper()

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(time.Second)))
	var resp RPCResponse
	require.NoError(t, conn.ReadJSON(&resp))
	return resp
}

type handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsFailingListener struct {
	err error
}

func (l handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsFailingListener) Accept() (net.Conn, error) {
	return nil, l.err
}

func (handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsFailingListener) Close() error {
	return nil
}

func (handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsFailingListener) Addr() net.Addr {
	return handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsAddr("runChatUsesProgramSeamAndCleansUpSession1-listener")
}

type handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsAddr string

func (a handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsAddr) Network() string {
	return string(a)
}

func (a handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsAddr) String() string {
	return string(a)
}

type handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore struct {
	mu     sync.Mutex
	traces map[string]turntrace.Trace
	events map[string][]turntrace.Event
}

func newHandleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore() *handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore {
	return &handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore{
		traces: make(map[string]turntrace.Trace),
		events: make(map[string][]turntrace.Event),
	}
}

func (s *handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore) CreateTrace(_ context.Context, trace turntrace.Trace) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.traces[trace.TraceID] = trace
	return nil
}

func (s *handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore) AppendEvent(_ context.Context, event turntrace.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events[event.TraceID] = append(s.events[event.TraceID], event)
	return nil
}

func (s *handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore) FinishTrace(
	_ context.Context,
	traceID string,
	outcome turntrace.Outcome,
	summary, errorCode, causeClass, causeDetail string,
	endedAt time.Time,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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

func (*handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore) RecentFailures(context.Context, int) ([]turntrace.Trace, error) {
	return nil, nil
}

func (*handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore) IsolationLeakCount(context.Context, []string) (int, error) {
	return 0, nil
}

func (s *handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore) EventsForTrace(_ context.Context, traceID string) ([]turntrace.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]turntrace.Event(nil), s.events[traceID]...), nil
}

func (*handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore) TracesForSession(context.Context, string) ([]turntrace.Trace, error) {
	return nil, nil
}

func (*handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore) PurgeTraces(context.Context, []string) error {
	return nil
}

func (*handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore) TraceCount(context.Context) (int, error) {
	return 0, nil
}

func (*handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore) OldTraces(context.Context, time.Time, bool, int) ([]string, error) {
	return nil, nil
}

func (*handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore) RecentByOutcome(
	context.Context,
	turntrace.Outcome,
	time.Time,
	int,
) ([]turntrace.Trace, error) {
	return nil, nil
}

var _ turntrace.Store = (*handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsTraceStore)(nil)
var _ net.Listener = handleChatMessageRunsTurnRunnerAndBroadcastsLifecycleEventsFailingListener{}
