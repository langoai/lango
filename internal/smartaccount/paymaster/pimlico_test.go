package paymaster

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestPimlicoProvider_SponsorUserOp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give     string
		policyID string
		handler  http.HandlerFunc
		wantErr  bool
	}{
		{
			give: "success without policy",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Verify request body
				var req jsonrpcRequest
				json.NewDecoder(r.Body).Decode(&req)
				if req.Method != "pm_sponsorUserOperation" {
					t.Errorf("want method pm_sponsorUserOperation, got %s", req.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      1,
					"result": map[string]interface{}{
						"paymasterAndData": "0xaabbccddaabbccddaabbccddaabbccddaabbccdd0011",
					},
				})
			},
		},
		{
			give:     "success with policy ID",
			policyID: "sp_test_123",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var req jsonrpcRequest
				json.NewDecoder(r.Body).Decode(&req)

				// Should have 3 params (opMap, entryPoint, context)
				if len(req.Params) != 3 {
					t.Errorf("want 3 params with policy, got %d", len(req.Params))
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      1,
					"result": map[string]interface{}{
						"paymasterAndData": "0xaabbccddaabbccddaabbccddaabbccddaabbccdd0011",
					},
				})
			},
		},
		{
			give: "RPC error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      1,
					"error": map[string]interface{}{
						"code":    -32601,
						"message": "method not found",
					},
				})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			provider := NewPimlicoProvider(srv.URL, tt.policyID)

			req := &SponsorRequest{
				UserOp:     testUserOp(),
				EntryPoint: common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032"),
				ChainID:    84532,
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
			if len(result.PaymasterAndData) == 0 {
				t.Error("want non-empty paymasterAndData")
			}
		})
	}
}

func TestPimlicoProvider_Type(t *testing.T) {
	t.Parallel()
	p := NewPimlicoProvider("http://localhost", "")
	if p.Type() != "pimlico" {
		t.Errorf("want type 'pimlico', got %q", p.Type())
	}
}

func testUserOp() *UserOpData {
	return &UserOpData{
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
	}
}
