package settings

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	StepComplete
)

// Editor is the main bubbletea model for the settings editor.
type Editor struct {
	step  EditorStep
	state *tuicore.ConfigState

	// Sub-models
	menu                 MenuModel
	providersList        ProvidersListModel
	authProvidersList    AuthProvidersListModel
	activeForm           *tuicore.FormModel
	activeProviderID     string
	activeAuthProviderID string

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
	return &Editor{
		step:  StepWelcome,
		state: tuicore.NewConfigState(),
		menu:  NewMenuModel(),
	}
}

// NewEditorWithConfig creates a new settings editor pre-loaded with the given config.
func NewEditorWithConfig(cfg *config.Config) *Editor {
	return &Editor{
		step:  StepWelcome,
		state: tuicore.NewConfigStateWith(cfg),
		menu:  NewMenuModel(),
	}
}

// Init implements tea.Model.
func (e *Editor) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (e *Editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
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
				return e, tea.Quit
			case StepProvidersList:
				e.step = StepMenu
				return e, nil
			case StepAuthProvidersList:
				e.step = StepMenu
				return e, nil
			case StepForm:
				if e.activeForm != nil {
					if e.activeAuthProviderID != "" || e.isAuthProviderForm() {
						e.state.UpdateAuthProviderFromForm(e.activeAuthProviderID, e.activeForm)
					} else if e.activeProviderID != "" || e.isProviderForm() {
						e.state.UpdateProviderFromForm(e.activeProviderID, e.activeForm)
					} else {
						e.state.UpdateConfigFromForm(e.activeForm)
					}
				}
				if e.activeAuthProviderID != "" || e.isAuthProviderForm() {
					e.step = StepAuthProvidersList
					e.authProvidersList = NewAuthProvidersListModel(e.state.Current)
				} else if e.activeProviderID != "" || e.isProviderForm() {
					e.step = StepProvidersList
					e.providersList = NewProvidersListModel(e.state.Current)
				} else {
					e.step = StepMenu
				}
				e.activeForm = nil
				e.activeProviderID = ""
				e.activeAuthProviderID = ""
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
		if e.activeForm != nil {
			var formCmd tea.Cmd
			*e.activeForm, formCmd = e.activeForm.Update(msg)
			cmd = formCmd
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
	}

	return e, cmd
}

func (e *Editor) handleMenuSelection(id string) tea.Cmd {
	switch id {
	case "agent":
		e.activeForm = NewAgentForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "server":
		e.activeForm = NewServerForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "channels":
		e.activeForm = NewChannelsForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "tools":
		e.activeForm = NewToolsForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "session":
		e.activeForm = NewSessionForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "security":
		e.activeForm = NewSecurityForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "knowledge":
		e.activeForm = NewKnowledgeForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "skill":
		e.activeForm = NewSkillForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "observational_memory":
		e.activeForm = NewObservationalMemoryForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "embedding":
		e.activeForm = NewEmbeddingForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "graph":
		e.activeForm = NewGraphForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "multi_agent":
		e.activeForm = NewMultiAgentForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "a2a":
		e.activeForm = NewA2AForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "payment":
		e.activeForm = NewPaymentForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "cron":
		e.activeForm = NewCronForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "background":
		e.activeForm = NewBackgroundForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "workflow":
		e.activeForm = NewWorkflowForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "librarian":
		e.activeForm = NewLibrarianForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "p2p":
		e.activeForm = NewP2PForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "p2p_zkp":
		e.activeForm = NewP2PZKPForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "p2p_pricing":
		e.activeForm = NewP2PPricingForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "p2p_owner":
		e.activeForm = NewP2POwnerProtectionForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "p2p_sandbox":
		e.activeForm = NewP2PSandboxForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "security_keyring":
		e.activeForm = NewKeyringForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "security_db":
		e.activeForm = NewDBEncryptionForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "security_kms":
		e.activeForm = NewKMSForm(e.state.Current)
		e.activeForm.Focus = true
		e.step = StepForm
	case "auth":
		e.authProvidersList = NewAuthProvidersListModel(e.state.Current)
		e.step = StepAuthProvidersList
	case "providers":
		e.providersList = NewProvidersListModel(e.state.Current)
		e.step = StepProvidersList
	case "save":
		e.Completed = true
		return tea.Quit
	case "cancel":
		e.err = fmt.Errorf("settings cancelled")
		e.Cancelled = true
		return tea.Quit
	}
	return nil
}

// View implements tea.Model.
func (e *Editor) View() string {
	var b strings.Builder

	// Dynamic breadcrumb header
	switch e.step {
	case StepWelcome, StepMenu:
		b.WriteString(tui.Breadcrumb("Settings"))
	case StepForm:
		formTitle := ""
		if e.activeForm != nil {
			formTitle = e.activeForm.Title
		}
		b.WriteString(tui.Breadcrumb("Settings", formTitle))
	case StepProvidersList:
		b.WriteString(tui.Breadcrumb("Settings", "Providers"))
	case StepAuthProvidersList:
		b.WriteString(tui.Breadcrumb("Settings", "Auth Providers"))
	default:
		b.WriteString(tui.Breadcrumb("Settings"))
	}
	b.WriteString("\n\n")

	// Content
	switch e.step {
	case StepWelcome:
		b.WriteString(e.viewWelcome())

	case StepMenu:
		b.WriteString(e.menu.View())

	case StepForm:
		if e.activeForm != nil {
			b.WriteString(e.activeForm.View())
		}

	case StepProvidersList:
		b.WriteString(e.providersList.View())

	case StepAuthProvidersList:
		b.WriteString(e.authProvidersList.View())
	}

	return b.String()
}

func (e *Editor) viewWelcome() string {
	var b strings.Builder

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.Primary).
		Padding(1, 3)

	title := lipgloss.NewStyle().Foreground(tui.Foreground).Bold(true).Render("Welcome to Lango Settings")
	desc := tui.MutedStyle.Render("Configure your agent, providers, channels, and more.\nAll settings are saved to an encrypted local profile.")

	b.WriteString(box.Render(title + "\n\n" + desc))
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
