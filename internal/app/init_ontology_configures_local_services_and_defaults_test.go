package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/ontology"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/testutil"
)

func TestInitOntologyDisabledAndUnavailableStorageBranches(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Ontology.Enabled = false

	disabled, err := initOntology(context.Background(), nil, cfg, nil)
	require.NoError(t, err)
	require.NotNil(t, disabled)
	assert.Nil(t, disabled.Service)
	assert.Nil(t, disabled.Registry)
	assert.Nil(t, disabled.Bridge)

	cfg.Ontology.Enabled = true
	missing, err := initOntology(context.Background(), nil, cfg, nil)
	require.Error(t, err)
	assert.Nil(t, missing)
	assert.EqualError(t, err, "ontology storage unavailable")

	missingRegistry, err := initOntology(context.Background(), &storage.OntologyDeps{}, cfg, nil)
	require.Error(t, err)
	assert.Nil(t, missingRegistry)
	assert.EqualError(t, err, "ontology storage unavailable")
}

func TestInitOntologyConfiguresLocalServicesAndDefaults(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := testutil.TestEntClient(t)
	cfg := config.DefaultConfig()
	cfg.Ontology.Enabled = true
	cfg.Ontology.ACL.Enabled = true
	cfg.Ontology.ACL.Roles = map[string]string{
		"system":   "admin",
		"observer": "read",
	}
	cfg.Ontology.ACL.P2PPermission = "write"
	cfg.Ontology.Governance.Enabled = true
	cfg.Ontology.Governance.MaxNewPerDay = 7
	cfg.Ontology.Governance.QuarantinePeriodHrs = 2
	cfg.Ontology.Governance.ShadowModeDurationHrs = 3
	cfg.Ontology.Governance.MinUsageForPromotion = 4
	cfg.Ontology.Governance.SchemaExplosionBudget = 5
	cfg.Ontology.Exchange.Enabled = true

	graphStore := testutil.NewMockGraphStore()
	result, err := initOntology(ctx, initOntologyConfiguresLocalServicesAndDefaultsDeps(client), cfg, graphStore)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Service)
	require.NotNil(t, result.Registry)
	require.NotNil(t, result.Bridge)

	actions, err := result.Service.ListActions(ctx)
	require.NoError(t, err)
	assert.Len(t, actions, 2)

	types, err := result.Service.ListTypes(ctx)
	require.NoError(t, err)
	assert.Len(t, types, 6)
	predicates, err := result.Service.ListPredicates(ctx)
	require.NoError(t, err)
	assert.Len(t, predicates, 9)

	err = result.Service.StoreTriple(ctx, graph.Triple{
		Subject:   "entity:initOntologyConfiguresLocalServicesAndDefaults",
		Predicate: "related_to",
		Object:    "entity:target",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, graphStore.TripleCount())

	err = result.Service.SetEntityProperty(ctx, "entity:initOntologyConfiguresLocalServicesAndDefaults", "Tool", "name", "search")
	require.NoError(t, err)
	props, err := result.Service.GetEntityProperties(ctx, "entity:initOntologyConfiguresLocalServicesAndDefaults")
	require.NoError(t, err)
	assert.Equal(t, "search", props["name"])

	deniedCtx := ctxkeys.WithPrincipal(ctx, "observer")
	err = result.Service.StoreTriple(deniedCtx, graph.Triple{
		Subject:   "entity:denied",
		Predicate: "related_to",
		Object:    "entity:target",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func initOntologyConfiguresLocalServicesAndDefaultsDeps(client *ent.Client) *storage.OntologyDeps {
	return &storage.OntologyDeps{
		Registry:  ontology.NewEntRegistry(client),
		Conflict:  ontology.NewConflictStore(client),
		Alias:     ontology.NewAliasStore(client),
		Property:  ontology.NewPropertyStore(client),
		ActionLog: ontology.NewActionLogStore(client),
	}
}
