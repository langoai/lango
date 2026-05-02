package cockpit

import (
	"sync"
	"time"

	"github.com/langoai/lango/internal/eventbus"
)

const (
	learningSuggestionCapacity = 20
	learningSuggestionTTL      = 30 * time.Minute
)

// LearningSuggestionBuffer stores recent learning suggestions for cockpit
// projection. It is session-scoped and concurrency-safe.
type LearningSuggestionBuffer struct {
	mu        sync.Mutex
	items     []eventbus.LearningSuggestionEvent
	dismissed map[string]struct{}
	nowFn     func() time.Time
}

func NewLearningSuggestionBuffer(nowFn func() time.Time) *LearningSuggestionBuffer {
	if nowFn == nil {
		nowFn = time.Now
	}
	return &LearningSuggestionBuffer{
		dismissed: make(map[string]struct{}),
		nowFn:     nowFn,
	}
}

func (b *LearningSuggestionBuffer) Append(event eventbus.LearningSuggestionEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.pruneLocked()
	if len(b.items) >= learningSuggestionCapacity {
		b.items = append([]eventbus.LearningSuggestionEvent(nil), b.items[len(b.items)-learningSuggestionCapacity+1:]...)
	}
	b.items = append(b.items, event)
}

func (b *LearningSuggestionBuffer) Snapshot() []eventbus.LearningSuggestionEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.pruneLocked()
	if len(b.items) == 0 {
		return nil
	}
	out := make([]eventbus.LearningSuggestionEvent, len(b.items))
	copy(out, b.items)
	return out
}

func (b *LearningSuggestionBuffer) Find(id string) *eventbus.LearningSuggestionEvent {
	if id == "" {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	b.pruneLocked()
	for _, item := range b.items {
		if item.SuggestionID != id {
			continue
		}
		itemCopy := item
		return &itemCopy
	}
	return nil
}

func (b *LearningSuggestionBuffer) Dismiss(id string) {
	if id == "" {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.dismissed[id] = struct{}{}
	b.pruneLocked()
}

func (b *LearningSuggestionBuffer) pruneLocked() {
	if len(b.items) == 0 {
		return
	}
	cutoff := b.nowFn().Add(-learningSuggestionTTL)
	filtered := b.items[:0]
	for _, item := range b.items {
		if item.Timestamp.Before(cutoff) {
			continue
		}
		if _, dismissed := b.dismissed[item.SuggestionID]; dismissed {
			continue
		}
		filtered = append(filtered, item)
	}
	b.items = filtered
}
