// Package wallet provides blockchain wallet management for the payment system.
// Private keys never leave the wallet layer — the agent only sees addresses and receipts.
package wallet

import (
	"context"
	"math/big"
)

// WalletProvider abstracts blockchain wallet operations.
// Implementations must ensure private keys are never exposed to callers.
type WalletProvider interface {
	// Address returns the wallet's public address.
	Address(ctx context.Context) (string, error)

	// Balance returns the native token balance in wei.
	Balance(ctx context.Context) (*big.Int, error)

	// SignTransaction signs a raw transaction and returns the signed bytes.
	SignTransaction(ctx context.Context, rawTx []byte) ([]byte, error)

	// SignMessage signs an arbitrary message and returns the signature.
	SignMessage(ctx context.Context, message []byte) ([]byte, error)
}

// WalletInfo holds public wallet metadata.
type WalletInfo struct {
	Address string `json:"address"`
	ChainID int64  `json:"chainId"`
	Network string `json:"network"`
}

// NetworkName returns a human-readable network name for common chain IDs.
func NetworkName(chainID int64) string {
	switch chainID {
	case 1:
		return "Ethereum Mainnet"
	case 8453:
		return "Base"
	case 84532:
		return "Base Sepolia"
	case 11155111:
		return "Sepolia"
	default:
		return "Unknown"
	}
}

// ConnectionChecker determines whether a remote companion is connected.
type ConnectionChecker interface {
	IsConnected() bool
}
