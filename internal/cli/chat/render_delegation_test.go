package chat

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderDelegationBlock(t *testing.T) {
	tests := []struct {
		give           string
		giveFrom       string
		giveTo         string
		giveReason     string
		giveWidth      int
		wantContain    []string
		wantNotContain []string
	}{
		{
			give:        "basic delegation without reason",
			giveFrom:    "operator",
			giveTo:      "librarian",
			giveReason:  "",
			giveWidth:   80,
			wantContain: []string{"operator", "\u2192", "librarian"},
		},
		{
			give:        "delegation with reason",
			giveFrom:    "operator",
			giveTo:      "librarian",
			giveReason:  "search query needed",
			giveWidth:   80,
			wantContain: []string{"operator", "\u2192", "librarian", "search query needed"},
		},
		{
			give:        "empty from and to does not crash",
			giveFrom:    "",
			giveTo:      "",
			giveReason:  "",
			giveWidth:   80,
			wantContain: []string{"\u2192"},
		},
		{
			give:        "long reason is truncated",
			giveFrom:    "a",
			giveTo:      "b",
			giveReason:  strings.Repeat("abcdefghij", 20),
			giveWidth:   40,
			wantContain: []string{"\u2026"},
		},
		{
			give:        "narrow width still renders",
			giveFrom:    "operator",
			giveTo:      "librarian",
			giveReason:  "some reason",
			giveWidth:   20,
			wantContain: []string{"operator", "librarian"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := renderDelegationBlock(tt.giveFrom, tt.giveTo, tt.giveReason, tt.giveWidth)
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

func TestRenderDelegationBlock_ZeroWidth(t *testing.T) {
	got := renderDelegationBlock("a", "b", "reason", 0)
	assert.NotEmpty(t, got, "should not panic with zero width")
}

func TestRenderDelegationBlock_ContainsIcon(t *testing.T) {
	got := renderDelegationBlock("from", "to", "", 80)
	assert.Contains(t, got, "\U0001F500", "should contain shuffle icon")
}
