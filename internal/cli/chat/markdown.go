package chat

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
)

// cachedRenderer holds a width-keyed glamour renderer so that repeated
// renderMarkdown calls at the same width (e.g., every 400ms cursor tick)
// reuse the renderer instead of rebuilding it. Bubbletea dispatches
// messages on a single goroutine, so no synchronization is needed.
var (
	cachedRenderer      *glamour.TermRenderer
	cachedRendererWidth int
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
	if cachedRenderer == nil || cachedRendererWidth != width {
		r, err := glamour.NewTermRenderer(
			glamour.WithStandardStyle(styles.DarkStyle),
			glamour.WithWordWrap(width),
		)
		if err != nil {
			return content
		}
		cachedRenderer = r
		cachedRendererWidth = width
	}
	out, err := cachedRenderer.Render(content)
	if err != nil {
		return content
	}
	return out
}
