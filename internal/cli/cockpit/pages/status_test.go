package pages

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/observability"
	"github.com/langoai/lango/internal/types"
)

func newTestPage() *StatusPage {
	provider := func() []types.FeatureStatus {
		return []types.FeatureStatus{
			{Name: "Knowledge", Enabled: true},
			{Name: "Graph", Enabled: false, Reason: "disabled by config"},
		}
	}
	collector := observability.NewCollector()
	cfg := &config.Config{}
	cfg.Agent.Provider = "openai"
	cfg.Agent.Model = "gpt-4o"

	return NewStatusPage(provider, collector, cfg)
}

func TestStatusPage_Title(t *testing.T) {
	p := newTestPage()
	assert.Equal(t, "Status", p.Title())
}

func TestStatusPage_ShortHelp(t *testing.T) {
	p := newTestPage()
	assert.Empty(t, p.ShortHelp())
}

func TestStatusPage_Init(t *testing.T) {
	p := newTestPage()
	cmd := p.Init()
	assert.Nil(t, cmd)
}

func TestStatusPage_Activate(t *testing.T) {
	p := newTestPage()
	cmd := p.Activate()

	assert.True(t, p.tickActive, "Activate should set tickActive=true")
	require.NotNil(t, cmd, "Activate should return a tick command")

	// Feature statuses should be populated after Activate.
	assert.Len(t, p.featureStatuses, 2)
}

func TestStatusPage_Deactivate(t *testing.T) {
	p := newTestPage()
	p.Activate()
	p.Deactivate()
	assert.False(t, p.tickActive, "Deactivate should set tickActive=false")
}

func TestStatusPage_TickWhenInactive(t *testing.T) {
	p := newTestPage()
	// tickActive is false by default; sending a tickMsg should produce no cmd.
	model, cmd := p.Update(tickMsg(time.Now()))
	assert.Nil(t, cmd, "tickMsg when inactive should return nil cmd")
	assert.NotNil(t, model)
}

func TestStatusPage_TickWhenActive(t *testing.T) {
	p := newTestPage()
	p.Activate()

	model, cmd := p.Update(tickMsg(time.Now()))
	require.NotNil(t, cmd, "tickMsg when active should return next tick cmd")
	assert.NotNil(t, model)
}

func TestStatusPage_WindowSizeMsg(t *testing.T) {
	p := newTestPage()
	_, _ = p.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	assert.Equal(t, 120, p.width)
	assert.Equal(t, 40, p.height)
}

func TestStatusPage_ViewContainsSections(t *testing.T) {
	p := newTestPage()
	p.Activate()

	view := p.View()
	assert.Contains(t, view, "Feature Status")
	assert.Contains(t, view, "Token Usage")
	assert.Contains(t, view, "Tool Execution")
	assert.Contains(t, view, "System")
	assert.Contains(t, view, "openai")
	assert.Contains(t, view, "gpt-4o")
}

func TestStatusPage_ViewShowsFeatureStatuses(t *testing.T) {
	p := newTestPage()
	p.Activate()

	view := p.View()
	assert.Contains(t, view, "Knowledge")
	assert.Contains(t, view, "enabled")
	assert.Contains(t, view, "Graph")
	assert.Contains(t, view, "disabled")
	assert.Contains(t, view, "disabled by config")
}

func TestStatusPage_ViewWithToolMetrics(t *testing.T) {
	provider := func() []types.FeatureStatus { return nil }
	collector := observability.NewCollector()
	collector.RecordToolExecution("read_file", "main", 100*time.Millisecond, true)
	collector.RecordToolExecution("read_file", "main", 200*time.Millisecond, true)
	collector.RecordToolExecution("exec_command", "main", 2*time.Second, false)

	p := NewStatusPage(provider, collector, &config.Config{})
	p.Activate()

	view := p.View()
	assert.Contains(t, view, "read_file")
	assert.Contains(t, view, "exec_command")
	assert.Contains(t, view, "Total executions")
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		give time.Duration
		want string
	}{
		{give: 500 * time.Millisecond, want: "500ms"},
		{give: 0, want: "0ms"},
		{give: 3*time.Minute + 42*time.Second, want: "3m 42s"},
		{give: 2*time.Hour + 15*time.Minute, want: "2h 15m"},
		{give: 5 * time.Second, want: "0m 5s"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, formatDuration(tt.give))
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		give int64
		want string
	}{
		{give: 0, want: "0"},
		{give: 999, want: "999"},
		{give: 1000, want: "1,000"},
		{give: 12345, want: "12,345"},
		{give: 1234567, want: "1,234,567"},
		{give: -5000, want: "-5,000"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, formatNumber(tt.give))
		})
	}
}

func TestStatusPage_NilProvider(t *testing.T) {
	p := NewStatusPage(nil, observability.NewCollector(), &config.Config{})
	p.Activate()
	// Should not panic with nil provider.
	view := p.View()
	assert.Contains(t, view, "Feature Status")
}

func TestStatusPage_NilCollector(t *testing.T) {
	provider := func() []types.FeatureStatus {
		return []types.FeatureStatus{{Name: "Test", Enabled: true}}
	}
	p := NewStatusPage(provider, nil, &config.Config{})
	p.Activate()
	// Should not panic with nil collector.
	view := p.View()
	assert.Contains(t, view, "Test")
}
