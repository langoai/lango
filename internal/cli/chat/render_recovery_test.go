package chat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRenderRecoveryBlock(t *testing.T) {
	tests := []struct {
		give           string
		giveAction     string
		giveCauseClass string
		giveAttempt    int
		giveBackoff    time.Duration
		wantContain    []string
		wantNotContain []string
	}{
		{
			give:           "retry action",
			giveAction:     "retry",
			giveCauseClass: "rate_limit",
			giveAttempt:    2,
			giveBackoff:    0,
			wantContain:    []string{"Retry", "#2", "rate_limit"},
			wantNotContain: []string{"backoff"},
		},
		{
			give:           "reroute action",
			giveAction:     "retry_with_hint",
			giveCauseClass: "tool_error",
			giveAttempt:    1,
			giveBackoff:    0,
			wantContain:    []string{"Reroute", "#1", "tool_error"},
		},
		{
			give:           "with backoff duration",
			giveAction:     "retry",
			giveCauseClass: "timeout",
			giveAttempt:    3,
			giveBackoff:    2 * time.Second,
			wantContain:    []string{"Retry", "#3", "timeout", "2s", "backoff"},
		},
		{
			give:           "zero backoff omits backoff text",
			giveAction:     "retry",
			giveCauseClass: "rate_limit",
			giveAttempt:    1,
			giveBackoff:    0,
			wantNotContain: []string{"backoff"},
		},
		{
			give:           "escalate action",
			giveAction:     "escalate",
			giveCauseClass: "unrecoverable",
			giveAttempt:    1,
			giveBackoff:    0,
			wantContain:    []string{"Escalate", "#1", "unrecoverable"},
		},
		{
			give:           "unknown action falls back to Recovery",
			giveAction:     "unknown",
			giveCauseClass: "mystery",
			giveAttempt:    1,
			giveBackoff:    0,
			wantContain:    []string{"Recovery", "#1", "mystery"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := renderRecoveryBlock(tt.giveAction, tt.giveCauseClass, tt.giveAttempt, tt.giveBackoff, 80)
			assert.NotEmpty(t, got)
			for _, want := range tt.wantContain {
				assert.Contains(t, got, want)
			}
			for _, notWant := range tt.wantNotContain {
				assert.NotContains(t, got, notWant)
			}
		})
	}
}

func TestRenderRecoveryBlock_ContainsIcon(t *testing.T) {
	got := renderRecoveryBlock("retry", "rate_limit", 1, 0, 80)
	assert.Contains(t, got, "\U0001F504", "should contain recycle icon")
}

func TestRenderRecoveryBlock_DirectAnswer(t *testing.T) {
	got := renderRecoveryBlock("direct_answer", "budget_exceeded", 1, 0, 80)
	assert.Contains(t, got, "Direct Answer")
}

func TestRecoveryActionDisplayName(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{give: "retry", want: "Retry"},
		{give: "retry_with_hint", want: "Reroute"},
		{give: "direct_answer", want: "Direct Answer"},
		{give: "escalate", want: "Escalate"},
		{give: "anything_else", want: "Recovery"},
		{give: "", want: "Recovery"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := recoveryActionDisplayName(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}
