package payment

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/coinbase/x402/go/mechanisms/evm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ent/enttest"
	corepayment "github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/paymentgate"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/wallet"
	"github.com/langoai/lango/internal/x402"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

func TestWave50DenyAllPaymentExecutionGate_DeniesMissingReceipt(t *testing.T) {
	t.Parallel()

	got, err := (DenyAllPaymentExecutionGate{}).EvaluateDirectPayment(
		context.Background(),
		paymentgate.Request{
			TransactionReceiptID: "tx-1",
			SubmissionReceiptID:  "sub-1",
			ToolName:             "payment_send",
		},
	)

	require.NoError(t, err)
	assert.Equal(t, paymentgate.Deny, got.Decision)
	assert.Equal(t, paymentgate.ReasonMissingReceipt, got.Reason)
	assert.Empty(t, got.SubmissionReceiptID)
}

func TestWave50CreateWalletTool_CreatesAndReportsExistingWallet(t *testing.T) {
	secrets := newWave50SecretsStore(t)
	tool := buildCreateWalletTool(secrets, 84532)

	assert.Equal(t, "payment_create_wallet", tool.Name)
	assert.Equal(t, agent.SafetyLevelDangerous, tool.SafetyLevel)
	assert.Equal(t, "payment", tool.Capability.Category)
	assert.Equal(t, agent.ActivityExecute, tool.Capability.Activity)
	assert.Contains(t, tool.Capability.RequiredCapabilities, "payment")

	result, err := tool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	created := requireWave50Map(t, result)
	assert.Equal(t, "created", created["status"])
	assert.Regexp(t, regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`), created["address"])
	assert.Equal(t, int64(84532), created["chainId"])
	assert.Equal(t, wallet.NetworkName(84532), created["network"])

	result, err = tool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	existing := requireWave50Map(t, result)
	assert.Equal(t, "exists", existing["status"])
	assert.Equal(t, created["address"], existing["address"])
	assert.Contains(t, existing["message"], "Wallet already exists")
}

func TestWave50X402FetchTool_FetchesTruncatesAndRecordsPaymentResponse(t *testing.T) {
	const signerAddress = "0x0000000000000000000000000000000000000050"
	handlerErr := make(chan error, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			handlerErr <- errors.New("unexpected method: " + r.Method)
			http.Error(w, "unexpected method", http.StatusBadRequest)
			return
		}
		if got := r.Header.Get("X-Test"); got != "yes" {
			handlerErr <- errors.New("unexpected X-Test header: " + got)
			http.Error(w, "unexpected header", http.StatusBadRequest)
			return
		}
		if got := r.Header.Get("Ignored"); got != "" {
			handlerErr <- errors.New("unexpected ignored header: " + got)
			http.Error(w, "unexpected ignored header", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			handlerErr <- err
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}
		if string(body) != "payload" {
			handlerErr <- errors.New("unexpected body: " + string(body))
			http.Error(w, "unexpected body", http.StatusBadRequest)
			return
		}

		w.Header().Set("Payment-Response", "paid")
		w.Header().Set("X-Reply", "ok")
		_, _ = w.Write([]byte(strings.Repeat("x", 8200)))
		handlerErr <- nil
	}))
	defer srv.Close()

	interceptor := x402.NewInterceptor(
		wave50SignerProvider{signer: wave50EvmSigner{address: signerAddress}},
		nil,
		x402.Config{Enabled: true, ChainID: 84532},
		zap.NewNop().Sugar(),
	)
	svc := &wave50PaymentRecorder{}
	tool := buildX402FetchTool(interceptor, svc)

	assert.Equal(t, "payment_x402_fetch", tool.Name)
	assert.Equal(t, agent.SafetyLevelDangerous, tool.SafetyLevel)
	assert.Equal(t, agent.ActivityExecute, tool.Capability.Activity)
	assert.Contains(t, tool.Capability.RequiredCapabilities, "payment")

	result, err := tool.Handler(context.Background(), map[string]interface{}{
		"url":    srv.URL,
		"method": http.MethodPost,
		"body":   "payload",
		"headers": map[string]interface{}{
			"X-Test":  "yes",
			"Ignored": 42,
		},
	})
	require.NoError(t, err)
	require.NoError(t, <-handlerErr)

	payload := requireWave50Map(t, result)
	assert.Equal(t, http.StatusOK, payload["statusCode"])
	assert.Len(t, payload["body"], 8192)
	assert.Equal(t, true, payload["truncated"])

	headers, ok := payload["headers"].(map[string]string)
	require.True(t, ok)
	assert.Equal(t, "paid", headers["Payment-Response"])
	assert.Equal(t, "ok", headers["X-Reply"])

	require.Len(t, svc.records, 1)
	assert.Equal(t, srv.URL, svc.records[0].URL)
	assert.Equal(t, signerAddress, svc.records[0].From)
	assert.Equal(t, int64(84532), svc.records[0].ChainID)
}

func TestWave50X402FetchTool_PropagatesSignerProviderError(t *testing.T) {
	t.Parallel()

	interceptor := x402.NewInterceptor(
		wave50SignerProvider{err: errors.New("signer unavailable")},
		nil,
		x402.Config{Enabled: true, ChainID: 84532},
		zap.NewNop().Sugar(),
	)
	tool := buildX402FetchTool(interceptor, nil)

	got, err := tool.Handler(context.Background(), map[string]interface{}{"url": "http://example.test"})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "create X402 HTTP client")
	assert.ErrorContains(t, err, "signer unavailable")
}

func newWave50SecretsStore(t *testing.T) *security.SecretsStore {
	t.Helper()

	client := enttest.Open(t, "sqlite3", "file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	cryptoProvider := security.NewLocalCryptoProvider()
	require.NoError(t, cryptoProvider.Initialize("wave50-passphrase"))

	registry := security.NewKeyRegistry(client)
	_, err := registry.RegisterKey(context.Background(), "default", "local", security.KeyTypeEncryption)
	require.NoError(t, err)

	return security.NewSecretsStore(client, registry, cryptoProvider)
}

func requireWave50Map(t *testing.T, result interface{}) map[string]interface{} {
	t.Helper()

	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	return payload
}

type wave50PaymentRecorder struct {
	records []corepayment.X402PaymentRecord
}

func (s *wave50PaymentRecorder) Send(context.Context, corepayment.PaymentRequest) (*corepayment.PaymentReceipt, error) {
	return nil, errors.New("unexpected send")
}

func (s *wave50PaymentRecorder) Balance(context.Context) (string, error) {
	return "", errors.New("unexpected balance")
}

func (s *wave50PaymentRecorder) History(context.Context, int) ([]corepayment.TransactionInfo, error) {
	return nil, errors.New("unexpected history")
}

func (s *wave50PaymentRecorder) WalletAddress(context.Context) (string, error) {
	return "", errors.New("unexpected wallet address")
}

func (s *wave50PaymentRecorder) ChainID() int64 { return 84532 }

func (s *wave50PaymentRecorder) RecordX402Payment(_ context.Context, record corepayment.X402PaymentRecord) error {
	s.records = append(s.records, record)
	return nil
}

type wave50SignerProvider struct {
	signer evm.ClientEvmSigner
	err    error
}

func (p wave50SignerProvider) EvmSigner(context.Context) (evm.ClientEvmSigner, error) {
	if p.err != nil {
		return nil, p.err
	}
	if p.signer != nil {
		return p.signer, nil
	}
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return wave50PrivateKeySigner{key: key}, nil
}

type wave50EvmSigner struct {
	address string
}

func (s wave50EvmSigner) Address() string { return s.address }

func (s wave50EvmSigner) SignTypedData(context.Context, evm.TypedDataDomain, map[string][]evm.TypedDataField, string, map[string]interface{}) ([]byte, error) {
	return []byte("signature"), nil
}

type wave50PrivateKeySigner struct {
	key *ecdsa.PrivateKey
}

func (s wave50PrivateKeySigner) Address() string {
	return crypto.PubkeyToAddress(s.key.PublicKey).Hex()
}

func (s wave50PrivateKeySigner) SignTypedData(context.Context, evm.TypedDataDomain, map[string][]evm.TypedDataField, string, map[string]interface{}) ([]byte, error) {
	return make([]byte, 65), nil
}
