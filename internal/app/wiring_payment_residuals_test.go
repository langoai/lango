package app

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/testutil"
)

func TestInitPaymentWithStorageResidualEarlyExits(t *testing.T) {
	t.Parallel()

	secrets := newPaymentResidualsSecretsStore(t)

	tests := []struct {
		name    string
		secrets *security.SecretsStore
		facade  *storage.Facade
		edit    func(*config.Config)
	}{
		{
			name:    "nil storage facade",
			secrets: secrets,
		},
		{
			name:    "missing payment tx store",
			secrets: secrets,
			facade:  storage.NewFacade(nil, nil),
			edit: func(cfg *config.Config) {
				cfg.Payment.Network.RPCURL = "http://127.0.0.1:1"
			},
		},
		{
			name:    "missing rpc url",
			secrets: secrets,
			facade:  storage.NewFacade(nil, nil, storage.WithEntClient(testutil.TestEntClient(t))),
		},
		{
			name:    "rpc connection failure",
			secrets: secrets,
			facade:  storage.NewFacade(nil, nil, storage.WithEntClient(testutil.TestEntClient(t))),
			edit: func(cfg *config.Config) {
				cfg.Payment.Network.RPCURL = "://invalid-rpc-url"
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.DefaultConfig()
			cfg.Payment.Enabled = true
			cfg.Payment.Network.RPCURL = ""
			if tt.edit != nil {
				tt.edit(cfg)
			}

			assert.Nil(t, initPaymentWithStorage(cfg, tt.secrets, tt.facade))
		})
	}
}

func TestInitX402ResidualDisabledAndDefaultLimitBranches(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Payment.Enabled = false
	assert.Nil(t, initX402(cfg, newPaymentResidualsSecretsStore(t), nil))

	cfg.Payment.Enabled = true
	assert.Nil(t, initX402(cfg, nil, nil))

	cfg.Payment.Limits.MaxPerTx = ""
	cfg.Payment.Network.ChainID = 84532
	components := initX402(cfg, newPaymentResidualsSecretsStore(t), nil)
	require.NotNil(t, components)
	require.NotNil(t, components.interceptor)
	assert.True(t, components.interceptor.IsEnabled())
	assert.Equal(t, int64(84532), components.interceptor.ChainID())
	assert.Equal(t, "1.00", paymentResidualsX402MaxAutoPayAmount(t, components.interceptor))
}

func paymentResidualsX402MaxAutoPayAmount(t *testing.T, interceptor any) string {
	t.Helper()

	value := reflect.ValueOf(interceptor)
	require.Equal(t, reflect.Ptr, value.Kind())
	configValue := value.Elem().FieldByName("config")
	require.True(t, configValue.IsValid())
	maxAutoPay := configValue.FieldByName("MaxAutoPayAmount")
	require.True(t, maxAutoPay.IsValid())
	return maxAutoPay.String()
}

func newPaymentResidualsSecretsStore(t *testing.T) *security.SecretsStore {
	t.Helper()

	client := testutil.TestEntClient(t)
	keys := security.NewKeyRegistry(client)
	_, err := keys.RegisterKey(context.Background(), "default-enc", "local", security.KeyTypeEncryption)
	require.NoError(t, err)

	crypto := security.NewLocalCryptoProvider()
	require.NoError(t, crypto.Initialize("payment-residuals-password"))
	return security.NewSecretsStore(client, keys, crypto)
}
