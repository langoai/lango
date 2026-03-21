package app

import (
	"context"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/provenance"
	"github.com/langoai/lango/internal/runledger"
)

// provenanceValues holds the outputs of the provenance module.
type provenanceValues struct {
	checkpointStore   provenance.CheckpointStore
	checkpointService *provenance.CheckpointService
	sessionTree       *provenance.SessionTree
}

// provenanceModule initializes the session provenance subsystem.
type provenanceModule struct {
	cfg  *config.Config
	boot *bootstrap.Result
}

func (m *provenanceModule) Name() string { return "provenance" }
func (m *provenanceModule) Provides() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesProvenance}
}
func (m *provenanceModule) DependsOn() []appinit.Provides {
	return []appinit.Provides{appinit.ProvidesRunLedger}
}
func (m *provenanceModule) Enabled() bool { return m.cfg.Provenance.Enabled }

func (m *provenanceModule) Init(_ context.Context, r appinit.Resolver) (*appinit.ModuleResult, error) {
	cpStore := provenance.CheckpointStore(provenance.NewMemoryStore())
	if m.boot != nil && m.boot.DBClient != nil {
		cpStore = provenance.NewEntCheckpointStore(m.boot.DBClient)
	}
	treeStore := provenance.SessionTreeStore(provenance.NewMemoryTreeStore())

	// Resolve RunLedger store if available.
	var ledgerStore runledger.RunLedgerStore
	if rlVals := r.Resolve(appinit.ProvidesRunLedger); rlVals != nil {
		if rv, ok := rlVals.(*runLedgerValues); ok {
			ledgerStore = rv.store
		}
	}

	cpService := provenance.NewCheckpointService(cpStore, ledgerStore, m.cfg.Provenance.Checkpoints)
	sessionTree := provenance.NewSessionTree(treeStore)

	// Register auto-checkpoint hook on the RunLedger store (post-construction).
	if ledgerStore != nil {
		if setter, ok := ledgerStore.(runledger.AppendHookSetter); ok {
			setter.SetAppendHook(cpService.OnJournalEvent)
		}
	}

	vals := &provenanceValues{
		checkpointStore:   cpStore,
		checkpointService: cpService,
		sessionTree:       sessionTree,
	}

	return &appinit.ModuleResult{
		Values: map[appinit.Provides]interface{}{
			appinit.ProvidesProvenance: vals,
		},
		CatalogEntries: []appinit.CatalogEntry{
			{
				Category:    "provenance",
				Description: "Session provenance: checkpoints, session tree, attribution",
				ConfigKey:   "provenance.enabled",
				Enabled:     true,
			},
		},
	}, nil
}
