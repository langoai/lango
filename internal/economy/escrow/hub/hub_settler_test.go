package hub

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

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

func TestHubSettler_SetDealMappingByDID(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1)

	s.SetDealMappingByDID("did:test:buyer", big.NewInt(99))

	id, ok := s.GetDealID("did:test:buyer")
	require.True(t, ok)
	assert.Equal(t, big.NewInt(99), id)
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

func TestHubSettler_Lock_NilHub(t *testing.T) {
	t.Parallel()
	s := NewHubSettlerOffline(common.HexToAddress("0x2"), 1,
		WithHubLogger(zap.NewNop().Sugar()))

	err := s.Lock(context.Background(), "did:test:buyer", big.NewInt(1000))
	require.NoError(t, err)
}

func TestHubSettler_Lock_CreatesAndDeposits(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.writeResult.Data = []interface{}{big.NewInt(7)} // dealID
	mc.writeResult.TxHash = "0xmocktx"

	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1,
		WithHubLogger(zap.NewNop().Sugar()))

	err := s.Lock(context.Background(), "did:test:buyer", big.NewInt(1000))
	require.NoError(t, err)

	// Two write calls: createDeal + deposit.
	mc.mu.Lock()
	assert.Len(t, mc.writeCalls, 2)
	assert.Equal(t, "createDeal", mc.writeCalls[0].Method)
	assert.Equal(t, "deposit", mc.writeCalls[1].Method)
	mc.mu.Unlock()

	// Deal mapping should be set.
	id, ok := s.GetDealID("did:test:buyer")
	require.True(t, ok)
	assert.Equal(t, big.NewInt(7), id)
}

func TestHubSettler_Lock_CreateDealError(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.writeErr = fmt.Errorf("rpc error")

	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1)

	err := s.Lock(context.Background(), "did:test:buyer", big.NewInt(500))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create deal")
}

func TestHubSettler_Release_NilHub(t *testing.T) {
	t.Parallel()
	s := NewHubSettlerOffline(common.HexToAddress("0x2"), 1,
		WithHubLogger(zap.NewNop().Sugar()))

	err := s.Release(context.Background(), "did:test:seller", big.NewInt(1000))
	require.NoError(t, err)
}

func TestHubSettler_Release_WithMapping(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.writeResult.TxHash = "0xreleasetx"

	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1,
		WithHubLogger(zap.NewNop().Sugar()))

	s.SetDealMappingByDID("did:test:seller", big.NewInt(42))

	err := s.Release(context.Background(), "did:test:seller", big.NewInt(1000))
	require.NoError(t, err)

	mc.mu.Lock()
	require.Len(t, mc.writeCalls, 1)
	assert.Equal(t, "release", mc.writeCalls[0].Method)
	mc.mu.Unlock()
}

func TestHubSettler_Release_NoMapping(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()

	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1)

	err := s.Release(context.Background(), "did:test:unknown", big.NewInt(1000))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no deal mapping")
}

func TestHubSettler_Release_HubError(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.writeErr = fmt.Errorf("hub unavailable")

	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1)
	s.SetDealMappingByDID("did:test:seller", big.NewInt(5))

	err := s.Release(context.Background(), "did:test:seller", big.NewInt(100))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "release deal")
}

func TestHubSettler_Refund_NilHub(t *testing.T) {
	t.Parallel()
	s := NewHubSettlerOffline(common.HexToAddress("0x2"), 1,
		WithHubLogger(zap.NewNop().Sugar()))

	err := s.Refund(context.Background(), "did:test:buyer", big.NewInt(1000))
	require.NoError(t, err)
}

func TestHubSettler_Refund_WithMapping(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.writeResult.TxHash = "0xrefundtx"

	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1,
		WithHubLogger(zap.NewNop().Sugar()))

	s.SetDealMappingByDID("did:test:buyer", big.NewInt(77))

	err := s.Refund(context.Background(), "did:test:buyer", big.NewInt(500))
	require.NoError(t, err)

	mc.mu.Lock()
	require.Len(t, mc.writeCalls, 1)
	assert.Equal(t, "refund", mc.writeCalls[0].Method)
	mc.mu.Unlock()
}

func TestHubSettler_Refund_NoMapping(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()

	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1)

	err := s.Refund(context.Background(), "did:test:unknown", big.NewInt(500))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no deal mapping")
}

func TestHubSettler_Refund_HubError(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	mc.writeErr = fmt.Errorf("network failure")

	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1)
	s.SetDealMappingByDID("did:test:buyer", big.NewInt(3))

	err := s.Refund(context.Background(), "did:test:buyer", big.NewInt(100))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "refund deal")
}

func TestHubSettler_HubClient_Accessor(t *testing.T) {
	t.Parallel()
	mc := newMockCaller()
	s := NewHubSettler(mc, common.HexToAddress("0x1"), common.HexToAddress("0x2"), 1)

	hub := s.HubClient()
	assert.NotNil(t, hub)
}

func TestHubSettler_HubClient_NilOffline(t *testing.T) {
	t.Parallel()
	s := NewHubSettlerOffline(common.HexToAddress("0x2"), 1)
	assert.Nil(t, s.HubClient())
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
