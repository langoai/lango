package chat

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/tui"
)

func TestToolStateVisual(t *testing.T) {
	tests := []struct {
		give      string
		state     ToolItemState
		wantIcon  string
		wantColor lipgloss.Color
	}{
		{
			give:      "running state returns gear icon and warning color",
			state:     toolStateRunning,
			wantIcon:  "\u2699",
			wantColor: tui.Warning,
		},
		{
			give:      "success state returns checkmark icon and success color",
			state:     toolStateSuccess,
			wantIcon:  "\u2713",
			wantColor: tui.Success,
		},
		{
			give:      "error state returns cross icon and error color",
			state:     toolStateError,
			wantIcon:  "\u2717",
			wantColor: tui.Error,
		},
		{
			give:      "canceled state returns circle-slash icon and muted color",
			state:     toolStateCanceled,
			wantIcon:  "\u2298",
			wantColor: tui.Muted,
		},
		{
			give:      "awaiting_approval state returns lock icon and warning color",
			state:     toolStateAwaitingApproval,
			wantIcon:  "\U0001F512",
			wantColor: tui.Warning,
		},
		{
			give:      "unknown state falls back to gear icon and muted color",
			state:     ToolItemState("unknown"),
			wantIcon:  "\u2699",
			wantColor: tui.Muted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			icon, color := toolStateVisual(tt.state)
			assert.Equal(t, tt.wantIcon, icon)
			assert.Equal(t, tt.wantColor, color)
		})
	}
}

func TestRenderToolBlock_AllStates(t *testing.T) {
	tests := []struct {
		give     string
		state    ToolItemState
		wantIcon string
	}{
		{
			give:     "running contains gear icon",
			state:    toolStateRunning,
			wantIcon: "⚙",
		},
		{
			give:     "success contains checkmark icon",
			state:    toolStateSuccess,
			wantIcon: "✓",
		},
		{
			give:     "error contains cross icon",
			state:    toolStateError,
			wantIcon: "✗",
		},
		{
			give:     "canceled contains circle-slash icon",
			state:    toolStateCanceled,
			wantIcon: "⊘",
		},
		{
			give:     "awaiting_approval contains lock icon",
			state:    toolStateAwaitingApproval,
			wantIcon: "🔒",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result := renderToolBlock("test_tool", tt.state, "1.2s", "", 80)
			assert.Contains(t, result, tt.wantIcon)
		})
	}
}

func TestRenderToolBlock_OutputTruncation(t *testing.T) {
	longOutput := strings.Repeat("a", 200)
	result := renderToolBlock("tool", toolStateSuccess, "1s", longOutput, 80)

	// The output line should contain the ellipsis character from truncation.
	assert.Contains(t, result, "…")
	// The original 200-char string should NOT appear in full.
	assert.NotContains(t, result, longOutput)
}

func TestRenderToolBlock_EmptyOutput(t *testing.T) {
	result := renderToolBlock("tool", toolStateSuccess, "1s", "", 80)

	// With empty output, the result should be a single line (no newline).
	require.NotEmpty(t, result)
	assert.NotContains(t, result, "\n")
}

func TestRenderToolBlock_NarrowWidth(t *testing.T) {
	// width=15 should not panic and should produce valid output.
	require.NotPanics(t, func() {
		result := renderToolBlock("tool", toolStateSuccess, "1s", "some output", 15)
		assert.NotEmpty(t, result)
	})
}

func TestRenderToolBlock_ZeroWidth(t *testing.T) {
	// width=0 should not panic.
	require.NotPanics(t, func() {
		result := renderToolBlock("tool", toolStateSuccess, "1s", "some output", 0)
		assert.NotEmpty(t, result)
	})
}

func TestRenderToolBlock_MultilineOutput(t *testing.T) {
	multiline := "line one\nline two\nline three"
	result := renderToolBlock("tool", toolStateSuccess, "1s", multiline, 80)

	// The output section should have newlines replaced with spaces.
	// Split by the first newline (which separates the header from the output line).
	parts := strings.SplitN(result, "\n", 2)
	require.Len(t, parts, 2, "expected header + output line")

	outputLine := parts[1]
	// The rendered output line should not contain literal newlines from the original.
	assert.NotContains(t, outputLine, "\n")
	// It should contain the words from all lines joined by spaces.
	assert.Contains(t, outputLine, "line one line two line three")
}

func TestRenderToolBlock_UnicodeOutput(t *testing.T) {
	// Korean chars are double-width, so truncation must use visual width.
	koreanOutput := strings.Repeat("가", 100)
	result := renderToolBlock("tool", toolStateSuccess, "1s", koreanOutput, 80)

	// Should be truncated (contains ellipsis).
	assert.Contains(t, result, "…")
	// The full 100-char Korean string should not appear.
	assert.NotContains(t, result, koreanOutput)

	// Emoji output should also truncate correctly.
	emojiOutput := strings.Repeat("🎉", 100)
	result2 := renderToolBlock("tool", toolStateSuccess, "1s", emojiOutput, 80)
	assert.Contains(t, result2, "…")
	assert.NotContains(t, result2, emojiOutput)
}
