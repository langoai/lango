package x402

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/coinbase/x402/go/mechanisms/evm"
	evmsigners "github.com/coinbase/x402/go/signers/evm"

	"github.com/langowarny/lango/internal/security"
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
		keyName: "wallet.privatekey",
	}
}

// EvmSigner loads the private key, creates an SDK ClientEvmSigner, then zeros the key material.
func (p *LocalSignerProvider) EvmSigner(ctx context.Context) (evm.ClientEvmSigner, error) {
	keyBytes, err := p.secrets.Get(ctx, p.keyName)
	if err != nil {
		return nil, fmt.Errorf("load wallet key: %w", err)
	}

	keyHex := hex.EncodeToString(keyBytes)
	// Zero raw key bytes immediately.
	for i := range keyBytes {
		keyBytes[i] = 0
	}

	signer, err := evmsigners.NewClientSignerFromPrivateKey(keyHex)
	// Zero hex string by overwriting the backing array.
	keyHexBytes := []byte(keyHex)
	for i := range keyHexBytes {
		keyHexBytes[i] = 0
	}

	if err != nil {
		return nil, fmt.Errorf("create EVM signer: %w", err)
	}

	return signer, nil
}

var _ SignerProvider = (*LocalSignerProvider)(nil)
