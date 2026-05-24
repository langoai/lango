package pages

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/cockpit"
	"github.com/langoai/lango/internal/cli/workbenchstart"
	"github.com/langoai/lango/internal/config"
)

func newWorkbenchHelperTestPage(t *testing.T, snapshot cockpit.MissionControlSnapshot) *MissionControlPage {
	t.Helper()

	page := loadedMissionControlPageForSurface(t, snapshot, true)
	page.starterPrompts = []string{
		"Summarize the repository",
		"Identify the next milestone",
		"Draft the implementation plan",
	}
	page.defaultStarterPrompt = "Summarize the repository"
	return page
}

func putWorkbenchComposerInFlight(t *testing.T, page *MissionControlPage) {
	t.Helper()

	require.NotNil(t, page.composer)
	page.composer.SetComposerValue("active request")
	cmd := page.composer.SubmitComposerWithParent(context.Background())
	require.NotNil(t, cmd)
	require.True(t, page.hasWorkbenchTurnInFlight())
	require.Empty(t, page.composer.ComposerValue())
}

func TestMissionControlWorkbenchStarterPromptSelectionGuardsAndApplies(t *testing.T) {
	t.Parallel()

	page := newWorkbenchHelperTestPage(t, cockpit.MissionControlSnapshot{})

	prompt, ok := page.workbenchStarterPromptForKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	require.True(t, ok)
	assert.Equal(t, "Identify the next milestone", prompt)

	for _, msg := range []tea.KeyMsg{
		{Type: tea.KeyEnter},
		{Type: tea.KeyRunes, Runes: []rune{'0'}},
		{Type: tea.KeyRunes, Runes: []rune{'4'}},
		{Type: tea.KeyRunes, Runes: []rune{'1', '2'}},
	} {
		got, selected := page.workbenchStarterPromptForKey(msg)
		assert.False(t, selected)
		assert.Empty(t, got)
	}

	assert.True(t, page.applyWorkbenchStarterPromptKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}))
	assert.Equal(t, "Identify the next milestone", page.composer.ComposerValue())
	assert.Equal(t, missionControlFocusComposer, page.focus)

	assert.True(t, page.hasWorkbenchStarterPromptArmed())
	assert.True(t, page.applyWorkbenchStarterPrompt("Draft the implementation plan"))
	assert.Equal(t, "Draft the implementation plan", page.composer.ComposerValue())

	page.composer.SetComposerValue("custom draft")
	assert.False(t, page.hasWorkbenchStarterPromptArmed())
	assert.False(t, page.applyWorkbenchStarterPrompt("Summarize the repository"))
	assert.Equal(t, "custom draft", page.composer.ComposerValue())

	page.surface = missionControlSurfaceCockpit
	assert.False(t, page.applyWorkbenchStarterPromptKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}))
	assert.Equal(t, "custom draft", page.composer.ComposerValue())
}

func TestMissionControlWorkbenchDefaultSeedAndFollowUpBranches(t *testing.T) {
	t.Parallel()

	page := newWorkbenchHelperTestPage(t, cockpit.MissionControlSnapshot{})
	assert.True(t, page.seedDefaultWorkbenchStarterPrompt())
	assert.Equal(t, "Summarize the repository", page.composer.ComposerValue())

	page.composer.SetComposerValue("")
	page.defaultStarterPrompt = " \t\n "
	assert.False(t, page.seedDefaultWorkbenchStarterPrompt())
	assert.Empty(t, page.composer.ComposerValue())

	page.defaultStarterPrompt = "Summarize the repository"
	putWorkbenchComposerInFlight(t, page)
	assert.True(t, page.queueDefaultWorkbenchFollowUp())
	assert.Equal(t, "Summarize the repository", page.composer.ComposerValue())
	assert.True(t, page.hasWorkbenchQueuedFollowUpDraft())
	assert.False(t, page.queueDefaultWorkbenchFollowUp())

	emptyDefaultPage := newWorkbenchHelperTestPage(t, cockpit.MissionControlSnapshot{})
	emptyDefaultPage.defaultStarterPrompt = ""
	putWorkbenchComposerInFlight(t, emptyDefaultPage)
	assert.False(t, emptyDefaultPage.queueDefaultWorkbenchFollowUp())
	assert.Empty(t, emptyDefaultPage.composer.ComposerValue())
}

func TestMissionControlWorkbenchActivityDefaultsPreferRecoveryAndPostTurn(t *testing.T) {
	t.Parallel()

	recoveryPage := newWorkbenchHelperTestPage(t, cockpit.MissionControlSnapshot{
		Activities: []cockpit.ActivityView{{
			Kind:    cockpit.MissionActivityAssistant,
			Summary: " Turn failed: inspect logs ",
		}},
	})
	assert.True(t, recoveryPage.hasWorkbenchTurnNeedingAttention())
	assert.Equal(t, workbenchstart.RecoveryDefaultPrompt(""), recoveryPage.defaultWorkbenchSeedPrompt())
	assert.True(t, recoveryPage.seedDefaultWorkbenchStarterPrompt())
	assert.Equal(t, workbenchstart.RecoveryDefaultPrompt(""), recoveryPage.composer.ComposerValue())

	postTurnPage := newWorkbenchHelperTestPage(t, cockpit.MissionControlSnapshot{
		Activities: []cockpit.ActivityView{{
			Kind:    cockpit.MissionActivityTurn,
			Summary: "Turn completed",
		}},
	})
	assert.True(t, postTurnPage.hasWorkbenchCompletedTurn())
	assert.False(t, postTurnPage.hasWorkbenchTurnNeedingAttention())
	assert.Equal(t, workbenchstart.PostTurnDefaultPrompt(""), postTurnPage.defaultWorkbenchSeedPrompt())
}

func TestMissionControlWorkbenchHintsCoverSetupStarterRunningAndComposerStates(t *testing.T) {
	t.Parallel()

	setupPage := newWorkbenchHelperTestPage(t, cockpit.MissionControlSnapshot{})
	setupPage.cfg = &config.Config{}
	assert.Contains(t, setupPage.setupHint(), "lango onboard")
	assert.Empty(t, setupPage.starterHint())
	assert.Equal(t, "Setup first: `lango onboard`, `lango settings`, or `lango doctor`.", setupPage.composerEmptyStateHint())

	starterPage := newWorkbenchHelperTestPage(t, cockpit.MissionControlSnapshot{})
	assert.Empty(t, starterPage.setupHint())
	assert.Contains(t, starterPage.starterHint(), "Quick start")
	assert.Contains(t, starterPage.composerEmptyStateHint(), "Press `Enter`")
	assert.Contains(t, starterPage.footerSurfaceHint(), "Enter default starter")

	assert.True(t, starterPage.seedDefaultWorkbenchStarterPrompt())
	assert.Equal(t, "Starter ready: press `Enter` to run it, use `1-3` to replace it, or edit the prompt before sending.", starterPage.seededStarterHint())
	assert.Equal(t, "Enter submits starter  1-3 replace starter  Type to edit prompt  `lango chat` focused chat  `lango cockpit` dashboard", starterPage.footerSurfaceHint())
	assert.Empty(t, starterPage.starterHint())

	runningPage := newWorkbenchHelperTestPage(t, cockpit.MissionControlSnapshot{})
	putWorkbenchComposerInFlight(t, runningPage)
	assert.Equal(t, "Current request running: type the next prompt now, or press `Enter` to interrupt and run it.", runningPage.runningStarterHint())
	assert.Equal(t, "Current request running... Type the next prompt, or press Enter to interrupt.", runningPage.composerEmptyStateHint())
	assert.Contains(t, runningPage.footerSurfaceHint(), "Request running")

	assert.True(t, runningPage.queueDefaultWorkbenchFollowUp())
	assert.Equal(t, "Follow-up ready: press `Enter` to interrupt and run it, use `1-3` to replace it, or keep editing before sending.", runningPage.runningStarterHint())
	assert.Empty(t, runningPage.composerEmptyStateHint())
	assert.Contains(t, runningPage.footerSurfaceHint(), "Enter runs staged follow-up")
}

func TestMissionControlWorkbenchFooterAndStarterHintsReflectProposalAndHistory(t *testing.T) {
	t.Parallel()

	proposalPage := newWorkbenchHelperTestPage(t, cockpit.MissionControlSnapshot{
		Missions: []cockpit.MissionView{{
			ID:    "proposal-1",
			Title: "Review proposal",
			Kind:  cockpit.MissionKindProposed,
		}},
	})
	proposalPage.focus = missionControlFocusMissions
	assert.Contains(t, proposalPage.footerSurfaceHint(), "Enter accepts proposal")
	assert.Contains(t, proposalPage.footerSurfaceHint(), "`lango cockpit` dashboard")

	recoveryPage := newWorkbenchHelperTestPage(t, cockpit.MissionControlSnapshot{
		Activities: []cockpit.ActivityView{{
			Kind:    cockpit.MissionActivityAssistant,
			Summary: "Turn failed: tool error",
		}},
	})
	assert.Contains(t, recoveryPage.starterHint(), "Recovery step")
	assert.Contains(t, recoveryPage.composerEmptyStateHint(), "Recovery step")
	assert.Contains(t, recoveryPage.footerSurfaceHint(), "Enter recovery starter")

	completedPage := newWorkbenchHelperTestPage(t, cockpit.MissionControlSnapshot{
		Activities: []cockpit.ActivityView{{
			Kind:    cockpit.MissionActivityTurn,
			Summary: "Turn completed",
		}},
	})
	assert.Contains(t, completedPage.starterHint(), "Next step")
	assert.Contains(t, completedPage.composerEmptyStateHint(), "Next step")
	assert.Contains(t, completedPage.footerSurfaceHint(), "Enter next-step starter")
}
