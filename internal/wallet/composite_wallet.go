package wallet

import (
	"context"
	"math/big"
	"sync"
)

// CompositeWallet implements WalletProvider with primary/fallback logic.
// When the companion app is connected (checked via ConnectionChecker),
// the primary (RPC) wallet is used. Otherwise, falls back to local.
type CompositeWallet struct {
	primary  WalletProvider
	fallback WalletProvider
	checker  ConnectionChecker

	mu        sync.RWMutex
	usedLocal bool
}

// NewCompositeWallet creates a composite wallet with primary/fallback providers.
func NewCompositeWallet(primary, fallback WalletProvider, checker ConnectionChecker) *CompositeWallet {
	return &CompositeWallet{
		primary:  primary,
		fallback: fallback,
		checker:  checker,
	}
}

// Address returns the wallet address from the active provider.
func (w *CompositeWallet) Address(ctx context.Context) (string, error) {
	if w.checker != nil && w.checker.IsConnected() {
		addr, err := w.primary.Address(ctx)
		if err == nil {
			return addr, nil
		}
	}

	w.mu.Lock()
	w.usedLocal = true
	w.mu.Unlock()

	return w.fallback.Address(ctx)
}

// Balance returns the balance from the active provider.
func (w *CompositeWallet) Balance(ctx context.Context) (*big.Int, error) {
	// Balance always uses local RPC node — companion may not support it.
	return w.fallback.Balance(ctx)
}

// SignTransaction signs using the active provider.
func (w *CompositeWallet) SignTransaction(ctx context.Context, rawTx []byte) ([]byte, error) {
	if w.checker != nil && w.checker.IsConnected() {
		sig, err := w.primary.SignTransaction(ctx, rawTx)
		if err == nil {
			return sig, nil
		}
	}

	w.mu.Lock()
	w.usedLocal = true
	w.mu.Unlock()

	return w.fallback.SignTransaction(ctx, rawTx)
}

// SignMessage signs using the active provider.
func (w *CompositeWallet) SignMessage(ctx context.Context, message []byte) ([]byte, error) {
	if w.checker != nil && w.checker.IsConnected() {
		sig, err := w.primary.SignMessage(ctx, message)
		if err == nil {
			return sig, nil
		}
	}

	w.mu.Lock()
	w.usedLocal = true
	w.mu.Unlock()

	return w.fallback.SignMessage(ctx, message)
}

// PublicKey returns the compressed public key from the active provider.
func (w *CompositeWallet) PublicKey(ctx context.Context) ([]byte, error) {
	if w.checker != nil && w.checker.IsConnected() {
		pk, err := w.primary.PublicKey(ctx)
		if err == nil {
			return pk, nil
		}
	}

	w.mu.Lock()
	w.usedLocal = true
	w.mu.Unlock()

	return w.fallback.PublicKey(ctx)
}

// UsedLocal returns true if the fallback (local) wallet was used at any point.
func (w *CompositeWallet) UsedLocal() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.usedLocal
}
