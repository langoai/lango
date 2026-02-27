//go:build !darwin || !cgo

package keyring

// BiometricProvider is a stub on platforms without macOS Touch ID support.
type BiometricProvider struct{}

// NewBiometricProvider always returns ErrBiometricNotAvailable on non-Darwin
// or non-CGO platforms.
func NewBiometricProvider() (*BiometricProvider, error) {
	return nil, ErrBiometricNotAvailable
}

// Get is a no-op stub that always returns ErrBiometricNotAvailable.
func (*BiometricProvider) Get(string, string) (string, error) {
	return "", ErrBiometricNotAvailable
}

// Set is a no-op stub that always returns ErrBiometricNotAvailable.
func (*BiometricProvider) Set(string, string, string) error {
	return ErrBiometricNotAvailable
}

// Delete is a no-op stub that always returns ErrBiometricNotAvailable.
func (*BiometricProvider) Delete(string, string) error {
	return ErrBiometricNotAvailable
}

// HasKey is a no-op stub that always returns false.
func (*BiometricProvider) HasKey(string, string) bool {
	return false
}
