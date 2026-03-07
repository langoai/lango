package hub

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/contract"
)

func TestFactoryClient_CreateVault_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	vaultAddr := common.HexToAddress("0xVault")
	mc.writeResult = &contract.ContractCallResult{
		Data:   []interface{}{big.NewInt(0), vaultAddr},
		TxHash: "0xfactory",
	}

	client := NewFactoryClient(mc, common.HexToAddress("0xF"), 1)
	info, txHash, err := client.CreateVault(
		context.Background(),
		common.HexToAddress("0x2"),
		common.HexToAddress("0x3"),
		big.NewInt(1000),
		big.NewInt(9999),
		common.HexToAddress("0xA"),
	)

	require.NoError(t, err)
	assert.Equal(t, "0xfactory", txHash)
	assert.Equal(t, big.NewInt(0), info.VaultID)
	assert.Equal(t, vaultAddr, info.VaultAddress)
	assert.Equal(t, "createVault", mc.writeCalls[0].Method)
}

func TestFactoryClient_CreateVault_Error(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.writeErr = errors.New("fail")

	client := NewFactoryClient(mc, common.HexToAddress("0xF"), 1)
	_, _, err := client.CreateVault(
		context.Background(),
		common.HexToAddress("0x2"),
		common.HexToAddress("0x3"),
		big.NewInt(1000),
		big.NewInt(9999),
		common.HexToAddress("0xA"),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "create vault")
}

func TestFactoryClient_CreateVault_EmptyResult(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.writeResult = &contract.ContractCallResult{
		Data:   []interface{}{},
		TxHash: "0xempty",
	}

	client := NewFactoryClient(mc, common.HexToAddress("0xF"), 1)
	info, txHash, err := client.CreateVault(
		context.Background(),
		common.HexToAddress("0x2"),
		common.HexToAddress("0x3"),
		big.NewInt(1000),
		big.NewInt(9999),
		common.HexToAddress("0xA"),
	)

	require.NoError(t, err)
	assert.Equal(t, "0xempty", txHash)
	assert.Nil(t, info.VaultID)
}

func TestFactoryClient_GetVault_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	expected := common.HexToAddress("0xVaultAddr")
	mc.readResult = &contract.ContractCallResult{
		Data: []interface{}{expected},
	}

	client := NewFactoryClient(mc, common.HexToAddress("0xF"), 1)
	addr, err := client.GetVault(context.Background(), big.NewInt(0))

	require.NoError(t, err)
	assert.Equal(t, expected, addr)
	assert.Equal(t, "getVault", mc.readCalls[0].Method)
}

func TestFactoryClient_GetVault_Error(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.readErr = errors.New("fail")

	client := NewFactoryClient(mc, common.HexToAddress("0xF"), 1)
	_, err := client.GetVault(context.Background(), big.NewInt(0))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "get vault")
}

func TestFactoryClient_VaultCount_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.readResult = &contract.ContractCallResult{
		Data: []interface{}{big.NewInt(5)},
	}

	client := NewFactoryClient(mc, common.HexToAddress("0xF"), 1)
	count, err := client.VaultCount(context.Background())

	require.NoError(t, err)
	assert.Equal(t, big.NewInt(5), count)
	assert.Equal(t, "vaultCount", mc.readCalls[0].Method)
}

func TestFactoryClient_VaultCount_Error(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.readErr = errors.New("fail")

	client := NewFactoryClient(mc, common.HexToAddress("0xF"), 1)
	_, err := client.VaultCount(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "vault count")
}

func TestFactoryClient_VaultCount_EmptyResult(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.readResult = &contract.ContractCallResult{Data: []interface{}{}}

	client := NewFactoryClient(mc, common.HexToAddress("0xF"), 1)
	count, err := client.VaultCount(context.Background())

	require.NoError(t, err)
	assert.Equal(t, big.NewInt(0), count)
}

func TestFactoryClient_PassesCorrectChainID(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	client := NewFactoryClient(mc, common.HexToAddress("0xF"), 31337)

	_, _ = client.VaultCount(context.Background())

	require.Len(t, mc.readCalls, 1)
	assert.Equal(t, int64(31337), mc.readCalls[0].ChainID)
}
