// Package permit implements EIP-2612 permit signing for USDC paymaster interactions.
// It follows the same pattern as internal/payment/eip3009/builder.go.
package permit

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// PermitSigner abstracts wallet signing to avoid direct wallet package imports.
type PermitSigner interface {
	// SignTransaction signs raw bytes (no additional hashing).
	SignTransaction(ctx context.Context, rawTx []byte) ([]byte, error)
	// Address returns the signer's public address.
	Address(ctx context.Context) (string, error)
}

// EthCaller abstracts eth_call for on-chain reads.
type EthCaller interface {
	CallContract(ctx context.Context, msg ethereum.CallMsg, block *big.Int) ([]byte, error)
}

// EIP-712 type hashes for USDC v2 domain and Permit.
var (
	eip712DomainTypeHash = crypto.Keccak256([]byte(
		"EIP712Domain(string name,string version,uint256 chainId," +
			"address verifyingContract)",
	))

	permitTypeHash = crypto.Keccak256([]byte(
		"Permit(address owner,address spender,uint256 value," +
			"uint256 nonce,uint256 deadline)",
	))

	usdcName    = crypto.Keccak256([]byte("USD Coin"))
	usdcVersion = crypto.Keccak256([]byte("2"))

	// noncesSelector is the 4-byte function selector for nonces(address) → 0x7ecebe00.
	noncesSelector = common.FromHex("0x7ecebe00")
)

// DomainSeparator computes the EIP-712 domain separator for USDC v2.
func DomainSeparator(chainID int64, usdcAddr common.Address) []byte {
	buf := make([]byte, 5*32)
	copy(buf[:32], eip712DomainTypeHash)
	copy(buf[32:64], usdcName)
	copy(buf[64:96], usdcVersion)
	big.NewInt(chainID).FillBytes(buf[96:128])
	copy(buf[128+12:160], usdcAddr.Bytes())
	return crypto.Keccak256(buf)
}

// PermitStructHash computes the struct hash for an EIP-2612 Permit.
func PermitStructHash(owner, spender common.Address, value, nonce, deadline *big.Int) []byte {
	buf := make([]byte, 6*32)
	copy(buf[:32], permitTypeHash)
	copy(buf[32+12:64], owner.Bytes())
	copy(buf[64+12:96], spender.Bytes())
	value.FillBytes(buf[96:128])
	nonce.FillBytes(buf[128:160])
	deadline.FillBytes(buf[160:192])
	return crypto.Keccak256(buf)
}

// TypedDataHash computes the EIP-712 hash to be signed for a Permit.
func TypedDataHash(domainSep, structHash []byte) []byte {
	msg := make([]byte, 2+32+32)
	msg[0] = 0x19
	msg[1] = 0x01
	copy(msg[2:34], domainSep)
	copy(msg[34:66], structHash)
	return crypto.Keccak256(msg)
}

// Sign computes the typed data hash and signs it with the provided signer.
// Returns (v, r, s) where v is 27 or 28.
func Sign(
	ctx context.Context,
	signer PermitSigner,
	owner, spender common.Address,
	value, nonce, deadline *big.Int,
	chainID int64,
	usdcAddr common.Address,
) (v uint8, r, s [32]byte, err error) {
	domainSep := DomainSeparator(chainID, usdcAddr)
	structHash := PermitStructHash(owner, spender, value, nonce, deadline)
	hash := TypedDataHash(domainSep, structHash)

	sig, err := signer.SignTransaction(ctx, hash)
	if err != nil {
		return 0, r, s, fmt.Errorf("sign permit: %w", err)
	}

	if len(sig) != 65 {
		return 0, r, s, fmt.Errorf("invalid signature length %d, want 65", len(sig))
	}

	copy(r[:], sig[:32])
	copy(s[:], sig[32:64])

	// go-ethereum uses V=0/1 (recovery id); EIP-2612 expects 27/28.
	v = sig[64]
	if v < 27 {
		v += 27
	}

	return v, r, s, nil
}

// GetPermitNonce queries the USDC nonces(address) function to retrieve the
// current EIP-2612 permit nonce for an owner.
func GetPermitNonce(
	ctx context.Context,
	caller EthCaller,
	usdcAddr common.Address,
	owner common.Address,
) (*big.Int, error) {
	// ABI-encode: nonces(address) — selector(4) + address(32) = 36 bytes.
	calldata := make([]byte, 0, 36)
	calldata = append(calldata, noncesSelector...)
	addrPadded := make([]byte, 32)
	copy(addrPadded[12:], owner.Bytes())
	calldata = append(calldata, addrPadded...)

	to := usdcAddr
	result, err := caller.CallContract(ctx, ethereum.CallMsg{
		To:   &to,
		Data: calldata,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("call nonces(%s): %w", owner.Hex(), err)
	}

	if len(result) < 32 {
		return nil, fmt.Errorf("nonces result too short: %d bytes", len(result))
	}

	return new(big.Int).SetBytes(result[:32]), nil
}
