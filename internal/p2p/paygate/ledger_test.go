package paygate

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeferredLedger_Add(t *testing.T) {
	l := NewDeferredLedger()
	id := l.Add("did:peer:a", "tool-x", "0.50")

	assert.NotEmpty(t, id)
	pending := l.Pending()
	require.Len(t, pending, 1)
	assert.Equal(t, "did:peer:a", pending[0].PeerDID)
	assert.Equal(t, "tool-x", pending[0].ToolName)
	assert.Equal(t, "0.50", pending[0].Price)
	assert.False(t, pending[0].Settled)
}

func TestDeferredLedger_Settle(t *testing.T) {
	l := NewDeferredLedger()
	id := l.Add("did:peer:b", "tool-y", "1.00")

	ok := l.Settle(id, "0xabc123")
	assert.True(t, ok)

	pending := l.Pending()
	assert.Empty(t, pending)
}

func TestDeferredLedger_Settle_NotFound(t *testing.T) {
	l := NewDeferredLedger()
	ok := l.Settle("nonexistent-id", "0xabc")
	assert.False(t, ok)
}

func TestDeferredLedger_PendingByPeer(t *testing.T) {
	l := NewDeferredLedger()
	l.Add("did:peer:alice", "tool-1", "0.10")
	l.Add("did:peer:bob", "tool-2", "0.20")
	l.Add("did:peer:alice", "tool-3", "0.30")

	alice := l.PendingByPeer("did:peer:alice")
	assert.Len(t, alice, 2)

	bob := l.PendingByPeer("did:peer:bob")
	assert.Len(t, bob, 1)

	unknown := l.PendingByPeer("did:peer:unknown")
	assert.Empty(t, unknown)
}

func TestDeferredLedger_ConcurrentAccess(t *testing.T) {
	l := NewDeferredLedger()
	var wg sync.WaitGroup
	ids := make([]string, 100)

	// Concurrent adds.
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ids[idx] = l.Add("did:peer:concurrent", "tool", "0.01")
		}(i)
	}
	wg.Wait()

	pending := l.Pending()
	assert.Len(t, pending, 100)

	// Concurrent settles.
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			l.Settle(ids[idx], "0xhash")
		}(i)
	}
	wg.Wait()

	pending = l.Pending()
	assert.Empty(t, pending)
}

func TestDeferredLedger_MultipleAdds(t *testing.T) {
	l := NewDeferredLedger()
	id1 := l.Add("did:peer:a", "tool-1", "0.50")
	id2 := l.Add("did:peer:a", "tool-2", "1.00")

	assert.NotEqual(t, id1, id2)
	assert.Len(t, l.Pending(), 2)

	l.Settle(id1, "0xhash1")
	assert.Len(t, l.Pending(), 1)
	assert.Equal(t, id2, l.Pending()[0].ID)
}
