package pages

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/cli/chat"
	"github.com/langoai/lango/internal/cli/cockpit"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/mission"
	"github.com/langoai/lango/internal/proposal"
)

func TestMissionControlConstructorsWireSurfaceDependenciesAndComposer(t *testing.T) {
	t.Parallel()

	learningBuffer := cockpit.NewLearningSuggestionBuffer(nil)
	missionSvc := &stubMissionLifecycleService{}
	proposalReader := &stubMissionProposalReader{}
	proposalSvc := &stubMissionProposalService{}
	deps := cockpit.Deps{
		Config:          readyWorkbenchConfig(),
		WorkDir:         "  /tmp/lango-project  ",
		SessionKey:      "mission-session",
		MissionService:  missionSvc,
		ProposalReader:  proposalReader,
		ProposalService: proposalSvc,
		LearningBuffer:  learningBuffer,
	}

	cockpitPage := NewMissionControlPage(deps, nil)
	workbenchPage := NewWorkbenchMissionControlPage(deps, nil)

	require.NotNil(t, cockpitPage.composer)
	assert.Equal(t, missionControlSurfaceCockpit, cockpitPage.surface)
	assert.Equal(t, "mission-session", cockpitPage.sessionKey)
	assert.Same(t, missionSvc, cockpitPage.missionService)
	assert.Same(t, proposalReader, cockpitPage.proposalReader)
	assert.Same(t, proposalSvc, cockpitPage.proposalSvc)
	assert.Same(t, learningBuffer, cockpitPage.learningBuffer)

	require.NotNil(t, workbenchPage.composer)
	assert.Equal(t, missionControlSurfaceWorkbench, workbenchPage.surface)
	assert.Equal(t, "/tmp/lango-project", workbenchPage.workDir)
	assert.NotEmpty(t, workbenchPage.starterPrompts)
	assert.NotEmpty(t, workbenchPage.defaultStarterPrompt)
}

func TestMissionControlLifecycleAndRefreshCommandsAreDeterministic(t *testing.T) {
	t.Parallel()

	projector := &stubMissionControlProjector{
		snapshot: cockpit.MissionControlSnapshot{
			Missions: []cockpit.MissionView{{ID: "m-1", Title: "Loaded mission"}},
		},
	}
	taskSource := stubMissionTaskSource{
		tasks: []background.TaskSnapshot{{ID: "task-1"}},
	}
	page := newMissionControlPage(projector, taskSource, nil)

	assert.Equal(t, "Mission Control", page.Title())
	assert.Nil(t, page.Init())
	assert.Contains(t, page.View(), "Waiting for terminal size")

	updated, cmd := page.Update(tea.WindowSizeMsg{Width: 130, Height: 30})
	page = updated.(*MissionControlPage)
	require.Nil(t, cmd)

	page.Deactivate()
	updated, cmd = page.Update(missionControlTickMsg(time.Now()))
	page = updated.(*MissionControlPage)
	assert.Nil(t, cmd)
	assert.False(t, page.hasLoaded)
	assert.Equal(t, 0, projector.calls)

	updated, cmd = page.Update(cockpit.MissionControlRefreshMsg{})
	page = updated.(*MissionControlPage)
	assert.Nil(t, cmd)
	assert.True(t, page.hasLoaded)
	assert.Equal(t, 1, projector.calls)
	require.Len(t, projector.tasks, 1)
	assert.Equal(t, "task-1", projector.tasks[0].ID)
	assert.Contains(t, page.View(), "Loaded mission")
}

func TestMissionControlWideViewRendersBothTopLanes(t *testing.T) {
	t.Parallel()

	page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{
		Missions: []cockpit.MissionView{{
			ID:     "m-1",
			Title:  "Wide mission",
			Status: cockpit.MissionStatusRunning,
		}},
		Decision: &cockpit.DecisionView{
			ID:         "apr-1",
			Title:      "Approve deploy",
			Reason:     "Deployment needs approval.",
			EffectText: "Ships the release.",
			RiskLabel:  "Moderate",
		},
	})
	updated, _ := page.Update(tea.WindowSizeMsg{Width: 132, Height: 32})
	page = updated.(*MissionControlPage)

	view := page.View()
	assert.Contains(t, view, "Wide mission")
	assert.Contains(t, view, "Action: Approve deploy")
	assert.Contains(t, view, "Ships the release.")
}

func TestMissionControlNavigationClampsMissionAndDecisionCursors(t *testing.T) {
	t.Parallel()

	page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{
		Missions: []cockpit.MissionView{
			{ID: "m-1", Title: "First"},
			{ID: "m-2", Title: "Second"},
		},
	})

	updated, _ := page.Update(tea.KeyMsg{Type: tea.KeyUp})
	page = updated.(*MissionControlPage)
	assert.Equal(t, 0, page.missionCursor)

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyDown})
	page = updated.(*MissionControlPage)
	assert.Equal(t, 1, page.missionCursor)

	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyDown})
	page = updated.(*MissionControlPage)
	assert.Equal(t, 1, page.missionCursor)

	page.focus = missionControlFocusDecisions
	page.decisionCursor = 4
	updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyUp})
	page = updated.(*MissionControlPage)
	assert.Equal(t, 0, page.decisionCursor)
}

func TestMissionControlProposalCommandErrorsReturnSystemMessages(t *testing.T) {
	t.Parallel()

	t.Run("missing proposal row", func(t *testing.T) {
		t.Parallel()

		reader := &stubMissionProposalReader{items: make(map[string]proposal.Proposal)}
		proposalSvc := &stubMissionProposalService{}
		page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{
			Missions: []cockpit.MissionView{{
				ID:    "proposal-missing",
				Kind:  cockpit.MissionKindProposed,
				Title: "Missing proposal",
			}},
		})
		page.missionService = &stubMissionLifecycleService{}
		page.proposalReader = reader
		page.proposalSvc = proposalSvc
		page.focus = missionControlFocusMissions

		updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
		page = updated.(*MissionControlPage)
		require.NotNil(t, cmd)
		assert.Equal(t, 0, proposalSvc.acceptCalls)

		msgs := collectMissionControlConstructorsWireSurfaceDependenciesAndComposerImmediateMsgs(cmd)
		require.Len(t, msgs, 1)
		sys, ok := msgs[0].(chat.SystemMsg)
		require.True(t, ok)
		assert.Contains(t, sys.Text, `Proposal "proposal-missing" is no longer available`)
	})

	t.Run("proposal accept fails", func(t *testing.T) {
		t.Parallel()

		reader := &stubMissionProposalReader{
			items: map[string]proposal.Proposal{
				"proposal-1": {
					ProposalID: "proposal-1",
					SessionKey: "mission-session",
					Status:     proposal.ProposalStatusPrepared,
				},
			},
		}
		proposalSvc := &stubMissionProposalService{acceptErr: errors.New("store offline")}
		page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{
			Missions: []cockpit.MissionView{{
				ID:    "proposal-1",
				Kind:  cockpit.MissionKindProposed,
				Title: "Accept me",
			}},
		})
		page.missionService = &stubMissionLifecycleService{}
		page.proposalReader = reader
		page.proposalSvc = proposalSvc
		page.focus = missionControlFocusMissions

		updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
		page = updated.(*MissionControlPage)
		require.NotNil(t, cmd)
		assert.Equal(t, 1, proposalSvc.acceptCalls)

		msgs := collectMissionControlConstructorsWireSurfaceDependenciesAndComposerImmediateMsgs(cmd)
		require.Len(t, msgs, 1)
		sys, ok := msgs[0].(chat.SystemMsg)
		require.True(t, ok)
		assert.Contains(t, sys.Text, "Proposal acceptance failed: store offline")
	})

	t.Run("dismiss fails", func(t *testing.T) {
		t.Parallel()

		proposalSvc := &stubMissionProposalService{dismissErr: errors.New("permission denied")}
		page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{
			Missions: []cockpit.MissionView{{
				ID:    "proposal-2",
				Kind:  cockpit.MissionKindProposed,
				Title: "Dismiss me",
			}},
		})
		page.proposalSvc = proposalSvc
		page.focus = missionControlFocusMissions

		updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		page = updated.(*MissionControlPage)
		require.NotNil(t, cmd)
		assert.Equal(t, 1, proposalSvc.dismissCalls)

		msgs := collectMissionControlConstructorsWireSurfaceDependenciesAndComposerImmediateMsgs(cmd)
		require.Len(t, msgs, 1)
		sys, ok := msgs[0].(chat.SystemMsg)
		require.True(t, ok)
		assert.Contains(t, sys.Text, "Proposal dismiss failed: permission denied")
	})
}

func TestMissionControlProposalRestoreFailureIsReported(t *testing.T) {
	t.Parallel()

	reader := &stubMissionProposalReader{
		items: map[string]proposal.Proposal{
			"proposal-1": {
				ProposalID: "proposal-1",
				SessionKey: "mission-session",
				Source: proposal.ProposalSource{
					Kind: "learning",
					Ref:  "s-1",
				},
				Title:  "Restore me",
				Status: proposal.ProposalStatusPrepared,
			},
		},
	}
	proposalSvc := &stubMissionProposalService{
		proposalReader: reader,
		restoreErr:     errors.New("restore failed"),
	}
	page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{
		Missions: []cockpit.MissionView{{
			ID:    "proposal-1",
			Kind:  cockpit.MissionKindProposed,
			Title: "Restore me",
		}},
	})
	page.missionService = &stubMissionLifecycleService{acceptErr: errors.New("mission offline")}
	page.proposalReader = reader
	page.proposalSvc = proposalSvc
	page.focus = missionControlFocusMissions

	updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(*MissionControlPage)
	require.NotNil(t, cmd)
	assert.Equal(t, 1, proposalSvc.restoreCalls)

	msgs := collectMissionControlConstructorsWireSurfaceDependenciesAndComposerImmediateMsgs(cmd)
	require.Len(t, msgs, 1)
	sys, ok := msgs[0].(chat.SystemMsg)
	require.True(t, ok)
	assert.Contains(t, sys.Text, "mission offline")
	assert.Contains(t, sys.Text, "restore failed")
}

func TestMissionControlLegacyLearningProposalUsesSuggestionFallback(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	learningBuffer := cockpit.NewLearningSuggestionBuffer(func() time.Time { return now })
	learningBuffer.Append(eventbus.LearningSuggestionEvent{
		SuggestionID: "suggestion-1",
		Rationale:    "Retry after bounded backoff",
		Timestamp:    now,
	})
	svc := &stubMissionLifecycleService{
		acceptResult: &mission.Mission{ID: uuid.MustParse("55555555-5555-5555-5555-555555555555")},
	}
	page := loadedMissionControlPage(t, cockpit.MissionControlSnapshot{
		Missions: []cockpit.MissionView{{
			ID:    "learn:suggestion-1",
			Kind:  cockpit.MissionKindProposed,
			Title: "Apply learning",
		}},
	})
	page.missionService = svc
	page.learningBuffer = learningBuffer
	page.sessionKey = "mission-session"
	page.focus = missionControlFocusMissions

	updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
	page = updated.(*MissionControlPage)

	assert.Nil(t, cmd)
	assert.Equal(t, 1, svc.acceptCalls)
	assert.Equal(t, "proposed_learning", svc.lastAcceptInput.SourceKind)
	assert.Equal(t, "suggestion-1", svc.lastAcceptInput.SourceRef)
	assert.Equal(t, "Retry after bounded backoff", svc.lastAcceptInput.Description)
	assert.Nil(t, page.lookupLearningSuggestion("suggestion-1"))
}

func TestMissionControlSubmitCommandAndErrorBranches(t *testing.T) {
	t.Parallel()

	t.Run("slash command bypasses durable mission creation", func(t *testing.T) {
		t.Parallel()

		composer, executor := newMissionComposerWithExecutor(t, nil)
		composer.SetComposerValue("/help")
		svc := &stubMissionLifecycleService{}
		page := newMissionControlPage(&stubMissionControlProjector{}, stubMissionTaskSource{}, composer)
		page.missionService = svc
		page.focus = missionControlFocusComposer

		updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
		page = updated.(*MissionControlPage)
		require.NotNil(t, cmd)
		assert.Equal(t, 0, svc.startCalls)
		assert.Equal(t, 0, executor.calls, "command execution is intentionally not forced in this branch test")
	})

	t.Run("mission start failure keeps composer text", func(t *testing.T) {
		t.Parallel()

		composer, executor := newMissionComposerWithExecutor(t, nil)
		composer.SetComposerValue("start durable mission")
		page := newMissionControlPage(&stubMissionControlProjector{}, stubMissionTaskSource{}, composer)
		page.missionService = &stubMissionLifecycleService{startErr: errors.New("database locked")}
		page.sessionKey = "mission-session"
		page.focus = missionControlFocusComposer

		updated, cmd := page.Update(tea.KeyMsg{Type: tea.KeyEnter})
		page = updated.(*MissionControlPage)
		require.NotNil(t, cmd)
		assert.Equal(t, 0, executor.calls)
		assert.Equal(t, "start durable mission", page.composer.ComposerValue())

		msgs := collectMissionControlConstructorsWireSurfaceDependenciesAndComposerImmediateMsgs(cmd)
		require.Len(t, msgs, 1)
		sys, ok := msgs[0].(chat.SystemMsg)
		require.True(t, ok)
		assert.Contains(t, sys.Text, "Mission start failed: database locked")
	})
}

func TestMissionControlHelperBoundaries(t *testing.T) {
	t.Parallel()

	assert.False(t, (*MissionControlPage)(nil).focusedLaneHasAlternativeRow())
	assert.Equal(t, "submit", (*MissionControlPage)(nil).enterHelpDesc())
	assert.Empty(t, (*MissionControlPage)(nil).visibleDegradedNote())
	assert.Empty(t, (*MissionControlPage)(nil).latestWorkbenchAssistantSummary())
	assert.Equal(t, "Type a request here, or use `lango chat` for focused chat.", (*MissionControlPage)(nil).focusedChatHint())

	page := &MissionControlPage{focus: missionControlFocusMissions}
	assert.False(t, page.composerVisibleInCompact())

	page.composer = chat.New(chat.Deps{})
	assert.False(t, page.composerVisibleInCompact())
	page.composer.SetComposerValue("draft")
	assert.True(t, page.composerVisibleInCompact())
	page.focus = missionControlFocusComposer
	page.composer.SetComposerValue("")
	assert.True(t, page.composerVisibleInCompact())

	assert.Nil(t, visibleActivities(nil, 0, 6))
	activities := []cockpit.ActivityView{
		{Summary: "one"},
		{Summary: "two"},
		{Summary: "three"},
	}
	assert.Equal(t, activities, visibleActivities(activities, 2, 0))
	assert.Equal(t, []cockpit.ActivityView{{Summary: "three"}}, visibleActivities(activities, 9, 2))

	assert.False(t, isMissionControlPrintableKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}, Alt: true}))
	assert.False(t, isMissionControlPrintableKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\n'}}))
	assert.True(t, isMissionControlComposerEditingKey(tea.KeyMsg{Type: tea.KeyBackspace}))
	assert.False(t, isMissionControlComposerEditingKey(tea.KeyMsg{Type: tea.KeyEsc}))

	assert.Equal(t, "", cond(false, "hidden"))
	assert.Equal(t, 0, clamp(4, 5, 1))
	assert.Equal(t, 3, clamp(9, 1, 3))
	assert.Equal(t, 2, min(7, 2))
}

func collectMissionControlConstructorsWireSurfaceDependenciesAndComposerImmediateMsgs(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	switch msg := msg.(type) {
	case nil:
		return nil
	case tea.BatchMsg:
		out := make([]tea.Msg, 0, len(msg))
		for _, child := range msg {
			out = append(out, collectMissionControlConstructorsWireSurfaceDependenciesAndComposerImmediateMsgs(child)...)
		}
		return out
	default:
		return []tea.Msg{msg}
	}
}
