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

func TestHubClient_CreateDeal_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.writeResult = &contract.ContractCallResult{
		Data:   []interface{}{big.NewInt(42)},
		TxHash: "0xabc",
	}

	client := NewHubClient(mc, common.HexToAddress("0x1"), 1)
	dealID, txHash, err := client.CreateDeal(
		context.Background(),
		common.HexToAddress("0x2"),
		common.HexToAddress("0x3"),
		big.NewInt(1000),
		big.NewInt(9999),
	)

	require.NoError(t, err)
	assert.Equal(t, big.NewInt(42), dealID)
	assert.Equal(t, "0xabc", txHash)
	assert.Len(t, mc.writeCalls, 1)
	assert.Equal(t, "createDeal", mc.writeCalls[0].Method)
}

func TestHubClient_CreateDeal_Error(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.writeErr = errors.New("rpc down")

	client := NewHubClient(mc, common.HexToAddress("0x1"), 1)
	_, _, err := client.CreateDeal(
		context.Background(),
		common.HexToAddress("0x2"),
		common.HexToAddress("0x3"),
		big.NewInt(1000),
		big.NewInt(9999),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "create deal")
}

func TestHubClient_Deposit_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.writeResult = &contract.ContractCallResult{TxHash: "0xdep"}

	client := NewHubClient(mc, common.HexToAddress("0x1"), 1)
	txHash, err := client.Deposit(context.Background(), big.NewInt(0))

	require.NoError(t, err)
	assert.Equal(t, "0xdep", txHash)
	assert.Equal(t, "deposit", mc.writeCalls[0].Method)
}

func TestHubClient_Deposit_Error(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.writeErr = errors.New("fail")

	client := NewHubClient(mc, common.HexToAddress("0x1"), 1)
	_, err := client.Deposit(context.Background(), big.NewInt(0))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "deposit deal")
}

func TestHubClient_SubmitWork_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	client := NewHubClient(mc, common.HexToAddress("0x1"), 1)

	var wh [32]byte
	copy(wh[:], []byte("workhash"))
	txHash, err := client.SubmitWork(context.Background(), big.NewInt(0), wh)

	require.NoError(t, err)
	assert.Equal(t, "0xmocktx", txHash)
	assert.Equal(t, "submitWork", mc.writeCalls[0].Method)
}

func TestHubClient_Release_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	client := NewHubClient(mc, common.HexToAddress("0x1"), 1)

	txHash, err := client.Release(context.Background(), big.NewInt(5))
	require.NoError(t, err)
	assert.Equal(t, "0xmocktx", txHash)
	assert.Equal(t, "release", mc.writeCalls[0].Method)
}

func TestHubClient_Refund_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	client := NewHubClient(mc, common.HexToAddress("0x1"), 1)

	txHash, err := client.Refund(context.Background(), big.NewInt(5))
	require.NoError(t, err)
	assert.Equal(t, "0xmocktx", txHash)
	assert.Equal(t, "refund", mc.writeCalls[0].Method)
}

func TestHubClient_Dispute_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	client := NewHubClient(mc, common.HexToAddress("0x1"), 1)

	txHash, err := client.Dispute(context.Background(), big.NewInt(5))
	require.NoError(t, err)
	assert.Equal(t, "0xmocktx", txHash)
	assert.Equal(t, "dispute", mc.writeCalls[0].Method)
}

func TestHubClient_ResolveDispute_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	client := NewHubClient(mc, common.HexToAddress("0x1"), 1)

	txHash, err := client.ResolveDispute(context.Background(), big.NewInt(5), true, big.NewInt(800), big.NewInt(200))
	require.NoError(t, err)
	assert.Equal(t, "0xmocktx", txHash)
	assert.Equal(t, "resolveDispute", mc.writeCalls[0].Method)
}

func TestHubClient_GetDeal_Error(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.readErr = errors.New("network error")

	client := NewHubClient(mc, common.HexToAddress("0x1"), 1)
	_, err := client.GetDeal(context.Background(), big.NewInt(0))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "get deal")
}

func TestHubClient_NextDealID_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.readResult = &contract.ContractCallResult{
		Data: []interface{}{big.NewInt(7)},
	}

	client := NewHubClient(mc, common.HexToAddress("0x1"), 1)
	id, err := client.NextDealID(context.Background())

	require.NoError(t, err)
	assert.Equal(t, big.NewInt(7), id)
	assert.Equal(t, "nextDealId", mc.readCalls[0].Method)
}

func TestHubClient_NextDealID_EmptyResult(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.readResult = &contract.ContractCallResult{Data: []interface{}{}}

	client := NewHubClient(mc, common.HexToAddress("0x1"), 1)
	id, err := client.NextDealID(context.Background())

	require.NoError(t, err)
	assert.Equal(t, big.NewInt(0), id)
}

func TestHubClient_WriteMethods_PassCorrectArgs(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	client := NewHubClient(mc, common.HexToAddress("0x1"), 31337)

	seller := common.HexToAddress("0x2")
	token := common.HexToAddress("0x3")
	amount := big.NewInt(5000)
	dl := big.NewInt(12345)

	_, _, _ = client.CreateDeal(context.Background(), seller, token, amount, dl)

	require.Len(t, mc.writeCalls, 1)
	call := mc.writeCalls[0]
	assert.Equal(t, int64(31337), call.ChainID)
	assert.Equal(t, common.HexToAddress("0x1"), call.Address)
	assert.Len(t, call.Args, 4)
}
