package eip3009

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testFrom     = common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	testTo       = common.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	testUSDCAddr = common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	testChainID  = int64(1)
)

func TestNewUnsigned(t *testing.T) {
	value := big.NewInt(1_000_000) // 1 USDC
	deadline := time.Now().Add(1 * time.Hour)

	auth := NewUnsigned(testFrom, testTo, value, deadline)

	assert.Equal(t, testFrom, auth.From)
	assert.Equal(t, testTo, auth.To)
	assert.Equal(t, value, auth.Value)
	assert.Equal(t, big.NewInt(deadline.Unix()), auth.ValidBefore)
	assert.NotEqual(t, [32]byte{}, auth.Nonce, "nonce should be random, not zero")

	// Value should be a copy, not shared reference.
	value.SetInt64(999)
	assert.NotEqual(t, value, auth.Value)
}

func TestTypedDataHash(t *testing.T) {
	auth := &UnsignedAuth{
		From:        testFrom,
		To:          testTo,
		Value:       big.NewInt(1_000_000),
		ValidAfter:  big.NewInt(1000),
		ValidBefore: big.NewInt(2000),
		Nonce:       [32]byte{0x01, 0x02, 0x03},
	}

	hash1, err := TypedDataHash(auth, testChainID, testUSDCAddr)
	require.NoError(t, err)
	assert.Len(t, hash1, 32)

	// Same inputs produce same hash (deterministic).
	hash2, err := TypedDataHash(auth, testChainID, testUSDCAddr)
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2)

	// Different chain ID produces different hash.
	hash3, err := TypedDataHash(auth, 8453, testUSDCAddr)
	require.NoError(t, err)
	assert.NotEqual(t, hash1, hash3)

	// Different nonce produces different hash.
	auth2 := *auth
	auth2.Nonce = [32]byte{0x04, 0x05, 0x06}
	hash4, err := TypedDataHash(&auth2, testChainID, testUSDCAddr)
	require.NoError(t, err)
	assert.NotEqual(t, hash1, hash4)
}

// testWallet implements WalletSigner using a raw ECDSA key for testing.
type testWallet struct {
	key *ecdsa.PrivateKey
}

func (w *testWallet) SignMessage(_ context.Context, message []byte) ([]byte, error) {
	return crypto.Sign(message, w.key)
}

func (w *testWallet) Address(_ context.Context) (string, error) {
	addr := crypto.PubkeyToAddress(w.key.PublicKey)
	return addr.Hex(), nil
}

func TestSignAndVerify(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)

	wallet := &testWallet{key: key}
	fromAddr := crypto.PubkeyToAddress(key.PublicKey)

	auth := &UnsignedAuth{
		From:        fromAddr,
		To:          testTo,
		Value:       big.NewInt(5_000_000),
		ValidAfter:  big.NewInt(100),
		ValidBefore: big.NewInt(9999),
		Nonce:       [32]byte{0xAA, 0xBB},
	}

	signed, err := Sign(
		context.Background(), wallet, auth, testChainID, testUSDCAddr,
	)
	require.NoError(t, err)

	assert.Equal(t, fromAddr, signed.From)
	assert.Equal(t, testTo, signed.To)
	assert.True(t, signed.V == 27 || signed.V == 28, "V should be 27 or 28")
	assert.NotEqual(t, [32]byte{}, signed.R)
	assert.NotEqual(t, [32]byte{}, signed.S)

	// Verify should succeed with the correct from address.
	err = Verify(signed, fromAddr, testChainID, testUSDCAddr)
	require.NoError(t, err)

	// Verify should fail with wrong expected address.
	wrongAddr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	err = Verify(signed, wrongAddr, testChainID, testUSDCAddr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signer mismatch")
}

func TestVerifyBadSignature(t *testing.T) {
	auth := &Authorization{
		From:        testFrom,
		To:          testTo,
		Value:       big.NewInt(1_000_000),
		ValidAfter:  big.NewInt(0),
		ValidBefore: big.NewInt(9999),
		Nonce:       [32]byte{0x01},
		V:           27,
		R:           [32]byte{0xFF},
		S:           [32]byte{0xFF},
	}

	err := Verify(auth, testFrom, testChainID, testUSDCAddr)
	require.Error(t, err)
}

func TestEncodeCalldata(t *testing.T) {
	auth := &Authorization{
		From:        testFrom,
		To:          testTo,
		Value:       big.NewInt(1_000_000),
		ValidAfter:  big.NewInt(100),
		ValidBefore: big.NewInt(200),
		Nonce:       [32]byte{0x01},
		V:           28,
		R:           [32]byte{0xAA},
		S:           [32]byte{0xBB},
	}

	data := EncodeCalldata(auth)

	// 4-byte selector + 9 * 32-byte params = 292 bytes
	assert.Len(t, data, 4+9*32)

	// Verify selector matches transferWithAuthorization.
	assert.Equal(t, transferWithAuthSelector, data[:4])

	// Verify from address is at offset 4, left-padded.
	assert.Equal(t, testFrom.Bytes(), data[4+12:4+32])

	// Verify to address is at offset 36, left-padded.
	assert.Equal(t, testTo.Bytes(), data[36+12:36+32])
}

func TestEncodeCalldataRoundTrip(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)

	wallet := &testWallet{key: key}
	fromAddr := crypto.PubkeyToAddress(key.PublicKey)

	unsigned := &UnsignedAuth{
		From:        fromAddr,
		To:          testTo,
		Value:       big.NewInt(10_000_000),
		ValidAfter:  big.NewInt(0),
		ValidBefore: big.NewInt(99999),
		Nonce:       [32]byte{0xDE, 0xAD},
	}

	signed, err := Sign(
		context.Background(), wallet, unsigned, testChainID, testUSDCAddr,
	)
	require.NoError(t, err)

	calldata := EncodeCalldata(signed)
	assert.Len(t, calldata, 4+9*32)

	// Verify the signed auth still passes verification.
	err = Verify(signed, fromAddr, testChainID, testUSDCAddr)
	require.NoError(t, err)
}
