package chat

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderThinkingBlock_Active(t *testing.T) {
	tests := []struct {
		give           string
		giveContent    string
		giveWidth      int
		wantContain    []string
		wantNotContain []string
	}{
		{
			give:        "active with content shows preview",
			giveContent: "analyzing data",
			giveWidth:   80,
			wantContain: []string{"💭", "Thinking...", "analyzing data"},
		},
		{
			give:           "active empty content no extra space",
			giveContent:    "",
			giveWidth:      80,
			wantContain:    []string{"💭", "Thinking..."},
			wantNotContain: []string{"Thinking...  "},
		},
		{
			give:        "active long content truncated",
			giveContent: "very long summary text that exceeds width and should be truncated properly",
			giveWidth:   40,
			wantContain: []string{"💭", "Thinking...", "…"},
		},
		{
			give:        "active narrow width still works",
			giveContent: "some content here",
			giveWidth:   20,
			wantContain: []string{"💭", "Thinking..."},
		},
		{
			give:        "active with different content",
			giveContent: "reasoning about approach",
			giveWidth:   80,
			wantContain: []string{"💭", "Thinking...", "reasoning about approach"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := renderThinkingBlock(tt.giveContent, "active", "", tt.giveWidth)
			for _, want := range tt.wantContain {
				assert.Contains(t, got, want)
			}
			for _, notWant := range tt.wantNotContain {
				assert.NotContains(t, got, notWant)
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
	// Done state shows duration but ignores content (unchanged behavior).
	got := renderThinkingBlock("summary of reasoning", "done", "5s", 80)
	assert.Contains(t, got, "💭")
	assert.Contains(t, got, "5s")
	assert.Contains(t, got, "Thinking")
	assert.NotContains(t, got, "summary of reasoning")
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
