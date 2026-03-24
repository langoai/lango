package chat

import (
	"github.com/charmbracelet/glamour"
)

// renderMarkdown renders markdown content using glamour with automatic
// terminal style detection. During streaming, raw text is displayed; this
// function is called once on the final DoneMsg for polished rendering.
func renderMarkdown(content string, width int) string {
	if content == "" {
		return ""
	}
	if width < 10 {
		width = 10
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content
	}
	out, err := r.Render(content)
	if err != nil {
		return content
	}
	return out
}
