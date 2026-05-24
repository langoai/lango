package pages

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/langoai/lango/internal/cli/cockpit"
	"github.com/langoai/lango/internal/cli/workbenchstart"
	"github.com/langoai/lango/internal/config"
)

func (p *MissionControlPage) setupHint() string {
	if p == nil || p.surface != missionControlSurfaceWorkbench || !p.needsInitialSetup() {
		return ""
	}
	return "Run `lango onboard` for guided setup, `lango settings` for the full editor, or `lango doctor` to inspect the current profile."
}

func (p *MissionControlPage) starterHint() string {
	if p == nil || p.surface != missionControlSurfaceWorkbench || p.needsInitialSetup() || p.hasWorkbenchStarterPromptArmed() || p.hasWorkbenchTurnInFlight() {
		return ""
	}
	prompts := p.workbenchStarterPromptSet()
	defaultPrompt := p.defaultWorkbenchSeedPrompt()
	if strings.TrimSpace(defaultPrompt) == "" {
		defaultPrompt = prompts[0]
	}
	if p.hasWorkbenchTurnNeedingAttention() {
		return "Recovery step: press `Enter` for `" + defaultPrompt + "`, or use `1` for `" + prompts[0] + "`, `2` for `" + prompts[1] + "`, and `3` for `" + prompts[2] + "`."
	}
	if p.hasWorkbenchCompletedTurn() {
		return "Next step: press `Enter` for `" + defaultPrompt + "`, or use `1` for `" + prompts[0] + "`, `2` for `" + prompts[1] + "`, and `3` for `" + prompts[2] + "`."
	}
	return "Quick start: press `Enter` for `" + defaultPrompt + "`, or use `1` for `" + prompts[0] + "`, `2` for `" + prompts[1] + "`, and `3` for `" + prompts[2] + "`."
}

func (p *MissionControlPage) seededStarterHint() string {
	if p.hasWorkbenchTurnInFlight() {
		return ""
	}
	if !p.hasWorkbenchStarterPromptArmed() {
		return ""
	}
	return "Starter ready: press `Enter` to run it, use `1-3` to replace it, or edit the prompt before sending."
}

func (p *MissionControlPage) runningStarterHint() string {
	if !p.hasWorkbenchTurnInFlight() {
		return ""
	}
	if p.hasWorkbenchQueuedFollowUpDraft() {
		return "Follow-up ready: press `Enter` to interrupt and run it, use `1-3` to replace it, or keep editing before sending."
	}
	return "Current request running: type the next prompt now, or press `Enter` to interrupt and run it."
}

func (p *MissionControlPage) composerEmptyStateHint() string {
	if p == nil || p.surface != missionControlSurfaceWorkbench || !p.isEmpty() {
		return ""
	}
	if p.needsInitialSetup() {
		return "Setup first: `lango onboard`, `lango settings`, or `lango doctor`."
	}
	if p.hasWorkbenchTurnInFlight() {
		if p.hasWorkbenchQueuedFollowUpDraft() {
			return ""
		}
		return "Current request running... Type the next prompt, or press Enter to interrupt."
	}
	prompts := p.workbenchStarterPromptSet()
	defaultPrompt := p.defaultWorkbenchSeedPrompt()
	if strings.TrimSpace(defaultPrompt) == "" {
		defaultPrompt = prompts[0]
	}
	if p.hasWorkbenchTurnNeedingAttention() {
		return "Recovery step: press `Enter` for `" + defaultPrompt + "`, or use 1-3 to load: `" + prompts[0] + "`, `" + prompts[1] + "`, or `" + prompts[2] + "`."
	}
	if p.hasWorkbenchCompletedTurn() {
		return "Next step: press `Enter` for `" + defaultPrompt + "`, or use 1-3 to load: `" + prompts[0] + "`, `" + prompts[1] + "`, or `" + prompts[2] + "`."
	}
	return "Press `Enter` for `" + defaultPrompt + "`, or use 1-3 to load: `" + prompts[0] + "`, `" + prompts[1] + "`, or `" + prompts[2] + "`."
}

func (p *MissionControlPage) needsInitialSetup() bool {
	if p == nil {
		return false
	}
	return !config.EvaluateAgentSetup(p.cfg).Ready()
}

func (p *MissionControlPage) dashboardHint() string {
	if p.surface != missionControlSurfaceWorkbench {
		return ""
	}
	return "For the advanced multi-page dashboard, use `lango cockpit`."
}

func (p *MissionControlPage) footerSurfaceHint() string {
	if p.focus == missionControlFocusMissions {
		if mission := p.selectedMission(); mission != nil && mission.Kind == cockpit.MissionKindProposed {
			if p.surface == missionControlSurfaceWorkbench {
				return "Enter accepts proposal  d dismisses proposal  Type request here  `lango chat` focused chat  `lango cockpit` dashboard"
			}
			return "Enter accepts proposal  d dismisses proposal  Type request here  `lango chat` focused chat"
		}
	}
	if p.surface == missionControlSurfaceWorkbench {
		if p.hasWorkbenchTurnInFlight() {
			if p.hasWorkbenchQueuedFollowUpDraft() {
				return "Enter runs staged follow-up  1-3 replace follow-up  Type to edit it  `lango chat` focused chat  `lango cockpit` dashboard"
			}
			return "Request running  Type next prompt  Enter interrupts and runs it  `lango chat` focused chat  `lango cockpit` dashboard"
		}
		if p.hasWorkbenchStarterPromptArmed() {
			return "Enter submits starter  1-3 replace starter  Type to edit prompt  `lango chat` focused chat  `lango cockpit` dashboard"
		}
		if p.canUseWorkbenchStarterPromptKeys() {
			if p.hasWorkbenchTurnNeedingAttention() {
				return "Enter recovery starter  1-3 starter prompts  Type recovery prompt here  `lango chat` focused chat  `lango cockpit` dashboard"
			}
			if p.hasWorkbenchCompletedTurn() {
				return "Enter next-step starter  1-3 starter prompts  Type next prompt here  `lango chat` focused chat  `lango cockpit` dashboard"
			}
			return "Enter default starter  1-3 starter prompts  Type request here  `lango chat` focused chat  `lango cockpit` dashboard"
		}
		return "Type request here  `lango chat` focused chat  `lango cockpit` dashboard"
	}
	return "Type request here  `lango chat` focused chat"
}

func (p *MissionControlPage) hasWorkbenchTurnInFlight() bool {
	return p != nil && p.surface == missionControlSurfaceWorkbench && p.composer != nil && p.composer.IsStreamingTurn()
}

func (p *MissionControlPage) hasWorkbenchQueuedFollowUpDraft() bool {
	return p.hasWorkbenchTurnInFlight() && p.composer != nil && strings.TrimSpace(p.composer.ComposerValue()) != ""
}

func (p *MissionControlPage) applyWorkbenchStarterPromptKey(msg tea.KeyMsg) bool {
	if !p.canUseWorkbenchStarterPromptSelection() {
		return false
	}
	prompt, ok := p.workbenchStarterPromptForKey(msg)
	if !ok {
		return false
	}
	return p.applyWorkbenchStarterPrompt(prompt)
}

func (p *MissionControlPage) canUseWorkbenchStarterPromptKeys() bool {
	if p == nil || p.surface != missionControlSurfaceWorkbench || p.needsInitialSetup() || !p.isEmpty() || p.composer == nil {
		return false
	}
	return strings.TrimSpace(p.composer.ComposerValue()) == ""
}

func (p *MissionControlPage) canUseWorkbenchStarterPromptSelection() bool {
	if p == nil || p.surface != missionControlSurfaceWorkbench || p.needsInitialSetup() || !p.isEmpty() || p.composer == nil {
		return false
	}
	return strings.TrimSpace(p.composer.ComposerValue()) == "" || p.hasWorkbenchStarterPromptArmed() || p.hasWorkbenchQueuedFollowUpDraft()
}

func (p *MissionControlPage) hasWorkbenchStarterPromptArmed() bool {
	if p == nil || p.surface != missionControlSurfaceWorkbench || p.needsInitialSetup() || !p.isEmpty() || p.composer == nil {
		return false
	}
	value := strings.TrimSpace(p.composer.ComposerValue())
	if value == "" {
		return false
	}
	for _, prompt := range p.workbenchStarterPromptSet() {
		if value == strings.TrimSpace(prompt) {
			return true
		}
	}
	return false
}

func (p *MissionControlPage) workbenchStarterPromptForKey(msg tea.KeyMsg) (string, bool) {
	if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 {
		return "", false
	}
	idx := int(msg.Runes[0] - '1')
	prompts := p.workbenchStarterPromptSet()
	if idx < 0 || idx >= len(prompts) {
		return "", false
	}
	return prompts[idx], true
}

func (p *MissionControlPage) workbenchStarterPromptSet() []string {
	if len(p.starterPrompts) > 0 {
		return p.starterPrompts
	}
	return workbenchstart.DefaultPrompts()
}

func (p *MissionControlPage) seedDefaultWorkbenchStarterPrompt() bool {
	prompt := p.defaultWorkbenchSeedPrompt()
	if strings.TrimSpace(prompt) == "" {
		return false
	}
	return p.applyWorkbenchStarterPrompt(prompt)
}

func (p *MissionControlPage) defaultWorkbenchSeedPrompt() string {
	if p.hasWorkbenchTurnNeedingAttention() {
		return strings.TrimSpace(workbenchstart.RecoveryDefaultPrompt(p.workDir))
	}
	if p.hasWorkbenchCompletedTurn() {
		return strings.TrimSpace(workbenchstart.PostTurnDefaultPrompt(p.workDir))
	}
	return strings.TrimSpace(p.defaultStarterPrompt)
}

func (p *MissionControlPage) hasWorkbenchCompletedTurn() bool {
	if p == nil || p.surface != missionControlSurfaceWorkbench {
		return false
	}
	for _, item := range p.snapshot.Activities {
		if item.Kind == cockpit.MissionActivityAssistant || item.Kind == cockpit.MissionActivityTurn {
			return true
		}
	}
	return false
}

func (p *MissionControlPage) hasWorkbenchTurnNeedingAttention() bool {
	if p == nil || p.surface != missionControlSurfaceWorkbench || !p.hasWorkbenchCompletedTurn() {
		return false
	}
	for _, item := range p.snapshot.Activities {
		if item.Kind == cockpit.MissionActivityAssistant {
			return strings.HasPrefix(strings.TrimSpace(item.Summary), "Turn ")
		}
	}
	return false
}

func (p *MissionControlPage) queueDefaultWorkbenchFollowUp() bool {
	if !p.hasWorkbenchTurnInFlight() || p.hasWorkbenchQueuedFollowUpDraft() || p.composer == nil {
		return false
	}
	if !p.seedDefaultWorkbenchStarterPrompt() {
		return false
	}
	return true
}

func (p *MissionControlPage) applyWorkbenchStarterPrompt(prompt string) bool {
	if p.composer == nil || !p.canUseWorkbenchStarterPromptSelection() {
		return false
	}
	p.composer.SetComposerValue(prompt)
	p.focus = missionControlFocusComposer
	p.refreshSnapshot()
	return true
}
