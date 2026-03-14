package x402

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/coinbase/x402/go/mechanisms/evm"
	evmsigners "github.com/coinbase/x402/go/signers/evm"

	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/wallet"
)

// SignerProvider creates an EVM signer for X402 payments.
type SignerProvider interface {
	EvmSigner(ctx context.Context) (evm.ClientEvmSigner, error)
}

// LocalSignerProvider loads the private key from SecretsStore and creates an SDK signer.
type LocalSignerProvider struct {
	secrets *security.SecretsStore
	keyName string
}

// NewLocalSignerProvider creates a signer provider backed by the local secrets store.
func NewLocalSignerProvider(secrets *security.SecretsStore) *LocalSignerProvider {
	return &LocalSignerProvider{
		secrets: secrets,
		keyName: wallet.WalletKeyName,
	}
}

// EvmSigner loads the private key, creates an SDK ClientEvmSigner, then zeros the key material.
func (p *LocalSignerProvider) EvmSigner(ctx context.Context) (evm.ClientEvmSigner, error) {
	keyBytes, err := p.secrets.Get(ctx, p.keyName)
	if err != nil {
		return nil, fmt.Errorf("load wallet key: %w", err)
	}

	// Encode to hex using a mutable []byte buffer so we can zero it after use.
	// Go strings are immutable, so []byte(str) creates a copy — zeroing the
	// copy does not clear the original. Work with []byte throughout.
	keyHexBytes := make([]byte, hex.EncodedLen(len(keyBytes)))
	hex.Encode(keyHexBytes, keyBytes)

	// Zero raw key bytes immediately.
	for i := range keyBytes {
		keyBytes[i] = 0
	}

	signer, err := evmsigners.NewClientSignerFromPrivateKey(string(keyHexBytes))
	// Zero hex buffer.
	for i := range keyHexBytes {
		keyHexBytes[i] = 0
	}

	if err != nil {
		return nil, fmt.Errorf("create EVM signer: %w", err)
	}

	return signer, nil
}

var _ SignerProvider = (*LocalSignerProvider)(nil)
