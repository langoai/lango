package settings

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestEditorViewRendersPrimaryStepBranches(t *testing.T) {
	enabled := true
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"anthropic": {Type: "anthropic"},
		},
		Auth: config.AuthConfig{
			Providers: map[string]config.OIDCProviderConfig{
				"corp": {IssuerURL: "https://issuer.example"},
			},
		},
		MCP: config.MCPConfig{
			Servers: map[string]config.MCPServerConfig{
				"filesystem": {Transport: "stdio", Enabled: &enabled},
			},
		},
	}

	tests := []struct {
		name      string
		configure func(*Editor)
		want      []string
	}{
		{
			name: "welcome",
			configure: func(e *Editor) {
				e.step = StepWelcome
			},
			want: []string{"Settings", "Configure your agent", "categories", "Enter", "Start"},
		},
		{
			name: "menu",
			configure: func(e *Editor) {
				e.step = StepMenu
			},
			want: []string{"Settings", "Core", "Save & Exit"},
		},
		{
			name: "providers list",
			configure: func(e *Editor) {
				e.step = StepProvidersList
				e.providersList = NewProvidersListModel(e.state.Current)
			},
			want: []string{"Settings", "Providers", "anthropic (anthropic)", "+ Add New Provider"},
		},
		{
			name: "auth providers list",
			configure: func(e *Editor) {
				e.step = StepAuthProvidersList
				e.authProvidersList = NewAuthProvidersListModel(e.state.Current)
			},
			want: []string{"Settings", "Auth Providers", "corp (https://issuer.example)", "+ Add New OIDC Provider"},
		},
		{
			name: "mcp servers list",
			configure: func(e *Editor) {
				e.step = StepMCPServersList
				e.mcpServersList = NewMCPServersListModel(e.state.Current)
			},
			want: []string{"Settings", "MCP Servers", "filesystem (stdio) [enabled]", "+ Add New MCP Server"},
		},
		{
			name: "setup flow",
			configure: func(e *Editor) {
				e.step = StepSetupFlow
				e.setupFlow = NewSetupFlow("smartaccount", []DepResult{
					{
						Dependency: Dependency{CategoryID: "security", Label: "Security"},
						Status:     DepNotEnabled,
					},
				}, e.state)
			},
			want: []string{"Settings", "Setup", "smartaccount", "Guided Setup", "Security Configuration", "Ctrl+N"},
		},
		{
			name: "complete",
			configure: func(e *Editor) {
				e.step = StepComplete
			},
			want: []string{"Settings"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEditorWithConfig(cfg)
			tt.configure(e)

			view := e.View()
			for _, want := range tt.want {
				assert.Contains(t, view, want)
			}
		})
	}
}

func TestEditorViewRendersFormBranchesWithAndWithoutActiveModels(t *testing.T) {
	t.Run("form without active form still renders shell", func(t *testing.T) {
		e := NewEditorWithConfig(&config.Config{})
		e.step = StepForm

		view := e.View()

		assert.Contains(t, view, "Settings")
		assert.NotContains(t, view, "Configuration")
	})

	t.Run("form with active form renders form content", func(t *testing.T) {
		e := NewEditorWithConfig(&config.Config{})
		e.step = StepForm
		e.activeForm = NewAgentForm(e.state.Current)

		view := e.View()

		assert.Contains(t, view, "Settings")
		assert.Contains(t, view, "Agent Configuration")
		assert.Contains(t, view, "Provider")
	})

	t.Run("form with dependency panel renders panel before form", func(t *testing.T) {
		e := NewEditorWithConfig(&config.Config{})
		e.handleMenuSelection("smartaccount")
		require.Equal(t, StepForm, e.step)
		require.NotNil(t, e.depPanel)

		view := e.View()

		assert.Contains(t, view, "Prerequisites")
		assert.Contains(t, view, "Payment")
		assert.Contains(t, view, "Smart Account Configuration")
		assert.Less(t, strings.Index(view, "Payment"), strings.LastIndex(view, "Smart Account Configuration"))
	})

	t.Run("setup flow without model renders breadcrumb only", func(t *testing.T) {
		e := NewEditorWithConfig(&config.Config{})
		e.step = StepSetupFlow

		view := e.View()

		assert.Contains(t, view, "Settings")
		assert.Contains(t, view, "Setup")
		assert.NotContains(t, view, "Guided Setup")
	})
}

func TestEditorViewRendersEmbeddedSaveBannersOnlyOnMenu(t *testing.T) {
	t.Run("success banner is scoped to menu", func(t *testing.T) {
		e := NewEditorForEmbedding(config.DefaultConfig(), func(_ *config.Config, _ map[string]bool) error {
			return nil
		})
		e.handleMenuSelection("save")

		assert.Contains(t, e.View(), "Settings saved")

		e.step = StepForm
		e.activeForm = NewAgentForm(e.state.Current)

		assert.NotContains(t, e.View(), "Settings saved")
		assert.Contains(t, e.View(), "Agent Configuration")
	})

	t.Run("error banner is scoped to menu", func(t *testing.T) {
		e := NewEditorForEmbedding(config.DefaultConfig(), func(_ *config.Config, _ map[string]bool) error {
			return fmt.Errorf("write denied")
		})
		e.handleMenuSelection("save")

		assert.Contains(t, e.View(), "Save failed")
		assert.Contains(t, e.View(), "write denied")

		e.step = StepProvidersList
		e.providersList = NewProvidersListModel(e.state.Current)

		assert.NotContains(t, e.View(), "Save failed")
		assert.Contains(t, e.View(), "+ Add New Provider")
	})
}

func TestEditorUpdateFormPanelFocusTransitions(t *testing.T) {
	e := NewEditorWithConfig(&config.Config{})
	e.handleMenuSelection("smartaccount")
	require.Equal(t, StepForm, e.step)
	require.NotNil(t, e.depPanel)
	require.True(t, e.panelFocus)

	model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyTab})
	e = model.(*Editor)
	require.Nil(t, cmd)
	assert.Equal(t, StepForm, e.step)
	assert.False(t, e.panelFocus)

	model, cmd = e.Update(tea.KeyMsg{Type: tea.KeyTab})
	e = model.(*Editor)
	require.Nil(t, cmd)
	assert.Equal(t, StepForm, e.step)
	assert.True(t, e.panelFocus)

	model, cmd = e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	e = model.(*Editor)
	require.Nil(t, cmd)
	assert.Equal(t, StepForm, e.step)
	assert.False(t, e.panelFocus)
	assert.NotNil(t, e.activeForm)
}

func TestEditorUpdateListDeleteExitAndSelectionBranches(t *testing.T) {
	t.Run("provider existing selection opens edit form", func(t *testing.T) {
		e := NewEditorWithConfig(&config.Config{
			Providers: map[string]config.ProviderConfig{
				"anthropic": {Type: "anthropic"},
			},
		})
		e.step = StepProvidersList
		e.providersList = NewProvidersListModel(e.state.Current)

		model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyEnter})
		e = model.(*Editor)

		require.Nil(t, cmd)
		assert.Equal(t, StepForm, e.step)
		assert.Equal(t, "anthropic", e.activeProviderID)
		require.NotNil(t, e.activeForm)
		assert.Equal(t, "Edit Provider: anthropic", e.activeForm.Title)
	})

	t.Run("auth delete and exit update state", func(t *testing.T) {
		e := NewEditorWithConfig(&config.Config{
			Auth: config.AuthConfig{
				Providers: map[string]config.OIDCProviderConfig{
					"corp": {IssuerURL: "https://issuer.example"},
				},
			},
		})
		e.step = StepAuthProvidersList
		e.authProvidersList = NewAuthProvidersListModel(e.state.Current)

		model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		e = model.(*Editor)
		require.Nil(t, cmd)
		assert.NotContains(t, e.state.Current.Auth.Providers, "corp")
		assert.True(t, e.state.IsDirty("auth"))
		assert.Empty(t, e.authProvidersList.Deleted)

		model, cmd = e.Update(tea.KeyMsg{Type: tea.KeyEsc})
		e = model.(*Editor)
		require.Nil(t, cmd)
		assert.Equal(t, StepMenu, e.step)
		assert.False(t, e.authProvidersList.Exit)
	})

	t.Run("mcp delete and exit update state", func(t *testing.T) {
		e := NewEditorWithConfig(&config.Config{
			MCP: config.MCPConfig{
				Servers: map[string]config.MCPServerConfig{
					"filesystem": {Transport: "stdio"},
				},
			},
		})
		e.step = StepMCPServersList
		e.mcpServersList = NewMCPServersListModel(e.state.Current)

		model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		e = model.(*Editor)
		require.Nil(t, cmd)
		assert.NotContains(t, e.state.Current.MCP.Servers, "filesystem")
		assert.True(t, e.state.IsDirty("mcp"))
		assert.Empty(t, e.mcpServersList.Deleted)

		model, cmd = e.Update(tea.KeyMsg{Type: tea.KeyEsc})
		e = model.(*Editor)
		require.Nil(t, cmd)
		assert.Equal(t, StepMenu, e.step)
		assert.False(t, e.mcpServersList.Exit)
	})
}

func TestEditorUpdateSetupFlowForwardsFormMessages(t *testing.T) {
	e := NewEditorWithConfig(&config.Config{})
	e.step = StepSetupFlow
	e.setupFlow = NewSetupFlow("smartaccount", []DepResult{
		{
			Dependency: Dependency{CategoryID: "security", Label: "Security"},
			Status:     DepNotEnabled,
		},
	}, e.state)
	require.NotNil(t, e.setupFlow)
	require.NotNil(t, e.setupFlow.ActiveForm())

	before := e.setupFlow.ActiveForm().Cursor
	model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyDown})
	e = model.(*Editor)

	require.Nil(t, cmd)
	assert.Equal(t, StepSetupFlow, e.step)
	require.NotNil(t, e.setupFlow)
	require.NotNil(t, e.setupFlow.ActiveForm())
	assert.Greater(t, e.setupFlow.ActiveForm().Cursor, before)
}
