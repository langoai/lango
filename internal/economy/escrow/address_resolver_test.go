package escrow

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/types"
)

func TestResolveAddress(t *testing.T) {
	t.Parallel()

	// Generate a real key for the valid case.
	privKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	compressed := crypto.CompressPubkey(&privKey.PublicKey)
	wantAddr := crypto.PubkeyToAddress(privKey.PublicKey)
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
