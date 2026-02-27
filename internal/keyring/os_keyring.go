package keyring

import (
	"errors"
	"runtime"

	gokeyring "github.com/zalando/go-keyring"
)

// OSProvider implements Provider using the OS keyring backend
// (macOS Keychain, Linux secret-service, Windows DPAPI).
type OSProvider struct{}

// NewOSProvider returns a new OSProvider.
func NewOSProvider() *OSProvider {
	return &OSProvider{}
}

// Get retrieves a secret from the OS keyring.
func (p *OSProvider) Get(service, key string) (string, error) {
	val, err := gokeyring.Get(service, key)
	if err != nil {
		if errors.Is(err, gokeyring.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}
	return val, nil
}

// Set stores a secret in the OS keyring.
func (p *OSProvider) Set(service, key, value string) error {
	return gokeyring.Set(service, key, value)
}

// Delete removes a secret from the OS keyring.
func (p *OSProvider) Delete(service, key string) error {
	err := gokeyring.Delete(service, key)
	if err != nil {
		if errors.Is(err, gokeyring.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// probeKey is the key used to test keyring availability.
const probeKey = "__lango_probe__"

// IsAvailable probes the OS keyring with a test write/read/delete cycle.
// Returns a Status describing keyring availability.
func IsAvailable() Status {
	provider := NewOSProvider()

	testValue := "probe"
	if err := provider.Set(Service, probeKey, testValue); err != nil {
		return Status{Available: false, Error: err.Error()}
	}

	got, err := provider.Get(Service, probeKey)
	if err != nil {
		_ = provider.Delete(Service, probeKey)
		return Status{Available: false, Error: err.Error()}
	}
	if got != testValue {
		_ = provider.Delete(Service, probeKey)
		return Status{Available: false, Error: "probe value mismatch"}
	}

	_ = provider.Delete(Service, probeKey)

	// Detect security tier for status reporting.
	_, tier := DetectSecureProvider()

	return Status{
		Available:    true,
		Backend:      backendName(),
		SecurityTier: tier,
	}
}

// backendName returns a human-readable name for the OS keyring backend.
func backendName() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS Keychain"
	case "linux":
		return "secret-service (D-Bus)"
	case "windows":
		return "Windows Credential Manager"
	default:
		return runtime.GOOS
	}
}
