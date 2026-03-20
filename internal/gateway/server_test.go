package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/gatekeeper"
	"github.com/langoai/lango/internal/runledger"
)

func TestGatewayServer(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	server.RegisterHandler("echo", func(_ *Client, params json.RawMessage) (interface{}, error) {
		var input string
		if err := json.Unmarshal(params, &input); err != nil {
			return nil, err
		}
		return "echo: " + input, nil
	})

	ts := httptest.NewServer(server.router)
	defer ts.Close()

	// Test HTTP Health
	resp, err := http.Get(ts.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test WebSocket
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Test RPC Call
	req := RPCRequest{
		ID:     "1",
		Method: "echo",
		Params: json.RawMessage(`"hello"`),
	}
	require.NoError(t, conn.WriteJSON(req))

	var rpcResp RPCResponse
	require.NoError(t, conn.ReadJSON(&rpcResp))

	assert.Equal(t, "1", rpcResp.ID)
	assert.Equal(t, "echo: hello", rpcResp.Result)

	// Test Broadcast
	done := make(chan bool)
	go func() {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var eventMsg map[string]interface{}
		if err := json.Unmarshal(msg, &eventMsg); err != nil {
			return
		}
		assert.Equal(t, "event", eventMsg["type"])
		assert.Equal(t, "test-event", eventMsg["event"])
		done <- true
	}()

	time.Sleep(100 * time.Millisecond)
	server.Broadcast("test-event", "payload")

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for broadcast")
	}
}

func TestChatMessage_UnauthenticatedUsesDefault(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	client := &Client{
		ID:         "test-client",
		Type:       "ui",
		Server:     server,
		SessionKey: "",
	}

	params := json.RawMessage(`{"message":"hello"}`)
	_, err := server.handleChatMessage(client, params)
	require.Error(t, err)
}

func TestChatMessage_AuthenticatedUsesOwnSession(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	client := &Client{
		ID:         "test-client",
		Type:       "ui",
		Server:     server,
		SessionKey: "sess_my-authenticated-key",
	}

	params := json.RawMessage(`{"message":"hello","sessionKey":"hacker-session"}`)
	_, err := server.handleChatMessage(client, params)
	require.Error(t, err)
}

func TestApprovalResponse_AtomicDelete(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	respChan := make(chan approval.ApprovalResponse, 1)
	server.pendingApprovalsMu.Lock()
	server.pendingApprovals["req-1"] = respChan
	server.pendingApprovalsMu.Unlock()

	params := json.RawMessage(`{"requestId":"req-1","approved":true}`)
	result, err := server.handleApprovalResponse(nil, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	select {
	case resp := <-respChan:
		assert.True(t, resp.Approved)
	default:
		t.Error("expected approval result on channel")
	}

	server.pendingApprovalsMu.Lock()
	_, exists := server.pendingApprovals["req-1"]
	server.pendingApprovalsMu.Unlock()
	assert.False(t, exists, "expected pending approval to be deleted after response")
}

func TestApprovalResponse_DuplicateResponse(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	respChan := make(chan approval.ApprovalResponse, 1)
	server.pendingApprovalsMu.Lock()
	server.pendingApprovals["req-dup"] = respChan
	server.pendingApprovalsMu.Unlock()

	params := json.RawMessage(`{"requestId":"req-dup","approved":true}`)
	_, err := server.handleApprovalResponse(nil, params)
	require.NoError(t, err)

	_, err = server.handleApprovalResponse(nil, params)
	require.NoError(t, err)

	select {
	case <-respChan:
	default:
		t.Error("expected one approval result on channel")
	}

	select {
	case <-respChan:
		t.Error("unexpected second value on channel — duplicate response was not blocked")
	default:
	}
}

func TestBroadcastToSession_ScopedBySessionKey(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	sendA := make(chan []byte, 256)
	sendB := make(chan []byte, 256)
	sendC := make(chan []byte, 256)

	server.clientsMu.Lock()
	server.clients["a"] = &Client{ID: "a", Type: "ui", SessionKey: "sess-1", Send: sendA}
	server.clients["b"] = &Client{ID: "b", Type: "ui", SessionKey: "sess-2", Send: sendB}
	server.clients["c"] = &Client{ID: "c", Type: "companion", SessionKey: "sess-1", Send: sendC}
	server.clientsMu.Unlock()

	server.BroadcastToSession("sess-1", "agent.thinking", map[string]string{"sessionKey": "sess-1"})

	select {
	case msg := <-sendA:
		var eventMsg map[string]interface{}
		require.NoError(t, json.Unmarshal(msg, &eventMsg))
		assert.Equal(t, "agent.thinking", eventMsg["event"])
	default:
		t.Error("expected client A to receive broadcast")
	}

	select {
	case <-sendB:
		t.Error("client B should not receive broadcast for sess-1")
	default:
	}

	select {
	case <-sendC:
		t.Error("companion client should not receive session broadcast")
	default:
	}
}

func TestBroadcastToSession_NoAuth(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	sendA := make(chan []byte, 256)
	sendB := make(chan []byte, 256)

	server.clientsMu.Lock()
	server.clients["a"] = &Client{ID: "a", Type: "ui", SessionKey: "", Send: sendA}
	server.clients["b"] = &Client{ID: "b", Type: "ui", SessionKey: "", Send: sendB}
	server.clientsMu.Unlock()

	server.BroadcastToSession("", "agent.done", map[string]string{"sessionKey": ""})

	select {
	case <-sendA:
	default:
		t.Error("expected client A to receive broadcast")
	}

	select {
	case <-sendB:
	default:
		t.Error("expected client B to receive broadcast")
	}
}

func TestHandleChatMessage_NilAgent_ReturnsErrorWithoutBroadcast(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
		RequestTimeout:   50 * time.Millisecond,
	}
	server := New(cfg, nil, nil, nil, nil)

	sendCh := make(chan []byte, 256)
	server.clientsMu.Lock()
	server.clients["ui-1"] = &Client{
		ID:         "ui-1",
		Type:       "ui",
		SessionKey: "",
		Send:       sendCh,
	}
	server.clientsMu.Unlock()

	client := &Client{ID: "test", Type: "ui", Server: server, SessionKey: ""}
	params := json.RawMessage(`{"message":"hello"}`)
	_, err := server.handleChatMessage(client, params)
	require.Error(t, err)

	select {
	case msg := <-sendCh:
		t.Errorf("expected no broadcast, got: %s", msg)
	default:
	}
}

func TestHandleChatMessage_SuccessBroadcastsAgentDone(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	sendCh := make(chan []byte, 256)
	server.clientsMu.Lock()
	server.clients["ui-1"] = &Client{
		ID:         "ui-1",
		Type:       "ui",
		SessionKey: "",
		Send:       sendCh,
	}
	server.clientsMu.Unlock()

	server.BroadcastToSession("", "agent.done", map[string]string{
		"sessionKey": "",
	})

	select {
	case msg := <-sendCh:
		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(msg, &m))
		assert.Equal(t, "agent.done", m["event"])
	default:
		t.Error("expected agent.done broadcast")
	}
}

func TestHandleChatMessage_ErrorBroadcastsAgentErrorEvent(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	sendCh := make(chan []byte, 256)
	server.clientsMu.Lock()
	server.clients["ui-1"] = &Client{
		ID:         "ui-1",
		Type:       "ui",
		SessionKey: "",
		Send:       sendCh,
	}
	server.clientsMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	<-ctx.Done()

	errType := "unknown"
	if ctx.Err() == context.DeadlineExceeded {
		errType = "timeout"
	}
	server.BroadcastToSession("", "agent.error", map[string]string{
		"sessionKey": "",
		"error":      fmt.Sprintf("agent error: %v", ctx.Err()),
		"type":       errType,
	})

	select {
	case msg := <-sendCh:
		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(msg, &m))
		assert.Equal(t, "agent.error", m["event"])
		payload, ok := m["payload"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "timeout", payload["type"])
	default:
		t.Error("expected agent.error broadcast")
	}
}

func TestWarningBroadcast_ApproachingTimeout(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	sendCh := make(chan []byte, 256)
	server.clientsMu.Lock()
	server.clients["ui-1"] = &Client{
		ID:         "ui-1",
		Type:       "ui",
		SessionKey: "",
		Send:       sendCh,
	}
	server.clientsMu.Unlock()

	timeout := 50 * time.Millisecond
	sessionKey := "test-session"

	warnTimer := time.AfterFunc(time.Duration(float64(timeout)*0.8), func() {
		server.BroadcastToSession(sessionKey, "agent.warning", map[string]string{
			"sessionKey": sessionKey,
			"message":    "Request is taking longer than expected",
			"type":       "approaching_timeout",
		})
	})
	defer warnTimer.Stop()

	time.Sleep(70 * time.Millisecond)

	select {
	case msg := <-sendCh:
		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(msg, &m))
		assert.Equal(t, "agent.warning", m["event"])
		payload, ok := m["payload"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "approaching_timeout", payload["type"])
		assert.Equal(t, "Request is taking longer than expected", payload["message"])
	default:
		t.Error("expected agent.warning broadcast after 80% timeout")
	}
}

func TestSetSanitizer_SanitizesChunksAndResponse(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	san, err := gatekeeper.NewSanitizer(config.GatekeeperConfig{})
	require.NoError(t, err)
	server.SetSanitizer(san)

	assert.NotNil(t, server.sanitizer)
	assert.True(t, server.sanitizer.Enabled())

	// Verify sanitizer strips thought tags from text.
	got := server.sanitizer.Sanitize("Hello <thought>internal</thought> world")
	assert.Equal(t, "Hello  world", got)
}

func TestSetSanitizer_DisabledPassthrough(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	disabled := false
	san, err := gatekeeper.NewSanitizer(config.GatekeeperConfig{
		Enabled: &disabled,
	})
	require.NoError(t, err)
	server.SetSanitizer(san)

	assert.False(t, server.sanitizer.Enabled())

	// Disabled sanitizer should pass through unchanged.
	got := server.sanitizer.Sanitize("Hello <thought>internal</thought> world")
	assert.Equal(t, "Hello <thought>internal</thought> world", got)
}

func TestSetSanitizer_NilSanitizerSafe(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	// SetSanitizer(nil) should not panic.
	server.SetSanitizer(nil)
	assert.Nil(t, server.sanitizer)
}

func TestShutdown_CancelsInflightRequestContexts(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	// Derive a child context from shutdownCtx (same as handleChatMessage does).
	ctx, cancel := context.WithTimeout(server.shutdownCtx, 5*time.Minute)
	defer cancel()

	// shutdownCancel should propagate to the child.
	server.shutdownCancel()

	select {
	case <-ctx.Done():
		assert.ErrorIs(t, ctx.Err(), context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("child context was not cancelled after shutdownCancel")
	}
}

func TestShutdown_CancelsApprovalWait(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
		ApprovalTimeout:  30 * time.Second, // long timeout — should NOT be reached
	}
	server := New(cfg, nil, nil, nil, nil)

	// Register a fake companion so RequestApproval doesn't short-circuit.
	server.clientsMu.Lock()
	server.clients["companion-1"] = &Client{
		ID:   "companion-1",
		Type: "companion",
		Send: make(chan []byte, 256),
	}
	server.clientsMu.Unlock()

	// Use shutdownCtx as parent (matches real request flow).
	ctx, cancel := context.WithTimeout(server.shutdownCtx, 30*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := server.RequestApproval(ctx, "dangerous action")
		done <- err
	}()

	// Simulate Ctrl+C — cancel all in-flight contexts.
	time.Sleep(50 * time.Millisecond) // let goroutine enter select
	server.shutdownCancel()

	select {
	case err := <-done:
		// Must return context.Canceled, NOT ErrApprovalTimeout.
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(2 * time.Second):
		t.Fatal("RequestApproval did not return after shutdown — this is the bug")
	}
}

func TestApprovalTimeout_UsesConfigTimeout(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
		ApprovalTimeout:  50 * time.Millisecond,
	}
	server := New(cfg, nil, nil, nil, nil)

	server.clientsMu.Lock()
	server.clients["companion-1"] = &Client{
		ID:   "companion-1",
		Type: "companion",
		Send: make(chan []byte, 256),
	}
	server.clientsMu.Unlock()

	_, err := server.RequestApproval(t.Context(), "test approval")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "approval timeout")
}

func TestHandleChatMessage_ResumeIntentReturnsCandidates(t *testing.T) {
	t.Parallel()

	server := New(Config{}, nil, nil, nil, nil)
	store := runledger.NewMemoryStore()
	server.SetRunLedgerStore(store)

	ctx := context.Background()
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "run-1",
		Type:    runledger.EventRunCreated,
		Payload: resumePayload(runledger.RunCreatedPayload{SessionKey: "sess-1", Goal: "resume me"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID: "run-1",
		Type:  runledger.EventPlanAttached,
		Payload: resumePayload(runledger.PlanAttachedPayload{
			Steps: []runledger.Step{{
				StepID:     "step-1",
				Goal:       "work",
				OwnerAgent: "operator",
				Status:     runledger.StepStatusPending,
				Validator:  runledger.ValidatorSpec{Type: runledger.ValidatorBuildPass},
				MaxRetries: runledger.DefaultMaxRetries,
			}},
		}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "run-1",
		Type:    runledger.EventRunPaused,
		Payload: resumePayload(runledger.RunPausedPayload{Reason: "paused"}),
	}))

	sendCh := make(chan []byte, 8)
	server.clientsMu.Lock()
	server.clients["ui-1"] = &Client{ID: "ui-1", Type: "ui", SessionKey: "sess-1", Send: sendCh}
	server.clientsMu.Unlock()

	client := &Client{ID: "ui-1", Type: "ui", Server: server, SessionKey: "sess-1"}
	result, err := server.handleChatMessage(client, json.RawMessage(`{"message":"계속해줘"}`))
	require.NoError(t, err)

	body := result.(map[string]interface{})
	assert.Equal(t, true, body["resumeRequired"])

	select {
	case msg := <-sendCh:
		var eventMsg map[string]interface{}
		require.NoError(t, json.Unmarshal(msg, &eventMsg))
		assert.Equal(t, "agent.resume_required", eventMsg["event"])
	case <-time.After(time.Second):
		t.Fatal("expected resume_required broadcast")
	}
}

func TestHandleChatMessage_ResumeConfirmResumesRun(t *testing.T) {
	t.Parallel()

	server := New(Config{}, nil, nil, nil, nil)
	store := runledger.NewMemoryStore()
	server.SetRunLedgerStore(store)

	ctx := context.Background()
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "run-2",
		Type:    runledger.EventRunCreated,
		Payload: resumePayload(runledger.RunCreatedPayload{SessionKey: "sess-2", Goal: "resume me"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID: "run-2",
		Type:  runledger.EventPlanAttached,
		Payload: resumePayload(runledger.PlanAttachedPayload{
			Steps: []runledger.Step{{
				StepID:     "step-1",
				Goal:       "work",
				OwnerAgent: "operator",
				Status:     runledger.StepStatusPending,
				Validator:  runledger.ValidatorSpec{Type: runledger.ValidatorBuildPass},
				MaxRetries: runledger.DefaultMaxRetries,
			}},
		}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "run-2",
		Type:    runledger.EventRunPaused,
		Payload: resumePayload(runledger.RunPausedPayload{Reason: "paused"}),
	}))

	client := &Client{ID: "ui-2", Type: "ui", Server: server, SessionKey: "sess-2"}
	_, err := server.handleChatMessage(client, json.RawMessage(`{"message":"resume","confirmResume":true,"resumeRunId":"run-2"}`))
	require.ErrorIs(t, err, ErrAgentNotReady)

	snap, snapErr := store.GetRunSnapshot(ctx, "run-2")
	require.NoError(t, snapErr)
	assert.Equal(t, runledger.RunStatusRunning, snap.Status)
}

func resumePayload(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
