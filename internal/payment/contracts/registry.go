// Package contracts provides canonical USDC contract addresses and on-chain
// verification utilities for supported EVM chains.
package contracts

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// CanonicalUSDC maps chain IDs to their official USDC contract addresses.
var CanonicalUSDC = map[int64]string{
	1:        "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", // Ethereum Mainnet
	8453:     "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", // Base
	84532:    "0x036CbD53842c5426634e7929541eC2318f3dCF7e", // Base Sepolia
	11155111: "0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238", // Sepolia
}

// symbolSelector is the function selector for symbol().
var symbolSelector = crypto.Keccak256([]byte("symbol()"))[:4]

// decimalsSelector is the function selector for decimals().
var decimalsSelector = crypto.Keccak256([]byte("decimals()"))[:4]

// LookupUSDC returns the canonical USDC contract address for the given chain.
func LookupUSDC(chainID int64) (common.Address, error) {
	addr, ok := CanonicalUSDC[chainID]
	if !ok {
		return common.Address{}, fmt.Errorf("unknown chain ID %d", chainID)
	}
	return common.HexToAddress(addr), nil
}

// IsCanonical checks whether the given address matches the canonical USDC
// contract for the specified chain.
func IsCanonical(chainID int64, addr common.Address) bool {
	canonical, ok := CanonicalUSDC[chainID]
	if !ok {
		return false
	}
	return common.HexToAddress(canonical) == addr
}

// ContractCaller abstracts the eth_call method for on-chain verification.
type ContractCaller interface {
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
}

// Ensure *ethclient.Client satisfies ContractCaller at compile time.
var _ ContractCaller = (*ethclient.Client)(nil)

// VerifyOnChain calls symbol() and decimals() on the contract to confirm it is
// a USDC token (symbol == "USDC", decimals == 6).
func VerifyOnChain(
	ctx context.Context,
	caller ContractCaller,
	addr common.Address,
) error {
	// Call symbol()
	symbolResult, err := caller.CallContract(ctx, ethereum.CallMsg{
		To:   &addr,
		Data: symbolSelector,
	}, nil)
	if err != nil {
		return fmt.Errorf("call symbol(): %w", err)
	}

	symbol, err := decodeABIString(symbolResult)
	if err != nil {
		return fmt.Errorf("decode symbol: %w", err)
	}
	if symbol != "USDC" {
		return fmt.Errorf("unexpected symbol %q, want \"USDC\"", symbol)
	}

	// Call decimals()
	decimalsResult, err := caller.CallContract(ctx, ethereum.CallMsg{
		To:   &addr,
		Data: decimalsSelector,
	}, nil)
	if err != nil {
		return fmt.Errorf("call decimals(): %w", err)
	}

	decimals, err := decodeABIUint8(decimalsResult)
	if err != nil {
		return fmt.Errorf("decode decimals: %w", err)
	}
	if decimals != 6 {
		return fmt.Errorf("unexpected decimals %d, want 6", decimals)
	}

	return nil
}

// decodeABIString decodes an ABI-encoded string return value.
// ABI layout: [32-byte offset][32-byte length][padded data...]
func decodeABIString(data []byte) (string, error) {
	if len(data) < 64 {
		return "", fmt.Errorf("response too short: %d bytes", len(data))
	}

	// First 32 bytes: offset to string data (should be 0x20 = 32)
	offset := new(big.Int).SetBytes(data[:32]).Int64()
	if offset < 0 || offset+32 > int64(len(data)) {
		return "", fmt.Errorf("invalid string offset: %d", offset)
	}

	// Next 32 bytes at offset: string length
	strLen := new(big.Int).SetBytes(data[offset : offset+32]).Int64()
	if strLen < 0 || offset+32+strLen > int64(len(data)) {
		return "", fmt.Errorf("invalid string length: %d", strLen)
	}

	raw := string(data[offset+32 : offset+32+strLen])
	return strings.TrimRight(raw, "\x00"), nil
}

// decodeABIUint8 decodes a uint8 (encoded as uint256) return value.
func decodeABIUint8(data []byte) (uint8, error) {
	if len(data) < 32 {
		return 0, fmt.Errorf("response too short: %d bytes", len(data))
	}
	val := new(big.Int).SetBytes(data[:32])
	if !val.IsInt64() || val.Int64() > 255 || val.Int64() < 0 {
		return 0, fmt.Errorf("value out of uint8 range: %s", val)
	}
	return uint8(val.Int64()), nil
}
