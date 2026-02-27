//go:build !linux

package keyring

// TPMProvider is a stub on platforms without TPM 2.0 support.
type TPMProvider struct{}

// NewTPMProvider always returns ErrTPMNotAvailable on non-Linux platforms.
func NewTPMProvider() (*TPMProvider, error) {
	return nil, ErrTPMNotAvailable
}

// Get is a no-op stub that always returns ErrTPMNotAvailable.
func (*TPMProvider) Get(string, string) (string, error) {
	return "", ErrTPMNotAvailable
}

// Set is a no-op stub that always returns ErrTPMNotAvailable.
func (*TPMProvider) Set(string, string, string) error {
	return ErrTPMNotAvailable
}

// Delete is a no-op stub that always returns ErrTPMNotAvailable.
func (*TPMProvider) Delete(string, string) error {
	return ErrTPMNotAvailable
}

// HasKey is a no-op stub that always returns false.
func (*TPMProvider) HasKey(string, string) bool {
	return false
}
