package chat

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
)

// renderMarkdown renders markdown content using a fixed dark Glamour style.
// This intentionally avoids auto-style terminal background probing, which can
// emit OSC 11 responses that leak into composer input on some terminals.
func renderMarkdown(content string, width int) string {
	if content == "" {
		return ""
	}
	if width < 10 {
		width = 10
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(styles.DarkStyle),
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
