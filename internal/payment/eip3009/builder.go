// Package eip3009 implements EIP-3009 transferWithAuthorization typed data
// building and signing for USDC gasless transfers.
package eip3009

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// WalletSigner abstracts wallet signing to avoid direct wallet package imports.
type WalletSigner interface {
	SignMessage(ctx context.Context, message []byte) ([]byte, error)
	Address(ctx context.Context) (string, error)
}

// Authorization is a fully signed EIP-3009 transferWithAuthorization.
type Authorization struct {
	From        common.Address
	To          common.Address
	Value       *big.Int
	ValidAfter  *big.Int
	ValidBefore *big.Int
	Nonce       [32]byte
	V           uint8
	R, S        [32]byte
}

// UnsignedAuth holds the authorization parameters before signing.
type UnsignedAuth struct {
	From        common.Address
	To          common.Address
	Value       *big.Int
	ValidAfter  *big.Int
	ValidBefore *big.Int
	Nonce       [32]byte
}

// EIP-712 type hashes for USDC v2 domain and TransferWithAuthorization.
var (
	eip712DomainTypeHash = crypto.Keccak256([]byte(
		"EIP712Domain(string name,string version,uint256 chainId," +
			"address verifyingContract)",
	))

	transferAuthTypeHash = crypto.Keccak256([]byte(
		"TransferWithAuthorization(address from,address to," +
			"uint256 value,uint256 validAfter,uint256 validBefore," +
			"bytes32 nonce)",
	))

	// transferWithAuthSelector is the 4-byte function selector for
	// transferWithAuthorization(address,address,uint256,uint256,uint256,bytes32,uint8,bytes32,bytes32).
	transferWithAuthSelector = crypto.Keccak256([]byte(
		"transferWithAuthorization(address,address,uint256,uint256," +
			"uint256,bytes32,uint8,bytes32,bytes32)",
	))[:4]

	usdcName    = crypto.Keccak256([]byte("USD Coin"))
	usdcVersion = crypto.Keccak256([]byte("2"))
)

// NewUnsigned creates an unsigned EIP-3009 authorization with a random nonce.
// validAfter is set to now; validBefore is set to the given deadline.
func NewUnsigned(
	from, to common.Address,
	value *big.Int,
	deadline time.Time,
) *UnsignedAuth {
	var nonce [32]byte
	// crypto/rand.Read never returns an error on supported platforms.
	_, _ = rand.Read(nonce[:])

	return &UnsignedAuth{
		From:        from,
		To:          to,
		Value:       new(big.Int).Set(value),
		ValidAfter:  big.NewInt(time.Now().Unix()),
		ValidBefore: big.NewInt(deadline.Unix()),
		Nonce:       nonce,
	}
}

// TypedDataHash computes the EIP-712 hash to be signed for a
// transferWithAuthorization on the given chain and USDC contract.
func TypedDataHash(
	auth *UnsignedAuth,
	chainID int64,
	usdcAddr common.Address,
) ([]byte, error) {
	domainSep := domainSeparator(chainID, usdcAddr)
	structHash := authStructHash(auth)

	// EIP-712: keccak256("\x19\x01" || domainSeparator || structHash)
	msg := make([]byte, 2+32+32)
	msg[0] = 0x19
	msg[1] = 0x01
	copy(msg[2:34], domainSep)
	copy(msg[34:66], structHash)

	return crypto.Keccak256(msg), nil
}

// Sign computes the typed data hash and signs it with the provided wallet.
func Sign(
	ctx context.Context,
	wallet WalletSigner,
	auth *UnsignedAuth,
	chainID int64,
	usdcAddr common.Address,
) (*Authorization, error) {
	hash, err := TypedDataHash(auth, chainID, usdcAddr)
	if err != nil {
		return nil, fmt.Errorf("typed data hash: %w", err)
	}

	sig, err := wallet.SignMessage(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("sign message: %w", err)
	}

	if len(sig) != 65 {
		return nil, fmt.Errorf("invalid signature length %d, want 65", len(sig))
	}

	result := &Authorization{
		From:        auth.From,
		To:          auth.To,
		Value:       new(big.Int).Set(auth.Value),
		ValidAfter:  new(big.Int).Set(auth.ValidAfter),
		ValidBefore: new(big.Int).Set(auth.ValidBefore),
		Nonce:       auth.Nonce,
	}

	copy(result.R[:], sig[:32])
	copy(result.S[:], sig[32:64])

	// go-ethereum uses V=0/1 (recovery id); EIP-3009 expects 27/28.
	v := sig[64]
	if v < 27 {
		v += 27
	}
	result.V = v

	return result, nil
}

// Verify recovers the signer from the authorization signature and checks that
// it matches the expected from address.
func Verify(
	auth *Authorization,
	expectedFrom common.Address,
	chainID int64,
	usdcAddr common.Address,
) error {
	// Reconstruct the unsigned auth to compute the hash.
	unsigned := &UnsignedAuth{
		From:        auth.From,
		To:          auth.To,
		Value:       auth.Value,
		ValidAfter:  auth.ValidAfter,
		ValidBefore: auth.ValidBefore,
		Nonce:       auth.Nonce,
	}

	hash, err := TypedDataHash(unsigned, chainID, usdcAddr)
	if err != nil {
		return fmt.Errorf("typed data hash: %w", err)
	}

	// Reconstruct the 65-byte signature with recovery id 0/1.
	var sig [65]byte
	copy(sig[:32], auth.R[:])
	copy(sig[32:64], auth.S[:])
	v := auth.V
	if v >= 27 {
		v -= 27
	}
	sig[64] = v

	pubKey, err := crypto.Ecrecover(hash, sig[:])
	if err != nil {
		return fmt.Errorf("ecrecover: %w", err)
	}

	// Convert uncompressed public key (65 bytes) to address.
	recovered := common.BytesToAddress(crypto.Keccak256(pubKey[1:])[12:])
	if recovered != expectedFrom {
		return fmt.Errorf(
			"signer mismatch: recovered %s, want %s",
			recovered.Hex(), expectedFrom.Hex(),
		)
	}

	return nil
}

// EncodeCalldata ABI-encodes the transferWithAuthorization call for on-chain
// submission. Layout: selector(4) + from(32) + to(32) + value(32) +
// validAfter(32) + validBefore(32) + nonce(32) + v(32) + r(32) + s(32).
func EncodeCalldata(auth *Authorization) []byte {
	// 4-byte selector + 9 * 32-byte parameters = 292 bytes
	data := make([]byte, 4+9*32)
	copy(data[:4], transferWithAuthSelector)

	off := 4
	// from (left-padded address)
	copy(data[off+12:off+32], auth.From.Bytes())
	off += 32

	// to (left-padded address)
	copy(data[off+12:off+32], auth.To.Bytes())
	off += 32

	// value
	auth.Value.FillBytes(data[off : off+32])
	off += 32

	// validAfter
	auth.ValidAfter.FillBytes(data[off : off+32])
	off += 32

	// validBefore
	auth.ValidBefore.FillBytes(data[off : off+32])
	off += 32

	// nonce
	copy(data[off:off+32], auth.Nonce[:])
	off += 32

	// v (uint8 as uint256)
	data[off+31] = auth.V
	off += 32

	// r
	copy(data[off:off+32], auth.R[:])
	off += 32

	// s
	copy(data[off:off+32], auth.S[:])

	return data
}

// domainSeparator computes the EIP-712 domain separator for USDC v2.
func domainSeparator(chainID int64, verifyingContract common.Address) []byte {
	// abi.encode(typeHash, nameHash, versionHash, chainId, verifyingContract)
	buf := make([]byte, 5*32)
	copy(buf[:32], eip712DomainTypeHash)
	copy(buf[32:64], usdcName)
	copy(buf[64:96], usdcVersion)
	big.NewInt(chainID).FillBytes(buf[96:128])
	copy(buf[128+12:160], verifyingContract.Bytes())
	return crypto.Keccak256(buf)
}

// authStructHash computes the struct hash for TransferWithAuthorization.
func authStructHash(auth *UnsignedAuth) []byte {
	buf := make([]byte, 7*32)
	copy(buf[:32], transferAuthTypeHash)
	copy(buf[32+12:64], auth.From.Bytes())
	copy(buf[64+12:96], auth.To.Bytes())
	auth.Value.FillBytes(buf[96:128])
	auth.ValidAfter.FillBytes(buf[128:160])
	auth.ValidBefore.FillBytes(buf[160:192])
	copy(buf[192:224], auth.Nonce[:])
	return crypto.Keccak256(buf)
}
