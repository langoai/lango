package wallet

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// UserOpSigner signs ERC-4337 UserOperation hashes.
type UserOpSigner interface {
	// SignUserOp signs a UserOp hash for the given entry point and chain.
	SignUserOp(
		ctx context.Context,
		userOpHash []byte,
		entryPoint common.Address,
		chainID *big.Int,
	) ([]byte, error)
}

// LocalUserOpSigner signs UserOps using a local ECDSA private key.
type LocalUserOpSigner struct {
	key *ecdsa.PrivateKey
}

// NewLocalUserOpSigner creates a signer from an ECDSA private key.
func NewLocalUserOpSigner(key *ecdsa.PrivateKey) *LocalUserOpSigner {
	return &LocalUserOpSigner{key: key}
}

// SignUserOp computes the ERC-4337 UserOp signature.
// The hash is: keccak256(abi.encode(userOpHash, entryPoint, chainId)).
func (s *LocalUserOpSigner) SignUserOp(
	_ context.Context,
	userOpHash []byte,
	entryPoint common.Address,
	chainID *big.Int,
) ([]byte, error) {
	// Pack: userOpHash (32 bytes) + entryPoint (32 bytes, left-padded)
	// + chainID (32 bytes)
	packed := make([]byte, 96)
	copy(packed[0:32], userOpHash)
	// 20 bytes right-aligned in 32-byte word
	copy(packed[44:64], entryPoint.Bytes())
	chainIDBytes := chainID.Bytes()
	copy(packed[96-len(chainIDBytes):96], chainIDBytes)

	finalHash := crypto.Keccak256(packed)
	// Ethereum personal_sign prefix
	prefixed := crypto.Keccak256(
		[]byte(fmt.Sprintf(
			"\x19Ethereum Signed Message:\n%d", len(finalHash),
		)),
		finalHash,
	)

	sig, err := crypto.Sign(prefixed, s.key)
	if err != nil {
		return nil, fmt.Errorf("sign user op: %w", err)
	}
	// Adjust v value for Ethereum (add 27)
	if sig[64] < 27 {
		sig[64] += 27
	}
	return sig, nil
}
