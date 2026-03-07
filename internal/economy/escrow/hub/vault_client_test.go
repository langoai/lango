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

func TestVaultClient_Deposit_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	client := NewVaultClient(mc, common.HexToAddress("0xV"), 1)

	txHash, err := client.Deposit(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "0xmocktx", txHash)
	assert.Equal(t, "deposit", mc.writeCalls[0].Method)
}

func TestVaultClient_Deposit_Error(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.writeErr = errors.New("fail")

	client := NewVaultClient(mc, common.HexToAddress("0xV"), 1)
	_, err := client.Deposit(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vault deposit")
}

func TestVaultClient_SubmitWork_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	client := NewVaultClient(mc, common.HexToAddress("0xV"), 1)

	var wh [32]byte
	copy(wh[:], []byte("workhash"))
	txHash, err := client.SubmitWork(context.Background(), wh)
	require.NoError(t, err)
	assert.Equal(t, "0xmocktx", txHash)
	assert.Equal(t, "submitWork", mc.writeCalls[0].Method)
}

func TestVaultClient_SubmitWork_Error(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.writeErr = errors.New("fail")

	client := NewVaultClient(mc, common.HexToAddress("0xV"), 1)
	var wh [32]byte
	_, err := client.SubmitWork(context.Background(), wh)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vault submit work")
}

func TestVaultClient_Release_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	client := NewVaultClient(mc, common.HexToAddress("0xV"), 1)

	txHash, err := client.Release(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "0xmocktx", txHash)
	assert.Equal(t, "release", mc.writeCalls[0].Method)
}

func TestVaultClient_Refund_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	client := NewVaultClient(mc, common.HexToAddress("0xV"), 1)

	txHash, err := client.Refund(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "0xmocktx", txHash)
	assert.Equal(t, "refund", mc.writeCalls[0].Method)
}

func TestVaultClient_Dispute_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	client := NewVaultClient(mc, common.HexToAddress("0xV"), 1)

	txHash, err := client.Dispute(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "0xmocktx", txHash)
	assert.Equal(t, "dispute", mc.writeCalls[0].Method)
}

func TestVaultClient_Resolve_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	client := NewVaultClient(mc, common.HexToAddress("0xV"), 1)

	txHash, err := client.Resolve(context.Background(), true, big.NewInt(800), big.NewInt(200))
	require.NoError(t, err)
	assert.Equal(t, "0xmocktx", txHash)
	assert.Equal(t, "resolve", mc.writeCalls[0].Method)
}

func TestVaultClient_Status_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.readResult = &contract.ContractCallResult{
		Data: []interface{}{uint8(2)},
	}

	client := NewVaultClient(mc, common.HexToAddress("0xV"), 1)
	status, err := client.Status(context.Background())
	require.NoError(t, err)
	assert.Equal(t, DealStatusWorkSubmitted, status)
	assert.Equal(t, "status", mc.readCalls[0].Method)
}

func TestVaultClient_Status_Error(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.readErr = errors.New("fail")

	client := NewVaultClient(mc, common.HexToAddress("0xV"), 1)
	_, err := client.Status(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vault status")
}

func TestVaultClient_Amount_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.readResult = &contract.ContractCallResult{
		Data: []interface{}{big.NewInt(5000)},
	}

	client := NewVaultClient(mc, common.HexToAddress("0xV"), 1)
	amount, err := client.Amount(context.Background())
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(5000), amount)
	assert.Equal(t, "amount", mc.readCalls[0].Method)
}

func TestVaultClient_Amount_Error(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.readErr = errors.New("fail")

	client := NewVaultClient(mc, common.HexToAddress("0xV"), 1)
	_, err := client.Amount(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vault amount")
}
