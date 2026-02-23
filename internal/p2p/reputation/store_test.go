package reputation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateScore(t *testing.T) {
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
			got := CalculateScore(tt.successes, tt.failures, tt.timeouts)
			assert.InDelta(t, tt.want, got, 1e-9)
		})
	}
}

func TestCalculateScore_Progression(t *testing.T) {
	// Score should monotonically increase as successes grow with no failures.
	var prev float64
	for i := 1; i <= 100; i++ {
		score := CalculateScore(i, 0, 0)
		assert.Greater(t, score, prev, "score should increase at successes=%d", i)
		prev = score
	}

	// Score should approach 1.0 with many successes.
	score := CalculateScore(10000, 0, 0)
	assert.Greater(t, score, 0.999, "score should approach 1.0 with many successes")
}

func TestCalculateScore_FailurePenalty(t *testing.T) {
	// Failures should penalize more heavily than timeouts.
	scoreWithFailure := CalculateScore(5, 1, 0)
	scoreWithTimeout := CalculateScore(5, 0, 1)
	assert.Less(t, scoreWithFailure, scoreWithTimeout,
		"failures (weight 2) should penalize more than timeouts (weight 1.5)")
}
