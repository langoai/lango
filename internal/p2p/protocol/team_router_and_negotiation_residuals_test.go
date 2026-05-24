package protocol

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamRouterHandleDispatchesTypedPayloadsAndErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	peerDID := "did:key:team-router-peer"
	deadline := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)

	var invite TeamInvitePayload
	var accept TeamAcceptPayload
	var task TeamTaskPayload
	var result TeamResultPayload
	var disband TeamDisbandPayload
	router := &TeamRouter{
		OnInvite: func(_ context.Context, gotPeer string, payload TeamInvitePayload) (map[string]interface{}, error) {
			assert.Equal(t, peerDID, gotPeer)
			invite = payload
			return map[string]interface{}{"route": "invite"}, nil
		},
		OnAccept: func(_ context.Context, gotPeer string, payload TeamAcceptPayload) (map[string]interface{}, error) {
			assert.Equal(t, peerDID, gotPeer)
			accept = payload
			return map[string]interface{}{"route": "accept"}, nil
		},
		OnTask: func(_ context.Context, gotPeer string, payload TeamTaskPayload) (map[string]interface{}, error) {
			assert.Equal(t, peerDID, gotPeer)
			task = payload
			return map[string]interface{}{"route": "task"}, nil
		},
		OnResult: func(_ context.Context, gotPeer string, payload TeamResultPayload) (map[string]interface{}, error) {
			assert.Equal(t, peerDID, gotPeer)
			result = payload
			return map[string]interface{}{"route": "result"}, nil
		},
		OnDisband: func(_ context.Context, gotPeer string, payload TeamDisbandPayload) (map[string]interface{}, error) {
			assert.Equal(t, peerDID, gotPeer)
			disband = payload
			return map[string]interface{}{"route": "disband"}, nil
		},
	}

	got, err := router.Handle(ctx, peerDID, RequestTeamInvite, map[string]interface{}{
		"teamId":       "team-1",
		"teamName":     "Coverage Team",
		"goal":         "raise coverage",
		"leaderDid":    "did:key:leader",
		"role":         "tester",
		"capabilities": []string{"go-test", "review"},
	})
	require.NoError(t, err)
	assert.Equal(t, "invite", got["route"])
	assert.Equal(t, "team-1", invite.TeamID)
	assert.Equal(t, []string{"go-test", "review"}, invite.Capabilities)

	got, err = router.Handle(ctx, peerDID, RequestTeamAccept, map[string]interface{}{
		"teamId":    "team-1",
		"memberDid": "did:key:member",
		"accepted":  true,
		"reason":    "available",
	})
	require.NoError(t, err)
	assert.Equal(t, "accept", got["route"])
	assert.True(t, accept.Accepted)
	assert.Equal(t, "available", accept.Reason)

	got, err = router.Handle(ctx, peerDID, RequestTeamTask, map[string]interface{}{
		"teamId":   "team-1",
		"taskId":   "task-1",
		"toolName": "go_test",
		"params":   map[string]interface{}{"package": "./internal/p2p/protocol"},
		"deadline": deadline.Format(time.RFC3339),
	})
	require.NoError(t, err)
	assert.Equal(t, "task", got["route"])
	assert.Equal(t, "go_test", task.ToolName)
	assert.Equal(t, "./internal/p2p/protocol", task.Params["package"])
	assert.Equal(t, deadline, task.Deadline)

	got, err = router.Handle(ctx, peerDID, RequestTeamResult, map[string]interface{}{
		"teamId":     "team-1",
		"taskId":     "task-1",
		"memberDid":  "did:key:member",
		"result":     map[string]interface{}{"ok": true},
		"durationMs": int64(125),
	})
	require.NoError(t, err)
	assert.Equal(t, "result", got["route"])
	assert.Equal(t, true, result.Result["ok"])
	assert.Equal(t, int64(125), result.Duration)

	got, err = router.Handle(ctx, peerDID, RequestTeamDisband, map[string]interface{}{
		"teamId": "team-1",
		"reason": "done",
	})
	require.NoError(t, err)
	assert.Equal(t, "disband", got["route"])
	assert.Equal(t, "done", disband.Reason)

	_, err = (&TeamRouter{}).Handle(ctx, peerDID, RequestTeamInvite, map[string]interface{}{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "team invite handler not configured")

	_, err = (&TeamRouter{OnAccept: router.OnAccept}).Handle(ctx, peerDID, RequestTeamAccept, map[string]interface{}{"accepted": func() {}})
	require.Error(t, err)
	assert.ErrorContains(t, err, "marshal team payload")

	_, err = (&TeamRouter{OnTask: router.OnTask}).Handle(ctx, peerDID, RequestTeamTask, map[string]interface{}{"teamId": 42})
	require.Error(t, err)
	assert.ErrorContains(t, err, "decode team task")

	_, err = router.Handle(ctx, peerDID, RequestType("team_unknown"), map[string]interface{}{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "unknown team request type")
}

func TestHandlerHandleNegotiateBranches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	req := &Request{
		Type:      RequestNegotiatePropose,
		RequestID: "negotiation-1",
		Payload: map[string]interface{}{
			"action":   "propose",
			"toolName": "summarize",
			"price":    "0.100000",
		},
	}
	h := NewHandler(HandlerConfig{})

	missing := h.handleNegotiate(ctx, req, "did:key:peer")
	assert.Equal(t, ResponseStatusError, missing.Status)
	assert.Equal(t, "negotiation not configured", missing.Error)

	var seenPayload NegotiatePayload
	h.SetNegotiator(func(_ context.Context, peerDID string, payload NegotiatePayload) (map[string]interface{}, error) {
		assert.Equal(t, "did:key:peer", peerDID)
		seenPayload = payload
		return map[string]interface{}{"sessionId": "session-1", "phase": "proposed"}, nil
	})

	ok := h.handleNegotiate(ctx, req, "did:key:peer")
	assert.Equal(t, ResponseStatusOK, ok.Status)
	assert.Equal(t, "session-1", ok.Result["sessionId"])
	assert.Equal(t, "propose", seenPayload.Action)
	assert.Equal(t, "summarize", seenPayload.ToolName)
	assert.Equal(t, "0.100000", seenPayload.Price)

	h.SetNegotiator(func(context.Context, string, NegotiatePayload) (map[string]interface{}, error) {
		return nil, errors.New("negotiation backend unavailable")
	})
	failed := h.handleNegotiate(ctx, req, "did:key:peer")
	assert.Equal(t, ResponseStatusError, failed.Status)
	assert.Equal(t, "negotiation backend unavailable", failed.Error)

	h.SetNegotiator(func(_ context.Context, _ string, payload NegotiatePayload) (map[string]interface{}, error) {
		assert.Empty(t, payload.Action)
		assert.Empty(t, payload.ToolName)
		return map[string]interface{}{"decoded": false}, nil
	})
	tolerant := h.handleNegotiate(ctx, &Request{
		Type:      RequestNegotiateRespond,
		RequestID: "negotiation-bad-payload",
		Payload:   map[string]interface{}{"action": func() {}},
	}, "did:key:peer")
	assert.Equal(t, ResponseStatusOK, tolerant.Status)
	assert.Equal(t, false, tolerant.Result["decoded"])
}
