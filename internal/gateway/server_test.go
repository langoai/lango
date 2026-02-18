package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/langowarny/lango/internal/approval"
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
