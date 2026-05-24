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
			result := renderToolBlock("test_tool", tt.state, "1.2s", "", "", 80)
			assert.Contains(t, result, tt.wantIcon)
		})
	}
}

func TestRenderToolBlock_OutputTruncation(t *testing.T) {
	longOutput := strings.Repeat("a", 200)
	result := renderToolBlock("tool", toolStateSuccess, "1s", "", longOutput, 80)

	// The output line should contain the ellipsis character from truncation.
	assert.Contains(t, result, "…")
	// The original 200-char string should NOT appear in full.
	assert.NotContains(t, result, longOutput)
}

func TestRenderToolBlock_EmptyOutput(t *testing.T) {
	result := renderToolBlock("tool", toolStateSuccess, "1s", "", "", 80)

	// With empty output, the result should be a single line (no newline).
	require.NotEmpty(t, result)
	assert.NotContains(t, result, "\n")
}

func TestRenderToolBlock_NarrowWidth(t *testing.T) {
	// width=15 should not panic and should produce valid output.
	require.NotPanics(t, func() {
		result := renderToolBlock("tool", toolStateSuccess, "1s", "", "some output", 15)
		assert.NotEmpty(t, result)
		for _, line := range strings.Split(result, "\n") {
			assert.LessOrEqual(t, lipgloss.Width(line), 15)
		}
	})
}

func TestRenderToolBlock_ZeroWidth(t *testing.T) {
	// width=0 should not panic.
	require.NotPanics(t, func() {
		result := renderToolBlock("tool", toolStateSuccess, "1s", "", "some output", 0)
		assert.NotEmpty(t, result)
		for _, line := range strings.Split(result, "\n") {
			assert.LessOrEqual(t, lipgloss.Width(line), 10)
		}
	})
}

func TestRenderToolBlock_MultilineOutput(t *testing.T) {
	multiline := "line one\nline two\tline three"
	result := renderToolBlock("tool", toolStateSuccess, "1s", "", multiline, 80)

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

func TestRenderToolBlock_SanitizesOutput(t *testing.T) {
	result := renderToolBlock("tool", toolStateSuccess, "1s", "", "\x1b[31mline one\nline two\x1b[0m", 80)
	parts := strings.SplitN(result, "\n", 2)
	require.Len(t, parts, 2)
	assert.Contains(t, parts[1], "line one line two")
	assert.NotContains(t, result, "\x1b[31m")
	assert.NotContains(t, result, "\x1b[0m")
}

func TestRenderToolBlock_UnicodeOutput(t *testing.T) {
	// Korean chars are double-width, so truncation must use visual width.
	koreanOutput := strings.Repeat("가", 100)
	result := renderToolBlock("tool", toolStateSuccess, "1s", "", koreanOutput, 80)

	// Should be truncated (contains ellipsis).
	assert.Contains(t, result, "…")
	// The full 100-char Korean string should not appear.
	assert.NotContains(t, result, koreanOutput)

	// Emoji output should also truncate correctly.
	emojiOutput := strings.Repeat("🎉", 100)
	result2 := renderToolBlock("tool", toolStateSuccess, "1s", "", emojiOutput, 80)
	assert.Contains(t, result2, "…")
	assert.NotContains(t, result2, emojiOutput)
}

func TestRenderToolBlock_RunningPreviewUsesStableKeyOrder(t *testing.T) {
	preview := formatParamPreview(map[string]any{
		"zeta":  3,
		"alpha": 1,
		"beta":  2,
	})
	result := renderToolBlock("tool", toolStateRunning, "", preview, "", 120)

	alphaIdx := strings.Index(result, "alpha=1")
	betaIdx := strings.Index(result, "beta=2")
	zetaIdx := strings.Index(result, "zeta=3")
	require.NotEqual(t, -1, alphaIdx)
	require.NotEqual(t, -1, betaIdx)
	require.NotEqual(t, -1, zetaIdx)
	assert.Less(t, alphaIdx, betaIdx)
	assert.Less(t, betaIdx, zetaIdx)
}

func TestRenderToolBlock_SuccessKeepsPreview(t *testing.T) {
	preview := "alpha=1  beta=2"
	result := renderToolBlock("tool", toolStateSuccess, "1s", preview, "done", 120)
	assert.Contains(t, result, "alpha=1")
	assert.Contains(t, result, "beta=2")
	assert.Contains(t, result, "done")
}

func TestFormatParamPreview_NestedValuesRenderDeterministically(t *testing.T) {
	preview := formatParamPreview(map[string]any{
		"config": map[string]any{
			"zeta":  3,
			"alpha": 1,
		},
	})
	assert.Equal(t, `config={"alpha":1,"zeta":3}`, preview)
}

func TestFormatParamPreview_MultilineStringStaysSingleLine(t *testing.T) {
	preview := formatParamPreview(map[string]any{
		"content": "line one\nline two\nline three",
	})
	assert.Equal(t, "content=line one line two line three", preview)
}

func TestFormatParamPreview_SanitizesParamKey(t *testing.T) {
	preview := formatParamPreview(map[string]any{
		"path\x1b[31m\nops\x1b[0m": "/tmp/test.txt",
	})
	assert.Equal(t, "path ops=/tmp/test.txt", preview)
}

func TestRenderToolBlock_ErrorKeepsPreview(t *testing.T) {
	preview := "alpha=1  beta=2"
	result := renderToolBlock("tool", toolStateError, "1s", preview, "boom", 120)
	assert.Contains(t, result, "alpha=1")
	assert.Contains(t, result, "beta=2")
	assert.Contains(t, result, "boom")
}

func TestRenderToolBlock_AwaitingApprovalKeepsPreview(t *testing.T) {
	preview := "alpha=1  beta=2"
	result := renderToolBlock("tool", toolStateAwaitingApproval, "", preview, "", 120)
	assert.Contains(t, result, "awaiting approval")
	assert.Contains(t, result, "alpha=1")
	assert.Contains(t, result, "beta=2")
}

func TestRenderToolBlock_CanceledKeepsPreview(t *testing.T) {
	preview := "alpha=1  beta=2"
	result := renderToolBlock("tool", toolStateCanceled, "", preview, "", 120)
	assert.Contains(t, result, "canceled")
	assert.Contains(t, result, "alpha=1")
	assert.Contains(t, result, "beta=2")
}

func TestRenderToolBlock_NarrowPreviewLineStaysWidthSafe(t *testing.T) {
	result := renderToolBlock("tool", toolStateRunning, "", "line one\nline two\tline three", "", 18)
	lines := strings.Split(result, "\n")
	require.Len(t, lines, 2)
	for _, line := range lines {
		assert.LessOrEqual(t, lipgloss.Width(line), 18)
	}
	assert.Contains(t, lines[1], "line one line…")
}

func TestRenderToolBlock_SanitizesPreview(t *testing.T) {
	result := renderToolBlock("tool", toolStateRunning, "", "\x1b[31mline one\nline two\x1b[0m", "", 80)
	parts := strings.SplitN(result, "\n", 2)
	require.Len(t, parts, 2)
	assert.Contains(t, parts[1], "line one line two")
	assert.NotContains(t, result, "\x1b[31m")
	assert.NotContains(t, result, "\x1b[0m")
}

func TestRenderToolBlock_NarrowOutputLineStaysWidthSafe(t *testing.T) {
	result := renderToolBlock("tool", toolStateError, "1s", "", "line one\nline two\tline three", 18)
	lines := strings.Split(result, "\n")
	require.Len(t, lines, 2)
	for _, line := range lines {
		assert.LessOrEqual(t, lipgloss.Width(line), 18)
	}
	assert.Contains(t, lines[1], "line one line…")
}

func TestRenderToolBlock_SanitizesToolName(t *testing.T) {
	result := renderToolBlock("fs_\x1b[31mread\nops\x1b[0m", toolStateRunning, "", "", "", 120)
	assert.Contains(t, result, "fs_read ops")
	assert.NotContains(t, result, "\x1b[31m")
	assert.NotContains(t, result, "\x1b[0m")
}
