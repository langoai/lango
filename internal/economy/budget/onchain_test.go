package budget

import (
	"math/big"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnChainTracker_Record(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		giveAmounts []int64
		wantTotal  int64
	}{
		{
			give:       "single record",
			giveAmounts: []int64{100},
			wantTotal:  100,
		},
		{
			give:       "multiple records accumulate",
			giveAmounts: []int64{100, 200, 300},
			wantTotal:  600,
		},
		{
			give:       "zero amount",
			giveAmounts: []int64{0},
			wantTotal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			tracker := NewOnChainTracker()
			for _, amount := range tt.giveAmounts {
				tracker.Record("session-1", big.NewInt(amount))
			}

			got := tracker.GetSpent("session-1")
			assert.Equal(t, 0, got.Cmp(big.NewInt(tt.wantTotal)))
		})
	}
}

func TestOnChainTracker_GetSpent_UnknownSession(t *testing.T) {
	t.Parallel()

	tracker := NewOnChainTracker()
	got := tracker.GetSpent("nonexistent")
	assert.Equal(t, 0, got.Sign(), "unknown session should return zero")
}

func TestOnChainTracker_GetSpent_DefensiveCopy(t *testing.T) {
	t.Parallel()

	tracker := NewOnChainTracker()
	tracker.Record("session-1", big.NewInt(500))

	got := tracker.GetSpent("session-1")
	got.SetInt64(0) // mutate returned value

	// Internal state should not be affected.
	assert.Equal(t, 0, tracker.GetSpent("session-1").Cmp(big.NewInt(500)))
}

func TestOnChainTracker_Callback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give           string
		giveAmounts    []int64
		wantCalls      int
		wantLastSpent  int64
	}{
		{
			give:          "callback called on each record",
			giveAmounts:   []int64{100, 200},
			wantCalls:     2,
			wantLastSpent: 300,
		},
		{
			give:          "single record callback",
			giveAmounts:   []int64{42},
			wantCalls:     1,
			wantLastSpent: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			tracker := NewOnChainTracker()

			var mu sync.Mutex
			var callCount int
			var lastSessionID string
			var lastSpent *big.Int

			tracker.SetCallback(func(sessionID string, spent *big.Int) {
				mu.Lock()
				defer mu.Unlock()
				callCount++
				lastSessionID = sessionID
				lastSpent = new(big.Int).Set(spent)
			})

			for _, amount := range tt.giveAmounts {
				tracker.Record("session-1", big.NewInt(amount))
			}

			mu.Lock()
			defer mu.Unlock()
			assert.Equal(t, tt.wantCalls, callCount)
			assert.Equal(t, "session-1", lastSessionID)
			require.NotNil(t, lastSpent)
			assert.Equal(t, 0, lastSpent.Cmp(big.NewInt(tt.wantLastSpent)))
		})
	}
}

func TestOnChainTracker_Callback_NotSet(t *testing.T) {
	t.Parallel()

	tracker := NewOnChainTracker()
	// Should not panic when callback is nil.
	tracker.Record("session-1", big.NewInt(100))
	assert.Equal(t, 0, tracker.GetSpent("session-1").Cmp(big.NewInt(100)))
}

func TestOnChainTracker_Reset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give          string
		giveSession   string
		giveAmount    int64
		resetSession  string
		wantAfterReset int64
	}{
		{
			give:           "reset clears tracked session",
			giveSession:    "session-1",
			giveAmount:     500,
			resetSession:   "session-1",
			wantAfterReset: 0,
		},
		{
			give:           "reset nonexistent session is safe",
			giveSession:    "session-1",
			giveAmount:     500,
			resetSession:   "session-other",
			wantAfterReset: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			tracker := NewOnChainTracker()
			tracker.Record(tt.giveSession, big.NewInt(tt.giveAmount))
			tracker.Reset(tt.resetSession)

			got := tracker.GetSpent(tt.giveSession)
			assert.Equal(t, 0, got.Cmp(big.NewInt(tt.wantAfterReset)))
		})
	}
}

func TestOnChainTracker_MultipleSessions(t *testing.T) {
	t.Parallel()

	tracker := NewOnChainTracker()
	tracker.Record("session-a", big.NewInt(100))
	tracker.Record("session-b", big.NewInt(200))
	tracker.Record("session-a", big.NewInt(50))

	assert.Equal(t, 0, tracker.GetSpent("session-a").Cmp(big.NewInt(150)))
	assert.Equal(t, 0, tracker.GetSpent("session-b").Cmp(big.NewInt(200)))
}
