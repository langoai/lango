package app

import (
	"context"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/runledger"
)

// runLedgerValues holds the outputs of the RunLedger module.
type runLedgerValues struct {
	store runledger.RunLedgerStore
	pev   *runledger.PEVEngine
}

// runLedgerModule initializes the RunLedger Task OS subsystem.
type runLedgerModule struct {
	cfg *config.Config
}

func (m *runLedgerModule) Name() string { return "runledger" }
func (m *runLedgerModule) Provides() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesRunLedger}
}
func (m *runLedgerModule) DependsOn() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesSupervisor}
}
func (m *runLedgerModule) Enabled() bool { return m.cfg.RunLedger.Enabled }

func (m *runLedgerModule) Init(_ context.Context, _ appinit.Resolver) (*appinit.ModuleResult, error) {
	// Use in-memory store for Phase 1 (shadow mode).
	// Ent-backed store will be implemented in Phase 2.
	// Workspace-aware validation remains phase-gated: the PEV engine supports
	// WithWorkspace(), but Phase 1 intentionally keeps runtime isolation disabled.
	// Phase 4 activates workspace wiring as part of the execution-isolation rollout.
	store := runledger.NewMemoryStore()
	validators := runledger.DefaultValidators()
	pev := runledger.NewPEVEngine(store, validators)

	tools := runledger.BuildTools(store, pev)

	vals := &runLedgerValues{
		store: store,
		pev:   pev,
	}

	return &appinit.ModuleResult{
		Tools: tools,
		Values: map[appinit.Provides]interface{}{
			appinit.ProvidesRunLedger: vals,
		},
		CatalogEntries: []appinit.CatalogEntry{
			{
				Category:    "runledger",
				Description: "Task OS: durable execution with PEV verification",
				ConfigKey:   "runLedger.enabled",
				Enabled:     true,
				Tools:       tools,
			},
		},
	}, nil
}
