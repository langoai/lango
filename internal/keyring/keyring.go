package keyring

import "errors"

// Service is the service name used for all keyring operations.
const Service = "lango"

// KeyMasterPassphrase is the keyring key for the master passphrase.
const KeyMasterPassphrase = "master-passphrase"

// ErrNotFound is returned when the requested key does not exist in the keyring.
var ErrNotFound = errors.New("keyring: key not found")

// ErrBiometricNotAvailable is returned when biometric authentication hardware
// (e.g., Touch ID on macOS) is not available on the current system.
var ErrBiometricNotAvailable = errors.New("keyring: biometric authentication not available")

// ErrTPMNotAvailable is returned when no TPM 2.0 device is accessible on the current system.
var ErrTPMNotAvailable = errors.New("keyring: TPM device not available")

// Provider abstracts OS keyring operations for testability.
type Provider interface {
	// Get retrieves a secret for the given service and key.
	// Returns ErrNotFound if the key does not exist.
	Get(service, key string) (string, error)

	// Set stores a secret for the given service and key.
	Set(service, key, value string) error

	// Delete removes a secret for the given service and key.
	// Returns ErrNotFound if the key does not exist.
	Delete(service, key string) error
}

// KeyChecker is an optional interface that secure providers can implement
// to check key existence without triggering authentication (e.g., Touch ID).
// CLI status commands should prefer HasKey over Get to avoid unnecessary
// biometric prompts.
type KeyChecker interface {
	HasKey(service, key string) bool
}

// SecurityTier represents the level of hardware-backed security available
// for keyring storage.
type SecurityTier int

const (
	// TierNone indicates no secure hardware backend; keyfile or interactive prompt only.
	TierNone SecurityTier = iota
	// TierTPM indicates TPM 2.0 sealed storage is available (Linux).
	TierTPM
	// TierBiometric indicates biometric-protected keyring is available (macOS Touch ID).
	TierBiometric
)

// String returns a human-readable label for the security tier.
func (t SecurityTier) String() string {
	switch t {
	case TierBiometric:
		return "biometric"
	case TierTPM:
		return "tpm"
	default:
		return "none"
	}
}

// Status describes the availability of the OS keyring.
type Status struct {
	Available    bool         `json:"available"`
	Backend      string       `json:"backend,omitempty"`
	SecurityTier SecurityTier `json:"securityTier"`
	Error        string       `json:"error,omitempty"`
}
