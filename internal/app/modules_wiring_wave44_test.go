package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/agentrt"
	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/toolcatalog"
)

func TestWave44FoundationCatalogEntriesOmitWebCategoryWithoutWebTools(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Tools.Browser.Enabled = true

	entries := buildFoundationCatalogEntries(cfg, []*agent.Tool{
		{Name: "exec_run"},
		{Name: "fs_read"},
		{Name: "browser_open"},
		{Name: "custom_tool"},
	}, nil, nil)

	assert.Equal(t, []string{"exec_run"}, catalogEntryToolNames(requireCatalogEntry(t, entries, "exec")))
	assert.Equal(t, []string{"fs_read"}, catalogEntryToolNames(requireCatalogEntry(t, entries, "filesystem")))
	assert.Equal(t, []string{"browser_open"}, catalogEntryToolNames(requireCatalogEntry(t, entries, "browser")))
	assert.Nil(t, wave44CatalogEntry(entries, "web"))
}

func TestWave44AutomationAgentRunStoreMirrorPreconditions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configure  func(*config.Config)
		values     func() *runLedgerValues
		wantMirror bool
	}{
		{
			name:      "nil run ledger values",
			configure: enableWave44RunLedgerMirrorConfig,
		},
		{
			name:      "nil run ledger store",
			configure: enableWave44RunLedgerMirrorConfig,
			values: func() *runLedgerValues {
				return &runLedgerValues{}
			},
		},
		{
			name: "run ledger disabled",
			configure: func(cfg *config.Config) {
				cfg.RunLedger.Enabled = false
				cfg.RunLedger.WriteThrough = true
			},
			values: func() *runLedgerValues {
				return &runLedgerValues{store: runledger.NewMemoryStore()}
			},
		},
		{
			name: "write through disabled",
			configure: func(cfg *config.Config) {
				cfg.RunLedger.Enabled = true
				cfg.RunLedger.WriteThrough = false
			},
			values: func() *runLedgerValues {
				return &runLedgerValues{store: runledger.NewMemoryStore()}
			},
		},
		{
			name:       "write through enabled with store",
			configure:  enableWave44RunLedgerMirrorConfig,
			values:     func() *runLedgerValues { return &runLedgerValues{store: runledger.NewMemoryStore()} },
			wantMirror: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.DefaultConfig()
			if tt.configure != nil {
				tt.configure(cfg)
			}
			var values *runLedgerValues
			if tt.values != nil {
				values = tt.values()
			}

			store := newAutomationAgentRunStore(cfg, values, nil)
			require.NotNil(t, store)

			_, mirrored := store.(*agentrt.RunLedgerMirrorStore)
			assert.Equal(t, tt.wantMirror, mirrored)
		})
	}
}

func TestWave44CatalogSourceModeWithNilConfigFallsBackToVisibleTools(t *testing.T) {
	t.Parallel()

	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{
		Name:        "review",
		Description: "Review tools",
		Enabled:     true,
	})
	catalog.Register("review", []*agent.Tool{
		{Name: "review_read", Description: "read", SafetyLevel: agent.SafetyLevelSafe},
		{Name: "review_write", Description: "write", SafetyLevel: agent.SafetyLevelModerate},
	})

	section := (&catalogSourceAdapter{catalog: catalog}).BuildToolCatalogSection("focused")

	assert.Contains(t, section, "## Tools Available in `focused` Mode")
	assert.Contains(t, section, "review_read")
	assert.Contains(t, section, "review_write")
	assert.Contains(t, section, "Only tools in this mode's allowlist")
	assert.NotContains(t, section, "Disabled categories")
}

func TestWave44BuildSubAgentPromptFuncDefaultPathStripsGlobalSections(t *testing.T) {
	t.Parallel()

	buildPrompt := buildSubAgentPromptFunc(&config.AgentConfig{})
	got := buildPrompt("worker-b", "Worker B should focus on deterministic helper tests.")

	assert.Contains(t, got, "Worker B should focus on deterministic helper tests.")
	assert.Contains(t, got, "Safety Guidelines")
	assert.NotContains(t, got, "production-grade AI assistant")
	assert.NotContains(t, got, "Tool Usage Guidelines")
}

func enableWave44RunLedgerMirrorConfig(cfg *config.Config) {
	cfg.RunLedger.Enabled = true
	cfg.RunLedger.WriteThrough = true
}

func wave44CatalogEntry(entries []appinit.CatalogEntry, category string) *appinit.CatalogEntry {
	for i := range entries {
		if entries[i].Category == category {
			return &entries[i]
		}
	}
	return nil
}
