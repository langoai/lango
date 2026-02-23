package contracts

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupUSDC(t *testing.T) {
	tests := []struct {
		give     int64
		wantAddr string
		wantErr  bool
	}{
		{
			give:     1,
			wantAddr: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
		},
		{
			give:     8453,
			wantAddr: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		},
		{
			give:     84532,
			wantAddr: "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		},
		{
			give:     11155111,
			wantAddr: "0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238",
		},
		{
			give:    999999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(big.NewInt(tt.give).String(), func(t *testing.T) {
			addr, err := LookupUSDC(tt.give)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, common.HexToAddress(tt.wantAddr), addr)
		})
	}
}

func TestIsCanonical(t *testing.T) {
	tests := []struct {
		give       string
		giveChain  int64
		giveAddr   common.Address
		wantResult bool
	}{
		{
			give:       "matching mainnet USDC",
			giveChain:  1,
			giveAddr:   common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
			wantResult: true,
		},
		{
			give:       "matching base USDC",
			giveChain:  8453,
			giveAddr:   common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
			wantResult: true,
		},
		{
			give:       "wrong address on mainnet",
			giveChain:  1,
			giveAddr:   common.HexToAddress("0x0000000000000000000000000000000000000001"),
			wantResult: false,
		},
		{
			give:       "unknown chain",
			giveChain:  42,
			giveAddr:   common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result := IsCanonical(tt.giveChain, tt.giveAddr)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}

// mockCaller implements ContractCaller for testing VerifyOnChain.
type mockCaller struct {
	calls   []ethereum.CallMsg
	results [][]byte
	errs    []error
	idx     int
}

func (m *mockCaller) CallContract(
	_ context.Context,
	msg ethereum.CallMsg,
	_ *big.Int,
) ([]byte, error) {
	i := m.idx
	m.calls = append(m.calls, msg)
	m.idx++
	if i < len(m.errs) && m.errs[i] != nil {
		return nil, m.errs[i]
	}
	if i < len(m.results) {
		return m.results[i], nil
	}
	return nil, nil
}

// encodeABIString produces an ABI-encoded string suitable for contract return.
func encodeABIString(s string) []byte {
	// offset (32 bytes) + length (32 bytes) + padded data (32 bytes per chunk)
	padded := (len(s) + 31) / 32 * 32
	if padded == 0 {
		padded = 32
	}
	buf := make([]byte, 64+padded)
	// Offset = 0x20
	big.NewInt(32).FillBytes(buf[:32])
	// Length
	big.NewInt(int64(len(s))).FillBytes(buf[32:64])
	// Data
	copy(buf[64:], s)
	return buf
}

// encodeABIUint8 produces an ABI-encoded uint256 for a uint8 value.
func encodeABIUint8(v uint8) []byte {
	buf := make([]byte, 32)
	big.NewInt(int64(v)).FillBytes(buf)
	return buf
}

func TestVerifyOnChain(t *testing.T) {
	addr := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")

	t.Run("valid USDC contract", func(t *testing.T) {
		caller := &mockCaller{
			results: [][]byte{
				encodeABIString("USDC"),
				encodeABIUint8(6),
			},
		}
		err := VerifyOnChain(context.Background(), caller, addr)
		require.NoError(t, err)
		assert.Len(t, caller.calls, 2)
	})

	t.Run("wrong symbol", func(t *testing.T) {
		caller := &mockCaller{
			results: [][]byte{
				encodeABIString("USDT"),
				encodeABIUint8(6),
			},
		}
		err := VerifyOnChain(context.Background(), caller, addr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected symbol")
	})

	t.Run("wrong decimals", func(t *testing.T) {
		caller := &mockCaller{
			results: [][]byte{
				encodeABIString("USDC"),
				encodeABIUint8(18),
			},
		}
		err := VerifyOnChain(context.Background(), caller, addr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected decimals")
	})

	t.Run("symbol call error", func(t *testing.T) {
		caller := &mockCaller{
			errs: []error{assert.AnError},
		}
		err := VerifyOnChain(context.Background(), caller, addr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "call symbol()")
	})
}

func TestDecodeABIString(t *testing.T) {
	tests := []struct {
		give    []byte
		want    string
		wantErr bool
	}{
		{
			give: encodeABIString("USDC"),
			want: "USDC",
		},
		{
			give: encodeABIString(""),
			want: "",
		},
		{
			give:    []byte{0x01, 0x02},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		got, err := decodeABIString(tt.give)
		if tt.wantErr {
			require.Error(t, err)
			continue
		}
		require.NoError(t, err)
		assert.Equal(t, tt.want, got)
	}
}

func TestDecodeABIUint8(t *testing.T) {
	tests := []struct {
		give    []byte
		want    uint8
		wantErr bool
	}{
		{
			give: encodeABIUint8(6),
			want: 6,
		},
		{
			give: encodeABIUint8(18),
			want: 18,
		},
		{
			give:    []byte{0x01},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		got, err := decodeABIUint8(tt.give)
		if tt.wantErr {
			require.Error(t, err)
			continue
		}
		require.NoError(t, err)
		assert.Equal(t, tt.want, got)
	}
}
