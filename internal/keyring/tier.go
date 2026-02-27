package keyring

// DetectSecureProvider probes available security backends and returns the
// highest-tier provider. Returns (nil, TierNone) if no secure hardware backend
// is available — callers should fall back to keyfile or interactive prompt.
func DetectSecureProvider() (Provider, SecurityTier) {
	// 1. Try biometric (macOS Touch ID).
	if p, err := NewBiometricProvider(); err == nil {
		return p, TierBiometric
	}

	// 2. Try TPM 2.0 (Linux).
	if p, err := NewTPMProvider(); err == nil {
		return p, TierTPM
	}

	// 3. No secure provider available.
	return nil, TierNone
}
