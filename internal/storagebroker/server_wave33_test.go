package storagebroker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	sec "github.com/langoai/lango/internal/security"
	"github.com/stretchr/testify/require"
)

func TestWave33ServerHandleEncodesResultAndPayloadErrors(t *testing.T) {
	t.Parallel()

	srv := NewServer()
	resp := srv.handle(Request{ID: 3301, Method: methodHealth})
	require.True(t, resp.OK)
	require.Equal(t, uint64(3301), resp.ID)
	require.Empty(t, resp.Error)
	require.JSONEq(t, `{"opened":false}`, string(resp.Result))

	var health HealthResult
	require.NoError(t, json.Unmarshal(resp.Result, &health))
	require.False(t, health.Opened)

	errResp := srv.handle(Request{
		ID:      3302,
		Method:  methodDBStatus,
		Payload: json.RawMessage(`{"db_path":`),
	})
	require.False(t, errResp.OK)
	require.Equal(t, uint64(3302), errResp.ID)
	require.Empty(t, errResp.Result)
	require.Contains(t, errResp.Error, "decode broker payload")
}

func TestWave33DecodePayloadEmptyValidAndMalformedInputs(t *testing.T) {
	t.Parallel()

	req := DBStatusSummaryRequest{DBPath: "unchanged.db"}
	require.NoError(t, decodePayload(nil, &req))
	require.Equal(t, "unchanged.db", req.DBPath)

	raw := mustPayload(t, EncryptPayloadRequest{Plaintext: []byte("plain")})
	require.JSONEq(t, `{"plaintext":"cGxhaW4="}`, string(raw))

	var encReq EncryptPayloadRequest
	require.NoError(t, decodePayload(raw, &encReq))
	require.Equal(t, []byte("plain"), encReq.Plaintext)

	err := decodePayload(json.RawMessage(`{"plaintext":`), &encReq)
	require.ErrorContains(t, err, "decode broker payload")
}

func TestWave33ServerPayloadProtectorEncryptDecryptAndVersionMismatch(t *testing.T) {
	t.Parallel()

	key := bytes.Repeat([]byte{0x72}, 32)
	defaultVersion := &serverPayloadProtector{key: key}

	ciphertext, nonce, version, err := defaultVersion.EncryptPayload([]byte("server secret"))
	require.NoError(t, err)
	require.Equal(t, sec.PayloadKeyVersionV1, version)
	require.NotEqual(t, []byte("server secret"), ciphertext)
	require.NotEmpty(t, nonce)

	plaintext, err := defaultVersion.DecryptPayload(ciphertext, nonce, sec.PayloadKeyVersionV1)
	require.NoError(t, err)
	require.Equal(t, []byte("server secret"), plaintext)

	_, err = defaultVersion.DecryptPayload(ciphertext, nonce, 99)
	require.ErrorContains(t, err, "unsupported payload key version 99")

	versioned := &serverPayloadProtector{key: key, version: 7}
	ciphertext, nonce, version, err = versioned.EncryptPayload([]byte("versioned secret"))
	require.NoError(t, err)
	require.Equal(t, 7, version)

	plaintext, err = versioned.DecryptPayload(ciphertext, nonce, 7)
	require.NoError(t, err)
	require.Equal(t, []byte("versioned secret"), plaintext)

	_, err = versioned.DecryptPayload(ciphertext, nonce, 8)
	require.ErrorContains(t, err, "unsupported payload key version 8")
}

func TestWave33PayloadProtectorWrapsBrokerAPI(t *testing.T) {
	t.Parallel()

	require.Nil(t, NewPayloadProtector(nil))

	api := &wave33PayloadAPI{}
	protector := NewPayloadProtector(api)
	ciphertext, nonce, version, err := protector.EncryptPayload([]byte("wrapped plain"))
	require.NoError(t, err)
	require.Equal(t, []byte("wrapped plain"), api.encryptPlaintext)
	require.Equal(t, []byte("wrapped cipher"), ciphertext)
	require.Equal(t, []byte("wrapped nonce"), nonce)
	require.Equal(t, 12, version)

	plaintext, err := protector.DecryptPayload([]byte("cipher in"), []byte("nonce in"), 12)
	require.NoError(t, err)
	require.Equal(t, []byte("cipher in"), api.decryptCiphertext)
	require.Equal(t, []byte("nonce in"), api.decryptNonce)
	require.Equal(t, 12, api.decryptVersion)
	require.Equal(t, []byte("wrapped decrypted"), plaintext)

	api.encryptErr = errors.New("encrypt unavailable")
	_, _, _, err = protector.EncryptPayload([]byte("fail"))
	require.ErrorContains(t, err, "encrypt unavailable")

	api.decryptErr = errors.New("decrypt unavailable")
	_, err = protector.DecryptPayload([]byte("cipher"), []byte("nonce"), 12)
	require.ErrorContains(t, err, "decrypt unavailable")
}

func TestWave33OperationalDispatchRequiresOpenedDatabase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	srv := NewServer()
	now := time.Unix(3300, 0).UTC()

	tests := []struct {
		name    string
		method  string
		payload any
	}{
		{name: "learning history", method: methodLearningHistory, payload: LearningHistoryRequest{Limit: 1}},
		{name: "pending inquiries", method: methodPendingInquiries, payload: PendingInquiriesRequest{Limit: 1}},
		{name: "workflow runs", method: methodWorkflowRuns, payload: WorkflowRunsRequest{Limit: 1}},
		{name: "alerts", method: methodAlerts, payload: AlertsRequest{From: now}},
		{name: "reputation get", method: methodReputationGet, payload: ReputationGetRequest{PeerDID: "did:example:missing"}},
		{name: "payment history", method: methodPaymentHistory, payload: PaymentHistoryRequest{Limit: 1}},
		{name: "payment usage", method: methodPaymentUsage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := Request{Method: tt.method}
			if tt.payload != nil {
				req.Payload = mustPayload(t, tt.payload)
			}

			_, err := srv.dispatch(ctx, req)
			require.ErrorContains(t, err, "database not opened")
		})
	}
}

type wave33PayloadAPI struct {
	API

	encryptPlaintext  []byte
	encryptErr        error
	decryptCiphertext []byte
	decryptNonce      []byte
	decryptVersion    int
	decryptErr        error
}

func (a *wave33PayloadAPI) EncryptPayload(_ context.Context, plaintext []byte) (EncryptPayloadResult, error) {
	a.encryptPlaintext = append([]byte(nil), plaintext...)
	if a.encryptErr != nil {
		return EncryptPayloadResult{}, a.encryptErr
	}
	return EncryptPayloadResult{
		Ciphertext: []byte("wrapped cipher"),
		Nonce:      []byte("wrapped nonce"),
		KeyVersion: 12,
	}, nil
}

func (a *wave33PayloadAPI) DecryptPayload(_ context.Context, ciphertext, nonce []byte, keyVersion int) (DecryptPayloadResult, error) {
	a.decryptCiphertext = append([]byte(nil), ciphertext...)
	a.decryptNonce = append([]byte(nil), nonce...)
	a.decryptVersion = keyVersion
	if a.decryptErr != nil {
		return DecryptPayloadResult{}, a.decryptErr
	}
	return DecryptPayloadResult{Plaintext: []byte("wrapped decrypted")}, nil
}
