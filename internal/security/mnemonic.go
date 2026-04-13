package security

import (
	"fmt"

	"github.com/tyler-smith/go-bip39"
)

// RecoveryMnemonicBits is the BIP39 entropy size that yields a 24-word mnemonic.
const RecoveryMnemonicBits = 256

// GenerateRecoveryMnemonic returns a fresh 24-word BIP39 mnemonic.
// The returned string contains all information needed to re-derive a KEK;
// callers MUST display it to the user and ensure it is recorded before persisting
// the associated envelope slot.
func GenerateRecoveryMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(RecoveryMnemonicBits)
	if err != nil {
		return "", fmt.Errorf("generate mnemonic entropy: %w", err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("generate mnemonic: %w", err)
	}
	return mnemonic, nil
}

// ValidateMnemonic returns nil if the mnemonic is valid BIP39.
// It does NOT verify that the mnemonic matches any envelope slot.
func ValidateMnemonic(mnemonic string) error {
	if !bip39.IsMnemonicValid(mnemonic) {
		return fmt.Errorf("invalid BIP39 mnemonic")
	}
	return nil
}
