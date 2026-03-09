package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatMrkdwn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want string
	}{
		{
			give: "**bold text**",
			want: "*bold text*",
		},
		{
			give: "~~strikethrough~~",
			want: "~strikethrough~",
		},
		{
			give: "[click here](https://example.com)",
			want: "<https://example.com|click here>",
		},
		{
			give: "# Heading",
			want: "*Heading*",
		},
		{
			give: "## Sub Heading",
			want: "*Sub Heading*",
		},
		{
			give: "`inline code`",
			want: "`inline code`",
		},
		{
			give: "",
			want: "",
		},
		{
			give: "plain text",
			want: "plain text",
		},
		{
			give: "```\n**bold inside code**\n~~strike~~\n```",
			want: "```\n**bold inside code**\n~~strike~~\n```",
		},
		{
			give: "before\n```\n**code**\n```\nafter **bold**",
			want: "before\n```\n**code**\n```\nafter *bold*",
		},
		{
			give: "Hello **bold** and ~~strike~~ with [link](https://go.dev)",
			want: "Hello *bold* and ~strike~ with <https://go.dev|link>",
		},
		{
			give: "# Title\nSome **bold** and [docs](https://docs.go.dev)\n- ~~old~~ item",
			want: "*Title*\nSome *bold* and <https://docs.go.dev|docs>\n- ~old~ item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got := FormatMrkdwn(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}
