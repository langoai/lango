package paygate

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// DeferredEntry tracks a post-pay obligation that is settled asynchronously
// after tool execution completes.
type DeferredEntry struct {
	ID        string    `json:"id"`
	PeerDID   string    `json:"peerDid"`
	ToolName  string    `json:"toolName"`
	Price     string    `json:"price"`
	CreatedAt time.Time `json:"createdAt"`
	Settled   bool      `json:"settled"`
	TxHash    string    `json:"txHash,omitempty"`
}

// DeferredLedger is an in-memory ledger tracking post-pay obligations.
// It is safe for concurrent use.
type DeferredLedger struct {
	mu      sync.Mutex
	entries map[string]*DeferredEntry
}

// NewDeferredLedger creates an empty deferred ledger.
func NewDeferredLedger() *DeferredLedger {
	return &DeferredLedger{
		entries: make(map[string]*DeferredEntry),
	}
}

// Add records a new deferred payment obligation and returns the entry ID.
func (l *DeferredLedger) Add(peerDID, toolName, price string) string {
	l.mu.Lock()
	defer l.mu.Unlock()

	id := uuid.New().String()
	l.entries[id] = &DeferredEntry{
		ID:        id,
		PeerDID:   peerDID,
		ToolName:  toolName,
		Price:     price,
		CreatedAt: time.Now(),
	}
	return id
}

// Settle marks an entry as settled with the given transaction hash.
// Returns false if the entry does not exist.
func (l *DeferredLedger) Settle(id, txHash string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.entries[id]
	if !ok {
		return false
	}
	entry.Settled = true
	entry.TxHash = txHash
	return true
}

// Pending returns all unsettled entries.
func (l *DeferredLedger) Pending() []*DeferredEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	var result []*DeferredEntry
	for _, e := range l.entries {
		if !e.Settled {
			result = append(result, e)
		}
	}
	return result
}

// PendingByPeer returns unsettled entries for a specific peer.
func (l *DeferredLedger) PendingByPeer(peerDID string) []*DeferredEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	var result []*DeferredEntry
	for _, e := range l.entries {
		if !e.Settled && e.PeerDID == peerDID {
			result = append(result, e)
		}
	}
	return result
}
