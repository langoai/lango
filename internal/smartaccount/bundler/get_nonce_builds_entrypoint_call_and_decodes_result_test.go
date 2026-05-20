package bundler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func TestGetNonceBuildsEntrypointCallAndDecodesResult(t *testing.T) {
	t.Parallel()

	entryPoint := common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032")
	account := common.HexToAddress("0x1111111111111111111111111111111111111111")
	var captured jsonrpcRequest
	checks := &getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !decodeGetNonceBuildsEntrypointCallAndDecodesResultRequest(checks, w, r, &captured) {
			return
		}
		if captured.Method != "eth_call" {
			checks.fail(w, "method = %q, want eth_call", captured.Method)
			return
		}
		if len(captured.Params) != 2 {
			checks.fail(w, "params len = %d, want 2", len(captured.Params))
			return
		}
		callMsg, ok := captured.Params[0].(map[string]interface{})
		if !ok {
			checks.fail(w, "params[0] = %T, want object", captured.Params[0])
			return
		}
		if got := callMsg["to"]; got != entryPoint.Hex() {
			checks.fail(w, "to = %v, want %s", got, entryPoint.Hex())
			return
		}
		wantCalldata := expectedGetNonceCalldata(account)
		if got := callMsg["data"]; got != wantCalldata {
			checks.fail(w, "data = %v, want %s", got, wantCalldata)
			return
		}
		if got := captured.Params[1]; got != "latest" {
			checks.fail(w, "block tag = %v, want latest", got)
			return
		}

		writeGetNonceBuildsEntrypointCallAndDecodesResultRPCResult(checks, w, captured.ID, hexutil.EncodeBig(big.NewInt(42)))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, entryPoint)
	nonce, err := client.GetNonce(context.Background(), account)
	checks.requireClean(t)
	if err != nil {
		t.Fatalf("GetNonce returned error: %v", err)
	}
	if nonce.Cmp(big.NewInt(42)) != 0 {
		t.Fatalf("nonce = %s, want 42", nonce)
	}
}

func TestGetNoncePropagatesRPCAndDecodeErrors(t *testing.T) {
	t.Parallel()

	entryPoint := common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032")
	account := common.HexToAddress("0x1111111111111111111111111111111111111111")
	tests := []struct {
		name       string
		write      func(checks *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks, w http.ResponseWriter, req jsonrpcRequest)
		wantErr    string
		wantIsBund bool
	}{
		{
			name: "rpc error",
			write: func(checks *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks, w http.ResponseWriter, req jsonrpcRequest) {
				writeGetNonceBuildsEntrypointCallAndDecodesResultRPCError(checks, w, req.ID, -32000, "execution reverted", nil)
			},
			wantErr:    "get entrypoint nonce: bundler RPC error -32000: execution reverted",
			wantIsBund: true,
		},
		{
			name: "non string result",
			write: func(checks *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks, w http.ResponseWriter, req jsonrpcRequest) {
				writeGetNonceBuildsEntrypointCallAndDecodesResultRawResult(checks, w, req.ID, json.RawMessage(`123`))
			},
			wantErr: "decode nonce result",
		},
		{
			name: "invalid hex result",
			write: func(checks *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks, w http.ResponseWriter, req jsonrpcRequest) {
				writeGetNonceBuildsEntrypointCallAndDecodesResultRPCResult(checks, w, req.ID, "not-hex")
			},
			wantErr: "parse nonce",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			checks := &getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req jsonrpcRequest
				if !decodeGetNonceBuildsEntrypointCallAndDecodesResultRequest(checks, w, r, &req) {
					return
				}
				if req.Method != "eth_call" {
					checks.fail(w, "method = %q, want eth_call", req.Method)
					return
				}
				tt.write(checks, w, req)
			}))
			defer srv.Close()

			client := NewClient(srv.URL, entryPoint)
			_, err := client.GetNonce(context.Background(), account)
			checks.requireClean(t)
			if err == nil {
				t.Fatal("GetNonce returned nil error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want substring %q", err, tt.wantErr)
			}
			if tt.wantIsBund && !errors.Is(err, ErrBundlerError) {
				t.Fatalf("error = %v, want errors.Is ErrBundlerError", err)
			}
		})
	}
}

func TestGetGasFeesUsesPriorityFeeAndLatestBaseFee(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var calls []jsonrpcRequest
	checks := &getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req jsonrpcRequest
		if !decodeGetNonceBuildsEntrypointCallAndDecodesResultRequest(checks, w, r, &req) {
			return
		}
		mu.Lock()
		calls = append(calls, req)
		mu.Unlock()

		switch req.Method {
		case "eth_maxPriorityFeePerGas":
			if len(req.Params) != 0 {
				checks.fail(w, "priority fee params = %#v, want empty", req.Params)
				return
			}
			writeGetNonceBuildsEntrypointCallAndDecodesResultRPCResult(checks, w, req.ID, "0x3b9aca00")
		case "eth_getBlockByNumber":
			wantParams := []interface{}{"latest", false}
			if !reflect.DeepEqual(req.Params, wantParams) {
				checks.fail(w, "block params = %#v, want %#v", req.Params, wantParams)
				return
			}
			writeGetNonceBuildsEntrypointCallAndDecodesResultRPCResult(checks, w, req.ID, map[string]string{
				"baseFeePerGas": "0x77359400",
			})
		default:
			checks.fail(w, "unexpected method %q", req.Method)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, common.Address{})
	fees, err := client.GetGasFees(context.Background())
	checks.requireClean(t)
	if err != nil {
		t.Fatalf("GetGasFees returned error: %v", err)
	}
	assertBigIntString(t, "max priority fee", fees.MaxPriorityFeePerGas, "1000000000")
	assertBigIntString(t, "max fee", fees.MaxFeePerGas, "5000000000")

	mu.Lock()
	defer mu.Unlock()
	if got, want := len(calls), 2; got != want {
		t.Fatalf("call count = %d, want %d", got, want)
	}
	if calls[0].Method != "eth_maxPriorityFeePerGas" || calls[1].Method != "eth_getBlockByNumber" {
		t.Fatalf("methods = [%s %s], want priority then latest block", calls[0].Method, calls[1].Method)
	}
}

func TestGetGasFeesFallsBackForPriorityFeeAndBaseFee(t *testing.T) {
	t.Parallel()

	checks := &getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req jsonrpcRequest
		if !decodeGetNonceBuildsEntrypointCallAndDecodesResultRequest(checks, w, r, &req) {
			return
		}
		switch req.Method {
		case "eth_maxPriorityFeePerGas":
			writeGetNonceBuildsEntrypointCallAndDecodesResultRPCError(checks, w, req.ID, -32601, "method not found", nil)
		case "eth_getBlockByNumber":
			writeGetNonceBuildsEntrypointCallAndDecodesResultRPCResult(checks, w, req.ID, map[string]string{
				"baseFeePerGas": "not-hex",
			})
		default:
			checks.fail(w, "unexpected method %q", req.Method)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, common.Address{})
	fees, err := client.GetGasFees(context.Background())
	checks.requireClean(t)
	if err != nil {
		t.Fatalf("GetGasFees returned error: %v", err)
	}
	assertBigIntString(t, "fallback priority fee", fees.MaxPriorityFeePerGas, "1500000000")
	assertBigIntString(t, "fallback max fee", fees.MaxFeePerGas, "3500000000")
}

func TestGetGasFeesPropagatesLatestBlockErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		write   func(checks *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks, w http.ResponseWriter, req jsonrpcRequest)
		wantErr string
	}{
		{
			name: "rpc error",
			write: func(checks *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks, w http.ResponseWriter, req jsonrpcRequest) {
				if req.Method == "eth_maxPriorityFeePerGas" {
					writeGetNonceBuildsEntrypointCallAndDecodesResultRPCResult(checks, w, req.ID, "0x1")
					return
				}
				writeGetNonceBuildsEntrypointCallAndDecodesResultRPCError(checks, w, req.ID, -32000, "block unavailable", nil)
			},
			wantErr: "get latest block: bundler RPC error -32000: block unavailable",
		},
		{
			name: "decode error",
			write: func(checks *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks, w http.ResponseWriter, req jsonrpcRequest) {
				if req.Method == "eth_maxPriorityFeePerGas" {
					writeGetNonceBuildsEntrypointCallAndDecodesResultRPCResult(checks, w, req.ID, "0x1")
					return
				}
				writeGetNonceBuildsEntrypointCallAndDecodesResultRawResult(checks, w, req.ID, json.RawMessage(`123`))
			},
			wantErr: "decode block",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			checks := &getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req jsonrpcRequest
				if !decodeGetNonceBuildsEntrypointCallAndDecodesResultRequest(checks, w, r, &req) {
					return
				}
				tt.write(checks, w, req)
			}))
			defer srv.Close()

			client := NewClient(srv.URL, common.Address{})
			_, err := client.GetGasFees(context.Background())
			checks.requireClean(t)
			if err == nil {
				t.Fatal("GetGasFees returned nil error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestUserOpToMapSplitsV07FactoryAndPaymasterFields(t *testing.T) {
	t.Parallel()

	factory := common.HexToAddress("0x2222222222222222222222222222222222222222")
	paymaster := common.HexToAddress("0x3333333333333333333333333333333333333333")
	verificationGasBytes := common.LeftPadBytes(big.NewInt(123456789).Bytes(), 16)
	postOpGasBytes := common.LeftPadBytes(big.NewInt(987654321).Bytes(), 16)
	op := newTestOp()
	op.InitCode = append(factory.Bytes(), []byte{0xde, 0xad, 0xbe, 0xef}...)
	op.PaymasterAndData = append(paymaster.Bytes(), verificationGasBytes...)
	op.PaymasterAndData = append(op.PaymasterAndData, postOpGasBytes...)
	op.PaymasterAndData = append(op.PaymasterAndData, []byte{0xca, 0xfe}...)

	got := userOpToMap(op)

	assertMapValue(t, got, "sender", op.Sender.Hex())
	assertMapValue(t, got, "nonce", "0x1")
	assertMapValue(t, got, "callData", "0x0102")
	assertMapValue(t, got, "factory", factory.Hex())
	assertMapValue(t, got, "factoryData", "0xdeadbeef")
	assertMapValue(t, got, "paymaster", paymaster.Hex())
	assertMapValue(t, got, "paymasterVerificationGasLimit", "0x75bcd15")
	assertMapValue(t, got, "paymasterPostOpGasLimit", "0x3ade68b1")
	assertMapValue(t, got, "paymasterData", "0xcafe")
	if _, ok := got["initCode"]; ok {
		t.Fatal("map contains initCode; v0.7 should split factory fields")
	}
	if _, ok := got["paymasterAndData"]; ok {
		t.Fatal("map contains paymasterAndData; v0.7 should split paymaster fields")
	}
}

func TestUserOpToMapUsesEmptyV07FieldsForShortCompositeData(t *testing.T) {
	t.Parallel()

	op := newTestOp()
	op.Nonce = nil
	op.InitCode = []byte{0x01, 0x02, 0x03}
	op.PaymasterAndData = common.LeftPadBytes([]byte{0x01}, 51)

	got := userOpToMap(op)

	assertMapValue(t, got, "nonce", "0x0")
	assertMapValue(t, got, "factory", "0x")
	assertMapValue(t, got, "factoryData", "0x")
	assertMapValue(t, got, "paymaster", "0x")
	assertMapValue(t, got, "paymasterVerificationGasLimit", "0x0")
	assertMapValue(t, got, "paymasterPostOpGasLimit", "0x0")
	assertMapValue(t, got, "paymasterData", "0x")
}

func TestEncodeUint128Hex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []byte
		want string
	}{
		{name: "nil", in: nil, want: "0x0"},
		{name: "zero", in: make([]byte, 16), want: "0x0"},
		{name: "trims leading zeros", in: common.LeftPadBytes([]byte{0x12, 0x34}, 16), want: "0x1234"},
		{name: "full uint128", in: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, want: "0xffffffffffffffffffffffffffffffff"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := encodeUint128Hex(tt.in); got != tt.want {
				t.Fatalf("encodeUint128Hex(%x) = %s, want %s", tt.in, got, tt.want)
			}
		})
	}
}

func TestCallSurfacesHTTPDecodeAndRevertReasonErrors(t *testing.T) {
	t.Parallel()

	revertData, err := json.Marshal("0x08c379a00000000000000000000000000000000000000000000000000000000000000020" +
		"000000000000000000000000000000000000000000000000000000000000001843616c6c6572206973206e6f7420746865206f776e65720000000000000000")
	if err != nil {
		t.Fatalf("marshal revert data: %v", err)
	}

	tests := []struct {
		name       string
		handler    func(checks *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks) http.HandlerFunc
		wantErr    string
		wantIsBund bool
	}{
		{
			name: "http status",
			handler: func(*getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "upstream unavailable", http.StatusBadGateway)
				}
			},
			wantErr:    "bundler HTTP 502",
			wantIsBund: true,
		},
		{
			name: "invalid json response",
			handler: func(*getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte("{"))
				}
			},
			wantErr: "decode response",
		},
		{
			name: "rpc revert reason",
			handler: func(checks *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					var req jsonrpcRequest
					if !decodeGetNonceBuildsEntrypointCallAndDecodesResultRequest(checks, w, r, &req) {
						return
					}
					raw := json.RawMessage(revertData)
					writeGetNonceBuildsEntrypointCallAndDecodesResultRPCError(checks, w, req.ID, -32500, "execution reverted", &raw)
				}
			},
			wantErr:    "reason: Caller is not the owner",
			wantIsBund: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			checks := &getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks{}
			srv := httptest.NewServer(tt.handler(checks))
			defer srv.Close()
			client := NewClient(srv.URL, common.Address{})

			_, err := client.call(context.Background(), "eth_test", nil)
			checks.requireClean(t)
			if err == nil {
				t.Fatal("call returned nil error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want substring %q", err, tt.wantErr)
			}
			if tt.wantIsBund && !errors.Is(err, ErrBundlerError) {
				t.Fatalf("error = %v, want errors.Is ErrBundlerError", err)
			}
		})
	}
}

func expectedGetNonceCalldata(account common.Address) string {
	calldata := make([]byte, 0, 68)
	calldata = append(calldata, getNonceSelector...)
	addrPadded := make([]byte, 32)
	copy(addrPadded[12:], account.Bytes())
	calldata = append(calldata, addrPadded...)
	calldata = append(calldata, make([]byte, 32)...)
	return hexutil.Encode(calldata)
}

type getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks struct {
	mu       sync.Mutex
	failures []string
}

func (c *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks) fail(w http.ResponseWriter, format string, args ...interface{}) {
	c.mu.Lock()
	c.failures = append(c.failures, fmt.Sprintf(format, args...))
	c.mu.Unlock()
	http.Error(w, "runChatUsesProgramSeamAndCleansUpSession1 handler assertion failed", http.StatusInternalServerError)
}

func (c *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks) requireClean(t *testing.T) {
	t.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.failures) > 0 {
		t.Fatalf("handler assertions failed: %s", strings.Join(c.failures, "; "))
	}
}

func decodeGetNonceBuildsEntrypointCallAndDecodesResultRequest(checks *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks, w http.ResponseWriter, r *http.Request, req *jsonrpcRequest) bool {
	if r.Method != http.MethodPost {
		checks.fail(w, "method = %s, want POST", r.Method)
		return false
	}
	if got := r.Header.Get("Content-Type"); got != "application/json" {
		checks.fail(w, "content-type = %q, want application/json", got)
		return false
	}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		checks.fail(w, "decode request: %v", err)
		return false
	}
	if req.JSONRPC != "2.0" {
		checks.fail(w, "jsonrpc = %q, want 2.0", req.JSONRPC)
		return false
	}
	if req.ID <= 0 {
		checks.fail(w, "id = %d, want positive", req.ID)
		return false
	}
	return true
}

func writeGetNonceBuildsEntrypointCallAndDecodesResultRPCResult(checks *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks, w http.ResponseWriter, id int, result interface{}) {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		checks.fail(w, "marshal result: %v", err)
		return
	}
	writeGetNonceBuildsEntrypointCallAndDecodesResultRawResult(checks, w, id, resultJSON)
}

func writeGetNonceBuildsEntrypointCallAndDecodesResultRawResult(checks *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks, w http.ResponseWriter, id int, result json.RawMessage) {
	writeGetNonceBuildsEntrypointCallAndDecodesResultResponse(checks, w, jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func writeGetNonceBuildsEntrypointCallAndDecodesResultRPCError(checks *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks, w http.ResponseWriter, id, code int, message string, data *json.RawMessage) {
	writeGetNonceBuildsEntrypointCallAndDecodesResultResponse(checks, w, jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &jsonrpcError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	})
}

func writeGetNonceBuildsEntrypointCallAndDecodesResultResponse(checks *getNonceBuildsEntrypointCallAndDecodesResultHandlerChecks, w http.ResponseWriter, resp jsonrpcResponse) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		checks.fail(w, "encode response: %v", err)
	}
}

func assertBigIntString(t *testing.T, name string, got *big.Int, want string) {
	t.Helper()

	if got == nil {
		t.Fatalf("%s = nil, want %s", name, want)
	}
	if got.String() != want {
		t.Fatalf("%s = %s, want %s", name, got, want)
	}
}

func assertMapValue(t *testing.T, got map[string]interface{}, key string, want interface{}) {
	t.Helper()

	if got[key] != want {
		t.Fatalf("%s = %s, want %s", key, fmt.Sprint(got[key]), fmt.Sprint(want))
	}
}
