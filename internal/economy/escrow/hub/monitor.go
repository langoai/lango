package hub

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
)

// BlockchainClient abstracts the RPC calls needed by EventMonitor.
// *ethclient.Client satisfies this interface.
type BlockchainClient interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
}

// OnChainStore provides escrow ID resolution from on-chain deal IDs.
type OnChainStore interface {
	GetByOnChainDealID(dealID string) (escrowID string, err error)
}

// EventMonitor watches on-chain escrow contract events and publishes them
// to the event bus. Uses eth_getLogs polling with confirmation depth buffer
// to protect against L2 reorgs.
type EventMonitor struct {
	rpc          BlockchainClient
	bus          *eventbus.Bus
	store        OnChainStore
	hubAddr      common.Address
	hubABI       *ethabi.ABI
	pollInterval time.Duration
	logger       *zap.SugaredLogger

	confirmationDepth uint64
	blockHashes       map[uint64]common.Hash // block hash cache for reorg detection
	maxHashCache      int                    // max entries in blockHashes

	lastBlock uint64
	stopCh    chan struct{}
	wg        sync.WaitGroup
	mu        sync.Mutex
	running   bool
}

// MonitorOption configures an EventMonitor.
type MonitorOption func(*EventMonitor)

// WithPollInterval sets the polling interval.
func WithPollInterval(d time.Duration) MonitorOption {
	return func(m *EventMonitor) {
		if d > 0 {
			m.pollInterval = d
		}
	}
}

// WithMonitorLogger sets a structured logger.
func WithMonitorLogger(l *zap.SugaredLogger) MonitorOption {
	return func(m *EventMonitor) {
		if l != nil {
			m.logger = l
		}
	}
}

// WithConfirmationDepth sets the number of blocks to wait before processing events.
// This protects against L2 reorgs by only processing blocks that are depth-confirmed.
func WithConfirmationDepth(depth uint64) MonitorOption {
	return func(m *EventMonitor) {
		m.confirmationDepth = depth
	}
}

// NewEventMonitor creates a new contract event monitor.
func NewEventMonitor(
	rpc BlockchainClient,
	bus *eventbus.Bus,
	store OnChainStore,
	hubAddr common.Address,
	opts ...MonitorOption,
) (*EventMonitor, error) {
	abi, err := ParseHubABI()
	if err != nil {
		return nil, fmt.Errorf("monitor parse ABI: %w", err)
	}

	m := &EventMonitor{
		rpc:          rpc,
		bus:          bus,
		store:        store,
		hubAddr:      hubAddr,
		hubABI:       abi,
		pollInterval: 15 * time.Second,
		logger:       zap.NewNop().Sugar(),
		stopCh:       make(chan struct{}),
		blockHashes:  make(map[uint64]common.Hash),
		maxHashCache: 256,
	}
	for _, o := range opts {
		o(m)
	}
	return m, nil
}

// Name implements lifecycle.Component.
func (m *EventMonitor) Name() string { return "escrow-event-monitor" }

// Start begins polling for contract events.
func (m *EventMonitor) Start(ctx context.Context, wg *sync.WaitGroup) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return nil
	}

	// Get current block as starting point.
	header, err := m.rpc.HeaderByNumber(ctx, nil)
	if err != nil {
		return fmt.Errorf("get latest block: %w", err)
	}
	m.lastBlock = header.Number.Uint64()
	m.running = true

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if wg != nil {
			wg.Done()
		}
		m.poll()
	}()

	m.logger.Infow("event monitor started", "startBlock", m.lastBlock, "interval", m.pollInterval)
	return nil
}

// Stop halts the polling loop.
func (m *EventMonitor) Stop(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}
	close(m.stopCh)
	m.wg.Wait()
	m.running = false
	m.logger.Info("event monitor stopped")
	return nil
}

// Running returns whether the monitor is active.
func (m *EventMonitor) Running() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// poll is the main polling loop.
func (m *EventMonitor) poll() {
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			if err := m.fetchAndPublish(); err != nil {
				m.logger.Warnw("poll error", "error", err)
			}
		}
	}
}

// fetchAndPublish queries logs from lastBlock+1 to safeBlock and publishes events.
// safeBlock = latest - confirmationDepth to protect against reorgs.
func (m *EventMonitor) fetchAndPublish() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	header, err := m.rpc.HeaderByNumber(ctx, nil)
	if err != nil {
		return fmt.Errorf("get latest block: %w", err)
	}

	latest := header.Number.Uint64()

	// Apply confirmation depth buffer.
	safeBlock := latest
	if m.confirmationDepth > 0 && latest > m.confirmationDepth {
		safeBlock = latest - m.confirmationDepth
	}

	// Reorg detection: if safeBlock fell behind our lastBlock, the chain reorganized.
	if safeBlock < m.lastBlock {
		reorgDepth := m.lastBlock - safeBlock
		exceeds := m.confirmationDepth > 0 && reorgDepth > m.confirmationDepth
		m.logger.Warnw("reorg detected",
			"previousBlock", m.lastBlock, "newSafeBlock", safeBlock,
			"reorgDepth", reorgDepth, "exceedsConfirmationDepth", exceeds)
		m.bus.Publish(eventbus.EscrowReorgDetectedEvent{
			PreviousBlock: m.lastBlock,
			NewBlock:      safeBlock,
			Depth:         reorgDepth,
			ExceedsDepth:  exceeds,
		})
		// Roll back to safeBlock so we re-process from the correct point.
		m.lastBlock = safeBlock
		return nil
	}

	if safeBlock <= m.lastBlock {
		return nil
	}

	fromBlock := m.lastBlock + 1

	// Block hash continuity check for silent reorg detection.
	if err := m.checkBlockHashContinuity(ctx, fromBlock); err != nil {
		m.logger.Warnw("block hash continuity check failed", "block", fromBlock, "error", err)
	}

	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(safeBlock),
		Addresses: []common.Address{m.hubAddr},
	}

	logs, err := m.rpc.FilterLogs(ctx, query)
	if err != nil {
		return fmt.Errorf("filter logs [%d, %d]: %w", fromBlock, safeBlock, err)
	}

	for _, log := range logs {
		m.processLog(log)
	}

	// Cache block hashes for future continuity checks.
	m.cacheBlockHash(ctx, safeBlock)

	m.lastBlock = safeBlock
	return nil
}

// checkBlockHashContinuity verifies that the parent block hash hasn't changed,
// which would indicate a silent reorg within the confirmation window.
func (m *EventMonitor) checkBlockHashContinuity(ctx context.Context, fromBlock uint64) error {
	if fromBlock == 0 {
		return nil
	}
	prevBlock := fromBlock - 1
	cachedHash, ok := m.blockHashes[prevBlock]
	if !ok {
		return nil
	}

	header, err := m.rpc.HeaderByNumber(ctx, new(big.Int).SetUint64(prevBlock))
	if err != nil {
		return fmt.Errorf("get block %d header: %w", prevBlock, err)
	}

	currentHash := header.Hash()
	if cachedHash != currentHash {
		m.logger.Warnw("silent reorg detected via hash mismatch",
			"block", prevBlock, "cachedHash", cachedHash.Hex(), "currentHash", currentHash.Hex())
		m.bus.Publish(eventbus.EscrowReorgDetectedEvent{
			PreviousBlock: m.lastBlock,
			NewBlock:      prevBlock,
			Depth:         m.lastBlock - prevBlock,
			ExceedsDepth:  false,
		})
		m.lastBlock = prevBlock
	}
	return nil
}

// cacheBlockHash stores a block's hash for future reorg detection.
func (m *EventMonitor) cacheBlockHash(ctx context.Context, blockNum uint64) {
	header, err := m.rpc.HeaderByNumber(ctx, new(big.Int).SetUint64(blockNum))
	if err != nil {
		return
	}
	m.blockHashes[blockNum] = header.Hash()
	m.trimBlockHashCache()
}

// trimBlockHashCache removes old entries to keep cache bounded.
func (m *EventMonitor) trimBlockHashCache() {
	if len(m.blockHashes) <= m.maxHashCache {
		return
	}
	// Find minimum block number and remove entries below threshold.
	var minBlock uint64
	first := true
	for b := range m.blockHashes {
		if first || b < minBlock {
			minBlock = b
			first = false
		}
	}
	// Remove oldest half of entries.
	threshold := minBlock + uint64(m.maxHashCache/2)
	for b := range m.blockHashes {
		if b < threshold {
			delete(m.blockHashes, b)
		}
	}
}

// processLog decodes a single log entry and publishes the corresponding event.
func (m *EventMonitor) processLog(log types.Log) {
	if len(log.Topics) == 0 {
		return
	}

	eventID := log.Topics[0]

	// Match against known event signatures.
	for _, ev := range m.hubABI.Events {
		if ev.ID == eventID {
			m.handleEvent(ev.Name, log)
			return
		}
	}
}

// handleEvent publishes typed events to the event bus.
// Supports both V1 (topic layout: [sig, dealId, addr]) and V2 (topic layout: [sig, refId, dealId, addr]) events.
func (m *EventMonitor) handleEvent(eventName string, log types.Log) {
	txHash := log.TxHash.Hex()

	// V2 events have refId as first indexed parameter after the event signature.
	// Detect V2 by checking if the event has an extra indexed topic (4 topics for V2 vs 3 for V1).
	isV2 := m.isV2Event(eventName, log)

	switch eventName {
	case "Deposited":
		dealID, buyer := m.extractDealAndAddress(log, isV2)
		amount := m.decodeAmount(log)
		escrowID := m.resolveEscrowID(dealID)
		m.bus.Publish(eventbus.EscrowOnChainDepositEvent{
			EscrowID: escrowID,
			DealID:   dealID,
			Buyer:    buyer,
			Amount:   amount,
			TxHash:   txHash,
		})

	case "WorkSubmitted":
		dealID, seller := m.extractDealAndAddress(log, isV2)
		escrowID := m.resolveEscrowID(dealID)
		m.bus.Publish(eventbus.EscrowOnChainWorkEvent{
			EscrowID: escrowID,
			DealID:   dealID,
			Seller:   seller,
			TxHash:   txHash,
		})

	case "Released":
		dealID, seller := m.extractDealAndAddress(log, isV2)
		amount := m.decodeAmount(log)
		escrowID := m.resolveEscrowID(dealID)
		m.bus.Publish(eventbus.EscrowOnChainReleaseEvent{
			EscrowID: escrowID,
			DealID:   dealID,
			Seller:   seller,
			Amount:   amount,
			TxHash:   txHash,
		})

	case "Refunded":
		dealID, buyer := m.extractDealAndAddress(log, isV2)
		amount := m.decodeAmount(log)
		escrowID := m.resolveEscrowID(dealID)
		m.bus.Publish(eventbus.EscrowOnChainRefundEvent{
			EscrowID: escrowID,
			DealID:   dealID,
			Buyer:    buyer,
			Amount:   amount,
			TxHash:   txHash,
		})

	case "Disputed", "DisputeRaised":
		// Dispute events have different V2 layout: initiator is in non-indexed data.
		var dealID, initiator string
		if isV2 {
			dealID = m.topicToBigInt(log, 2)
			initiator = m.decodeAddress(log)
		} else {
			dealID = m.topicToBigInt(log, 1)
			initiator = m.topicToAddress(log, 2)
		}
		escrowID := m.resolveEscrowID(dealID)
		m.bus.Publish(eventbus.EscrowOnChainDisputeEvent{
			EscrowID:  escrowID,
			DealID:    dealID,
			Initiator: initiator,
			TxHash:    txHash,
		})

	case "DealResolved", "SettlementFinalized":
		dealID := m.extractDealID(log, isV2)
		escrowID := m.resolveEscrowID(dealID)
		m.bus.Publish(eventbus.EscrowOnChainResolvedEvent{
			EscrowID: escrowID,
			DealID:   dealID,
			TxHash:   txHash,
		})

	case "EscrowOpened":
		dealID := m.topicToBigInt(log, 2)
		m.logger.Debugw("escrow opened on-chain", "dealID", dealID, "txHash", txHash)

	case "MilestoneReached":
		dealID := m.topicToBigInt(log, 2)
		m.logger.Debugw("milestone reached on-chain", "dealID", dealID, "txHash", txHash)

	case "DealCreated":
		m.logger.Debugw("deal created on-chain", "txHash", txHash)
	}
}

// extractDealAndAddress extracts dealID and address from log topics,
// accounting for V2's extra refId topic at index 1.
func (m *EventMonitor) extractDealAndAddress(log types.Log, isV2 bool) (dealID, addr string) {
	if isV2 {
		return m.topicToBigInt(log, 2), m.topicToAddress(log, 3)
	}
	return m.topicToBigInt(log, 1), m.topicToAddress(log, 2)
}

// extractDealID extracts only the dealID from log topics.
func (m *EventMonitor) extractDealID(log types.Log, isV2 bool) string {
	if isV2 {
		return m.topicToBigInt(log, 2)
	}
	return m.topicToBigInt(log, 1)
}

// isV2Event detects V2 events by topic count.
// V2 events always have refId as an indexed parameter, giving them one extra topic.
func (m *EventMonitor) isV2Event(eventName string, log types.Log) bool {
	switch eventName {
	case "Deposited", "Released", "Refunded", "WorkSubmitted":
		// V1: 3 topics [sig, dealId, addr], V2: 4 topics [sig, refId, dealId, addr]
		return len(log.Topics) >= 4
	case "Disputed":
		// V1: 3 topics [sig, dealId, initiator], V2 "DisputeRaised": 3 topics [sig, refId, dealId]
		return false
	case "DisputeRaised":
		return true
	case "DealResolved":
		return false
	case "SettlementFinalized", "EscrowOpened", "MilestoneReached":
		return true
	default:
		return false
	}
}

// topicToBigInt extracts a uint256 value from an indexed topic.
func (m *EventMonitor) topicToBigInt(log types.Log, idx int) string {
	if idx >= len(log.Topics) {
		return ""
	}
	return new(big.Int).SetBytes(log.Topics[idx].Bytes()).String()
}

// topicToAddress extracts an address from an indexed topic.
func (m *EventMonitor) topicToAddress(log types.Log, idx int) string {
	if idx >= len(log.Topics) {
		return ""
	}
	return common.BytesToAddress(log.Topics[idx].Bytes()).Hex()
}

// decodeAmount extracts amount from non-indexed log data.
func (m *EventMonitor) decodeAmount(log types.Log) *big.Int {
	if len(log.Data) >= 32 {
		return new(big.Int).SetBytes(log.Data[:32])
	}
	return new(big.Int)
}

// decodeAddress extracts an address from the first 32 bytes of non-indexed log data.
func (m *EventMonitor) decodeAddress(log types.Log) string {
	if len(log.Data) >= 32 {
		return common.BytesToAddress(log.Data[:32]).Hex()
	}
	return ""
}

// resolveEscrowID maps an on-chain deal ID string to a local escrow ID.
func (m *EventMonitor) resolveEscrowID(dealID string) string {
	if m.store == nil {
		return ""
	}
	escrowID, err := m.store.GetByOnChainDealID(dealID)
	if err != nil {
		m.logger.Debugw("resolve escrow ID", "dealID", dealID, "error", err)
		return ""
	}
	return escrowID
}
