package chat

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/stretchr/testify/assert"
)

func TestChannelColor_KnownChannels(t *testing.T) {
	tests := []struct {
		give    string
		wantNot lipgloss.Color
	}{
		{give: "telegram", wantNot: tui.Muted},
		{give: "discord", wantNot: tui.Muted},
		{give: "slack", wantNot: tui.Muted},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := channelColor(tt.give)
			assert.NotEqual(t, tt.wantNot, got, "known channel %q should not return tui.Muted", tt.give)
		})
	}

	// Each known channel returns a distinct color.
	colors := map[lipgloss.Color]string{}
	for _, ch := range []string{"telegram", "discord", "slack"} {
		c := channelColor(ch)
		if existing, ok := colors[c]; ok {
			t.Errorf("channels %q and %q share the same color %v", existing, ch, c)
		}
		colors[c] = ch
	}
}

func TestChannelColor_Unknown(t *testing.T) {
	got := channelColor("unknown-platform")
	assert.Equal(t, tui.Muted, got)
}

func TestRenderChannelBlock_Telegram(t *testing.T) {
	out := renderChannelBlock("hello world", "telegram", "alice", 80)
	assert.Contains(t, out, "telegram")
}

func TestRenderChannelBlock_Discord(t *testing.T) {
	out := renderChannelBlock("hello world", "discord", "bob", 80)
	assert.Contains(t, out, "discord")
}

func TestRenderChannelBlock_WithSender(t *testing.T) {
	out := renderChannelBlock("hello world", "telegram", "alice", 80)
	assert.Contains(t, out, "@alice")
}

func TestRenderChannelBlock_EmptySender(t *testing.T) {
	out := renderChannelBlock("hello world", "telegram", "", 80)
	assert.NotContains(t, out, "@")
}

func TestRenderChannelBlock_NarrowWidth(t *testing.T) {
	out := renderChannelBlock("some text here", "slack", "user", 20)
	assert.NotEmpty(t, out)
	assert.LessOrEqual(t, lipgloss.Width(out), 20)
}

func TestRenderChannelBlock_ZeroWidth(t *testing.T) {
	out := renderChannelBlock("some text here", "discord", "user", 0)
	assert.NotEmpty(t, out)
	assert.LessOrEqual(t, lipgloss.Width(out), 10)
}

func TestRenderChannelBlock_LongText(t *testing.T) {
	longText := strings.Repeat("abcdefghij", 20) // 200 chars
	out := renderChannelBlock(longText, "telegram", "alice", 80)
	assert.Contains(t, out, "…")
}

func TestRenderChannelBlock_SingleLine(t *testing.T) {
	out := renderChannelBlock("short message", "slack", "bob", 120)
	assert.Equal(t, 1, lipgloss.Height(out), "channel block should be a single line")
}

func TestRenderChannelBlock_MultilinePayloadCollapsesWhitespace(t *testing.T) {
	out := renderChannelBlock("line one\nline two\t line three", "telegram", "alice\nops", 120)
	assert.Equal(t, 1, lipgloss.Height(out), "channel block should stay single-line")
	assert.Contains(t, out, "@alice ops")
	assert.Contains(t, out, "line one line two line three")
}

func TestRenderChannelBlock_LongSenderStaysWidthSafe(t *testing.T) {
	out := renderChannelBlock("hello", "discord", strings.Repeat("operator", 8), 24)
	assert.LessOrEqual(t, lipgloss.Width(out), 24)
	assert.Contains(t, out, "…")
}

func TestRenderChannelBlock_StripsSenderEscapeSequences(t *testing.T) {
	out := renderChannelBlock("hello world", "telegram", "\x1b[31malice\x1b[0m", 80)
	assert.Contains(t, out, "@alice")
	assert.NotContains(t, out, "\x1b[31m")
	assert.NotContains(t, out, "\x1b[0m")
}

func TestRenderChannelBlock_SanitizesChannelBadgeText(t *testing.T) {
	out := renderChannelBlock("hello world", "\x1b[31mtele\ngram\x1b[0m", "alice", 120)
	assert.Contains(t, out, "tele gram")
	assert.NotContains(t, out, "\x1b[31m")
	assert.NotContains(t, out, "\x1b[0m")
}

func TestChannelColor_UsesSanitizedKnownChannelName(t *testing.T) {
	got := channelColor(sanitizeDisplayText("\x1b[31mtelegram\x1b[0m"))
	assert.Equal(t, lipgloss.Color("#0088cc"), got)
}
