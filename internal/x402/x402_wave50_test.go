package x402

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/coinbase/x402/go/mechanisms/evm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

func TestWave50InterceptorAccessorsReturnEnabledStateAndSignerAddress(t *testing.T) {
	t.Parallel()

	enabled := NewInterceptor(
		wave50X402SignerProvider{signer: wave50X402Signer{address: "0x0000000000000000000000000000000000000050"}},
		nil,
		Config{Enabled: true, ChainID: 84532},
		zap.NewNop().Sugar(),
	)
	disabled := NewInterceptor(wave50X402SignerProvider{}, nil, Config{Enabled: false}, zap.NewNop().Sugar())

	assert.True(t, enabled.IsEnabled())
	assert.False(t, disabled.IsEnabled())

	address, err := enabled.SignerAddress(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "0x0000000000000000000000000000000000000050", address)
}

func TestWave50InterceptorSignerAddressPropagatesProviderError(t *testing.T) {
	t.Parallel()

	interceptor := NewInterceptor(
		wave50X402SignerProvider{err: errors.New("wallet unavailable")},
		nil,
		Config{Enabled: true, ChainID: 84532},
		zap.NewNop().Sugar(),
	)

	address, err := interceptor.SignerAddress(context.Background())
	require.Error(t, err)
	assert.Empty(t, address)
	assert.ErrorContains(t, err, "wallet unavailable")
}

func TestWave50LocalSignerProviderLoadsWalletKey(t *testing.T) {
	secrets := newWave50X402SecretsStore(t)
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	keyBytes := crypto.FromECDSA(privateKey)
	require.NoError(t, secrets.Store(context.Background(), wallet.WalletKeyName, keyBytes))

	provider := NewLocalSignerProvider(secrets)
	assert.Same(t, secrets, provider.secrets)
	assert.Equal(t, wallet.WalletKeyName, provider.keyName)

	signer, err := provider.EvmSigner(context.Background())
	require.NoError(t, err)
	assert.Equal(t, crypto.PubkeyToAddress(privateKey.PublicKey).Hex(), signer.Address())
}

func TestWave50LocalSignerProviderWrapsMissingWalletKey(t *testing.T) {
	secrets := newWave50X402SecretsStore(t)
	provider := NewLocalSignerProvider(secrets)

	signer, err := provider.EvmSigner(context.Background())
	require.Error(t, err)
	assert.Nil(t, signer)
	assert.ErrorContains(t, err, "load wallet key")
	assert.ErrorContains(t, err, wallet.WalletKeyName)
}

func newWave50X402SecretsStore(t *testing.T) *security.SecretsStore {
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

type wave50X402SignerProvider struct {
	signer evm.ClientEvmSigner
	err    error
}

func (p wave50X402SignerProvider) EvmSigner(context.Context) (evm.ClientEvmSigner, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.signer, nil
}

type wave50X402Signer struct {
	address string
}

func (s wave50X402Signer) Address() string { return s.address }

func (s wave50X402Signer) SignTypedData(context.Context, evm.TypedDataDomain, map[string][]evm.TypedDataField, string, map[string]interface{}) ([]byte, error) {
	return []byte("signature"), nil
}
