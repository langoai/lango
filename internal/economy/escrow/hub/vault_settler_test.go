package hub

import (
	"context"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/contract"
	"github.com/langoai/lango/internal/economy/escrow"
)

func TestVaultSettler_InterfaceCompliance(t *testing.T) {
	t.Parallel()
	var _ escrow.SettlementExecutor = (*VaultSettler)(nil)
}

func TestVaultSettler_SetAndGetVaultMapping(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewVaultSettler(mc,
		common.HexToAddress("0xF"),
		common.HexToAddress("0xI"),
		common.HexToAddress("0xT"),
		common.HexToAddress("0xA"),
		1,
	)

	vaultAddr := common.HexToAddress("0xVAULT")
	s.SetVaultMapping("esc-1", vaultAddr)

	addr, ok := s.GetVaultAddress("esc-1")
	require.True(t, ok)
	assert.Equal(t, vaultAddr, addr)
}

func TestVaultSettler_GetVaultAddress_NotFound(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewVaultSettler(mc,
		common.HexToAddress("0xF"),
		common.HexToAddress("0xI"),
		common.HexToAddress("0xT"),
		common.HexToAddress("0xA"),
		1,
	)

	_, ok := s.GetVaultAddress("nonexistent")
	assert.False(t, ok)
}

func TestVaultSettler_Lock_NoOp(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewVaultSettler(mc,
		common.HexToAddress("0xF"),
		common.HexToAddress("0xI"),
		common.HexToAddress("0xT"),
		common.HexToAddress("0xA"),
		1,
	)

	err := s.Lock(context.Background(), "did:test:buyer", big.NewInt(1000))
	require.NoError(t, err)
}

func TestVaultSettler_Release_NoOp(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewVaultSettler(mc,
		common.HexToAddress("0xF"),
		common.HexToAddress("0xI"),
		common.HexToAddress("0xT"),
		common.HexToAddress("0xA"),
		1,
	)

	err := s.Release(context.Background(), "did:test:seller", big.NewInt(1000))
	require.NoError(t, err)
}

func TestVaultSettler_Refund_NoOp(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewVaultSettler(mc,
		common.HexToAddress("0xF"),
		common.HexToAddress("0xI"),
		common.HexToAddress("0xT"),
		common.HexToAddress("0xA"),
		1,
	)

	err := s.Refund(context.Background(), "did:test:buyer", big.NewInt(1000))
	require.NoError(t, err)
}

func TestVaultSettler_CreateVault_Success(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	vaultAddr := common.HexToAddress("0xNEWVAULT")
	mc.writeResult = &contract.ContractCallResult{
		Data:   []interface{}{big.NewInt(0), vaultAddr},
		TxHash: "0xfactory",
	}

	s := NewVaultSettler(mc,
		common.HexToAddress("0xF"),
		common.HexToAddress("0xI"),
		common.HexToAddress("0xT"),
		common.HexToAddress("0xA"),
		1,
	)

	addr, txHash, err := s.CreateVault(
		context.Background(),
		common.HexToAddress("0xSELLER"),
		big.NewInt(1000),
		big.NewInt(9999),
	)

	require.NoError(t, err)
	assert.Equal(t, vaultAddr, addr)
	assert.Equal(t, "0xfactory", txHash)
}

func TestVaultSettler_VaultClientFor(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewVaultSettler(mc,
		common.HexToAddress("0xF"),
		common.HexToAddress("0xI"),
		common.HexToAddress("0xT"),
		common.HexToAddress("0xA"),
		1,
	)

	vc := s.VaultClientFor(common.HexToAddress("0xV"))
	assert.NotNil(t, vc)
}

func TestVaultSettler_FactoryClient_Accessor(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewVaultSettler(mc,
		common.HexToAddress("0xF"),
		common.HexToAddress("0xI"),
		common.HexToAddress("0xT"),
		common.HexToAddress("0xA"),
		1,
	)

	fc := s.FactoryClient()
	assert.NotNil(t, fc)
}

func TestVaultSettler_TokenAddress(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	tokenAddr := common.HexToAddress("0xTOKEN")
	s := NewVaultSettler(mc,
		common.HexToAddress("0xF"),
		common.HexToAddress("0xI"),
		tokenAddr,
		common.HexToAddress("0xA"),
		1,
	)

	assert.Equal(t, tokenAddr, s.TokenAddress())
}

func TestVaultSettler_ConcurrentMapping(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewVaultSettler(mc,
		common.HexToAddress("0xF"),
		common.HexToAddress("0xI"),
		common.HexToAddress("0xT"),
		common.HexToAddress("0xA"),
		1,
	)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			addr := common.BigToAddress(big.NewInt(int64(n)))
			key := "esc-concurrent"
			s.SetVaultMapping(key, addr)
			s.GetVaultAddress(key)
		}(i)
	}
	wg.Wait()
}
