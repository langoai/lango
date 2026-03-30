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
func initOntology(ctx context.Context, client *ent.Client, cfg *config.Config, graphStore graph.Store) (*ontology.ServiceImpl, error) {
	if !cfg.Ontology.Enabled {
		return nil, nil
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
	if cfg.Ontology.ACL.Enabled {
		roles := make(map[string]ontology.Permission, len(cfg.Ontology.ACL.Roles))
		for principal, level := range cfg.Ontology.ACL.Roles {
			roles[principal] = ontology.ParsePermission(level)
		}
		svc.SetACLPolicy(ontology.NewRoleBasedPolicy(roles))
		logger().Infow("ontology ACL enabled", "roles", len(roles))
	}

	return svc, nil
}
