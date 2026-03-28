package pages

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/langoai/lango/internal/cli/settings"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/configstore"
)

// SettingsPage embeds a settings.Editor in the cockpit.
// Save operations use the OnSave callback instead of quitting.
type SettingsPage struct {
	editor *settings.Editor
}

// NewSettingsPage creates a SettingsPage with an embedded Editor.
// The save callback persists to ConfigStore without exiting the TUI.
func NewSettingsPage(
	cfg *config.Config,
	store *configstore.Store,
	profileName string,
) *SettingsPage {
	onSave := func(cfg *config.Config, dirtyKeys map[string]bool) error {
		return store.Save(context.Background(), profileName, cfg, dirtyKeys)
	}
	return &SettingsPage{
		editor: settings.NewEditorForEmbedding(cfg, onSave),
	}
}

// Title returns the page tab label.
func (m *SettingsPage) Title() string { return "Settings" }

// ShortHelp returns key bindings for the help bar.
func (m *SettingsPage) ShortHelp() []key.Binding { return nil }

// Init satisfies tea.Model.
func (m *SettingsPage) Init() tea.Cmd { return nil }

// Activate is a no-op — the editor is always ready.
func (m *SettingsPage) Activate() tea.Cmd { return nil }

// Deactivate is a no-op.
func (m *SettingsPage) Deactivate() {}

// Update delegates all messages to the embedded Editor.
func (m *SettingsPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.editor.Update(msg)
	m.editor = updated.(*settings.Editor)
	return m, cmd
}

// View delegates rendering to the embedded Editor.
func (m *SettingsPage) View() string {
	return m.editor.View()
}
