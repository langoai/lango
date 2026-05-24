package chat

import (
	"sync"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/x/ansi"
)

// cachedRenderer holds a width-keyed glamour renderer so that repeated
// renderMarkdown calls at the same width (e.g., every 400ms cursor tick)
// reuse the renderer instead of rebuilding it. Bubbletea dispatches
// messages on a single goroutine in production, but tests exercise the
// renderer from parallel goroutines, so a mutex is required — glamour's
// internal BlockStack is not safe for concurrent use.
var (
	rendererMu          sync.Mutex
	cachedRenderer      *glamour.TermRenderer
	cachedRendererWidth int
	renderWithGlamour   func(r *glamour.TermRenderer, content string) (string, error) = func(r *glamour.TermRenderer, content string) (string, error) {
		return r.Render(content)
	}
)

func sanitizeMarkdownInput(content string) string {
	return ansi.Strip(content)
}

// renderMarkdown renders markdown content using a fixed dark Glamour style.
// This intentionally avoids auto-style terminal background probing, which can
// emit OSC 11 responses that leak into composer input on some terminals.
func renderMarkdown(content string, width int) (out string) {
	content = sanitizeMarkdownInput(content)
	out = content
	defer func() {
		if recover() != nil {
			out = content
		}
	}()

	if content == "" {
		return ""
	}
	if width < 10 {
		width = 10
	}
	rendererMu.Lock()
	defer rendererMu.Unlock()
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
	out, err := renderWithGlamour(cachedRenderer, content)
	if err != nil {
		return content
	}
	return out
}
