package session

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/types"
)

// fakeCompactor implements MessageCompactor for tests.
type fakeCompactor struct {
	mu           sync.Mutex
	history      []Message
	compactCalls int
	compactDelay time.Duration
	compactErr   error
}

func (f *fakeCompactor) Get(_ string) (*Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make([]Message, len(f.history))
	copy(cp, f.history)
	return &Session{History: cp}, nil
}

func (f *fakeCompactor) CompactMessages(_ string, upToIndex int, summary string) error {
	if f.compactDelay > 0 {
		time.Sleep(f.compactDelay)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.compactCalls++
	if f.compactErr != nil {
		return f.compactErr
	}
	if upToIndex < 0 || upToIndex >= len(f.history) {
		return nil
	}
	remaining := append([]Message{{
		Role:    types.RoleAssistant,
		Content: "[Compacted Summary]\n" + summary,
	}}, f.history[upToIndex+1:]...)
	f.history = remaining
	return nil
}

func newFakeCompactor(n int) *fakeCompactor {
	f := &fakeCompactor{}
	for i := 0; i < n; i++ {
		f.history = append(f.history, Message{
			Role:    types.RoleUser,
			Content: "lorem ipsum dolor sit amet " + string(rune('a'+i)),
		})
	}
	return f
}

func TestCompactionBuffer_EnqueueTriggersCompaction(t *testing.T) {
	t.Parallel()

	comp := newFakeCompactor(20)
	bus := eventbus.New()
	buf := NewCompactionBuffer(comp, bus, zap.NewNop().Sugar())

	var wg sync.WaitGroup
	buf.Start(&wg)
	t.Cleanup(func() {
		buf.Stop()
		wg.Wait()
	})

	var received atomic.Int32
	eventbus.SubscribeTyped(bus, func(_ eventbus.CompactionCompletedEvent) {
		received.Add(1)
	})

	buf.EnqueueCompaction("sess-1", 9)

	assert.Eventually(t, func() bool { return received.Load() == 1 }, 2*time.Second, 10*time.Millisecond)
	assert.Equal(t, 1, comp.compactCalls)
}

func TestCompactionBuffer_WaitForSession_CompletesWithinTimeout(t *testing.T) {
	t.Parallel()

	comp := newFakeCompactor(20)
	comp.compactDelay = 50 * time.Millisecond
	buf := NewCompactionBuffer(comp, nil, zap.NewNop().Sugar())

	var wg sync.WaitGroup
	buf.Start(&wg)
	t.Cleanup(func() {
		buf.Stop()
		wg.Wait()
	})

	buf.EnqueueCompaction("sess-1", 9)
	// Give the worker a moment to pick up the job.
	time.Sleep(5 * time.Millisecond)

	done, waited := buf.WaitForSession(context.Background(), "sess-1", 500*time.Millisecond)
	assert.True(t, done)
	assert.Less(t, waited, 500*time.Millisecond)
}

func TestCompactionBuffer_WaitForSession_Timeout(t *testing.T) {
	t.Parallel()

	comp := newFakeCompactor(20)
	comp.compactDelay = 300 * time.Millisecond
	buf := NewCompactionBuffer(comp, nil, zap.NewNop().Sugar())

	var wg sync.WaitGroup
	buf.Start(&wg)
	t.Cleanup(func() {
		buf.Stop()
		wg.Wait()
	})

	buf.EnqueueCompaction("sess-1", 9)
	time.Sleep(5 * time.Millisecond)

	done, _ := buf.WaitForSession(context.Background(), "sess-1", 50*time.Millisecond)
	assert.False(t, done, "wait should time out when compaction is slower than timeout")
}

func TestCompactionBuffer_WaitForSession_NoInFlight(t *testing.T) {
	t.Parallel()

	buf := NewCompactionBuffer(newFakeCompactor(0), nil, zap.NewNop().Sugar())
	done, waited := buf.WaitForSession(context.Background(), "missing", time.Second)
	assert.True(t, done)
	assert.Equal(t, time.Duration(0), waited)
}

func TestCompactionBuffer_EventCarriesReclaimedTokens(t *testing.T) {
	t.Parallel()

	comp := newFakeCompactor(30)
	bus := eventbus.New()
	buf := NewCompactionBuffer(comp, bus, zap.NewNop().Sugar())

	var wg sync.WaitGroup
	buf.Start(&wg)
	t.Cleanup(func() {
		buf.Stop()
		wg.Wait()
	})

	received := make(chan eventbus.CompactionCompletedEvent, 1)
	eventbus.SubscribeTyped(bus, func(e eventbus.CompactionCompletedEvent) {
		received <- e
	})

	buf.EnqueueCompaction("sess-1", 14)

	select {
	case e := <-received:
		require.Equal(t, "sess-1", e.SessionKey)
		assert.Equal(t, 14, e.UpToIndex)
		assert.Greater(t, e.ReclaimedTokens, 0)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for CompactionCompletedEvent")
	}
}
