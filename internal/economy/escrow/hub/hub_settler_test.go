package hub

import (
	"context"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/economy/escrow"
)

func TestHubSettler_InterfaceCompliance(t *testing.T) {
	t.Parallel()
	var _ escrow.SettlementExecutor = (*HubSettler)(nil)
}

func TestHubSettler_SetAndGetDealMapping(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1)

	s.SetDealMapping("esc-1", big.NewInt(42))

	id, ok := s.GetDealID("esc-1")
	require.True(t, ok)
	assert.Equal(t, big.NewInt(42), id)
}

func TestHubSettler_GetDealID_NotFound(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1)

	_, ok := s.GetDealID("nonexistent")
	assert.False(t, ok)
}

func TestHubSettler_SetDealMapping_Overwrite(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1)

	s.SetDealMapping("esc-1", big.NewInt(10))
	s.SetDealMapping("esc-1", big.NewInt(20))

	id, ok := s.GetDealID("esc-1")
	require.True(t, ok)
	assert.Equal(t, big.NewInt(20), id)
}

func TestHubSettler_Lock_NoOp(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1)

	err := s.Lock(context.Background(), "did:test:buyer", big.NewInt(1000))
	require.NoError(t, err)
	assert.Empty(t, mc.writeCalls)
}

func TestHubSettler_Release_NoOp(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1)

	err := s.Release(context.Background(), "did:test:seller", big.NewInt(1000))
	require.NoError(t, err)
	assert.Empty(t, mc.writeCalls)
}

func TestHubSettler_Refund_NoOp(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1)

	err := s.Refund(context.Background(), "did:test:buyer", big.NewInt(1000))
	require.NoError(t, err)
	assert.Empty(t, mc.writeCalls)
}

func TestHubSettler_HubClient_Accessor(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1)

	hub := s.HubClient()
	assert.NotNil(t, hub)
}

func TestHubSettler_TokenAddress(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	tokenAddr := common.HexToAddress("0xTOKEN")
	s := NewHubSettler(mc, common.HexToAddress("0x1"), tokenAddr, 1)

	assert.Equal(t, tokenAddr, s.TokenAddress())
}

func TestHubSettler_ConcurrentMapping(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := big.NewInt(int64(n))
			key := "esc-concurrent"
			s.SetDealMapping(key, id)
			s.GetDealID(key)
		}(i)
	}
	wg.Wait()
}
