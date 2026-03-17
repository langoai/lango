package wallet

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalUserOpSigner_SignUserOp(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)

	signer := NewLocalUserOpSigner(key)

	tests := []struct {
		give       string
		userOpHash []byte
		entryPoint common.Address
		chainID    *big.Int
	}{
		{
			give:       "base sepolia",
			userOpHash: crypto.Keccak256([]byte("test-op-1")),
			entryPoint: common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032"),
			chainID:    big.NewInt(84532),
		},
		{
			give:       "mainnet",
			userOpHash: crypto.Keccak256([]byte("test-op-2")),
			entryPoint: common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032"),
			chainID:    big.NewInt(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			sig, err := signer.SignUserOp(
				context.Background(),
				tt.userOpHash,
				tt.entryPoint,
				tt.chainID,
			)
			require.NoError(t, err)
			assert.Len(t, sig, 65)
			assert.True(t, sig[64] >= 27, "v value should be >= 27")
		})
	}
}
