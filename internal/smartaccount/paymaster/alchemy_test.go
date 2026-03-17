package paymaster

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestAlchemyProvider_SponsorUserOp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			give: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var req jsonrpcRequest
				json.NewDecoder(r.Body).Decode(&req)
				if req.Method != "alchemy_requestGasAndPaymasterAndData" {
					t.Errorf("want method alchemy_requestGasAndPaymasterAndData, got %s", req.Method)
				}

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
						"message": "policy not found",
					},
				})
			},
			wantErr: true,
		},
		{
			give: "HTTP error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte("bad gateway"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			provider := NewAlchemyProvider(srv.URL, "policy_123")

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
			if result.GasOverrides == nil {
				t.Error("want gas overrides from alchemy")
			}
		})
	}
}

func TestAlchemyProvider_Type(t *testing.T) {
	t.Parallel()
	p := NewAlchemyProvider("http://localhost", "")
	if p.Type() != "alchemy" {
		t.Errorf("want type 'alchemy', got %q", p.Type())
	}
}
