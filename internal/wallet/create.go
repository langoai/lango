package wallet

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/langowarny/lango/internal/security"
)

const walletKeyName = "wallet.privatekey"

// ErrWalletExists is returned when attempting to create a wallet that already exists.
var ErrWalletExists = errors.New("wallet already exists")

// CreateWallet generates a new ECDSA private key, stores it encrypted in
// SecretsStore, and returns the derived public address. If a wallet already
// exists, it returns ErrWalletExists along with the existing address.
func CreateWallet(ctx context.Context, secrets *security.SecretsStore) (string, error) {
	// Check if wallet already exists
	existing, err := secrets.Get(ctx, walletKeyName)
	if err == nil {
		defer zeroBytes(existing)

		key, parseErr := crypto.ToECDSA(existing)
		if parseErr != nil {
			return "", fmt.Errorf("parse existing wallet key: %w", parseErr)
		}
		addr := crypto.PubkeyToAddress(key.PublicKey).Hex()
		return addr, ErrWalletExists
	}

	// Generate new ECDSA key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return "", fmt.Errorf("generate key: %w", err)
	}

	keyBytes := crypto.FromECDSA(privateKey)
	defer zeroBytes(keyBytes)

	// Store encrypted in SecretsStore
	if err := secrets.Store(ctx, walletKeyName, keyBytes); err != nil {
		return "", fmt.Errorf("store wallet key: %w", err)
	}

	addr := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	return addr, nil
}
