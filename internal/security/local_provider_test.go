package security

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalCryptoProvider_Initialize(t *testing.T) {
	t.Parallel()

	p := NewLocalCryptoProvider()

	// Test short passphrase
	err := p.Initialize("short")
	assert.Error(t, err, "expected error for short passphrase")

	// Test valid passphrase
	err = p.Initialize("secure-passphrase-123")
	require.NoError(t, err)
	assert.True(t, p.IsInitialized(), "expected provider to be initialized")
	assert.Len(t, p.Salt(), SaltSize)
}

func TestLocalCryptoProvider_EncryptDecrypt(t *testing.T) {
	t.Parallel()

	p := NewLocalCryptoProvider()
	require.NoError(t, p.Initialize("test-passphrase-123"))

	ctx := context.Background()
	plaintext := []byte("secret message to encrypt")

	// Encrypt
	ciphertext, err := p.Encrypt(ctx, "local", plaintext)
	require.NoError(t, err)
	assert.Greater(t, len(ciphertext), len(plaintext), "ciphertext should be longer than plaintext")

	// Decrypt
	decrypted, err := p.Decrypt(ctx, "local", ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestLocalCryptoProvider_DecryptWithWrongKey(t *testing.T) {
	t.Parallel()

	p1 := NewLocalCryptoProvider()
	require.NoError(t, p1.Initialize("passphrase-one-123"))

	p2 := NewLocalCryptoProvider()
	require.NoError(t, p2.Initialize("passphrase-two-456"))

	ctx := context.Background()
	plaintext := []byte("secret message")

	// Encrypt with p1
	ciphertext, err := p1.Encrypt(ctx, "local", plaintext)
	require.NoError(t, err)

	// Try to decrypt with p2 - should fail
	_, err = p2.Decrypt(ctx, "local", ciphertext)
	assert.Error(t, err, "expected decryption to fail with wrong key")
}

func TestLocalCryptoProvider_Sign(t *testing.T) {
	t.Parallel()

	p := NewLocalCryptoProvider()
	require.NoError(t, p.Initialize("test-passphrase-123"))

	ctx := context.Background()
	payload := []byte("data to sign")

	sig1, err := p.Sign(ctx, "local", payload)
	require.NoError(t, err)

	// Same payload should produce same signature
	sig2, err := p.Sign(ctx, "local", payload)
	require.NoError(t, err)
	assert.Equal(t, sig1, sig2, "signatures should match for same payload")

	// Different payload should produce different signature
	sig3, err := p.Sign(ctx, "local", []byte("different data"))
	require.NoError(t, err)
	assert.NotEqual(t, sig1, sig3, "signatures should differ for different payloads")
}

func TestLocalCryptoProvider_NotInitialized(t *testing.T) {
	t.Parallel()

	p := NewLocalCryptoProvider()
	ctx := context.Background()

	_, err := p.Encrypt(ctx, "local", []byte("test"))
	assert.Error(t, err, "expected error for uninitialized provider")

	_, err = p.Decrypt(ctx, "local", []byte("test"))
	assert.Error(t, err, "expected error for uninitialized provider")

	_, err = p.Sign(ctx, "local", []byte("test"))
	assert.Error(t, err, "expected error for uninitialized provider")
}

func TestLocalCryptoProvider_InitializeWithSalt(t *testing.T) {
	t.Parallel()

	p1 := NewLocalCryptoProvider()
	passphrase := "test-passphrase-123"
	require.NoError(t, p1.Initialize(passphrase))

	salt := p1.Salt()
	ctx := context.Background()
	plaintext := []byte("secret message")

	ciphertext, err := p1.Encrypt(ctx, "local", plaintext)
	require.NoError(t, err)

	// Create new provider with same passphrase and salt
	p2 := NewLocalCryptoProvider()
	require.NoError(t, p2.InitializeWithSalt(passphrase, salt))

	// Should be able to decrypt
	decrypted, err := p2.Decrypt(ctx, "local", ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}
