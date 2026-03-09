package paymaster

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
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
				EntryPoint: common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789"),
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
		EntryPoint: common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789"),
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
