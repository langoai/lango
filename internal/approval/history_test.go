package approval

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoryStore_AppendAndList(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tests := []struct {
		give      string
		wantCount int
		wantFirst string // newest entry's RequestID
		wantLast  string // oldest entry's RequestID
	}{
		{
			give:      "three entries newest-first",
			wantCount: 3,
			wantFirst: "req-3",
			wantLast:  "req-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			store := NewHistoryStore(10)

			store.Append(HistoryEntry{Timestamp: now, RequestID: "req-1", Outcome: "granted"})
			store.Append(HistoryEntry{Timestamp: now.Add(time.Second), RequestID: "req-2", Outcome: "denied"})
			store.Append(HistoryEntry{Timestamp: now.Add(2 * time.Second), RequestID: "req-3", Outcome: "granted"})

			result := store.List()
			require.Len(t, result, tt.wantCount)
			assert.Equal(t, tt.wantFirst, result[0].RequestID)
			assert.Equal(t, tt.wantLast, result[len(result)-1].RequestID)
		})
	}
}

func TestHistoryStore_ListEmpty(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore(10)
	result := store.List()
	assert.Nil(t, result)
}

func TestHistoryStore_RingBufferOverflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give        string
		giveMax     int
		giveCount   int
		wantCount   int
		wantFirstID string // newest
		wantLastID  string // oldest surviving
	}{
		{
			give:        "5 entries into maxSize=3 keeps last 3",
			giveMax:     3,
			giveCount:   5,
			wantCount:   3,
			wantFirstID: "req-4",
			wantLastID:  "req-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			store := NewHistoryStore(tt.giveMax)

			for i := 0; i < tt.giveCount; i++ {
				store.Append(HistoryEntry{RequestID: fmt.Sprintf("req-%d", i)})
			}

			result := store.List()
			require.Len(t, result, tt.wantCount)
			assert.Equal(t, tt.wantFirstID, result[0].RequestID)
			assert.Equal(t, tt.wantLastID, result[len(result)-1].RequestID)
		})
	}
}

func TestHistoryStore_Count(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		giveMax   int
		giveCount int
		wantCount int
	}{
		{
			give:      "count within capacity",
			giveMax:   10,
			giveCount: 5,
			wantCount: 5,
		},
		{
			give:      "count capped at maxSize",
			giveMax:   3,
			giveCount: 7,
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			store := NewHistoryStore(tt.giveMax)

			for i := 0; i < tt.giveCount; i++ {
				store.Append(HistoryEntry{RequestID: fmt.Sprintf("req-%d", i)})
			}

			assert.Equal(t, tt.wantCount, store.Count())
		})
	}
}

func TestHistoryStore_CountByOutcome(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give        string
		giveEntries []HistoryEntry
		wantCounts  map[string]int
	}{
		{
			give: "mixed outcomes",
			giveEntries: []HistoryEntry{
				{Outcome: "granted"},
				{Outcome: "denied"},
				{Outcome: "granted"},
				{Outcome: "timeout"},
				{Outcome: "granted"},
				{Outcome: "denied"},
			},
			wantCounts: map[string]int{
				"granted": 3,
				"denied":  2,
				"timeout": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			store := NewHistoryStore(100)

			for _, e := range tt.giveEntries {
				store.Append(e)
			}

			assert.Equal(t, tt.wantCounts, store.CountByOutcome())
		})
	}
}

func TestHistoryStore_ConcurrentSafety(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore(100)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(id int) {
			defer wg.Done()
			store.Append(HistoryEntry{
				RequestID: fmt.Sprintf("req-%d", id),
				Outcome:   "granted",
			})
		}(i)
		go func() {
			defer wg.Done()
			_ = store.List()
			_ = store.Count()
			_ = store.CountByOutcome()
		}()
	}

	wg.Wait()

	assert.Equal(t, 100, store.Count())
}

func TestHistoryStore_DefaultMaxSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give        string
		giveMaxSize int
		wantMaxSize int
	}{
		{give: "zero defaults to 500", giveMaxSize: 0, wantMaxSize: 500},
		{give: "negative defaults to 500", giveMaxSize: -1, wantMaxSize: 500},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			store := NewHistoryStore(tt.giveMaxSize)
			assert.Equal(t, tt.wantMaxSize, store.maxSize)
		})
	}
}
