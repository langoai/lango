package memory

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/types"
)

func BenchmarkEstimateTokens(b *testing.B) {
	tests := []struct {
		name string
		give string
	}{
		{"Short", "Hello, world!"},
		{"Medium", strings.Repeat("word ", 100)},
		{"Long", strings.Repeat("word ", 1000)},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				EstimateTokens(tt.give)
			}
		})
	}
}

func BenchmarkCountMessageTokens(b *testing.B) {
	tests := []struct {
		name string
		give session.Message
	}{
		{
			name: "Simple",
			give: session.Message{
				Role:      types.RoleUser,
				Content:   "What is the weather today?",
				Timestamp: time.Now(),
			},
		},
		{
			name: "WithToolCalls",
			give: session.Message{
				Role:    types.RoleAssistant,
				Content: "Let me check that for you.",
				ToolCalls: []session.ToolCall{
					{ID: "call_1", Name: "weather", Input: `{"city":"Seoul"}`, Output: `{"temp":22,"condition":"sunny"}`},
					{ID: "call_2", Name: "calendar", Input: `{"date":"today"}`, Output: `{"events":["meeting at 3pm"]}`},
				},
				Timestamp: time.Now(),
			},
		},
		{
			name: "LargeContent",
			give: session.Message{
				Role:      types.RoleAssistant,
				Content:   strings.Repeat("This is a long response with detailed information. ", 100),
				Timestamp: time.Now(),
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				CountMessageTokens(tt.give)
			}
		})
	}
}

func BenchmarkCountMessagesTokens(b *testing.B) {
	sizes := []int{5, 20, 100}

	for _, size := range sizes {
		msgs := make([]session.Message, size)
		for i := range msgs {
			msgs[i] = session.Message{
				Role:      types.RoleUser,
				Content:   fmt.Sprintf("Message number %d with some content to estimate tokens for.", i),
				Timestamp: time.Now(),
			}
		}

		b.Run(fmt.Sprintf("Messages_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				CountMessagesTokens(msgs)
			}
		})
	}
}
