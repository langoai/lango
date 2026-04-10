package app

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/observability/token"
	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/provenance"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/runledger"
	"github.com/langoai/lango/internal/toolchain"
)

// provenanceValues holds the outputs of the provenance module.
type provenanceValues struct {
	checkpointStore   provenance.CheckpointStore
	checkpointService *provenance.CheckpointService
	sessionTreeStore  provenance.SessionTreeStore
	sessionTree       *provenance.SessionTree
	attributionStore  provenance.AttributionStore
	attribution       *provenance.AttributionService
	bundle            *provenance.BundleService

	// configMetadata is the cached metadata for session config checkpoints.
	// Contains "config_fingerprint" and "hook_registry" keys.
	configMetadata map[string]string
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
	treeStore := provenance.SessionTreeStore(provenance.NewMemoryTreeStore())
	attrStore := provenance.AttributionStore(provenance.NewMemoryAttributionStore())
	var tokenStore provenance.TokenUsageReader
	if m.boot != nil && m.boot.DBClient != nil {
		cpStore = provenance.NewEntCheckpointStore(m.boot.DBClient)
		treeStore = provenance.NewEntSessionTreeStore(m.boot.DBClient)
		attrStore = provenance.NewEntAttributionStore(m.boot.DBClient)
		tokenStore = token.NewEntTokenStore(m.boot.DBClient)
	}

	// Resolve RunLedger store if available.
	var ledgerStore runledger.RunLedgerStore
	if rlVals := r.Resolve(appinit.ProvidesRunLedger); rlVals != nil {
		if rv, ok := rlVals.(*runLedgerValues); ok {
			ledgerStore = rv.store
		}
	}

	cpService := provenance.NewCheckpointService(cpStore, ledgerStore, m.cfg.Provenance.Checkpoints)
	sessionTree := provenance.NewSessionTree(treeStore)
	attribution := provenance.NewAttributionService(attrStore, cpStore, tokenStore)
	ed25519Verifier := func(didStr string, payload, signature []byte) error {
		pubkey, err := identity.ParseDIDPublicKey(didStr)
		if err != nil {
			return err
		}
		return security.VerifyEd25519(pubkey, payload, signature)
	}
	verifiers := map[string]provenance.SignatureVerifyFunc{
		security.AlgorithmSecp256k1Keccak256: identity.VerifyMessageSignature,
		security.AlgorithmEd25519:            ed25519Verifier,
	}
	bundle := provenance.NewBundleService(cpStore, treeStore, attrStore, attribution, verifiers)

	// Register auto-checkpoint hook on the RunLedger store (post-construction).
	if ledgerStore != nil {
		if setter, ok := ledgerStore.(runledger.AppendHookSetter); ok {
			setter.SetAppendHook(cpService.OnJournalEvent)
		}
	}

	vals := &provenanceValues{
		checkpointStore:   cpStore,
		checkpointService: cpService,
		sessionTreeStore:  treeStore,
		sessionTree:       sessionTree,
		attributionStore:  attrStore,
		attribution:       attribution,
		bundle:            bundle,
	}

	// Compute and cache config fingerprint + hook snapshot for session checkpoints.
	vals.configMetadata = computeConfigMetadata(m.boot, nil)

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

// hookEntry describes a registered hook for snapshot serialization.
type hookEntry struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

// computeConfigFingerprint computes a SHA-256 hex digest of the config state
// relevant to session reproducibility: ExplicitKeys, AutoEnabled, and HooksConfig.
func computeConfigFingerprint(boot *bootstrap.Result) string {
	h := sha256.New()

	// json.Marshal sorts map keys deterministically (Go 1.12+).
	if boot.ExplicitKeys != nil {
		data, _ := json.Marshal(boot.ExplicitKeys)
		h.Write(data)
	}

	autoData, _ := json.Marshal(boot.AutoEnabled)
	h.Write(autoData)

	hooksData, _ := json.Marshal(boot.Config.Hooks)
	h.Write(hooksData)

	return fmt.Sprintf("%x", h.Sum(nil))
}

// buildHookRegistrySnapshot serializes the current hook registry state as a
// JSON array of {name, priority} objects.
func buildHookRegistrySnapshot(registry *toolchain.HookRegistry) string {
	if registry == nil {
		return "[]"
	}

	var entries []hookEntry
	for _, h := range registry.PreHooks() {
		entries = append(entries, hookEntry{Name: h.Name(), Priority: h.Priority()})
	}
	for _, h := range registry.PostHooks() {
		entries = append(entries, hookEntry{Name: h.Name(), Priority: h.Priority()})
	}
	if entries == nil {
		return "[]"
	}

	data, _ := json.Marshal(entries)
	return string(data)
}

// computeConfigMetadata builds the metadata map for session config checkpoints.
// hookRegistry may be nil during module init (hooks are registered later in Phase B).
func computeConfigMetadata(boot *bootstrap.Result, registry *toolchain.HookRegistry) map[string]string {
	if boot == nil {
		return nil
	}
	return map[string]string{
		"config_fingerprint": computeConfigFingerprint(boot),
		"hook_registry":      buildHookRegistrySnapshot(registry),
	}
}
