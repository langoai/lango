package security

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockConnectionChecker for testing
type mockConnectionChecker struct {
	connected bool
}

func (m *mockConnectionChecker) IsConnected() bool {
	return m.connected
}

// mockCryptoProvider for testing
type mockCryptoProvider struct {
	signResult    []byte
	encryptResult []byte
	decryptResult []byte
	signErr       error
	encryptErr    error
	decryptErr    error
	called        bool
}

func (m *mockCryptoProvider) Sign(ctx context.Context, keyID string, payload []byte) ([]byte, error) {
	m.called = true
	return m.signResult, m.signErr
}

func (m *mockCryptoProvider) Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	m.called = true
	return m.encryptResult, m.encryptErr
}

func (m *mockCryptoProvider) Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	m.called = true
	return m.decryptResult, m.decryptErr
}

func TestCompositeProvider_UsesPrimaryWhenConnected(t *testing.T) {
	t.Parallel()

	primary := &mockCryptoProvider{encryptResult: []byte("primary-encrypted")}
	fallback := &mockCryptoProvider{encryptResult: []byte("fallback-encrypted")}
	checker := &mockConnectionChecker{connected: true}

	composite := NewCompositeCryptoProvider(primary, fallback, checker)

	result, err := composite.Encrypt(context.Background(), "key1", []byte("data"))
	require.NoError(t, err)
	assert.Equal(t, "primary-encrypted", string(result))
	assert.True(t, primary.called, "primary should have been called")
	assert.False(t, fallback.called, "fallback should not have been called")
	assert.False(t, composite.UsedLocal(), "should not have used local")
}

func TestCompositeProvider_UsesFallbackWhenDisconnected(t *testing.T) {
	t.Parallel()

	primary := &mockCryptoProvider{encryptResult: []byte("primary-encrypted")}
	fallback := &mockCryptoProvider{encryptResult: []byte("fallback-encrypted")}
	checker := &mockConnectionChecker{connected: false}

	composite := NewCompositeCryptoProvider(primary, fallback, checker)

	result, err := composite.Encrypt(context.Background(), "key1", []byte("data"))
	require.NoError(t, err)
	assert.Equal(t, "fallback-encrypted", string(result))
	assert.False(t, primary.called, "primary should not have been called")
	assert.True(t, fallback.called, "fallback should have been called")
	assert.True(t, composite.UsedLocal(), "should have used local")
}

func TestCompositeProvider_ErrorsWhenNoProvider(t *testing.T) {
	t.Parallel()

	checker := &mockConnectionChecker{connected: false}
	composite := NewCompositeCryptoProvider(nil, nil, checker)

	_, err := composite.Encrypt(context.Background(), "key1", []byte("data"))
	assert.Error(t, err, "expected error when no provider available")
}

func TestCompositeProvider_Sign(t *testing.T) {
	t.Parallel()

	primary := &mockCryptoProvider{signResult: []byte("primary-sig")}
	fallback := &mockCryptoProvider{signResult: []byte("fallback-sig")}
	checker := &mockConnectionChecker{connected: true}

	composite := NewCompositeCryptoProvider(primary, fallback, checker)

	result, err := composite.Sign(context.Background(), "key1", []byte("data"))
	require.NoError(t, err)
	assert.Equal(t, "primary-sig", string(result))
}

func TestCompositeProvider_Decrypt(t *testing.T) {
	t.Parallel()

	fallback := &mockCryptoProvider{decryptResult: []byte("decrypted-data")}
	checker := &mockConnectionChecker{connected: false}

	composite := NewCompositeCryptoProvider(nil, fallback, checker)

	result, err := composite.Decrypt(context.Background(), "key1", []byte("encrypted"))
	require.NoError(t, err)
	assert.Equal(t, "decrypted-data", string(result))
}
