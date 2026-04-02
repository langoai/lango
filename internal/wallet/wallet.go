// Package wallet provides blockchain wallet management for the payment system.
// Private keys never leave the wallet layer — the agent only sees addresses and receipts.
package wallet

import (
	"context"
	"math/big"

	"github.com/langoai/lango/internal/finance"
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

	// PublicKey returns the compressed public key bytes.
	// Used for P2P identity derivation (DID). Private key is never exposed.
	PublicKey(ctx context.Context) ([]byte, error)
}

// ChainID identifies a blockchain network.
type ChainID int64

const (
	ChainEthereumMainnet ChainID = 1
	ChainBase            ChainID = 8453
	ChainBaseSepolia     ChainID = 84532
	ChainSepolia         ChainID = 11155111
)

// CurrencyUSDC is the ticker symbol for the USDC stablecoin used across the
// payment system.
// Deprecated: Use finance.CurrencyUSDC instead.
const CurrencyUSDC = finance.CurrencyUSDC

// WalletInfo holds public wallet metadata.
type WalletInfo struct {
	Address string `json:"address"`
	ChainID int64  `json:"chainId"`
	Network string `json:"network"`
}

// NetworkName returns a human-readable network name for common chain IDs.
func NetworkName(chainID int64) string {
	switch ChainID(chainID) {
	case ChainEthereumMainnet:
		return "Ethereum Mainnet"
	case ChainBase:
		return "Base"
	case ChainBaseSepolia:
		return "Base Sepolia"
	case ChainSepolia:
		return "Sepolia"
	default:
		return "Unknown"
	}
}

// ConnectionChecker determines whether a remote companion is connected.
type ConnectionChecker interface {
	IsConnected() bool
}
