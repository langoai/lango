package gitbundle

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type protocolTestStream struct {
	in     *bytes.Reader
	out    bytes.Buffer
	closed bool
}

func newProtocolTestStream(input []byte) *protocolTestStream {
	return &protocolTestStream{in: bytes.NewReader(input)}
}

func (s *protocolTestStream) Read(p []byte) (int, error) {
	return s.in.Read(p)
}

func (s *protocolTestStream) Write(p []byte) (int, error) {
	return s.out.Write(p)
}

func (s *protocolTestStream) Close() error {
	s.closed = true
	return nil
}

func (s *protocolTestStream) CloseRead() error {
	return nil
}

func (s *protocolTestStream) CloseWrite() error {
	return nil
}

func (s *protocolTestStream) Reset() error {
	return nil
}

func (s *protocolTestStream) ResetWithError(network.StreamErrorCode) error {
	return nil
}

func (s *protocolTestStream) SetDeadline(time.Time) error {
	return nil
}

func (s *protocolTestStream) SetReadDeadline(time.Time) error {
	return nil
}

func (s *protocolTestStream) SetWriteDeadline(time.Time) error {
	return nil
}

func (s *protocolTestStream) ID() string {
	return "test-stream"
}

func (s *protocolTestStream) Protocol() protocol.ID {
	return ProtocolID
}

func (s *protocolTestStream) SetProtocol(protocol.ID) error {
	return nil
}

func (s *protocolTestStream) Stat() network.Stats {
	return network.Stats{}
}

func (s *protocolTestStream) Conn() network.Conn {
	return nil
}

func (s *protocolTestStream) Scope() network.StreamScope {
	return nil
}

func encodeProtocolRequest(t *testing.T, req Request) []byte {
	t.Helper()

	b, err := json.Marshal(req)
	require.NoError(t, err)
	return b
}

func encodeProtocolPayload(t *testing.T, payload any) json.RawMessage {
	t.Helper()

	b, err := json.Marshal(payload)
	require.NoError(t, err)
	return b
}

func decodeProtocolResponse(t *testing.T, r io.Reader) Response {
	t.Helper()

	var resp Response
	require.NoError(t, json.NewDecoder(r).Decode(&resp))
	return resp
}

func decodeProtocolData[T any](t *testing.T, resp Response) T {
	t.Helper()

	var data T
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	return data
}

func newProtocolTestHandler(service *Service) *Handler {
	return NewHandler(HandlerConfig{
		Service: service,
		Logger:  zap.NewNop(),
	})
}

func TestHandler_StreamHandler_InvalidJSON(t *testing.T) {
	stream := newProtocolTestStream([]byte(`{"type":`))
	handler := newProtocolTestHandler(nil)

	handler.StreamHandler()(stream)

	resp := decodeProtocolResponse(t, &stream.out)
	assert.True(t, stream.closed)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "decode request")
	assert.Empty(t, resp.Data)
}

func TestHandler_StreamHandler_InvalidToken(t *testing.T) {
	req := Request{
		Type:        RequestListCommits,
		WorkspaceID: "ws-1",
		Token:       "bad-token",
	}
	stream := newProtocolTestStream(encodeProtocolRequest(t, req))
	handler := NewHandler(HandlerConfig{
		Validator: func(token string) (string, bool) {
			assert.Equal(t, "bad-token", token)
			return "", false
		},
		Logger: zap.NewNop(),
	})

	handler.StreamHandler()(stream)

	resp := decodeProtocolResponse(t, &stream.out)
	assert.True(t, stream.closed)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "invalid or expired session token")
	assert.Empty(t, resp.Data)
}

func TestHandler_StreamHandler_UnknownType(t *testing.T) {
	req := Request{
		Type:        RequestType("unsupported"),
		WorkspaceID: "ws-1",
	}
	stream := newProtocolTestStream(encodeProtocolRequest(t, req))
	handler := newProtocolTestHandler(nil)

	handler.StreamHandler()(stream)

	resp := decodeProtocolResponse(t, &stream.out)
	assert.True(t, stream.closed)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "unknown request type: unsupported")
	assert.Empty(t, resp.Data)
}

func TestHandler_HandlePushBundle_RejectsOversizedBundle(t *testing.T) {
	payload := encodeProtocolPayload(t, PushBundlePayload{
		Bundle: []byte("too-large"),
	})
	stream := newProtocolTestStream(nil)
	handler := NewHandler(HandlerConfig{
		MaxBundleSize: 3,
		Logger:        zap.NewNop(),
	})

	handler.handlePushBundle(context.Background(), stream, Request{Payload: payload})

	resp := decodeProtocolResponse(t, &stream.out)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "bundle too large: 9 > 3")
	assert.Empty(t, resp.Data)
}

func TestHandler_HandleListCommits_DefaultLimit_EmptyRepo(t *testing.T) {
	service := newTestService(t)
	require.NoError(t, service.Init(context.Background(), "ws-1"))
	stream := newProtocolTestStream(nil)
	handler := newProtocolTestHandler(service)

	handler.handleListCommits(context.Background(), stream, Request{WorkspaceID: "ws-1"})

	resp := decodeProtocolResponse(t, &stream.out)
	assert.Equal(t, StatusOK, resp.Status)
	assert.Empty(t, resp.Error)

	data := decodeProtocolData[ListCommitsResponse](t, resp)
	assert.Empty(t, data.Commits)
}

func TestHandler_HandleFetchByHash_EmptyRepo(t *testing.T) {
	skipIfNoGit(t)

	service := newTestService(t)
	require.NoError(t, service.Init(context.Background(), "ws-1"))
	stream := newProtocolTestStream(nil)
	handler := newProtocolTestHandler(service)

	handler.handleFetchByHash(context.Background(), stream, Request{WorkspaceID: "ws-1"})

	resp := decodeProtocolResponse(t, &stream.out)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "empty repository")
	assert.Empty(t, resp.Data)
}

func TestHandler_HandleDiff_BadPayload(t *testing.T) {
	stream := newProtocolTestStream(nil)
	handler := newProtocolTestHandler(nil)

	handler.handleDiff(context.Background(), stream, Request{
		Payload: json.RawMessage(`{"from":`),
	})

	resp := decodeProtocolResponse(t, &stream.out)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "unmarshal diff payload")
	assert.Empty(t, resp.Data)
}

func TestHandler_HandleVerifyBundle_InvalidBundle(t *testing.T) {
	skipIfNoGit(t)

	service := newTestService(t)
	require.NoError(t, service.Init(context.Background(), "ws-1"))
	payload := encodeProtocolPayload(t, VerifyBundlePayload{
		Bundle: []byte("not-a-valid-bundle"),
	})
	stream := newProtocolTestStream(nil)
	handler := newProtocolTestHandler(service)

	handler.handleVerifyBundle(context.Background(), stream, Request{
		WorkspaceID: "ws-1",
		Payload:     payload,
	})

	resp := decodeProtocolResponse(t, &stream.out)
	assert.Equal(t, StatusOK, resp.Status)
	assert.Empty(t, resp.Error)

	data := decodeProtocolData[VerifyBundleResponse](t, resp)
	assert.False(t, data.Valid)
	assert.NotEmpty(t, data.Message)
}

func TestHandler_HandleHasCommit_InvalidPayload(t *testing.T) {
	stream := newProtocolTestStream(nil)
	handler := newProtocolTestHandler(nil)

	handler.handleHasCommit(context.Background(), stream, Request{
		Payload: json.RawMessage(`{"commitHash":`),
	})

	resp := decodeProtocolResponse(t, &stream.out)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "unmarshal has_commit payload")
	assert.Empty(t, resp.Data)
}

func TestHandler_HandleHasCommit_EmptyRepo(t *testing.T) {
	service := newTestService(t)
	require.NoError(t, service.Init(context.Background(), "ws-1"))
	hash := strings.Repeat("a", 40)
	payload := encodeProtocolPayload(t, HasCommitPayload{CommitHash: hash})
	stream := newProtocolTestStream(nil)
	handler := newProtocolTestHandler(service)

	handler.handleHasCommit(context.Background(), stream, Request{
		WorkspaceID: "ws-1",
		Payload:     payload,
	})

	resp := decodeProtocolResponse(t, &stream.out)
	assert.Equal(t, StatusOK, resp.Status)
	assert.Empty(t, resp.Error)

	data := decodeProtocolData[HasCommitResponse](t, resp)
	assert.False(t, data.Exists)
	assert.Equal(t, hash, data.Hash)
}
