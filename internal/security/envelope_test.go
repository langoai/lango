package security

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

const testPassphrase = "correct-horse-battery-staple"

func TestGenerateMasterKey_Properties(t *testing.T) {
	mk1, err := GenerateMasterKey()
	if err != nil {
		t.Fatalf("GenerateMasterKey: %v", err)
	}
	if len(mk1) != KeySize {
		t.Fatalf("expected %d bytes, got %d", KeySize, len(mk1))
	}
	mk2, err := GenerateMasterKey()
	if err != nil {
		t.Fatalf("GenerateMasterKey: %v", err)
	}
	if bytes.Equal(mk1, mk2) {
		t.Fatal("two successive GenerateMasterKey calls returned identical bytes")
	}
}

func TestWrapUnwrap_RoundTrip(t *testing.T) {
	mk, err := GenerateMasterKey()
	if err != nil {
		t.Fatal(err)
	}
	kek, err := GenerateMasterKey() // reuse 32-byte random as test KEK
	if err != nil {
		t.Fatal(err)
	}
	wrapped, nonce, err := WrapMasterKey(mk, kek)
	if err != nil {
		t.Fatalf("WrapMasterKey: %v", err)
	}
	got, err := UnwrapMasterKey(wrapped, nonce, kek)
	if err != nil {
		t.Fatalf("UnwrapMasterKey: %v", err)
	}
	if !bytes.Equal(got, mk) {
		t.Fatal("round trip mismatch")
	}
}

func TestUnwrap_WrongKEK(t *testing.T) {
	mk, _ := GenerateMasterKey()
	kek, _ := GenerateMasterKey()
	wrongKEK, _ := GenerateMasterKey()
	wrapped, nonce, err := WrapMasterKey(mk, kek)
	if err != nil {
		t.Fatal(err)
	}
	_, err = UnwrapMasterKey(wrapped, nonce, wrongKEK)
	if err == nil {
		t.Fatal("expected error with wrong KEK")
	}
	if !errors.Is(err, ErrUnwrapFailed) {
		t.Fatalf("expected ErrUnwrapFailed, got %v", err)
	}
}

func TestUnwrap_TamperedCiphertext(t *testing.T) {
	mk, _ := GenerateMasterKey()
	kek, _ := GenerateMasterKey()
	wrapped, nonce, err := WrapMasterKey(mk, kek)
	if err != nil {
		t.Fatal(err)
	}
	wrapped[0] ^= 0xFF
	_, err = UnwrapMasterKey(wrapped, nonce, kek)
	if !errors.Is(err, ErrUnwrapFailed) {
		t.Fatalf("expected ErrUnwrapFailed on tampered ciphertext, got %v", err)
	}
}

func TestNewEnvelope_HasPassphraseSlot(t *testing.T) {
	env, mk, err := NewEnvelope(testPassphrase)
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	defer ZeroBytes(mk)
	if env.Version != EnvelopeVersion {
		t.Fatalf("expected version %d, got %d", EnvelopeVersion, env.Version)
	}
	if env.SlotCount() != 1 {
		t.Fatalf("expected 1 slot, got %d", env.SlotCount())
	}
	slot := env.Slots[0]
	if slot.Type != KEKSlotPassphrase {
		t.Fatalf("expected passphrase slot, got %q", slot.Type)
	}
	if slot.KDFAlg != KDFAlgPBKDF2SHA256 {
		t.Fatalf("unexpected KDF alg %q", slot.KDFAlg)
	}
	if slot.WrapAlg != WrapAlgAES256GCM {
		t.Fatalf("unexpected wrap alg %q", slot.WrapAlg)
	}
	if slot.KDFParams.Iterations != Iterations {
		t.Fatalf("expected %d iterations, got %d", Iterations, slot.KDFParams.Iterations)
	}
	if len(slot.Salt) != SaltSize {
		t.Fatalf("expected salt size %d, got %d", SaltSize, len(slot.Salt))
	}
	if slot.ID == "" {
		t.Fatal("slot ID must not be empty")
	}
}

func TestEnvelope_UnwrapFromPassphrase(t *testing.T) {
	env, mk, err := NewEnvelope(testPassphrase)
	if err != nil {
		t.Fatal(err)
	}
	defer ZeroBytes(mk)
	got, _, err := env.UnwrapFromPassphrase(testPassphrase)
	if err != nil {
		t.Fatalf("UnwrapFromPassphrase: %v", err)
	}
	defer ZeroBytes(got)
	if !bytes.Equal(got, mk) {
		t.Fatal("unwrapped MK mismatch")
	}
}

func TestEnvelope_UnwrapWrongPassphrase(t *testing.T) {
	env, mk, err := NewEnvelope(testPassphrase)
	if err != nil {
		t.Fatal(err)
	}
	ZeroBytes(mk)
	_, _, err = env.UnwrapFromPassphrase("wrong-passphrase")
	if !errors.Is(err, ErrUnwrapFailed) {
		t.Fatalf("expected ErrUnwrapFailed, got %v", err)
	}
}

func TestEnvelope_AddMnemonicSlot(t *testing.T) {
	env, mk, err := NewEnvelope(testPassphrase)
	if err != nil {
		t.Fatal(err)
	}
	defer ZeroBytes(mk)
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	if err := env.AddSlot(KEKSlotMnemonic, "recovery", mk, mnemonic, NewDefaultKDFParams()); err != nil {
		t.Fatalf("AddSlot: %v", err)
	}
	if env.SlotCount() != 2 {
		t.Fatalf("expected 2 slots, got %d", env.SlotCount())
	}
	if !env.HasSlotType(KEKSlotMnemonic) {
		t.Fatal("expected HasSlotType(mnemonic) to return true")
	}
	gotPass, _, err := env.UnwrapFromPassphrase(testPassphrase)
	if err != nil {
		t.Fatal(err)
	}
	ZeroBytes(gotPass)
	gotMnem, _, err := env.UnwrapFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("UnwrapFromMnemonic: %v", err)
	}
	defer ZeroBytes(gotMnem)
	if !bytes.Equal(gotMnem, mk) {
		t.Fatal("mnemonic unwrap returned different MK than original")
	}
}

func TestEnvelope_RemoveLastSlotRejected(t *testing.T) {
	env, mk, err := NewEnvelope(testPassphrase)
	if err != nil {
		t.Fatal(err)
	}
	defer ZeroBytes(mk)
	slotID := env.Slots[0].ID
	if err := env.RemoveSlot(slotID); !errors.Is(err, ErrLastSlot) {
		t.Fatalf("expected ErrLastSlot, got %v", err)
	}
	if env.SlotCount() != 1 {
		t.Fatalf("slot count changed after failed removal: %d", env.SlotCount())
	}
}

func TestEnvelope_RemoveSlot_Success(t *testing.T) {
	env, mk, err := NewEnvelope(testPassphrase)
	if err != nil {
		t.Fatal(err)
	}
	defer ZeroBytes(mk)
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	if err := env.AddSlot(KEKSlotMnemonic, "", mk, mnemonic, NewDefaultKDFParams()); err != nil {
		t.Fatal(err)
	}
	mnemSlotID := env.Slots[1].ID
	if err := env.RemoveSlot(mnemSlotID); err != nil {
		t.Fatalf("RemoveSlot: %v", err)
	}
	if env.SlotCount() != 1 {
		t.Fatalf("expected 1 slot after removal, got %d", env.SlotCount())
	}
	if env.HasSlotType(KEKSlotMnemonic) {
		t.Fatal("mnemonic slot should be removed")
	}
}

func TestEnvelope_ChangePassphrase(t *testing.T) {
	env, mk, err := NewEnvelope("old-passphrase-1234")
	if err != nil {
		t.Fatal(err)
	}
	defer ZeroBytes(mk)
	if err := env.ChangePassphraseSlot(mk, "new-passphrase-5678"); err != nil {
		t.Fatalf("ChangePassphraseSlot: %v", err)
	}
	if _, _, err := env.UnwrapFromPassphrase("old-passphrase-1234"); !errors.Is(err, ErrUnwrapFailed) {
		t.Fatal("old passphrase should no longer unwrap")
	}
	got, _, err := env.UnwrapFromPassphrase("new-passphrase-5678")
	if err != nil {
		t.Fatalf("new passphrase unwrap failed: %v", err)
	}
	defer ZeroBytes(got)
	if !bytes.Equal(got, mk) {
		t.Fatal("unwrapped MK differs after passphrase change")
	}
}

func TestDeriveDBKey_Deterministic(t *testing.T) {
	mk, _ := GenerateMasterKey()
	k1 := DeriveDBKey(mk)
	k2 := DeriveDBKey(mk)
	if !bytes.Equal(k1, k2) {
		t.Fatal("DeriveDBKey is not deterministic")
	}
	if len(k1) != KeySize {
		t.Fatalf("expected %d bytes, got %d", KeySize, len(k1))
	}
}

func TestDeriveDBKey_DifferentMKs(t *testing.T) {
	mk1, _ := GenerateMasterKey()
	mk2, _ := GenerateMasterKey()
	if bytes.Equal(DeriveDBKey(mk1), DeriveDBKey(mk2)) {
		t.Fatal("different MKs produced identical DB keys")
	}
}

func TestDeriveDBKeyHex_Length(t *testing.T) {
	mk, _ := GenerateMasterKey()
	hex := DeriveDBKeyHex(mk)
	if len(hex) != KeySize*2 {
		t.Fatalf("expected %d hex chars, got %d", KeySize*2, len(hex))
	}
}

func TestZeroBytes(t *testing.T) {
	b := []byte{1, 2, 3, 4, 5}
	ZeroBytes(b)
	for i, v := range b {
		if v != 0 {
			t.Fatalf("byte %d not zero: %d", i, v)
		}
	}
}

// mockKMSProvider implements CryptoProvider for testing KMS slot operations.
// Uses LocalCryptoProvider internally for actual encrypt/decrypt.
type mockKMSProvider struct {
	local *LocalCryptoProvider
}

func newMockKMSProvider(t *testing.T) *mockKMSProvider {
	t.Helper()
	p := NewLocalCryptoProvider()
	if err := p.Initialize("mock-kms-passphrase-1234"); err != nil {
		t.Fatal(err)
	}
	return &mockKMSProvider{local: p}
}

func (m *mockKMSProvider) Sign(ctx context.Context, keyID string, payload []byte) ([]byte, error) {
	return m.local.Sign(ctx, keyID, payload)
}

func (m *mockKMSProvider) Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	return m.local.Encrypt(ctx, keyID, plaintext)
}

func (m *mockKMSProvider) Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	return m.local.Decrypt(ctx, keyID, ciphertext)
}

// failingKMSProvider always returns an error on Decrypt.
type failingKMSProvider struct{}

func (f *failingKMSProvider) Sign(context.Context, string, []byte) ([]byte, error) {
	return nil, fmt.Errorf("failing mock")
}
func (f *failingKMSProvider) Encrypt(context.Context, string, []byte) ([]byte, error) {
	return nil, fmt.Errorf("failing mock")
}
func (f *failingKMSProvider) Decrypt(context.Context, string, []byte) ([]byte, error) {
	return nil, fmt.Errorf("failing mock: KMS decrypt")
}

func TestEnvelope_AddKMSSlot(t *testing.T) {
	env, mk, err := NewEnvelope(testPassphrase)
	if err != nil {
		t.Fatal(err)
	}
	defer ZeroBytes(mk)

	kms := newMockKMSProvider(t)
	ctx := context.Background()

	err = env.AddKMSSlot(ctx, "test-kms", mk, kms, "aws-kms", "arn:aws:kms:us-east-1:123:key/abc")
	if err != nil {
		t.Fatalf("AddKMSSlot: %v", err)
	}

	if env.SlotCount() != 2 {
		t.Fatalf("expected 2 slots, got %d", env.SlotCount())
	}
	if !env.HasSlotType(KEKSlotHardware) {
		t.Fatal("expected HasSlotType(hardware) to return true")
	}

	kmsSlot := env.Slots[1]
	if kmsSlot.Type != KEKSlotHardware {
		t.Fatalf("expected hardware slot, got %q", kmsSlot.Type)
	}
	if kmsSlot.KDFAlg != KDFAlgNone {
		t.Fatalf("expected KDFAlg %q, got %q", KDFAlgNone, kmsSlot.KDFAlg)
	}
	if kmsSlot.WrapAlg != WrapAlgKMSEnvelope {
		t.Fatalf("expected WrapAlg %q, got %q", WrapAlgKMSEnvelope, kmsSlot.WrapAlg)
	}
	if kmsSlot.KMSProvider != "aws-kms" {
		t.Fatalf("expected KMSProvider %q, got %q", "aws-kms", kmsSlot.KMSProvider)
	}
	if kmsSlot.KMSKeyID != "arn:aws:kms:us-east-1:123:key/abc" {
		t.Fatalf("expected KMSKeyID, got %q", kmsSlot.KMSKeyID)
	}
	if len(kmsSlot.WrappedMK) == 0 {
		t.Fatal("expected non-empty WrappedMK")
	}
	if kmsSlot.ID == "" {
		t.Fatal("expected non-empty slot ID")
	}
}

func TestEnvelope_UnwrapFromKMS_RoundTrip(t *testing.T) {
	env, mk, err := NewEnvelope(testPassphrase)
	if err != nil {
		t.Fatal(err)
	}
	defer ZeroBytes(mk)

	kms := newMockKMSProvider(t)
	ctx := context.Background()
	keyID := "test-key-1"

	err = env.AddKMSSlot(ctx, "kms", mk, kms, "aws-kms", keyID)
	if err != nil {
		t.Fatal(err)
	}

	got, slotID, err := env.UnwrapFromKMS(ctx, kms, "aws-kms", keyID)
	if err != nil {
		t.Fatalf("UnwrapFromKMS: %v", err)
	}
	defer ZeroBytes(got)

	if !bytes.Equal(got, mk) {
		t.Fatal("unwrapped MK mismatch")
	}
	if slotID == "" {
		t.Fatal("expected non-empty slot ID")
	}
}

func TestEnvelope_UnwrapFromKMS_Failure(t *testing.T) {
	env, mk, err := NewEnvelope(testPassphrase)
	if err != nil {
		t.Fatal(err)
	}
	defer ZeroBytes(mk)

	kms := newMockKMSProvider(t)
	ctx := context.Background()

	err = env.AddKMSSlot(ctx, "kms", mk, kms, "aws-kms", "key-1")
	if err != nil {
		t.Fatal(err)
	}

	// Use a failing provider for unwrap.
	_, _, err = env.UnwrapFromKMS(ctx, &failingKMSProvider{}, "aws-kms", "key-1")
	if !errors.Is(err, ErrKMSSlotUnavailable) {
		t.Fatalf("expected ErrKMSSlotUnavailable, got %v", err)
	}
}

func TestEnvelope_UnwrapFromKMS_TierMatching(t *testing.T) {
	env, mk, err := NewEnvelope(testPassphrase)
	if err != nil {
		t.Fatal(err)
	}
	defer ZeroBytes(mk)

	kms := newMockKMSProvider(t)
	ctx := context.Background()

	// Add two KMS slots with same provider but different key IDs.
	err = env.AddKMSSlot(ctx, "kms-a", mk, kms, "aws-kms", "key-A")
	if err != nil {
		t.Fatal(err)
	}
	err = env.AddKMSSlot(ctx, "kms-b", mk, kms, "aws-kms", "key-B")
	if err != nil {
		t.Fatal(err)
	}

	// Tier 1: exact match on key-A.
	got, _, err := env.UnwrapFromKMS(ctx, kms, "aws-kms", "key-A")
	if err != nil {
		t.Fatalf("Tier 1 exact match failed: %v", err)
	}
	if !bytes.Equal(got, mk) {
		t.Fatal("Tier 1 MK mismatch")
	}
	ZeroBytes(got)

	// Tier 2: provider-only fallback (env keyID doesn't match any slot).
	got, _, err = env.UnwrapFromKMS(ctx, kms, "aws-kms", "key-nonexistent")
	if err != nil {
		t.Fatalf("Tier 2 provider-only fallback failed: %v", err)
	}
	if !bytes.Equal(got, mk) {
		t.Fatal("Tier 2 MK mismatch")
	}
	ZeroBytes(got)

	// No match: wrong provider.
	_, _, err = env.UnwrapFromKMS(ctx, kms, "gcp-kms", "key-A")
	if !errors.Is(err, ErrKMSSlotUnavailable) {
		t.Fatalf("expected ErrKMSSlotUnavailable for wrong provider, got %v", err)
	}
}

func TestEnvelope_KMSSlot_JSON_RoundTrip(t *testing.T) {
	env, mk, err := NewEnvelope(testPassphrase)
	if err != nil {
		t.Fatal(err)
	}
	defer ZeroBytes(mk)

	kms := newMockKMSProvider(t)
	ctx := context.Background()

	err = env.AddKMSSlot(ctx, "kms-test", mk, kms, "gcp-kms", "projects/x/locations/y/keyRings/z/cryptoKeys/w")
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var loaded MasterKeyEnvelope
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if loaded.SlotCount() != 2 {
		t.Fatalf("expected 2 slots after JSON roundtrip, got %d", loaded.SlotCount())
	}

	kmsSlot := loaded.Slots[1]
	if kmsSlot.KMSProvider != "gcp-kms" {
		t.Fatalf("KMSProvider lost in JSON: got %q", kmsSlot.KMSProvider)
	}
	if kmsSlot.KMSKeyID != "projects/x/locations/y/keyRings/z/cryptoKeys/w" {
		t.Fatalf("KMSKeyID lost in JSON: got %q", kmsSlot.KMSKeyID)
	}

	// Verify unwrap still works after JSON roundtrip.
	got, _, err := loaded.UnwrapFromKMS(ctx, kms, "gcp-kms", kmsSlot.KMSKeyID)
	if err != nil {
		t.Fatalf("UnwrapFromKMS after JSON roundtrip: %v", err)
	}
	defer ZeroBytes(got)

	if !bytes.Equal(got, mk) {
		t.Fatal("MK mismatch after JSON roundtrip")
	}
}

func TestEnvelope_KMSSlot_BackwardCompat(t *testing.T) {
	// Simulate an old envelope JSON without KMS fields.
	oldJSON := `{
		"version": 1,
		"slots": [{
			"id": "test-id",
			"type": "passphrase",
			"kdf_alg": "pbkdf2-sha256",
			"kdf_params": {"iterations": 100000},
			"wrap_alg": "aes-256-gcm",
			"domain": "passphrase",
			"salt": "AAAAAAAAAAAAAAAAAAAAAA==",
			"wrapped_mk": "AAAAAAAAAAAAAAAAAAAAAA==",
			"nonce": "AAAAAAAAAAAAAAA=",
			"created_at": "2024-01-01T00:00:00Z"
		}],
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-01-01T00:00:00Z"
	}`

	var env MasterKeyEnvelope
	if err := json.Unmarshal([]byte(oldJSON), &env); err != nil {
		t.Fatalf("failed to unmarshal old envelope: %v", err)
	}

	if env.SlotCount() != 1 {
		t.Fatalf("expected 1 slot, got %d", env.SlotCount())
	}

	slot := env.Slots[0]
	if slot.KMSProvider != "" {
		t.Fatalf("expected empty KMSProvider in old envelope, got %q", slot.KMSProvider)
	}
	if slot.KMSKeyID != "" {
		t.Fatalf("expected empty KMSKeyID in old envelope, got %q", slot.KMSKeyID)
	}
	if !env.HasSlotType(KEKSlotPassphrase) {
		t.Fatal("passphrase slot should still be present")
	}
	if env.HasSlotType(KEKSlotHardware) {
		t.Fatal("hardware slot should not be present in old envelope")
	}
}
