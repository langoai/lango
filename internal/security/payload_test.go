package security

import (
	"bytes"
	"testing"
)

func TestPayloadKeyDerivationAndRoundTrip(t *testing.T) {
	mk := bytes.Repeat([]byte{0x42}, KeySize)
	key1 := DerivePayloadKey(mk)
	key2 := DerivePayloadKey(mk)
	if !bytes.Equal(key1, key2) {
		t.Fatal("payload key derivation is not deterministic")
	}

	plaintext := []byte("top secret payload")
	ciphertext, nonce, err := EncryptPayloadWithKey(key1, plaintext)
	if err != nil {
		t.Fatalf("EncryptPayloadWithKey: %v", err)
	}
	got, err := DecryptPayloadWithKey(key2, ciphertext, nonce)
	if err != nil {
		t.Fatalf("DecryptPayloadWithKey: %v", err)
	}
	if string(got) != string(plaintext) {
		t.Fatalf("want %q, got %q", plaintext, got)
	}
}

func TestPayloadDecryptTamperFails(t *testing.T) {
	mk := bytes.Repeat([]byte{0x11}, KeySize)
	key := DerivePayloadKey(mk)
	ciphertext, nonce, err := EncryptPayloadWithKey(key, []byte("tamper target"))
	if err != nil {
		t.Fatalf("EncryptPayloadWithKey: %v", err)
	}
	ciphertext[0] ^= 0xFF
	if _, err := DecryptPayloadWithKey(key, ciphertext, nonce); err == nil {
		t.Fatal("expected tampered ciphertext to fail decryption")
	}
}
