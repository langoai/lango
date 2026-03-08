package hub

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/contract"
	"github.com/langoai/lango/internal/economy/escrow"
)

// Compile-time check.
var _ escrow.SettlementExecutor = (*VaultSettler)(nil)

// VaultSettler implements SettlementExecutor using per-deal LangoVault contracts
// created via the LangoVaultFactory.
type VaultSettler struct {
	factory    *FactoryClient
	caller     contract.ContractCaller
	tokenAddr  common.Address
	implAddr   common.Address
	arbitrator common.Address
	chainID    int64
	logger     *zap.SugaredLogger

	// vaultMap tracks escrowID → vault address.
	vaultMap map[string]common.Address
	mu       sync.RWMutex
}

// VaultSettlerOption configures a VaultSettler.
type VaultSettlerOption func(*VaultSettler)

// WithVaultLogger sets a structured logger.
func WithVaultLogger(l *zap.SugaredLogger) VaultSettlerOption {
	return func(s *VaultSettler) {
		if l != nil {
			s.logger = l
		}
	}
}

// NewVaultSettler creates a vault-mode settler.
func NewVaultSettler(
	caller contract.ContractCaller,
	factoryAddr, implAddr, tokenAddr, arbitrator common.Address,
	chainID int64,
	opts ...VaultSettlerOption,
) *VaultSettler {
	s := &VaultSettler{
		factory:    NewFactoryClient(caller, factoryAddr, chainID),
		caller:     caller,
		tokenAddr:  tokenAddr,
		implAddr:   implAddr,
		arbitrator: arbitrator,
		chainID:    chainID,
		logger:     zap.NewNop().Sugar(),
		vaultMap:   make(map[string]common.Address),
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// SetVaultMapping associates a local escrow ID with a vault address.
func (s *VaultSettler) SetVaultMapping(escrowID string, vaultAddr common.Address) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vaultMap[escrowID] = vaultAddr
}

// GetVaultAddress returns the vault address for a local escrow ID.
func (s *VaultSettler) GetVaultAddress(escrowID string) (common.Address, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	addr, ok := s.vaultMap[escrowID]
	return addr, ok
}

// Lock is a no-op for vault mode; actual vault creation + deposit
// is done by the escrow tools layer.
func (s *VaultSettler) Lock(_ context.Context, buyerDID string, amount *big.Int) error {
	s.logger.Infow("vault settler lock",
		"buyerDID", buyerDID, "amount", amount.String())
	return nil
}

// Release is a no-op at this level; actual vault release is done
// by the tools layer which has vault address context.
func (s *VaultSettler) Release(ctx context.Context, sellerDID string, amount *big.Int) error {
	s.logger.Infow("vault settler release",
		"sellerDID", sellerDID, "amount", amount.String())
	return nil
}

// Refund is a no-op at this level; actual vault refund is done
// by the tools layer which has vault address context.
func (s *VaultSettler) Refund(ctx context.Context, buyerDID string, amount *big.Int) error {
	s.logger.Infow("vault settler refund",
		"buyerDID", buyerDID, "amount", amount.String())
	return nil
}

// CreateVault creates a new vault via the factory and returns its address.
func (s *VaultSettler) CreateVault(ctx context.Context, seller common.Address, amount, deadline *big.Int) (common.Address, string, error) {
	info, txHash, err := s.factory.CreateVault(ctx, seller, s.tokenAddr, amount, deadline, s.arbitrator)
	if err != nil {
		return common.Address{}, "", fmt.Errorf("create vault: %w", err)
	}
	return info.VaultAddress, txHash, nil
}

// VaultClientFor creates a VaultClient for a specific vault address.
func (s *VaultSettler) VaultClientFor(vaultAddr common.Address) *VaultClient {
	return NewVaultClient(s.caller, vaultAddr, s.chainID)
}

// FactoryClient exposes the underlying factory client.
func (s *VaultSettler) FactoryClient() *FactoryClient {
	return s.factory
}

// TokenAddress returns the configured ERC-20 token address.
func (s *VaultSettler) TokenAddress() common.Address {
	return s.tokenAddr
}
