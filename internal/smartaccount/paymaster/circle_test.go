package paymaster

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestCircleProvider_SponsorUserOp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		handler    http.HandlerFunc
		wantErr    bool
		wantPMLen  int
		wantGasOvr bool
	}{
		{
			give: "success with paymasterAndData only",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      1,
					"result": map[string]interface{}{
						"paymasterAndData": "0xaabbccddaabbccddaabbccddaabbccddaabbccdd0011223344",
					},
				})
			},
			wantPMLen:  25,
			wantGasOvr: false,
		},
		{
			give: "success with gas overrides",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      1,
					"result": map[string]interface{}{
						"paymasterAndData":     "0xaabbccddaabbccddaabbccddaabbccddaabbccdd0011223344",
						"callGasLimit":         "0x30d40",
						"verificationGasLimit": "0x186a0",
						"preVerificationGas":   "0x5208",
					},
				})
			},
			wantPMLen:  25,
			wantGasOvr: true,
		},
		{
			give: "RPC error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      1,
					"error": map[string]interface{}{
						"code":    -32000,
						"message": "insufficient USDC balance",
					},
				})
			},
			wantErr: true,
		},
		{
			give: "HTTP error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal server error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			provider := NewCircleProvider(srv.URL)

			req := &SponsorRequest{
				UserOp: &UserOpData{
					Sender:               common.HexToAddress("0x1234"),
					Nonce:                big.NewInt(1),
					InitCode:             []byte{},
					CallData:             []byte{0x01},
					CallGasLimit:         big.NewInt(100000),
					VerificationGasLimit: big.NewInt(50000),
					PreVerificationGas:   big.NewInt(21000),
					MaxFeePerGas:         big.NewInt(2000000000),
					MaxPriorityFeePerGas: big.NewInt(1000000000),
					PaymasterAndData:     []byte{},
					Signature:            []byte{},
				},
				EntryPoint: common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032"),
				ChainID:    84532,
				Stub:       false,
			}

			result, err := provider.SponsorUserOp(context.Background(), req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("want error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.PaymasterAndData) != tt.wantPMLen {
				t.Errorf("want paymasterAndData len %d, got %d", tt.wantPMLen, len(result.PaymasterAndData))
			}
			if tt.wantGasOvr && result.GasOverrides == nil {
				t.Error("want gas overrides, got nil")
			}
			if !tt.wantGasOvr && result.GasOverrides != nil {
				t.Error("want no gas overrides, got non-nil")
			}
		})
	}
}

func TestCircleProvider_Timeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer srv.Close()

	provider := &CircleProvider{
		url:        srv.URL,
		httpClient: &http.Client{Timeout: 100 * time.Millisecond},
	}

	req := &SponsorRequest{
		UserOp: &UserOpData{
			Sender:               common.HexToAddress("0x1234"),
			Nonce:                big.NewInt(0),
			InitCode:             []byte{},
			CallData:             []byte{},
			CallGasLimit:         big.NewInt(0),
			VerificationGasLimit: big.NewInt(0),
			PreVerificationGas:   big.NewInt(0),
			MaxFeePerGas:         big.NewInt(0),
			MaxPriorityFeePerGas: big.NewInt(0),
			PaymasterAndData:     []byte{},
			Signature:            []byte{},
		},
		EntryPoint: common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032"),
		ChainID:    84532,
	}

	_, err := provider.SponsorUserOp(context.Background(), req)
	if err == nil {
		t.Fatal("want timeout error, got nil")
	}
}

func TestCircleProvider_Type(t *testing.T) {
	t.Parallel()
	p := NewCircleProvider("http://localhost")
	if p.Type() != "circle" {
		t.Errorf("want type 'circle', got %q", p.Type())
	}
}

func TestBuildApproveCalldata(t *testing.T) {
	t.Parallel()

	spender := common.HexToAddress("0xaabbccddaabbccddaabbccddaabbccddaabbccdd")
	amount := big.NewInt(1000000) // 1 USDC

	data := BuildApproveCalldata(spender, amount)

	// Should be 4 (selector) + 32 (address) + 32 (amount) = 68 bytes
	if len(data) != 68 {
		t.Fatalf("want calldata len 68, got %d", len(data))
	}

	// First 4 bytes should be approve selector
	wantSelector := []byte{0x09, 0x5e, 0xa7, 0xb3}
	for i := 0; i < 4; i++ {
		if data[i] != wantSelector[i] {
			t.Errorf("selector byte %d: want 0x%02x, got 0x%02x", i, wantSelector[i], data[i])
		}
	}
}

func TestNewApprovalCall(t *testing.T) {
	t.Parallel()

	token := common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")
	pm := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	amount := big.NewInt(1000000000) // 1000 USDC

	call := NewApprovalCall(token, pm, amount)

	if call.TokenAddress != token {
		t.Errorf("want token %s, got %s", token.Hex(), call.TokenAddress.Hex())
	}
	if call.PaymasterAddr != pm {
		t.Errorf("want paymaster %s, got %s", pm.Hex(), call.PaymasterAddr.Hex())
	}
	if call.Amount.Cmp(amount) != 0 {
		t.Errorf("want amount %s, got %s", amount.String(), call.Amount.String())
	}
	if len(call.ApproveCalldata) != 68 {
		t.Errorf("want calldata len 68, got %d", len(call.ApproveCalldata))
	}
}

// --- CirclePermitProvider tests ---

type testPermitSigner struct {
	key     *ecdsa.PrivateKey
	address common.Address
}

func newTestPermitSigner(t *testing.T) *testPermitSigner {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return &testPermitSigner{
		key:     key,
		address: crypto.PubkeyToAddress(key.PublicKey),
	}
}

func (s *testPermitSigner) SignTransaction(_ context.Context, rawTx []byte) ([]byte, error) {
	return crypto.Sign(rawTx, s.key)
}

func (s *testPermitSigner) Address(_ context.Context) (string, error) {
	return s.address.Hex(), nil
}

type testEthCaller struct {
	nonce *big.Int
}

func (c *testEthCaller) CallContract(_ context.Context, _ ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	result := make([]byte, 32)
	if c.nonce != nil {
		c.nonce.FillBytes(result)
	}
	return result, nil
}

func TestCirclePermitProvider_Type(t *testing.T) {
	t.Parallel()

	p := NewCirclePermitProvider(
		common.HexToAddress("0x31BE08D380A21fc740883c0BC434FcFc88740b58"),
		common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"),
		84532,
		newTestPermitSigner(t),
		&testEthCaller{},
	)
	if p.Type() != "circle-permit" {
		t.Errorf("want type 'circle-permit', got %q", p.Type())
	}
}

func TestCirclePermitProvider_SponsorUserOp_Stub(t *testing.T) {
	t.Parallel()

	signer := newTestPermitSigner(t)
	p := NewCirclePermitProvider(
		common.HexToAddress("0x31BE08D380A21fc740883c0BC434FcFc88740b58"),
		common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"),
		84532,
		signer,
		&testEthCaller{},
	)

	req := &SponsorRequest{
		UserOp: &UserOpData{
			Sender:               common.HexToAddress("0x1234"),
			Nonce:                big.NewInt(0),
			InitCode:             []byte{},
			CallData:             []byte{0x01},
			CallGasLimit:         big.NewInt(100000),
			VerificationGasLimit: big.NewInt(50000),
			PreVerificationGas:   big.NewInt(21000),
			MaxFeePerGas:         big.NewInt(2000000000),
			MaxPriorityFeePerGas: big.NewInt(1000000000),
			PaymasterAndData:     []byte{},
			Signature:            []byte{},
		},
		EntryPoint: common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032"),
		ChainID:    84532,
		Stub:       true,
	}

	result, err := p.SponsorUserOp(context.Background(), req)
	if err != nil {
		t.Fatalf("stub sponsor: %v", err)
	}

	// Stub PaymasterAndData: prefix(52) + paymasterData(118) = 170 bytes.
	if len(result.PaymasterAndData) != 170 {
		t.Errorf("want stub PaymasterAndData len 170, got %d", len(result.PaymasterAndData))
	}

	// First 20 bytes should be the paymaster address.
	gotAddr := common.BytesToAddress(result.PaymasterAndData[:20])
	wantAddr := common.HexToAddress("0x31BE08D380A21fc740883c0BC434FcFc88740b58")
	if gotAddr != wantAddr {
		t.Errorf("want paymaster addr %s, got %s", wantAddr.Hex(), gotAddr.Hex())
	}
}

func TestCirclePermitProvider_SponsorUserOp_Real(t *testing.T) {
	t.Parallel()

	signer := newTestPermitSigner(t)
	caller := &testEthCaller{nonce: big.NewInt(3)}
	p := NewCirclePermitProvider(
		common.HexToAddress("0x31BE08D380A21fc740883c0BC434FcFc88740b58"),
		common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"),
		84532,
		signer,
		caller,
	)

	req := &SponsorRequest{
		UserOp: &UserOpData{
			Sender:               signer.address,
			Nonce:                big.NewInt(0),
			InitCode:             []byte{},
			CallData:             []byte{0x01},
			CallGasLimit:         big.NewInt(100000),
			VerificationGasLimit: big.NewInt(50000),
			PreVerificationGas:   big.NewInt(21000),
			MaxFeePerGas:         big.NewInt(2000000000),
			MaxPriorityFeePerGas: big.NewInt(1000000000),
			PaymasterAndData:     []byte{},
			Signature:            []byte{},
		},
		EntryPoint: common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032"),
		ChainID:    84532,
		Stub:       false,
	}

	result, err := p.SponsorUserOp(context.Background(), req)
	if err != nil {
		t.Fatalf("real sponsor: %v", err)
	}

	// Full PaymasterAndData: prefix(52) + paymasterData(118) = 170 bytes.
	if len(result.PaymasterAndData) != 170 {
		t.Errorf("want PaymasterAndData len 170, got %d", len(result.PaymasterAndData))
	}

	pmd := result.PaymasterAndData

	// Verify paymaster address.
	gotAddr := common.BytesToAddress(pmd[:20])
	wantAddr := common.HexToAddress("0x31BE08D380A21fc740883c0BC434FcFc88740b58")
	if gotAddr != wantAddr {
		t.Errorf("want paymaster addr %s, got %s", wantAddr.Hex(), gotAddr.Hex())
	}

	// Verify mode byte.
	if pmd[52] != 0x01 {
		t.Errorf("want mode 0x01, got 0x%02x", pmd[52])
	}

	// Verify token address in paymasterData.
	gotToken := common.BytesToAddress(pmd[53:73])
	wantToken := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")
	if gotToken != wantToken {
		t.Errorf("want token addr %s, got %s", wantToken.Hex(), gotToken.Hex())
	}

	// Verify amount (10 USDC = 10_000_000 in 6 decimals).
	amount := new(big.Int).SetBytes(pmd[73:105])
	if amount.Int64() != 10_000_000 {
		t.Errorf("want amount 10000000, got %d", amount.Int64())
	}

	// Signature should be non-zero (65 bytes at offset 105).
	sig := pmd[105:170]
	allZero := true
	for _, b := range sig {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("signature should not be all zeros")
	}
}
