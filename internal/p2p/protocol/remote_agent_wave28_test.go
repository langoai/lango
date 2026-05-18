package protocol

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	ma "github.com/multiformats/go-multiaddr"
)

var _ host.Host = (*wave28Host)(nil)
var _ network.Stream = (*wave28Stream)(nil)

type wave28Host struct {
	stream       network.Stream
	err          error
	calls        int
	gotPeer      peer.ID
	gotProtocols []libp2pprotocol.ID
}

func (h *wave28Host) ID() peer.ID                                  { return peer.ID("wave28-local") }
func (h *wave28Host) Peerstore() peerstore.Peerstore               { return nil }
func (h *wave28Host) Addrs() []ma.Multiaddr                        { return nil }
func (h *wave28Host) Network() network.Network                     { return nil }
func (h *wave28Host) Mux() libp2pprotocol.Switch                   { return nil }
func (h *wave28Host) Connect(context.Context, peer.AddrInfo) error { return nil }
func (h *wave28Host) SetStreamHandler(libp2pprotocol.ID, network.StreamHandler) {
}
func (h *wave28Host) SetStreamHandlerMatch(libp2pprotocol.ID, func(libp2pprotocol.ID) bool, network.StreamHandler) {
}
func (h *wave28Host) RemoveStreamHandler(libp2pprotocol.ID) {}
func (h *wave28Host) NewStream(ctx context.Context, p peer.ID, pids ...libp2pprotocol.ID) (network.Stream, error) {
	h.calls++
	h.gotPeer = p
	h.gotProtocols = append([]libp2pprotocol.ID(nil), pids...)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if h.err != nil {
		return nil, h.err
	}
	return h.stream, nil
}
func (h *wave28Host) Close() error                     { return nil }
func (h *wave28Host) ConnManager() connmgr.ConnManager { return nil }
func (h *wave28Host) EventBus() event.Bus              { return nil }

type wave28Stream struct {
	reader            io.Reader
	writer            io.Writer
	responseTemplate  *Response
	responseRequestID string
	onWrite           func()
	closed            bool
}

func (s *wave28Stream) Read(p []byte) (int, error) {
	if s.reader == nil {
		return 0, io.EOF
	}
	return s.reader.Read(p)
}

func (s *wave28Stream) Write(p []byte) (int, error) {
	if s.writer == nil {
		if s.onWrite != nil {
			s.onWrite()
		}
		return len(p), nil
	}
	n, err := s.writer.Write(p)
	if err != nil {
		return n, err
	}
	if s.responseTemplate != nil {
		var req Request
		if decodeErr := json.Unmarshal(p, &req); decodeErr == nil {
			resp := *s.responseTemplate
			resp.RequestID = req.RequestID
			s.responseRequestID = req.RequestID
			var response bytes.Buffer
			if encodeErr := json.NewEncoder(&response).Encode(resp); encodeErr == nil {
				s.reader = &response
			}
		}
	}
	if s.onWrite != nil {
		s.onWrite()
	}
	return n, nil
}

func (s *wave28Stream) Close() error {
	s.closed = true
	return nil
}

func (s *wave28Stream) CloseWrite() error                            { return nil }
func (s *wave28Stream) CloseRead() error                             { return nil }
func (s *wave28Stream) Reset() error                                 { s.closed = true; return nil }
func (s *wave28Stream) ResetWithError(network.StreamErrorCode) error { s.closed = true; return nil }
func (s *wave28Stream) SetDeadline(time.Time) error                  { return nil }
func (s *wave28Stream) SetReadDeadline(time.Time) error              { return nil }
func (s *wave28Stream) SetWriteDeadline(time.Time) error             { return nil }
func (s *wave28Stream) ID() string                                   { return "wave28-stream" }
func (s *wave28Stream) Protocol() libp2pprotocol.ID                  { return ProtocolID }
func (s *wave28Stream) SetProtocol(libp2pprotocol.ID) error          { return nil }
func (s *wave28Stream) Stat() network.Stats                          { return network.Stats{} }
func (s *wave28Stream) Conn() network.Conn                           { return &wave10ProtocolConn{} }
func (s *wave28Stream) Scope() network.StreamScope                   { return &network.NullScope{} }

func wave28AgentWithResponse(t *testing.T, resp Response) (*P2PRemoteAgent, *wave28Host, *wave28Stream, *bytes.Buffer) {
	t.Helper()

	sent := &bytes.Buffer{}
	stream := &wave28Stream{writer: sent, responseTemplate: &resp}
	fakeHost := &wave28Host{stream: stream}
	agent := NewRemoteAgent(RemoteAgentConfig{
		Name:         "wave28-remote",
		DID:          "did:key:wave28-remote",
		PeerID:       peer.ID("wave28-peer"),
		SessionToken: "wave28-token",
		Host:         fakeHost,
		Logger:       zap.NewNop().Sugar(),
	})

	return agent, fakeHost, stream, sent
}

func wave28DecodeRequest(t *testing.T, sent *bytes.Buffer) Request {
	t.Helper()

	var req Request
	require.NoError(t, json.NewDecoder(sent).Decode(&req))
	require.NotEmpty(t, req.RequestID)
	return req
}

func wave28AssertOpenedAndClosed(t *testing.T, h *wave28Host, s *wave28Stream) {
	t.Helper()

	require.Equal(t, 1, h.calls)
	assert.Equal(t, peer.ID("wave28-peer"), h.gotPeer)
	assert.Equal(t, []libp2pprotocol.ID{ProtocolID}, h.gotProtocols)
	assert.True(t, s.closed)
}

func TestWave28RemoteAgentInvokeToolSendsRequestAndDecodesResponse(t *testing.T) {
	t.Parallel()

	agent, fakeHost, stream, sent := wave28AgentWithResponse(t, Response{
		Status: ResponseStatusOK,
		Result: map[string]interface{}{"answer": "done", "count": 2},
	})

	result, err := agent.InvokeTool(context.Background(), "summarize", map[string]interface{}{
		"text":  "hello",
		"limit": 3,
	})
	require.NoError(t, err)
	assert.Equal(t, "done", result["answer"])
	assert.Equal(t, float64(2), result["count"])
	wave28AssertOpenedAndClosed(t, fakeHost, stream)

	req := wave28DecodeRequest(t, sent)
	assert.Equal(t, req.RequestID, stream.responseRequestID)
	assert.Equal(t, RequestToolInvoke, req.Type)
	assert.Equal(t, "wave28-token", req.SessionToken)
	assert.Equal(t, "summarize", req.Payload["toolName"])
	assert.Equal(t, map[string]interface{}{"text": "hello", "limit": float64(3)}, req.Payload["params"])
}

func TestWave28RemoteAgentInvokeToolPropagatesOpenRemoteAndEncodingErrors(t *testing.T) {
	t.Parallel()

	t.Run("open stream failure", func(t *testing.T) {
		t.Parallel()

		fakeHost := &wave28Host{err: errors.New("peer unreachable")}
		agent := NewRemoteAgent(RemoteAgentConfig{
			Name:         "wave28-remote",
			PeerID:       peer.ID("wave28-peer"),
			SessionToken: "wave28-token",
			Host:         fakeHost,
			Logger:       zap.NewNop().Sugar(),
		})

		result, err := agent.InvokeTool(context.Background(), "echo", nil)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "open stream to")
		assert.Contains(t, err.Error(), "peer unreachable")
	})

	t.Run("context canceled before stream", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		fakeHost := &wave28Host{}
		agent := NewRemoteAgent(RemoteAgentConfig{
			Name:         "wave28-remote",
			PeerID:       peer.ID("wave28-peer"),
			SessionToken: "wave28-token",
			Host:         fakeHost,
			Logger:       zap.NewNop().Sugar(),
		})

		result, err := agent.InvokeTool(ctx, "echo", nil)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "open stream to")
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("context canceled after stream write", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		sent := &bytes.Buffer{}
		stream := &wave28Stream{
			writer: sent,
			onWrite: func() {
				cancel()
			},
			responseTemplate: &Response{Status: ResponseStatusOK},
		}
		fakeHost := &wave28Host{stream: stream}
		agent := NewRemoteAgent(RemoteAgentConfig{
			Name:         "wave28-remote",
			PeerID:       peer.ID("wave28-peer"),
			SessionToken: "wave28-token",
			Host:         fakeHost,
			Logger:       zap.NewNop().Sugar(),
		})

		result, err := agent.InvokeTool(ctx, "echo", map[string]interface{}{"value": "after-open"})
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Contains(t, err.Error(), "request context")
		assert.True(t, stream.closed)

		req := wave28DecodeRequest(t, sent)
		assert.Equal(t, RequestToolInvoke, req.Type)
	})

	t.Run("remote error without message uses unknown", func(t *testing.T) {
		t.Parallel()

		agent, _, _, _ := wave28AgentWithResponse(t, Response{Status: ResponseStatusError})
		result, err := agent.InvokeTool(context.Background(), "missing", nil)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "remote tool missing error: unknown error")
	})

	t.Run("json encoding failure closes stream", func(t *testing.T) {
		t.Parallel()

		agent, _, stream, _ := wave28AgentWithResponse(t, Response{Status: ResponseStatusOK})
		result, err := agent.InvokeTool(context.Background(), "bad-json", map[string]interface{}{
			"unsupported": func() {},
		})
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "tool invoke bad-json on wave28-remote")
		assert.Contains(t, err.Error(), "send request")
		assert.True(t, stream.closed)
	})
}

func TestWave28RemoteAgentQueryCapabilitiesAndFetchAgentCard(t *testing.T) {
	t.Parallel()

	t.Run("query capabilities success", func(t *testing.T) {
		t.Parallel()

		agent, fakeHost, stream, sent := wave28AgentWithResponse(t, Response{
			Status: ResponseStatusOK,
			Result: map[string]interface{}{"capabilities": []interface{}{"search", "write"}},
		})

		result, err := agent.QueryCapabilities(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []interface{}{"search", "write"}, result["capabilities"])
		wave28AssertOpenedAndClosed(t, fakeHost, stream)

		req := wave28DecodeRequest(t, sent)
		assert.Equal(t, RequestCapabilityQuery, req.Type)
		assert.Equal(t, "wave28-token", req.SessionToken)
		assert.Nil(t, req.Payload)
	})

	t.Run("query capabilities remote error", func(t *testing.T) {
		t.Parallel()

		agent, _, _, _ := wave28AgentWithResponse(t, Response{
			Status: ResponseStatusDenied,
			Error:  "not allowed",
		})

		result, err := agent.QueryCapabilities(context.Background())
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "capability query error: not allowed")
	})

	t.Run("fetch agent card success", func(t *testing.T) {
		t.Parallel()

		agent, fakeHost, stream, sent := wave28AgentWithResponse(t, Response{
			Status: ResponseStatusOK,
			Result: map[string]interface{}{"name": "Remote Agent", "version": "1"},
		})

		result, err := agent.FetchAgentCard(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "Remote Agent", result["name"])
		wave28AssertOpenedAndClosed(t, fakeHost, stream)

		req := wave28DecodeRequest(t, sent)
		assert.Equal(t, RequestAgentCard, req.Type)
		assert.Nil(t, req.Payload)
	})

	t.Run("fetch agent card receive error", func(t *testing.T) {
		t.Parallel()

		sent := &bytes.Buffer{}
		stream := &wave28Stream{reader: bytes.NewBufferString("{not-json"), writer: sent}
		fakeHost := &wave28Host{stream: stream}
		agent := NewRemoteAgent(RemoteAgentConfig{
			Name:         "wave28-remote",
			PeerID:       peer.ID("wave28-peer"),
			SessionToken: "wave28-token",
			Host:         fakeHost,
			Logger:       zap.NewNop().Sugar(),
		})

		result, err := agent.FetchAgentCard(context.Background())
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "agent card fetch wave28-remote")
		assert.Contains(t, err.Error(), "receive response")
		assert.True(t, stream.closed)
	})
}

func TestWave28RemoteAgentQueryPriceSendsPayloadAndParsesQuote(t *testing.T) {
	t.Parallel()

	agent, fakeHost, stream, sent := wave28AgentWithResponse(t, Response{
		Status: ResponseStatusOK,
		Result: map[string]interface{}{
			"toolName":     "translate",
			"price":        "0.25",
			"currency":     "USDC",
			"usdcContract": "0xUSDC",
			"chainId":      8453,
			"sellerAddr":   "0xSeller",
			"quoteExpiry":  123456789,
			"isFree":       false,
		},
	})

	quote, err := agent.QueryPrice(context.Background(), "translate")
	require.NoError(t, err)
	assert.Equal(t, &PriceQuoteResult{
		ToolName:     "translate",
		Price:        "0.25",
		Currency:     "USDC",
		USDCContract: "0xUSDC",
		ChainID:      8453,
		SellerAddr:   "0xSeller",
		QuoteExpiry:  123456789,
		IsFree:       false,
	}, quote)
	wave28AssertOpenedAndClosed(t, fakeHost, stream)

	req := wave28DecodeRequest(t, sent)
	assert.Equal(t, RequestPriceQuery, req.Type)
	assert.Equal(t, "translate", req.Payload["toolName"])
}

func TestWave28RemoteAgentQueryPriceErrorBranches(t *testing.T) {
	t.Parallel()

	t.Run("remote error without message uses unknown", func(t *testing.T) {
		t.Parallel()

		agent, _, _, _ := wave28AgentWithResponse(t, Response{Status: ResponseStatusPaymentRequired})
		quote, err := agent.QueryPrice(context.Background(), "expensive")
		require.Error(t, err)
		assert.Nil(t, quote)
		assert.Contains(t, err.Error(), "price query expensive error: unknown error")
	})

	t.Run("invalid quote result returns unmarshal error", func(t *testing.T) {
		t.Parallel()

		agent, _, _, _ := wave28AgentWithResponse(t, Response{
			Status: ResponseStatusOK,
			Result: map[string]interface{}{"toolName": "bad-price", "chainId": "not-a-number"},
		})

		quote, err := agent.QueryPrice(context.Background(), "bad-price")
		require.Error(t, err)
		assert.Nil(t, quote)
		assert.Contains(t, err.Error(), "unmarshal price quote")
	})
}

func TestWave28RemoteAgentInvokeToolPaidReturnsResponseAndPayload(t *testing.T) {
	t.Parallel()

	agent, fakeHost, stream, sent := wave28AgentWithResponse(t, Response{
		Status: ResponseStatusPaymentRequired,
		Error:  "payment missing",
		Result: map[string]interface{}{"invoice": "inv-1"},
	})

	resp, err := agent.InvokeToolPaid(
		context.Background(),
		"render",
		map[string]interface{}{"prompt": "sun"},
		map[string]interface{}{"txHash": "0xabc"},
	)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, ResponseStatusPaymentRequired, resp.Status)
	assert.Equal(t, "payment missing", resp.Error)
	assert.Equal(t, "inv-1", resp.Result["invoice"])
	wave28AssertOpenedAndClosed(t, fakeHost, stream)

	req := wave28DecodeRequest(t, sent)
	assert.Equal(t, RequestToolInvokePaid, req.Type)
	assert.Equal(t, "render", req.Payload["toolName"])
	assert.Equal(t, map[string]interface{}{"prompt": "sun"}, req.Payload["params"])
	assert.Equal(t, map[string]interface{}{"txHash": "0xabc"}, req.Payload["paymentAuth"])
}

func TestWave28RemoteAgentTeamMessagesSendExpectedPayloads(t *testing.T) {
	t.Parallel()

	t.Run("invite", func(t *testing.T) {
		t.Parallel()

		agent, fakeHost, stream, sent := wave28AgentWithResponse(t, Response{
			Status: ResponseStatusOK,
			Result: map[string]interface{}{"accepted": true},
		})

		resp, err := agent.SendTeamInvite(context.Background(), TeamInvitePayload{
			TeamID:       "team-1",
			TeamName:     "Alpha",
			Goal:         "ship tests",
			LeaderDID:    "did:key:leader",
			Role:         "reviewer",
			Capabilities: []string{"go", "p2p"},
		})
		require.NoError(t, err)
		assert.Equal(t, true, resp.Result["accepted"])
		wave28AssertOpenedAndClosed(t, fakeHost, stream)

		req := wave28DecodeRequest(t, sent)
		assert.Equal(t, RequestTeamInvite, req.Type)
		assert.Equal(t, "team-1", req.Payload["teamId"])
		assert.Equal(t, "Alpha", req.Payload["teamName"])
		assert.Equal(t, "ship tests", req.Payload["goal"])
		assert.Equal(t, "did:key:leader", req.Payload["leaderDid"])
		assert.Equal(t, "reviewer", req.Payload["role"])
		assert.Equal(t, []interface{}{"go", "p2p"}, req.Payload["capabilities"])
	})

	t.Run("task with deadline", func(t *testing.T) {
		t.Parallel()

		deadline := time.Date(2026, 5, 19, 10, 30, 0, 123, time.UTC)
		agent, fakeHost, stream, sent := wave28AgentWithResponse(t, Response{
			Status: ResponseStatusOK,
			Result: map[string]interface{}{"queued": true},
		})

		resp, err := agent.SendTeamTask(context.Background(), TeamTaskPayload{
			TeamID:   "team-1",
			TaskID:   "task-9",
			ToolName: "search",
			Params:   map[string]interface{}{"query": "lango"},
			Deadline: deadline,
		})
		require.NoError(t, err)
		assert.Equal(t, true, resp.Result["queued"])
		wave28AssertOpenedAndClosed(t, fakeHost, stream)

		req := wave28DecodeRequest(t, sent)
		assert.Equal(t, RequestTeamTask, req.Type)
		assert.Equal(t, "team-1", req.Payload["teamId"])
		assert.Equal(t, "task-9", req.Payload["taskId"])
		assert.Equal(t, "search", req.Payload["toolName"])
		assert.Equal(t, map[string]interface{}{"query": "lango"}, req.Payload["params"])
		assert.Equal(t, deadline.Format(time.RFC3339Nano), req.Payload["deadline"])
	})

	t.Run("disband", func(t *testing.T) {
		t.Parallel()

		agent, fakeHost, stream, sent := wave28AgentWithResponse(t, Response{
			Status: ResponseStatusOK,
			Result: map[string]interface{}{"disbanded": true},
		})

		resp, err := agent.SendTeamDisband(context.Background(), TeamDisbandPayload{
			TeamID: "team-1",
			Reason: "complete",
		})
		require.NoError(t, err)
		assert.Equal(t, true, resp.Result["disbanded"])
		wave28AssertOpenedAndClosed(t, fakeHost, stream)

		req := wave28DecodeRequest(t, sent)
		assert.Equal(t, RequestTeamDisband, req.Type)
		assert.Equal(t, "team-1", req.Payload["teamId"])
		assert.Equal(t, "complete", req.Payload["reason"])
	})
}

func TestWave28RemoteAgentTeamMethodsPropagateSendRequestErrors(t *testing.T) {
	t.Parallel()

	sent := &bytes.Buffer{}
	stream := &wave28Stream{reader: bytes.NewBufferString("{bad-response"), writer: sent}
	fakeHost := &wave28Host{stream: stream}
	agent := NewRemoteAgent(RemoteAgentConfig{
		Name:         "wave28-remote",
		PeerID:       peer.ID("wave28-peer"),
		SessionToken: "wave28-token",
		Host:         fakeHost,
		Logger:       zap.NewNop().Sugar(),
	})

	resp, err := agent.SendTeamDisband(context.Background(), TeamDisbandPayload{TeamID: "team-err"})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "team disband to wave28-remote")
	assert.Contains(t, err.Error(), "receive response")
	assert.True(t, stream.closed)

	req := wave28DecodeRequest(t, sent)
	assert.Equal(t, RequestTeamDisband, req.Type)
	assert.Equal(t, "team-err", req.Payload["teamId"])
}
