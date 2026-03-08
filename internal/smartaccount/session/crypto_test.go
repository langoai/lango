package session

import (
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSessionKey(t *testing.T) {
	t.Parallel()

	key, err := GenerateSessionKey()
	require.NoError(t, err)
	require.NotNil(t, key)

	assert.NotNil(t, key.PublicKey.X, "public key X must be set")
	assert.NotNil(t, key.PublicKey.Y, "public key Y must be set")
	assert.Equal(t, crypto.S256(), key.Curve,
		"key must use secp256k1 curve")
}

func TestGenerateSessionKey_Unique(t *testing.T) {
	t.Parallel()

	key1, err := GenerateSessionKey()
	require.NoError(t, err)
	key2, err := GenerateSessionKey()
	require.NoError(t, err)

	assert.NotEqual(t, key1.D.Bytes(), key2.D.Bytes(),
		"two generated keys must differ")
}

func TestSerializeDeserializePrivateKey_Roundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
	}{
		{give: "roundtrip_1"},
		{give: "roundtrip_2"},
		{give: "roundtrip_3"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			original, err := GenerateSessionKey()
			require.NoError(t, err)

			serialized := SerializePrivateKey(original)
			require.NotEmpty(t, serialized,
				"serialized key must not be empty")

			restored, err := DeserializePrivateKey(serialized)
			require.NoError(t, err)

			assert.Equal(t, original.D.Bytes(), restored.D.Bytes(),
				"private key D must match after roundtrip")
			assert.Equal(t, original.PublicKey.X.Bytes(), restored.PublicKey.X.Bytes(),
				"public key X must match after roundtrip")
			assert.Equal(t, original.PublicKey.Y.Bytes(), restored.PublicKey.Y.Bytes(),
				"public key Y must match after roundtrip")
		})
	}
}

func TestDeserializePrivateKey_InvalidData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give     string
		giveData []byte
	}{
		{give: "empty", giveData: []byte{}},
		{give: "too_short", giveData: []byte{0x01, 0x02}},
		{give: "too_long", giveData: make([]byte, 64)},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			_, err := DeserializePrivateKey(tt.giveData)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "deserialize session key")
		})
	}
}

func TestAddressFromPublicKey(t *testing.T) {
	t.Parallel()

	key, err := GenerateSessionKey()
	require.NoError(t, err)

	addr := AddressFromPublicKey(&key.PublicKey)
	assert.NotEqual(t, common.Address{}, addr,
		"address must not be zero")
	assert.Len(t, addr.Bytes(), 20, "Ethereum address must be 20 bytes")
}

func TestAddressFromPublicKey_Deterministic(t *testing.T) {
	t.Parallel()

	key, err := GenerateSessionKey()
	require.NoError(t, err)

	addr1 := AddressFromPublicKey(&key.PublicKey)
	addr2 := AddressFromPublicKey(&key.PublicKey)

	assert.Equal(t, addr1, addr2,
		"same public key must produce same address")
}

func TestAddressFromPublicKey_DifferentKeys(t *testing.T) {
	t.Parallel()

	key1, err := GenerateSessionKey()
	require.NoError(t, err)
	key2, err := GenerateSessionKey()
	require.NoError(t, err)

	addr1 := AddressFromPublicKey(&key1.PublicKey)
	addr2 := AddressFromPublicKey(&key2.PublicKey)

	assert.NotEqual(t, addr1, addr2,
		"different keys must produce different addresses")
}

func TestAddressFromPublicKey_MatchesCryptoLib(t *testing.T) {
	t.Parallel()

	key, err := GenerateSessionKey()
	require.NoError(t, err)

	got := AddressFromPublicKey(&key.PublicKey)
	want := crypto.PubkeyToAddress(key.PublicKey)

	assert.Equal(t, want, got,
		"must match go-ethereum's PubkeyToAddress")
}

func TestSerializePublicKey(t *testing.T) {
	t.Parallel()

	key, err := GenerateSessionKey()
	require.NoError(t, err)

	serialized := SerializePublicKey(&key.PublicKey)

	// Compressed public key is 33 bytes (0x02 or 0x03 prefix + 32 byte X).
	require.Len(t, serialized, 33,
		"compressed public key must be 33 bytes")

	prefix := serialized[0]
	assert.True(t, prefix == 0x02 || prefix == 0x03,
		"compressed key must start with 0x02 or 0x03, got 0x%02x", prefix)
}

func TestSerializePublicKey_Recoverable(t *testing.T) {
	t.Parallel()

	key, err := GenerateSessionKey()
	require.NoError(t, err)

	serialized := SerializePublicKey(&key.PublicKey)

	// Decompress using go-ethereum's secp256k1 decompressor.
	recovered, err := crypto.DecompressPubkey(serialized)
	require.NoError(t, err, "decompression must succeed")
	require.NotNil(t, recovered, "recovered key must not be nil")

	assert.Equal(t, key.PublicKey.X.Bytes(), recovered.X.Bytes(),
		"recovered X must match original")
	assert.Equal(t, key.PublicKey.Y.Bytes(), recovered.Y.Bytes(),
		"recovered Y must match original")
}

func TestSerializePrivateKey_Length(t *testing.T) {
	t.Parallel()

	key, err := GenerateSessionKey()
	require.NoError(t, err)

	serialized := SerializePrivateKey(key)
	assert.Len(t, serialized, 32,
		"serialized private key must be 32 bytes")
}

func TestFullCryptoRoundtrip(t *testing.T) {
	t.Parallel()

	// Generate -> serialize -> deserialize -> derive address -> verify
	key, err := GenerateSessionKey()
	require.NoError(t, err)

	privBytes := SerializePrivateKey(key)
	pubBytes := SerializePublicKey(&key.PublicKey)
	origAddr := AddressFromPublicKey(&key.PublicKey)

	// Restore private key.
	restored, err := DeserializePrivateKey(privBytes)
	require.NoError(t, err)

	// Restored key produces the same address.
	restoredAddr := AddressFromPublicKey(&restored.PublicKey)
	assert.Equal(t, origAddr, restoredAddr,
		"restored key must produce same address")

	// Restored key produces the same serialized public key.
	restoredPub := SerializePublicKey(&restored.PublicKey)
	assert.Equal(t, pubBytes, restoredPub,
		"restored key must produce same serialized public key")
}

func TestAddressFromPublicKey_KnownKey(t *testing.T) {
	t.Parallel()

	// Use a well-known test private key to verify address derivation.
	hexKey := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	key, err := crypto.HexToECDSA(hexKey)
	require.NoError(t, err)

	addr := AddressFromPublicKey(&key.PublicKey)
	want := crypto.PubkeyToAddress(key.PublicKey)

	assert.Equal(t, want, addr)
}

func TestDeserializePrivateKey_PreservesPublicKey(t *testing.T) {
	t.Parallel()

	key, err := GenerateSessionKey()
	require.NoError(t, err)

	data := SerializePrivateKey(key)
	restored, err := DeserializePrivateKey(data)
	require.NoError(t, err)

	// Public key should be fully reconstructed.
	assert.True(t, restored.PublicKey.IsOnCurve(
		restored.PublicKey.X, restored.PublicKey.Y,
	), "restored public key must be on curve")

	// Signing with restored key should be verifiable with original pub key.
	msg := crypto.Keccak256([]byte("test message"))
	sig, err := crypto.Sign(msg, restored)
	require.NoError(t, err)

	recoveredPub, err := crypto.Ecrecover(msg, sig)
	require.NoError(t, err)

	origPub := crypto.FromECDSAPub(&key.PublicKey)
	assert.Equal(t, origPub, recoveredPub,
		"signature from restored key must be verifiable with original public key")
}

func TestSerializePublicKey_DifferentKeys(t *testing.T) {
	t.Parallel()

	key1, err := GenerateSessionKey()
	require.NoError(t, err)
	key2, err := GenerateSessionKey()
	require.NoError(t, err)

	pub1 := SerializePublicKey(&key1.PublicKey)
	pub2 := SerializePublicKey(&key2.PublicKey)

	assert.NotEqual(t, pub1, pub2,
		"different keys must produce different serialized forms")
}

func TestSignAndVerifyWithSessionKey(t *testing.T) {
	t.Parallel()

	key, err := GenerateSessionKey()
	require.NoError(t, err)

	// Sign a message.
	msg := crypto.Keccak256([]byte("hello world"))
	sig, err := crypto.Sign(msg, key)
	require.NoError(t, err)

	// Recover the public key from the signature.
	recoveredPubBytes, err := crypto.Ecrecover(msg, sig)
	require.NoError(t, err)

	recoveredPub, err := crypto.UnmarshalPubkey(recoveredPubBytes)
	require.NoError(t, err)

	recoveredAddr := AddressFromPublicKey(recoveredPub)
	expectedAddr := AddressFromPublicKey(&key.PublicKey)

	assert.Equal(t, expectedAddr, recoveredAddr,
		"recovered address must match original session key address")
}

// Compile-time check that GenerateSessionKey returns secp256k1 keys.
func TestGenerateSessionKey_Secp256k1(t *testing.T) {
	t.Parallel()

	key, err := GenerateSessionKey()
	require.NoError(t, err)

	// Verify the curve parameters match secp256k1.
	want := crypto.S256().Params()
	got := key.Curve.Params()
	assert.Equal(t, want.N, got.N, "curve order N must match secp256k1")
	assert.Equal(t, want.P, got.P, "field prime P must match secp256k1")
}

func TestDeserializePrivateKey_NilInput(t *testing.T) {
	t.Parallel()

	_, err := DeserializePrivateKey(nil)
	assert.Error(t, err, "nil input must error")
}

// Verify that AddressFromPublicKey works with a manually constructed key.
func TestAddressFromPublicKey_ManualKey(t *testing.T) {
	t.Parallel()

	// Create a key from a known hex.
	privHex := "4c0883a69102937d6231471b5dbb6204fe512961708279f5c6a7b2e1ce66ac1f"
	key, err := crypto.HexToECDSA(privHex)
	require.NoError(t, err)

	pub := &ecdsa.PublicKey{
		Curve: key.PublicKey.Curve,
		X:     key.PublicKey.X,
		Y:     key.PublicKey.Y,
	}

	addr := AddressFromPublicKey(pub)
	want := crypto.PubkeyToAddress(*pub)
	assert.Equal(t, want, addr)
}
