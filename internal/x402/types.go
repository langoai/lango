// Package x402 implements the X402 HTTP payment protocol for blockchain micropayments.
// It uses the Coinbase X402 V2 Go SDK for automatic payment handling when agents
// encounter HTTP 402 responses from X402-enabled services.
package x402

import "fmt"

// Config holds X402 interceptor configuration.
type Config struct {
	// Enabled controls whether X402 automatic payment is active.
	Enabled bool

	// ChainID is the numeric EVM chain ID (e.g. 84532 for Base Sepolia).
	ChainID int64

	// MaxAutoPayAmount is the maximum USDC amount for automatic payments (e.g. "1.00").
	MaxAutoPayAmount string
}

// CAIP2Network converts a numeric chain ID to a CAIP-2 network identifier.
// Example: 84532 -> "eip155:84532"
func CAIP2Network(chainID int64) string {
	return fmt.Sprintf("eip155:%d", chainID)
}
