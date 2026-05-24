package smartaccount

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/contract"
	"github.com/langoai/lango/internal/smartaccount/bindings"
	"github.com/langoai/lango/internal/smartaccount/bundler"
)

func TestManagerInstallModuleSubmitsAndRejectsDuplicate(t *testing.T) {
	t.Parallel()

	m := newModuleSubmitManager(t)
	moduleAddr := common.HexToAddress("0x9999999999999999999999999999999999999999")

	txHash, err := m.InstallModule(context.Background(), ModuleTypeValidator, moduleAddr, []byte{0x01})
	require.NoError(t, err)
	require.Equal(t, "0x2222222222222222222222222222222222222222222222222222222222222222", txHash)
	require.Len(t, m.modules, 1)
	require.Equal(t, moduleAddr, m.modules[0].Address)
	require.Equal(t, ModuleTypeValidator, m.modules[0].Type)
	require.Equal(t, "validator", m.modules[0].Name)

	_, err = m.InstallModule(context.Background(), ModuleTypeValidator, moduleAddr, nil)
	require.ErrorIs(t, err, ErrModuleAlreadyInstalled)
}

func TestManagerUninstallModuleSubmitsAndRemovesOnlyMatch(t *testing.T) {
	t.Parallel()

	m := newModuleSubmitManager(t)
	target := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	other := common.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	m.modules = []ModuleInfo{
		{Address: target, Type: ModuleTypeExecutor, Name: "executor"},
		{Address: other, Type: ModuleTypeHook, Name: "hook"},
	}

	txHash, err := m.UninstallModule(context.Background(), ModuleTypeExecutor, target, []byte{0x02})
	require.NoError(t, err)
	require.Equal(t, "0x2222222222222222222222222222222222222222222222222222222222222222", txHash)
	require.Len(t, m.modules, 1)
	require.Equal(t, other, m.modules[0].Address)
	require.Equal(t, ModuleTypeHook, m.modules[0].Type)
}

func TestManagerEncodeCallsUsesSingleAndBatchModes(t *testing.T) {
	t.Parallel()

	m := &Manager{}
	singleTarget := common.HexToAddress("0x1111111111111111111111111111111111111111")
	single, err := m.encodeCalls([]ContractCall{{
		Target: singleTarget,
		Value:  big.NewInt(3),
		Data:   []byte{0xaa},
	}})
	require.NoError(t, err)

	parsed, err := contract.ParseABI(bindings.Safe7579ABI)
	require.NoError(t, err)
	execute := parsed.Methods["execute"]
	require.True(t, bytes.Equal(single[:4], execute.ID))
	values, err := execute.Inputs.Unpack(single[4:])
	require.NoError(t, err)
	singleMode := values[0].([32]byte)
	require.Zero(t, singleMode[0])

	batch, err := m.encodeCalls([]ContractCall{
		{Target: singleTarget, Data: []byte{0xaa}},
		{Target: common.HexToAddress("0x2222222222222222222222222222222222222222"), Data: []byte{0xbb}},
	})
	require.NoError(t, err)
	values, err = execute.Inputs.Unpack(batch[4:])
	require.NoError(t, err)
	batchMode := values[0].([32]byte)
	require.Equal(t, byte(0x01), batchMode[0])
}

func newModuleSubmitManager(t *testing.T) *Manager {
	t.Helper()

	entryPoint := common.HexToAddress("0x0000000071727De22E5E9d8BAf0edAc6f37da032")
	handlerErr := make(chan error, 8)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     int    `json:"id"`
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			handlerErr <- err
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		var err error
		switch req.Method {
		case "eth_call":
			err = writeRPCResult(w, req.ID, "0x00")
		case "eth_maxPriorityFeePerGas":
			err = writeRPCResult(w, req.ID, "0x59682f00")
		case "eth_getBlockByNumber":
			err = writeRPCResult(w, req.ID, map[string]any{"baseFeePerGas": "0x3b9aca00"})
		case "eth_estimateUserOperationGas":
			err = writeRPCResult(w, req.ID, map[string]any{
				"callGasLimit":         "0x30d40",
				"verificationGasLimit": "0x186a0",
				"preVerificationGas":   "0x5208",
			})
		case "eth_sendUserOperation":
			err = writeRPCResult(w, req.ID, "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
		case "eth_getUserOperationReceipt":
			err = writeRPCResult(w, req.ID, map[string]any{
				"userOpHash":      "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				"transactionHash": "0x2222222222222222222222222222222222222222222222222222222222222222",
				"success":         true,
				"actualGasUsed":   "0x5208",
			})
		default:
			err = fmt.Errorf("unexpected bundler method %q", req.Method)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if err != nil {
			handlerErr <- err
		}
	}))
	t.Cleanup(func() {
		srv.Close()
		close(handlerErr)
		for err := range handlerErr {
			require.NoError(t, err)
		}
	})

	m := NewManager(
		nil,
		bundler.NewClient(srv.URL, entryPoint),
		nil,
		&mockWallet{addr: "0x1234567890abcdef1234567890abcdef12345678"},
		84532,
		entryPoint,
	)
	m.accountAddr = common.HexToAddress("0xabcdabcdabcdabcdabcdabcdabcdabcdabcdabcd")
	return m
}

func writeRPCResult(w http.ResponseWriter, id int, result any) error {
	return json.NewEncoder(w).Encode(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	})
}
