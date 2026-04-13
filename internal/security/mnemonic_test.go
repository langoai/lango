package security

import (
	"strings"
	"testing"
)

func TestGenerateRecoveryMnemonic_24Words(t *testing.T) {
	mnemonic, err := GenerateRecoveryMnemonic()
	if err != nil {
		t.Fatalf("GenerateRecoveryMnemonic: %v", err)
	}
	words := strings.Fields(mnemonic)
	if len(words) != 24 {
		t.Fatalf("expected 24 words, got %d: %q", len(words), mnemonic)
	}
	if err := ValidateMnemonic(mnemonic); err != nil {
		t.Fatalf("generated mnemonic failed validation: %v", err)
	}
}

func TestGenerateRecoveryMnemonic_Distinct(t *testing.T) {
	m1, err := GenerateRecoveryMnemonic()
	if err != nil {
		t.Fatal(err)
	}
	m2, err := GenerateRecoveryMnemonic()
	if err != nil {
		t.Fatal(err)
	}
	if m1 == m2 {
		t.Fatal("two successive calls returned identical mnemonics")
	}
}

func TestValidateMnemonic_Invalid(t *testing.T) {
	cases := []string{
		"",
		"not enough words",
		"abandon abandon abandon",
		// 24-word mnemonic with invalid checksum
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon",
		// word not in BIP39 wordlist
		"notaword notaword notaword notaword notaword notaword notaword notaword notaword notaword notaword notaword",
	}
	for _, c := range cases {
		if err := ValidateMnemonic(c); err == nil {
			t.Errorf("expected error for %q", c)
		}
	}
}

func TestValidateMnemonic_KnownValid(t *testing.T) {
	// Canonical BIP39 test vector (all-abandon).
	valid := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	if err := ValidateMnemonic(valid); err != nil {
		t.Fatalf("known valid mnemonic rejected: %v", err)
	}
}
