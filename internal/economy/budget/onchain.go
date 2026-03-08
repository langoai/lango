package budget

import (
	"math/big"
	"sync"
)

// OnChainSyncCallback syncs on-chain spending data to off-chain tracking.
type OnChainSyncCallback func(sessionID string, spent *big.Int)

// OnChainTracker tracks spending from on-chain SpendingHook data.
type OnChainTracker struct {
	mu       sync.RWMutex
	sessions map[string]*big.Int // sessionID -> cumulative spent
	callback OnChainSyncCallback
}

// NewOnChainTracker creates a new on-chain spending tracker.
func NewOnChainTracker() *OnChainTracker {
	return &OnChainTracker{
		sessions: make(map[string]*big.Int),
	}
}

// SetCallback sets the sync callback.
func (t *OnChainTracker) SetCallback(fn OnChainSyncCallback) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.callback = fn
}

// Record records a spend for a session.
func (t *OnChainTracker) Record(sessionID string, amount *big.Int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	current, ok := t.sessions[sessionID]
	if !ok {
		current = new(big.Int)
		t.sessions[sessionID] = current
	}
	current.Add(current, amount)

	if t.callback != nil {
		t.callback(sessionID, new(big.Int).Set(current))
	}
}

// GetSpent returns the cumulative amount spent for a session.
func (t *OnChainTracker) GetSpent(sessionID string) *big.Int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if spent, ok := t.sessions[sessionID]; ok {
		return new(big.Int).Set(spent)
	}
	return new(big.Int)
}

// Reset resets the tracker for a session (e.g., after on-chain sync).
func (t *OnChainTracker) Reset(sessionID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.sessions, sessionID)
}
