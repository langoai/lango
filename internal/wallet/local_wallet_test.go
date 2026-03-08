package wallet

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// newTestSecretsStore creates an in-memory ent-backed SecretsStore for testing.
func newTestSecretsStore(t *testing.T) *security.SecretsStore {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	cryptoProvider := security.NewLocalCryptoProvider()
	require.NoError(t, cryptoProvider.Initialize("test-passphrase-12345"))

	registry := security.NewKeyRegistry(client)
	ctx := context.Background()
	_, err := registry.RegisterKey(ctx, "default", "local", security.KeyTypeEncryption)
	require.NoError(t, err)

	return security.NewSecretsStore(client, registry, cryptoProvider)
}

// storeTestKey generates and stores a private key in the SecretsStore, returning
// the key bytes for verification.
func storeTestKey(t *testing.T, secrets *security.SecretsStore) []byte {
	t.Helper()

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	keyBytes := crypto.FromECDSA(privateKey)
	require.NoError(t, secrets.Store(context.Background(), WalletKeyName, keyBytes))

	// Return a copy so deferred zeroBytes in wallet code won't affect test assertions.
	cp := make([]byte, len(keyBytes))
	copy(cp, keyBytes)
	return cp
}

func TestLocalWallet_Address(t *testing.T) {
	secrets := newTestSecretsStore(t)
	keyBytes := storeTestKey(t, secrets)

	// Derive expected address from the same key.
	expectedKey, err := crypto.ToECDSA(keyBytes)
	require.NoError(t, err)
	expectedAddr := crypto.PubkeyToAddress(expectedKey.PublicKey).Hex()

	w := NewLocalWallet(secrets, "http://localhost:8545", 1)
	ctx := context.Background()

	addr, err := w.Address(ctx)
	require.NoError(t, err)
	assert.Equal(t, expectedAddr, addr)
}

func TestLocalWallet_Address_Deterministic(t *testing.T) {
	secrets := newTestSecretsStore(t)
	storeTestKey(t, secrets)

	w := NewLocalWallet(secrets, "http://localhost:8545", 1)
	ctx := context.Background()

	addr1, err := w.Address(ctx)
	require.NoError(t, err)

	addr2, err := w.Address(ctx)
	require.NoError(t, err)

	assert.Equal(t, addr1, addr2, "Address should be deterministic")
}

func TestLocalWallet_Address_NoKey(t *testing.T) {
	secrets := newTestSecretsStore(t)
	// Do not store any key.

	w := NewLocalWallet(secrets, "http://localhost:8545", 1)
	ctx := context.Background()

	_, err := w.Address(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load wallet key")
}

func TestLocalWallet_SignTransaction(t *testing.T) {
	secrets := newTestSecretsStore(t)
	keyBytes := storeTestKey(t, secrets)

	w := NewLocalWallet(secrets, "http://localhost:8545", 1)
	ctx := context.Background()

	// Use a 32-byte hash as transaction data (typical for signing).
	txHash := crypto.Keccak256([]byte("test transaction"))

	sig, err := w.SignTransaction(ctx, txHash)
	require.NoError(t, err)
	assert.Len(t, sig, 65, "ECDSA signature should be 65 bytes (R + S + V)")

	// Verify the signature can recover the correct public key.
	expectedKey, err := crypto.ToECDSA(keyBytes)
	require.NoError(t, err)
	expectedPubBytes := crypto.CompressPubkey(&expectedKey.PublicKey)

	recoveredPub, err := crypto.Ecrecover(txHash, sig)
	require.NoError(t, err)

	pubKey, err := crypto.UnmarshalPubkey(recoveredPub)
	require.NoError(t, err)
	recoveredCompressed := crypto.CompressPubkey(pubKey)
	assert.Equal(t, expectedPubBytes, recoveredCompressed)
}

func TestLocalWallet_SignTransaction_NoKey(t *testing.T) {
	secrets := newTestSecretsStore(t)

	w := NewLocalWallet(secrets, "http://localhost:8545", 1)
	ctx := context.Background()

	txHash := crypto.Keccak256([]byte("test"))
	_, err := w.SignTransaction(ctx, txHash)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load wallet key")
}

func TestLocalWallet_SignMessage(t *testing.T) {
	secrets := newTestSecretsStore(t)
	keyBytes := storeTestKey(t, secrets)

	w := NewLocalWallet(secrets, "http://localhost:8545", 1)
	ctx := context.Background()

	message := []byte("hello world")

	sig, err := w.SignMessage(ctx, message)
	require.NoError(t, err)
	assert.Len(t, sig, 65)

	// Verify signature. SignMessage internally does crypto.Keccak256(message)
	// before signing.
	hash := crypto.Keccak256(message)
	recoveredPub, err := crypto.Ecrecover(hash, sig)
	require.NoError(t, err)

	expectedKey, err := crypto.ToECDSA(keyBytes)
	require.NoError(t, err)
	expectedPub := crypto.FromECDSAPub(&expectedKey.PublicKey)
	assert.Equal(t, expectedPub, recoveredPub)
}

func TestLocalWallet_SignMessage_NoKey(t *testing.T) {
	secrets := newTestSecretsStore(t)

	w := NewLocalWallet(secrets, "http://localhost:8545", 1)
	ctx := context.Background()

	_, err := w.SignMessage(ctx, []byte("test"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load wallet key")
}

func TestLocalWallet_PublicKey(t *testing.T) {
	secrets := newTestSecretsStore(t)
	keyBytes := storeTestKey(t, secrets)

	w := NewLocalWallet(secrets, "http://localhost:8545", 1)
	ctx := context.Background()

	pubKey, err := w.PublicKey(ctx)
	require.NoError(t, err)
	assert.Len(t, pubKey, 33, "compressed public key should be 33 bytes")

	// Verify it matches the expected compressed public key.
	expectedKey, err := crypto.ToECDSA(keyBytes)
	require.NoError(t, err)
	expectedPub := crypto.CompressPubkey(&expectedKey.PublicKey)
	assert.Equal(t, expectedPub, pubKey)
}

func TestLocalWallet_PublicKey_NoKey(t *testing.T) {
	secrets := newTestSecretsStore(t)

	w := NewLocalWallet(secrets, "http://localhost:8545", 1)
	ctx := context.Background()

	_, err := w.PublicKey(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load wallet key")
}

func TestLocalWallet_PublicKey_Deterministic(t *testing.T) {
	secrets := newTestSecretsStore(t)
	storeTestKey(t, secrets)

	w := NewLocalWallet(secrets, "http://localhost:8545", 1)
	ctx := context.Background()

	pk1, err := w.PublicKey(ctx)
	require.NoError(t, err)

	pk2, err := w.PublicKey(ctx)
	require.NoError(t, err)

	assert.Equal(t, pk1, pk2, "PublicKey should be deterministic")
}

func TestLocalWallet_KeyNameDefault(t *testing.T) {
	secrets := newTestSecretsStore(t)
	w := NewLocalWallet(secrets, "http://localhost:8545", 1)
	assert.Equal(t, WalletKeyName, w.keyName)
}

func TestLocalWallet_SignTransaction_DifferentMessages(t *testing.T) {
	secrets := newTestSecretsStore(t)
	storeTestKey(t, secrets)

	w := NewLocalWallet(secrets, "http://localhost:8545", 1)
	ctx := context.Background()

	hash1 := crypto.Keccak256([]byte("message one"))
	hash2 := crypto.Keccak256([]byte("message two"))

	sig1, err := w.SignTransaction(ctx, hash1)
	require.NoError(t, err)

	sig2, err := w.SignTransaction(ctx, hash2)
	require.NoError(t, err)

	assert.NotEqual(t, sig1, sig2, "different messages should produce different signatures")
}
