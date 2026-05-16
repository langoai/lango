package pages

import (
	"context"
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/session"
)

func fakeListFn(sessions []session.SessionSummary, err error) func(context.Context) ([]session.SessionSummary, error) {
	return func(_ context.Context) ([]session.SessionSummary, error) {
		return sessions, err
	}
}

func TestSessionsPage_Title(t *testing.T) {
	p := NewSessionsPage(nil)
	assert.Equal(t, "Sessions", p.Title())
}

func TestSessionsPage_ShortHelp(t *testing.T) {
	now := time.Now()
	p := NewSessionsPage(nil)
	p.sessions = []session.SessionSummary{
		{Key: "s1", UpdatedAt: now},
		{Key: "s2", UpdatedAt: now.Add(-time.Minute)},
	}
	bindings := p.ShortHelp()
	assert.Len(t, bindings, 2)
	assert.Equal(t, "↑/k", bindings[0].Help().Key)
	assert.Equal(t, "↓/j", bindings[1].Help().Key)
}

func TestSessionsPage_ShortHelpHiddenWithoutRows(t *testing.T) {
	p := NewSessionsPage(nil)
	assert.Empty(t, p.ShortHelp())
}

func TestSessionsPage_ShortHelpHiddenWithSingleRow(t *testing.T) {
	now := time.Now()
	p := NewSessionsPage(nil)
	p.sessions = []session.SessionSummary{{Key: "s1", UpdatedAt: now}}
	assert.Empty(t, p.ShortHelp())
}

func TestSessionsPage_Init(t *testing.T) {
	p := NewSessionsPage(nil)
	cmd := p.Init()
	assert.Nil(t, cmd)
}

func TestSessionsPage_Activate_ReturnsCmd(t *testing.T) {
	now := time.Now()
	listFn := fakeListFn([]session.SessionSummary{
		{Key: "s1", CreatedAt: now, UpdatedAt: now},
	}, nil)
	p := NewSessionsPage(listFn)
	cmd := p.Activate()
	require.NotNil(t, cmd, "Activate should return a load command")

	msg := cmd()
	loaded, ok := msg.(sessionsLoadedMsg)
	require.True(t, ok)
	assert.Nil(t, loaded.err)
	assert.Len(t, loaded.sessions, 1)
	assert.Equal(t, "s1", loaded.sessions[0].Key)
}

func TestSessionsPage_Activate_Error(t *testing.T) {
	listFn := fakeListFn(nil, fmt.Errorf("db error"))
	p := NewSessionsPage(listFn)
	cmd := p.Activate()
	require.NotNil(t, cmd)

	msg := cmd()
	loaded := msg.(sessionsLoadedMsg)
	assert.NotNil(t, loaded.err)
	assert.Nil(t, loaded.sessions)
}

func TestSessionsPage_Activate_NilListFn(t *testing.T) {
	p := NewSessionsPage(nil)
	cmd := p.Activate()
	require.NotNil(t, cmd)

	msg := cmd()
	loaded := msg.(sessionsLoadedMsg)
	assert.NotNil(t, loaded.err)
}

func TestSessionsPage_Deactivate(t *testing.T) {
	p := NewSessionsPage(nil)
	p.Deactivate()
}

func TestSessionsPage_UpdateLoadedMsg(t *testing.T) {
	now := time.Now()
	p := NewSessionsPage(nil)
	msg := sessionsLoadedMsg{
		sessions: []session.SessionSummary{
			{Key: "b", UpdatedAt: now.Add(-time.Hour)},
			{Key: "a", UpdatedAt: now},
		},
	}
	model, _ := p.Update(msg)
	sp := model.(*SessionsPage)
	assert.Len(t, sp.sessions, 2)
	assert.Nil(t, sp.loadErr)
	assert.Equal(t, "a", sp.sessions[0].Key)
	assert.Equal(t, "b", sp.sessions[1].Key)
}

func TestSessionsPage_UpdateLoadedError(t *testing.T) {
	p := NewSessionsPage(nil)
	msg := sessionsLoadedMsg{err: fmt.Errorf("fail")}
	model, _ := p.Update(msg)
	sp := model.(*SessionsPage)
	assert.NotNil(t, sp.loadErr)
	assert.Nil(t, sp.sessions)
}

func TestSessionsPage_CursorNavigation(t *testing.T) {
	now := time.Now()
	p := NewSessionsPage(nil)
	p.sessions = []session.SessionSummary{
		{Key: "a", UpdatedAt: now},
		{Key: "b", UpdatedAt: now},
		{Key: "c", UpdatedAt: now},
	}

	p.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, p.cursor)

	p.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 2, p.cursor)

	p.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 2, p.cursor)

	p.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 1, p.cursor)

	p.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, p.cursor)
	p.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, p.cursor)
}

func TestSessionsPage_WindowSizeMsg(t *testing.T) {
	p := NewSessionsPage(nil)
	_, _ = p.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	assert.Equal(t, 120, p.width)
	assert.Equal(t, 40, p.height)
}

func TestSessionsPage_ViewEmpty(t *testing.T) {
	p := NewSessionsPage(nil)
	view := p.View()
	assert.Contains(t, view, "Sessions")
	assert.Contains(t, view, "Session list is not configured.")
}

func TestSessionsPage_ViewEmptyConfiguredList(t *testing.T) {
	p := NewSessionsPage(fakeListFn([]session.SessionSummary{}, nil))
	view := p.View()
	assert.Contains(t, view, "No sessions found.")
}

func TestSessionsPage_ViewWithSessions(t *testing.T) {
	now := time.Now()
	p := NewSessionsPage(fakeListFn(nil, nil))
	p.sessions = []session.SessionSummary{
		{Key: "session-\x1b[31malpha\nops", UpdatedAt: now.Add(-5 * time.Minute)},
		{Key: "session-beta", UpdatedAt: now.Add(-2 * time.Hour)},
	}
	p.width = 80
	view := p.View()
	assert.Contains(t, view, "session-alpha ops")
	assert.Contains(t, view, "session-beta")
	assert.Contains(t, view, "5m ago")
	assert.Contains(t, view, "2h ago")
	assert.NotContains(t, view, "\x1b")
}

func TestSessionsPage_ViewWithError(t *testing.T) {
	p := NewSessionsPage(fakeListFn(nil, nil))
	p.loadErr = fmt.Errorf("connection \x1b[31mrefused\nnow")
	view := p.View()
	assert.Contains(t, view, "Session list failed to load")
	assert.Contains(t, view, "connection refused now")
	assert.NotContains(t, view, "\x1b")
}

func TestSessionsRelativeTime(t *testing.T) {
	tests := []struct {
		give time.Time
		want string
	}{
		{give: time.Now().Add(-30 * time.Second), want: "just now"},
		{give: time.Now().Add(-5 * time.Minute), want: "5m ago"},
		{give: time.Now().Add(-90 * time.Minute), want: "1h ago"},
		{give: time.Now().Add(-3 * 24 * time.Hour), want: "3d ago"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tui.RelativeTimeHuman(time.Now(), tt.give))
		})
	}
}
