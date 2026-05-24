package reputation_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/p2p/reputation"
	"github.com/langoai/lango/internal/testutil"
)

func TestCalculateScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		successes int
		failures  int
		timeouts  int
		want      float64
	}{
		{
			give:      "zero history",
			successes: 0,
			failures:  0,
			timeouts:  0,
			want:      0.0,
		},
		{
			give:      "one success",
			successes: 1,
			failures:  0,
			timeouts:  0,
			want:      0.5, // 1 / (1 + 0 + 0 + 1)
		},
		{
			give:      "one success one failure",
			successes: 1,
			failures:  1,
			timeouts:  0,
			want:      0.25, // 1 / (1 + 2 + 0 + 1)
		},
		{
			give:      "one success one timeout",
			successes: 1,
			failures:  0,
			timeouts:  1,
			want:      1.0 / 3.5, // 1 / (1 + 0 + 1.5 + 1)
		},
		{
			give:      "ten successes no failures",
			successes: 10,
			failures:  0,
			timeouts:  0,
			want:      10.0 / 11.0, // 10 / (10 + 0 + 0 + 1)
		},
		{
			give:      "ten successes two failures one timeout",
			successes: 10,
			failures:  2,
			timeouts:  1,
			want:      10.0 / (10.0 + 4.0 + 1.5 + 1.0), // 10 / 16.5
		},
		{
			give:      "only failures",
			successes: 0,
			failures:  5,
			timeouts:  0,
			want:      0.0, // 0 / (0 + 10 + 0 + 1)
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := reputation.CalculateScore(tt.successes, tt.failures, tt.timeouts)
			assert.InDelta(t, tt.want, got, 1e-9)
		})
	}
}

func TestCalculateScore_Progression(t *testing.T) {
	t.Parallel()

	// Score should monotonically increase as successes grow with no failures.
	var prev float64
	for i := 1; i <= 100; i++ {
		score := reputation.CalculateScore(i, 0, 0)
		assert.Greater(t, score, prev, "score should increase at successes=%d", i)
		prev = score
	}

	// Score should approach 1.0 with many successes.
	score := reputation.CalculateScore(10000, 0, 0)
	assert.Greater(t, score, 0.999, "score should approach 1.0 with many successes")
}

func TestCalculateScore_FailurePenalty(t *testing.T) {
	t.Parallel()

	// Failures should penalize more heavily than timeouts.
	scoreWithFailure := reputation.CalculateScore(5, 1, 0)
	scoreWithTimeout := reputation.CalculateScore(5, 0, 1)
	assert.Less(t, scoreWithFailure, scoreWithTimeout,
		"failures (weight 2) should penalize more than timeouts (weight 1.5)")
}

func TestStore_ConcurrentUpdatesOnSinglePeerPreserved(t *testing.T) {
	t.Parallel()

	client := testutil.TestEntClient(t)
	store := reputation.NewStore(client, testutil.NopLogger())
	ctx := context.Background()

	const (
		successCount = 25
		failureCount = 20
		timeoutCount = 15
	)

	var wg sync.WaitGroup
	errCh := make(chan error, successCount+failureCount+timeoutCount)
	for i := 0; i < successCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh <- store.RecordSuccess(ctx, "did:lango:peer")
		}()
	}
	for i := 0; i < failureCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh <- store.RecordFailure(ctx, "did:lango:peer")
		}()
	}
	for i := 0; i < timeoutCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh <- store.RecordTimeout(ctx, "did:lango:peer")
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	details, err := store.GetDetails(ctx, "did:lango:peer")
	require.NoError(t, err)
	require.NotNil(t, details)
	assert.Equal(t, successCount, details.SuccessfulExchanges)
	assert.Equal(t, failureCount*reputation.FailurePenaltyUnits, details.FailedExchanges)
	assert.Equal(t, timeoutCount, details.TimeoutCount)
}
