package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/p2p/firewall"
	"github.com/langoai/lango/internal/p2p/handshake"
)

// testHandler creates a Handler with pre-configured sessions and firewall.
func testHandler() (*Handler, *handshake.SessionStore) {
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	sessions, err := handshake.NewSessionStore(time.Hour)
	if err != nil {
		panic(fmt.Sprintf("create session store: %v", err))
	}

	fw := firewall.New([]firewall.ACLRule{
		{PeerDID: "did:key:peer-1", Action: firewall.ACLActionAllow, Tools: []string{firewall.WildcardAll}},
		{PeerDID: "did:key:peer-2", Action: firewall.ACLActionAllow, Tools: []string{firewall.WildcardAll}},
		{PeerDID: "did:key:peer-3", Action: firewall.ACLActionAllow, Tools: []string{firewall.WildcardAll}},
		{PeerDID: "did:key:peer-4", Action: firewall.ACLActionAllow, Tools: []string{firewall.WildcardAll}},
		{PeerDID: "did:key:peer-5", Action: firewall.ACLActionAllow, Tools: []string{firewall.WildcardAll}},
		{PeerDID: "did:key:peer-6", Action: firewall.ACLActionAllow, Tools: []string{firewall.WildcardAll}},
		{PeerDID: "did:key:peer-json", Action: firewall.ACLActionAllow, Tools: []string{firewall.WildcardAll}},
	}, sugar)

	h := NewHandler(HandlerConfig{
		Sessions: sessions,
		Firewall: fw,
		Executor: func(_ context.Context, toolName string, _ map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{"tool": toolName, "executed": true}, nil
		},
		LocalDID: "did:key:local",
		Logger:   sugar,
	})

	return h, sessions
}

// createSession adds a session and returns the token.
func createSession(sessions *handshake.SessionStore, peerDID string) string {
	sess, err := sessions.Create(peerDID, false)
	if err != nil {
		panic(fmt.Sprintf("create session: %v", err))
	}
	return sess.Token
}

func TestHandleToolInvoke_NilApprovalFn_DefaultDeny(t *testing.T) {
	t.Parallel()

	h, sessions := testHandler()
	// Do NOT set approvalFn — it stays nil.

	peerDID := "did:key:peer-1"
	token := createSession(sessions, peerDID)

	req := &Request{
		Type:         RequestToolInvoke,
		SessionToken: token,
		RequestID:    "req-1",
		Payload:      map[string]interface{}{"toolName": "echo"},
	}

	resp := h.handleRequest(context.Background(), nil, req)
	assert.Equal(t, ResponseStatusDenied, resp.Status)
	assert.Equal(t, ErrNoApprovalHandler.Error(), resp.Error)
}

func TestHandleToolInvokePaid_NilApprovalFn_DefaultDeny(t *testing.T) {
	t.Parallel()

	h, sessions := testHandler()
	// Do NOT set approvalFn.

	peerDID := "did:key:peer-2"
	token := createSession(sessions, peerDID)

	req := &Request{
		Type:         RequestToolInvokePaid,
		SessionToken: token,
		RequestID:    "req-2",
		Payload:      map[string]interface{}{"toolName": "paid_tool"},
	}

	resp := h.handleRequest(context.Background(), nil, req)
	assert.Equal(t, ResponseStatusDenied, resp.Status)
	assert.Equal(t, ErrNoApprovalHandler.Error(), resp.Error)
}

func TestHandleToolInvoke_Approved(t *testing.T) {
	t.Parallel()

	h, sessions := testHandler()
	h.SetApprovalFunc(func(_ context.Context, _, _ string, _ map[string]interface{}) (bool, error) {
		return true, nil
	})

	peerDID := "did:key:peer-3"
	token := createSession(sessions, peerDID)

	req := &Request{
		Type:         RequestToolInvoke,
		SessionToken: token,
		RequestID:    "req-3",
		Payload:      map[string]interface{}{"toolName": "echo"},
	}

	resp := h.handleRequest(context.Background(), nil, req)
	assert.Equal(t, ResponseStatusOK, resp.Status, "error: %s", resp.Error)
}

func TestHandleToolInvoke_Denied(t *testing.T) {
	t.Parallel()

	h, sessions := testHandler()
	h.SetApprovalFunc(func(_ context.Context, _, _ string, _ map[string]interface{}) (bool, error) {
		return false, nil
	})

	peerDID := "did:key:peer-4"
	token := createSession(sessions, peerDID)

	req := &Request{
		Type:         RequestToolInvoke,
		SessionToken: token,
		RequestID:    "req-4",
		Payload:      map[string]interface{}{"toolName": "exec"},
	}

	resp := h.handleRequest(context.Background(), nil, req)
	assert.Equal(t, ResponseStatusDenied, resp.Status)
	assert.Equal(t, ErrDeniedByOwner.Error(), resp.Error)
}

func TestHandleToolInvoke_ApprovalError(t *testing.T) {
	t.Parallel()

	h, sessions := testHandler()
	h.SetApprovalFunc(func(_ context.Context, _, _ string, _ map[string]interface{}) (bool, error) {
		return false, fmt.Errorf("approval service unavailable")
	})

	peerDID := "did:key:peer-5"
	token := createSession(sessions, peerDID)

	req := &Request{
		Type:         RequestToolInvoke,
		SessionToken: token,
		RequestID:    "req-5",
		Payload:      map[string]interface{}{"toolName": "echo"},
	}

	resp := h.handleRequest(context.Background(), nil, req)
	assert.Equal(t, ResponseStatusError, resp.Status)
}

func TestHandleToolInvokePaid_Approved(t *testing.T) {
	t.Parallel()

	h, sessions := testHandler()
	h.SetApprovalFunc(func(_ context.Context, _, _ string, _ map[string]interface{}) (bool, error) {
		return true, nil
	})

	peerDID := "did:key:peer-6"
	token := createSession(sessions, peerDID)

	req := &Request{
		Type:         RequestToolInvokePaid,
		SessionToken: token,
		RequestID:    "req-6",
		Payload:      map[string]interface{}{"toolName": "paid_echo"},
	}

	resp := h.handleRequest(context.Background(), nil, req)
	assert.Equal(t, ResponseStatusOK, resp.Status, "error: %s", resp.Error)
}

func TestResponseJSON_DefaultDeny(t *testing.T) {
	t.Parallel()

	h, sessions := testHandler()
	peerDID := "did:key:peer-json"
	token := createSession(sessions, peerDID)

	req := &Request{
		Type:         RequestToolInvoke,
		SessionToken: token,
		RequestID:    "req-json",
		Payload:      map[string]interface{}{"toolName": "echo"},
	}

	resp := h.handleRequest(context.Background(), nil, req)

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded Response
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, ResponseStatusDenied, decoded.Status)
}
