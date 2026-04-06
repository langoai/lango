package tui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		give       string
		giveMaxLen int
		want       string
	}{
		{give: "hello", giveMaxLen: 10, want: "hello"},
		{give: "hello", giveMaxLen: 5, want: "hello"},
		{give: "hello world", giveMaxLen: 8, want: "hello..."},
		{give: "abcdef", giveMaxLen: 3, want: "abc"},
		{give: "abcdef", giveMaxLen: 2, want: "ab"},
		{give: "", giveMaxLen: 5, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, Truncate(tt.give, tt.giveMaxLen))
		})
	}
}

func TestWordWrap(t *testing.T) {
	tests := []struct {
		give      string
		giveWidth int
		want      string
	}{
		{give: "", giveWidth: 10, want: ""},
		{give: "hello world", giveWidth: 0, want: "hello world"},
		{give: "hello world", giveWidth: 20, want: "hello world"},
		{give: "hello world", giveWidth: 5, want: "hello\nworld"},
		{give: "a b c d", giveWidth: 3, want: "a b\nc d"},
		{give: "line1\nline2", giveWidth: 20, want: "line1\nline2"},
		{give: "line1\n\nline3", giveWidth: 20, want: "line1\n\nline3"},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, WordWrap(tt.give, tt.giveWidth))
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		give int64
		want string
	}{
		{give: 0, want: "0"},
		{give: 999, want: "999"},
		{give: 1000, want: "1,000"},
		{give: 12345, want: "12,345"},
		{give: 1234567, want: "1,234,567"},
		{give: -12345, want: "-12,345"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatNumber(tt.give))
		})
	}
}

func TestFormatTokens(t *testing.T) {
	tests := []struct {
		give int
		want string
	}{
		{give: 0, want: "0"},
		{give: 500, want: "500"},
		{give: 1500, want: "1,500"},
		{give: 1234567, want: "1,234,567"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatTokens(tt.give))
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		give time.Duration
		want string
	}{
		{give: 150 * time.Millisecond, want: "150ms"},
		{give: 0, want: "0ms"},
		{give: 3*time.Minute + 42*time.Second, want: "3m 42s"},
		{give: 2*time.Hour + 15*time.Minute, want: "2h 15m"},
		{give: 5 * time.Second, want: "0m 5s"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatDuration(tt.give))
		})
	}
}

func TestRelativeTime(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		give string
		giveT time.Time
		want  string
	}{
		{give: "seconds", giveT: now.Add(-5 * time.Second), want: "5s ago"},
		{give: "minutes", giveT: now.Add(-3 * time.Minute), want: "3m ago"},
		{give: "hours", giveT: now.Add(-2 * time.Hour), want: "2h ago"},
		{give: "days", giveT: now.Add(-48 * time.Hour), want: "2d ago"},
		{give: "future", giveT: now.Add(5 * time.Second), want: "0s ago"},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, RelativeTime(now, tt.giveT))
		})
	}
}

func TestRelativeTimeHuman(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		give  string
		giveT time.Time
		want  string
	}{
		{give: "just now", giveT: now.Add(-5 * time.Second), want: "just now"},
		{give: "minutes", giveT: now.Add(-3 * time.Minute), want: "3m ago"},
		{give: "hours", giveT: now.Add(-2 * time.Hour), want: "2h ago"},
		{give: "days", giveT: now.Add(-48 * time.Hour), want: "2d ago"},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, RelativeTimeHuman(now, tt.giveT))
		})
	}
}
