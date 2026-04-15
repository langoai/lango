package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResolveCompactionDefaults(t *testing.T) {
	t.Parallel()

	got := ContextCompactionConfig{}.ResolveCompaction()

	assert.NotNil(t, got.Enabled)
	assert.True(t, *got.Enabled)
	assert.InDelta(t, 0.5, got.Threshold, 0.001)
	assert.Equal(t, 2*time.Second, got.SyncTimeout)
	assert.Equal(t, 1, got.WorkerCount)
}

func TestResolveCompactionClamps(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		give       ContextCompactionConfig
		wantThresh float64
		wantSync   time.Duration
	}{
		{
			name:       "threshold above max clamps to 0.95",
			give:       ContextCompactionConfig{Threshold: 1.5, SyncTimeout: time.Second},
			wantThresh: 0.95,
			wantSync:   time.Second,
		},
		{
			name:       "threshold below min clamps to 0.1",
			give:       ContextCompactionConfig{Threshold: 0.01, SyncTimeout: time.Second},
			wantThresh: 0.1,
			wantSync:   time.Second,
		},
		{
			name:       "syncTimeout above max clamps to 10s",
			give:       ContextCompactionConfig{Threshold: 0.5, SyncTimeout: time.Minute},
			wantThresh: 0.5,
			wantSync:   10 * time.Second,
		},
		{
			name:       "syncTimeout below min clamps to 100ms",
			give:       ContextCompactionConfig{Threshold: 0.5, SyncTimeout: 10 * time.Millisecond},
			wantThresh: 0.5,
			wantSync:   100 * time.Millisecond,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.give.ResolveCompaction()
			assert.InDelta(t, tt.wantThresh, got.Threshold, 0.001)
			assert.Equal(t, tt.wantSync, got.SyncTimeout)
		})
	}
}

func TestResolveRecallDefaults(t *testing.T) {
	t.Parallel()

	got := ContextRecallConfig{}.ResolveRecall()

	assert.NotNil(t, got.Enabled)
	assert.True(t, *got.Enabled)
	assert.Equal(t, 3, got.TopN)
	assert.InDelta(t, 0.2, got.MinRank, 0.001)
}

func TestResolveRecallClamps(t *testing.T) {
	t.Parallel()

	got := ContextRecallConfig{TopN: 50, MinRank: 5.0}.ResolveRecall()
	assert.Equal(t, 10, got.TopN)
	assert.InDelta(t, 1.0, got.MinRank, 0.001)
}

func TestResolveSuggestionsDefaults(t *testing.T) {
	t.Parallel()

	got := LearningSuggestionsConfig{}.ResolveSuggestions()

	assert.NotNil(t, got.Enabled)
	assert.True(t, *got.Enabled)
	assert.InDelta(t, 0.5, got.Threshold, 0.001)
	assert.Equal(t, 10, got.RateLimit)
	assert.Equal(t, time.Hour, got.DedupWindow)
}

func TestResolveSuggestionsClamps(t *testing.T) {
	t.Parallel()

	got := LearningSuggestionsConfig{
		Threshold:   2.0,
		RateLimit:   500,
		DedupWindow: 48 * time.Hour,
	}.ResolveSuggestions()

	assert.InDelta(t, 0.9, got.Threshold, 0.001)
	assert.Equal(t, 100, got.RateLimit)
	assert.Equal(t, 24*time.Hour, got.DedupWindow)
}
