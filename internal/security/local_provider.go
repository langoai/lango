package security

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"sync"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// KeySize is the size of AES-256 key in bytes.
	KeySize = 32
	// NonceSize is the size of GCM nonce in bytes.
	NonceSize = 12
	// SaltSize is the size of PBKDF2 salt in bytes.
	SaltSize = 16
	// Iterations is the PBKDF2 iteration count.
	Iterations = 100000
)

// LocalCryptoProvider implements CryptoProvider using local AES-256-GCM encryption.
// Key is derived from a passphrase using PBKDF2.
type LocalCryptoProvider struct {
	mu          sync.RWMutex
	keys        map[string][]byte // keyID -> derived key
	salt        []byte
	initialized bool
}

// NewLocalCryptoProvider creates a new LocalCryptoProvider.
func NewLocalCryptoProvider() *LocalCryptoProvider {
	return &LocalCryptoProvider{
		keys: make(map[string][]byte),
	}
}

// Initialize sets up the provider with a passphrase.
// The passphrase is used to derive an encryption key using PBKDF2.
func (p *LocalCryptoProvider) Initialize(passphrase string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(passphrase) < 8 {
		return fmt.Errorf("passphrase must be at least 8 characters")
	}

	// Generate salt for PBKDF2
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return fmt.Errorf("generate salt: %w", err)
	}

	p.salt = salt
	p.initialized = true

	// Derive and store default key
	key := pbkdf2.Key([]byte(passphrase), salt, Iterations, KeySize, sha256.New)
	p.keys["local"] = key

	return nil
}

// InitializeWithSalt sets up the provider with existing salt (for loading saved state).
func (p *LocalCryptoProvider) InitializeWithSalt(passphrase string, salt []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(passphrase) < 8 {
		return fmt.Errorf("passphrase must be at least 8 characters")
	}

	if len(salt) != SaltSize {
		return fmt.Errorf("invalid salt size")
	}

	p.salt = salt
	p.initialized = true

	// Derive and store default key
	key := pbkdf2.Key([]byte(passphrase), salt, Iterations, KeySize, sha256.New)
	p.keys["local"] = key

	return nil
}

// CalculateChecksum computes the checksum for a given passphrase and salt.
// Uses HMAC-SHA256 with salt as key to avoid length extension attacks.
// NOTE: Changing this algorithm requires migrating existing stored checksums.
func (p *LocalCryptoProvider) CalculateChecksum(passphrase string, salt []byte) []byte {
	mac := hmac.New(sha256.New, salt)
	mac.Write([]byte(passphrase))
	return mac.Sum(nil)
}

// Salt returns the current salt for persistence.
func (p *LocalCryptoProvider) Salt() []byte {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.salt == nil {
		return nil
	}
	// Return copy to prevent modification
	salt := make([]byte, len(p.salt))
	copy(salt, p.salt)
	return salt
}

// IsInitialized returns true if the provider has been initialized.
func (p *LocalCryptoProvider) IsInitialized() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.initialized
}

// Sign generates a signature using HMAC-SHA256 (local signing).
func (p *LocalCryptoProvider) Sign(ctx context.Context, keyID string, payload []byte) ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized {
		return nil, fmt.Errorf("local provider not initialized")
	}

	key, ok := p.keys[keyID]
	if !ok {
		key = p.keys["local"]
	}

	h := hmac.New(sha256.New, key)
	h.Write(payload)
	return h.Sum(nil), nil
}

// Encrypt encrypts data using AES-256-GCM.
func (p *LocalCryptoProvider) Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized {
		return nil, fmt.Errorf("local provider not initialized")
	}

	key, ok := p.keys[keyID]
	if !ok {
		key = p.keys["local"]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	// Prepend nonce to ciphertext
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts data using AES-256-GCM.
func (p *LocalCryptoProvider) Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized {
		return nil, fmt.Errorf("local provider not initialized")
	}

	if len(ciphertext) < NonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	key, ok := p.keys[keyID]
	if !ok {
		key = p.keys["local"]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	nonce := ciphertext[:NonceSize]
	ciphertext = ciphertext[NonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}
