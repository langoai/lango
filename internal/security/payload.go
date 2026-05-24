package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

const (
	// PayloadKeyVersionV1 is the fixed payload-protection key version for v1.
	PayloadKeyVersionV1 = 1
	domainPayloadKeyV1  = "lango-payload-protection-v1"
)

// PayloadProtector encrypts and decrypts sensitive payloads.
type PayloadProtector interface {
	EncryptPayload(plaintext []byte) (ciphertext []byte, nonce []byte, keyVersion int, err error)
	DecryptPayload(ciphertext []byte, nonce []byte, keyVersion int) ([]byte, error)
}

// DerivePayloadKey derives a 32-byte AEAD key for payload protection.
func DerivePayloadKey(mk []byte) []byte {
	h := hkdf.New(sha256.New, mk, nil, []byte(domainPayloadKeyV1))
	out := make([]byte, KeySize)
	_, _ = io.ReadFull(h, out)
	return out
}

// EncryptPayloadWithKey encrypts a payload using AES-256-GCM and returns the
// detached ciphertext and nonce.
func EncryptPayloadWithKey(key, plaintext []byte) ([]byte, []byte, error) {
	if len(key) != KeySize {
		return nil, nil, fmt.Errorf("encrypt payload: invalid key size %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("encrypt payload: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("encrypt payload: new gcm: %w", err)
	}
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("encrypt payload: generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// DecryptPayloadWithKey decrypts a detached ciphertext/nonce pair.
func DecryptPayloadWithKey(key, ciphertext, nonce []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("decrypt payload: invalid key size %d", len(key))
	}
	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("decrypt payload: invalid nonce size %d", len(nonce))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("decrypt payload: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("decrypt payload: new gcm: %w", err)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt payload: %w", err)
	}
	return plaintext, nil
}
