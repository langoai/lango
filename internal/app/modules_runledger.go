package app

import (
	"context"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/bootstrap"
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
	cfg  *config.Config
	boot *bootstrap.Result
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
	// Phase 2 uses an Ent-backed store when the shared app database is available.
	// MemoryStore remains as a fallback for tests and non-bootstrapped contexts.
	// Workspace-aware validation remains phase-gated: the PEV engine supports
	// WithWorkspace(), but Phase 1 intentionally keeps runtime isolation disabled.
	// Phase 4 activates workspace wiring as part of the execution-isolation rollout.
	store := runledger.RunLedgerStore(runledger.NewMemoryStore())
	if m.boot != nil && m.boot.DBClient != nil {
		store = runledger.NewEntStore(m.boot.DBClient)
	}
	validators := runledger.DefaultValidators()
	pev := runledger.NewPEVEngine(store, validators)
	if m.cfg.RunLedger.WorkspaceIsolation {
		pev.WithWorkspace(runledger.NewWorkspaceManager())
	}

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
