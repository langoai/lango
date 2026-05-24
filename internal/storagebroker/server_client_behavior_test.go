package storagebroker

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
	"github.com/stretchr/testify/require"
)

func TestServerHandleReturnsErrorsForUnknownMethodAndBadPayload(t *testing.T) {
	t.Parallel()

	srv := NewServer()

	unknown := srv.handle(Request{ID: 11, Method: "missing_method"})
	require.False(t, unknown.OK)
	require.Equal(t, uint64(11), unknown.ID)
	require.Contains(t, unknown.Error, `unknown broker method "missing_method"`)

	badPayload := srv.handle(Request{
		ID:      12,
		Method:  methodOpenDB,
		Payload: []byte(`{"db_path":`),
	})
	require.False(t, badPayload.OK)
	require.Equal(t, uint64(12), badPayload.ID)
	require.Contains(t, badPayload.Error, "decode broker payload")
}

func TestServerSessionCRUDThroughDispatch(t *testing.T) {
	t.Parallel()

	srv := NewServer()
	t.Cleanup(func() {
		_ = srv.shutdown()
	})
	_, err := srv.dispatch(context.Background(), Request{
		Method:  methodOpenDB,
		Payload: mustPayload(t, OpenDBRequest{DBPath: t.TempDir() + "/broker.db"}),
	})
	require.NoError(t, err)

	createdAt := time.Unix(100, 0)
	createReq := SessionCreateRequest{Session: session.Session{
		Key:   "session-1",
		Model: "gpt-test",
		History: []session.Message{{
			Role:      types.RoleUser,
			Content:   "hello",
			Timestamp: createdAt,
		}},
	}}
	_, err = srv.dispatch(context.Background(), Request{
		Method:  methodSessionCreate,
		Payload: mustPayload(t, createReq),
	})
	require.NoError(t, err)

	getAny, err := srv.dispatch(context.Background(), Request{
		Method:  methodSessionGet,
		Payload: mustPayload(t, SessionGetRequest{Key: "session-1"}),
	})
	require.NoError(t, err)
	got := getAny.(SessionGetResult).Session
	require.Equal(t, "gpt-test", got.Model)
	require.Len(t, got.History, 1)
	require.Equal(t, "hello", got.History[0].Content)

	_, err = srv.dispatch(context.Background(), Request{
		Method: methodSessionAppend,
		Payload: mustPayload(t, SessionAppendMessageRequest{
			Key: "session-1",
			Message: session.Message{
				Role:      types.RoleAssistant,
				Content:   "world",
				Timestamp: createdAt.Add(time.Second),
			},
		}),
	})
	require.NoError(t, err)

	listAny, err := srv.dispatch(context.Background(), Request{Method: methodSessionList})
	require.NoError(t, err)
	list := listAny.(SessionListResult)
	require.Len(t, list.Sessions, 1)
	require.Equal(t, "session-1", list.Sessions[0].Key)

	_, err = srv.dispatch(context.Background(), Request{
		Method:  methodSessionDelete,
		Payload: mustPayload(t, SessionDeleteRequest{Key: "session-1"}),
	})
	require.NoError(t, err)

	_, err = srv.dispatch(context.Background(), Request{
		Method:  methodSessionGet,
		Payload: mustPayload(t, SessionGetRequest{Key: "session-1"}),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "session not found")
}

func TestServerPayloadProtectionRequiresInitializedKeyAndVersionMatches(t *testing.T) {
	t.Parallel()

	srv := NewServer()
	_, err := srv.dispatch(context.Background(), Request{
		Method:  methodEncryptPayload,
		Payload: mustPayload(t, EncryptPayloadRequest{Plaintext: []byte("plain")}),
	})
	require.ErrorContains(t, err, "payload protection key not initialized")

	_, err = srv.dispatch(context.Background(), Request{
		Method: methodOpenDB,
		Payload: mustPayload(t, OpenDBRequest{
			DBPath:         t.TempDir() + "/broker.db",
			PayloadKey:     bytes.Repeat([]byte{0x44}, 32),
			PayloadVersion: 7,
		}),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = srv.shutdown()
	})

	encAny, err := srv.dispatch(context.Background(), Request{
		Method:  methodEncryptPayload,
		Payload: mustPayload(t, EncryptPayloadRequest{Plaintext: []byte("plain")}),
	})
	require.NoError(t, err)
	enc := encAny.(EncryptPayloadResult)
	require.Equal(t, 7, enc.KeyVersion)

	_, err = srv.dispatch(context.Background(), Request{
		Method: methodDecryptPayload,
		Payload: mustPayload(t, DecryptPayloadRequest{
			Ciphertext: enc.Ciphertext,
			Nonce:      enc.Nonce,
			KeyVersion: 8,
		}),
	})
	require.ErrorContains(t, err, "unsupported payload key version 8")
}

func TestClientCallSurfacesBrokerAndDecodeErrors(t *testing.T) {
	t.Parallel()

	brokerErrClient, cleanup := newPipeClient(t, func(req Request) Response {
		return Response{ID: req.ID, OK: false, Error: "backend unavailable"}
	})
	defer cleanup()

	_, err := brokerErrClient.Health(context.Background())
	require.ErrorContains(t, err, "backend unavailable")

	decodeErrClient, cleanup := newPipeClient(t, func(req Request) Response {
		return Response{ID: req.ID, OK: true, Result: []byte(`"not an object"`)}
	})
	defer cleanup()

	_, err = decodeErrClient.Health(context.Background())
	require.ErrorContains(t, err, "decode broker result")
}

func TestClientCallCanceledContextReleasesPendingRequest(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	stdin := &bufferWriteCloser{}
	c := &Client{
		stdin:   stdin,
		stdout:  io.NopCloser(strings.NewReader("")),
		pending: make(map[uint64]chan Response),
	}

	err := c.call(ctx, methodHealth, nil, &HealthResult{})
	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, c.pending)
	require.Contains(t, stdin.String(), `"method":"health"`)
}

type bufferWriteCloser struct {
	bytes.Buffer
}

func (w *bufferWriteCloser) Close() error {
	return nil
}
