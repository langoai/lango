package discovery

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

func TestNewAdService(t *testing.T) {
	svc := NewAdService(AdServiceConfig{Logger: testLogger()})
	require.NotNil(t, svc)
	assert.NotNil(t, svc.ads)
}

func TestStoreAd_Valid(t *testing.T) {
	svc := NewAdService(AdServiceConfig{Logger: testLogger()})

	ad := &AgentAd{
		DID:       "did:lango:abc123",
		Name:      "test-agent",
		Tags:      []string{"search", "code"},
		Timestamp: time.Now(),
	}
	err := svc.StoreAd(ad)
	require.NoError(t, err)

	results, err := svc.Discover(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "test-agent", results[0].Name)
}

func TestStoreAd_MissingDID(t *testing.T) {
	svc := NewAdService(AdServiceConfig{Logger: testLogger()})

	ad := &AgentAd{Name: "no-did-agent"}
	err := svc.StoreAd(ad)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing DID")
}

func TestStoreAd_ZKVerification_Pass(t *testing.T) {
	verifier := func(cred *ZKCredential) (bool, error) { return true, nil }
	svc := NewAdService(AdServiceConfig{
		Logger:   testLogger(),
		Verifier: verifier,
	})

	ad := &AgentAd{
		DID:       "did:lango:abc123",
		Name:      "verified-agent",
		Timestamp: time.Now(),
		ZKCredentials: []ZKCredential{
			{
				CapabilityID: "search",
				Proof:        []byte("valid-proof"),
				IssuedAt:     time.Now(),
				ExpiresAt:    time.Now().Add(time.Hour),
			},
		},
	}
	err := svc.StoreAd(ad)
	assert.NoError(t, err)
}

func TestStoreAd_ZKVerification_Fail(t *testing.T) {
	verifier := func(cred *ZKCredential) (bool, error) { return false, nil }
	svc := NewAdService(AdServiceConfig{
		Logger:   testLogger(),
		Verifier: verifier,
	})

	ad := &AgentAd{
		DID:       "did:lango:abc123",
		Name:      "unverified-agent",
		Timestamp: time.Now(),
		ZKCredentials: []ZKCredential{
			{
				CapabilityID: "search",
				Proof:        []byte("invalid-proof"),
				IssuedAt:     time.Now(),
				ExpiresAt:    time.Now().Add(time.Hour),
			},
		},
	}
	err := svc.StoreAd(ad)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ZK credential")
}

func TestStoreAd_ExpiredCredential_Skipped(t *testing.T) {
	called := false
	verifier := func(cred *ZKCredential) (bool, error) {
		called = true
		return true, nil
	}
	svc := NewAdService(AdServiceConfig{
		Logger:   testLogger(),
		Verifier: verifier,
	})

	ad := &AgentAd{
		DID:       "did:lango:abc123",
		Name:      "expired-cred-agent",
		Timestamp: time.Now(),
		ZKCredentials: []ZKCredential{
			{
				CapabilityID: "search",
				Proof:        []byte("proof"),
				IssuedAt:     time.Now().Add(-2 * time.Hour),
				ExpiresAt:    time.Now().Add(-1 * time.Hour), // already expired
			},
		},
	}
	err := svc.StoreAd(ad)
	assert.NoError(t, err)
	assert.False(t, called, "verifier should not be called for expired credentials")
}

func TestStoreAd_TimestampOrdering(t *testing.T) {
	svc := NewAdService(AdServiceConfig{Logger: testLogger()})

	older := &AgentAd{
		DID:       "did:lango:abc123",
		Name:      "old-name",
		Timestamp: time.Now().Add(-time.Hour),
	}
	newer := &AgentAd{
		DID:       "did:lango:abc123",
		Name:      "new-name",
		Timestamp: time.Now(),
	}

	require.NoError(t, svc.StoreAd(newer))
	require.NoError(t, svc.StoreAd(older)) // older should not overwrite

	results, _ := svc.Discover(context.Background(), nil)
	require.Len(t, results, 1)
	assert.Equal(t, "new-name", results[0].Name, "newer ad should be retained")
}

func TestDiscover_EmptyTags_ReturnsAll(t *testing.T) {
	svc := NewAdService(AdServiceConfig{Logger: testLogger()})

	for _, did := range []string{"did:lango:a", "did:lango:b", "did:lango:c"} {
		require.NoError(t, svc.StoreAd(&AgentAd{DID: did, Timestamp: time.Now()}))
	}

	results, err := svc.Discover(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestDiscover_WithTags_Filters(t *testing.T) {
	svc := NewAdService(AdServiceConfig{Logger: testLogger()})

	require.NoError(t, svc.StoreAd(&AgentAd{
		DID: "did:lango:a", Tags: []string{"search", "code"}, Timestamp: time.Now(),
	}))
	require.NoError(t, svc.StoreAd(&AgentAd{
		DID: "did:lango:b", Tags: []string{"translate"}, Timestamp: time.Now(),
	}))

	results, err := svc.Discover(context.Background(), []string{"code"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "did:lango:a", results[0].DID)
}

func TestDiscover_NoMatches(t *testing.T) {
	svc := NewAdService(AdServiceConfig{Logger: testLogger()})

	require.NoError(t, svc.StoreAd(&AgentAd{
		DID: "did:lango:a", Tags: []string{"search"}, Timestamp: time.Now(),
	}))

	results, err := svc.Discover(context.Background(), []string{"nonexistent"})
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestDiscoverByCapability(t *testing.T) {
	svc := NewAdService(AdServiceConfig{Logger: testLogger()})

	require.NoError(t, svc.StoreAd(&AgentAd{
		DID: "did:lango:a", Capabilities: []string{"search", "summarize"}, Timestamp: time.Now(),
	}))
	require.NoError(t, svc.StoreAd(&AgentAd{
		DID: "did:lango:b", Capabilities: []string{"translate"}, Timestamp: time.Now(),
	}))

	matches := svc.DiscoverByCapability(context.Background(), "search")
	assert.Len(t, matches, 1)
	assert.Equal(t, "did:lango:a", matches[0].DID)

	noMatch := svc.DiscoverByCapability(context.Background(), "unknown")
	assert.Empty(t, noMatch)
}

func TestMatchesTags(t *testing.T) {
	tests := []struct {
		name     string
		adTags   []string
		query    []string
		expected bool
	}{
		{"overlap", []string{"a", "b"}, []string{"b", "c"}, true},
		{"no overlap", []string{"a", "b"}, []string{"c", "d"}, false},
		{"empty ad tags", nil, []string{"a"}, false},
		{"empty query", []string{"a"}, nil, false},
		{"both empty", nil, nil, false},
		{"exact match", []string{"search"}, []string{"search"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, matchesTags(tt.adTags, tt.query))
		})
	}
}
