package security

import (
	"bytes"
	"errors"
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
