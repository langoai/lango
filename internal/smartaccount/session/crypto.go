package session

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// GenerateSessionKey creates a new ECDSA key pair for session signing.
func GenerateSessionKey() (*ecdsa.PrivateKey, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("generate session key: %w", err)
	}
	return key, nil
}

// AddressFromPublicKey derives the Ethereum address from a public key.
func AddressFromPublicKey(pub *ecdsa.PublicKey) common.Address {
	return crypto.PubkeyToAddress(*pub)
}

// SerializePrivateKey serializes an ECDSA private key to bytes.
func SerializePrivateKey(key *ecdsa.PrivateKey) []byte {
	return crypto.FromECDSA(key)
}

// DeserializePrivateKey restores an ECDSA private key from bytes.
func DeserializePrivateKey(data []byte) (*ecdsa.PrivateKey, error) {
	key, err := crypto.ToECDSA(data)
	if err != nil {
		return nil, fmt.Errorf("deserialize session key: %w", err)
	}
	return key, nil
}

// SerializePublicKey serializes a public key to compressed bytes.
func SerializePublicKey(pub *ecdsa.PublicKey) []byte {
	return elliptic.MarshalCompressed(pub.Curve, pub.X, pub.Y)
}
