package escrow

import (
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/types"
)

func TestResolveAddress(t *testing.T) {
	t.Parallel()

	// Generate a real key for the valid case.
	privKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)

	compressed := ethcrypto.CompressPubkey(&privKey.PublicKey)
	wantAddr := ethcrypto.PubkeyToAddress(privKey.PublicKey)
	validDID := types.DIDPrefix + hex.EncodeToString(compressed)

	tests := []struct {
		give    string
		wantErr bool
		wantDID error
	}{
		{
			give:    validDID,
			wantErr: false,
		},
		{
			give:    "did:other:abc123",
			wantErr: true,
			wantDID: ErrInvalidDID,
		},
		{
			give:    "random-string",
			wantErr: true,
			wantDID: ErrInvalidDID,
		},
		{
			give:    types.DIDPrefix,
			wantErr: true,
			wantDID: ErrInvalidDID,
		},
		{
			give:    types.DIDPrefix + "zzzz-not-hex",
			wantErr: true,
			wantDID: ErrInvalidDID,
		},
		{
			give:    types.DIDPrefix + "deadbeef",
			wantErr: true,
			wantDID: ErrInvalidDID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			addr, err := ResolveAddress(tt.give)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidDID), "expected ErrInvalidDID, got: %v", err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, wantAddr, addr)
		})
	}
}

func TestDefaultAddressResolver_V1(t *testing.T) {
	t.Parallel()
	resolver := NewDefaultAddressResolver(nil) // v1-only mode

	key, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	compressed := ethcrypto.CompressPubkey(&key.PublicKey)
	did := "did:lango:" + hex.EncodeToString(compressed)

	addr, err := resolver.ResolveAddress(did)
	require.NoError(t, err)
	assert.Equal(t, ethcrypto.PubkeyToAddress(key.PublicKey), addr)
}

func TestDefaultAddressResolver_V2_WithBundle(t *testing.T) {
	t.Parallel()

	// Create a settlement key (secp256k1).
	key, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	compressed := ethcrypto.CompressPubkey(&key.PublicKey)
	wantAddr := ethcrypto.PubkeyToAddress(key.PublicKey)

	// Mock settlement key lookup.
	lookup := SettlementKeyLookup(func(did string) ([]byte, error) {
		return compressed, nil
	})

	resolver := NewDefaultAddressResolver(lookup)
	didV2 := "did:lango:v2:abcdef1234567890abcdef1234567890abcdef12"

	addr, err := resolver.ResolveAddress(didV2)
	require.NoError(t, err)
	assert.Equal(t, wantAddr, addr)
}

func TestDefaultAddressResolver_V2_NoBundleResolver(t *testing.T) {
	t.Parallel()
	resolver := NewDefaultAddressResolver(nil)
	didV2 := "did:lango:v2:abcdef1234567890abcdef1234567890abcdef12"

	_, err := resolver.ResolveAddress(didV2)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBundleNotFound)
}

func TestDefaultAddressResolver_V2_BundleNotFound(t *testing.T) {
	t.Parallel()
	lookup := SettlementKeyLookup(func(did string) ([]byte, error) {
		return nil, fmt.Errorf("not found")
	})
	resolver := NewDefaultAddressResolver(lookup)
	didV2 := "did:lango:v2:abcdef1234567890abcdef1234567890abcdef12"

	_, err := resolver.ResolveAddress(didV2)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBundleNotFound)
}
