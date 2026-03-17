package permit

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// mockSigner implements PermitSigner for testing.
type mockSigner struct {
	key     *ecdsa.PrivateKey
	address common.Address
}

func newMockSigner(t *testing.T) *mockSigner {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return &mockSigner{key: key, address: addr}
}

func (m *mockSigner) SignTransaction(_ context.Context, rawTx []byte) ([]byte, error) {
	return crypto.Sign(rawTx, m.key)
}

func (m *mockSigner) Address(_ context.Context) (string, error) {
	return m.address.Hex(), nil
}

// mockCaller implements EthCaller for testing.
type mockCaller struct {
	result []byte
	err    error
}

func (m *mockCaller) CallContract(_ context.Context, _ ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	return m.result, m.err
}

func TestDomainSeparator(t *testing.T) {
	t.Parallel()

	chainID := int64(84532)
	usdcAddr := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")

	ds := DomainSeparator(chainID, usdcAddr)
	if len(ds) != 32 {
		t.Fatalf("want domain separator length 32, got %d", len(ds))
	}

	// Same inputs should produce the same separator.
	ds2 := DomainSeparator(chainID, usdcAddr)
	if !bytesEqual(ds, ds2) {
		t.Error("domain separator not deterministic")
	}

	// Different chain should produce different separator.
	ds3 := DomainSeparator(1, usdcAddr)
	if bytesEqual(ds, ds3) {
		t.Error("different chain IDs should produce different separators")
	}
}

func TestPermitStructHash(t *testing.T) {
	t.Parallel()

	owner := common.HexToAddress("0x1111111111111111111111111111111111111111")
	spender := common.HexToAddress("0x2222222222222222222222222222222222222222")
	value := big.NewInt(1000000)
	nonce := big.NewInt(0)
	deadline := big.NewInt(1700000000)

	hash := PermitStructHash(owner, spender, value, nonce, deadline)
	if len(hash) != 32 {
		t.Fatalf("want struct hash length 32, got %d", len(hash))
	}

	// Same inputs should produce the same hash.
	hash2 := PermitStructHash(owner, spender, value, nonce, deadline)
	if !bytesEqual(hash, hash2) {
		t.Error("struct hash not deterministic")
	}
}

func TestTypedDataHash(t *testing.T) {
	t.Parallel()

	domainSep := make([]byte, 32)
	domainSep[0] = 0xAA
	structHash := make([]byte, 32)
	structHash[0] = 0xBB

	hash := TypedDataHash(domainSep, structHash)
	if len(hash) != 32 {
		t.Fatalf("want typed data hash length 32, got %d", len(hash))
	}
}

func TestSign(t *testing.T) {
	t.Parallel()

	signer := newMockSigner(t)
	owner := signer.address
	spender := common.HexToAddress("0x31BE08D380A21fc740883c0BC434FcFc88740b58")
	value := big.NewInt(1000000)
	nonce := big.NewInt(0)
	deadline := big.NewInt(1700000000)
	chainID := int64(84532)
	usdcAddr := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")

	v, r, s, err := Sign(
		context.Background(), signer,
		owner, spender, value, nonce, deadline,
		chainID, usdcAddr,
	)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	// V should be 27 or 28.
	if v != 27 && v != 28 {
		t.Errorf("want V 27 or 28, got %d", v)
	}

	// R and S should be non-zero.
	zeroBytes := [32]byte{}
	if r == zeroBytes {
		t.Error("R is zero")
	}
	if s == zeroBytes {
		t.Error("S is zero")
	}

	// Verify signature recovery.
	domainSep := DomainSeparator(chainID, usdcAddr)
	structHash := PermitStructHash(owner, spender, value, nonce, deadline)
	hash := TypedDataHash(domainSep, structHash)

	var sig [65]byte
	copy(sig[:32], r[:])
	copy(sig[32:64], s[:])
	recV := v
	if recV >= 27 {
		recV -= 27
	}
	sig[64] = recV

	pubKey, err := crypto.Ecrecover(hash, sig[:])
	if err != nil {
		t.Fatalf("ecrecover: %v", err)
	}
	recovered := common.BytesToAddress(crypto.Keccak256(pubKey[1:])[12:])
	if recovered != owner {
		t.Errorf("recovered %s, want %s", recovered.Hex(), owner.Hex())
	}
}

func TestSign_InvalidSignatureLength(t *testing.T) {
	t.Parallel()

	badSigner := &shortSigSigner{}
	owner := common.HexToAddress("0x1111111111111111111111111111111111111111")
	spender := common.HexToAddress("0x2222222222222222222222222222222222222222")

	_, _, _, err := Sign(
		context.Background(), badSigner,
		owner, spender, big.NewInt(1), big.NewInt(0), big.NewInt(1700000000),
		84532, common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"),
	)
	if err == nil {
		t.Fatal("want error for short signature")
	}
}

type shortSigSigner struct{}

func (s *shortSigSigner) SignTransaction(_ context.Context, _ []byte) ([]byte, error) {
	return []byte{0x01, 0x02}, nil // too short
}

func (s *shortSigSigner) Address(_ context.Context) (string, error) {
	return "0x0000000000000000000000000000000000000000", nil
}

func TestGetPermitNonce(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		result    []byte
		err       error
		wantNonce int64
		wantErr   bool
	}{
		{
			give:      "nonce is zero",
			result:    make([]byte, 32),
			wantNonce: 0,
		},
		{
			give: "nonce is 5",
			result: func() []byte {
				b := make([]byte, 32)
				big.NewInt(5).FillBytes(b)
				return b
			}(),
			wantNonce: 5,
		},
		{
			give:    "result too short",
			result:  []byte{0x01},
			wantErr: true,
		},
		{
			give:    "call error",
			err:     context.DeadlineExceeded,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			caller := &mockCaller{result: tt.result, err: tt.err}
			usdcAddr := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")
			owner := common.HexToAddress("0x1111111111111111111111111111111111111111")

			nonce, err := GetPermitNonce(context.Background(), caller, usdcAddr, owner)
			if tt.wantErr {
				if err == nil {
					t.Fatal("want error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if nonce.Int64() != tt.wantNonce {
				t.Errorf("want nonce %d, got %d", tt.wantNonce, nonce.Int64())
			}
		})
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
