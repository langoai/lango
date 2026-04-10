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
//
// Two initialization modes are supported:
//
//   - Envelope mode (preferred): the Master Key is unwrapped from a
//     MasterKeyEnvelope slot and installed as keys["local"]. All data encryption
//     uses the MK directly. This mode supports key rotation (change-passphrase),
//     recovery mnemonics, and DB key derivation via DeriveDBKey.
//
//   - Legacy mode: the key is derived directly from a passphrase via PBKDF2 and
//     stored in keys["local"]. This mode is kept for backward compatibility and
//     is replaced by envelope mode during migration.
//
// In both modes, Encrypt/Decrypt/Sign use the same keys["local"] lookup, so
// consumers (SecretsStore, ConfigStore, tools) are unaware of the underlying
// source of the key.
type LocalCryptoProvider struct {
	mu          sync.RWMutex
	keys        map[string][]byte // keyID -> key material
	salt        []byte            // legacy PBKDF2 salt (empty in envelope mode)
	masterKey   []byte             // unwrapped MK (envelope mode)
	envelope    *MasterKeyEnvelope // envelope reference (envelope mode)
	initialized bool
	legacy      bool
}

// NewLocalCryptoProvider creates a new LocalCryptoProvider.
func NewLocalCryptoProvider() *LocalCryptoProvider {
	return &LocalCryptoProvider{
		keys: make(map[string][]byte),
	}
}

// Initialize sets up the provider with a passphrase using the legacy direct-key model.
// The passphrase is used to derive an encryption key via PBKDF2.
// Marks the provider as legacy — envelope-based bootstrap should use
// InitializeNewEnvelope or InitializeWithEnvelope instead.
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
	p.legacy = true

	// Derive and store default key
	key := pbkdf2.Key([]byte(passphrase), salt, Iterations, KeySize, sha256.New)
	p.keys["local"] = key

	return nil
}

// InitializeWithSalt sets up the provider with existing salt (legacy direct-key model).
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
	p.legacy = true

	// Derive and store default key
	key := pbkdf2.Key([]byte(passphrase), salt, Iterations, KeySize, sha256.New)
	p.keys["local"] = key

	return nil
}

// InitializeWithEnvelope installs an already-unwrapped Master Key as the local key.
// The provider takes ownership of the MK bytes and zeroes them in Close().
// The envelope reference is kept for diagnostics (SlotCount, recovery status).
func (p *LocalCryptoProvider) InitializeWithEnvelope(mk []byte, envelope *MasterKeyEnvelope) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(mk) != KeySize {
		return fmt.Errorf("initialize with envelope: invalid mk size %d", len(mk))
	}

	// Copy the MK so the caller can zero its own buffer independently.
	stored := make([]byte, KeySize)
	copy(stored, mk)

	p.masterKey = stored
	p.envelope = envelope
	p.keys["local"] = stored
	p.initialized = true
	p.legacy = false
	return nil
}

// InitializeNewEnvelope generates a fresh envelope from a passphrase and
// installs the Master Key as the local key. Returns the envelope so the
// caller can persist it via StoreEnvelopeFile.
func (p *LocalCryptoProvider) InitializeNewEnvelope(passphrase string) (*MasterKeyEnvelope, error) {
	env, mk, err := NewEnvelope(passphrase)
	if err != nil {
		return nil, err
	}
	if err := p.InitializeWithEnvelope(mk, env); err != nil {
		ZeroBytes(mk)
		return nil, err
	}
	ZeroBytes(mk)
	return env, nil
}

// Envelope returns the currently installed envelope, or nil in legacy mode.
// Callers MUST NOT mutate the returned envelope without acquiring exclusive access.
func (p *LocalCryptoProvider) Envelope() *MasterKeyEnvelope {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.envelope
}

// IsLegacy reports whether the provider was initialized via the legacy
// direct-passphrase-derived-key path.
func (p *LocalCryptoProvider) IsLegacy() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.legacy
}

// Close zeroes all key material held by the provider. After Close, the
// provider is unusable and must not be referenced.
func (p *LocalCryptoProvider) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for k, v := range p.keys {
		ZeroBytes(v)
		delete(p.keys, k)
	}
	if p.masterKey != nil {
		ZeroBytes(p.masterKey)
		p.masterKey = nil
	}
	p.envelope = nil
	p.initialized = false
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
