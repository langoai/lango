package security

import (
	"context"
	"sync"
	"time"
)

// KMSHealthChecker implements ConnectionChecker for KMS providers.
// It caches the connection status with a configurable probe interval.
type KMSHealthChecker struct {
	mu            sync.RWMutex
	provider      CryptoProvider
	probeInterval time.Duration
	lastCheck     time.Time
	lastResult    bool
	testKeyID     string
}

// NewKMSHealthChecker creates a health checker that probes the KMS provider
// by attempting a small encrypt/decrypt roundtrip on probeInterval.
func NewKMSHealthChecker(provider CryptoProvider, testKeyID string, probeInterval time.Duration) *KMSHealthChecker {
	if probeInterval <= 0 {
		probeInterval = 30 * time.Second
	}
	return &KMSHealthChecker{
		provider:      provider,
		probeInterval: probeInterval,
		testKeyID:     testKeyID,
	}
}

// IsConnected implements ConnectionChecker. Returns the cached result if fresh,
// otherwise performs a synchronous probe.
func (h *KMSHealthChecker) IsConnected() bool {
	h.mu.RLock()
	if !h.lastCheck.IsZero() && time.Since(h.lastCheck) < h.probeInterval {
		result := h.lastResult
		h.mu.RUnlock()
		return result
	}
	h.mu.RUnlock()

	h.mu.Lock()
	defer h.mu.Unlock()

	// Double-check after acquiring write lock.
	if !h.lastCheck.IsZero() && time.Since(h.lastCheck) < h.probeInterval {
		return h.lastResult
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testData := []byte("kms-health-probe")
	encrypted, err := h.provider.Encrypt(ctx, h.testKeyID, testData)
	if err != nil {
		h.lastResult = false
		h.lastCheck = time.Now()
		return false
	}

	_, err = h.provider.Decrypt(ctx, h.testKeyID, encrypted)
	h.lastResult = err == nil
	h.lastCheck = time.Now()
	return h.lastResult
}
