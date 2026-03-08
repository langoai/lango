package bundler

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func newTestOp() *UserOperation {
	return &UserOperation{
		Sender:               common.HexToAddress("0x1234"),
		Nonce:                big.NewInt(1),
		InitCode:             []byte{},
		CallData:             []byte{0x01, 0x02},
		CallGasLimit:         big.NewInt(100000),
		VerificationGasLimit: big.NewInt(50000),
		PreVerificationGas:   big.NewInt(21000),
		MaxFeePerGas:         big.NewInt(2000000000),
		MaxPriorityFeePerGas: big.NewInt(1000000000),
		PaymasterAndData:     []byte{},
		Signature:            []byte{0xAA, 0xBB},
	}
}

func TestSendUserOperation(t *testing.T) {
	t.Parallel()

	opHash := "0xabcdef1234567890abcdef1234567890" +
		"abcdef1234567890abcdef1234567890"

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req jsonrpcRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("decode request: %v", err)
				return
			}
			if req.Method != "eth_sendUserOperation" {
				t.Errorf(
					"want method eth_sendUserOperation, got %s",
					req.Method,
				)
			}
			resp := jsonrpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
			}
			hashJSON, _ := json.Marshal(opHash)
			resp.Result = hashJSON
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}),
	)
	defer srv.Close()

	entryPoint := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
	c := NewClient(srv.URL, entryPoint)

	result, err := c.SendUserOperation(
		context.Background(), newTestOp(),
	)
	if err != nil {
		t.Fatalf("send user op: %v", err)
	}
	if !result.Success {
		t.Error("want success=true")
	}
	if result.UserOpHash != common.HexToHash(opHash) {
		t.Errorf(
			"want hash %s, got %s",
			opHash, result.UserOpHash.Hex(),
		)
	}
}

func TestEstimateGas(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req jsonrpcRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("decode request: %v", err)
				return
			}
			if req.Method != "eth_estimateUserOperationGas" {
				t.Errorf(
					"want method eth_estimateUserOperationGas, got %s",
					req.Method,
				)
			}
			gasResult := map[string]string{
				"callGasLimit":         "0x186a0",
				"verificationGasLimit": "0xc350",
				"preVerificationGas":   "0x5208",
			}
			resultJSON, _ := json.Marshal(gasResult)
			resp := jsonrpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  resultJSON,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}),
	)
	defer srv.Close()

	entryPoint := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
	c := NewClient(srv.URL, entryPoint)

	estimate, err := c.EstimateGas(
		context.Background(), newTestOp(),
	)
	if err != nil {
		t.Fatalf("estimate gas: %v", err)
	}
	if estimate.CallGasLimit.Int64() != 100000 {
		t.Errorf(
			"want callGasLimit 100000, got %d",
			estimate.CallGasLimit.Int64(),
		)
	}
	if estimate.VerificationGasLimit.Int64() != 50000 {
		t.Errorf(
			"want verificationGasLimit 50000, got %d",
			estimate.VerificationGasLimit.Int64(),
		)
	}
	if estimate.PreVerificationGas.Int64() != 21000 {
		t.Errorf(
			"want preVerificationGas 21000, got %d",
			estimate.PreVerificationGas.Int64(),
		)
	}
}

func TestSendUserOperationRPCError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req jsonrpcRequest
			json.NewDecoder(r.Body).Decode(&req)
			resp := jsonrpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &jsonrpcError{
					Code:    -32602,
					Message: "invalid params",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}),
	)
	defer srv.Close()

	entryPoint := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
	c := NewClient(srv.URL, entryPoint)

	_, err := c.SendUserOperation(
		context.Background(), newTestOp(),
	)
	if err == nil {
		t.Fatal("want error for RPC error response")
	}
}

func TestGetUserOperationReceipt(t *testing.T) {
	t.Parallel()

	opHash := "0xabcdef1234567890abcdef1234567890" +
		"abcdef1234567890abcdef1234567890"
	txHash := "0x1111111111111111111111111111111111111111" +
		"111111111111111111111111"

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receipt := map[string]interface{}{
				"userOpHash":      opHash,
				"transactionHash": txHash,
				"success":         true,
				"actualGasUsed":   "0x5208",
			}
			resultJSON, _ := json.Marshal(receipt)
			var req jsonrpcRequest
			json.NewDecoder(r.Body).Decode(&req)
			resp := jsonrpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  resultJSON,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}),
	)
	defer srv.Close()

	entryPoint := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
	c := NewClient(srv.URL, entryPoint)

	result, err := c.GetUserOperationReceipt(
		context.Background(),
		common.HexToHash(opHash),
	)
	if err != nil {
		t.Fatalf("get receipt: %v", err)
	}
	if !result.Success {
		t.Error("want success=true")
	}
	if result.GasUsed != 21000 {
		t.Errorf("want gasUsed 21000, got %d", result.GasUsed)
	}
}

func TestSupportedEntryPoints(t *testing.T) {
	t.Parallel()

	ep := "0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789"
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req jsonrpcRequest
			json.NewDecoder(r.Body).Decode(&req)
			addrs := []string{ep}
			resultJSON, _ := json.Marshal(addrs)
			resp := jsonrpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  resultJSON,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}),
	)
	defer srv.Close()

	c := NewClient(
		srv.URL,
		common.HexToAddress(ep),
	)
	addrs, err := c.SupportedEntryPoints(context.Background())
	if err != nil {
		t.Fatalf("supported entry points: %v", err)
	}
	if len(addrs) != 1 {
		t.Fatalf("want 1 entry point, got %d", len(addrs))
	}
	if addrs[0] != common.HexToAddress(ep) {
		t.Errorf(
			"want %s, got %s", ep, addrs[0].Hex(),
		)
	}
}

func TestSendUserOperationNilOp(t *testing.T) {
	t.Parallel()

	c := NewClient(
		"http://localhost:1234",
		common.Address{},
	)
	_, err := c.SendUserOperation(
		context.Background(), nil,
	)
	if err == nil {
		t.Fatal("want error for nil op")
	}
}
