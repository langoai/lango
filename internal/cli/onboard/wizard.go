package onboard

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/langowarny/lango/internal/cli/tui"
	"github.com/langowarny/lango/internal/config"
)

// WizardStep represents the current step in the wizard.
type WizardStep int

const (
	StepWelcome WizardStep = iota
	StepMenu
	StepForm
	StepProvidersList
	StepComplete
)

// Wizard is the main bubbletea model for the onboard wizard.
type Wizard struct {
	step  WizardStep
	state *ConfigState

	// Sub-models
	menu             MenuModel
	providersList    ProvidersListModel
	activeForm       *FormModel
	activeProviderID string // Track which provider is being edited

	// UI State
	width  int
	height int
	err    error

	// Public status
	Completed bool
	Cancelled bool
}

// NewWizard creates a new onboard wizard with default config.
func NewWizard() *Wizard {
	return &Wizard{
		step:  StepWelcome,
		state: NewConfigState(),
		menu:  NewMenuModel(),
	}
}

// NewWizardWithConfig creates a new onboard wizard pre-loaded with the given config.
func NewWizardWithConfig(cfg *config.Config) *Wizard {
	return &Wizard{
		step:  StepWelcome,
		state: NewConfigStateWith(cfg),
		menu:  NewMenuModel(),
	}
}

// Init implements tea.Model.
func (w *Wizard) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (w *Wizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			w.Cancelled = true
			return w, tea.Quit
		}

		// ESC handling depends on context
		if msg.String() == "esc" {
			switch w.step {
			case StepWelcome:
				return w, tea.Quit
			case StepMenu:
				// Maybe verify exit? For now quit.
				// Or go back to welcome?
				// Actually, "Cancel" in menu handles quit.
				// Let's make ESC go back to welcome or quit.
				return w, tea.Quit
			case StepProvidersList:
				w.step = StepMenu
				return w, nil
			case StepForm:
				// Save form state to config
				if w.activeForm != nil {
					if w.activeProviderID != "" || w.isProviderForm() {
						w.state.UpdateProviderFromForm(w.activeProviderID, w.activeForm)
					} else {
						w.state.UpdateConfigFromForm(w.activeForm)
					}
				}
				// Go back to menu or providers list
				if w.activeProviderID != "" || w.isProviderForm() {
					w.step = StepProvidersList
					// Refresh list
					w.providersList = NewProvidersListModel(w.state.Current)
				} else {
					w.step = StepMenu
				}
				w.activeForm = nil
				w.activeProviderID = ""
				return w, nil
			}
		}

	case tea.WindowSizeMsg:
		w.width = msg.Width
		w.height = msg.Height
	}

	// Route events based on step
	switch w.step {
	case StepWelcome:
		// Simple start/quit logic
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "enter":
				// Default to Advanced (Editor) mode as that's what we are building
				w.step = StepMenu
			}
		}

	case StepMenu:
		var menuCmd tea.Cmd
		w.menu, menuCmd = w.menu.Update(msg)
		cmd = menuCmd

		if w.menu.Selected != "" {
			cmd = w.handleMenuSelection(w.menu.Selected)
			w.menu.Selected = "" // Reset selection signal
		}

	case StepForm:
		if w.activeForm != nil {
			var formCmd tea.Cmd
			*w.activeForm, formCmd = w.activeForm.Update(msg)
			cmd = formCmd
		}

	case StepProvidersList:
		var plCmd tea.Cmd
		w.providersList, plCmd = w.providersList.Update(msg)
		cmd = plCmd

		if w.providersList.Exit {
			w.providersList.Exit = false
			w.step = StepMenu
		} else if w.providersList.Selected != "" {
			id := w.providersList.Selected
			if id == "NEW" {
				w.activeProviderID = "" // New provider
				w.activeForm = NewProviderForm("", config.ProviderConfig{})
			} else {
				w.activeProviderID = id
				if p, ok := w.state.Current.Providers[id]; ok {
					w.activeForm = NewProviderForm(id, p)
				}
			}
			w.activeForm.Focus = true
			w.step = StepForm
			w.providersList.Selected = ""
		}
	}

	return w, cmd
}

func (w *Wizard) handleMenuSelection(id string) tea.Cmd {
	switch id {
	case "agent":
		w.activeForm = NewAgentForm(w.state.Current)
		w.activeForm.Focus = true
		w.step = StepForm
	case "server":
		w.activeForm = NewServerForm(w.state.Current)
		w.activeForm.Focus = true
		w.step = StepForm
	case "channels":
		w.activeForm = NewChannelsForm(w.state.Current)
		w.activeForm.Focus = true
		w.step = StepForm
	case "tools":
		w.activeForm = NewToolsForm(w.state.Current)
		w.activeForm.Focus = true
		w.step = StepForm
	case "security":
		w.activeForm = NewSecurityForm(w.state.Current)
		w.activeForm.Focus = true
		w.step = StepForm
	case "knowledge":
		w.activeForm = NewKnowledgeForm(w.state.Current)
		w.activeForm.Focus = true
		w.step = StepForm
	case "observational_memory":
		w.activeForm = NewObservationalMemoryForm(w.state.Current)
		w.activeForm.Focus = true
		w.step = StepForm
	case "embedding":
		w.activeForm = NewEmbeddingForm(w.state.Current)
		w.activeForm.Focus = true
		w.step = StepForm
	case "providers":
		w.providersList = NewProvidersListModel(w.state.Current)
		w.step = StepProvidersList
	case "save":
		// Save is handled by the caller (onboard.go) after wizard completes.
		w.Completed = true // Signal completion to main loop
		return tea.Quit
	case "cancel":
		w.err = fmt.Errorf("setup cancelled")
		w.Cancelled = true
		return tea.Quit
	}
	return nil
}

// View implements tea.Model.
func (w *Wizard) View() string {
	var b strings.Builder

	// Global Header
	b.WriteString(tui.TitleStyle.Render("🚀 Lango Configuration Editor"))
	b.WriteString("\n\n")

	switch w.step {
	case StepWelcome:
		b.WriteString(tui.SubtitleStyle.Render("Welcome!"))
		b.WriteString("\n\n")
		b.WriteString("Press [Enter] to start configuring Lango.\n")
		b.WriteString(tui.MutedStyle.Render("This editor allows you to configure Agent, Server, Tools, and more."))

	case StepMenu:
		b.WriteString(w.menu.View())

	case StepForm:
		if w.activeForm != nil {
			b.WriteString(w.activeForm.View())
		}

	case StepProvidersList:
		b.WriteString(w.providersList.View())
	}

	// Global Footer (if needed)

	return b.String()
}

// Config returns the current configuration from the wizard state.
func (w *Wizard) Config() *config.Config {
	return w.state.Current
}

func (w *Wizard) isProviderForm() bool {
	if w.activeForm == nil {
		return false
	}
	// Heuristic: check if title contains "Provider"
	return strings.Contains(w.activeForm.Title, "Provider")
}
