package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
)

// EnvelopeVersion is the current MasterKeyEnvelope format version.
const EnvelopeVersion = 1

// KDF and wrap algorithm identifiers.
const (
	KDFAlgPBKDF2SHA256 = "pbkdf2-sha256"
	WrapAlgAES256GCM   = "aes-256-gcm"

	// Domain separation strings.
	domainPassphrase = "passphrase"
	domainMnemonic   = "mnemonic"
	domainDBKey      = "lango-db-encryption"
)

// KEKSlotType identifies the key-derivation source for a KEK slot.
type KEKSlotType string

const (
	KEKSlotPassphrase   KEKSlotType = "passphrase"
	KEKSlotMnemonic     KEKSlotType = "mnemonic"
	KEKSlotRecoveryFile KEKSlotType = "recovery_file" // reserved for follow-up
	KEKSlotHardware     KEKSlotType = "hardware"      // reserved for follow-up
)

// KDFParams carries algorithm-specific KDF parameters.
type KDFParams struct {
	Iterations int `json:"iterations,omitempty"` // PBKDF2
	Memory     int `json:"memory,omitempty"`     // Argon2id (KiB)
	Time       int `json:"time,omitempty"`       // Argon2id
	Threads    int `json:"threads,omitempty"`    // Argon2id
}

// NewDefaultKDFParams returns PBKDF2 parameters matching the current security baseline.
func NewDefaultKDFParams() KDFParams {
	return KDFParams{Iterations: Iterations}
}

// KEKSlot is a single envelope slot that can independently unwrap the Master Key.
type KEKSlot struct {
	ID        string      `json:"id"`
	Type      KEKSlotType `json:"type"`
	KDFAlg    string      `json:"kdf_alg"`
	KDFParams KDFParams   `json:"kdf_params"`
	WrapAlg   string      `json:"wrap_alg"`
	Domain    string      `json:"domain"`
	Salt      []byte      `json:"salt"`
	WrappedMK []byte      `json:"wrapped_mk"`
	Nonce     []byte      `json:"nonce"`
	CreatedAt time.Time   `json:"created_at"`
	Label     string      `json:"label,omitempty"`
}

// MasterKeyEnvelope holds the wrapped Master Key across one or more KEK slots.
// Persisted as JSON at <LangoDir>/envelope.json with 0600 permissions.
type MasterKeyEnvelope struct {
	Version          int       `json:"version"`
	Slots            []KEKSlot `json:"slots"`
	PendingMigration bool      `json:"pending_migration,omitempty"`
	PendingRekey     bool      `json:"pending_rekey,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// GenerateMasterKey returns 32 cryptographically random bytes for use as a Master Key.
func GenerateMasterKey() ([]byte, error) {
	mk := make([]byte, KeySize)
	if _, err := io.ReadFull(rand.Reader, mk); err != nil {
		return nil, fmt.Errorf("generate master key: %w", err)
	}
	return mk, nil
}

// DeriveKEK derives a Key Encryption Key from a secret using the slot's KDF parameters.
// The caller is responsible for providing the correct secret (passphrase or mnemonic).
func DeriveKEK(secret string, slot *KEKSlot) ([]byte, error) {
	if slot == nil {
		return nil, ErrInvalidSlot
	}
	switch slot.KDFAlg {
	case KDFAlgPBKDF2SHA256:
		iters := slot.KDFParams.Iterations
		if iters <= 0 {
			iters = Iterations
		}
		return pbkdf2.Key([]byte(secret), slot.Salt, iters, KeySize, sha256.New), nil
	default:
		return nil, fmt.Errorf("%w: unsupported kdf alg %q", ErrInvalidSlot, slot.KDFAlg)
	}
}

// DeriveDBKey derives the SQLCipher database encryption key from the Master Key.
// Uses HKDF-SHA256 with a domain-separated info label. Returns 32 raw bytes.
func DeriveDBKey(mk []byte) []byte {
	h := hkdf.New(sha256.New, mk, nil, []byte(domainDBKey))
	out := make([]byte, KeySize)
	_, _ = io.ReadFull(h, out)
	return out
}

// DeriveDBKeyHex returns DeriveDBKey(mk) hex-encoded for use with
// SQLCipher `PRAGMA key = "x'<hex>'"` (raw key mode).
func DeriveDBKeyHex(mk []byte) string {
	return hex.EncodeToString(DeriveDBKey(mk))
}

// WrapMasterKey wraps the Master Key with the given KEK using AES-256-GCM.
// Returns the ciphertext and nonce separately.
func WrapMasterKey(mk, kek []byte) (wrapped, nonce []byte, err error) {
	if len(kek) != KeySize {
		return nil, nil, fmt.Errorf("wrap mk: invalid kek size %d", len(kek))
	}
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, nil, fmt.Errorf("wrap mk: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("wrap mk: new gcm: %w", err)
	}
	nonce = make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("wrap mk: generate nonce: %w", err)
	}
	wrapped = gcm.Seal(nil, nonce, mk, nil)
	return wrapped, nonce, nil
}

// UnwrapMasterKey verifies the GCM tag and recovers the Master Key.
// Wraps failures in ErrUnwrapFailed so callers can match with errors.Is.
func UnwrapMasterKey(wrapped, nonce, kek []byte) ([]byte, error) {
	if len(kek) != KeySize {
		return nil, fmt.Errorf("%w: invalid kek size", ErrUnwrapFailed)
	}
	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("%w: invalid nonce size", ErrUnwrapFailed)
	}
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, fmt.Errorf("%w: new cipher", ErrUnwrapFailed)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: new gcm", ErrUnwrapFailed)
	}
	mk, err := gcm.Open(nil, nonce, wrapped, nil)
	if err != nil {
		return nil, fmt.Errorf("%w", ErrUnwrapFailed)
	}
	if len(mk) != KeySize {
		ZeroBytes(mk)
		return nil, fmt.Errorf("%w: unexpected mk size %d", ErrUnwrapFailed, len(mk))
	}
	return mk, nil
}

// NewEnvelope generates a fresh Master Key, wraps it with a passphrase-derived KEK,
// and returns a new envelope alongside the raw MK. Callers MUST zero the returned
// MK with ZeroBytes when done.
func NewEnvelope(passphrase string) (*MasterKeyEnvelope, []byte, error) {
	if len(passphrase) < 8 {
		return nil, nil, fmt.Errorf("passphrase must be at least 8 characters")
	}
	mk, err := GenerateMasterKey()
	if err != nil {
		return nil, nil, err
	}
	now := time.Now().UTC()
	env := &MasterKeyEnvelope{
		Version:   EnvelopeVersion,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := env.AddSlot(KEKSlotPassphrase, "", mk, passphrase, NewDefaultKDFParams()); err != nil {
		ZeroBytes(mk)
		return nil, nil, fmt.Errorf("new envelope: %w", err)
	}
	return env, mk, nil
}

// AddSlot adds a new KEK slot that wraps the provided MK.
// The caller must hold the unwrapped MK in memory. The MK is not modified.
func (e *MasterKeyEnvelope) AddSlot(slotType KEKSlotType, label string, mk []byte, secret string, params KDFParams) error {
	if len(mk) != KeySize {
		return fmt.Errorf("%w: invalid mk size", ErrInvalidSlot)
	}
	if params.Iterations == 0 {
		params = NewDefaultKDFParams()
	}
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return fmt.Errorf("add slot: generate salt: %w", err)
	}
	domain := domainForSlotType(slotType)
	slot := KEKSlot{
		ID:        uuid.NewString(),
		Type:      slotType,
		KDFAlg:    KDFAlgPBKDF2SHA256,
		KDFParams: params,
		WrapAlg:   WrapAlgAES256GCM,
		Domain:    domain,
		Salt:      salt,
		CreatedAt: time.Now().UTC(),
		Label:     label,
	}
	kek, err := DeriveKEK(secret, &slot)
	if err != nil {
		return fmt.Errorf("add slot: derive kek: %w", err)
	}
	defer ZeroBytes(kek)
	wrapped, nonce, err := WrapMasterKey(mk, kek)
	if err != nil {
		return fmt.Errorf("add slot: wrap mk: %w", err)
	}
	slot.WrappedMK = wrapped
	slot.Nonce = nonce
	e.Slots = append(e.Slots, slot)
	e.UpdatedAt = time.Now().UTC()
	return nil
}

// RemoveSlot removes the slot with the given ID. Returns ErrLastSlot if removal
// would leave zero slots.
func (e *MasterKeyEnvelope) RemoveSlot(id string) error {
	if len(e.Slots) <= 1 {
		return ErrLastSlot
	}
	for i, s := range e.Slots {
		if s.ID == id {
			e.Slots = append(e.Slots[:i], e.Slots[i+1:]...)
			e.UpdatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("%w: slot %q not found", ErrInvalidSlot, id)
}

// UnwrapFromPassphrase attempts each passphrase slot in order until one unwraps
// the MK. Returns (mk, slotID, nil) on success or (nil, "", ErrUnwrapFailed).
// Callers MUST ZeroBytes the returned MK when done.
func (e *MasterKeyEnvelope) UnwrapFromPassphrase(passphrase string) ([]byte, string, error) {
	return e.unwrapBySlotType(passphrase, KEKSlotPassphrase)
}

// UnwrapFromMnemonic attempts each mnemonic slot until one unwraps the MK.
func (e *MasterKeyEnvelope) UnwrapFromMnemonic(mnemonic string) ([]byte, string, error) {
	return e.unwrapBySlotType(mnemonic, KEKSlotMnemonic)
}

func (e *MasterKeyEnvelope) unwrapBySlotType(secret string, slotType KEKSlotType) ([]byte, string, error) {
	for i := range e.Slots {
		slot := &e.Slots[i]
		if slot.Type != slotType {
			continue
		}
		kek, err := DeriveKEK(secret, slot)
		if err != nil {
			continue
		}
		mk, err := UnwrapMasterKey(slot.WrappedMK, slot.Nonce, kek)
		ZeroBytes(kek)
		if err == nil {
			return mk, slot.ID, nil
		}
	}
	return nil, "", ErrUnwrapFailed
}

// ChangePassphraseSlot replaces the first passphrase slot (or adds one if missing)
// with a new KEK derived from newPassphrase. The MK is unchanged and all non-passphrase
// slots remain intact.
func (e *MasterKeyEnvelope) ChangePassphraseSlot(mk []byte, newPassphrase string) error {
	if len(newPassphrase) < 8 {
		return fmt.Errorf("passphrase must be at least 8 characters")
	}
	if len(mk) != KeySize {
		return fmt.Errorf("%w: invalid mk size", ErrInvalidSlot)
	}
	// Find and remove existing passphrase slot, if any.
	for i, s := range e.Slots {
		if s.Type == KEKSlotPassphrase {
			e.Slots = append(e.Slots[:i], e.Slots[i+1:]...)
			break
		}
	}
	return e.AddSlot(KEKSlotPassphrase, "", mk, newPassphrase, NewDefaultKDFParams())
}

// HasSlotType reports whether the envelope contains at least one slot of the given type.
func (e *MasterKeyEnvelope) HasSlotType(slotType KEKSlotType) bool {
	for _, s := range e.Slots {
		if s.Type == slotType {
			return true
		}
	}
	return false
}

// SlotCount returns the number of KEK slots.
func (e *MasterKeyEnvelope) SlotCount() int {
	return len(e.Slots)
}

// ZeroBytes overwrites every byte in b with 0x00. Exported so wallet, p2p, and
// other packages can share a single implementation instead of maintaining private copies.
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func domainForSlotType(t KEKSlotType) string {
	switch t {
	case KEKSlotPassphrase:
		return domainPassphrase
	case KEKSlotMnemonic:
		return domainMnemonic
	default:
		return string(t)
	}
}
