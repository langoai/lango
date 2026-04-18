package app

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/ontology"
	"github.com/langoai/lango/internal/p2p/ontologybridge"
	"github.com/langoai/lango/internal/storage"
)

// initOntology creates the ontology service and seeds default types/predicates.
// Returns nil service if ontology is disabled. Errors are non-fatal (logged, graph continues).
// initOntologyResult holds the ontology service and action registry for tool generation.
type initOntologyResult struct {
	Service  *ontology.ServiceImpl
	Registry *ontology.ActionRegistry
	Bridge   *ontologybridge.Bridge // non-nil when exchange is enabled
}

func initOntology(ctx context.Context, deps *storage.OntologyDeps, cfg *config.Config, graphStore graph.Store) (*initOntologyResult, error) {
	if !cfg.Ontology.Enabled {
		return &initOntologyResult{}, nil
	}
	if deps == nil || deps.Registry == nil {
		return nil, fmt.Errorf("ontology storage unavailable")
	}

	reg := deps.Registry
	svc := ontology.NewService(reg, graphStore)

	if err := ontology.SeedDefaults(ctx, svc); err != nil {
		return nil, fmt.Errorf("seed ontology: %w", err)
	}

	// Truth Maintenance — requires graph store for triple CRUD.
	if graphStore != nil {
		conflictStore := deps.Conflict
		tm := ontology.NewTruthMaintainer(svc, graphStore, conflictStore)
		svc.SetTruthMaintainer(tm)
		logger().Info("truth maintenance initialized")

		// Entity Resolution — requires truth maintainer for Merge retraction.
		aliasStore := deps.Alias
		resolver := ontology.NewEntityResolver(aliasStore, graphStore, tm)
		svc.SetEntityResolver(resolver)
		logger().Info("entity resolution initialized")
	}

	// Property Store — EAV for per-entity property values.
	propStore := deps.Property
	svc.SetPropertyStore(propStore)
	logger().Info("property store initialized")

	// ACL — operation-level access control.
	var acl ontology.ACLPolicy
	if cfg.Ontology.ACL.Enabled {
		roles := make(map[string]ontology.Permission, len(cfg.Ontology.ACL.Roles))
		for principal, level := range cfg.Ontology.ACL.Roles {
			roles[principal] = ontology.ParsePermission(level)
		}
		rbp := ontology.NewRoleBasedPolicy(roles)
		if cfg.Ontology.ACL.P2PPermission != "" {
			rbp.SetP2PPermission(ontology.ParsePermission(cfg.Ontology.ACL.P2PPermission))
		}
		acl = rbp
		svc.SetACLPolicy(acl)
		logger().Infow("ontology ACL enabled", "roles", len(roles))
	}

	// Action Types — registry, built-in actions, executor.
	actionReg := ontology.NewActionRegistry()
	if err := actionReg.Register(ontology.BuiltinLinkEntities()); err != nil {
		return nil, fmt.Errorf("register link_entities: %w", err)
	}
	if err := actionReg.Register(ontology.BuiltinSetEntityStatus()); err != nil {
		return nil, fmt.Errorf("register set_entity_status: %w", err)
	}
	logStore := deps.ActionLog
	executor := ontology.NewActionExecutor(svc, actionReg, acl, logStore)
	svc.SetActionExecutor(executor)
	logger().Infow("action executor initialized", "actions", len(actionReg.List()))

	// Governance — schema lifecycle FSM. Must be set AFTER SeedDefaults
	// so seed types/predicates register with active status.
	if cfg.Ontology.Governance.Enabled {
		gov := ontology.NewGovernanceEngine(ontology.GovernancePolicy{
			MaxNewPerDay:          cfg.Ontology.Governance.MaxNewPerDay,
			QuarantinePeriodHrs:   cfg.Ontology.Governance.QuarantinePeriodHrs,
			ShadowModeDurationHrs: cfg.Ontology.Governance.ShadowModeDurationHrs,
			MinUsageForPromotion:  cfg.Ontology.Governance.MinUsageForPromotion,
			SchemaExplosionBudget: cfg.Ontology.Governance.SchemaExplosionBudget,
		})
		svc.SetGovernanceEngine(gov)
		logger().Infow("ontology governance enabled",
			"maxNewPerDay", cfg.Ontology.Governance.MaxNewPerDay)
	}

	// Ontology Exchange Bridge — P2P schema exchange.
	// Bridge is created here but handler injection happens in modules.go
	// where the P2P protocol handler is accessible.
	var bridge *ontologybridge.Bridge
	if cfg.Ontology.Exchange.Enabled {
		bridgeCfg := ontologybridge.Config{
			MinTrustForSchema: cfg.Ontology.Exchange.MinTrustForSchema,
			MinTrustForFacts:  cfg.Ontology.Exchange.MinTrustForFacts,
			AutoImportMode:    cfg.Ontology.Exchange.AutoImportMode,
			MaxTypesPerImport: cfg.Ontology.Exchange.MaxTypesPerImport,
		}
		if bridgeCfg.MinTrustForSchema == 0 {
			bridgeCfg.MinTrustForSchema = 0.5
		}
		if bridgeCfg.MinTrustForFacts == 0 {
			bridgeCfg.MinTrustForFacts = 0.7
		}
		if bridgeCfg.AutoImportMode == "" {
			bridgeCfg.AutoImportMode = string(ontology.ImportShadow)
		}
		if bridgeCfg.MaxTypesPerImport == 0 {
			bridgeCfg.MaxTypesPerImport = 10
		}
		bridge = ontologybridge.New(svc, nil, bridgeCfg) // reputation store injected later
		logger().Infow("ontology exchange bridge created",
			"autoImportMode", bridgeCfg.AutoImportMode,
			"minTrustForSchema", bridgeCfg.MinTrustForSchema)
	}

	return &initOntologyResult{Service: svc, Registry: actionReg, Bridge: bridge}, nil
}
