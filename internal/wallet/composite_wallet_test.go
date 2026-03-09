package wallet

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompositeWallet_Address_UsesPrimary_WhenConnected(t *testing.T) {
	primary := &mockWalletProvider{address: "0xPRIMARY"}
	fallback := &mockWalletProvider{address: "0xFALLBACK"}
	checker := &mockConnectionChecker{connected: true}

	cw := NewCompositeWallet(primary, fallback, checker)
	ctx := context.Background()

	addr, err := cw.Address(ctx)
	require.NoError(t, err)
	assert.Equal(t, "0xPRIMARY", addr)
	assert.False(t, cw.UsedLocal())
}

func TestCompositeWallet_Address_UsesFallback_WhenDisconnected(t *testing.T) {
	primary := &mockWalletProvider{address: "0xPRIMARY"}
	fallback := &mockWalletProvider{address: "0xFALLBACK"}
	checker := &mockConnectionChecker{connected: false}

	cw := NewCompositeWallet(primary, fallback, checker)
	ctx := context.Background()

	addr, err := cw.Address(ctx)
	require.NoError(t, err)
	assert.Equal(t, "0xFALLBACK", addr)
	assert.True(t, cw.UsedLocal())
}

func TestCompositeWallet_Address_UsesFallback_WhenPrimaryErrors(t *testing.T) {
	primary := &mockWalletProvider{
		addressFn: func(_ context.Context) (string, error) {
			return "", errors.New("primary unavailable")
		},
	}
	fallback := &mockWalletProvider{address: "0xFALLBACK"}
	checker := &mockConnectionChecker{connected: true}

	cw := NewCompositeWallet(primary, fallback, checker)
	ctx := context.Background()

	addr, err := cw.Address(ctx)
	require.NoError(t, err)
	assert.Equal(t, "0xFALLBACK", addr)
	assert.True(t, cw.UsedLocal())
}

func TestCompositeWallet_Address_UsesFallback_WhenNilChecker(t *testing.T) {
	primary := &mockWalletProvider{address: "0xPRIMARY"}
	fallback := &mockWalletProvider{address: "0xFALLBACK"}

	cw := NewCompositeWallet(primary, fallback, nil)
	ctx := context.Background()

	addr, err := cw.Address(ctx)
	require.NoError(t, err)
	assert.Equal(t, "0xFALLBACK", addr)
	assert.True(t, cw.UsedLocal())
}

func TestCompositeWallet_Balance_AlwaysUsesFallback(t *testing.T) {
	tests := []struct {
		give      string
		connected bool
	}{
		{give: "connected", connected: true},
		{give: "disconnected", connected: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			primary := &mockWalletProvider{
				balanceFn: func(_ context.Context) (*big.Int, error) {
					return big.NewInt(999), nil
				},
			}
			fallback := &mockWalletProvider{balance: big.NewInt(42)}
			checker := &mockConnectionChecker{connected: tt.connected}

			cw := NewCompositeWallet(primary, fallback, checker)
			ctx := context.Background()

			bal, err := cw.Balance(ctx)
			require.NoError(t, err)
			assert.Equal(t, big.NewInt(42), bal, "Balance should always use fallback")
		})
	}
}

func TestCompositeWallet_SignTransaction_UsesPrimary_WhenConnected(t *testing.T) {
	primarySig := []byte("primary-sig")
	primary := &mockWalletProvider{
		signTxFn: func(_ context.Context, _ []byte) ([]byte, error) {
			return primarySig, nil
		},
	}
	fallback := &mockWalletProvider{
		signTxFn: func(_ context.Context, _ []byte) ([]byte, error) {
			return []byte("fallback-sig"), nil
		},
	}
	checker := &mockConnectionChecker{connected: true}

	cw := NewCompositeWallet(primary, fallback, checker)
	ctx := context.Background()

	sig, err := cw.SignTransaction(ctx, []byte("raw-tx"))
	require.NoError(t, err)
	assert.Equal(t, primarySig, sig)
	assert.False(t, cw.UsedLocal())
}

func TestCompositeWallet_SignTransaction_FallsBack_WhenPrimaryErrors(t *testing.T) {
	primary := &mockWalletProvider{
		signTxFn: func(_ context.Context, _ []byte) ([]byte, error) {
			return nil, errors.New("primary sign error")
		},
	}
	fallbackSig := []byte("fallback-sig")
	fallback := &mockWalletProvider{
		signTxFn: func(_ context.Context, _ []byte) ([]byte, error) {
			return fallbackSig, nil
		},
	}
	checker := &mockConnectionChecker{connected: true}

	cw := NewCompositeWallet(primary, fallback, checker)
	ctx := context.Background()

	sig, err := cw.SignTransaction(ctx, []byte("raw-tx"))
	require.NoError(t, err)
	assert.Equal(t, fallbackSig, sig)
	assert.True(t, cw.UsedLocal())
}

func TestCompositeWallet_SignTransaction_UsesFallback_WhenDisconnected(t *testing.T) {
	primary := &mockWalletProvider{
		signTxFn: func(_ context.Context, _ []byte) ([]byte, error) {
			return []byte("should-not-be-called"), nil
		},
	}
	fallbackSig := []byte("fallback-sig")
	fallback := &mockWalletProvider{
		signTxFn: func(_ context.Context, _ []byte) ([]byte, error) {
			return fallbackSig, nil
		},
	}
	checker := &mockConnectionChecker{connected: false}

	cw := NewCompositeWallet(primary, fallback, checker)
	ctx := context.Background()

	sig, err := cw.SignTransaction(ctx, []byte("raw-tx"))
	require.NoError(t, err)
	assert.Equal(t, fallbackSig, sig)
	assert.True(t, cw.UsedLocal())
}

func TestCompositeWallet_SignMessage_UsesPrimary_WhenConnected(t *testing.T) {
	primarySig := []byte("primary-msg-sig")
	primary := &mockWalletProvider{
		signMsgFn: func(_ context.Context, _ []byte) ([]byte, error) {
			return primarySig, nil
		},
	}
	fallback := &mockWalletProvider{
		signMsgFn: func(_ context.Context, _ []byte) ([]byte, error) {
			return []byte("fallback-msg-sig"), nil
		},
	}
	checker := &mockConnectionChecker{connected: true}

	cw := NewCompositeWallet(primary, fallback, checker)
	ctx := context.Background()

	sig, err := cw.SignMessage(ctx, []byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, primarySig, sig)
	assert.False(t, cw.UsedLocal())
}

func TestCompositeWallet_SignMessage_FallsBack_WhenPrimaryErrors(t *testing.T) {
	primary := &mockWalletProvider{
		signMsgFn: func(_ context.Context, _ []byte) ([]byte, error) {
			return nil, errors.New("primary sign msg error")
		},
	}
	fallbackSig := []byte("fallback-msg-sig")
	fallback := &mockWalletProvider{
		signMsgFn: func(_ context.Context, _ []byte) ([]byte, error) {
			return fallbackSig, nil
		},
	}
	checker := &mockConnectionChecker{connected: true}

	cw := NewCompositeWallet(primary, fallback, checker)
	ctx := context.Background()

	sig, err := cw.SignMessage(ctx, []byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, fallbackSig, sig)
	assert.True(t, cw.UsedLocal())
}

func TestCompositeWallet_SignMessage_UsesFallback_WhenDisconnected(t *testing.T) {
	primary := &mockWalletProvider{}
	fallbackSig := []byte("fallback-msg-sig")
	fallback := &mockWalletProvider{
		signMsgFn: func(_ context.Context, _ []byte) ([]byte, error) {
			return fallbackSig, nil
		},
	}
	checker := &mockConnectionChecker{connected: false}

	cw := NewCompositeWallet(primary, fallback, checker)
	ctx := context.Background()

	sig, err := cw.SignMessage(ctx, []byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, fallbackSig, sig)
	assert.True(t, cw.UsedLocal())
}

func TestCompositeWallet_PublicKey_UsesPrimary_WhenConnected(t *testing.T) {
	primaryPK := []byte{0x02, 0xAA, 0xBB}
	primary := &mockWalletProvider{
		pubKeyFn: func(_ context.Context) ([]byte, error) {
			return primaryPK, nil
		},
	}
	fallback := &mockWalletProvider{
		pubKeyFn: func(_ context.Context) ([]byte, error) {
			return []byte{0x02, 0xCC, 0xDD}, nil
		},
	}
	checker := &mockConnectionChecker{connected: true}

	cw := NewCompositeWallet(primary, fallback, checker)
	ctx := context.Background()

	pk, err := cw.PublicKey(ctx)
	require.NoError(t, err)
	assert.Equal(t, primaryPK, pk)
	assert.False(t, cw.UsedLocal())
}

func TestCompositeWallet_PublicKey_FallsBack_WhenPrimaryErrors(t *testing.T) {
	primary := &mockWalletProvider{
		pubKeyFn: func(_ context.Context) ([]byte, error) {
			return nil, errors.New("primary pk error")
		},
	}
	fallbackPK := []byte{0x03, 0xEE, 0xFF}
	fallback := &mockWalletProvider{
		pubKeyFn: func(_ context.Context) ([]byte, error) {
			return fallbackPK, nil
		},
	}
	checker := &mockConnectionChecker{connected: true}

	cw := NewCompositeWallet(primary, fallback, checker)
	ctx := context.Background()

	pk, err := cw.PublicKey(ctx)
	require.NoError(t, err)
	assert.Equal(t, fallbackPK, pk)
	assert.True(t, cw.UsedLocal())
}

func TestCompositeWallet_PublicKey_UsesFallback_WhenNilChecker(t *testing.T) {
	primary := &mockWalletProvider{
		pubKeyFn: func(_ context.Context) ([]byte, error) {
			return []byte{0x02, 0x11}, nil
		},
	}
	fallbackPK := []byte{0x03, 0x22}
	fallback := &mockWalletProvider{
		pubKeyFn: func(_ context.Context) ([]byte, error) {
			return fallbackPK, nil
		},
	}

	cw := NewCompositeWallet(primary, fallback, nil)
	ctx := context.Background()

	pk, err := cw.PublicKey(ctx)
	require.NoError(t, err)
	assert.Equal(t, fallbackPK, pk)
	assert.True(t, cw.UsedLocal())
}

func TestCompositeWallet_UsedLocal_InitiallyFalse(t *testing.T) {
	cw := NewCompositeWallet(
		&mockWalletProvider{},
		&mockWalletProvider{},
		&mockConnectionChecker{connected: true},
	)
	assert.False(t, cw.UsedLocal())
}

func TestCompositeWallet_UsedLocal_Sticky(t *testing.T) {
	primary := &mockWalletProvider{address: "0xPRIMARY"}
	fallback := &mockWalletProvider{address: "0xFALLBACK"}
	checker := &mockConnectionChecker{connected: false}

	cw := NewCompositeWallet(primary, fallback, checker)
	ctx := context.Background()

	// First call: disconnected, uses fallback.
	_, err := cw.Address(ctx)
	require.NoError(t, err)
	assert.True(t, cw.UsedLocal())

	// Now reconnect.
	checker.connected = true

	// Second call: connected, uses primary.
	addr, err := cw.Address(ctx)
	require.NoError(t, err)
	assert.Equal(t, "0xPRIMARY", addr)

	// usedLocal should remain true (sticky).
	assert.True(t, cw.UsedLocal())
}

func TestCompositeWallet_BothFallbackErrors_Propagate(t *testing.T) {
	primary := &mockWalletProvider{
		addressFn: func(_ context.Context) (string, error) {
			return "", errors.New("primary error")
		},
	}
	fallback := &mockWalletProvider{
		addressFn: func(_ context.Context) (string, error) {
			return "", errors.New("fallback error")
		},
	}
	checker := &mockConnectionChecker{connected: true}

	cw := NewCompositeWallet(primary, fallback, checker)
	ctx := context.Background()

	_, err := cw.Address(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fallback error")
}
