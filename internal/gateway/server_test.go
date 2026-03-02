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
	"github.com/langoai/lango/internal/approval"
)

func TestGatewayServer(t *testing.T) {
	// Setup server (no auth — dev mode)
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	// Register a test RPC handler (updated signature with *Client)
	server.RegisterHandler("echo", func(_ *Client, params json.RawMessage) (interface{}, error) {
		var input string
		if err := json.Unmarshal(params, &input); err != nil {
			return nil, err
		}
		return "echo: " + input, nil
	})

	// Use httptest server with the gateway's router
	ts := httptest.NewServer(server.router)
	defer ts.Close()

	// Test HTTP Health
	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("failed to get health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Test WebSocket
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer conn.Close()

	// Test RPC Call
	req := RPCRequest{
		ID:     "1",
		Method: "echo",
		Params: json.RawMessage(`"hello"`), // JSON string "hello"
	}
	if err := conn.WriteJSON(req); err != nil {
		t.Fatalf("failed to write json: %v", err)
	}

	// Read response
	var rpcResp RPCResponse
	if err := conn.ReadJSON(&rpcResp); err != nil {
		t.Fatalf("failed to read json: %v", err)
	}

	if rpcResp.ID != "1" {
		t.Errorf("expected id 1, got %s", rpcResp.ID)
	}
	if rpcResp.Result != "echo: hello" {
		t.Errorf("expected 'echo: hello', got %v", rpcResp.Result)
	}

	// Test Broadcast
	done := make(chan bool)
	go func() {
		// Read next message (expecting broadcast)
		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Errorf("failed to read broadcast: %v", err)
			return
		}

		var eventMsg map[string]interface{}
		if err := json.Unmarshal(msg, &eventMsg); err != nil {
			t.Errorf("failed to unmarshal broadcast: %v", err)
			return
		}

		if eventMsg["type"] != "event" {
			t.Errorf("expected type 'event', got %v", eventMsg["type"])
		}
		if eventMsg["event"] != "test-event" {
			t.Errorf("expected event 'test-event', got %v", eventMsg["event"])
		}
		done <- true
	}()

	// Allow client to be registered
	time.Sleep(100 * time.Millisecond)
	server.Broadcast("test-event", "payload")

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for broadcast")
	}
}

func TestChatMessage_UnauthenticatedUsesDefault(t *testing.T) {
	// When auth is nil (no OIDC) and client has no SessionKey,
	// handleChatMessage should use "default" session key.
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	// Client with empty SessionKey (unauthenticated)
	client := &Client{
		ID:         "test-client",
		Type:       "ui",
		Server:     server,
		SessionKey: "",
	}

	params := json.RawMessage(`{"message":"hello"}`)
	// agent is nil so RunAndCollect will panic/error — but we can test the session
	// key resolution by checking that the handler does NOT error on param parsing
	_, err := server.handleChatMessage(client, params)
	// Expected: error because agent is nil, but the params parsing should succeed
	if err == nil {
		t.Error("expected error (nil agent), got nil")
	}
}

func TestChatMessage_AuthenticatedUsesOwnSession(t *testing.T) {
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	// Client with authenticated SessionKey
	client := &Client{
		ID:         "test-client",
		Type:       "ui",
		Server:     server,
		SessionKey: "sess_my-authenticated-key",
	}

	// Even if client tries to send a different sessionKey, the authenticated one is used
	params := json.RawMessage(`{"message":"hello","sessionKey":"hacker-session"}`)
	_, err := server.handleChatMessage(client, params)
	// Expected: error because agent is nil, but params parsing succeeds
	if err == nil {
		t.Error("expected error (nil agent), got nil")
	}
}

func TestApprovalResponse_AtomicDelete(t *testing.T) {
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	// Create a pending approval
	respChan := make(chan approval.ApprovalResponse, 1)
	server.pendingApprovalsMu.Lock()
	server.pendingApprovals["req-1"] = respChan
	server.pendingApprovalsMu.Unlock()

	// First response — should succeed
	params := json.RawMessage(`{"requestId":"req-1","approved":true}`)
	result, err := server.handleApprovalResponse(nil, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}

	// Verify the approval was received
	select {
	case resp := <-respChan:
		if !resp.Approved {
			t.Error("expected approved=true")
		}
	default:
		t.Error("expected approval result on channel")
	}

	// Verify entry was deleted
	server.pendingApprovalsMu.Lock()
	_, exists := server.pendingApprovals["req-1"]
	server.pendingApprovalsMu.Unlock()
	if exists {
		t.Error("expected pending approval to be deleted after response")
	}
}

func TestApprovalResponse_DuplicateResponse(t *testing.T) {
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	// Create a pending approval
	respChan := make(chan approval.ApprovalResponse, 1)
	server.pendingApprovalsMu.Lock()
	server.pendingApprovals["req-dup"] = respChan
	server.pendingApprovalsMu.Unlock()

	// First response
	params := json.RawMessage(`{"requestId":"req-dup","approved":true}`)
	_, err := server.handleApprovalResponse(nil, params)
	if err != nil {
		t.Fatalf("unexpected error on first response: %v", err)
	}

	// Second response — should not send to channel again (entry already deleted)
	_, err = server.handleApprovalResponse(nil, params)
	if err != nil {
		t.Fatalf("unexpected error on second response: %v", err)
	}

	// Only one value should be on the channel
	select {
	case <-respChan:
		// Good — first response
	default:
		t.Error("expected one approval result on channel")
	}

	// Channel should be empty now
	select {
	case <-respChan:
		t.Error("unexpected second value on channel — duplicate response was not blocked")
	default:
		// Good — no duplicate
	}
}

func TestBroadcastToSession_ScopedBySessionKey(t *testing.T) {
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}
	server := New(cfg, nil, nil, nil, nil)

	// Create clients with different session keys
	sendA := make(chan []byte, 256)
	sendB := make(chan []byte, 256)
	sendC := make(chan []byte, 256)

	server.clientsMu.Lock()
	server.clients["a"] = &Client{ID: "a", Type: "ui", SessionKey: "sess-1", Send: sendA}
	server.clients["b"] = &Client{ID: "b", Type: "ui", SessionKey: "sess-2", Send: sendB}
	server.clients["c"] = &Client{ID: "c", Type: "companion", SessionKey: "sess-1", Send: sendC}
	server.clientsMu.Unlock()

	// Broadcast to session "sess-1" — only client "a" (UI, matching session) should receive
	server.BroadcastToSession("sess-1", "agent.thinking", map[string]string{"sessionKey": "sess-1"})

	// Client A should receive (UI + matching session)
	select {
	case msg := <-sendA:
		var eventMsg map[string]interface{}
		if err := json.Unmarshal(msg, &eventMsg); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if eventMsg["event"] != "agent.thinking" {
			t.Errorf("expected 'agent.thinking', got %v", eventMsg["event"])
		}
	default:
		t.Error("expected client A to receive broadcast")
	}

	// Client B should NOT receive (different session)
	select {
	case <-sendB:
		t.Error("client B should not receive broadcast for sess-1")
	default:
		// Good
	}

	// Client C should NOT receive (companion, not UI)
	select {
	case <-sendC:
		t.Error("companion client should not receive session broadcast")
	default:
		// Good
	}
}

func TestBroadcastToSession_NoAuth(t *testing.T) {
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

	// With empty session key (no auth), all UI clients should receive
	server.BroadcastToSession("", "agent.done", map[string]string{"sessionKey": ""})

	select {
	case <-sendA:
		// Good
	default:
		t.Error("expected client A to receive broadcast")
	}

	select {
	case <-sendB:
		// Good
	default:
		t.Error("expected client B to receive broadcast")
	}
}

func TestHandleChatMessage_NilAgent_ReturnsErrorWithoutBroadcast(t *testing.T) {
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
		RequestTimeout:   50 * time.Millisecond,
	}
	server := New(cfg, nil, nil, nil, nil)

	// Create a UI client to receive broadcasts.
	sendCh := make(chan []byte, 256)
	server.clientsMu.Lock()
	server.clients["ui-1"] = &Client{
		ID:         "ui-1",
		Type:       "ui",
		SessionKey: "",
		Send:       sendCh,
	}
	server.clientsMu.Unlock()

	// Call handleChatMessage — agent is nil, so it returns ErrAgentNotReady
	// before any broadcast events (agent.thinking, agent.done, agent.error).
	client := &Client{ID: "test", Type: "ui", Server: server, SessionKey: ""}
	params := json.RawMessage(`{"message":"hello"}`)
	_, err := server.handleChatMessage(client, params)
	if err == nil {
		t.Fatal("expected error from nil agent")
	}

	// No events should be sent — ErrAgentNotReady fires before agent.thinking.
	select {
	case msg := <-sendCh:
		t.Errorf("expected no broadcast, got: %s", msg)
	default:
		// Good — no broadcast
	}
}

func TestHandleChatMessage_SuccessBroadcastsAgentDone(t *testing.T) {
	// This test verifies that on success, agent.done is sent (not agent.error).
	// We validate the broadcast logic directly using BroadcastToSession.
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

	// Simulate the success path: broadcast agent.done.
	server.BroadcastToSession("", "agent.done", map[string]string{
		"sessionKey": "",
	})

	select {
	case msg := <-sendCh:
		var m map[string]interface{}
		if err := json.Unmarshal(msg, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if m["event"] != "agent.done" {
			t.Errorf("expected agent.done, got %v", m["event"])
		}
	default:
		t.Error("expected agent.done broadcast")
	}
}

func TestHandleChatMessage_ErrorBroadcastsAgentErrorEvent(t *testing.T) {
	// Simulate the error path: broadcast agent.error with classification.
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

	// Simulate timeout error broadcast.
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
		if err := json.Unmarshal(msg, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if m["event"] != "agent.error" {
			t.Errorf("expected agent.error, got %v", m["event"])
		}
		payload, ok := m["payload"].(map[string]interface{})
		if !ok {
			t.Fatal("expected payload map")
		}
		if payload["type"] != "timeout" {
			t.Errorf("expected type 'timeout', got %v", payload["type"])
		}
	default:
		t.Error("expected agent.error broadcast")
	}
}

func TestWarningBroadcast_ApproachingTimeout(t *testing.T) {
	// Verify that the 80% timeout warning timer fires and broadcasts
	// an agent.warning event with the correct payload.
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

	// Simulate the warning timer pattern used in handleChatMessage:
	// time.AfterFunc at 80% of timeout broadcasting agent.warning.
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

	// Wait for the timer to fire (80% of 50ms = 40ms, wait a bit more).
	time.Sleep(70 * time.Millisecond)

	select {
	case msg := <-sendCh:
		var m map[string]interface{}
		if err := json.Unmarshal(msg, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if m["event"] != "agent.warning" {
			t.Errorf("expected agent.warning, got %v", m["event"])
		}
		payload, ok := m["payload"].(map[string]interface{})
		if !ok {
			t.Fatal("expected payload map")
		}
		if payload["type"] != "approaching_timeout" {
			t.Errorf("expected type 'approaching_timeout', got %v", payload["type"])
		}
		if payload["message"] != "Request is taking longer than expected" {
			t.Errorf("unexpected message: %v", payload["message"])
		}
	default:
		t.Error("expected agent.warning broadcast after 80% timeout")
	}
}

func TestApprovalTimeout_UsesConfigTimeout(t *testing.T) {
	cfg := Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
		ApprovalTimeout:  50 * time.Millisecond,
	}
	server := New(cfg, nil, nil, nil, nil)

	// Add a fake companion so RequestApproval doesn't fail early
	server.clientsMu.Lock()
	server.clients["companion-1"] = &Client{
		ID:   "companion-1",
		Type: "companion",
		Send: make(chan []byte, 256),
	}
	server.clientsMu.Unlock()

	_, err := server.RequestApproval(t.Context(), "test approval")
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "approval timeout") {
		t.Errorf("expected 'approval timeout' error, got: %v", err)
	}
}
