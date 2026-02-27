package onboard

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/types"
)

// WizardStep represents the current step in the wizard.
type WizardStep int

const (
	StepProvider WizardStep = iota // Step 1
	StepAgent                      // Step 2
	StepChannel                    // Step 3
	StepSecurity                   // Step 4
	StepTest                       // Step 5
	StepComplete
)

// channelOption defines a channel choice in the selector.
type channelOption struct {
	ID   string
	Name string
	Desc string
}

var channelOptions = []channelOption{
	{ID: string(types.ChannelTelegram), Name: "Telegram", Desc: "Bot via BotFather"},
	{ID: string(types.ChannelDiscord), Name: "Discord", Desc: "Bot via Developer Portal"},
	{ID: string(types.ChannelSlack), Name: "Slack", Desc: "App via Socket Mode"},
	{ID: "skip", Name: "Skip", Desc: "Configure later in settings"},
}

// Wizard is the main bubbletea model for the 5-step onboard wizard.
type Wizard struct {
	step  WizardStep
	state *tuicore.ConfigState

	// Current step's form
	activeForm *tuicore.FormModel

	// Channel selection state (Step 3)
	channelChoice     string // "telegram", "discord", "slack", "skip", or ""
	channelSelectMode bool   // true when showing channel picker (before form)
	channelCursor     int    // cursor for channel selector

	// Test results (Step 5)
	testResults []TestResult

	// UI dimensions
	width  int
	height int

	// Public status
	Completed bool
	Cancelled bool
}

// NewWizard creates a new 5-step onboard wizard.
func NewWizard(cfg *config.Config) *Wizard {
	w := &Wizard{
		step:              StepProvider,
		state:             tuicore.NewConfigStateWith(cfg),
		channelSelectMode: false,
	}
	w.enterStep(StepProvider)
	return w
}

// Init implements tea.Model.
func (w *Wizard) Init() tea.Cmd {
	return tea.ClearScreen
}

// Update implements tea.Model.
func (w *Wizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			w.Cancelled = true
			return w, tea.Quit
		}

		// Global navigation
		switch msg.String() {
		case "ctrl+n":
			return w.nextStep()
		case "ctrl+p":
			return w.prevStep()
		}

		// Step-specific handling
		switch w.step {
		case StepProvider, StepAgent, StepSecurity:
			return w.handleFormStep(msg)
		case StepChannel:
			return w.handleChannelStep(msg)
		case StepTest:
			return w.handleTestStep(msg)
		}

	case tea.WindowSizeMsg:
		w.width = msg.Width
		w.height = msg.Height
	}

	return w, nil
}

func (w *Wizard) handleFormStep(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		if w.step == StepProvider {
			// On first step, Esc quits
			w.Cancelled = true
			return w, tea.Quit
		}
		return w.prevStep()
	}

	if w.activeForm != nil {
		var cmd tea.Cmd
		*w.activeForm, cmd = w.activeForm.Update(msg)
		return w, cmd
	}
	return w, nil
}

func (w *Wizard) handleChannelStep(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		if w.channelSelectMode {
			// Go back to previous step
			return w.prevStep()
		}
		// Go back to channel selector
		w.channelSelectMode = true
		w.activeForm = nil
		return w, nil
	}

	if w.channelSelectMode {
		switch msg.String() {
		case "up", "k":
			if w.channelCursor > 0 {
				w.channelCursor--
			}
		case "down", "j":
			if w.channelCursor < len(channelOptions)-1 {
				w.channelCursor++
			}
		case "enter":
			choice := channelOptions[w.channelCursor]
			w.channelChoice = choice.ID
			if choice.ID == "skip" {
				return w.nextStep()
			}
			// Enable the selected channel
			w.enableChannel(choice.ID)
			w.activeForm = NewChannelStepForm(choice.ID, w.state.Current)
			if w.activeForm != nil {
				w.activeForm.Focus = true
			}
			w.channelSelectMode = false
		}
		return w, nil
	}

	// Channel form mode
	if w.activeForm != nil {
		var cmd tea.Cmd
		*w.activeForm, cmd = w.activeForm.Update(msg)
		return w, cmd
	}
	return w, nil
}

func (w *Wizard) handleTestStep(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		w.Completed = true
		return w, tea.Quit
	case "esc":
		return w.prevStep()
	}
	return w, nil
}

func (w *Wizard) nextStep() (tea.Model, tea.Cmd) {
	// Save current form before advancing
	w.saveCurrentForm()

	next := w.step + 1
	if next > StepTest {
		w.Completed = true
		return w, tea.Quit
	}

	w.enterStep(next)
	return w, nil
}

func (w *Wizard) prevStep() (tea.Model, tea.Cmd) {
	// Save current form before going back
	w.saveCurrentForm()

	if w.step <= StepProvider {
		return w, nil
	}

	prev := w.step - 1
	w.enterStep(prev)
	return w, nil
}

func (w *Wizard) enterStep(step WizardStep) {
	w.step = step

	switch step {
	case StepProvider:
		w.activeForm = NewProviderStepForm(w.state.Current)
		w.activeForm.Focus = true
	case StepAgent:
		w.activeForm = NewAgentStepForm(w.state.Current)
		w.activeForm.Focus = true
	case StepChannel:
		w.channelSelectMode = true
		w.activeForm = nil
	case StepSecurity:
		w.activeForm = NewSecurityStepForm(w.state.Current)
		w.activeForm.Focus = true
	case StepTest:
		w.activeForm = nil
		w.testResults = RunConfigTests(w.state.Current)
	}
}

func (w *Wizard) saveCurrentForm() {
	if w.activeForm == nil {
		return
	}

	switch w.step {
	case StepProvider:
		w.state.UpdateProviderFromForm("", w.activeForm)
		// Also set agent provider to match
		for _, f := range w.activeForm.Fields {
			if f.Key == "id" {
				w.state.Current.Agent.Provider = f.Value
			}
		}
	case StepAgent:
		w.state.UpdateConfigFromForm(w.activeForm)
	case StepChannel:
		w.state.UpdateConfigFromForm(w.activeForm)
	case StepSecurity:
		w.state.UpdateConfigFromForm(w.activeForm)
	}
}

func (w *Wizard) enableChannel(ch string) {
	switch types.ChannelType(ch) {
	case types.ChannelTelegram:
		w.state.Current.Channels.Telegram.Enabled = true
	case types.ChannelDiscord:
		w.state.Current.Channels.Discord.Enabled = true
	case types.ChannelSlack:
		w.state.Current.Channels.Slack.Enabled = true
	}
}

// View implements tea.Model.
func (w *Wizard) View() string {
	var b strings.Builder

	// Banner + subtitle
	b.WriteString(tui.Banner())
	b.WriteString("\n")
	b.WriteString(tui.SubtitleStyle.Render("Setup Wizard"))
	b.WriteString("\n\n")

	if w.step <= StepTest {
		// Progress bar
		b.WriteString(renderProgress(int(w.step), w.width))
		// Step list
		b.WriteString(renderStepList(int(w.step)))
	}

	// Step content
	switch w.step {
	case StepProvider, StepAgent, StepSecurity:
		if w.activeForm != nil {
			b.WriteString(w.activeForm.View())
		}
	case StepChannel:
		b.WriteString(w.viewChannel())
	case StepTest:
		b.WriteString(w.viewTest())
	}

	// Footer
	b.WriteString("\n")
	footer := tui.MutedStyle.Render("ctrl+n: next \u2022 ctrl+p: back \u2022 ctrl+c: quit")
	b.WriteString(footer)

	return b.String()
}

func (w *Wizard) viewChannel() string {
	var b strings.Builder

	if w.channelSelectMode {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.Primary).MarginBottom(1)
		b.WriteString(titleStyle.Render("Select a Channel"))
		b.WriteString("\n\n")

		for i, opt := range channelOptions {
			cursor := "  "
			style := lipgloss.NewStyle()
			descStyle := tui.MutedStyle

			if i == w.channelCursor {
				cursor = "\u25b8 "
				style = lipgloss.NewStyle().Foreground(tui.Success).Bold(true)
			}

			b.WriteString(cursor)
			b.WriteString(style.Render(opt.Name))
			b.WriteString("    ")
			b.WriteString(descStyle.Render(opt.Desc))
			b.WriteString("\n")
		}
	} else if w.activeForm != nil {
		b.WriteString(w.activeForm.View())
	}

	return b.String()
}

func (w *Wizard) viewTest() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.Primary).MarginBottom(1)
	b.WriteString(titleStyle.Render("Configuration Test Results"))
	b.WriteString("\n\n")

	for _, r := range w.testResults {
		var line string
		switch r.Status {
		case "pass":
			line = tui.FormatPass(fmt.Sprintf("%s: %s", r.Name, r.Message))
		case "warn":
			line = tui.FormatWarn(fmt.Sprintf("%s: %s", r.Name, r.Message))
		case "fail":
			line = tui.FormatFail(fmt.Sprintf("%s: %s", r.Name, r.Message))
		}
		b.WriteString("  ")
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(tui.MutedStyle.Render("Press [Enter] to save and complete, [Esc] to go back"))

	return b.String()
}

// Config returns the current configuration.
func (w *Wizard) Config() *config.Config {
	return w.state.Current
}
