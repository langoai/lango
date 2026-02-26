package security

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCryptoProvider implements CryptoProvider for testing the health checker.
type mockHealthCryptoProvider struct {
	encryptErr error
	decryptErr error
}

func (m *mockHealthCryptoProvider) Encrypt(_ context.Context, _ string, data []byte) ([]byte, error) {
	if m.encryptErr != nil {
		return nil, m.encryptErr
	}
	return data, nil
}

func (m *mockHealthCryptoProvider) Decrypt(_ context.Context, _ string, data []byte) ([]byte, error) {
	if m.decryptErr != nil {
		return nil, m.decryptErr
	}
	return data, nil
}

func (m *mockHealthCryptoProvider) Sign(_ context.Context, _ string, _ []byte) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func TestNewKMSHealthChecker_DefaultProbeInterval(t *testing.T) {
	checker := NewKMSHealthChecker(&mockHealthCryptoProvider{}, "test-key", 0)
	require.NotNil(t, checker)
	assert.Equal(t, 30*time.Second, checker.probeInterval)
}

func TestNewKMSHealthChecker_CustomProbeInterval(t *testing.T) {
	checker := NewKMSHealthChecker(&mockHealthCryptoProvider{}, "test-key", 10*time.Second)
	assert.Equal(t, 10*time.Second, checker.probeInterval)
}

func TestKMSHealthChecker_Healthy(t *testing.T) {
	provider := &mockHealthCryptoProvider{}
	checker := NewKMSHealthChecker(provider, "test-key", time.Minute)

	assert.True(t, checker.IsConnected())
}

func TestKMSHealthChecker_Unhealthy_EncryptFails(t *testing.T) {
	provider := &mockHealthCryptoProvider{encryptErr: fmt.Errorf("kms unreachable")}
	checker := NewKMSHealthChecker(provider, "test-key", time.Minute)

	assert.False(t, checker.IsConnected())
}

func TestKMSHealthChecker_Unhealthy_DecryptFails(t *testing.T) {
	provider := &mockHealthCryptoProvider{decryptErr: fmt.Errorf("decrypt failed")}
	checker := NewKMSHealthChecker(provider, "test-key", time.Minute)

	assert.False(t, checker.IsConnected())
}

func TestKMSHealthChecker_CacheFresh(t *testing.T) {
	callCount := 0
	provider := &countingCryptoProvider{count: &callCount}
	checker := NewKMSHealthChecker(provider, "test-key", time.Minute)

	// First call triggers probe.
	result1 := checker.IsConnected()
	assert.True(t, result1)
	assert.Equal(t, 1, callCount)

	// Second call within probe interval returns cached result.
	result2 := checker.IsConnected()
	assert.True(t, result2)
	assert.Equal(t, 1, callCount, "should not re-probe within interval")
}

func TestKMSHealthChecker_CacheExpired(t *testing.T) {
	callCount := 0
	provider := &countingCryptoProvider{count: &callCount}
	checker := NewKMSHealthChecker(provider, "test-key", 10*time.Millisecond)

	checker.IsConnected()
	assert.Equal(t, 1, callCount)

	// Wait for cache to expire.
	time.Sleep(20 * time.Millisecond)

	checker.IsConnected()
	assert.Equal(t, 2, callCount, "should re-probe after interval")
}

// countingCryptoProvider counts encrypt calls for cache testing.
type countingCryptoProvider struct {
	count *int
}

func (c *countingCryptoProvider) Encrypt(_ context.Context, _ string, data []byte) ([]byte, error) {
	*c.count++
	return data, nil
}

func (c *countingCryptoProvider) Decrypt(_ context.Context, _ string, data []byte) ([]byte, error) {
	return data, nil
}

func (c *countingCryptoProvider) Sign(_ context.Context, _ string, _ []byte) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}
