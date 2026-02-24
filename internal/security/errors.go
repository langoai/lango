package security

import (
	"errors"
	"fmt"
)

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrNoEncryptionKeys = errors.New("no encryption keys available")
	ErrDecryptionFailed = errors.New("decryption failed")

	// KMS errors
	ErrKMSUnavailable = errors.New("KMS service unavailable")
	ErrKMSAccessDenied = errors.New("KMS access denied")
	ErrKMSKeyDisabled  = errors.New("KMS key is disabled")
	ErrKMSThrottled    = errors.New("KMS request throttled")
	ErrKMSInvalidKey   = errors.New("KMS invalid key")
	ErrPKCS11Module    = errors.New("PKCS#11 module error")
	ErrPKCS11Session   = errors.New("PKCS#11 session error")
)

// KMSError wraps a KMS operation error with context.
type KMSError struct {
	Provider string
	Op       string
	KeyID    string
	Err      error
}

func (e *KMSError) Error() string {
	return fmt.Sprintf("kms %s %s (key=%s): %v", e.Provider, e.Op, e.KeyID, e.Err)
}

func (e *KMSError) Unwrap() error {
	return e.Err
}

// IsTransient reports whether err is a transient KMS error eligible for retry.
func IsTransient(err error) bool {
	return errors.Is(err, ErrKMSUnavailable) || errors.Is(err, ErrKMSThrottled)
}
