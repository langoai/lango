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

	return svc, nil
}
