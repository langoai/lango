package contract

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/wallet"
)

// Compile-time interface check.
var _ wallet.WalletProvider = (*wallet.LocalWallet)(nil)

func TestNewCaller(t *testing.T) {
	cache := NewABICache()
	caller := NewCaller(nil, nil, 8453, cache)

	assert.NotNil(t, caller)
	assert.Equal(t, int64(8453), caller.chainID.Int64())
	assert.Equal(t, DefaultTimeout, caller.timeout)
	assert.Equal(t, MaxRetries, caller.maxRetries)
}

func TestCaller_LoadABI(t *testing.T) {
	tests := []struct {
		give    string
		wantErr bool
	}{
		{give: testERC20ABI, wantErr: false},
		{give: "invalid json", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			cache := NewABICache()
			caller := NewCaller(nil, nil, 1, cache)
			addr := common.HexToAddress("0x1111111111111111111111111111111111111111")

			err := caller.LoadABI(1, addr, tt.give)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				// Verify it was cached.
				_, ok := cache.Get(1, addr)
				assert.True(t, ok)
			}
		})
	}
}
