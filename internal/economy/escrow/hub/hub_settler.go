package hub

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/contract"
	"github.com/langoai/lango/internal/economy/escrow"
)

// Compile-time check.
var _ escrow.SettlementExecutor = (*HubSettler)(nil)

// HubSettler implements SettlementExecutor using the LangoEscrowHub contract.
// Lock creates a deal + deposits on the hub. Release/Refund delegate to the hub.
type HubSettler struct {
	hub       *HubClient
	tokenAddr common.Address
	chainID   int64
	logger    *zap.SugaredLogger

	// dealMap tracks key → on-chain dealID.
	// Keys are either escrow IDs (via SetDealMapping) or DIDs (via Lock).
	dealMap map[string]*big.Int
	mu      sync.RWMutex
}

// HubSettlerOption configures a HubSettler.
type HubSettlerOption func(*HubSettler)

// WithHubLogger sets a structured logger for the settler.
func WithHubLogger(l *zap.SugaredLogger) HubSettlerOption {
	return func(s *HubSettler) {
		if l != nil {
			s.logger = l
		}
	}
}

// NewHubSettler creates a hub-mode settler.
func NewHubSettler(caller contract.ContractCaller, hubAddr, tokenAddr common.Address, chainID int64, opts ...HubSettlerOption) *HubSettler {
	s := &HubSettler{
		hub:       NewHubClient(caller, hubAddr, chainID),
		tokenAddr: tokenAddr,
		chainID:   chainID,
		logger:    zap.NewNop().Sugar(),
		dealMap:   make(map[string]*big.Int),
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// NewHubSettlerOffline creates a hub settler without a hub client (offline/test mode).
// All on-chain operations become no-ops with warning logs.
func NewHubSettlerOffline(tokenAddr common.Address, chainID int64, opts ...HubSettlerOption) *HubSettler {
	s := &HubSettler{
		tokenAddr: tokenAddr,
		chainID:   chainID,
		logger:    zap.NewNop().Sugar(),
		dealMap:   make(map[string]*big.Int),
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// SetDealMapping associates a local escrow ID with an on-chain deal ID.
func (s *HubSettler) SetDealMapping(escrowID string, dealID *big.Int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dealMap[escrowID] = new(big.Int).Set(dealID)
}

// SetDealMappingByDID associates a DID with an on-chain deal ID.
func (s *HubSettler) SetDealMappingByDID(did string, dealID *big.Int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dealMap[did] = new(big.Int).Set(dealID)
}

// GetDealID returns the on-chain deal ID for a local escrow ID or DID.
func (s *HubSettler) GetDealID(key string) (*big.Int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.dealMap[key]
	return id, ok
}

// Lock creates an on-chain deal and deposits funds.
// If the hub client is nil (offline mode), this is a no-op.
func (s *HubSettler) Lock(ctx context.Context, buyerDID string, amount *big.Int) error {
	if s.hub == nil {
		s.logger.Warnw("hub client nil, skipping on-chain lock", "buyer", buyerDID, "amount", amount)
		return nil
	}

	deadline := new(big.Int).SetInt64(time.Now().Add(24 * time.Hour).Unix())
	dealID, txHash, err := s.hub.CreateDeal(ctx, common.Address{}, s.tokenAddr, amount, deadline)
	if err != nil {
		return fmt.Errorf("create deal: %w", err)
	}

	depositTx, err := s.hub.Deposit(ctx, dealID)
	if err != nil {
		return fmt.Errorf("deposit deal %s: %w", dealID, err)
	}

	s.mu.Lock()
	s.dealMap[buyerDID] = dealID
	s.mu.Unlock()

	s.logger.Infow("funds locked on-chain",
		"dealID", dealID, "createTx", txHash, "depositTx", depositTx,
		"buyerDID", buyerDID, "amount", amount)
	return nil
}

// Release releases funds on the hub contract for the given seller.
// If the hub client is nil (offline mode), this is a no-op.
func (s *HubSettler) Release(ctx context.Context, sellerDID string, amount *big.Int) error {
	if s.hub == nil {
		s.logger.Warnw("hub client nil, skipping on-chain release", "seller", sellerDID)
		return nil
	}

	s.mu.RLock()
	dealID, ok := s.dealMap[sellerDID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("release: no deal mapping for seller %s", sellerDID)
	}

	txHash, err := s.hub.Release(ctx, dealID)
	if err != nil {
		return fmt.Errorf("release deal %s: %w", dealID, err)
	}

	s.mu.Lock()
	delete(s.dealMap, sellerDID)
	s.mu.Unlock()

	s.logger.Infow("funds released on-chain",
		"dealID", dealID, "txHash", txHash, "seller", sellerDID)
	return nil
}

// Refund refunds funds on the hub contract to the given buyer.
// If the hub client is nil (offline mode), this is a no-op.
func (s *HubSettler) Refund(ctx context.Context, buyerDID string, amount *big.Int) error {
	if s.hub == nil {
		s.logger.Warnw("hub client nil, skipping on-chain refund", "buyer", buyerDID)
		return nil
	}

	s.mu.RLock()
	dealID, ok := s.dealMap[buyerDID]
	s.mu.RUnlock()
	if !ok {
		return fmt.Errorf("refund: no deal mapping for buyer %s", buyerDID)
	}

	txHash, err := s.hub.Refund(ctx, dealID)
	if err != nil {
		return fmt.Errorf("refund deal %s: %w", dealID, err)
	}

	s.mu.Lock()
	delete(s.dealMap, buyerDID)
	s.mu.Unlock()

	s.logger.Infow("funds refunded on-chain",
		"dealID", dealID, "txHash", txHash, "buyer", buyerDID)
	return nil
}

// HubClient exposes the underlying hub client for direct operations.
func (s *HubSettler) HubClient() *HubClient {
	return s.hub
}

// TokenAddress returns the configured ERC-20 token address.
func (s *HubSettler) TokenAddress() common.Address {
	return s.tokenAddr
}
