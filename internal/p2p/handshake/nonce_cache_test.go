package handshake

import (
	"crypto/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeNonce(t *testing.T) []byte {
	t.Helper()
	nonce := make([]byte, NonceSize)
	_, err := rand.Read(nonce)
	require.NoError(t, err)
	return nonce
}

func TestNonceCache_FirstNonce(t *testing.T) {
	nc := NewNonceCache(5 * time.Minute)

	nonce := makeNonce(t)
	ok := nc.CheckAndRecord(nonce)
	assert.True(t, ok, "first occurrence of a nonce should return true")
}

func TestNonceCache_DuplicateNonce(t *testing.T) {
	nc := NewNonceCache(5 * time.Minute)

	nonce := makeNonce(t)
	ok := nc.CheckAndRecord(nonce)
	require.True(t, ok)

	ok = nc.CheckAndRecord(nonce)
	assert.False(t, ok, "duplicate nonce should return false")
}

func TestNonceCache_DifferentNonces(t *testing.T) {
	nc := NewNonceCache(5 * time.Minute)

	nonce1 := makeNonce(t)
	nonce2 := makeNonce(t)

	ok1 := nc.CheckAndRecord(nonce1)
	ok2 := nc.CheckAndRecord(nonce2)

	assert.True(t, ok1, "first nonce should return true")
	assert.True(t, ok2, "second different nonce should return true")
}

func TestNonceCache_InvalidLength(t *testing.T) {
	nc := NewNonceCache(5 * time.Minute)

	tests := []struct {
		give string
		data []byte
	}{
		{give: "nil", data: nil},
		{give: "empty", data: []byte{}},
		{give: "too_short", data: make([]byte, 16)},
		{give: "too_long", data: make([]byte, 64)},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			ok := nc.CheckAndRecord(tt.data)
			assert.False(t, ok, "invalid nonce length should return false")
		})
	}
}

func TestNonceCache_Cleanup(t *testing.T) {
	ttl := 50 * time.Millisecond
	nc := NewNonceCache(ttl)

	nonce := makeNonce(t)
	ok := nc.CheckAndRecord(nonce)
	require.True(t, ok)

	// Wait for TTL to expire.
	time.Sleep(ttl + 10*time.Millisecond)
	nc.Cleanup()

	// After cleanup, the nonce should be accepted again.
	ok = nc.CheckAndRecord(nonce)
	assert.True(t, ok, "nonce should be accepted again after TTL expiry and cleanup")
}

func TestNonceCache_StartStop(t *testing.T) {
	ttl := 50 * time.Millisecond
	nc := NewNonceCache(ttl)

	nc.Start()

	nonce := makeNonce(t)
	ok := nc.CheckAndRecord(nonce)
	require.True(t, ok)

	// Duplicate while running should be rejected.
	ok = nc.CheckAndRecord(nonce)
	assert.False(t, ok)

	// Wait for automatic cleanup via ticker.
	time.Sleep(ttl + 30*time.Millisecond)

	// After automatic cleanup, the nonce should be accepted again.
	ok = nc.CheckAndRecord(nonce)
	assert.True(t, ok, "nonce should be accepted after automatic cleanup")

	nc.Stop()
}

func TestNonceCache_Concurrent(t *testing.T) {
	nc := NewNonceCache(5 * time.Minute)
	nc.Start()
	defer nc.Stop()

	const goroutines = 50
	nonces := make([][]byte, goroutines)
	for i := range nonces {
		nonces[i] = makeNonce(t)
	}

	var wg sync.WaitGroup
	results := make([]bool, goroutines)

	// Each goroutine records a unique nonce.
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx] = nc.CheckAndRecord(nonces[idx])
		}(i)
	}
	wg.Wait()

	for i, ok := range results {
		assert.True(t, ok, "unique nonce %d should succeed", i)
	}

	// Now try duplicates concurrently — exactly one should succeed per nonce.
	shared := makeNonce(t)
	ok := nc.CheckAndRecord(shared)
	require.True(t, ok)

	dupResults := make([]bool, goroutines)
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			dupResults[idx] = nc.CheckAndRecord(shared)
		}(i)
	}
	wg.Wait()

	// All duplicate attempts should return false since the nonce is already recorded.
	for i, ok := range dupResults {
		assert.False(t, ok, "duplicate nonce attempt %d should fail", i)
	}
}
