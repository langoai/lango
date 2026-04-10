package handshake

import (
	"bytes"
	"encoding/binary"
	"fmt"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// ResponseVerifyFunc verifies a signature against a claimed public key.
// pubkey is the claimed compressed public key, nonce is the challenge nonce,
// and signature is the response signature to verify.
type ResponseVerifyFunc func(pubkey, nonce, signature []byte) error

// VerifySecp256k1Signature is the default response verifier using secp256k1+keccak256.
// It recovers the public key from the ECDSA signature and compares it with the
// claimed key. This is the same algorithm used by wallet.SignMessage.
func VerifySecp256k1Signature(pubkey, nonce, signature []byte) error {
	if len(signature) != 65 {
		return fmt.Errorf("invalid signature length: %d (expected 65)", len(signature))
	}
	hash := ethcrypto.Keccak256(nonce)
	recoveredPub, err := ethcrypto.SigToPub(hash, signature)
	if err != nil {
		return fmt.Errorf("recover public key from signature: %w", err)
	}
	recoveredCompressed := ethcrypto.CompressPubkey(recoveredPub)
	if !bytes.Equal(recoveredCompressed, pubkey) {
		return fmt.Errorf("signature public key mismatch")
	}
	return nil
}

// verifyChallengeSignature verifies the ECDSA signature on a v1.1 challenge.
func verifyChallengeSignature(c *Challenge) error {
	if len(c.Signature) != 65 {
		return fmt.Errorf("invalid signature length: %d (expected 65)", len(c.Signature))
	}

	payload := challengeSignPayload(c.Nonce, c.Timestamp, c.SenderDID)
	recovered, err := ethcrypto.SigToPub(payload, c.Signature)
	if err != nil {
		return fmt.Errorf("recover public key: %w", err)
	}

	recoveredCompressed := ethcrypto.CompressPubkey(recovered)
	if !bytes.Equal(recoveredCompressed, c.PublicKey) {
		return fmt.Errorf("public key mismatch")
	}

	return nil
}

// challengeSignPayload constructs the canonical bytes for challenge signing:
// nonce || bigEndian(timestamp, 8) || utf8(senderDID)
func challengeSignPayload(nonce []byte, timestamp int64, senderDID string) []byte {
	buf := make([]byte, 0, len(nonce)+8+len(senderDID))
	buf = append(buf, nonce...)
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, uint64(timestamp))
	buf = append(buf, ts...)
	buf = append(buf, []byte(senderDID)...)
	return ethcrypto.Keccak256(buf)
}
