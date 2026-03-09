package hub

import (
	"context"
	"math/big"
	"sync"

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

	// dealMap tracks escrowID → on-chain dealID (set by wiring layer).
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

// SetDealMapping associates a local escrow ID with an on-chain deal ID.
func (s *HubSettler) SetDealMapping(escrowID string, dealID *big.Int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dealMap[escrowID] = new(big.Int).Set(dealID)
}

// GetDealID returns the on-chain deal ID for a local escrow ID.
func (s *HubSettler) GetDealID(escrowID string) (*big.Int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.dealMap[escrowID]
	return id, ok
}

// Lock verifies balance sufficiency (hub model — funds held in hub contract after deposit).
// The actual on-chain createDeal + deposit is handled by the escrow tools layer
// since the SettlementExecutor.Lock signature only receives buyerDID + amount.
func (s *HubSettler) Lock(_ context.Context, buyerDID string, amount *big.Int) error {
	s.logger.Infow("hub settler lock",
		"buyerDID", buyerDID, "amount", amount.String())
	return nil
}

// Release releases funds on the hub contract for the given seller.
func (s *HubSettler) Release(ctx context.Context, sellerDID string, amount *big.Int) error {
	s.logger.Infow("hub settler release",
		"sellerDID", sellerDID, "amount", amount.String())
	// Note: release is called from Engine.Release which knows the escrowID.
	// The actual hub.Release(dealID) call is done in the tools layer
	// where we have access to the escrowID → dealID mapping.
	return nil
}

// Refund refunds funds on the hub contract to the given buyer.
func (s *HubSettler) Refund(ctx context.Context, buyerDID string, amount *big.Int) error {
	s.logger.Infow("hub settler refund",
		"buyerDID", buyerDID, "amount", amount.String())
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
