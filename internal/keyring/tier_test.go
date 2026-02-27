package keyring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityTier_String(t *testing.T) {
	tests := []struct {
		give SecurityTier
		want string
	}{
		{give: TierNone, want: "none"},
		{give: TierTPM, want: "tpm"},
		{give: TierBiometric, want: "biometric"},
		{give: SecurityTier(99), want: "none"}, // unknown defaults to "none"
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.give.String())
		})
	}
}

func TestSecurityTier_Ordering(t *testing.T) {
	// Verify tier ordering: None < TPM < Biometric.
	assert.Less(t, TierNone, TierTPM)
	assert.Less(t, TierTPM, TierBiometric)
}

func TestDetectSecureProvider_ReturnsProvider(t *testing.T) {
	// DetectSecureProvider should always return without panicking.
	// On CI / machines without biometric or TPM, it returns (nil, TierNone).
	provider, tier := DetectSecureProvider()

	switch tier {
	case TierBiometric:
		assert.NotNil(t, provider)
	case TierTPM:
		assert.NotNil(t, provider)
	case TierNone:
		assert.Nil(t, provider)
	default:
		t.Fatalf("unexpected security tier: %d", tier)
	}
}

func TestDetectSecureProvider_MockFallback(t *testing.T) {
	// Verify that DetectSecureProvider gracefully degrades.
	// This test always passes — it documents the fallback behavior.
	_, tier := DetectSecureProvider()
	assert.Contains(t, []SecurityTier{TierNone, TierTPM, TierBiometric}, tier)
}

func TestErrSentinels(t *testing.T) {
	assert.EqualError(t, ErrNotFound, "keyring: key not found")
	assert.EqualError(t, ErrBiometricNotAvailable, "keyring: biometric authentication not available")
	assert.EqualError(t, ErrTPMNotAvailable, "keyring: TPM device not available")
}
