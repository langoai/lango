package chat

import (
	"errors"
	"testing"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestRenderMarkdown_FallsBackOnRendererError(t *testing.T) {
	original := renderWithGlamour
	t.Cleanup(func() { renderWithGlamour = original })

	renderWithGlamour = func(_ *glamour.TermRenderer, _ string) (string, error) {
		return "", errors.New("renderer failed")
	}

	got := renderMarkdown("plain text", 80)
	assert.Equal(t, "plain text", got)
}

func TestRenderMarkdown_FallsBackOnRendererPanic(t *testing.T) {
	original := renderWithGlamour
	t.Cleanup(func() { renderWithGlamour = original })

	renderWithGlamour = func(_ *glamour.TermRenderer, _ string) (string, error) {
		panic("renderer panic")
	}

	got := renderMarkdown("plain text", 80)
	assert.Equal(t, "plain text", got)
}

func TestRenderMarkdown_StripsEscapeSequencesBeforeRendering(t *testing.T) {
	got := renderMarkdown("\x1b[31mplain text\x1b[0m", 80)
	assert.Contains(t, ansi.Strip(got), "plain text")
	assert.NotContains(t, got, "\x1b[31mplain text")
}

func TestRenderMarkdown_FallbackUsesSanitizedText(t *testing.T) {
	original := renderWithGlamour
	t.Cleanup(func() { renderWithGlamour = original })

	renderWithGlamour = func(_ *glamour.TermRenderer, _ string) (string, error) {
		return "", errors.New("renderer failed")
	}

	got := renderMarkdown("\x1b[31mplain text\x1b[0m", 80)
	assert.Equal(t, "plain text", got)
}
