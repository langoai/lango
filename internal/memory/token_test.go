package memory

import (
	"testing"

	"github.com/langoai/lango/internal/session"
	"github.com/stretchr/testify/assert"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		give string
		want int
	}{
		{
			give: "",
			want: 0,
		},
		{
			give: "Hello world",
			want: 2, // 11 ASCII chars / 4 = 2 (integer division)
		},
		{
			give: "안녕하세요",
			want: 2, // 5 Korean runes / 2 = 2 (integer division)
		},
		{
			give: "Hello 안녕",
			want: 2, // 6 ASCII chars / 4 = 1, plus 2 Korean runes / 2 = 1 => 2
		},
		{
			give: "The quick brown fox jumps over the lazy dog",
			want: 10, // 43 ASCII chars / 4 = 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := EstimateTokens(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCountMessageTokens(t *testing.T) {
	tests := []struct {
		give string
		msg  session.Message
		want int
	}{
		{
			give: "content only",
			msg: session.Message{
				Role:    "user",
				Content: "Hello world",
			},
			want: 4 + 2, // overhead + 11/4
		},
		{
			give: "with tool calls",
			msg: session.Message{
				Role:    "assistant",
				Content: "result",
				ToolCalls: []session.ToolCall{
					{
						ID:     "tc01",
						Name:   "search",
						Input:  `{"query":"test"}`,
						Output: `{"results":[]}`,
					},
				},
			},
			// overhead(4) + "result"(1) + "tc01"(1) + "search"(1) + input(4) + output(3) = 14
			want: 14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := CountMessageTokens(tt.msg)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCountMessagesTokens(t *testing.T) {
	msgs := []session.Message{
		{Role: "user", Content: "Hello world"},
		{Role: "assistant", Content: "Hi there!!"},
	}
	// msg1: 4 + 2 = 6
	// msg2: 4 + 2 = 6 ("Hi there!!" is 10 chars / 4 = 2)
	// total: 12
	got := CountMessagesTokens(msgs)
	assert.Equal(t, 12, got)
}
