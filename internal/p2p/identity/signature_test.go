package identity

import (
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestVerifyMessageSignature(t *testing.T) {
	key, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	pub := ethcrypto.CompressPubkey(&key.PublicKey)
	did, err := DIDFromPublicKey(pub)
	require.NoError(t, err)

	msg := []byte("signed provenance payload")
	sig, err := ethcrypto.Sign(ethcrypto.Keccak256(msg), key)
	require.NoError(t, err)

	require.NoError(t, VerifyMessageSignature(did.ID, msg, sig))

	bad := append([]byte(nil), msg...)
	bad[0] ^= 0x01
	require.Error(t, VerifyMessageSignature(did.ID, bad, sig))
}
