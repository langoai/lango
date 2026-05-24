package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/turnrunner"
	"github.com/langoai/lango/internal/turntrace"
)

type signResponseCompletesPendingProviderRequestCryptoRequest struct {
	event   string
	payload interface{}
}

func TestSignResponseCompletesPendingProviderRequest(t *testing.T) {
	t.Parallel()

	provider, requests := newSignResponseCompletesPendingProviderRequestCapturedProvider()
	server := New(Config{}, nil, provider, nil, nil)
	installSignResponseCompletesPendingProviderRequestCaptureSender(provider, requests)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	type signResult struct {
		signature []byte
		err       error
	}
	done := make(chan signResult, 1)
	go func() {
		signature, err := provider.Sign(ctx, "key-sign", []byte("payload"))
		done <- signResult{signature: signature, err: err}
	}()

	captured := requireSignResponseCompletesPendingProviderRequestCryptoRequest(t, requests)
	assert.Equal(t, "sign.request", captured.event)
	req, ok := captured.payload.(security.SignRequest)
	require.True(t, ok)
	assert.Equal(t, "key-sign", req.KeyID)
	assert.Equal(t, []byte("payload"), req.Payload)

	params := mustSignResponseCompletesPendingProviderRequestJSON(t, security.SignResponse{
		ID:        req.ID,
		Signature: []byte("signed"),
	})
	result, err := server.handleSignResponse(nil, params)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"status": "ok"}, result)

	select {
	case got := <-done:
		require.NoError(t, got.err)
		assert.Equal(t, []byte("signed"), got.signature)
	case <-time.After(time.Second):
		t.Fatal("sign request did not receive handler response")
	}
}

func TestEncryptResponseCompletesPendingProviderRequest(t *testing.T) {
	t.Parallel()

	provider, requests := newSignResponseCompletesPendingProviderRequestCapturedProvider()
	server := New(Config{}, nil, provider, nil, nil)
	installSignResponseCompletesPendingProviderRequestCaptureSender(provider, requests)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	type encryptResult struct {
		ciphertext []byte
		err        error
	}
	done := make(chan encryptResult, 1)
	go func() {
		ciphertext, err := provider.Encrypt(ctx, "key-encrypt", []byte("plain"))
		done <- encryptResult{ciphertext: ciphertext, err: err}
	}()

	captured := requireSignResponseCompletesPendingProviderRequestCryptoRequest(t, requests)
	assert.Equal(t, "encrypt.request", captured.event)
	req, ok := captured.payload.(security.EncryptRequest)
	require.True(t, ok)
	assert.Equal(t, "key-encrypt", req.KeyID)
	assert.Equal(t, []byte("plain"), req.Plaintext)

	params := mustSignResponseCompletesPendingProviderRequestJSON(t, security.EncryptResponse{
		ID:         req.ID,
		Ciphertext: []byte("cipher"),
	})
	result, err := server.handleEncryptResponse(nil, params)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"status": "ok"}, result)

	select {
	case got := <-done:
		require.NoError(t, got.err)
		assert.Equal(t, []byte("cipher"), got.ciphertext)
	case <-time.After(time.Second):
		t.Fatal("encrypt request did not receive handler response")
	}
}

func TestDecryptResponseCompletesPendingProviderRequest(t *testing.T) {
	t.Parallel()

	provider, requests := newSignResponseCompletesPendingProviderRequestCapturedProvider()
	server := New(Config{}, nil, provider, nil, nil)
	installSignResponseCompletesPendingProviderRequestCaptureSender(provider, requests)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	type decryptResult struct {
		plaintext []byte
		err       error
	}
	done := make(chan decryptResult, 1)
	go func() {
		plaintext, err := provider.Decrypt(ctx, "key-decrypt", []byte("cipher"))
		done <- decryptResult{plaintext: plaintext, err: err}
	}()

	captured := requireSignResponseCompletesPendingProviderRequestCryptoRequest(t, requests)
	assert.Equal(t, "decrypt.request", captured.event)
	req, ok := captured.payload.(security.DecryptRequest)
	require.True(t, ok)
	assert.Equal(t, "key-decrypt", req.KeyID)
	assert.Equal(t, []byte("cipher"), req.Ciphertext)

	params := mustSignResponseCompletesPendingProviderRequestJSON(t, security.DecryptResponse{
		ID:        req.ID,
		Plaintext: []byte("plain"),
	})
	result, err := server.handleDecryptResponse(nil, params)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"status": "ok"}, result)

	select {
	case got := <-done:
		require.NoError(t, got.err)
		assert.Equal(t, []byte("plain"), got.plaintext)
	case <-time.After(time.Second):
		t.Fatal("decrypt request did not receive handler response")
	}
}

func TestResponseHandlersRejectMissingProviderAndInvalidParams(t *testing.T) {
	t.Parallel()

	noProvider := New(Config{}, nil, nil, nil, nil)
	for name, handler := range map[string]RPCHandler{
		"sign":    noProvider.handleSignResponse,
		"encrypt": noProvider.handleEncryptResponse,
		"decrypt": noProvider.handleDecryptResponse,
	} {
		t.Run(name+"_missing_provider", func(t *testing.T) {
			_, err := handler(nil, json.RawMessage(`{"id":"missing"}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "provider not configured")
		})
	}

	withProvider := New(Config{}, nil, security.NewRPCProvider(), nil, nil)
	for name, handler := range map[string]RPCHandler{
		"sign":    withProvider.handleSignResponse,
		"encrypt": withProvider.handleEncryptResponse,
		"decrypt": withProvider.handleDecryptResponse,
	} {
		t.Run(name+"_invalid_params", func(t *testing.T) {
			_, err := handler(nil, json.RawMessage(`{`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid params")
		})
	}
}

func TestRouterSetupServesConfiguredRoutesWithoutListener(t *testing.T) {
	t.Parallel()

	server := New(Config{HTTPEnabled: true, WebSocketEnabled: true}, nil, security.NewRPCProvider(), nil, nil)
	router := server.Router()
	require.Same(t, server.router, router)

	router.Get("/configCommandLocalValidationDoesNotBootstrap8/probe", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	var health map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &health))
	assert.Equal(t, "ok", health["status"])

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/status", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	var status map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &status))
	assert.Equal(t, "running", status["status"])
	assert.Equal(t, float64(0), status["clients"])

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/configCommandLocalValidationDoesNotBootstrap8/probe", nil))
	assert.Equal(t, http.StatusNoContent, rec.Code)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/companion", nil))
	assert.NotEqual(t, http.StatusNotFound, rec.Code)

	noProvider := New(Config{HTTPEnabled: true, WebSocketEnabled: true}, nil, nil, nil, nil)
	rec = httptest.NewRecorder()
	noProvider.Router().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/companion", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestOnTurnCompleteDelegatesToConfiguredRunner(t *testing.T) {
	t.Parallel()

	runner := turnrunner.New(
		turnrunner.Config{HardCeiling: time.Second},
		signResponseCompletesPendingProviderRequestExecutor{},
		nil,
		nil,
	)
	server := New(Config{}, nil, nil, nil, nil)
	server.SetTurnRunner(runner)

	called := make(chan string, 1)
	server.OnTurnComplete(func(sessionKey string) {
		called <- sessionKey
	})
	require.Empty(t, server.turnCallbacks)

	result, err := runner.Run(context.Background(), turnrunner.Request{
		SessionKey: "sess-configCommandLocalValidationDoesNotBootstrap8",
		Input:      "hello",
	})
	require.NoError(t, err)
	assert.Equal(t, turntrace.OutcomeSuccess, result.Outcome)

	select {
	case got := <-called:
		assert.Equal(t, "sess-configCommandLocalValidationDoesNotBootstrap8", got)
	case <-time.After(time.Second):
		t.Fatal("turn completion callback was not fired by runner")
	}
}

func TestStartBackgroundReturnsListenErrorWithoutStartingGoroutine(t *testing.T) {
	t.Parallel()

	server := New(Config{Host: "bad\x00host", Port: 18789}, nil, nil, nil, nil)
	var wg sync.WaitGroup
	onErrorCalled := make(chan error, 1)

	err := server.StartBackground(&wg, func(err error) {
		onErrorCalled <- err
	})
	require.Error(t, err)

	wg.Wait()
	select {
	case err := <-onErrorCalled:
		t.Fatalf("onError should not be called when listen fails before goroutine start: %v", err)
	default:
	}
}

type signResponseCompletesPendingProviderRequestExecutor struct{}

func (signResponseCompletesPendingProviderRequestExecutor) RunStreamingDetailed(
	_ context.Context,
	_, _ string,
	onChunk adk.ChunkCallback,
	opts ...adk.RunOption,
) (adk.RunReport, error) {
	hooks := adk.ResolveRunHooks(opts...)
	defer func() {
		if hooks.OnFinish != nil {
			hooks.OnFinish()
		}
	}()
	if onChunk != nil {
		onChunk("chunk")
	}
	return adk.RunReport{Response: "complete"}, nil
}

func newSignResponseCompletesPendingProviderRequestCapturedProvider() (*security.RPCProvider, chan signResponseCompletesPendingProviderRequestCryptoRequest) {
	requests := make(chan signResponseCompletesPendingProviderRequestCryptoRequest, 1)
	provider := security.NewRPCProvider()
	installSignResponseCompletesPendingProviderRequestCaptureSender(provider, requests)
	return provider, requests
}

func installSignResponseCompletesPendingProviderRequestCaptureSender(provider *security.RPCProvider, requests chan<- signResponseCompletesPendingProviderRequestCryptoRequest) {
	provider.SetSender(func(event string, payload interface{}) error {
		select {
		case requests <- signResponseCompletesPendingProviderRequestCryptoRequest{event: event, payload: payload}:
			return nil
		default:
			return fmt.Errorf("unexpected extra request for %s", event)
		}
	})
}

func requireSignResponseCompletesPendingProviderRequestCryptoRequest(t *testing.T, requests <-chan signResponseCompletesPendingProviderRequestCryptoRequest) signResponseCompletesPendingProviderRequestCryptoRequest {
	t.Helper()

	select {
	case req := <-requests:
		return req
	case <-time.After(time.Second):
		t.Fatal("provider sender was not invoked")
		return signResponseCompletesPendingProviderRequestCryptoRequest{}
	}
}

func mustSignResponseCompletesPendingProviderRequestJSON(t *testing.T, v interface{}) json.RawMessage {
	t.Helper()

	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}
