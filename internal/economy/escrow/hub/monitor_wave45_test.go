package hub

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/eventbus"
)

type wave45BlockchainClient struct {
	mu sync.Mutex

	headers   map[uint64]*types.Header
	logs      []types.Log
	headerErr error
	logErr    error
	queries   []ethereum.FilterQuery
}

func newWave45BlockchainClient() *wave45BlockchainClient {
	return &wave45BlockchainClient{
		headers: make(map[uint64]*types.Header),
	}
}

func (c *wave45BlockchainClient) setHeader(num uint64, extra string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.headers[num] = wave45Header(num, extra)
}

func (c *wave45BlockchainClient) setLatest(num uint64) {
	c.setHeader(num, "latest")
}

func (c *wave45BlockchainClient) HeaderByNumber(
	_ context.Context,
	number *big.Int,
) (*types.Header, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.headerErr != nil {
		return nil, c.headerErr
	}
	if number != nil {
		header, ok := c.headers[number.Uint64()]
		if !ok {
			return wave45Header(number.Uint64(), ""), nil
		}
		return header, nil
	}

	var latest uint64
	var latestHeader *types.Header
	for block, header := range c.headers {
		if latestHeader == nil || block > latest {
			latest = block
			latestHeader = header
		}
	}
	if latestHeader == nil {
		return wave45Header(0, ""), nil
	}
	return latestHeader, nil
}

func (c *wave45BlockchainClient) FilterLogs(
	_ context.Context,
	query ethereum.FilterQuery,
) ([]types.Log, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.queries = append(c.queries, query)
	if c.logErr != nil {
		return nil, c.logErr
	}
	logs := make([]types.Log, len(c.logs))
	copy(logs, c.logs)
	return logs, nil
}

func wave45Header(num uint64, extra string) *types.Header {
	return &types.Header{
		Number: new(big.Int).SetUint64(num),
		Extra:  []byte(extra),
	}
}

type wave45ErrorStore struct {
	err error
}

func (s wave45ErrorStore) GetByOnChainDealID(string) (string, error) {
	return "", s.err
}

func TestWave45FetchAndPublishFiltersRangePublishesABIEventAndAdvances(t *testing.T) {
	t.Parallel()

	client := newWave45BlockchainClient()
	client.setLatest(15)
	hubAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	buyer := common.HexToAddress("0x2222222222222222222222222222222222222222")
	amount := big.NewInt(12345)
	store := newMockOnChainStore()
	store.Set("42", "escrow-42")
	bus := eventbus.New()

	var received eventbus.EscrowOnChainDepositEvent
	eventbus.SubscribeTyped(bus, func(e eventbus.EscrowOnChainDepositEvent) {
		received = e
	})

	monitor, err := NewEventMonitor(client, bus, store, hubAddr)
	require.NoError(t, err)
	monitor.lastBlock = 10
	deposited := monitor.hubABI.Events["Deposited"]
	client.logs = []types.Log{
		{
			Topics: []common.Hash{
				deposited.ID,
				common.BigToHash(big.NewInt(42)),
				common.BytesToHash(buyer.Bytes()),
			},
			Data:   common.LeftPadBytes(amount.Bytes(), 32),
			TxHash: common.HexToHash("0xabc"),
		},
	}

	require.NoError(t, monitor.fetchAndPublish())

	require.Len(t, client.queries, 1)
	assert.Equal(t, uint64(11), client.queries[0].FromBlock.Uint64())
	assert.Equal(t, uint64(15), client.queries[0].ToBlock.Uint64())
	assert.Equal(t, []common.Address{hubAddr}, client.queries[0].Addresses)
	assert.Equal(t, uint64(15), monitor.lastBlock)
	assert.Equal(t, client.headers[15].Hash(), monitor.blockHashes[15])
	assert.Equal(t, "escrow-42", received.EscrowID)
	assert.Equal(t, "42", received.DealID)
	assert.Equal(t, buyer.Hex(), received.Buyer)
	assert.Equal(t, amount, received.Amount)
	assert.Equal(t, common.HexToHash("0xabc").Hex(), received.TxHash)
}

func TestWave45FetchAndPublishReturnsLatestHeaderErrorWithoutAdvancing(t *testing.T) {
	t.Parallel()

	client := newWave45BlockchainClient()
	client.headerErr = errors.New("header unavailable")
	monitor, err := NewEventMonitor(client, eventbus.New(), nil, common.Address{})
	require.NoError(t, err)
	monitor.lastBlock = 7

	err = monitor.fetchAndPublish()

	require.Error(t, err)
	assert.ErrorContains(t, err, "get latest block")
	assert.ErrorContains(t, err, "header unavailable")
	assert.Equal(t, uint64(7), monitor.lastBlock)
	assert.Empty(t, client.queries)
}

func TestWave45FetchAndPublishReturnsFilterErrorWithoutAdvancing(t *testing.T) {
	t.Parallel()

	client := newWave45BlockchainClient()
	client.setLatest(12)
	client.logErr = errors.New("filter unavailable")
	monitor, err := NewEventMonitor(client, eventbus.New(), nil, common.Address{})
	require.NoError(t, err)
	monitor.lastBlock = 10

	err = monitor.fetchAndPublish()

	require.Error(t, err)
	assert.ErrorContains(t, err, "filter logs [11, 12]")
	assert.ErrorContains(t, err, "filter unavailable")
	assert.Equal(t, uint64(10), monitor.lastBlock)
	require.Len(t, client.queries, 1)
	assert.Equal(t, uint64(11), client.queries[0].FromBlock.Uint64())
	assert.Equal(t, uint64(12), client.queries[0].ToBlock.Uint64())
}

func TestWave45FetchAndPublishNoopsWhenSafeBlockDoesNotAdvance(t *testing.T) {
	t.Parallel()

	client := newWave45BlockchainClient()
	client.setLatest(12)
	monitor, err := NewEventMonitor(client, eventbus.New(), nil, common.Address{},
		WithConfirmationDepth(2),
	)
	require.NoError(t, err)
	monitor.lastBlock = 10

	require.NoError(t, monitor.fetchAndPublish())

	assert.Equal(t, uint64(10), monitor.lastBlock)
	assert.Empty(t, client.queries)
}

func TestWave45HandleEventDisputeRaisedDecodesV2InitiatorFromData(t *testing.T) {
	t.Parallel()

	store := newMockOnChainStore()
	store.Set("77", "escrow-77")
	bus := eventbus.New()
	monitor, err := NewEventMonitor(nil, bus, store, common.Address{})
	require.NoError(t, err)
	initiator := common.HexToAddress("0x3333333333333333333333333333333333333333")

	var received eventbus.EscrowOnChainDisputeEvent
	eventbus.SubscribeTyped(bus, func(e eventbus.EscrowOnChainDisputeEvent) {
		received = e
	})

	monitor.handleEvent("DisputeRaised", types.Log{
		Topics: []common.Hash{
			{},
			common.BigToHash(big.NewInt(5)),
			common.BigToHash(big.NewInt(77)),
		},
		Data:   common.LeftPadBytes(initiator.Bytes(), 32),
		TxHash: common.HexToHash("0xdef"),
	})

	assert.Equal(t, "escrow-77", received.EscrowID)
	assert.Equal(t, "77", received.DealID)
	assert.Equal(t, initiator.Hex(), received.Initiator)
	assert.Equal(t, common.HexToHash("0xdef").Hex(), received.TxHash)
}

func TestWave45HandleEventSettlementFinalizedUsesV2DealIDTopic(t *testing.T) {
	t.Parallel()

	store := newMockOnChainStore()
	store.Set("88", "escrow-88")
	bus := eventbus.New()
	monitor, err := NewEventMonitor(nil, bus, store, common.Address{})
	require.NoError(t, err)

	var received eventbus.EscrowOnChainResolvedEvent
	eventbus.SubscribeTyped(bus, func(e eventbus.EscrowOnChainResolvedEvent) {
		received = e
	})

	monitor.handleEvent("SettlementFinalized", types.Log{
		Topics: []common.Hash{
			{},
			common.BigToHash(big.NewInt(6)),
			common.BigToHash(big.NewInt(88)),
		},
		TxHash: common.HexToHash("0x123"),
	})

	assert.Equal(t, "escrow-88", received.EscrowID)
	assert.Equal(t, "88", received.DealID)
	assert.Equal(t, common.HexToHash("0x123").Hex(), received.TxHash)
}

func TestWave45ResolveEscrowIDReturnsEmptyWhenStoreErrors(t *testing.T) {
	t.Parallel()

	monitor := testMonitor(t, wave45ErrorStore{err: errors.New("store unavailable")})

	assert.Equal(t, "", monitor.resolveEscrowID("99"))
}

func TestWave45CheckBlockHashContinuityPublishesSilentReorgAndRollsBack(t *testing.T) {
	t.Parallel()

	client := newWave45BlockchainClient()
	client.setHeader(19, "new-parent")
	bus := eventbus.New()
	monitor, err := NewEventMonitor(client, bus, nil, common.Address{})
	require.NoError(t, err)
	monitor.lastBlock = 20
	monitor.blockHashes[19] = wave45Header(19, "old-parent").Hash()

	var received eventbus.EscrowReorgDetectedEvent
	eventbus.SubscribeTyped(bus, func(e eventbus.EscrowReorgDetectedEvent) {
		received = e
	})

	require.NoError(t, monitor.checkBlockHashContinuity(context.Background(), 20))

	assert.Equal(t, uint64(19), monitor.lastBlock)
	assert.Equal(t, uint64(20), received.PreviousBlock)
	assert.Equal(t, uint64(19), received.NewBlock)
	assert.Equal(t, uint64(1), received.Depth)
	assert.False(t, received.ExceedsDepth)
}

func TestWave45StartStopLifecycleIsIdempotent(t *testing.T) {
	t.Parallel()

	client := newWave45BlockchainClient()
	client.setLatest(8)
	monitor, err := NewEventMonitor(client, eventbus.New(), nil, common.Address{},
		WithPollInterval(time.Hour),
	)
	require.NoError(t, err)

	var started sync.WaitGroup
	started.Add(1)
	require.NoError(t, monitor.Start(context.Background(), &started))
	wave45WaitGroup(t, &started)
	require.NoError(t, monitor.Start(context.Background(), nil))

	assert.True(t, monitor.Running())
	assert.Equal(t, uint64(8), monitor.lastBlock)

	require.NoError(t, monitor.Stop(context.Background()))
	require.NoError(t, monitor.Stop(context.Background()))
	assert.False(t, monitor.Running())
}

func wave45WaitGroup(t *testing.T, wg *sync.WaitGroup) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()

	timeout := time.Second
	if deadline, ok := t.Deadline(); ok {
		remaining := time.Until(deadline) / 10
		if remaining > 0 && remaining < timeout {
			timeout = remaining
		}
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-done:
	case <-timer.C:
		t.Fatal("timed out waiting for monitor goroutine start")
	}
}
