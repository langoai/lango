package wallet

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetworkName_AllChainIDs(t *testing.T) {
	tests := []struct {
		give int64
		want string
	}{
		{give: int64(ChainEthereumMainnet), want: "Ethereum Mainnet"},
		{give: int64(ChainBase), want: "Base"},
		{give: int64(ChainBaseSepolia), want: "Base Sepolia"},
		{give: int64(ChainSepolia), want: "Sepolia"},
		{give: 0, want: "Unknown"},
		{give: -1, want: "Unknown"},
		{give: 42161, want: "Unknown"},
		{give: 137, want: "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := NetworkName(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChainIDConstants(t *testing.T) {
	assert.Equal(t, ChainID(1), ChainEthereumMainnet)
	assert.Equal(t, ChainID(8453), ChainBase)
	assert.Equal(t, ChainID(84532), ChainBaseSepolia)
	assert.Equal(t, ChainID(11155111), ChainSepolia)
}

func TestCurrencyUSDC(t *testing.T) {
	assert.Equal(t, "USDC", CurrencyUSDC)
}

func TestWalletInfo_Fields(t *testing.T) {
	info := WalletInfo{
		Address: "0x1234567890abcdef1234567890abcdef12345678",
		ChainID: 1,
		Network: "Ethereum Mainnet",
	}

	assert.Equal(t, "0x1234567890abcdef1234567890abcdef12345678", info.Address)
	assert.Equal(t, int64(1), info.ChainID)
	assert.Equal(t, "Ethereum Mainnet", info.Network)
}

func TestZeroBytes(t *testing.T) {
	tests := []struct {
		give string
		size int
	}{
		{give: "empty slice", size: 0},
		{give: "single byte", size: 1},
		{give: "32 bytes (key-sized)", size: 32},
		{give: "64 bytes (sig-sized)", size: 64},
		{give: "256 bytes", size: 256},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			b := make([]byte, tt.size)
			// Fill with non-zero values.
			for i := range b {
				b[i] = 0xFF
			}

			zeroBytes(b)

			for i, v := range b {
				assert.Equal(t, byte(0), v, "byte at index %d should be zero", i)
			}
		})
	}
}

func TestZeroBytes_PreservesLength(t *testing.T) {
	b := make([]byte, 42)
	for i := range b {
		b[i] = byte(i)
	}

	zeroBytes(b)

	assert.Len(t, b, 42)
}

// mockWalletProvider implements WalletProvider for testing composite logic.
type mockWalletProvider struct {
	address   string
	addressFn func(ctx context.Context) (string, error)
	balance   *big.Int
	balanceFn func(ctx context.Context) (*big.Int, error)
	signTxFn  func(ctx context.Context, rawTx []byte) ([]byte, error)
	signMsgFn func(ctx context.Context, message []byte) ([]byte, error)
	pubKeyFn  func(ctx context.Context) ([]byte, error)
}

func (m *mockWalletProvider) Address(ctx context.Context) (string, error) {
	if m.addressFn != nil {
		return m.addressFn(ctx)
	}
	if m.address != "" {
		return m.address, nil
	}
	return "", errors.New("no address configured")
}

func (m *mockWalletProvider) Balance(ctx context.Context) (*big.Int, error) {
	if m.balanceFn != nil {
		return m.balanceFn(ctx)
	}
	if m.balance != nil {
		return m.balance, nil
	}
	return nil, errors.New("no balance configured")
}

func (m *mockWalletProvider) SignTransaction(ctx context.Context, rawTx []byte) ([]byte, error) {
	if m.signTxFn != nil {
		return m.signTxFn(ctx, rawTx)
	}
	return nil, errors.New("sign tx not configured")
}

func (m *mockWalletProvider) SignMessage(ctx context.Context, message []byte) ([]byte, error) {
	if m.signMsgFn != nil {
		return m.signMsgFn(ctx, message)
	}
	return nil, errors.New("sign msg not configured")
}

func (m *mockWalletProvider) PublicKey(ctx context.Context) ([]byte, error) {
	if m.pubKeyFn != nil {
		return m.pubKeyFn(ctx)
	}
	return nil, errors.New("public key not configured")
}

// mockConnectionChecker implements ConnectionChecker for testing.
type mockConnectionChecker struct {
	connected bool
}

func (m *mockConnectionChecker) IsConnected() bool {
	return m.connected
}

// Compile-time interface compliance checks.
var _ WalletProvider = (*mockWalletProvider)(nil)
var _ ConnectionChecker = (*mockConnectionChecker)(nil)

func TestWalletProviderInterface(t *testing.T) {
	// Verify mock satisfies the interface.
	mock := &mockWalletProvider{
		address: "0xABCD",
		balance: big.NewInt(1000),
	}

	ctx := context.Background()

	addr, err := mock.Address(ctx)
	require.NoError(t, err)
	assert.Equal(t, "0xABCD", addr)

	bal, err := mock.Balance(ctx)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(1000), bal)
}
