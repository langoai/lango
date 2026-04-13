package settings

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// EditorStep represents the current step in the settings editor.
type EditorStep int

const (
	StepWelcome EditorStep = iota
	StepMenu
	StepForm
	StepProvidersList
	StepAuthProvidersList
	StepMCPServersList
	StepSetupFlow
	StepComplete
)

// OnSaveFunc is the callback type for embedded-mode save.
// It receives the current config and a map of dirty (modified) field keys.
type OnSaveFunc func(cfg *config.Config, dirtyKeys map[string]bool) error

// Editor is the main bubbletea model for the settings editor.
type Editor struct {
	step  EditorStep
	state *tuicore.ConfigState

	// Sub-models
	menu                 MenuModel
	providersList        ProvidersListModel
	authProvidersList    AuthProvidersListModel
	mcpServersList       MCPServersListModel
	activeForm           *tuicore.FormModel
	activeProviderID     string
	activeAuthProviderID string
	activeMCPServerName  string

	// Dependency discovery
	depIndex   *DependencyIndex
	depPanel   *DependencyPanel
	panelFocus bool     // true when the dependency panel has focus (vs. the form)
	navStack   []string // navigation stack for jump-to-dependency flow

	// Guided setup flow
	setupFlow *SetupFlow

	// Embedded mode
	OnSave      OnSaveFunc // if set, save calls this instead of tea.Quit
	saveSuccess bool       // true after a successful embedded save

	// UI State
	width  int
	height int
	err    error

	// Public status
	Completed bool
	Cancelled bool
}

// NewEditor creates a new settings editor with default config.
func NewEditor() *Editor {
	e := &Editor{
		step:     StepWelcome,
		state:    tuicore.NewConfigState(),
		menu:     NewMenuModel(),
		depIndex: NewDependencyIndex(),
	}
	e.wireMenuCheckers()
	return e
}

// NewEditorWithConfig creates a new settings editor pre-loaded with the given config.
func NewEditorWithConfig(cfg *config.Config) *Editor {
	e := &Editor{
		step:     StepWelcome,
		state:    tuicore.NewConfigStateWith(cfg),
		menu:     NewMenuModel(),
		depIndex: NewDependencyIndex(),
	}
	e.wireMenuCheckers()
	return e
}

// NewEditorForEmbedding creates a settings editor suitable for embedding in another
// TUI (e.g. cockpit). It skips the welcome step and uses the provided callback
// for save instead of calling tea.Quit.
func NewEditorForEmbedding(cfg *config.Config, onSave OnSaveFunc) *Editor {
	e := NewEditorWithConfig(cfg.Clone()) // deep copy to avoid mutating live config
	e.OnSave = onSave
	e.step = StepMenu // skip StepWelcome
	return e
}

// wireMenuCheckers connects the dirty/enabled/dependency checkers to the menu.
func (e *Editor) wireMenuCheckers() {
	e.menu.DirtyChecker = func(id string) bool {
		return e.state.IsDirty(id)
	}
	e.menu.EnabledChecker = func(id string) bool {
		return categoryIsEnabled(e.state.Current, id)
	}
	e.menu.DependencyChecker = func(id string) int {
		if e.depIndex == nil {
			return 0
		}
		return e.depIndex.UnmetRequired(id, e.state.Current)
	}
}

// categoryIsEnabled returns true if the feature associated with the category is enabled.
func categoryIsEnabled(cfg *config.Config, id string) bool {
	switch id {
	case "channels":
		return cfg.Channels.Telegram.Enabled || cfg.Channels.Discord.Enabled || cfg.Channels.Slack.Enabled
	case "knowledge":
		return cfg.Knowledge.Enabled
	case "skill":
		return cfg.Skill.Enabled
	case "observational_memory":
		return cfg.ObservationalMemory.Enabled
	case "embedding":
		return cfg.Embedding.Provider != ""
	case "graph":
		return cfg.Graph.Enabled
	case "librarian":
		return cfg.Librarian.Enabled
	case "agent_memory":
		return cfg.AgentMemory.Enabled
	case "multi_agent":
		return cfg.Agent.MultiAgent
	case "a2a":
		return cfg.A2A.Enabled
	case "hooks":
		return cfg.Hooks.Enabled
	case "context_profile":
		return cfg.ContextProfile != "" && cfg.ContextProfile != "off"
	case "retrieval":
		return cfg.Retrieval.Enabled
	case "auto_adjust":
		return cfg.Retrieval.AutoAdjust.Enabled
	case "context_budget":
		return cfg.Context.ModelWindow > 0 || cfg.Context.Allocation.Knowledge > 0
	case "cron":
		return cfg.Cron.Enabled
	case "background":
		return cfg.Background.Enabled
	case "workflow":
		return cfg.Workflow.Enabled
	case "payment":
		return cfg.Payment.Enabled
	case "smartaccount", "smartaccount_session", "smartaccount_paymaster", "smartaccount_modules":
		return cfg.SmartAccount.Enabled
	case "p2p", "p2p_workspace", "p2p_zkp", "p2p_pricing", "p2p_owner", "p2p_sandbox":
		return cfg.P2P.Enabled
	case "economy", "economy_risk", "economy_negotiation", "economy_escrow", "economy_escrow_onchain", "economy_pricing":
		return cfg.Economy.Enabled
	case "mcp", "mcp_servers":
		return cfg.MCP.Enabled
	case "observability":
		return cfg.Observability.Enabled
	case "security":
		return cfg.Security.Interceptor.Enabled
	case "ontology":
		return cfg.Ontology.Enabled
	case "alerting":
		return cfg.Observability.Enabled && cfg.Alerting.Enabled
	case "gatekeeper":
		return derefBoolCfg(cfg.Gatekeeper.Enabled, true)
	case "output_manager":
		return derefBoolCfg(cfg.Tools.OutputManager.Enabled, true)
	case "server":
		return cfg.Server.HTTPEnabled
	default:
		return false
	}
}

// derefBoolCfg safely dereferences a *bool with a default.
func derefBoolCfg(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

// Init implements tea.Model.
func (e *Editor) Init() tea.Cmd {
	return tea.ClearScreen
}

// Update implements tea.Model.
func (e *Editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear inline save banner on next key press.
		e.saveSuccess = false
		e.err = nil

		if msg.String() == "ctrl+c" {
			e.Cancelled = true
			return e, tea.Quit
		}

		if msg.String() == "esc" {
			switch e.step {
			case StepWelcome:
				return e, tea.Quit
			case StepMenu:
				if e.menu.IsSearching() {
					// Let the menu handle esc to cancel search
					break
				}
				if e.menu.InCategoryLevel() {
					// Let menu.Update() handle Level2 → Level1 transition
					break
				}
				e.step = StepWelcome
				return e, nil
			case StepProvidersList:
				e.step = StepMenu
				return e, nil
			case StepAuthProvidersList:
				e.step = StepMenu
				return e, nil
			case StepMCPServersList:
				e.step = StepMenu
				return e, nil
			case StepSetupFlow:
				if e.setupFlow != nil {
					e.setupFlow.Cancel()
				}
				e.setupFlow = nil
				e.step = StepMenu
				return e, nil
			case StepForm:
				// If panel has focus, Esc switches focus back to the form
				if e.panelFocus && e.depPanel != nil {
					e.panelFocus = false
					return e, nil
				}
				// If a search-select dropdown is open, let the form handle Esc
				// (closes dropdown only, does not exit the form).
				if e.activeForm != nil && e.activeForm.HasOpenDropdown() {
					break
				}
				if e.activeForm != nil {
					if e.activeAuthProviderID != "" || e.isAuthProviderForm() {
						e.state.UpdateAuthProviderFromForm(e.activeAuthProviderID, e.activeForm)
					} else if e.activeProviderID != "" || e.isProviderForm() {
						e.state.UpdateProviderFromForm(e.activeProviderID, e.activeForm)
					} else if e.activeMCPServerName != "" || e.isMCPServerForm() {
						e.state.UpdateMCPServerFromForm(e.activeMCPServerName, e.activeForm)
					} else {
						e.state.UpdateConfigFromForm(e.activeForm)
					}
				}
				// Pop nav stack if we jumped to a dependency
				if len(e.navStack) > 0 {
					e.popNavStack()
					return e, nil
				}
				if e.activeAuthProviderID != "" || e.isAuthProviderForm() {
					e.step = StepAuthProvidersList
					e.authProvidersList = NewAuthProvidersListModel(e.state.Current)
				} else if e.activeProviderID != "" || e.isProviderForm() {
					e.step = StepProvidersList
					e.providersList = NewProvidersListModel(e.state.Current)
				} else if e.activeMCPServerName != "" || e.isMCPServerForm() {
					e.step = StepMCPServersList
					e.mcpServersList = NewMCPServersListModel(e.state.Current)
				} else {
					e.step = StepMenu
				}
				e.activeForm = nil
				e.activeProviderID = ""
				e.activeAuthProviderID = ""
				e.activeMCPServerName = ""
				e.depPanel = nil
				e.panelFocus = false
				return e, nil
			}
		}

	case tea.WindowSizeMsg:
		e.width = msg.Width
		e.height = msg.Height
	}

	switch e.step {
	case StepWelcome:
		if msg, ok := msg.(tea.KeyMsg); ok {
			if msg.String() == "enter" {
				e.step = StepMenu
			}
		}

	case StepMenu:
		var menuCmd tea.Cmd
		e.menu, menuCmd = e.menu.Update(msg)
		cmd = menuCmd

		if e.menu.Selected != "" {
			cmd = e.handleMenuSelection(e.menu.Selected)
			e.menu.Selected = ""
		}

	case StepForm:
		// Handle panel focus key events
		if e.panelFocus && e.depPanel != nil {
			if msg, ok := msg.(tea.KeyMsg); ok {
				switch msg.String() {
				case "up", "k":
					e.depPanel.MoveUp()
					return e, nil
				case "down", "j":
					e.depPanel.MoveDown()
					return e, nil
				case "enter":
					if e.depPanel.SelectedIsUnmet() {
						e.jumpToDependency(e.depPanel.SelectedCategoryID())
					}
					return e, nil
				case "s":
					// Start guided setup flow
					if e.depPanel.UnmetCount() > 0 {
						e.startSetupFlow()
					}
					return e, nil
				case "tab":
					// Switch focus to the form
					e.panelFocus = false
					return e, nil
				}
			}
		}

		// Handle tab key to switch focus to panel when panel exists
		if e.depPanel != nil && !e.panelFocus {
			if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "tab" {
				e.panelFocus = true
				return e, nil
			}
		}

		if e.activeForm != nil {
			var formCmd tea.Cmd
			*e.activeForm, formCmd = e.activeForm.Update(msg)
			cmd = formCmd
		}

	case StepSetupFlow:
		if e.setupFlow != nil {
			if msg, ok := msg.(tea.KeyMsg); ok {
				switch msg.String() {
				case "ctrl+n":
					e.setupFlow.NextStep()
					if e.setupFlow.State() == SetupCompleted {
						e.completeSetupFlow()
					}
					return e, nil
				case "ctrl+s":
					e.setupFlow.SkipStep()
					if e.setupFlow.State() == SetupCompleted {
						e.completeSetupFlow()
					}
					return e, nil
				}
			}
			// Forward other messages to the setup flow's active form
			if sf := e.setupFlow.ActiveForm(); sf != nil {
				var formCmd tea.Cmd
				*sf, formCmd = sf.Update(msg)
				cmd = formCmd
			}
		}

	case StepProvidersList:
		var plCmd tea.Cmd
		e.providersList, plCmd = e.providersList.Update(msg)
		cmd = plCmd

		if e.providersList.Deleted != "" {
			delete(e.state.Current.Providers, e.providersList.Deleted)
			e.state.MarkDirty("providers")
			e.providersList = NewProvidersListModel(e.state.Current)
		} else if e.providersList.Exit {
			e.providersList.Exit = false
			e.step = StepMenu
		} else if e.providersList.Selected != "" {
			id := e.providersList.Selected
			if id == "NEW" {
				e.activeProviderID = ""
				e.activeForm = NewProviderForm("", config.ProviderConfig{})
			} else {
				e.activeProviderID = id
				if p, ok := e.state.Current.Providers[id]; ok {
					e.activeForm = NewProviderForm(id, p)
				}
			}
			e.activeForm.Focus = true
			e.step = StepForm
			e.providersList.Selected = ""
		}

	case StepAuthProvidersList:
		var aplCmd tea.Cmd
		e.authProvidersList, aplCmd = e.authProvidersList.Update(msg)
		cmd = aplCmd

		if e.authProvidersList.Deleted != "" {
			delete(e.state.Current.Auth.Providers, e.authProvidersList.Deleted)
			e.state.MarkDirty("auth")
			e.authProvidersList = NewAuthProvidersListModel(e.state.Current)
		} else if e.authProvidersList.Exit {
			e.authProvidersList.Exit = false
			e.step = StepMenu
		} else if e.authProvidersList.Selected != "" {
			id := e.authProvidersList.Selected
			if id == "NEW" {
				e.activeAuthProviderID = ""
				e.activeForm = NewOIDCProviderForm("", config.OIDCProviderConfig{})
			} else {
				e.activeAuthProviderID = id
				if p, ok := e.state.Current.Auth.Providers[id]; ok {
					e.activeForm = NewOIDCProviderForm(id, p)
				}
			}
			e.activeForm.Focus = true
			e.step = StepForm
			e.authProvidersList.Selected = ""
		}

	case StepMCPServersList:
		var mslCmd tea.Cmd
		e.mcpServersList, mslCmd = e.mcpServersList.Update(msg)
		cmd = mslCmd

		if e.mcpServersList.Deleted != "" {
			delete(e.state.Current.MCP.Servers, e.mcpServersList.Deleted)
			e.state.MarkDirty("mcp")
			e.mcpServersList = NewMCPServersListModel(e.state.Current)
		} else if e.mcpServersList.Exit {
			e.mcpServersList.Exit = false
			e.step = StepMenu
		} else if e.mcpServersList.Selected != "" {
			name := e.mcpServersList.Selected
			if name == "NEW" {
				e.activeMCPServerName = ""
				e.activeForm = NewMCPServerForm("", config.MCPServerConfig{})
			} else {
				e.activeMCPServerName = name
				if srv, ok := e.state.Current.MCP.Servers[name]; ok {
					e.activeForm = NewMCPServerForm(name, srv)
				}
			}
			e.activeForm.Focus = true
			e.step = StepForm
			e.mcpServersList.Selected = ""
		}
	}

	return e, cmd
}

func (e *Editor) handleMenuSelection(id string) tea.Cmd {
	// Handle special (non-form) selections first.
	switch id {
	case "mcp_servers":
		e.mcpServersList = NewMCPServersListModel(e.state.Current)
		e.step = StepMCPServersList
		return nil
	case "auth":
		e.authProvidersList = NewAuthProvidersListModel(e.state.Current)
		e.step = StepAuthProvidersList
		return nil
	case "providers":
		e.providersList = NewProvidersListModel(e.state.Current)
		e.step = StepProvidersList
		return nil
	case "save":
		if e.OnSave != nil {
			cfg := e.Config()
			explicitKeys := make(map[string]bool, len(config.ContextRelatedKeys()))
			for _, k := range config.ContextRelatedKeys() {
				explicitKeys[k] = true
			}
			if err := e.OnSave(cfg, explicitKeys); err != nil {
				e.err = err
				return nil
			}
			e.saveSuccess = true
			return nil
		}
		e.Completed = true
		return tea.Quit
	case "cancel":
		e.err = fmt.Errorf("settings cancelled")
		e.Cancelled = true
		return tea.Quit
	}

	// Try to create a form for this category.
	if form := createFormForCategory(id, e.state.Current); form != nil {
		e.activeForm = form
		e.activeForm.Focus = true
		e.step = StepForm
		e.attachDependencyPanel(id)
	}
	return nil
}

// View implements tea.Model.
func (e *Editor) View() string {
	var b strings.Builder

	// Dynamic breadcrumb header
	switch e.step {
	case StepWelcome, StepMenu:
		if e.menu.InCategoryLevel() {
			b.WriteString(tui.Breadcrumb("Settings", e.menu.ActiveSectionTitle()))
		} else {
			b.WriteString(tui.Breadcrumb("Settings"))
		}
	case StepForm:
		segments := []string{"Settings"}
		// Show navigation chain if jumped from another form
		segments = append(segments, e.navStack...)
		formTitle := ""
		if e.activeForm != nil {
			formTitle = e.activeForm.Title
		}
		segments = append(segments, formTitle)
		b.WriteString(tui.Breadcrumb(segments...))
	case StepSetupFlow:
		targetLabel := ""
		if e.setupFlow != nil {
			targetLabel = e.setupFlow.TargetID()
		}
		b.WriteString(tui.Breadcrumb("Settings", "Setup", targetLabel))
	case StepProvidersList:
		b.WriteString(tui.Breadcrumb("Settings", "Providers"))
	case StepAuthProvidersList:
		b.WriteString(tui.Breadcrumb("Settings", "Auth Providers"))
	case StepMCPServersList:
		b.WriteString(tui.Breadcrumb("Settings", "MCP Servers"))
	default:
		b.WriteString(tui.Breadcrumb("Settings"))
	}
	b.WriteString("\n\n")

	// Content
	switch e.step {
	case StepWelcome:
		b.WriteString(e.viewWelcome())

	case StepMenu:
		if e.err != nil {
			b.WriteString(tui.FormatFail("Save failed: "+e.err.Error()) + "\n\n")
		} else if e.saveSuccess {
			b.WriteString(tui.FormatPass("Settings saved") + "\n\n")
		}
		b.WriteString(e.menu.View())

	case StepForm:
		// Render dependency panel above form if present
		if e.depPanel != nil {
			b.WriteString(e.depPanel.View())
			b.WriteString("\n")
		}
		if e.activeForm != nil {
			b.WriteString(e.activeForm.View())
		}

	case StepSetupFlow:
		if e.setupFlow != nil {
			b.WriteString(e.setupFlow.View())
		}

	case StepProvidersList:
		b.WriteString(e.providersList.View())

	case StepAuthProvidersList:
		b.WriteString(e.authProvidersList.View())

	case StepMCPServersList:
		b.WriteString(e.mcpServersList.View())
	}

	return b.String()
}

func (e *Editor) viewWelcome() string {
	var b strings.Builder

	b.WriteString(tui.BannerBox())
	b.WriteString("\n\n")
	b.WriteString(tui.MutedStyle.Render("Configure your agent, providers, channels, and more."))
	b.WriteString("\n")
	b.WriteString(tui.MutedStyle.Render("All settings are saved to an encrypted local profile."))
	b.WriteString("\n\n")

	// Category summary
	all := e.menu.allCategories()
	basic, adv := 0, 0
	for _, c := range all {
		if c.Tier == TierBasic {
			basic++
		} else {
			adv++
		}
	}
	total := basic + adv
	summary := fmt.Sprintf("%d categories (%d basic, %d advanced)", total, basic, adv)
	b.WriteString(tui.MutedStyle.Render(summary))
	b.WriteString("\n\n")

	// Tips
	b.WriteString(tui.MutedStyle.Render("Tips:"))
	b.WriteString("\n")
	b.WriteString(tui.MutedStyle.Render("  / Search across all categories"))
	b.WriteString("\n")
	b.WriteString(tui.MutedStyle.Render("  @basic @advanced @enabled @modified @ready — smart filters"))
	b.WriteString("\n")
	b.WriteString(tui.MutedStyle.Render("  Select a section, then browse its settings"))
	b.WriteString("\n\n")

	b.WriteString(tui.HelpBar(
		tui.HelpEntry("Enter", "Start"),
		tui.HelpEntry("Esc", "Quit"),
	))

	return b.String()
}

// Config returns the current configuration from the editor state.
func (e *Editor) Config() *config.Config {
	return e.state.Current
}

func (e *Editor) isProviderForm() bool {
	if e.activeForm == nil {
		return false
	}
	return strings.Contains(e.activeForm.Title, "Provider") && !strings.Contains(e.activeForm.Title, "OIDC")
}

func (e *Editor) isAuthProviderForm() bool {
	if e.activeForm == nil {
		return false
	}
	return strings.Contains(e.activeForm.Title, "OIDC")
}

func (e *Editor) isMCPServerForm() bool {
	if e.activeForm == nil {
		return false
	}
	return strings.Contains(e.activeForm.Title, "MCP Server")
}

// attachDependencyPanel creates a dependency panel for the given category ID.
func (e *Editor) attachDependencyPanel(categoryID string) {
	e.depPanel = nil
	e.panelFocus = false
	if e.depIndex == nil {
		return
	}
	results := e.depIndex.Evaluate(categoryID, e.state.Current)
	panel := NewDependencyPanel(categoryID, results)
	if panel != nil {
		e.depPanel = panel
		// Auto-focus panel if there are unmet required deps
		if panel.UnmetCount() > 0 {
			e.panelFocus = true
		}
	}
}

// jumpToDependency pushes the current form onto the nav stack and opens the dependency's form.
func (e *Editor) jumpToDependency(targetCategoryID string) {
	// Save current form
	if e.activeForm != nil {
		e.state.UpdateConfigFromForm(e.activeForm)
	}

	// Push current category onto nav stack
	if e.depPanel != nil {
		e.navStack = append(e.navStack, e.depPanel.CategoryID)
	}

	// Open the dependency's form
	e.activeForm = createFormForCategory(targetCategoryID, e.state.Current)
	if e.activeForm != nil {
		e.activeForm.Focus = true
	}

	// Attach dependency panel for the new form
	e.attachDependencyPanel(targetCategoryID)
}

// popNavStack returns to the previous form in the navigation stack.
func (e *Editor) popNavStack() {
	if len(e.navStack) == 0 {
		return
	}

	// Pop the last category
	prevID := e.navStack[len(e.navStack)-1]
	e.navStack = e.navStack[:len(e.navStack)-1]

	// Open the previous category's form
	e.activeForm = createFormForCategory(prevID, e.state.Current)
	if e.activeForm != nil {
		e.activeForm.Focus = true
	}

	// Re-evaluate dependency panel for the restored form
	e.attachDependencyPanel(prevID)
}

// startSetupFlow creates and enters a guided setup flow.
func (e *Editor) startSetupFlow() {
	if e.depPanel == nil || e.depIndex == nil {
		return
	}

	categoryID := e.depPanel.CategoryID

	// Save current form
	if e.activeForm != nil {
		e.state.UpdateConfigFromForm(e.activeForm)
	}

	// Collect transitive unmet dependencies
	unmetDeps := e.depIndex.AllTransitiveUnmet(categoryID, e.state.Current)
	sf := NewSetupFlow(categoryID, unmetDeps, e.state)
	if sf == nil {
		return
	}

	e.setupFlow = sf
	e.step = StepSetupFlow
	e.depPanel = nil
	e.panelFocus = false
}

// completeSetupFlow finishes the guided setup and opens the target form.
func (e *Editor) completeSetupFlow() {
	if e.setupFlow == nil {
		return
	}

	targetID := e.setupFlow.TargetID()
	e.setupFlow = nil

	// Open the original target form
	e.activeForm = createFormForCategory(targetID, e.state.Current)
	if e.activeForm != nil {
		e.activeForm.Focus = true
	}
	e.step = StepForm
	e.attachDependencyPanel(targetID)
}
