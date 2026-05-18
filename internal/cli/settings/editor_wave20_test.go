package settings

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

func TestWave20EditorInitWindowSizeAndWelcomeTransition(t *testing.T) {
	e := NewEditor()

	assert.NotNil(t, e.Init())

	model, cmd := e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	e = model.(*Editor)
	require.Nil(t, cmd)
	assert.Equal(t, 120, e.width)
	assert.Equal(t, 40, e.height)

	model, cmd = e.Update(tea.KeyMsg{Type: tea.KeyEnter})
	e = model.(*Editor)
	require.Nil(t, cmd)
	assert.Equal(t, StepMenu, e.step)
}

func TestWave20EditorMenuCheckersReflectCurrentState(t *testing.T) {
	e := NewEditorWithConfig(&config.Config{})
	e.state.MarkDirty("agent")
	e.state.Current.Knowledge.Enabled = true

	assert.True(t, e.menu.DirtyChecker("agent"))
	assert.True(t, e.menu.EnabledChecker("knowledge"))
	assert.Greater(t, e.menu.DependencyChecker("smartaccount"), 0)

	e.depIndex = nil
	assert.Zero(t, e.menu.DependencyChecker("smartaccount"))
}

func TestWave20CategoryIsEnabledCoversDeterministicBranches(t *testing.T) {
	trueValue := true
	tests := []struct {
		name      string
		category  string
		configure func(*config.Config)
	}{
		{
			name:     "skill",
			category: "skill",
			configure: func(cfg *config.Config) {
				cfg.Skill.Enabled = true
			},
		},
		{
			name:     "observational_memory",
			category: "observational_memory",
			configure: func(cfg *config.Config) {
				cfg.ObservationalMemory.Enabled = true
			},
		},
		{
			name:     "graph",
			category: "graph",
			configure: func(cfg *config.Config) {
				cfg.Graph.Enabled = true
			},
		},
		{
			name:     "agent_memory",
			category: "agent_memory",
			configure: func(cfg *config.Config) {
				cfg.AgentMemory.Enabled = true
			},
		},
		{
			name:     "multi_agent",
			category: "multi_agent",
			configure: func(cfg *config.Config) {
				cfg.Agent.MultiAgent = true
			},
		},
		{
			name:     "a2a",
			category: "a2a",
			configure: func(cfg *config.Config) {
				cfg.A2A.Enabled = true
			},
		},
		{
			name:     "hooks",
			category: "hooks",
			configure: func(cfg *config.Config) {
				cfg.Hooks.Enabled = true
			},
		},
		{
			name:     "retrieval",
			category: "retrieval",
			configure: func(cfg *config.Config) {
				cfg.Retrieval.Enabled = true
			},
		},
		{
			name:     "auto_adjust",
			category: "auto_adjust",
			configure: func(cfg *config.Config) {
				cfg.Retrieval.AutoAdjust.Enabled = true
			},
		},
		{
			name:     "cron",
			category: "cron",
			configure: func(cfg *config.Config) {
				cfg.Cron.Enabled = true
			},
		},
		{
			name:     "background",
			category: "background",
			configure: func(cfg *config.Config) {
				cfg.Background.Enabled = true
			},
		},
		{
			name:     "workflow",
			category: "workflow",
			configure: func(cfg *config.Config) {
				cfg.Workflow.Enabled = true
			},
		},
		{
			name:     "payment",
			category: "payment",
			configure: func(cfg *config.Config) {
				cfg.Payment.Enabled = true
			},
		},
		{
			name:     "mcp",
			category: "mcp",
			configure: func(cfg *config.Config) {
				cfg.MCP.Enabled = true
			},
		},
		{
			name:     "observability",
			category: "observability",
			configure: func(cfg *config.Config) {
				cfg.Observability.Enabled = true
			},
		},
		{
			name:     "security",
			category: "security",
			configure: func(cfg *config.Config) {
				cfg.Security.Interceptor.Enabled = true
			},
		},
		{
			name:     "ontology",
			category: "ontology",
			configure: func(cfg *config.Config) {
				cfg.Ontology.Enabled = true
			},
		},
		{
			name:     "output_manager_nil_defaults_enabled",
			category: "output_manager",
			configure: func(cfg *config.Config) {
				cfg.Tools.OutputManager.Enabled = nil
			},
		},
		{
			name:     "gatekeeper_explicit_true",
			category: "gatekeeper",
			configure: func(cfg *config.Config) {
				cfg.Gatekeeper.Enabled = &trueValue
			},
		},
		{
			name:     "server",
			category: "server",
			configure: func(cfg *config.Config) {
				cfg.Server.HTTPEnabled = true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			tt.configure(cfg)

			assert.True(t, categoryIsEnabled(cfg, tt.category))
		})
	}
}

func TestWave20EditorEscPersistsNewProviderFormAndReturnsToList(t *testing.T) {
	e := NewEditorWithConfig(&config.Config{})
	e.step = StepForm
	e.activeForm = NewProviderForm("", config.ProviderConfig{})
	e.activeForm.Focus = true
	setWave20FieldValue(e.activeForm, "id", "local-openai")
	setWave20FieldValue(e.activeForm, "type", "openai")
	setWave20FieldValue(e.activeForm, "apikey", "${OPENAI_API_KEY}")
	setWave20FieldValue(e.activeForm, "baseurl", "https://api.example.test/v1")

	model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	e = model.(*Editor)

	require.Nil(t, cmd)
	assert.Equal(t, StepProvidersList, e.step)
	assert.Nil(t, e.activeForm)
	assert.Empty(t, e.activeProviderID)
	require.Contains(t, e.state.Current.Providers, "local-openai")
	provider := e.state.Current.Providers["local-openai"]
	assert.Equal(t, "openai", string(provider.Type))
	assert.Equal(t, "${OPENAI_API_KEY}", provider.APIKey)
	assert.Equal(t, "https://api.example.test/v1", provider.BaseURL)
	assert.True(t, e.state.IsDirty("providers"))
}

func TestWave20EditorEscPersistsNewAuthProviderFormAndReturnsToList(t *testing.T) {
	e := NewEditorWithConfig(&config.Config{})
	e.step = StepForm
	e.activeForm = NewOIDCProviderForm("", config.OIDCProviderConfig{})
	e.activeForm.Focus = true
	setWave20FieldValue(e.activeForm, "oidc_id", "corp")
	setWave20FieldValue(e.activeForm, "oidc_issuer", "https://issuer.example.test")
	setWave20FieldValue(e.activeForm, "oidc_client_id", "client-id")
	setWave20FieldValue(e.activeForm, "oidc_redirect", "http://localhost/callback")
	setWave20FieldValue(e.activeForm, "oidc_scopes", "openid,email")

	model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	e = model.(*Editor)

	require.Nil(t, cmd)
	assert.Equal(t, StepAuthProvidersList, e.step)
	assert.Nil(t, e.activeForm)
	assert.Empty(t, e.activeAuthProviderID)
	require.Contains(t, e.state.Current.Auth.Providers, "corp")
	provider := e.state.Current.Auth.Providers["corp"]
	assert.Equal(t, "https://issuer.example.test", provider.IssuerURL)
	assert.Equal(t, "client-id", provider.ClientID)
	assert.Equal(t, "http://localhost/callback", provider.RedirectURL)
	assert.Equal(t, []string{"openid", "email"}, provider.Scopes)
	assert.True(t, e.state.IsDirty("auth"))
}

func TestWave20EditorEscPersistsNewMCPServerFormAndReturnsToList(t *testing.T) {
	e := NewEditorWithConfig(&config.Config{})
	e.step = StepForm
	e.activeForm = NewMCPServerForm("", config.MCPServerConfig{})
	e.activeForm.Focus = true
	setWave20FieldValue(e.activeForm, "mcp_srv_name", "filesystem")
	setWave20FieldValue(e.activeForm, "mcp_srv_transport", "stdio")
	setWave20FieldValue(e.activeForm, "mcp_srv_command", "npx")
	setWave20FieldValue(e.activeForm, "mcp_srv_args", "-y,@modelcontextprotocol/server-filesystem")
	setWave20FieldValue(e.activeForm, "mcp_srv_timeout", "15s")
	setWave20FieldChecked(e.activeForm, "mcp_srv_enabled", true)

	model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	e = model.(*Editor)

	require.Nil(t, cmd)
	assert.Equal(t, StepMCPServersList, e.step)
	assert.Nil(t, e.activeForm)
	assert.Empty(t, e.activeMCPServerName)
	require.Contains(t, e.state.Current.MCP.Servers, "filesystem")
	server := e.state.Current.MCP.Servers["filesystem"]
	assert.Equal(t, "stdio", server.Transport)
	assert.Equal(t, "npx", server.Command)
	assert.Equal(t, []string{"-y", "@modelcontextprotocol/server-filesystem"}, server.Args)
	assert.Equal(t, 15*time.Second, server.Timeout)
	require.NotNil(t, server.Enabled)
	assert.True(t, *server.Enabled)
	assert.True(t, e.state.IsDirty("mcp"))
}

func TestWave20EditorEscClosesOpenDropdownWithoutLeavingForm(t *testing.T) {
	form := tuicore.NewFormModel("Generic Configuration")
	form.Focus = true
	form.AddField(&tuicore.Field{
		Key:             "model",
		Label:           "Model",
		Type:            tuicore.InputSearchSelect,
		Value:           "gpt-4o",
		Options:         []string{"gpt-4o", "gpt-4.1"},
		FilteredOptions: []string{"gpt-4o", "gpt-4.1"},
		SelectOpen:      true,
	})

	e := NewEditorWithConfig(&config.Config{})
	e.step = StepForm
	e.activeForm = &form

	model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	e = model.(*Editor)

	require.Nil(t, cmd)
	assert.Equal(t, StepForm, e.step)
	require.NotNil(t, e.activeForm)
	assert.False(t, e.activeForm.Fields[0].SelectOpen)
}

func TestWave20EditorHelperBranchesAreNoOpsWhenStateIsMissing(t *testing.T) {
	e := NewEditorWithConfig(&config.Config{})

	assert.False(t, e.isProviderForm())
	assert.False(t, e.isAuthProviderForm())
	assert.False(t, e.isMCPServerForm())

	e.depIndex = nil
	e.attachDependencyPanel("smartaccount")
	assert.Nil(t, e.depPanel)
	assert.False(t, e.panelFocus)

	e.popNavStack()
	assert.Empty(t, e.navStack)

	e.startSetupFlow()
	assert.Equal(t, StepWelcome, e.step)
	assert.Nil(t, e.setupFlow)

	e.completeSetupFlow()
	assert.Equal(t, StepWelcome, e.step)
	assert.Nil(t, e.activeForm)
}

func TestWave20EditorSetupFlowKeyBranchesCompleteAndSkip(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyType
	}{
		{name: "next", key: tea.KeyCtrlN},
		{name: "skip", key: tea.KeyCtrlS},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEditorWithConfig(&config.Config{})
			e.step = StepSetupFlow
			e.setupFlow = NewSetupFlow("p2p", []DepResult{
				{
					Dependency: Dependency{
						CategoryID: "security",
						Label:      "Security",
						Required:   true,
					},
					Status: DepNotEnabled,
				},
			}, e.state)
			require.NotNil(t, e.setupFlow)

			model, cmd := e.Update(tea.KeyMsg{Type: tt.key})
			e = model.(*Editor)

			require.Nil(t, cmd)
			assert.Equal(t, StepForm, e.step)
			assert.Nil(t, e.setupFlow)
			require.NotNil(t, e.activeForm)
			assert.Equal(t, "P2P Network Configuration", e.activeForm.Title)
		})
	}
}

func TestWave20EditorSetupFlowEscCancelsAndReturnsToMenu(t *testing.T) {
	e := NewEditorWithConfig(&config.Config{})
	e.step = StepSetupFlow
	e.setupFlow = NewSetupFlow("p2p", []DepResult{
		{
			Dependency: Dependency{
				CategoryID: "security",
				Label:      "Security",
				Required:   true,
			},
			Status: DepNotEnabled,
		},
	}, e.state)
	require.NotNil(t, e.setupFlow)

	model, cmd := e.Update(tea.KeyMsg{Type: tea.KeyEsc})
	e = model.(*Editor)

	require.Nil(t, cmd)
	assert.Equal(t, StepMenu, e.step)
	assert.Nil(t, e.setupFlow)
}

func TestWave20EditorCancelSelectionSetsErrorAndQuitCommand(t *testing.T) {
	e := NewEditorWithConfig(&config.Config{})

	cmd := e.handleMenuSelection("cancel")

	assert.True(t, e.Cancelled)
	require.Error(t, e.err)
	assert.Equal(t, "settings cancelled", e.err.Error())
	assert.NotNil(t, cmd)
}

func setWave20FieldValue(form *tuicore.FormModel, key string, value string) {
	for _, field := range form.Fields {
		if field.Key != key {
			continue
		}
		field.Value = value
		if field.TextInput.Width > 0 {
			field.TextInput.SetValue(value)
		}
		return
	}
}

func setWave20FieldChecked(form *tuicore.FormModel, key string, checked bool) {
	for _, field := range form.Fields {
		if field.Key != key {
			continue
		}
		field.Checked = checked
		return
	}
}
