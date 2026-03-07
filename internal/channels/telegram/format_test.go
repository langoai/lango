package telegram

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatMarkdown(t *testing.T) {
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
			give: "_italic text_",
			want: "_italic text_",
		},
		{
			give: "`inline code`",
			want: "`inline code`",
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
			give: "### Deep Heading",
			want: "*Deep Heading*",
		},
		{
			give: "~~strikethrough~~",
			want: "strikethrough",
		},
		{
			give: "Hello **bold** and ~~strike~~",
			want: "Hello *bold* and strike",
		},
		{
			give: "",
			want: "",
		},
		{
			give: "plain text without markdown",
			want: "plain text without markdown",
		},
		{
			give: "```\n**bold inside code**\n~~strike inside code~~\n```",
			want: "```\n**bold inside code**\n~~strike inside code~~\n```",
		},
		{
			give: "before\n```\n**code**\n```\nafter **bold**",
			want: "before\n```\n**code**\n```\nafter *bold*",
		},
		{
			give: "keep `**bold** inside inline code`",
			want: "keep `**bold** inside inline code`",
		},
		{
			give: "# Heading\nSome **bold** text\n- item ~~removed~~",
			want: "*Heading*\nSome *bold* text\n- item removed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			got := FormatMarkdown(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}
