package settings

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// SetupFlowState tracks the guided setup flow lifecycle.
type SetupFlowState int

const (
	SetupInProgress SetupFlowState = iota
	SetupCompleted
	SetupCancelled
)

// SetupStep represents a single step in the guided setup flow.
type SetupStep struct {
	CategoryID string
	Label      string
	Completed  bool
	Skipped    bool
}

// SetupFlow chains dependency forms before the target form.
type SetupFlow struct {
	steps       []SetupStep
	currentStep int
	state       SetupFlowState
	targetID    string

	// Active form for the current step
	activeForm *tuicore.FormModel
	// Config reference for form creation
	cfg *config.Config
	// ConfigState for saving form values
	configState *tuicore.ConfigState
}

// NewSetupFlow creates a guided setup flow from unmet transitive dependencies.
// Returns nil if fewer than 1 unmet dependency (no flow needed).
func NewSetupFlow(targetID string, unmetDeps []DepResult, configState *tuicore.ConfigState) *SetupFlow {
	if len(unmetDeps) == 0 {
		return nil
	}

	// Deduplicate dependencies (transitive resolution may produce duplicates)
	seen := make(map[string]bool)
	var steps []SetupStep
	for _, dep := range unmetDeps {
		if seen[dep.CategoryID] {
			continue
		}
		seen[dep.CategoryID] = true
		steps = append(steps, SetupStep{
			CategoryID: dep.CategoryID,
			Label:      dep.Label,
		})
	}

	if len(steps) == 0 {
		return nil
	}

	sf := &SetupFlow{
		steps:       steps,
		currentStep: 0,
		state:       SetupInProgress,
		targetID:    targetID,
		cfg:         configState.Current,
		configState: configState,
	}
	sf.enterCurrentStep()
	return sf
}

// enterCurrentStep creates the form for the current step.
func (sf *SetupFlow) enterCurrentStep() {
	if sf.currentStep >= len(sf.steps) {
		sf.state = SetupCompleted
		return
	}
	step := sf.steps[sf.currentStep]
	sf.activeForm = createFormForCategory(step.CategoryID, sf.cfg)
	if sf.activeForm != nil {
		sf.activeForm.Focus = true
	}
}

// NextStep saves the current form and advances to the next step.
func (sf *SetupFlow) NextStep() {
	sf.saveCurrentForm()
	sf.steps[sf.currentStep].Completed = true
	sf.currentStep++
	sf.enterCurrentStep()
}

// SkipStep marks the current step as skipped and advances.
func (sf *SetupFlow) SkipStep() {
	sf.steps[sf.currentStep].Skipped = true
	sf.currentStep++
	sf.enterCurrentStep()
}

// Cancel cancels the setup flow.
func (sf *SetupFlow) Cancel() {
	sf.saveCurrentForm()
	sf.state = SetupCancelled
}

// saveCurrentForm saves values from the active form to config state.
func (sf *SetupFlow) saveCurrentForm() {
	if sf.activeForm != nil && sf.configState != nil {
		sf.configState.UpdateConfigFromForm(sf.activeForm)
	}
}

// State returns the current flow state.
func (sf *SetupFlow) State() SetupFlowState {
	return sf.state
}

// ActiveForm returns the form model for the current step.
func (sf *SetupFlow) ActiveForm() *tuicore.FormModel {
	return sf.activeForm
}

// TargetID returns the target category that triggered this flow.
func (sf *SetupFlow) TargetID() string {
	return sf.targetID
}

// View renders the setup flow UI.
func (sf *SetupFlow) View() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.Primary)
	b.WriteString(titleStyle.Render("Guided Setup"))
	b.WriteString("\n\n")

	// Progress bar
	total := len(sf.steps)
	completed := 0
	for _, s := range sf.steps {
		if s.Completed || s.Skipped {
			completed++
		}
	}
	b.WriteString(sf.renderProgressBar(completed, total))
	b.WriteString("\n")

	// Step list
	b.WriteString(sf.renderStepList())
	b.WriteString("\n")

	// Current form
	if sf.activeForm != nil {
		b.WriteString(sf.activeForm.View())
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(tui.HelpBar(
		tui.HelpEntry("Ctrl+N", "Save & Next"),
		tui.HelpEntry("Ctrl+S", "Skip step"),
		tui.HelpEntry("Esc", "Cancel setup"),
	))

	return b.String()
}

func (sf *SetupFlow) renderProgressBar(completed, total int) string {
	barWidth := 30
	filled := 0
	if total > 0 {
		filled = completed * barWidth / total
	}

	filledStyle := lipgloss.NewStyle().Foreground(tui.Success)
	mutedStyle := lipgloss.NewStyle().Foreground(tui.Muted)

	bar := filledStyle.Render(strings.Repeat("━", filled)) +
		mutedStyle.Render(strings.Repeat("━", barWidth-filled))

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.Primary)
	header := fmt.Sprintf("[%d/%d]", completed, total)

	return headerStyle.Render(header) + " " + bar
}

func (sf *SetupFlow) renderStepList() string {
	var b strings.Builder

	for i, step := range sf.steps {
		var indicator string
		var style lipgloss.Style

		switch {
		case step.Completed:
			indicator = tui.CheckPass
			style = lipgloss.NewStyle().Foreground(tui.Success)
		case step.Skipped:
			indicator = "○"
			style = lipgloss.NewStyle().Foreground(tui.Dim).Strikethrough(true)
		case i == sf.currentStep:
			indicator = "▸"
			style = lipgloss.NewStyle().Foreground(tui.Highlight).Bold(true)
		default:
			indicator = "○"
			style = lipgloss.NewStyle().Foreground(tui.Muted)
		}

		b.WriteString("  ")
		b.WriteString(style.Render(fmt.Sprintf("%s %s", indicator, step.Label)))
		b.WriteString("\n")
	}

	return b.String()
}

// createFormForCategory maps a category ID to its form constructor.
func createFormForCategory(categoryID string, cfg *config.Config) *tuicore.FormModel {
	switch categoryID {
	case "agent":
		return NewAgentForm(cfg)
	case "channels":
		return NewChannelsForm(cfg)
	case "tools":
		return NewToolsForm(cfg)
	case "server":
		return NewServerForm(cfg)
	case "session":
		return NewSessionForm(cfg)
	case "logging":
		return NewLoggingForm(cfg)
	case "gatekeeper":
		return NewGatekeeperForm(cfg)
	case "output_manager":
		return NewOutputManagerForm(cfg)
	case "security":
		return NewSecurityForm(cfg)
	case "knowledge":
		return NewKnowledgeForm(cfg)
	case "skill":
		return NewSkillForm(cfg)
	case "observational_memory":
		return NewObservationalMemoryForm(cfg)
	case "embedding":
		return NewEmbeddingForm(cfg)
	case "graph":
		return NewGraphForm(cfg)
	case "multi_agent":
		return NewMultiAgentForm(cfg)
	case "a2a":
		return NewA2AForm(cfg)
	case "payment":
		return NewPaymentForm(cfg)
	case "cron":
		return NewCronForm(cfg)
	case "background":
		return NewBackgroundForm(cfg)
	case "workflow":
		return NewWorkflowForm(cfg)
	case "runledger":
		return NewRunLedgerForm(cfg)
	case "smartaccount":
		return NewSmartAccountForm(cfg)
	case "smartaccount_session":
		return NewSmartAccountSessionForm(cfg)
	case "smartaccount_paymaster":
		return NewSmartAccountPaymasterForm(cfg)
	case "smartaccount_modules":
		return NewSmartAccountModulesForm(cfg)
	case "mcp":
		return NewMCPForm(cfg)
	case "hooks":
		return NewHooksForm(cfg)
	case "agent_memory":
		return NewAgentMemoryForm(cfg)
	case "librarian":
		return NewLibrarianForm(cfg)
	case "economy":
		return NewEconomyForm(cfg)
	case "economy_risk":
		return NewEconomyRiskForm(cfg)
	case "economy_negotiation":
		return NewEconomyNegotiationForm(cfg)
	case "economy_escrow":
		return NewEconomyEscrowForm(cfg)
	case "economy_escrow_onchain":
		return NewEconomyEscrowOnChainForm(cfg)
	case "economy_pricing":
		return NewEconomyPricingForm(cfg)
	case "observability":
		return NewObservabilityForm(cfg)
	case "p2p":
		return NewP2PForm(cfg)
	case "p2p_zkp":
		return NewP2PZKPForm(cfg)
	case "p2p_pricing":
		return NewP2PPricingForm(cfg)
	case "p2p_owner":
		return NewP2POwnerProtectionForm(cfg)
	case "p2p_sandbox":
		return NewP2PSandboxForm(cfg)
	case "p2p_workspace":
		return NewP2PWorkspaceForm(cfg)
	case "security_db":
		return NewDBEncryptionForm(cfg)
	case "security_kms":
		return NewKMSForm(cfg)
	default:
		return nil
	}
}
