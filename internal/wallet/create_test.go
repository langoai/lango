package wallet

import (
	"context"
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"

	"github.com/langoai/lango/internal/security"
)

// Uses newTestSecretsStore from local_wallet_test.go.

func TestCreateWallet_Success(t *testing.T) {
	secrets := newTestSecretsStore(t)
	ctx := context.Background()

	addr, err := CreateWallet(ctx, secrets)
	require.NoError(t, err)
	assert.NotEmpty(t, addr)
	// Address should be a valid hex address (0x + 40 hex chars).
	assert.Regexp(t, `^0x[0-9a-fA-F]{40}$`, addr)
}

func TestCreateWallet_StoresRecoverableKey(t *testing.T) {
	secrets := newTestSecretsStore(t)
	ctx := context.Background()

	addr, err := CreateWallet(ctx, secrets)
	require.NoError(t, err)

	// Retrieve the stored key and verify it derives the same address.
	keyBytes, err := secrets.Get(ctx, WalletKeyName)
	require.NoError(t, err)
	defer security.ZeroBytes(keyBytes)

	privateKey, err := crypto.ToECDSA(keyBytes)
	require.NoError(t, err)

	derivedAddr := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	assert.Equal(t, addr, derivedAddr)
}

func TestCreateWallet_AlreadyExists(t *testing.T) {
	secrets := newTestSecretsStore(t)
	ctx := context.Background()

	// Create first wallet.
	firstAddr, err := CreateWallet(ctx, secrets)
	require.NoError(t, err)

	// Second creation attempt should return ErrWalletExists with the existing address.
	secondAddr, err := CreateWallet(ctx, secrets)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrWalletExists))
	assert.Equal(t, firstAddr, secondAddr)
}

func TestCreateWallet_GeneratesUniqueKeys(t *testing.T) {
	// Each call to CreateWallet on a fresh store should produce a different address.
	secrets1 := newTestSecretsStore(t)
	secrets2 := newTestSecretsStore(t)
	ctx := context.Background()

	addr1, err := CreateWallet(ctx, secrets1)
	require.NoError(t, err)

	addr2, err := CreateWallet(ctx, secrets2)
	require.NoError(t, err)

	// Extremely unlikely for two random keys to collide.
	assert.NotEqual(t, addr1, addr2)
}

func TestWalletKeyName_Constant(t *testing.T) {
	assert.Equal(t, "wallet.privatekey", WalletKeyName)
}

func TestErrWalletExists_Sentinel(t *testing.T) {
	assert.Equal(t, "wallet already exists", ErrWalletExists.Error())
}
