package settings

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestEditor_EscAtWelcome_Quits(t *testing.T) {
	e := NewEditor()
	require.Equal(t, StepWelcome, e.step)

	model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	ed := model.(*Editor)

	assert.Equal(t, StepWelcome, ed.step)
	assert.NotNil(t, cmd, "esc at welcome should return quit cmd")
}

func TestEditor_EscAtMenu_NavigatesToWelcome(t *testing.T) {
	e := NewEditor()
	e.step = StepMenu

	model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	ed := model.(*Editor)

	assert.Equal(t, StepWelcome, ed.step)
	assert.Nil(t, cmd, "esc at menu should not quit, just navigate back")
}

func TestEditor_EscAtMenuWhileSearching_StaysAtMenu(t *testing.T) {
	e := NewEditor()
	e.step = StepMenu

	// Enter search mode by pressing /
	model, _ := e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	ed := model.(*Editor)
	require.True(t, ed.menu.IsSearching(), "should be in search mode")

	// Press esc — should cancel search, not navigate back
	model, cmd := ed.Update(tea.KeyMsg{Type: tea.KeyEsc})
	ed = model.(*Editor)

	assert.Equal(t, StepMenu, ed.step, "should stay at menu")
	assert.False(t, ed.menu.IsSearching(), "search should be cancelled")
	assert.Nil(t, cmd)
}

func TestEditor_EscAtMenuLevel2_StaysAtMenu(t *testing.T) {
	e := NewEditor()
	e.step = StepMenu

	// Enter section (cursor at 0 = Core)
	model, _ := e.Update(tea.KeyMsg{Type: tea.KeyEnter})
	ed := model.(*Editor)
	require.Equal(t, StepMenu, ed.step, "should still be at menu")
	require.True(t, ed.menu.InCategoryLevel(), "should be at category level")

	// Press Esc — should go back to section level, NOT to Welcome
	model, cmd := ed.Update(tea.KeyMsg{Type: tea.KeyEsc})
	ed = model.(*Editor)

	assert.Equal(t, StepMenu, ed.step, "should stay at menu")
	assert.False(t, ed.menu.InCategoryLevel(), "should be back at section level")
	assert.Nil(t, cmd)
}

func TestMenu_EnterSection_TransitionsToLevel2(t *testing.T) {
	m := NewMenuModel()

	// Cursor at 0 = Core section
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.True(t, m.InCategoryLevel(), "should be at category level")
	assert.Equal(t, "Core", m.ActiveSectionTitle())
	assert.Equal(t, 0, m.Cursor, "cursor should reset to 0")
}

func TestMenu_EscAtLevel2_ReturnsToLevel1(t *testing.T) {
	m := NewMenuModel()

	// Navigate to section 2 (Automation)
	m.Cursor = 2
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.True(t, m.InCategoryLevel())
	require.Equal(t, "Automation", m.ActiveSectionTitle())

	// Press Esc
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	assert.False(t, m.InCategoryLevel(), "should be back at section level")
	assert.Equal(t, 2, m.Cursor, "cursor should restore to section position")
}

func TestMenu_TabOnlyAtLevel2(t *testing.T) {
	m := NewMenuModel()

	// Tab at Level 1: should be no-op (showAdvanced stays true)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.True(t, m.ShowAdvanced(), "tab at Level 1 should be no-op")
	assert.False(t, m.InCategoryLevel())

	// Enter section to go to Level 2
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.True(t, m.InCategoryLevel())

	// Tab at Level 2: should toggle
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.False(t, m.ShowAdvanced(), "tab at Level 2 should toggle to basic")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.True(t, m.ShowAdvanced(), "tab again should toggle back")
}

func TestMenu_SearchAtBothLevels(t *testing.T) {
	m := NewMenuModel()

	// Search from Level 1
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.True(t, m.IsSearching(), "should enter search from Level 1")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, m.IsSearching())

	// Enter Level 2
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.True(t, m.InCategoryLevel())

	// Search from Level 2
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.True(t, m.IsSearching(), "should enter search from Level 2")
}

func TestMenu_SaveCancelFromLevel1(t *testing.T) {
	m := NewMenuModel()

	// Navigate to Save & Exit (after 7 named sections = index 7)
	items := m.selectableItems()
	saveIdx := -1
	for i, item := range items {
		if item.ID == "save" {
			saveIdx = i
			break
		}
	}
	require.NotEqual(t, -1, saveIdx, "should find save item")

	m.Cursor = saveIdx
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "save", m.Selected, "should select save")

	// Reset and test cancel
	m = NewMenuModel()
	cancelIdx := -1
	items = m.selectableItems()
	for i, item := range items {
		if item.ID == "cancel" {
			cancelIdx = i
			break
		}
	}
	require.NotEqual(t, -1, cancelIdx, "should find cancel item")

	m.Cursor = cancelIdx
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "cancel", m.Selected, "should select cancel")
}

func TestMenu_AutomationIncludesRunLedger(t *testing.T) {
	m := NewMenuModel()

	found := false
	for _, section := range m.Sections {
		if section.Title != "Automation" {
			continue
		}
		for _, category := range section.Categories {
			if category.ID == "runledger" {
				found = true
				assert.Equal(t, "RunLedger", category.Title)
				break
			}
		}
	}

	assert.True(t, found, "Automation section should include RunLedger category")
}

func TestMenu_AutomationIncludesProvenance(t *testing.T) {
	m := NewMenuModel()

	found := false
	for _, section := range m.Sections {
		if section.Title != "Automation" {
			continue
		}
		for _, category := range section.Categories {
			if category.ID == "provenance" {
				found = true
				assert.Equal(t, "Provenance", category.Title)
				break
			}
		}
	}

	assert.True(t, found, "Automation section should include Provenance category")
}

func TestEditor_CtrlC_AlwaysQuits(t *testing.T) {
	tests := []struct {
		give string
		step EditorStep
	}{
		{give: "welcome", step: StepWelcome},
		{give: "menu", step: StepMenu},
		{give: "form", step: StepForm},
		{give: "providers_list", step: StepProvidersList},
		{give: "auth_providers_list", step: StepAuthProvidersList},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			e := NewEditor()
			e.step = tt.step

			model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
			ed := model.(*Editor)

			assert.True(t, ed.Cancelled, "ctrl+c should set Cancelled")
			assert.NotNil(t, cmd, "ctrl+c should return quit cmd")
		})
	}
}

func TestCategoryIsEnabled_ReflectsConfigState(t *testing.T) {
	t.Parallel()

	falseValue := false
	trueValue := true
	tests := []struct {
		name      string
		category  string
		configure func(*config.Config)
		want      bool
	}{
		{
			name:     "channels enabled when any channel is enabled",
			category: "channels",
			configure: func(cfg *config.Config) {
				cfg.Channels.Slack.Enabled = true
			},
			want: true,
		},
		{
			name:     "embedding requires provider",
			category: "embedding",
			configure: func(cfg *config.Config) {
				cfg.Embedding.Provider = "openai"
			},
			want: true,
		},
		{
			name:     "context profile off is disabled",
			category: "context_profile",
			configure: func(cfg *config.Config) {
				cfg.ContextProfile = "off"
			},
		},
		{
			name:     "context budget enabled by model window",
			category: "context_budget",
			configure: func(cfg *config.Config) {
				cfg.Context.ModelWindow = 128000
			},
			want: true,
		},
		{
			name:     "smart account child category follows parent",
			category: "smartaccount_paymaster",
			configure: func(cfg *config.Config) {
				cfg.SmartAccount.Enabled = true
			},
			want: true,
		},
		{
			name:     "p2p child category follows parent",
			category: "p2p_workspace",
			configure: func(cfg *config.Config) {
				cfg.P2P.Enabled = true
			},
			want: true,
		},
		{
			name:     "economy child category follows parent",
			category: "economy_escrow",
			configure: func(cfg *config.Config) {
				cfg.Economy.Enabled = true
			},
			want: true,
		},
		{
			name:     "alerting requires observability and alerting",
			category: "alerting",
			configure: func(cfg *config.Config) {
				cfg.Observability.Enabled = true
				cfg.Alerting.Enabled = true
			},
			want: true,
		},
		{
			name:     "alerting disabled without observability",
			category: "alerting",
			configure: func(cfg *config.Config) {
				cfg.Alerting.Enabled = true
			},
		},
		{
			name:     "gatekeeper nil defaults enabled",
			category: "gatekeeper",
			configure: func(cfg *config.Config) {
				cfg.Gatekeeper.Enabled = nil
			},
			want: true,
		},
		{
			name:     "gatekeeper explicit false disables",
			category: "gatekeeper",
			configure: func(cfg *config.Config) {
				cfg.Gatekeeper.Enabled = &falseValue
			},
		},
		{
			name:     "output manager explicit true enables",
			category: "output_manager",
			configure: func(cfg *config.Config) {
				cfg.Tools.OutputManager.Enabled = &trueValue
			},
			want: true,
		},
		{
			name:     "unknown category disabled",
			category: "does_not_exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{}
			if tt.configure != nil {
				tt.configure(cfg)
			}

			assert.Equal(t, tt.want, categoryIsEnabled(cfg, tt.category))
		})
	}
}

func TestEditor_HandleMenuSelectionOpensListAndFormSurfaces(t *testing.T) {
	e := NewEditorWithConfig(&config.Config{
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
				"filesystem": {Transport: "stdio"},
			},
		},
	})

	e.handleMenuSelection("providers")
	require.Equal(t, StepProvidersList, e.step)
	require.Len(t, e.providersList.Providers, 1)
	assert.Equal(t, "anthropic", e.providersList.Providers[0].ID)

	e.handleMenuSelection("auth")
	require.Equal(t, StepAuthProvidersList, e.step)
	require.Len(t, e.authProvidersList.Providers, 1)
	assert.Equal(t, "corp", e.authProvidersList.Providers[0].ID)

	e.handleMenuSelection("mcp_servers")
	require.Equal(t, StepMCPServersList, e.step)
	require.Len(t, e.mcpServersList.Servers, 1)
	assert.Equal(t, "filesystem", e.mcpServersList.Servers[0].Name)

	e.handleMenuSelection("agent")
	require.Equal(t, StepForm, e.step)
	require.NotNil(t, e.activeForm)
	assert.Equal(t, "Agent Configuration", e.activeForm.Title)
	assert.True(t, e.activeForm.Focus)
}

func TestEditor_ProvidersListDeleteSelectAndExit(t *testing.T) {
	cfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"anthropic": {Type: "anthropic"},
			"openai":    {Type: "openai"},
		},
	}
	e := NewEditorWithConfig(cfg)
	e.step = StepProvidersList
	e.providersList = NewProvidersListModel(e.state.Current)

	model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	e = model.(*Editor)
	require.Nil(t, cmd)
	assert.NotContains(t, e.state.Current.Providers, "anthropic")
	assert.True(t, e.state.IsDirty("providers"))
	assert.Empty(t, e.providersList.Deleted)

	e.providersList.Cursor = len(e.providersList.Providers)
	model, cmd = e.Update(tea.KeyMsg{Type: tea.KeyEnter})
	e = model.(*Editor)
	require.Nil(t, cmd)
	require.Equal(t, StepForm, e.step)
	require.NotNil(t, e.activeForm)
	assert.Empty(t, e.activeProviderID)
	assert.Contains(t, e.activeForm.Title, "Provider")

	model, cmd = e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	e = model.(*Editor)
	require.Nil(t, cmd)
	assert.Equal(t, StepProvidersList, e.step)
	assert.Nil(t, e.activeForm)

	model, cmd = e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	e = model.(*Editor)
	require.Nil(t, cmd)
	assert.Equal(t, StepMenu, e.step)
}

func TestEditor_AuthAndMCPListSelectionsOpenTypedForms(t *testing.T) {
	e := NewEditorWithConfig(&config.Config{
		Auth: config.AuthConfig{
			Providers: map[string]config.OIDCProviderConfig{
				"corp": {IssuerURL: "https://issuer.example"},
			},
		},
		MCP: config.MCPConfig{
			Servers: map[string]config.MCPServerConfig{
				"filesystem": {Transport: "stdio"},
			},
		},
	})

	e.step = StepAuthProvidersList
	e.authProvidersList = NewAuthProvidersListModel(e.state.Current)
	model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyEnter})
	e = model.(*Editor)
	require.Nil(t, cmd)
	require.Equal(t, StepForm, e.step)
	require.NotNil(t, e.activeForm)
	assert.Equal(t, "corp", e.activeAuthProviderID)
	assert.True(t, e.isAuthProviderForm())

	model, cmd = e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	e = model.(*Editor)
	require.Nil(t, cmd)
	require.Equal(t, StepAuthProvidersList, e.step)

	e.step = StepMCPServersList
	e.mcpServersList = NewMCPServersListModel(e.state.Current)
	model, cmd = e.Update(tea.KeyMsg{Type: tea.KeyEnter})
	e = model.(*Editor)
	require.Nil(t, cmd)
	require.Equal(t, StepForm, e.step)
	require.NotNil(t, e.activeForm)
	assert.Equal(t, "filesystem", e.activeMCPServerName)
	assert.True(t, e.isMCPServerForm())
}

func TestEditor_ViewRendersStepSpecificContent(t *testing.T) {
	e := NewEditor()

	assert.Contains(t, e.View(), "categories")

	e.step = StepProvidersList
	e.providersList = NewProvidersListModel(e.state.Current)
	assert.Contains(t, e.View(), "Providers")

	e.step = StepAuthProvidersList
	e.authProvidersList = NewAuthProvidersListModel(e.state.Current)
	assert.Contains(t, e.View(), "Auth Providers")

	e.step = StepMCPServersList
	e.mcpServersList = NewMCPServersListModel(e.state.Current)
	assert.Contains(t, e.View(), "MCP Servers")
}

func TestEditor_DependencyNavigationPushesAndPopsForms(t *testing.T) {
	e := NewEditorWithConfig(&config.Config{})

	e.handleMenuSelection("p2p")
	require.Equal(t, StepForm, e.step)
	require.NotNil(t, e.activeForm)
	require.NotNil(t, e.depPanel)
	require.True(t, e.panelFocus)
	assert.Equal(t, "p2p", e.depPanel.CategoryID)

	e.jumpToDependency("security")
	require.Len(t, e.navStack, 1)
	assert.Equal(t, "p2p", e.navStack[0])
	require.NotNil(t, e.activeForm)
	assert.True(t, e.activeForm.Focus)

	e.popNavStack()
	assert.Empty(t, e.navStack)
	require.NotNil(t, e.activeForm)
	require.NotNil(t, e.depPanel)
	assert.Equal(t, "p2p", e.depPanel.CategoryID)
	assert.True(t, e.activeForm.Focus)
}

func TestEditor_DependencyPanelKeyHandlingAndSetupFlow(t *testing.T) {
	e := NewEditorWithConfig(&config.Config{})
	e.handleMenuSelection("smartaccount")
	require.Equal(t, StepForm, e.step)
	require.NotNil(t, e.depPanel)
	require.True(t, e.panelFocus)

	model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyDown})
	e = model.(*Editor)
	require.Nil(t, cmd)
	assert.True(t, e.panelFocus)
	assert.Equal(t, "security", e.depPanel.SelectedCategoryID())

	model, cmd = e.Update(tea.KeyMsg{Type: tea.KeyEnter})
	e = model.(*Editor)
	require.Nil(t, cmd)
	assert.Equal(t, StepForm, e.step)
	assert.Equal(t, []string{"smartaccount"}, e.navStack)
	require.NotNil(t, e.activeForm)

	model, cmd = e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	e = model.(*Editor)
	require.Nil(t, cmd)
	assert.Equal(t, StepForm, e.step)
	assert.Empty(t, e.navStack)
	require.NotNil(t, e.depPanel)
	assert.Equal(t, "smartaccount", e.depPanel.CategoryID)

	model, cmd = e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	e = model.(*Editor)
	require.Nil(t, cmd)
	assert.Equal(t, StepSetupFlow, e.step)
	require.NotNil(t, e.setupFlow)
	assert.Equal(t, "smartaccount", e.setupFlow.TargetID())
	assert.Nil(t, e.depPanel)
	assert.False(t, e.panelFocus)
}

func TestEditor_CompleteSetupFlowRestoresTargetForm(t *testing.T) {
	e := NewEditorWithConfig(&config.Config{})
	e.handleMenuSelection("smartaccount")
	require.NotNil(t, e.depPanel)

	e.startSetupFlow()
	require.Equal(t, StepSetupFlow, e.step)
	require.NotNil(t, e.setupFlow)

	e.completeSetupFlow()
	assert.Equal(t, StepForm, e.step)
	assert.Nil(t, e.setupFlow)
	require.NotNil(t, e.activeForm)
	assert.True(t, e.activeForm.Focus)
	require.NotNil(t, e.depPanel)
	assert.Equal(t, "smartaccount", e.depPanel.CategoryID)
}
