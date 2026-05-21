package session

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	sa "github.com/langoai/lango/internal/smartaccount"
)

func TestManagerEncryptionOptionsAndSignUserOp(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()
	ciphertext := []byte("encrypted session key")
	var encryptedPlaintext []byte
	var encryptKeyID string
	var decryptKeyID string
	var decryptCiphertext []byte

	mgr := NewManager(
		store,
		WithEncryption(
			func(_ context.Context, keyID string, plaintext []byte) ([]byte, error) {
				encryptKeyID = keyID
				encryptedPlaintext = append([]byte(nil), plaintext...)
				return ciphertext, nil
			},
			func(_ context.Context, keyID string, ciphertext []byte) ([]byte, error) {
				decryptKeyID = keyID
				decryptCiphertext = append([]byte(nil), ciphertext...)
				return encryptedPlaintext, nil
			},
		),
		WithEntryPoint(common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032")),
		WithChainID(84532),
	)

	key, err := mgr.Create(ctx, defaultPolicy(time.Hour), "")
	require.NoError(t, err)
	require.Equal(t, hex.EncodeToString(ciphertext), key.PrivateKeyRef)
	require.NotEqual(t, string(encryptedPlaintext), key.PrivateKeyRef)
	require.Equal(t, defaultCryptoKeyID, encryptKeyID)

	op := &sa.UserOperation{
		Sender:               key.Address,
		Nonce:                big.NewInt(1),
		InitCode:             []byte{0x01},
		CallData:             []byte{0x02},
		CallGasLimit:         big.NewInt(100),
		VerificationGasLimit: big.NewInt(200),
		PreVerificationGas:   big.NewInt(30),
		MaxFeePerGas:         big.NewInt(40),
		MaxPriorityFeePerGas: big.NewInt(5),
		PaymasterAndData:     []byte{0x03},
	}
	sig, err := mgr.SignUserOp(ctx, key.ID, op)
	require.NoError(t, err)
	require.Len(t, sig, 65)
	require.Equal(t, defaultCryptoKeyID, decryptKeyID)
	require.Equal(t, ciphertext, decryptCiphertext)
	requireSessionSignatureAddress(t, op, sig, key.Address)
}

func TestManagerSignUserOpWithoutEncryptionUsesStoredPrivateKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()
	mgr := NewManager(
		store,
		WithEntryPoint(common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032")),
		WithChainID(84532),
	)

	key, err := mgr.Create(ctx, defaultPolicy(time.Hour), "")
	require.NoError(t, err)
	require.NotEmpty(t, key.PrivateKeyRef)
	decoded, err := hex.DecodeString(key.PrivateKeyRef)
	require.NoError(t, err)
	require.Len(t, decoded, 32)

	op := &sa.UserOperation{
		Sender:               key.Address,
		Nonce:                big.NewInt(1),
		InitCode:             []byte{0x01},
		CallData:             []byte{0x02},
		CallGasLimit:         big.NewInt(100),
		VerificationGasLimit: big.NewInt(200),
		PreVerificationGas:   big.NewInt(30),
		MaxFeePerGas:         big.NewInt(40),
		MaxPriorityFeePerGas: big.NewInt(5),
		PaymasterAndData:     []byte{0x03},
	}
	sig, err := mgr.SignUserOp(ctx, key.ID, op)
	require.NoError(t, err)
	require.Len(t, sig, 65)
	requireSessionSignatureAddress(t, op, sig, key.Address)
}

func requireSessionSignatureAddress(
	t *testing.T, op *sa.UserOperation, sig []byte, want common.Address,
) {
	t.Helper()

	hash := sa.ComputeUserOpHash(
		op,
		common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032"),
		84532,
	)
	recoveredPub, err := crypto.Ecrecover(hash, sig)
	require.NoError(t, err)
	pubKey, err := crypto.UnmarshalPubkey(recoveredPub)
	require.NoError(t, err)
	require.Equal(t, want, crypto.PubkeyToAddress(*pubKey))
}

func TestManagerSignUserOpEncryptedKeyErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()
	active := makeSessionKey("bad-hex", "", false, time.Now().Add(time.Hour))
	active.PrivateKeyRef = "not hex"
	require.NoError(t, store.Save(ctx, active))

	mgr := NewManager(store, WithEncryption(nil, func(context.Context, string, []byte) ([]byte, error) {
		t.Fatal("decrypt should not run when key reference is not hex")
		return nil, nil
	}))

	_, err := mgr.SignUserOp(ctx, "bad-hex", &sa.UserOperation{})
	require.ErrorContains(t, err, "decode session key")

	decryptErr := errors.New("kms unavailable")
	encrypted := makeSessionKey("decrypt-error", "", false, time.Now().Add(time.Hour))
	encrypted.PrivateKeyRef = hex.EncodeToString([]byte("ciphertext"))
	require.NoError(t, store.Save(ctx, encrypted))

	mgr = NewManager(store, WithEncryption(nil, func(context.Context, string, []byte) ([]byte, error) {
		return nil, decryptErr
	}))
	_, err = mgr.SignUserOp(ctx, "decrypt-error", &sa.UserOperation{})
	require.ErrorIs(t, err, decryptErr)
	require.ErrorContains(t, err, "decrypt session key")
}

func TestManagerRevokeAllStopsOnOnChainError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()
	first := makeSessionKey("first", "", false, time.Now().Add(time.Hour))
	second := makeSessionKey("second", "", false, time.Now().Add(time.Hour))
	second.CreatedAt = first.CreatedAt.Add(time.Second)
	require.NoError(t, store.Save(ctx, first))
	require.NoError(t, store.Save(ctx, second))

	revokeErr := errors.New("rpc unavailable")
	var revoked []common.Address
	mgr := NewManager(store, WithOnChainRevocation(func(_ context.Context, addr common.Address) (string, error) {
		revoked = append(revoked, addr)
		return "", revokeErr
	}))

	err := mgr.RevokeAll(ctx)
	require.ErrorIs(t, err, revokeErr)
	require.ErrorContains(t, err, "revoke on-chain first")
	require.Equal(t, []common.Address{first.Address}, revoked)

	gotFirst, err := store.Get(ctx, "first")
	require.NoError(t, err)
	require.True(t, gotFirst.Revoked)
	gotSecond, err := store.Get(ctx, "second")
	require.NoError(t, err)
	require.False(t, gotSecond.Revoked)
}

func TestIntersectPoliciesCopiesParentOnlyConstraints(t *testing.T) {
	t.Parallel()

	parent := sa.SessionPolicy{
		AllowedTargets:   []common.Address{common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")},
		AllowedFunctions: []string{"0xaaaaaaaa"},
		SpendLimit:       big.NewInt(50),
		ValidAfter:       time.Unix(100, 0),
		ValidUntil:       time.Unix(200, 0),
	}
	child := sa.SessionPolicy{
		ValidAfter: time.Unix(50, 0),
		ValidUntil: time.Unix(300, 0),
	}

	got := intersectPolicies(parent, child)
	require.Equal(t, parent.ValidAfter, got.ValidAfter)
	require.Equal(t, parent.ValidUntil, got.ValidUntil)
	require.Equal(t, 0, got.SpendLimit.Cmp(big.NewInt(50)))
	require.Equal(t, parent.AllowedTargets, got.AllowedTargets)
	require.Equal(t, parent.AllowedFunctions, got.AllowedFunctions)

	parent.AllowedTargets[0] = common.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	parent.AllowedFunctions[0] = "0xbbbbbbbb"
	parent.SpendLimit.SetInt64(99)
	require.Equal(t, common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), got.AllowedTargets[0])
	require.Equal(t, "0xaaaaaaaa", got.AllowedFunctions[0])
	require.Equal(t, 0, got.SpendLimit.Cmp(big.NewInt(50)))
	require.False(t, bytes.Equal(parent.SpendLimit.Bytes(), got.SpendLimit.Bytes()))
}
