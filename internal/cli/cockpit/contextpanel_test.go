package cockpit

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/observability"
)

func TestContextPanel_NewContextPanel(t *testing.T) {
	collector := observability.NewCollector()
	panel := NewContextPanel(collector)

	require.NotNil(t, panel)
	assert.False(t, panel.visible)
	assert.False(t, panel.tickActive)
	assert.Equal(t, 28, panel.width)
}

func TestContextPanel_InitReturnsNil(t *testing.T) {
	panel := NewContextPanel(nil)
	cmd := panel.Init()
	assert.Nil(t, cmd)
}

func TestContextPanel_ViewEmptyWhenNotVisible(t *testing.T) {
	panel := NewContextPanel(nil)
	panel.visible = false
	view := panel.View()
	assert.Empty(t, view)
}

func TestContextPanel_ViewRendersSections(t *testing.T) {
	collector := observability.NewCollector()
	collector.RecordTokenUsage(observability.TokenUsage{
		InputTokens:  100,
		OutputTokens: 200,
		TotalTokens:  300,
		CacheTokens:  50,
	})
	collector.RecordToolExecution("web_search", "navigator", time.Millisecond*100, true)

	panel := NewContextPanel(collector)
	panel.visible = true
	panel.height = 30
	panel.refreshSnapshot()

	view := panel.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Token Usage")
	assert.Contains(t, view, "Tool Stats")
	assert.Contains(t, view, "System")
	assert.Contains(t, view, "web_search")
}

func TestContextPanel_StartActivatesTick(t *testing.T) {
	collector := observability.NewCollector()
	panel := NewContextPanel(collector)

	cmd := panel.Start()
	assert.True(t, panel.tickActive)
	assert.NotNil(t, cmd, "Start should return a tick command")
}

func TestContextPanel_StopDeactivatesTick(t *testing.T) {
	panel := NewContextPanel(nil)
	panel.tickActive = true

	panel.Stop()
	assert.False(t, panel.tickActive)
}

func TestContextPanel_TickRefreshesSnapshot(t *testing.T) {
	collector := observability.NewCollector()
	panel := NewContextPanel(collector)
	panel.tickActive = true

	// Record some token usage.
	collector.RecordTokenUsage(observability.TokenUsage{
		InputTokens: 500,
		TotalTokens: 500,
	})

	// Send a tick.
	updated, cmd := panel.Update(contextTickMsg(time.Now()))
	panel = updated.(*ContextPanel)

	assert.NotNil(t, cmd, "active tick should produce next tick command")
	assert.Equal(t, int64(500), panel.snapshot.TokenUsageTotal.InputTokens)
}

func TestContextPanel_TickStopsWhenInactive(t *testing.T) {
	panel := NewContextPanel(nil)
	panel.tickActive = false

	_, cmd := panel.Update(contextTickMsg(time.Now()))
	assert.Nil(t, cmd, "inactive tick should not produce next tick")
}

func TestContextPanel_WindowSizeMsg(t *testing.T) {
	panel := NewContextPanel(nil)
	msg := tea.WindowSizeMsg{Width: 30, Height: 50}

	updated, _ := panel.Update(msg)
	p := updated.(*ContextPanel)

	assert.Equal(t, 30, p.width)
	assert.Equal(t, 50, p.height)
}

func TestContextPanel_SetHeight(t *testing.T) {
	panel := NewContextPanel(nil)
	panel.SetHeight(42)
	assert.Equal(t, 42, panel.height)
}

func TestContextPanel_SetVisible(t *testing.T) {
	panel := NewContextPanel(nil)
	panel.SetVisible(true)
	assert.True(t, panel.visible)
	panel.SetVisible(false)
	assert.False(t, panel.visible)
}

func TestContextPanel_NilCollector(t *testing.T) {
	panel := NewContextPanel(nil)
	panel.visible = true
	panel.height = 20

	// Should not panic.
	panel.refreshSnapshot()
	view := panel.View()
	assert.Contains(t, view, "No tool executions")
}

func TestContextPanel_TopFiveTools(t *testing.T) {
	collector := observability.NewCollector()
	// Record 7 tools.
	for i := 0; i < 7; i++ {
		name := string(rune('a'+i)) + "_tool"
		for j := 0; j <= i; j++ {
			collector.RecordToolExecution(name, "", time.Millisecond, true)
		}
	}

	panel := NewContextPanel(collector)
	panel.visible = true
	panel.height = 30
	panel.refreshSnapshot()

	view := panel.View()
	// g_tool (7 calls) through c_tool (3 calls) should appear.
	assert.Contains(t, view, "g_tool")
	// a_tool (1 call) should NOT appear — it's rank 7.
	assert.NotContains(t, view, "a_tool")
}

func TestContextPanel_NoChannels(t *testing.T) {
	panel := NewContextPanel(nil)
	panel.visible = true
	panel.height = 30

	view := panel.View()
	assert.NotContains(t, view, "Channels")
}

func TestContextPanel_WithChannels(t *testing.T) {
	panel := NewContextPanel(nil)
	panel.visible = true
	panel.height = 30
	panel.SetChannelStatuses([]channelStatus{
		{Name: "slack", Connected: true, MessageCount: 10, LastActivity: time.Now()},
		{Name: "discord", Connected: true, MessageCount: 3, LastActivity: time.Now()},
	})

	view := panel.View()
	assert.Contains(t, view, "Channels")
	assert.Contains(t, view, "slack")
	assert.Contains(t, view, "discord")
	assert.Contains(t, view, "●")
}

func TestContextPanel_DisconnectedChannel(t *testing.T) {
	panel := NewContextPanel(nil)
	panel.visible = true
	panel.height = 30
	panel.SetChannelStatuses([]channelStatus{
		{Name: "email", Connected: false, MessageCount: 0, LastActivity: time.Time{}},
	})

	view := panel.View()
	assert.Contains(t, view, "Channels")
	assert.Contains(t, view, "email")
	assert.Contains(t, view, "○")
}

func TestContextPanel_MessageCount(t *testing.T) {
	panel := NewContextPanel(nil)
	panel.visible = true
	panel.height = 30
	panel.SetChannelStatuses([]channelStatus{
		{Name: "webhook", Connected: true, MessageCount: 5, LastActivity: time.Now()},
	})

	view := panel.View()
	assert.Contains(t, view, "5 msgs")
}

func TestContextPanel_SetChannelStatuses(t *testing.T) {
	panel := NewContextPanel(nil)

	statuses := []channelStatus{
		{Name: "slack", Connected: true, MessageCount: 42, LastActivity: time.Now()},
		{Name: "email", Connected: false, MessageCount: 0, LastActivity: time.Time{}},
	}
	panel.SetChannelStatuses(statuses)

	require.Len(t, panel.channelStatuses, 2)
	assert.Equal(t, "slack", panel.channelStatuses[0].Name)
	assert.True(t, panel.channelStatuses[0].Connected)
	assert.Equal(t, 42, panel.channelStatuses[0].MessageCount)
	assert.Equal(t, "email", panel.channelStatuses[1].Name)
	assert.False(t, panel.channelStatuses[1].Connected)

	// Verify defensive copy — mutating the original should not affect panel.
	statuses[0].Name = "mutated"
	assert.Equal(t, "slack", panel.channelStatuses[0].Name)
}

func TestFormatCompact(t *testing.T) {
	tests := []struct {
		give int64
		want string
	}{
		{give: 0, want: "0"},
		{give: 999, want: "999"},
		{give: 1000, want: "1,000"},
		{give: 12345, want: "12,345"},
		{give: 1234567, want: "1,234,567"},
		{give: -42, want: "-42"},
	}
	for _, tt := range tests {
		got := formatCompact(tt.give)
		assert.Equal(t, tt.want, got, "formatCompact(%d)", tt.give)
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		give time.Duration
		want string
	}{
		{give: 500 * time.Millisecond, want: "500ms"},
		{give: 90 * time.Second, want: "1m 30s"},
		{give: 2*time.Hour + 15*time.Minute, want: "2h 15m"},
	}
	for _, tt := range tests {
		got := formatUptime(tt.give)
		assert.Equal(t, tt.want, got, "formatUptime(%v)", tt.give)
	}
}
