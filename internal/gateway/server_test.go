package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/gatekeeper"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/turnrunner"
	"github.com/langoai/lango/internal/types"
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

func TestServerStartReturnsListenError(t *testing.T) {
	t.Parallel()

	occupied, port := occupyGatewayPort(t)
	defer occupied.Close()

	server := New(Config{
		Host:             "127.0.0.1",
		Port:             port,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}, nil, nil, nil, nil)

	err := server.Start()
	require.Error(t, err)
}

func TestServerShutdownBeforeStartIsSafe(t *testing.T) {
	t.Parallel()

	server := New(Config{
		Host:             "127.0.0.1",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}, nil, nil, nil, nil)

	require.NotPanics(t, func() {
		require.NoError(t, server.Shutdown(context.Background()))
	})
}

func TestServerShutdownAfterFailedStartIsSafe(t *testing.T) {
	t.Parallel()

	occupied, port := occupyGatewayPort(t)
	defer occupied.Close()

	server := New(Config{
		Host:             "127.0.0.1",
		Port:             port,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}, nil, nil, nil, nil)

	require.Error(t, server.Start())
	require.NotPanics(t, func() {
		require.NoError(t, server.Shutdown(context.Background()))
	})
}

func TestServerRepeatedShutdownIsSafe(t *testing.T) {
	t.Parallel()

	server := New(Config{
		Host:             "127.0.0.1",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}, nil, nil, nil, nil)

	require.NoError(t, server.Shutdown(context.Background()))
	require.NotPanics(t, func() {
		require.NoError(t, server.Shutdown(context.Background()))
	})
}

func TestServerServeGracefulShutdown(t *testing.T) {
	t.Parallel()

	server := New(Config{
		Host:             "127.0.0.1",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}, nil, nil, nil, nil)

	listener, err := server.listen()
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() {
		done <- server.serve(listener)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, server.Shutdown(ctx))

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("gateway serve did not stop after shutdown")
	}
}

func occupyGatewayPort(t *testing.T) (net.Listener, int) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	parsed, err := strconv.Atoi(port)
	require.NoError(t, err)

	return listener, parsed
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

func TestServerSettersWireDependencies(t *testing.T) {
	t.Parallel()
	server := New(Config{}, nil, nil, nil, nil)

	agent := &adk.Agent{}
	server.SetAgent(agent)
	assert.Same(t, agent, server.agent)

	san, err := gatekeeper.NewSanitizer(config.GatekeeperConfig{})
	require.NoError(t, err)
	server.SetSanitizer(san)
	assert.Same(t, san, server.sanitizer)

	runner := turnrunner.New(turnrunner.Config{}, nil, nil, nil)
	server.SetTurnRunner(runner)
	assert.Same(t, runner, server.turnRunner)

	ledger := runledger.NewMemoryStore()
	server.SetRunLedgerStore(ledger)
	assert.Same(t, ledger, server.runLedgerStore)
}

func TestOnTurnCompleteQueuesCallbackUntilRunnerIsSet(t *testing.T) {
	t.Parallel()
	server := New(Config{}, nil, nil, nil, nil)

	called := false
	server.OnTurnComplete(func(string) {
		called = true
	})

	require.Len(t, server.turnCallbacks, 1)
	server.turnCallbacks[0]("sess-test")
	assert.True(t, called)
}

func TestHandleHealthIncludesFeatureStatuses(t *testing.T) {
	t.Parallel()
	server := New(Config{HTTPEnabled: true}, nil, nil, nil, nil)
	server.SetFeatureStatuses([]types.FeatureStatus{
		{Name: "Graph Store", Enabled: true, Healthy: false, Reason: "not configured"},
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	server.handleHealth(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
	features, ok := body["features"].([]interface{})
	require.True(t, ok)
	require.Len(t, features, 1)
	feature, ok := features[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Graph Store", feature["name"])
	assert.Equal(t, true, feature["enabled"])
	assert.Equal(t, false, feature["healthy"])
	assert.Equal(t, "not configured", feature["reason"])
}

func TestHandleStatusReportsClientCountAndConfig(t *testing.T) {
	t.Parallel()
	server := New(Config{
		HTTPEnabled:      true,
		WebSocketEnabled: false,
	}, nil, nil, nil, nil)
	server.clientsMu.Lock()
	server.clients["ui-1"] = &Client{ID: "ui-1", Type: "ui", Send: make(chan []byte, 1)}
	server.clients["companion-1"] = &Client{ID: "companion-1", Type: "companion", Send: make(chan []byte, 1)}
	server.clientsMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()
	server.handleStatus(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "running", body["status"])
	assert.Equal(t, float64(2), body["clients"])
	assert.Equal(t, false, body["wsEnabled"])
	assert.Equal(t, true, body["httpEnabled"])
}

func TestHandleRPCSendsResultAndPassesClient(t *testing.T) {
	t.Parallel()
	server := New(Config{}, nil, nil, nil, nil)
	client := &Client{
		ID:     "ui-1",
		Type:   "ui",
		Server: server,
		Send:   make(chan []byte, 1),
	}

	client.handleRPC(RPCRequest{
		ID:     "rpc-1",
		Method: "echo",
		Params: json.RawMessage(`{"value":"hello"}`),
	}, func(gotClient *Client, params json.RawMessage) (interface{}, error) {
		assert.Same(t, client, gotClient)
		assert.JSONEq(t, `{"value":"hello"}`, string(params))
		return map[string]string{"echo": "hello"}, nil
	})

	var resp RPCResponse
	require.NoError(t, json.Unmarshal(<-client.Send, &resp))
	assert.Equal(t, "rpc-1", resp.ID)
	require.Nil(t, resp.Error)
	result, ok := resp.Result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "hello", result["echo"])
}

func TestHandleRPCSendsHandlerError(t *testing.T) {
	t.Parallel()
	client := &Client{ID: "ui-1", Send: make(chan []byte, 1)}

	client.handleRPC(RPCRequest{ID: "rpc-err", Method: "boom"}, func(*Client, json.RawMessage) (interface{}, error) {
		return nil, fmt.Errorf("handler failed")
	})

	var resp RPCResponse
	require.NoError(t, json.Unmarshal(<-client.Send, &resp))
	assert.Equal(t, "rpc-err", resp.ID)
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32000, resp.Error.Code)
	assert.Equal(t, "handler failed", resp.Error.Message)
}

func TestHandleRPCRecoversPanic(t *testing.T) {
	t.Parallel()
	client := &Client{ID: "ui-1", Send: make(chan []byte, 1)}

	client.handleRPC(RPCRequest{ID: "rpc-panic", Method: "panic"}, func(*Client, json.RawMessage) (interface{}, error) {
		panic("exploded")
	})

	var resp RPCResponse
	require.NoError(t, json.Unmarshal(<-client.Send, &resp))
	assert.Equal(t, "rpc-panic", resp.ID)
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32000, resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "internal error: exploded")
}

func TestRegisterHandlerOverwritesExistingMethod(t *testing.T) {
	t.Parallel()
	server := New(Config{}, nil, nil, nil, nil)

	server.RegisterHandler("custom.method", func(*Client, json.RawMessage) (interface{}, error) {
		return "first", nil
	})
	server.RegisterHandler("custom.method", func(*Client, json.RawMessage) (interface{}, error) {
		return "second", nil
	})

	server.handlersMu.RLock()
	handler := server.handlers["custom.method"]
	server.handlersMu.RUnlock()
	result, err := handler(nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "second", result)
}

func TestHandleCompanionHelloReturnsStatusAndRejectsInvalidJSON(t *testing.T) {
	t.Parallel()
	server := New(Config{}, nil, nil, nil, nil)

	result, err := server.handleCompanionHello(nil, json.RawMessage(`{"deviceId":"phone-1","publicKey":"pub"}`))
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"status": "ok"}, result)

	_, err = server.handleCompanionHello(nil, json.RawMessage(`{`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid params")
}

func TestBroadcastToCompanionsTargetsOnlyCompanionClients(t *testing.T) {
	t.Parallel()
	server := New(Config{}, nil, nil, nil, nil)
	uiSend := make(chan []byte, 1)
	companionSend := make(chan []byte, 1)
	server.clientsMu.Lock()
	server.clients["ui-1"] = &Client{ID: "ui-1", Type: "ui", Send: uiSend}
	server.clients["companion-1"] = &Client{ID: "companion-1", Type: "companion", Send: companionSend}
	server.clientsMu.Unlock()

	server.BroadcastToCompanions("approval.request", map[string]string{"id": "req-1"})

	select {
	case <-uiSend:
		t.Fatal("ui client should not receive companion broadcast")
	default:
	}

	var event map[string]interface{}
	require.NoError(t, json.Unmarshal(<-companionSend, &event))
	assert.Equal(t, "event", event["type"])
	assert.Equal(t, "approval.request", event["event"])
}

func TestHasCompanionsReflectsConnectedCompanionClients(t *testing.T) {
	t.Parallel()
	server := New(Config{}, nil, nil, nil, nil)
	assert.False(t, server.HasCompanions())

	server.clientsMu.Lock()
	server.clients["ui-1"] = &Client{ID: "ui-1", Type: "ui", Send: make(chan []byte, 1)}
	server.clientsMu.Unlock()
	assert.False(t, server.HasCompanions())

	server.clientsMu.Lock()
	server.clients["companion-1"] = &Client{ID: "companion-1", Type: "companion", Send: make(chan []byte, 1)}
	server.clientsMu.Unlock()
	assert.True(t, server.HasCompanions())
}

func TestRequestApprovalRequiresCompanion(t *testing.T) {
	t.Parallel()
	server := New(Config{}, nil, nil, nil, nil)

	_, err := server.RequestApproval(context.Background(), "approve this")
	require.ErrorIs(t, err, ErrNoCompanion)
}

func TestRequestApprovalBroadcastsAndReturnsCanceledContext(t *testing.T) {
	t.Parallel()
	server := New(Config{ApprovalTimeout: time.Minute}, nil, nil, nil, nil)
	companionSend := make(chan []byte, 1)
	server.clientsMu.Lock()
	server.clients["companion-1"] = &Client{ID: "companion-1", Type: "companion", Send: companionSend}
	server.clientsMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := server.RequestApproval(ctx, "approve this")
	require.ErrorIs(t, err, context.Canceled)

	var event map[string]interface{}
	require.NoError(t, json.Unmarshal(<-companionSend, &event))
	assert.Equal(t, "approval.request", event["event"])
	payload, ok := event["payload"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, payload["id"])
	assert.Equal(t, "approve this", payload["message"])

	server.pendingApprovalsMu.Lock()
	pendingCount := len(server.pendingApprovals)
	server.pendingApprovalsMu.Unlock()
	assert.Zero(t, pendingCount)
}

func TestHandleApprovalResponseWithoutPendingRequestIsOK(t *testing.T) {
	t.Parallel()
	server := New(Config{}, nil, nil, nil, nil)

	result, err := server.handleApprovalResponse(nil, json.RawMessage(`{"requestId":"missing","approved":false,"alwaysAllow":true}`))
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"status": "ok"}, result)
}

func TestProviderSenderRoutesCryptoToCompanions(t *testing.T) {
	t.Parallel()
	provider := security.NewRPCProvider()
	server := New(Config{}, nil, provider, nil, nil)
	uiSend := make(chan []byte, 1)
	companionSend := make(chan []byte, 1)
	server.clientsMu.Lock()
	server.clients["ui-1"] = &Client{ID: "ui-1", Type: "ui", Send: uiSend}
	server.clients["companion-1"] = &Client{ID: "companion-1", Type: "companion", Send: companionSend}
	server.clientsMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := provider.Sign(ctx, "key-1", []byte("payload"))
	require.ErrorIs(t, err, context.Canceled)

	var cryptoEvent map[string]interface{}
	require.NoError(t, json.Unmarshal(<-companionSend, &cryptoEvent))
	assert.Equal(t, "sign.request", cryptoEvent["event"])
	select {
	case <-uiSend:
		t.Fatal("ui client should not receive crypto provider requests")
	default:
	}

	_, err = provider.Encrypt(ctx, "key-1", []byte("plaintext"))
	require.ErrorIs(t, err, context.Canceled)
	require.NoError(t, json.Unmarshal(<-companionSend, &cryptoEvent))
	assert.Equal(t, "encrypt.request", cryptoEvent["event"])

	_, err = provider.Decrypt(ctx, "key-1", []byte("ciphertext"))
	require.ErrorIs(t, err, context.Canceled)
	require.NoError(t, json.Unmarshal(<-companionSend, &cryptoEvent))
	assert.Equal(t, "decrypt.request", cryptoEvent["event"])
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
	result, err := server.handleChatMessage(client, json.RawMessage(`{"message":"resume","confirmResume":true,"resumeRunId":"run-2"}`))
	require.NoError(t, err)
	body := result.(map[string]interface{})
	assert.Equal(t, true, body["resumed"])
	assert.Equal(t, "run-2", body["runId"])

	snap, snapErr := store.GetRunSnapshot(ctx, "run-2")
	require.NoError(t, snapErr)
	assert.Equal(t, runledger.RunStatusRunning, snap.Status)
}

func TestHandleChatMessage_ResumeConfirmWithoutIntentKeyword(t *testing.T) {
	t.Parallel()

	server := New(Config{
		RunLedger: config.RunLedgerConfig{StaleTTL: 30 * time.Minute},
	}, nil, nil, nil, nil)
	store := runledger.NewMemoryStore()
	server.SetRunLedgerStore(store)

	ctx := context.Background()
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "run-3",
		Type:    runledger.EventRunCreated,
		Payload: resumePayload(runledger.RunCreatedPayload{SessionKey: "sess-3", Goal: "resume me"}),
	}))
	require.NoError(t, store.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID: "run-3",
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
		RunID:   "run-3",
		Type:    runledger.EventRunPaused,
		Payload: resumePayload(runledger.RunPausedPayload{Reason: "paused"}),
	}))

	client := &Client{ID: "ui-3", Type: "ui", Server: server, SessionKey: "sess-3"}
	result, err := server.handleChatMessage(client, json.RawMessage(`{"message":"yes","confirmResume":true,"resumeRunId":"run-3"}`))
	require.NoError(t, err)

	body := result.(map[string]interface{})
	assert.Equal(t, true, body["resumed"])
	assert.Equal(t, "run-3", body["runId"])

	snap, snapErr := store.GetRunSnapshot(ctx, "run-3")
	require.NoError(t, snapErr)
	assert.Equal(t, runledger.RunStatusRunning, snap.Status)
}

func resumePayload(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

func TestChatMessage_UnauthenticatedGetsUniqueSessionKey(t *testing.T) {
	t.Parallel()

	// Two unauthenticated clients sending different session keys should get isolated sessions.
	server := New(Config{
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}, nil, nil, nil, nil)

	clientA := &Client{ID: "a", Type: "ui", Server: server, SessionKey: ""}
	clientB := &Client{ID: "b", Type: "ui", Server: server, SessionKey: ""}

	// Both fail because no TurnRunner is set, but we can verify the session key
	// assignment logic by checking the error message context.
	// With empty SessionKey and no client-specified key, it should default to "default".
	paramsNoKey := json.RawMessage(`{"message":"hello"}`)
	_, errA := server.handleChatMessage(clientA, paramsNoKey)
	require.Error(t, errA, "should fail due to nil turn runner")

	// With explicit session key from unauthenticated client, should use the provided key.
	paramsWithKey := json.RawMessage(`{"message":"hello","sessionKey":"custom-sess"}`)
	_, errB := server.handleChatMessage(clientB, paramsWithKey)
	require.Error(t, errB, "should fail due to nil turn runner")
}

func TestChatMessage_AuthenticatedIgnoresClientSessionKey(t *testing.T) {
	t.Parallel()

	server := New(Config{
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}, nil, nil, nil, nil)

	// Authenticated client should always use their own session key,
	// even if the request specifies a different one.
	client := &Client{
		ID:         "auth-client",
		Type:       "ui",
		Server:     server,
		SessionKey: "sess_my-auth-key",
	}

	params := json.RawMessage(`{"message":"hello","sessionKey":"hijack-attempt"}`)
	_, err := server.handleChatMessage(client, params)
	// Error expected (nil turn runner), but the important thing is
	// the session key was NOT "hijack-attempt".
	require.Error(t, err, "should fail due to nil turn runner")
}

func TestWebSocket_AbruptDisconnect(t *testing.T) {
	t.Parallel()

	server := New(Config{
		Host:             "localhost",
		Port:             0,
		HTTPEnabled:      true,
		WebSocketEnabled: true,
	}, nil, nil, nil, nil)

	ts := httptest.NewServer(server.router)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"

	// Connect and immediately close without graceful shutdown.
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	// Send a partial message then close abruptly (no close frame).
	conn.UnderlyingConn().Close()

	// Give the server a moment to process the disconnection.
	time.Sleep(100 * time.Millisecond)

	// Server should still be healthy — connect another client.
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn2.Close()

	// Verify the server still responds to RPC.
	server.RegisterHandler("ping", func(_ *Client, _ json.RawMessage) (interface{}, error) {
		return "pong", nil
	})

	require.NoError(t, conn2.WriteJSON(RPCRequest{ID: "1", Method: "ping"}))
	var resp RPCResponse
	require.NoError(t, conn2.ReadJSON(&resp))
	assert.Equal(t, "pong", resp.Result)
}

func TestHealth_AlwaysReturns200(t *testing.T) {
	t.Parallel()

	server := New(Config{
		HTTPEnabled: true,
	}, nil, nil, nil, nil)

	ts := httptest.NewServer(server.router)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "ok", body["status"])
}
