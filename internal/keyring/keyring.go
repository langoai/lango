package keyring

import "errors"

// Service is the service name used for all keyring operations.
const Service = "lango"

// KeyMasterPassphrase is the keyring key for the master passphrase.
const KeyMasterPassphrase = "master-passphrase"

// ErrNotFound is returned when the requested key does not exist in the keyring.
var ErrNotFound = errors.New("keyring: key not found")

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

// Status describes the availability of the OS keyring.
type Status struct {
	Available bool   `json:"available"`
	Backend   string `json:"backend,omitempty"`
	Error     string `json:"error,omitempty"`
}
