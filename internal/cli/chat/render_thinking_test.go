package chat

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderThinkingBlock_Active(t *testing.T) {
	tests := []struct {
		give        string
		giveContent string
		wantContain []string
	}{
		{
			give:        "active with content",
			giveContent: "analyzing the problem",
			wantContain: []string{"💭", "Thinking..."},
		},
		{
			give:        "active with different content",
			giveContent: "reasoning about approach",
			wantContain: []string{"💭", "Thinking..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := renderThinkingBlock(tt.giveContent, "active", "", 80)
			for _, want := range tt.wantContain {
				assert.Contains(t, got, want)
			}
		})
	}
}

func TestRenderThinkingBlock_Done(t *testing.T) {
	got := renderThinkingBlock("", "done", "2s", 80)
	assert.Contains(t, got, "💭")
	assert.Contains(t, got, "2s")
	assert.Contains(t, got, "Thinking")
}

func TestRenderThinkingBlock_DoneWithContent(t *testing.T) {
	got := renderThinkingBlock("summary of reasoning", "done", "5s", 80)
	assert.Contains(t, got, "💭")
	assert.Contains(t, got, "5s")
	assert.Contains(t, got, "Thinking")
	// In done state, the duration is shown but content is not rendered inline
	// (the implementation uses duration label only).
}

func TestRenderThinkingBlock_EmptyContent(t *testing.T) {
	got := renderThinkingBlock("", "active", "", 80)
	assert.NotEmpty(t, got, "should produce valid output even with empty content")
	assert.Contains(t, got, "💭")
}

func TestRenderThinkingBlock_UnknownState(t *testing.T) {
	// Unknown state falls through to default branch which renders content directly.
	got := renderThinkingBlock("some fallback text", "unknown", "", 80)
	assert.NotEmpty(t, got, "should produce output without panic for unknown state")
	assert.Contains(t, got, "💭")
	assert.Contains(t, got, "some fallback text")
}

func TestRenderThinkingBlock_NarrowWidth(t *testing.T) {
	tests := []struct {
		give      string
		giveState string
		giveWidth int
	}{
		{give: "narrow active", giveState: "active", giveWidth: 10},
		{give: "narrow done", giveState: "done", giveWidth: 10},
		{give: "zero width", giveState: "active", giveWidth: 0},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := renderThinkingBlock("content", tt.giveState, "1s", tt.giveWidth)
			assert.NotEmpty(t, got, "should not panic or produce empty output with narrow width")
		})
	}
}

func TestRenderPendingIndicator(t *testing.T) {
	got := renderPendingIndicator("3s")
	assert.Contains(t, got, "Working")
	assert.Contains(t, got, "3s")
	assert.True(t, strings.Contains(got, "⏳"), "should contain hourglass emoji")
}

func TestRenderPendingIndicator_EmptyElapsed(t *testing.T) {
	got := renderPendingIndicator("")
	assert.NotEmpty(t, got, "should produce valid output even with empty elapsed")
	assert.Contains(t, got, "Working")
	assert.Contains(t, got, "⏳")
}
