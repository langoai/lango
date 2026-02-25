package handshake

import (
	"sync"
	"time"
)

// NonceSize is the expected byte length of a nonce.
const NonceSize = 32

// NonceCache prevents nonce replay attacks by tracking recently seen nonces.
type NonceCache struct {
	mu     sync.Mutex
	seen   map[[NonceSize]byte]time.Time
	ttl    time.Duration
	ticker *time.Ticker
	stopCh chan struct{}
}

// NewNonceCache creates a new NonceCache with the given TTL.
func NewNonceCache(ttl time.Duration) *NonceCache {
	return &NonceCache{
		seen:   make(map[[NonceSize]byte]time.Time),
		ttl:    ttl,
		stopCh: make(chan struct{}),
	}
}

// CheckAndRecord returns true if the nonce has NOT been seen before (first occurrence).
// Returns false if the nonce was already recorded (replay detected).
// The nonce parameter must be exactly 32 bytes.
func (nc *NonceCache) CheckAndRecord(nonce []byte) bool {
	if len(nonce) != NonceSize {
		return false
	}

	var key [NonceSize]byte
	copy(key[:], nonce)

	nc.mu.Lock()
	defer nc.mu.Unlock()

	if _, exists := nc.seen[key]; exists {
		return false
	}

	nc.seen[key] = time.Now()
	return true
}

// Cleanup removes expired entries older than TTL.
func (nc *NonceCache) Cleanup() {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	for key, recorded := range nc.seen {
		if time.Since(recorded) > nc.ttl {
			delete(nc.seen, key)
		}
	}
}

// Start begins periodic cleanup using a ticker goroutine.
func (nc *NonceCache) Start() {
	nc.ticker = time.NewTicker(nc.ttl / 2)

	go func() {
		for {
			select {
			case <-nc.ticker.C:
				nc.Cleanup()
			case <-nc.stopCh:
				return
			}
		}
	}()
}

// Stop halts the periodic cleanup goroutine.
func (nc *NonceCache) Stop() {
	close(nc.stopCh)
	if nc.ticker != nil {
		nc.ticker.Stop()
	}
}
