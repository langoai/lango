package app

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/ontology"
)

// initOntology creates the ontology service and seeds default types/predicates.
// Returns nil service if ontology is disabled. Errors are non-fatal (logged, graph continues).
// initOntologyResult holds the ontology service and action registry for tool generation.
type initOntologyResult struct {
	Service  *ontology.ServiceImpl
	Registry *ontology.ActionRegistry
}

func initOntology(ctx context.Context, client *ent.Client, cfg *config.Config, graphStore graph.Store) (*initOntologyResult, error) {
	if !cfg.Ontology.Enabled {
		return &initOntologyResult{}, nil
	}

	reg := ontology.NewEntRegistry(client)
	svc := ontology.NewService(reg, graphStore)

	if err := ontology.SeedDefaults(ctx, svc); err != nil {
		return nil, fmt.Errorf("seed ontology: %w", err)
	}

	// Truth Maintenance — requires graph store for triple CRUD.
	if graphStore != nil {
		conflictStore := ontology.NewConflictStore(client)
		tm := ontology.NewTruthMaintainer(svc, graphStore, conflictStore)
		svc.SetTruthMaintainer(tm)
		logger().Info("truth maintenance initialized")

		// Entity Resolution — requires truth maintainer for Merge retraction.
		aliasStore := ontology.NewAliasStore(client)
		resolver := ontology.NewEntityResolver(aliasStore, graphStore, tm)
		svc.SetEntityResolver(resolver)
		logger().Info("entity resolution initialized")
	}

	// Property Store — EAV for per-entity property values.
	propStore := ontology.NewPropertyStore(client)
	svc.SetPropertyStore(propStore)
	logger().Info("property store initialized")

	// ACL — operation-level access control.
	var acl ontology.ACLPolicy
	if cfg.Ontology.ACL.Enabled {
		roles := make(map[string]ontology.Permission, len(cfg.Ontology.ACL.Roles))
		for principal, level := range cfg.Ontology.ACL.Roles {
			roles[principal] = ontology.ParsePermission(level)
		}
		acl = ontology.NewRoleBasedPolicy(roles)
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
	logStore := ontology.NewActionLogStore(client)
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

	return &initOntologyResult{Service: svc, Registry: actionReg}, nil
}
