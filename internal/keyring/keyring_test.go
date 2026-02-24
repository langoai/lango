package keyring

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider is an in-memory Provider for testing.
type mockProvider struct {
	store map[string]string
}

func newMockProvider() *mockProvider {
	return &mockProvider{store: make(map[string]string)}
}

func (m *mockProvider) Get(service, key string) (string, error) {
	k := service + "/" + key
	v, ok := m.store[k]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (m *mockProvider) Set(service, key, value string) error {
	k := service + "/" + key
	m.store[k] = value
	return nil
}

func (m *mockProvider) Delete(service, key string) error {
	k := service + "/" + key
	if _, ok := m.store[k]; !ok {
		return ErrNotFound
	}
	delete(m.store, k)
	return nil
}

func TestMockProvider_SetGetDelete(t *testing.T) {
	p := newMockProvider()

	// Get non-existent key returns ErrNotFound.
	_, err := p.Get(Service, KeyMasterPassphrase)
	assert.ErrorIs(t, err, ErrNotFound)

	// Set and Get.
	require.NoError(t, p.Set(Service, KeyMasterPassphrase, "my-secret"))
	got, err := p.Get(Service, KeyMasterPassphrase)
	require.NoError(t, err)
	assert.Equal(t, "my-secret", got)

	// Overwrite.
	require.NoError(t, p.Set(Service, KeyMasterPassphrase, "updated"))
	got, err = p.Get(Service, KeyMasterPassphrase)
	require.NoError(t, err)
	assert.Equal(t, "updated", got)

	// Delete.
	require.NoError(t, p.Delete(Service, KeyMasterPassphrase))
	_, err = p.Get(Service, KeyMasterPassphrase)
	assert.ErrorIs(t, err, ErrNotFound)

	// Delete non-existent key returns ErrNotFound.
	err = p.Delete(Service, KeyMasterPassphrase)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMockProvider_MultipleKeys(t *testing.T) {
	p := newMockProvider()

	require.NoError(t, p.Set(Service, "key-a", "val-a"))
	require.NoError(t, p.Set(Service, "key-b", "val-b"))

	a, err := p.Get(Service, "key-a")
	require.NoError(t, err)
	assert.Equal(t, "val-a", a)

	b, err := p.Get(Service, "key-b")
	require.NoError(t, err)
	assert.Equal(t, "val-b", b)

	// Delete one, other still exists.
	require.NoError(t, p.Delete(Service, "key-a"))
	_, err = p.Get(Service, "key-a")
	assert.ErrorIs(t, err, ErrNotFound)

	b, err = p.Get(Service, "key-b")
	require.NoError(t, err)
	assert.Equal(t, "val-b", b)
}

func TestProviderInterfaceCompliance(t *testing.T) {
	// Compile-time check that mockProvider satisfies Provider.
	var _ Provider = (*mockProvider)(nil)
	var _ Provider = (*OSProvider)(nil)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "lango", Service)
	assert.Equal(t, "master-passphrase", KeyMasterPassphrase)
}
