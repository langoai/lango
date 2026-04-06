package pages

import (
	"context"
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	p := NewSessionsPage(nil)
	bindings := p.ShortHelp()
	assert.Len(t, bindings, 2)
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
			{Key: "a", UpdatedAt: now},
			{Key: "b", UpdatedAt: now.Add(-time.Hour)},
		},
	}
	model, _ := p.Update(msg)
	sp := model.(*SessionsPage)
	assert.Len(t, sp.sessions, 2)
	assert.Nil(t, sp.loadErr)
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
	assert.Contains(t, view, "No sessions found")
}

func TestSessionsPage_ViewWithSessions(t *testing.T) {
	now := time.Now()
	p := NewSessionsPage(nil)
	p.sessions = []session.SessionSummary{
		{Key: "session-alpha", UpdatedAt: now.Add(-5 * time.Minute)},
		{Key: "session-beta", UpdatedAt: now.Add(-2 * time.Hour)},
	}
	p.width = 80
	view := p.View()
	assert.Contains(t, view, "session-alpha")
	assert.Contains(t, view, "session-beta")
	assert.Contains(t, view, "5m ago")
	assert.Contains(t, view, "2h ago")
}

func TestSessionsPage_ViewWithError(t *testing.T) {
	p := NewSessionsPage(nil)
	p.loadErr = fmt.Errorf("connection refused")
	view := p.View()
	assert.Contains(t, view, "Error")
	assert.Contains(t, view, "connection refused")
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
			assert.Equal(t, tt.want, sessionsRelativeTime(tt.give))
		})
	}
}
