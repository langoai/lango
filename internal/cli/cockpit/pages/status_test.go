package pages

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/tui"
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

func TestStatusPage_SanitizesFeatureStatusText(t *testing.T) {
	provider := func() []types.FeatureStatus {
		return []types.FeatureStatus{
			{Name: "\x1b[31mKnow\nledge\x1b[0m", Enabled: false, Reason: "\x1b[31mdisabled\nby\tconfig\x1b[0m"},
		}
	}
	p := NewStatusPage(provider, observability.NewCollector(), &config.Config{})
	p.Activate()

	view := p.View()
	assert.Contains(t, view, "Know ledge")
	assert.Contains(t, view, "disabled by config")
	assert.NotContains(t, view, "\x1b[31m")
	assert.NotContains(t, view, "\x1b[0m")
}

func TestStatusPage_SanitizesConfigSystemLabels(t *testing.T) {
	provider := func() []types.FeatureStatus { return nil }
	cfg := &config.Config{}
	cfg.Agent.Provider = "open\x1b[31mai\n"
	cfg.Agent.Model = "gpt\x1b[31m-5\n"

	p := NewStatusPage(provider, observability.NewCollector(), cfg)
	p.Activate()

	view := p.View()
	assert.Contains(t, view, "openai")
	assert.Contains(t, view, "gpt-5")
	assert.NotContains(t, view, "\x1b[31m")
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

func TestStatusPage_ViewWithGraphAdmissionMetrics(t *testing.T) {
	provider := func() []types.FeatureStatus { return nil }
	collector := observability.NewCollector()
	learningGroup := "learning"

	collector.RecordGraphAdmissionBatch(observability.GraphAdmissionBatchMetric{
		Source:           "learning",
		ProducerGroup:    &learningGroup,
		ValidatorSource:  "ontology_registry",
		BatchCount:       2,
		KnownCount:       5,
		UnknownCount:     1,
		UnvalidatedCount: 0,
	})
	collector.RecordGraphAdmissionBatch(observability.GraphAdmissionBatchMetric{
		Source:           "content_saved_extractor",
		ValidatorSource:  "unavailable",
		BatchCount:       1,
		KnownCount:       0,
		UnknownCount:     0,
		UnvalidatedCount: 3,
	})
	collector.RecordGraphExtractorDroppedUnknown("content_saved_extractor", 2)
	collector.RecordGraphAdmissionUnmappedSource("legacy_import", 1)
	collector.RecordGraphWriteFailure(1)

	p := NewStatusPage(provider, collector, &config.Config{})
	p.Activate()

	view := p.View()
	assert.Contains(t, view, "Graph Admission")
	assert.Contains(t, view, "learning  group=learning  validator=ontology_registry")
	assert.Contains(t, view, "content_saved_extractor  validator=unavailable")
	assert.NotContains(t, view, "content_saved_extractor  group=none")
	assert.NotContains(t, view, "group=none")
	assert.Contains(t, view, "Dropped unknown:")
	assert.Contains(t, view, "legacy_import  1 batches")
	assert.Contains(t, view, "1 failed batches")
}

func TestStatusPage_SanitizesGraphAdmissionText(t *testing.T) {
	provider := func() []types.FeatureStatus { return nil }
	collector := observability.NewCollector()
	group := "\x1b[31mlearn\ning\x1b[0m"

	collector.RecordGraphAdmissionBatch(observability.GraphAdmissionBatchMetric{
		Source:           "\x1b[31mlearn\ning\x1b[0m",
		ProducerGroup:    &group,
		ValidatorSource:  "\x1b[31montology\nregistry\x1b[0m",
		BatchCount:       1,
		KnownCount:       1,
		UnknownCount:     0,
		UnvalidatedCount: 0,
	})

	p := NewStatusPage(provider, collector, &config.Config{})
	p.Activate()

	view := p.View()
	assert.Contains(t, view, "learn ing  group=learn ing  validator=ontology registry")
	assert.NotContains(t, view, "\x1b[31m")
	assert.NotContains(t, view, "\x1b[0m")
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
			assert.Equal(t, tt.want, tui.FormatDuration(tt.give))
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
			assert.Equal(t, tt.want, tui.FormatNumber(tt.give))
		})
	}
}

func TestStatusPage_NilProvider(t *testing.T) {
	p := NewStatusPage(nil, observability.NewCollector(), &config.Config{})
	p.Activate()
	view := p.View()
	assert.Contains(t, view, "Feature Status")
	assert.Contains(t, view, "Feature status provider is not configured")
}

func TestStatusPage_NilCollector(t *testing.T) {
	provider := func() []types.FeatureStatus {
		return []types.FeatureStatus{{Name: "Test", Enabled: true}}
	}
	p := NewStatusPage(provider, nil, &config.Config{})
	p.Activate()
	view := p.View()
	assert.Contains(t, view, "Test")
	assert.Contains(t, view, "Metrics collector is not configured")
}
