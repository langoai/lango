package smartaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/smartaccount/bundler"
)

// mockWallet implements wallet.WalletProvider for testing.
type mockWallet struct {
	addr string
}

func (w *mockWallet) Address(_ context.Context) (string, error) {
	return w.addr, nil
}

func (w *mockWallet) Balance(_ context.Context) (*big.Int, error) {
	return big.NewInt(1000000000000000000), nil
}

func (w *mockWallet) SignTransaction(
	_ context.Context, _ []byte,
) ([]byte, error) {
	return make([]byte, 65), nil
}

func (w *mockWallet) SignMessage(
	_ context.Context, _ []byte,
) ([]byte, error) {
	return make([]byte, 65), nil
}

func (w *mockWallet) PublicKey(
	_ context.Context,
) ([]byte, error) {
	return make([]byte, 33), nil
}

func TestNewManager(t *testing.T) {
	t.Parallel()

	entryPoint := common.HexToAddress(
		"0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789",
	)
	wp := &mockWallet{
		addr: "0x1234567890abcdef1234567890abcdef12345678",
	}

	// Create a mock bundler server.
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  "0x0",
			})
		}),
	)
	defer srv.Close()

	bundlerClient := bundler.NewClient(srv.URL, entryPoint)

	m := NewManager(
		nil, // factory (not used in this test)
		bundlerClient,
		nil, // caller (not used in this test)
		wp,
		84532, // Base Sepolia
		entryPoint,
	)

	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.chainID != 84532 {
		t.Errorf("want chainID 84532, got %d", m.chainID)
	}
	if m.entryPoint != entryPoint {
		t.Errorf(
			"want entryPoint %s, got %s",
			entryPoint.Hex(), m.entryPoint.Hex(),
		)
	}
}

func TestManagerInstallModuleNotDeployed(t *testing.T) {
	t.Parallel()

	m := &Manager{
		modules: make([]ModuleInfo, 0),
	}

	_, err := m.InstallModule(
		context.Background(),
		ModuleTypeValidator,
		common.HexToAddress("0x1234"),
		nil,
	)
	if err != ErrAccountNotDeployed {
		t.Errorf(
			"want ErrAccountNotDeployed, got %v", err,
		)
	}
}

func TestManagerUninstallModuleNotFound(t *testing.T) {
	t.Parallel()

	m := &Manager{
		accountAddr: common.HexToAddress("0xABCD"),
		modules:     make([]ModuleInfo, 0),
	}

	_, err := m.UninstallModule(
		context.Background(),
		ModuleTypeValidator,
		common.HexToAddress("0x1234"),
		nil,
	)
	if err != ErrModuleNotInstalled {
		t.Errorf(
			"want ErrModuleNotInstalled, got %v", err,
		)
	}
}

func TestManagerExecuteEmpty(t *testing.T) {
	t.Parallel()

	m := &Manager{
		accountAddr: common.HexToAddress("0xABCD"),
		modules:     make([]ModuleInfo, 0),
	}

	_, err := m.Execute(
		context.Background(), []ContractCall{},
	)
	if err == nil {
		t.Fatal("want error for empty calls")
	}
}

func TestManagerExecuteNotDeployed(t *testing.T) {
	t.Parallel()

	m := &Manager{
		modules: make([]ModuleInfo, 0),
	}

	_, err := m.Execute(
		context.Background(),
		[]ContractCall{{
			Target: common.HexToAddress("0x1234"),
			Data:   []byte{0x01},
		}},
	)
	if err != ErrAccountNotDeployed {
		t.Errorf(
			"want ErrAccountNotDeployed, got %v", err,
		)
	}
}

func TestComputeUserOpHash(t *testing.T) {
	t.Parallel()

	entryPoint := common.HexToAddress(
		"0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789",
	)
	m := &Manager{
		chainID:    84532,
		entryPoint: entryPoint,
	}

	op := &UserOperation{
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
	}

	hash := m.computeUserOpHash(op)
	if len(hash) != 32 {
		t.Errorf("want 32-byte hash, got %d bytes", len(hash))
	}

	// Hash should be deterministic.
	hash2 := m.computeUserOpHash(op)
	if string(hash) != string(hash2) {
		t.Error("hash is not deterministic")
	}
}

func TestFactoryComputeAddress(t *testing.T) {
	t.Parallel()

	f := NewFactory(
		nil, // caller not used for compute
		nil, // rpc not used for compute
		common.HexToAddress("0xAAAA"),
		common.HexToAddress("0xBBBB"),
		common.HexToAddress("0xCCCC"),
		84532,
	)

	owner := common.HexToAddress(
		"0x1234567890abcdef1234567890abcdef12345678",
	)
	addr1 := f.ComputeAddress(owner, big.NewInt(0))
	addr2 := f.ComputeAddress(owner, big.NewInt(0))

	// Same inputs should produce same address.
	if addr1 != addr2 {
		t.Errorf(
			"deterministic address mismatch: %s != %s",
			addr1.Hex(), addr2.Hex(),
		)
	}

	// Different salt should produce different address.
	addr3 := f.ComputeAddress(owner, big.NewInt(1))
	if addr1 == addr3 {
		t.Error(
			"different salts should produce different addresses",
		)
	}
}

func TestSubmitUserOp_NoPaymaster(t *testing.T) {
	t.Parallel()

	// Mock bundler: getNonce → estimateGas → send
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		w.Header().Set("Content-Type", "application/json")
		callCount++

		switch req.Method {
		case "eth_call":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      callCount,
				"result":  "0x0000000000000000000000000000000000000000000000000000000000000005",
			})
		case "eth_maxPriorityFeePerGas":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      callCount,
				"result":  "0x59682f00",
			})
		case "eth_getBlockByNumber":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      callCount,
				"result": map[string]interface{}{
					"baseFeePerGas": "0x3b9aca00",
				},
			})
		case "eth_estimateUserOperationGas":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      callCount,
				"result": map[string]interface{}{
					"callGasLimit":         "0x30d40",
					"verificationGasLimit": "0x186a0",
					"preVerificationGas":   "0x5208",
				},
			})
		case "eth_sendUserOperation":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      callCount,
				"result":  "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			})
		}
	}))
	defer srv.Close()

	entryPoint := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
	wp := &mockWallet{addr: "0x1234567890abcdef1234567890abcdef12345678"}
	bundlerClient := bundler.NewClient(srv.URL, entryPoint)

	m := NewManager(nil, bundlerClient, nil, wp, 84532, entryPoint)
	m.accountAddr = common.HexToAddress("0xABCD")

	// No paymaster set — should use existing flow
	txHash, err := m.Execute(context.Background(), []ContractCall{{
		Target: common.HexToAddress("0x1111"),
		Data:   []byte{0x01, 0x02},
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txHash == "" {
		t.Error("want non-empty txHash")
	}
}

func TestSubmitUserOp_PaymasterTwoPhase(t *testing.T) {
	t.Parallel()

	stubCalled := false
	finalCalled := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case "eth_call":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  "0x0000000000000000000000000000000000000000000000000000000000000000",
			})
		case "eth_maxPriorityFeePerGas":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  "0x59682f00",
			})
		case "eth_getBlockByNumber":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result": map[string]interface{}{
					"baseFeePerGas": "0x3b9aca00",
				},
			})
		case "eth_estimateUserOperationGas":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result": map[string]interface{}{
					"callGasLimit":         "0x30d40",
					"verificationGasLimit": "0x186a0",
					"preVerificationGas":   "0x5208",
				},
			})
		case "eth_sendUserOperation":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      2,
				"result":  "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			})
		}
	}))
	defer srv.Close()

	entryPoint := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
	wp := &mockWallet{addr: "0x1234567890abcdef1234567890abcdef12345678"}
	bundlerClient := bundler.NewClient(srv.URL, entryPoint)

	m := NewManager(nil, bundlerClient, nil, wp, 84532, entryPoint)
	m.accountAddr = common.HexToAddress("0xABCD")

	stubPMData := make([]byte, 20)
	finalPMData := append(make([]byte, 20), 0x01, 0x02)

	m.SetPaymasterFunc(func(ctx context.Context, op *UserOperation, stub bool) ([]byte, *PaymasterGasOverrides, error) {
		if stub {
			stubCalled = true
			return stubPMData, nil, nil
		}
		finalCalled = true
		return finalPMData, nil, nil
	})

	txHash, err := m.Execute(context.Background(), []ContractCall{{
		Target: common.HexToAddress("0x1111"),
		Data:   []byte{0x01},
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txHash == "" {
		t.Error("want non-empty txHash")
	}
	if !stubCalled {
		t.Error("paymaster stub phase was not called")
	}
	if !finalCalled {
		t.Error("paymaster final phase was not called")
	}
}

func TestSubmitUserOp_PaymasterStubFails(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case "eth_getBlockByNumber":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result": map[string]interface{}{
					"baseFeePerGas": "0x3b9aca00",
				},
			})
		default:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0", "id": 1, "result": "0x0000000000000000000000000000000000000000000000000000000000000000",
			})
		}
	}))
	defer srv.Close()

	entryPoint := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
	wp := &mockWallet{addr: "0x1234567890abcdef1234567890abcdef12345678"}
	bundlerClient := bundler.NewClient(srv.URL, entryPoint)

	m := NewManager(nil, bundlerClient, nil, wp, 84532, entryPoint)
	m.accountAddr = common.HexToAddress("0xABCD")

	m.SetPaymasterFunc(func(ctx context.Context, op *UserOperation, stub bool) ([]byte, *PaymasterGasOverrides, error) {
		if stub {
			return nil, nil, fmt.Errorf("stub error: insufficient USDC")
		}
		return nil, nil, nil
	})

	_, err := m.Execute(context.Background(), []ContractCall{{
		Target: common.HexToAddress("0x1111"),
		Data:   []byte{0x01},
	}})
	if err == nil {
		t.Fatal("want error when paymaster stub fails")
	}
}

func TestSubmitUserOp_PaymasterFinalFails(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case "eth_estimateUserOperationGas":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result": map[string]interface{}{
					"callGasLimit":         "0x30d40",
					"verificationGasLimit": "0x186a0",
					"preVerificationGas":   "0x5208",
				},
			})
		case "eth_getBlockByNumber":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result": map[string]interface{}{
					"baseFeePerGas": "0x3b9aca00",
				},
			})
		default:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0", "id": 1, "result": "0x0000000000000000000000000000000000000000000000000000000000000000",
			})
		}
	}))
	defer srv.Close()

	entryPoint := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
	wp := &mockWallet{addr: "0x1234567890abcdef1234567890abcdef12345678"}
	bundlerClient := bundler.NewClient(srv.URL, entryPoint)

	m := NewManager(nil, bundlerClient, nil, wp, 84532, entryPoint)
	m.accountAddr = common.HexToAddress("0xABCD")

	m.SetPaymasterFunc(func(ctx context.Context, op *UserOperation, stub bool) ([]byte, *PaymasterGasOverrides, error) {
		if stub {
			return make([]byte, 20), nil, nil
		}
		return nil, nil, fmt.Errorf("final error: paymaster rejected")
	})

	_, err := m.Execute(context.Background(), []ContractCall{{
		Target: common.HexToAddress("0x1111"),
		Data:   []byte{0x01},
	}})
	if err == nil {
		t.Fatal("want error when paymaster final fails")
	}
}

func TestSubmitUserOp_PaymasterGasOverrides(t *testing.T) {
	t.Parallel()

	var sentOp map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case "eth_call":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  "0x0000000000000000000000000000000000000000000000000000000000000003",
			})
		case "eth_maxPriorityFeePerGas":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result":  "0x59682f00",
			})
		case "eth_getBlockByNumber":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result": map[string]interface{}{
					"baseFeePerGas": "0x3b9aca00",
				},
			})
		case "eth_estimateUserOperationGas":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"result": map[string]interface{}{
					"callGasLimit":         "0x30d40",
					"verificationGasLimit": "0x186a0",
					"preVerificationGas":   "0x5208",
				},
			})
		case "eth_sendUserOperation":
			// Capture the sent operation
			if len(req.Params) > 0 {
				json.Unmarshal(req.Params[0], &sentOp)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      2,
				"result":  "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			})
		}
	}))
	defer srv.Close()

	entryPoint := common.HexToAddress("0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789")
	wp := &mockWallet{addr: "0x1234567890abcdef1234567890abcdef12345678"}
	bundlerClient := bundler.NewClient(srv.URL, entryPoint)

	m := NewManager(nil, bundlerClient, nil, wp, 84532, entryPoint)
	m.accountAddr = common.HexToAddress("0xABCD")

	overriddenCallGas := big.NewInt(500000)

	m.SetPaymasterFunc(func(ctx context.Context, op *UserOperation, stub bool) ([]byte, *PaymasterGasOverrides, error) {
		if stub {
			return make([]byte, 20), nil, nil
		}
		return make([]byte, 22), &PaymasterGasOverrides{
			CallGasLimit: overriddenCallGas,
		}, nil
	})

	txHash, err := m.Execute(context.Background(), []ContractCall{{
		Target: common.HexToAddress("0x1111"),
		Data:   []byte{0x01},
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txHash == "" {
		t.Error("want non-empty txHash")
	}
}

func TestManagerModuleAlreadyInstalled(t *testing.T) {
	t.Parallel()

	moduleAddr := common.HexToAddress("0x9999")
	m := &Manager{
		accountAddr: common.HexToAddress("0xABCD"),
		modules: []ModuleInfo{
			{
				Address: moduleAddr,
				Type:    ModuleTypeValidator,
			},
		},
	}

	_, err := m.InstallModule(
		context.Background(),
		ModuleTypeValidator,
		moduleAddr,
		nil,
	)
	if err != ErrModuleAlreadyInstalled {
		t.Errorf(
			"want ErrModuleAlreadyInstalled, got %v", err,
		)
	}
}
