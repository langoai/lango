package security

import "errors"

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrNoEncryptionKeys = errors.New("no encryption keys available")
	ErrDecryptionFailed = errors.New("decryption failed")
)
