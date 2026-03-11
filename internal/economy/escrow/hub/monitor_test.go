package hub

import (
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/eventbus"
)

// testMonitor creates an EventMonitor with a real eventbus but no RPC.
// Only useful for testing helper functions and handleEvent.
func testMonitor(t *testing.T, store OnChainStore) *EventMonitor {
	t.Helper()
	bus := eventbus.New()
	m, err := NewEventMonitor(nil, bus, store, common.HexToAddress("0xHUB"))
	require.NoError(t, err)
	return m
}

// ---- helper function tests ----

func TestTopicToBigInt(t *testing.T) {
	t.Parallel()
	m := testMonitor(t, nil)

	log := types.Log{
		Topics: []common.Hash{
			common.BigToHash(big.NewInt(0)),
			common.BigToHash(big.NewInt(42)),
		},
	}

	result := m.topicToBigInt(log, 1)
	assert.Equal(t, "42", result)
}

func TestTopicToBigInt_OutOfRange(t *testing.T) {
	t.Parallel()
	m := testMonitor(t, nil)

	log := types.Log{Topics: []common.Hash{{}}}
	result := m.topicToBigInt(log, 5)
	assert.Equal(t, "", result)
}

func TestTopicToAddress(t *testing.T) {
	t.Parallel()
	m := testMonitor(t, nil)

	addr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	log := types.Log{
		Topics: []common.Hash{
			{},
			common.BytesToHash(addr.Bytes()),
		},
	}

	result := m.topicToAddress(log, 1)
	assert.Equal(t, addr.Hex(), result)
}

func TestTopicToAddress_OutOfRange(t *testing.T) {
	t.Parallel()
	m := testMonitor(t, nil)

	log := types.Log{Topics: []common.Hash{{}}}
	result := m.topicToAddress(log, 3)
	assert.Equal(t, "", result)
}

func TestDecodeAmount(t *testing.T) {
	t.Parallel()
	m := testMonitor(t, nil)

	amount := big.NewInt(1000000)
	data := common.LeftPadBytes(amount.Bytes(), 32)
	log := types.Log{Data: data}

	result := m.decodeAmount(log)
	assert.Equal(t, amount, result)
}

func TestDecodeAmount_ShortData(t *testing.T) {
	t.Parallel()
	m := testMonitor(t, nil)

	log := types.Log{Data: []byte{1, 2, 3}}
	result := m.decodeAmount(log)
	assert.Equal(t, new(big.Int), result)
}

// ---- resolveEscrowID tests ----

func TestResolveEscrowID_WithStore(t *testing.T) {
	t.Parallel()
	store := newMockOnChainStore()
	store.Set("42", "esc-abc")

	m := testMonitor(t, store)
	result := m.resolveEscrowID("42")
	assert.Equal(t, "esc-abc", result)
}

func TestResolveEscrowID_NilStore(t *testing.T) {
	t.Parallel()
	m := testMonitor(t, nil)

	result := m.resolveEscrowID("42")
	assert.Equal(t, "", result)
}

func TestResolveEscrowID_NotFound(t *testing.T) {
	t.Parallel()
	store := newMockOnChainStore()
	m := testMonitor(t, store)

	result := m.resolveEscrowID("999")
	assert.Equal(t, "", result)
}

// ---- handleEvent tests ----

func TestHandleEvent_Deposited(t *testing.T) {
	t.Parallel()
	store := newMockOnChainStore()
	store.Set("1", "esc-dep")

	bus := eventbus.New()
	m, err := NewEventMonitor(nil, bus, store, common.HexToAddress("0xHUB"))
	require.NoError(t, err)

	var received eventbus.EscrowOnChainDepositEvent
	var mu sync.Mutex
	eventbus.SubscribeTyped(bus, func(e eventbus.EscrowOnChainDepositEvent) {
		mu.Lock()
		received = e
		mu.Unlock()
	})

	amount := big.NewInt(5000)
	log := types.Log{
		Topics: []common.Hash{
			{}, // event ID (not checked in handleEvent)
			common.BigToHash(big.NewInt(1)),
			common.BytesToHash(common.HexToAddress("0xBuyer").Bytes()),
		},
		Data:   common.LeftPadBytes(amount.Bytes(), 32),
		TxHash: common.HexToHash("0xdeptx"),
	}

	m.handleEvent("Deposited", log)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "esc-dep", received.EscrowID)
	assert.Equal(t, "1", received.DealID)
	assert.Equal(t, amount, received.Amount)
}

func TestHandleEvent_WorkSubmitted(t *testing.T) {
	t.Parallel()
	bus := eventbus.New()
	m, err := NewEventMonitor(nil, bus, nil, common.HexToAddress("0xHUB"))
	require.NoError(t, err)

	var received eventbus.EscrowOnChainWorkEvent
	eventbus.SubscribeTyped(bus, func(e eventbus.EscrowOnChainWorkEvent) {
		received = e
	})

	log := types.Log{
		Topics: []common.Hash{
			{},
			common.BigToHash(big.NewInt(3)),
			common.BytesToHash(common.HexToAddress("0xSeller").Bytes()),
		},
		TxHash: common.HexToHash("0xworktx"),
	}

	m.handleEvent("WorkSubmitted", log)
	assert.Equal(t, "3", received.DealID)
}

func TestHandleEvent_Released(t *testing.T) {
	t.Parallel()
	bus := eventbus.New()
	m, err := NewEventMonitor(nil, bus, nil, common.HexToAddress("0xHUB"))
	require.NoError(t, err)

	var received eventbus.EscrowOnChainReleaseEvent
	eventbus.SubscribeTyped(bus, func(e eventbus.EscrowOnChainReleaseEvent) {
		received = e
	})

	amount := big.NewInt(2000)
	log := types.Log{
		Topics: []common.Hash{
			{},
			common.BigToHash(big.NewInt(5)),
			common.BytesToHash(common.HexToAddress("0xSeller").Bytes()),
		},
		Data:   common.LeftPadBytes(amount.Bytes(), 32),
		TxHash: common.HexToHash("0xreltx"),
	}

	m.handleEvent("Released", log)
	assert.Equal(t, "5", received.DealID)
	assert.Equal(t, amount, received.Amount)
}

func TestHandleEvent_Refunded(t *testing.T) {
	t.Parallel()
	bus := eventbus.New()
	m, err := NewEventMonitor(nil, bus, nil, common.HexToAddress("0xHUB"))
	require.NoError(t, err)

	var received eventbus.EscrowOnChainRefundEvent
	eventbus.SubscribeTyped(bus, func(e eventbus.EscrowOnChainRefundEvent) {
		received = e
	})

	amount := big.NewInt(3000)
	log := types.Log{
		Topics: []common.Hash{
			{},
			common.BigToHash(big.NewInt(7)),
			common.BytesToHash(common.HexToAddress("0xBuyer").Bytes()),
		},
		Data:   common.LeftPadBytes(amount.Bytes(), 32),
		TxHash: common.HexToHash("0xreftx"),
	}

	m.handleEvent("Refunded", log)
	assert.Equal(t, "7", received.DealID)
	assert.Equal(t, amount, received.Amount)
}

func TestHandleEvent_Disputed(t *testing.T) {
	t.Parallel()
	bus := eventbus.New()
	m, err := NewEventMonitor(nil, bus, nil, common.HexToAddress("0xHUB"))
	require.NoError(t, err)

	var received eventbus.EscrowOnChainDisputeEvent
	eventbus.SubscribeTyped(bus, func(e eventbus.EscrowOnChainDisputeEvent) {
		received = e
	})

	log := types.Log{
		Topics: []common.Hash{
			{},
			common.BigToHash(big.NewInt(9)),
			common.BytesToHash(common.HexToAddress("0xInit").Bytes()),
		},
		TxHash: common.HexToHash("0xdisptx"),
	}

	m.handleEvent("Disputed", log)
	assert.Equal(t, "9", received.DealID)
}

func TestHandleEvent_DealResolved(t *testing.T) {
	t.Parallel()
	bus := eventbus.New()
	m, err := NewEventMonitor(nil, bus, nil, common.HexToAddress("0xHUB"))
	require.NoError(t, err)

	var received eventbus.EscrowOnChainResolvedEvent
	eventbus.SubscribeTyped(bus, func(e eventbus.EscrowOnChainResolvedEvent) {
		received = e
	})

	log := types.Log{
		Topics: []common.Hash{
			{},
			common.BigToHash(big.NewInt(11)),
		},
		TxHash: common.HexToHash("0xrestx"),
	}

	m.handleEvent("DealResolved", log)
	assert.Equal(t, "11", received.DealID)
}

// ---- processLog tests ----

func TestProcessLog_EmptyTopics(t *testing.T) {
	t.Parallel()
	m := testMonitor(t, nil)

	// Should not panic on empty topics.
	m.processLog(types.Log{Topics: []common.Hash{}})
}

func TestProcessLog_UnknownEventID(t *testing.T) {
	t.Parallel()
	m := testMonitor(t, nil)

	// Unknown event ID should be silently ignored.
	log := types.Log{
		Topics: []common.Hash{common.HexToHash("0xdeadbeef")},
	}
	m.processLog(log)
}

func TestMonitor_Name(t *testing.T) {
	t.Parallel()
	m := testMonitor(t, nil)
	assert.Equal(t, "escrow-event-monitor", m.Name())
}

// ---- extractDealAndAddress tests ----

func TestExtractDealAndAddress_V1(t *testing.T) {
	t.Parallel()
	m := testMonitor(t, nil)

	// V1 layout: [sig, dealId, addr] — 3 topics.
	dealID := big.NewInt(42)
	addr := common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	log := types.Log{
		Topics: []common.Hash{
			{},
			common.BigToHash(dealID),
			common.BytesToHash(addr.Bytes()),
		},
	}

	gotDealID, gotAddr := m.extractDealAndAddress(log, false)
	assert.Equal(t, "42", gotDealID)
	assert.Equal(t, addr.Hex(), gotAddr)
}

func TestExtractDealAndAddress_V2(t *testing.T) {
	t.Parallel()
	m := testMonitor(t, nil)

	// V2 layout: [sig, refId, dealId, addr] — 4 topics.
	refID := common.BigToHash(big.NewInt(99))
	dealID := big.NewInt(55)
	addr := common.HexToAddress("0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC")
	log := types.Log{
		Topics: []common.Hash{
			{},
			refID,
			common.BigToHash(dealID),
			common.BytesToHash(addr.Bytes()),
		},
	}

	gotDealID, gotAddr := m.extractDealAndAddress(log, true)
	assert.Equal(t, "55", gotDealID)
	assert.Equal(t, addr.Hex(), gotAddr)
}

// ---- extractDealID tests ----

func TestExtractDealID_V1(t *testing.T) {
	t.Parallel()
	m := testMonitor(t, nil)

	// V1 layout: [sig, dealId, ...] — dealID at index 1.
	dealID := big.NewInt(100)
	log := types.Log{
		Topics: []common.Hash{
			{},
			common.BigToHash(dealID),
		},
	}

	got := m.extractDealID(log, false)
	assert.Equal(t, "100", got)
}

func TestExtractDealID_V2(t *testing.T) {
	t.Parallel()
	m := testMonitor(t, nil)

	// V2 layout: [sig, refId, dealId, ...] — dealID at index 2.
	refID := common.BigToHash(big.NewInt(77))
	dealID := big.NewInt(200)
	log := types.Log{
		Topics: []common.Hash{
			{},
			refID,
			common.BigToHash(dealID),
		},
	}

	got := m.extractDealID(log, true)
	assert.Equal(t, "200", got)
}

// ---- isV2Event tests ----

func TestIsV2Event(t *testing.T) {
	t.Parallel()
	m := testMonitor(t, nil)

	tests := []struct {
		give       string
		eventName  string
		topicCount int
		wantV2     bool
	}{
		{give: "Deposited 3 topics (V1)", eventName: "Deposited", topicCount: 3, wantV2: false},
		{give: "Deposited 4 topics (V2)", eventName: "Deposited", topicCount: 4, wantV2: true},
		{give: "Released 3 topics (V1)", eventName: "Released", topicCount: 3, wantV2: false},
		{give: "Released 4 topics (V2)", eventName: "Released", topicCount: 4, wantV2: true},
		{give: "Disputed always V1", eventName: "Disputed", topicCount: 3, wantV2: false},
		{give: "DisputeRaised always V2", eventName: "DisputeRaised", topicCount: 3, wantV2: true},
		{give: "DealResolved V1", eventName: "DealResolved", topicCount: 2, wantV2: false},
		{give: "SettlementFinalized V2", eventName: "SettlementFinalized", topicCount: 3, wantV2: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			topics := make([]common.Hash, tt.topicCount)
			log := types.Log{Topics: topics}
			got := m.isV2Event(tt.eventName, log)
			assert.Equal(t, tt.wantV2, got)
		})
	}
}
