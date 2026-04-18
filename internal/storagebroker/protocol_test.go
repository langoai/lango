package storagebroker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsBrokerMode(t *testing.T) {
	assert.False(t, IsBrokerMode())
}

func TestRequestResponseJSON(t *testing.T) {
	req := Request{ID: 1, Method: methodHealth}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded Request
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, req.ID, decoded.ID)
	assert.Equal(t, req.Method, decoded.Method)
}

func TestServerHealthAndShutdown(t *testing.T) {
	srv := NewServer()

	health, err := srv.dispatch(context.Background(), Request{Method: methodHealth})
	require.NoError(t, err)
	assert.Equal(t, HealthResult{Opened: false}, health)

	shutdown, err := srv.dispatch(context.Background(), Request{Method: methodShutdown})
	require.NoError(t, err)
	assert.Equal(t, ShutdownResult{ShuttingDown: true}, shutdown)
}

func TestServerDBStatusSummary_RequiresPath(t *testing.T) {
	srv := NewServer()
	_, err := srv.dispatch(context.Background(), Request{
		Method: methodDBStatus,
	})
	require.Error(t, err)
}

func TestServerSecurityStateRequiresOpenDB(t *testing.T) {
	srv := NewServer()
	_, err := srv.dispatch(context.Background(), Request{Method: methodLoadSecurityState})
	require.Error(t, err)
}

func TestServerRunRoundTrip(t *testing.T) {
	srv := NewServer()

	var in bytes.Buffer
	require.NoError(t, json.NewEncoder(&in).Encode(Request{ID: 1, Method: methodHealth}))
	require.NoError(t, json.NewEncoder(&in).Encode(Request{ID: 2, Method: methodShutdown}))

	var out bytes.Buffer
	require.NoError(t, srv.Run(&in, &out))

	dec := json.NewDecoder(&out)
	var resp1, resp2 Response
	require.NoError(t, dec.Decode(&resp1))
	require.NoError(t, dec.Decode(&resp2))

	assert.True(t, resp1.OK)
	assert.True(t, resp2.OK)
}

func TestServerOpenDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/broker.db"

	srv := NewServer()
	result, err := srv.openDB(context.Background(), OpenDBRequest{DBPath: dbPath})
	require.NoError(t, err)
	assert.True(t, result.Opened)

	_, statErr := os.Stat(dbPath)
	require.NoError(t, statErr)

	_ = srv.shutdown()
}

func TestServerEncryptDecryptPayloadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/broker.db"

	srv := NewServer()
	_, err := srv.openDB(context.Background(), OpenDBRequest{
		DBPath:         dbPath,
		PayloadKey:     bytes.Repeat([]byte{0x22}, 32),
		PayloadVersion: 1,
	})
	require.NoError(t, err)

	encAny, err := srv.dispatch(context.Background(), Request{
		Method:  methodEncryptPayload,
		Payload: mustPayload(t, EncryptPayloadRequest{Plaintext: []byte("secret payload")}),
	})
	require.NoError(t, err)
	enc := encAny.(EncryptPayloadResult)
	assert.Equal(t, 1, enc.KeyVersion)
	require.NotEmpty(t, enc.Ciphertext)
	require.NotEmpty(t, enc.Nonce)

	decAny, err := srv.dispatch(context.Background(), Request{
		Method: methodDecryptPayload,
		Payload: mustPayload(t, DecryptPayloadRequest{
			Ciphertext: enc.Ciphertext,
			Nonce:      enc.Nonce,
			KeyVersion: enc.KeyVersion,
		}),
	})
	require.NoError(t, err)
	dec := decAny.(DecryptPayloadResult)
	assert.Equal(t, []byte("secret payload"), dec.Plaintext)
}

func TestServerDecryptPayloadTamperFails(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/broker.db"

	srv := NewServer()
	_, err := srv.openDB(context.Background(), OpenDBRequest{
		DBPath:         dbPath,
		PayloadKey:     bytes.Repeat([]byte{0x33}, 32),
		PayloadVersion: 1,
	})
	require.NoError(t, err)

	encAny, err := srv.dispatch(context.Background(), Request{
		Method:  methodEncryptPayload,
		Payload: mustPayload(t, EncryptPayloadRequest{Plaintext: []byte("secret payload")}),
	})
	require.NoError(t, err)
	enc := encAny.(EncryptPayloadResult)
	enc.Ciphertext[0] ^= 0xFF

	_, err = srv.dispatch(context.Background(), Request{
		Method: methodDecryptPayload,
		Payload: mustPayload(t, DecryptPayloadRequest{
			Ciphertext: enc.Ciphertext,
			Nonce:      enc.Nonce,
			KeyVersion: enc.KeyVersion,
		}),
	})
	require.Error(t, err)
}

func mustPayload(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}

type overlapDetectWriter struct {
	target     io.WriteCloser
	active     int32
	overlapped atomic.Bool
}

func (w *overlapDetectWriter) Write(p []byte) (int, error) {
	if !atomic.CompareAndSwapInt32(&w.active, 0, 1) {
		w.overlapped.Store(true)
		return 0, errors.New("concurrent write detected")
	}
	defer atomic.StoreInt32(&w.active, 0)
	time.Sleep(10 * time.Millisecond)
	return w.target.Write(p)
}

func (w *overlapDetectWriter) Close() error {
	return w.target.Close()
}

func newPipeClient(t *testing.T, responder func(Request) Response) (*Client, func()) {
	t.Helper()
	reqR, reqW := io.Pipe()
	respR, respW := io.Pipe()

	c := &Client{
		stdin:   reqW,
		stdout:  respR,
		pending: make(map[uint64]chan Response),
	}
	go c.readLoop()
	go func() {
		dec := json.NewDecoder(reqR)
		enc := json.NewEncoder(respW)
		for {
			var req Request
			if err := dec.Decode(&req); err != nil {
				_ = respW.Close()
				return
			}
			_ = enc.Encode(responder(req))
		}
	}()
	return c, func() {
		_ = reqW.Close()
		_ = respW.Close()
	}
}

func TestClientCallSerializesWrites(t *testing.T) {
	reqR, reqW := io.Pipe()
	respR, respW := io.Pipe()
	writer := &overlapDetectWriter{target: reqW}
	c := &Client{
		stdin:   writer,
		stdout:  respR,
		pending: make(map[uint64]chan Response),
	}
	go c.readLoop()
	go func() {
		dec := json.NewDecoder(reqR)
		enc := json.NewEncoder(respW)
		for {
			var req Request
			if err := dec.Decode(&req); err != nil {
				_ = respW.Close()
				return
			}
			_ = enc.Encode(Response{ID: req.ID, OK: true})
		}
	}()
	defer func() {
		_ = writer.Close()
		_ = respW.Close()
	}()

	errCh := make(chan error, 2)
	go func() { errCh <- c.StoreSalt(context.Background(), []byte("salt")) }()
	go func() { errCh <- c.StoreChecksum(context.Background(), []byte("checksum")) }()

	require.NoError(t, <-errCh)
	require.NoError(t, <-errCh)
	assert.False(t, writer.overlapped.Load())
}

func TestClientReadLoopHandlesLargeResponses(t *testing.T) {
	largeCiphertext := bytes.Repeat([]byte("a"), 70*1024)
	c, cleanup := newPipeClient(t, func(req Request) Response {
		if req.Method != methodEncryptPayload {
			return Response{ID: req.ID, OK: false, Error: "unexpected method"}
		}
		result, err := json.Marshal(EncryptPayloadResult{
			Ciphertext: largeCiphertext,
			Nonce:      []byte("123456789012"),
			KeyVersion: 1,
		})
		require.NoError(t, err)
		return Response{ID: req.ID, OK: true, Result: result}
	})
	defer cleanup()

	resp, err := c.EncryptPayload(context.Background(), []byte("payload"))
	require.NoError(t, err)
	assert.Len(t, resp.Ciphertext, len(largeCiphertext))
}

func TestClientCloseSendsShutdownRPC(t *testing.T) {
	reqR, reqW := io.Pipe()
	respR, respW := io.Pipe()
	cmd := exec.Command("true")
	require.NoError(t, cmd.Start())

	c := &Client{
		cmd:     cmd,
		stdin:   reqW,
		stdout:  respR,
		pending: make(map[uint64]chan Response),
	}
	go c.readLoop()

	seen := make(chan string, 1)
	go func() {
		dec := json.NewDecoder(reqR)
		enc := json.NewEncoder(respW)
		for {
			var req Request
			if err := dec.Decode(&req); err != nil {
				return
			}
			seen <- req.Method
			_ = enc.Encode(Response{ID: req.ID, OK: true, Result: mustPayload(t, ShutdownResult{ShuttingDown: true})})
			return
		}
	}()

	require.NoError(t, c.Close(context.Background()))
	assert.Equal(t, methodShutdown, <-seen)
}
