package smartaccount

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/contract"
	"github.com/langoai/lango/internal/smartaccount/bindings"
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

type addressErrorWallet struct {
	mockWallet
	err error
}

func (w *addressErrorWallet) Address(_ context.Context) (string, error) {
	return "", w.err
}

func newEthCodeClient(t *testing.T, code string) (*ethclient.Client, func()) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     interface{} `json:"id"`
			Method string      `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode JSON-RPC request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if req.Method != "eth_getCode" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"error": map[string]interface{}{
					"code":    -32601,
					"message": "method not found",
				},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  code,
		})
	}))

	client, err := ethclient.Dial(srv.URL)
	if err != nil {
		srv.Close()
		t.Fatalf("dial ethclient: %v", err)
	}
	return client, func() {
		client.Close()
		srv.Close()
	}
}

func newManagerTestFactory(rpc *ethclient.Client, caller *stubContractCaller) *Factory {
	if caller == nil {
		caller = &stubContractCaller{}
	}
	return NewFactory(
		caller,
		rpc,
		common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		common.HexToAddress("0x1111111111111111111111111111111111111111"),
		common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
		common.HexToAddress("0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC"),
		84532,
	)
}

func TestNewManager(t *testing.T) {
	t.Parallel()

	entryPoint := common.HexToAddress(
		"0x0000000071727De22E5E9d8BAf0edAc6f37da032",
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

func TestManagerOwnerAddressReturnsWalletAddress(t *testing.T) {
	t.Parallel()

	want := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	m := &Manager{wallet: &mockWallet{addr: want.Hex()}}

	got, err := m.ownerAddress(context.Background())
	if err != nil {
		t.Fatalf("ownerAddress: %v", err)
	}
	if got != want {
		t.Fatalf("ownerAddress = %s, want %s", got.Hex(), want.Hex())
	}
}

func TestManagerOwnerAddressWrapsWalletError(t *testing.T) {
	t.Parallel()

	walletErr := errors.New("wallet locked")
	m := &Manager{wallet: &addressErrorWallet{err: walletErr}}

	_, err := m.ownerAddress(context.Background())
	if err == nil {
		t.Fatal("ownerAddress error = nil, want wallet error")
	}
	if !strings.Contains(err.Error(), "get owner address") || !errors.Is(err, walletErr) {
		t.Fatalf("ownerAddress error = %v, want wrapped wallet error", err)
	}
}

func TestManagerBuildInfoCopiesModules(t *testing.T) {
	t.Parallel()

	installedAt := time.Unix(123, 0)
	moduleAddr := common.HexToAddress("0x9999999999999999999999999999999999999999")
	accountAddr := common.HexToAddress("0xABCD")
	entryPoint := common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032")
	owner := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	m := &Manager{
		accountAddr: accountAddr,
		chainID:     84532,
		entryPoint:  entryPoint,
		modules: []ModuleInfo{{
			Address:     moduleAddr,
			Type:        ModuleTypeValidator,
			Name:        "validator",
			InstalledAt: installedAt,
		}},
	}

	info := m.buildInfo(owner, true)
	if info.Address != accountAddr || info.OwnerAddress != owner || info.ChainID != 84532 || info.EntryPoint != entryPoint || !info.IsDeployed {
		t.Fatalf("buildInfo returned unexpected account metadata: %+v", info)
	}
	if len(info.Modules) != 1 || info.Modules[0].Address != moduleAddr || info.Modules[0].InstalledAt != installedAt {
		t.Fatalf("buildInfo modules = %+v, want installed module copy", info.Modules)
	}

	info.Modules[0].Name = "mutated"
	info.Modules = append(info.Modules, ModuleInfo{Address: common.HexToAddress("0x1111")})
	if len(m.modules) != 1 {
		t.Fatalf("manager modules length = %d, want 1", len(m.modules))
	}
	if m.modules[0].Name != "validator" {
		t.Fatalf("manager module name = %q, want validator", m.modules[0].Name)
	}
}

func TestManagerPackSafe7579Call(t *testing.T) {
	t.Parallel()

	m := &Manager{}
	moduleAddr := common.HexToAddress("0x9999999999999999999999999999999999999999")
	initData := []byte{0x01, 0x02}

	data, err := m.packSafe7579Call("installModule", big.NewInt(int64(ModuleTypeValidator)), moduleAddr, initData)
	if err != nil {
		t.Fatalf("packSafe7579Call: %v", err)
	}
	parsed, err := contract.ParseABI(bindings.Safe7579ABI)
	if err != nil {
		t.Fatalf("ParseABI: %v", err)
	}
	method := parsed.Methods["installModule"]
	if !bytes.Equal(data[:4], method.ID) {
		t.Fatalf("method selector = %x, want %x", data[:4], method.ID)
	}

	values, err := method.Inputs.Unpack(data[4:])
	if err != nil {
		t.Fatalf("unpack installModule call: %v", err)
	}
	if values[0].(*big.Int).Cmp(big.NewInt(int64(ModuleTypeValidator))) != 0 {
		t.Fatalf("module type = %v, want %d", values[0], ModuleTypeValidator)
	}
	if values[1].(common.Address) != moduleAddr {
		t.Fatalf("module address = %s, want %s", values[1].(common.Address).Hex(), moduleAddr.Hex())
	}
	if !bytes.Equal(values[2].([]byte), initData) {
		t.Fatalf("initData = %x, want %x", values[2].([]byte), initData)
	}
}

func TestManagerPackSafe7579CallRejectsUnknownMethod(t *testing.T) {
	t.Parallel()

	_, err := (&Manager{}).packSafe7579Call("missingMethod")
	if err == nil {
		t.Fatal("packSafe7579Call error = nil, want unknown method error")
	}
	if !strings.Contains(err.Error(), "pack missingMethod call") {
		t.Fatalf("packSafe7579Call error = %v, want packed method context", err)
	}
}

func TestManagerEncodeBatchCallsPacksBatchExecutionModeAndPayload(t *testing.T) {
	t.Parallel()

	firstTarget := common.HexToAddress("0x1111111111111111111111111111111111111111")
	secondTarget := common.HexToAddress("0x2222222222222222222222222222222222222222")
	calls := []ContractCall{
		{Target: firstTarget, Value: big.NewInt(7), Data: []byte{0xaa, 0xbb}},
		{Target: secondTarget, Data: []byte{0xcc}},
	}

	data, err := (&Manager{}).encodeBatchCalls(calls)
	if err != nil {
		t.Fatalf("encodeBatchCalls: %v", err)
	}
	parsed, err := contract.ParseABI(bindings.Safe7579ABI)
	if err != nil {
		t.Fatalf("ParseABI: %v", err)
	}
	method := parsed.Methods["execute"]
	if !bytes.Equal(data[:4], method.ID) {
		t.Fatalf("method selector = %x, want %x", data[:4], method.ID)
	}
	values, err := method.Inputs.Unpack(data[4:])
	if err != nil {
		t.Fatalf("unpack execute call: %v", err)
	}
	mode := values[0].([32]byte)
	if mode[0] != 0x01 {
		t.Fatalf("batch mode first byte = 0x%x, want 0x01", mode[0])
	}
	payload := values[1].([]byte)
	firstLength := 32 + 32 + 32 + len(calls[0].Data)
	require.Len(t, payload, firstLength+32+32+32+len(calls[1].Data))

	assertBatchPayloadEntry(t, payload[:firstLength], firstTarget, big.NewInt(7), []byte{0xaa, 0xbb})
	assertBatchPayloadEntry(t, payload[firstLength:], secondTarget, big.NewInt(0), []byte{0xcc})
}

func assertBatchPayloadEntry(t *testing.T, payload []byte, target common.Address, value *big.Int, data []byte) {
	t.Helper()

	require.Len(t, payload, 32+32+32+len(data))
	if got := common.BytesToAddress(payload[12:32]); got != target {
		t.Fatalf("batch target = %s, want %s", got.Hex(), target.Hex())
	}
	if got := new(big.Int).SetBytes(payload[32:64]); got.Cmp(value) != 0 {
		t.Fatalf("batch value = %s, want %s", got, value)
	}
	if got := new(big.Int).SetBytes(payload[64:96]); got.Cmp(big.NewInt(int64(len(data)))) != 0 {
		t.Fatalf("batch calldata length = %s, want %d", got, len(data))
	}
	if !bytes.Equal(payload[96:], data) {
		t.Fatalf("batch calldata = %x, want %x", payload[96:], data)
	}
}

func TestManagerInfoComputesAddressAndReportsDeployment(t *testing.T) {
	t.Parallel()

	rpc, cleanup := newEthCodeClient(t, "0x6000")
	defer cleanup()
	caller := &stubContractCaller{}
	entryPoint := common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032")
	owner := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	m := NewManager(
		newManagerTestFactory(rpc, caller),
		nil,
		nil,
		&mockWallet{addr: owner.Hex()},
		84532,
		entryPoint,
	)

	info, err := m.Info(context.Background())
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if info.Address == (common.Address{}) {
		t.Fatal("Info address is zero, want computed account address")
	}
	if !info.IsDeployed {
		t.Fatal("Info IsDeployed = false, want true")
	}
	if info.OwnerAddress != owner || info.ChainID != 84532 || info.EntryPoint != entryPoint {
		t.Fatalf("Info metadata = %+v, want owner/chain/entry point populated", info)
	}
	if caller.readCalls != 1 {
		t.Fatalf("factory readCalls = %d, want 1 proxyCreationCode call", caller.readCalls)
	}
}

func TestManagerInfoWrapsDeploymentCheckError(t *testing.T) {
	t.Parallel()

	caller := &stubContractCaller{}
	m := NewManager(
		newManagerTestFactory(nil, caller),
		nil,
		nil,
		&mockWallet{addr: "0x1234567890abcdef1234567890abcdef12345678"},
		84532,
		common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032"),
	)

	_, err := m.Info(context.Background())
	if err == nil {
		t.Fatal("Info error = nil, want rpc configuration error")
	}
	if !strings.Contains(err.Error(), "check deployment") || !strings.Contains(err.Error(), "rpc client not configured") {
		t.Fatalf("Info error = %v, want wrapped deployment check error", err)
	}
}

func TestManagerGetOrDeployReturnsCachedDeployedAccount(t *testing.T) {
	t.Parallel()

	rpc, cleanup := newEthCodeClient(t, "0x6000")
	defer cleanup()
	caller := &stubContractCaller{}
	cached := common.HexToAddress("0x9999999999999999999999999999999999999999")
	owner := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	m := NewManager(
		newManagerTestFactory(rpc, caller),
		nil,
		nil,
		&mockWallet{addr: owner.Hex()},
		84532,
		common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032"),
	)
	m.accountAddr = cached

	info, err := m.GetOrDeploy(context.Background())
	if err != nil {
		t.Fatalf("GetOrDeploy: %v", err)
	}
	if info.Address != cached || !info.IsDeployed || info.OwnerAddress != owner {
		t.Fatalf("GetOrDeploy info = %+v, want cached deployed account", info)
	}
	if caller.readCalls != 0 || caller.writeCalls != 0 {
		t.Fatalf("factory calls read=%d write=%d, want no compute/deploy calls", caller.readCalls, caller.writeCalls)
	}
}

func TestManagerGetOrDeployErrorsWhenDeployVerificationFindsNoCode(t *testing.T) {
	t.Parallel()

	rpc, cleanup := newEthCodeClient(t, "0x")
	defer cleanup()
	deployedAddr := common.HexToAddress("0x8888888888888888888888888888888888888888")
	caller := &stubContractCaller{
		writeResult: &contract.ContractCallResult{
			TxHash: "0xabc",
			Data:   []interface{}{deployedAddr},
		},
	}
	m := NewManager(
		newManagerTestFactory(rpc, caller),
		nil,
		nil,
		&mockWallet{addr: "0x1234567890abcdef1234567890abcdef12345678"},
		84532,
		common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032"),
	)

	_, err := m.GetOrDeploy(context.Background())
	if err == nil {
		t.Fatal("GetOrDeploy error = nil, want verification error")
	}
	if !strings.Contains(err.Error(), "on-chain verification found no code") {
		t.Fatalf("GetOrDeploy error = %v, want verification error", err)
	}
	if caller.writeCalls != 1 {
		t.Fatalf("factory writeCalls = %d, want 1 deploy call", caller.writeCalls)
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
		"0x0000000071727De22E5E9d8BAf0edAc6f37da032",
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
		&stubContractCaller{},         // stub for proxyCreationCode
		nil,                           // rpc not used for compute
		common.HexToAddress("0xAAAA"), // factory
		common.HexToAddress("0x1111"), // singleton
		common.HexToAddress("0xBBBB"), // safe7579
		common.HexToAddress("0xCCCC"), // fallback
		84532,
	)

	owner := common.HexToAddress(
		"0x1234567890abcdef1234567890abcdef12345678",
	)
	ctx := context.Background()
	addr1, err := f.ComputeAddress(ctx, owner, big.NewInt(0))
	if err != nil {
		t.Fatalf("ComputeAddress: %v", err)
	}
	addr2, err := f.ComputeAddress(ctx, owner, big.NewInt(0))
	if err != nil {
		t.Fatalf("ComputeAddress: %v", err)
	}

	// Same inputs should produce same address.
	if addr1 != addr2 {
		t.Errorf(
			"deterministic address mismatch: %s != %s",
			addr1.Hex(), addr2.Hex(),
		)
	}

	// Different salt should produce different address.
	addr3, err := f.ComputeAddress(ctx, owner, big.NewInt(1))
	if err != nil {
		t.Fatalf("ComputeAddress: %v", err)
	}
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
		case "eth_getUserOperationReceipt":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      callCount,
				"result": map[string]interface{}{
					"userOpHash":      "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
					"transactionHash": "0x1111111111111111111111111111111111111111111111111111111111111111",
					"success":         true,
					"actualGasUsed":   "0x5208",
				},
			})
		}
	}))
	defer srv.Close()

	entryPoint := common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032")
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
	// txHash should now be the on-chain tx hash, not the userOp hash
	if txHash != "0x1111111111111111111111111111111111111111111111111111111111111111" {
		t.Errorf("want on-chain txHash, got %s", txHash)
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
		case "eth_getUserOperationReceipt":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      3,
				"result": map[string]interface{}{
					"userOpHash":      "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
					"transactionHash": "0x2222222222222222222222222222222222222222222222222222222222222222",
					"success":         true,
					"actualGasUsed":   "0x5208",
				},
			})
		}
	}))
	defer srv.Close()

	entryPoint := common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032")
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

	entryPoint := common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032")
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

	entryPoint := common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032")
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
		case "eth_getUserOperationReceipt":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      3,
				"result": map[string]interface{}{
					"userOpHash":      "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
					"transactionHash": "0x3333333333333333333333333333333333333333333333333333333333333333",
					"success":         true,
					"actualGasUsed":   "0x5208",
				},
			})
		}
	}))
	defer srv.Close()

	entryPoint := common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032")
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
