package gitbundle

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStreamHandler_ValidSessionDispatchesListCommits(t *testing.T) {
	service := &streamHandlerValidSessionDispatchesListCommitsProtocolService{
		commits: []CommitInfo{{
			Hash:      "abc123",
			Message:   "initial",
			Author:    "agent",
			Timestamp: time.Unix(1_893_456_000, 0).UTC(),
		}},
	}
	req := Request{
		Type:        RequestListCommits,
		WorkspaceID: "ws-1",
		Token:       "session-token",
	}
	stream := newProtocolTestStream(encodeProtocolRequest(t, req))

	var validatedToken string
	handler := newStreamHandlerValidSessionDispatchesListCommitsProtocolHandler(service)
	handler.validator = func(token string) (string, bool) {
		validatedToken = token
		return "did:lango:peer-1", true
	}

	handler.StreamHandler()(stream)

	resp := decodeProtocolResponse(t, &stream.out)
	assert.True(t, stream.closed)
	assert.Equal(t, "session-token", validatedToken)
	assert.Equal(t, StatusOK, resp.Status)
	assert.Empty(t, resp.Error)

	data := decodeProtocolData[ListCommitsResponse](t, resp)
	require.Len(t, data.Commits, 1)
	assert.Equal(t, "abc123", data.Commits[0].Hash)
	assert.Equal(t, []string{"ws-1:20"}, service.listCalls)
}

func TestStreamHandler_RequestBodyLimitWritesDecodeError(t *testing.T) {
	req := `{"type":"list_commits","workspaceId":"ws-1","payload":"` +
		strings.Repeat("x", 6000) + `"}`
	stream := newProtocolTestStream([]byte(req))
	handler := NewHandler(HandlerConfig{
		MaxBundleSize: 1,
		Logger:        zap.NewNop(),
	})

	handler.StreamHandler()(stream)

	resp := decodeProtocolResponse(t, &stream.out)
	assert.True(t, stream.closed)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "decode request")
	assert.Empty(t, resp.Data)
}

func TestStreamHandler_ValidSessionDispatchesServiceError(t *testing.T) {
	service := &streamHandlerValidSessionDispatchesListCommitsProtocolService{listErr: errors.New("store offline")}
	req := Request{
		Type:        RequestListCommits,
		WorkspaceID: "ws-error",
		Token:       "session-token",
	}
	stream := newProtocolTestStream(encodeProtocolRequest(t, req))

	var validatedToken string
	handler := newStreamHandlerValidSessionDispatchesListCommitsProtocolHandler(service)
	handler.validator = func(token string) (string, bool) {
		validatedToken = token
		return "did:lango:peer-1", true
	}

	handler.StreamHandler()(stream)

	resp := decodeProtocolResponse(t, &stream.out)
	assert.True(t, stream.closed)
	assert.Equal(t, "session-token", validatedToken)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "list commits: store offline")
	assert.Empty(t, resp.Data)
	assert.Equal(t, []string{"ws-error:20"}, service.listCalls)
}

func TestHandlePushBundle_InvalidPayloadWritesValidationError(t *testing.T) {
	stream := newProtocolTestStream(nil)
	handler := newStreamHandlerValidSessionDispatchesListCommitsProtocolHandler(nil)

	handler.handlePushBundle(context.Background(), stream, Request{
		Payload: json.RawMessage(`{"bundle":`),
	})

	resp := decodeProtocolResponse(t, &stream.out)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "unmarshal push payload")
	assert.Empty(t, resp.Data)
}

func TestHandlePushIncremental_RejectsOversizedBundle(t *testing.T) {
	payload := encodeProtocolPayload(t, PushIncrementalBundlePayload{
		Bundle: []byte("large"),
	})
	stream := newProtocolTestStream(nil)
	handler := NewHandler(HandlerConfig{
		MaxBundleSize: 4,
		Logger:        zap.NewNop(),
	})

	handler.handlePushIncrementalBundle(context.Background(), stream, Request{Payload: payload})

	resp := decodeProtocolResponse(t, &stream.out)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "bundle too large: 5 > 4")
	assert.Empty(t, resp.Data)
}

func TestHandlePushBundle_ServiceErrorWritesApplyError(t *testing.T) {
	service := &streamHandlerValidSessionDispatchesListCommitsProtocolService{applyErr: errors.New("apply denied")}
	payload := encodeProtocolPayload(t, PushBundlePayload{
		Bundle: []byte("bundle-bytes"),
	})
	stream := newProtocolTestStream(nil)
	handler := newStreamHandlerValidSessionDispatchesListCommitsProtocolHandler(service)

	handler.handlePushBundle(context.Background(), stream, Request{
		WorkspaceID: "ws-apply",
		Payload:     payload,
	})

	resp := decodeProtocolResponse(t, &stream.out)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "apply bundle: apply denied")
	assert.Empty(t, resp.Data)
	assert.Equal(t, []string{"ws-apply:12"}, service.applyCalls)
}

func TestHandleListCommits_ServiceErrorForMissingWorkspace(t *testing.T) {
	service := &streamHandlerValidSessionDispatchesListCommitsProtocolService{listErr: errors.New("workspace missing")}
	stream := newProtocolTestStream(nil)
	handler := newStreamHandlerValidSessionDispatchesListCommitsProtocolHandler(service)

	handler.handleListCommits(context.Background(), stream, Request{WorkspaceID: "missing"})

	resp := decodeProtocolResponse(t, &stream.out)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "list commits: workspace missing")
	assert.Empty(t, resp.Data)
	assert.Equal(t, []string{"missing:20"}, service.listCalls)
}

func TestHandleFetchIncremental_InvalidBaseCommitWritesServiceError(t *testing.T) {
	service := &streamHandlerValidSessionDispatchesListCommitsProtocolService{incrementalErr: errors.New("invalid base commit")}
	payload := encodeProtocolPayload(t, FetchIncrementalPayload{
		BaseCommit: "not-a-hash",
	})
	stream := newProtocolTestStream(nil)
	handler := newStreamHandlerValidSessionDispatchesListCommitsProtocolHandler(service)

	handler.handleFetchIncremental(context.Background(), stream, Request{
		WorkspaceID: "ws-1",
		Payload:     payload,
	})

	resp := decodeProtocolResponse(t, &stream.out)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "create incremental bundle: invalid base commit")
	assert.Empty(t, resp.Data)
	assert.Equal(t, []string{"ws-1:not-a-hash"}, service.incrementalCalls)
}

func TestHandleHasCommit_InvalidHashWritesServiceError(t *testing.T) {
	service := &streamHandlerValidSessionDispatchesListCommitsProtocolService{hasCommitErr: errors.New("invalid commit hash")}
	payload := encodeProtocolPayload(t, HasCommitPayload{
		CommitHash: "bad-hash",
	})
	stream := newProtocolTestStream(nil)
	handler := newStreamHandlerValidSessionDispatchesListCommitsProtocolHandler(service)

	handler.handleHasCommit(context.Background(), stream, Request{
		WorkspaceID: "ws-1",
		Payload:     payload,
	})

	resp := decodeProtocolResponse(t, &stream.out)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "has commit: invalid commit hash")
	assert.Empty(t, resp.Data)
	assert.Equal(t, []string{"ws-1:bad-hash"}, service.hasCommitCalls)
}

func TestWriteResponse_NilDataWritesOKWithoutData(t *testing.T) {
	stream := newProtocolTestStream(nil)
	handler := newStreamHandlerValidSessionDispatchesListCommitsProtocolHandler(nil)

	handler.writeResponse(stream, nil)

	resp := decodeProtocolResponse(t, &stream.out)
	assert.Equal(t, StatusOK, resp.Status)
	assert.Empty(t, resp.Error)
	assert.Empty(t, resp.Data)
}

func TestWriteResponse_MarshalFailureWritesError(t *testing.T) {
	stream := newProtocolTestStream(nil)
	handler := newStreamHandlerValidSessionDispatchesListCommitsProtocolHandler(nil)

	handler.writeResponse(stream, map[string]any{"bad": func() {}})

	resp := decodeProtocolResponse(t, &stream.out)
	assert.Equal(t, StatusError, resp.Status)
	assert.Contains(t, resp.Error, "marshal response")
	assert.Empty(t, resp.Data)
}

type streamHandlerValidSessionDispatchesListCommitsProtocolService struct {
	commits          []CommitInfo
	applyErr         error
	listErr          error
	incrementalErr   error
	hasCommitErr     error
	applyCalls       []string
	listCalls        []string
	incrementalCalls []string
	hasCommitCalls   []string
}

func newStreamHandlerValidSessionDispatchesListCommitsProtocolHandler(service protocolService) *Handler {
	return &Handler{
		service: service,
		validator: func(token string) (string, bool) {
			return token, token != ""
		},
		maxBundle: 50 * 1024 * 1024,
		logger:    zap.NewNop(),
	}
}

func (s *streamHandlerValidSessionDispatchesListCommitsProtocolService) ApplyBundle(_ context.Context, workspaceID string, bundle []byte) error {
	s.applyCalls = append(s.applyCalls, workspaceID+":"+strconv.Itoa(len(bundle)))
	return s.applyErr
}

func (s *streamHandlerValidSessionDispatchesListCommitsProtocolService) CreateBundle(context.Context, string) ([]byte, string, error) {
	return nil, "", errors.New("unexpected CreateBundle")
}

func (s *streamHandlerValidSessionDispatchesListCommitsProtocolService) Log(_ context.Context, workspaceID string, limit int) ([]CommitInfo, error) {
	s.listCalls = append(s.listCalls, workspaceID+":"+strconv.Itoa(limit))
	if s.listErr != nil {
		return nil, s.listErr
	}
	return append([]CommitInfo(nil), s.commits...), nil
}

func (s *streamHandlerValidSessionDispatchesListCommitsProtocolService) Leaves(context.Context, string) ([]string, error) {
	return nil, errors.New("unexpected Leaves")
}

func (s *streamHandlerValidSessionDispatchesListCommitsProtocolService) Diff(context.Context, string, string, string) (string, error) {
	return "", errors.New("unexpected Diff")
}

func (s *streamHandlerValidSessionDispatchesListCommitsProtocolService) SafeApplyBundle(context.Context, string, []byte) error {
	return errors.New("unexpected SafeApplyBundle")
}

func (s *streamHandlerValidSessionDispatchesListCommitsProtocolService) CreateIncrementalBundle(_ context.Context, workspaceID, baseCommit string) ([]byte, string, error) {
	s.incrementalCalls = append(s.incrementalCalls, workspaceID+":"+baseCommit)
	return nil, "", s.incrementalErr
}

func (s *streamHandlerValidSessionDispatchesListCommitsProtocolService) VerifyBundle(context.Context, string, []byte) error {
	return errors.New("unexpected VerifyBundle")
}

func (s *streamHandlerValidSessionDispatchesListCommitsProtocolService) HasCommit(_ context.Context, workspaceID, commitHash string) (bool, error) {
	s.hasCommitCalls = append(s.hasCommitCalls, workspaceID+":"+commitHash)
	return false, s.hasCommitErr
}
