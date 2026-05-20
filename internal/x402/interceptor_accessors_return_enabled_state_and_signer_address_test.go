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

func TestInterceptorAccessorsReturnEnabledStateAndSignerAddress(t *testing.T) {
	t.Parallel()

	enabled := NewInterceptor(
		interceptorAccessorsReturnEnabledStateAndSignerAddressX402SignerProvider{signer: interceptorAccessorsReturnEnabledStateAndSignerAddressX402Signer{address: "0x0000000000000000000000000000000000000050"}},
		nil,
		Config{Enabled: true, ChainID: 84532},
		zap.NewNop().Sugar(),
	)
	disabled := NewInterceptor(interceptorAccessorsReturnEnabledStateAndSignerAddressX402SignerProvider{}, nil, Config{Enabled: false}, zap.NewNop().Sugar())

	assert.True(t, enabled.IsEnabled())
	assert.False(t, disabled.IsEnabled())

	address, err := enabled.SignerAddress(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "0x0000000000000000000000000000000000000050", address)
}

func TestInterceptorSignerAddressPropagatesProviderError(t *testing.T) {
	t.Parallel()

	interceptor := NewInterceptor(
		interceptorAccessorsReturnEnabledStateAndSignerAddressX402SignerProvider{err: errors.New("wallet unavailable")},
		nil,
		Config{Enabled: true, ChainID: 84532},
		zap.NewNop().Sugar(),
	)

	address, err := interceptor.SignerAddress(context.Background())
	require.Error(t, err)
	assert.Empty(t, address)
	assert.ErrorContains(t, err, "wallet unavailable")
}

func TestLocalSignerProviderLoadsWalletKey(t *testing.T) {
	secrets := newInterceptorAccessorsReturnEnabledStateAndSignerAddressX402SecretsStore(t)
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

func TestLocalSignerProviderWrapsMissingWalletKey(t *testing.T) {
	secrets := newInterceptorAccessorsReturnEnabledStateAndSignerAddressX402SecretsStore(t)
	provider := NewLocalSignerProvider(secrets)

	signer, err := provider.EvmSigner(context.Background())
	require.Error(t, err)
	assert.Nil(t, signer)
	assert.ErrorContains(t, err, "load wallet key")
	assert.ErrorContains(t, err, wallet.WalletKeyName)
}

func newInterceptorAccessorsReturnEnabledStateAndSignerAddressX402SecretsStore(t *testing.T) *security.SecretsStore {
	t.Helper()

	client := enttest.Open(t, "sqlite3", "file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	cryptoProvider := security.NewLocalCryptoProvider()
	require.NoError(t, cryptoProvider.Initialize("newDeadLetterStatusLoaderWrapsBuildAppError0-passphrase"))

	registry := security.NewKeyRegistry(client)
	_, err := registry.RegisterKey(context.Background(), "default", "local", security.KeyTypeEncryption)
	require.NoError(t, err)

	return security.NewSecretsStore(client, registry, cryptoProvider)
}

type interceptorAccessorsReturnEnabledStateAndSignerAddressX402SignerProvider struct {
	signer evm.ClientEvmSigner
	err    error
}

func (p interceptorAccessorsReturnEnabledStateAndSignerAddressX402SignerProvider) EvmSigner(context.Context) (evm.ClientEvmSigner, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.signer, nil
}

type interceptorAccessorsReturnEnabledStateAndSignerAddressX402Signer struct {
	address string
}

func (s interceptorAccessorsReturnEnabledStateAndSignerAddressX402Signer) Address() string {
	return s.address
}

func (s interceptorAccessorsReturnEnabledStateAndSignerAddressX402Signer) SignTypedData(context.Context, evm.TypedDataDomain, map[string][]evm.TypedDataField, string, map[string]interface{}) ([]byte, error) {
	return []byte("signature"), nil
}
