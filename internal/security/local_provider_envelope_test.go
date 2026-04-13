package security

import (
	"bytes"
	"context"
	"testing"
)

func TestLocalProvider_InitializeNewEnvelope(t *testing.T) {
	p := NewLocalCryptoProvider()
	env, err := p.InitializeNewEnvelope(testPassphrase)
	if err != nil {
		t.Fatalf("InitializeNewEnvelope: %v", err)
	}
	if env == nil || env.SlotCount() != 1 {
		t.Fatalf("expected fresh envelope with 1 slot, got %+v", env)
	}
	if !p.IsInitialized() {
		t.Fatal("provider should be initialized")
	}
	if p.IsLegacy() {
		t.Fatal("envelope-mode provider must not be legacy")
	}
	if p.Envelope() == nil {
		t.Fatal("Envelope() should return the stored envelope")
	}

	// Sanity check: encrypt/decrypt via the CryptoProvider interface still works.
	ctx := context.Background()
	plaintext := []byte("hello world")
	ct, err := p.Encrypt(ctx, "local", plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	pt, err := p.Decrypt(ctx, "local", ct)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(pt, plaintext) {
		t.Fatal("round trip mismatch")
	}
}

func TestLocalProvider_InitializeWithEnvelope(t *testing.T) {
	env, mk, err := NewEnvelope(testPassphrase)
	if err != nil {
		t.Fatal(err)
	}
	defer ZeroBytes(mk)

	// Simulate bootstrap: caller holds the unwrapped MK and installs it.
	p := NewLocalCryptoProvider()
	if err := p.InitializeWithEnvelope(mk, env); err != nil {
		t.Fatalf("InitializeWithEnvelope: %v", err)
	}
	if p.IsLegacy() {
		t.Fatal("provider initialized from envelope must not be legacy")
	}

	// Verify the caller can still mutate / zero its own MK buffer without
	// breaking the provider (provider must hold its own copy).
	ZeroBytes(mk)

	ctx := context.Background()
	ct, err := p.Encrypt(ctx, "local", []byte("data"))
	if err != nil {
		t.Fatalf("Encrypt after caller zero: %v", err)
	}
	pt, err := p.Decrypt(ctx, "local", ct)
	if err != nil {
		t.Fatalf("Decrypt after caller zero: %v", err)
	}
	if string(pt) != "data" {
		t.Fatal("mismatch after caller zero")
	}
}

func TestLocalProvider_LegacyPathFlaggedLegacy(t *testing.T) {
	p := NewLocalCryptoProvider()
	if err := p.Initialize(testPassphrase); err != nil {
		t.Fatal(err)
	}
	if !p.IsLegacy() {
		t.Fatal("Initialize() should mark the provider as legacy")
	}
	if p.Envelope() != nil {
		t.Fatal("legacy provider should have nil Envelope()")
	}
}

func TestLocalProvider_Close(t *testing.T) {
	p := NewLocalCryptoProvider()
	if _, err := p.InitializeNewEnvelope(testPassphrase); err != nil {
		t.Fatal(err)
	}
	p.Close()
	if p.IsInitialized() {
		t.Fatal("provider must not be initialized after Close()")
	}
	// Encrypt should now fail.
	ctx := context.Background()
	if _, err := p.Encrypt(ctx, "local", []byte("x")); err == nil {
		t.Fatal("Encrypt should fail after Close()")
	}
}
