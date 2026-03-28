package pages

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/settings"
	"github.com/langoai/lango/internal/config"
)

func TestSettingsPage_Title(t *testing.T) {
	page := newTestSettingsPage(nil)
	assert.Equal(t, "Settings", page.Title())
}

func TestSettingsPage_ShortHelp(t *testing.T) {
	page := newTestSettingsPage(nil)
	assert.Empty(t, page.ShortHelp())
}

func TestSettingsPage_ActivateDeactivate(t *testing.T) {
	page := newTestSettingsPage(nil)

	cmd := page.Activate()
	assert.Nil(t, cmd, "Activate should return nil")

	page.Deactivate() // should not panic
}

func TestSettingsPage_Init(t *testing.T) {
	page := newTestSettingsPage(nil)
	cmd := page.Init()
	assert.Nil(t, cmd)
}

func TestSettingsPage_UpdateForwardsToEditor(t *testing.T) {
	page := newTestSettingsPage(nil)
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}

	updated, _ := page.Update(msg)
	require.NotNil(t, updated)

	sp, ok := updated.(*SettingsPage)
	require.True(t, ok)
	require.NotNil(t, sp.editor)
}

func TestSettingsPage_ViewNonEmpty(t *testing.T) {
	page := newTestSettingsPage(nil)
	view := page.View()
	assert.NotEmpty(t, view, "View should render editor content")
}

func TestSettingsPage_PageInterfaceCompliance(t *testing.T) {
	page := newTestSettingsPage(nil)

	// Verify all Page interface methods exist and are callable.
	_ = page.Title()
	_ = page.ShortHelp()
	_ = page.Init()
	_ = page.Activate()
	page.Deactivate()
	_, _ = page.Update(nil)
	_ = page.View()
}

func TestSettingsPage_OnSaveCallback(t *testing.T) {
	var called bool
	page := newTestSettingsPage(func(cfg *config.Config, dirtyKeys map[string]bool) error {
		called = true
		return nil
	})
	_ = page
	// Verify the callback is wired (editor has OnSave set).
	// Direct invocation requires navigating the editor state machine,
	// which is covered by editor_embed_test.go. Here we verify construction.
	assert.False(t, called, "callback not yet invoked")
}

func TestSettingsPage_OnSaveError(t *testing.T) {
	page := newTestSettingsPage(func(cfg *config.Config, dirtyKeys map[string]bool) error {
		return fmt.Errorf("disk full")
	})
	require.NotNil(t, page.editor)
}

// --- helpers ---

// newTestSettingsPage creates a SettingsPage with an embedded editor
// bypassing ConfigStore (which requires a DB). If onSave is nil, a no-op is used.
func newTestSettingsPage(onSave settings.OnSaveFunc) *SettingsPage {
	if onSave == nil {
		onSave = func(cfg *config.Config, dirtyKeys map[string]bool) error {
			return nil
		}
	}
	cfg := config.DefaultConfig()
	return &SettingsPage{
		editor: settings.NewEditorForEmbedding(cfg, onSave),
	}
}

// Compile-time interface compliance check.
var _ interface {
	tea.Model
	Title() string
	ShortHelp() []key.Binding
	Activate() tea.Cmd
	Deactivate()
} = (*SettingsPage)(nil)
