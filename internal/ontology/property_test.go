package ontology_test

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/ontology"
	"github.com/langoai/lango/internal/testutil"
)

func newPropertyTestEnv(t *testing.T) (ontology.OntologyService, *testutil.MockGraphStore) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	reg := ontology.NewEntRegistry(client)
	gs := testutil.NewMockGraphStore()
	svc := ontology.NewService(reg, gs)
	require.NoError(t, ontology.SeedDefaults(context.Background(), svc))

	// Wire truth + resolution + property store.
	cs := ontology.NewConflictStore(client)
	tm := ontology.NewTruthMaintainer(svc, gs, cs)
	svc.SetTruthMaintainer(tm)

	as := ontology.NewAliasStore(client)
	er := ontology.NewEntityResolver(as, gs, tm)
	svc.SetEntityResolver(er)

	ps := ontology.NewPropertyStore(client)
	svc.SetPropertyStore(ps)

	return svc, gs
}

// --- PropertyStore CRUD Tests ---

func TestPropertyStore_SetAndGet(t *testing.T) {
	svc, _ := newPropertyTestEnv(t)
	ctx := context.Background()

	err := svc.SetEntityProperty(ctx, "error:timeout", "ErrorPattern", "tool_name", "http_client")
	require.NoError(t, err)

	props, err := svc.GetEntityProperties(ctx, "error:timeout")
	require.NoError(t, err)
	assert.Equal(t, "http_client", props["tool_name"])
}

func TestPropertyStore_Upsert(t *testing.T) {
	svc, _ := newPropertyTestEnv(t)
	ctx := context.Background()

	err := svc.SetEntityProperty(ctx, "error:timeout", "ErrorPattern", "tool_name", "http_v1")
	require.NoError(t, err)

	err = svc.SetEntityProperty(ctx, "error:timeout", "ErrorPattern", "tool_name", "http_v2")
	require.NoError(t, err)

	props, err := svc.GetEntityProperties(ctx, "error:timeout")
	require.NoError(t, err)
	assert.Equal(t, "http_v2", props["tool_name"])
}

func TestPropertyStore_MultipleProperties(t *testing.T) {
	svc, _ := newPropertyTestEnv(t)
	ctx := context.Background()

	require.NoError(t, svc.SetEntityProperty(ctx, "error:timeout", "ErrorPattern", "tool_name", "http"))
	require.NoError(t, svc.SetEntityProperty(ctx, "error:timeout", "ErrorPattern", "pattern", "connection timeout"))

	props, err := svc.GetEntityProperties(ctx, "error:timeout")
	require.NoError(t, err)
	assert.Len(t, props, 2)
	assert.Equal(t, "http", props["tool_name"])
	assert.Equal(t, "connection timeout", props["pattern"])
}

func TestPropertyStore_GetEmpty(t *testing.T) {
	svc, _ := newPropertyTestEnv(t)
	ctx := context.Background()

	props, err := svc.GetEntityProperties(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Empty(t, props)
}

// --- Query Tests ---

func TestPropertyStore_QueryEq(t *testing.T) {
	svc, _ := newPropertyTestEnv(t)
	ctx := context.Background()

	require.NoError(t, svc.SetEntityProperty(ctx, "err:1", "ErrorPattern", "tool_name", "http_client"))
	require.NoError(t, svc.SetEntityProperty(ctx, "err:2", "ErrorPattern", "tool_name", "db_client"))
	require.NoError(t, svc.SetEntityProperty(ctx, "err:3", "ErrorPattern", "tool_name", "http_client"))

	results, err := svc.QueryEntities(ctx, ontology.PropertyQuery{
		EntityType: "ErrorPattern",
		Filters: []ontology.PropertyFilter{
			{Property: "tool_name", Op: ontology.FilterEq, Value: "http_client"},
		},
	})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestPropertyStore_QueryNeq(t *testing.T) {
	svc, _ := newPropertyTestEnv(t)
	ctx := context.Background()

	require.NoError(t, svc.SetEntityProperty(ctx, "err:1", "ErrorPattern", "tool_name", "http_client"))
	require.NoError(t, svc.SetEntityProperty(ctx, "err:2", "ErrorPattern", "tool_name", "db_client"))

	results, err := svc.QueryEntities(ctx, ontology.PropertyQuery{
		EntityType: "ErrorPattern",
		Filters: []ontology.PropertyFilter{
			{Property: "tool_name", Op: ontology.FilterNeq, Value: "http_client"},
		},
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "db_client", results[0].Properties["tool_name"])
}

func TestPropertyStore_QueryContains(t *testing.T) {
	svc, _ := newPropertyTestEnv(t)
	ctx := context.Background()

	require.NoError(t, svc.SetEntityProperty(ctx, "err:1", "ErrorPattern", "tool_name", "http_client_v2"))
	require.NoError(t, svc.SetEntityProperty(ctx, "err:2", "ErrorPattern", "tool_name", "db_client"))

	results, err := svc.QueryEntities(ctx, ontology.PropertyQuery{
		EntityType: "ErrorPattern",
		Filters: []ontology.PropertyFilter{
			{Property: "tool_name", Op: ontology.FilterContains, Value: "http"},
		},
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "err:1", results[0].EntityID)
}

func TestPropertyStore_QueryMultipleFilters(t *testing.T) {
	svc, _ := newPropertyTestEnv(t)
	ctx := context.Background()

	require.NoError(t, svc.SetEntityProperty(ctx, "err:1", "ErrorPattern", "tool_name", "http"))
	require.NoError(t, svc.SetEntityProperty(ctx, "err:1", "ErrorPattern", "pattern", "timeout"))
	require.NoError(t, svc.SetEntityProperty(ctx, "err:2", "ErrorPattern", "tool_name", "http"))
	require.NoError(t, svc.SetEntityProperty(ctx, "err:2", "ErrorPattern", "pattern", "connection refused"))

	results, err := svc.QueryEntities(ctx, ontology.PropertyQuery{
		EntityType: "ErrorPattern",
		Filters: []ontology.PropertyFilter{
			{Property: "tool_name", Op: ontology.FilterEq, Value: "http"},
			{Property: "pattern", Op: ontology.FilterContains, Value: "timeout"},
		},
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "err:1", results[0].EntityID)
}

func TestPropertyStore_QueryNoResults(t *testing.T) {
	svc, _ := newPropertyTestEnv(t)
	ctx := context.Background()

	results, err := svc.QueryEntities(ctx, ontology.PropertyQuery{
		EntityType: "ErrorPattern",
		Filters: []ontology.PropertyFilter{
			{Property: "tool_name", Op: ontology.FilterEq, Value: "nonexistent"},
		},
	})
	require.NoError(t, err)
	assert.Empty(t, results)
}

// --- Validation Tests ---

func TestSetEntityProperty_UnknownProperty(t *testing.T) {
	svc, _ := newPropertyTestEnv(t)
	ctx := context.Background()

	err := svc.SetEntityProperty(ctx, "err:1", "ErrorPattern", "nonexistent_prop", "val")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not defined in type")
}

func TestSetEntityProperty_UnknownType(t *testing.T) {
	svc, _ := newPropertyTestEnv(t)
	ctx := context.Background()

	err := svc.SetEntityProperty(ctx, "x", "UnknownType", "prop", "val")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown type")
}

// --- Alias-Aware Tests ---

func TestSetEntityProperty_AliasAware(t *testing.T) {
	svc, _ := newPropertyTestEnv(t)
	ctx := context.Background()

	// Create alias.
	err := svc.DeclareSameAs(ctx, "error:api_timeout", "error:timeout", "manual")
	require.NoError(t, err)

	// Set property via alias.
	err = svc.SetEntityProperty(ctx, "error:api_timeout", "ErrorPattern", "tool_name", "http")
	require.NoError(t, err)

	// Get via canonical should work.
	props, err := svc.GetEntityProperties(ctx, "error:timeout")
	require.NoError(t, err)
	assert.Equal(t, "http", props["tool_name"])

	// Get via alias should also work.
	props, err = svc.GetEntityProperties(ctx, "error:api_timeout")
	require.NoError(t, err)
	assert.Equal(t, "http", props["tool_name"])
}

func TestGetEntity_AliasAware(t *testing.T) {
	svc, gs := newPropertyTestEnv(t)
	ctx := context.Background()

	// Set property on canonical.
	require.NoError(t, svc.SetEntityProperty(ctx, "error:timeout", "ErrorPattern", "tool_name", "http"))

	// Add outgoing and incoming triples.
	require.NoError(t, gs.AddTriple(ctx, graph.Triple{
		Subject: "error:timeout", Predicate: graph.CausedBy, Object: "tool:http",
	}))
	require.NoError(t, gs.AddTriple(ctx, graph.Triple{
		Subject: "fix:retry", Predicate: graph.ResolvedBy, Object: "error:timeout",
	}))

	// Create alias.
	require.NoError(t, svc.DeclareSameAs(ctx, "error:api_timeout", "error:timeout", "manual"))

	// GetEntity via alias — should return properties + outgoing + incoming.
	entity, err := svc.GetEntity(ctx, "error:api_timeout")
	require.NoError(t, err)
	assert.Equal(t, "error:timeout", entity.EntityID)
	assert.Equal(t, "ErrorPattern", entity.EntityType)
	assert.Equal(t, "http", entity.Properties["tool_name"])
	assert.Len(t, entity.Outgoing, 1, "should have outgoing triple")
	assert.Equal(t, graph.CausedBy, entity.Outgoing[0].Predicate)
	assert.Len(t, entity.Incoming, 1, "should have incoming triple")
	assert.Equal(t, graph.ResolvedBy, entity.Incoming[0].Predicate)
}

// --- GetEntity Tests ---

func TestGetEntity_WithProperties(t *testing.T) {
	svc, _ := newPropertyTestEnv(t)
	ctx := context.Background()

	require.NoError(t, svc.SetEntityProperty(ctx, "tool:http", "Tool", "name", "HTTP Client"))

	entity, err := svc.GetEntity(ctx, "tool:http")
	require.NoError(t, err)
	assert.Equal(t, "tool:http", entity.EntityID)
	assert.Equal(t, "Tool", entity.EntityType)
	assert.Equal(t, "HTTP Client", entity.Properties["name"])
}

func TestGetEntity_NoProperties(t *testing.T) {
	svc, _ := newPropertyTestEnv(t)
	ctx := context.Background()

	entity, err := svc.GetEntity(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, "nonexistent", entity.EntityID)
	assert.Empty(t, entity.Properties)
}

func TestGetEntity_OutgoingAndIncoming(t *testing.T) {
	svc, gs := newPropertyTestEnv(t)
	ctx := context.Background()

	require.NoError(t, svc.SetEntityProperty(ctx, "tool:http", "Tool", "name", "HTTP Client"))

	// Outgoing: tool:http → ...
	require.NoError(t, gs.AddTriple(ctx, graph.Triple{
		Subject: "tool:http", Predicate: graph.RelatedTo, Object: "tool:grpc",
	}))
	// Incoming: ... → tool:http
	require.NoError(t, gs.AddTriple(ctx, graph.Triple{
		Subject: "error:timeout", Predicate: graph.CausedBy, Object: "tool:http",
	}))

	entity, err := svc.GetEntity(ctx, "tool:http")
	require.NoError(t, err)
	assert.Len(t, entity.Outgoing, 1, "should have 1 outgoing triple")
	assert.Len(t, entity.Incoming, 1, "should have 1 incoming triple")
	assert.Equal(t, "tool:grpc", entity.Outgoing[0].Object)
	assert.Equal(t, "error:timeout", entity.Incoming[0].Subject)
}

func TestQueryEntities_WithGraphJoin(t *testing.T) {
	svc, gs := newPropertyTestEnv(t)
	ctx := context.Background()

	require.NoError(t, svc.SetEntityProperty(ctx, "err:1", "ErrorPattern", "tool_name", "http"))

	// Add outgoing triple for the entity.
	require.NoError(t, gs.AddTriple(ctx, graph.Triple{
		Subject: "err:1", Predicate: graph.CausedBy, Object: "tool:http",
	}))

	results, err := svc.QueryEntities(ctx, ontology.PropertyQuery{
		EntityType: "ErrorPattern",
		Filters: []ontology.PropertyFilter{
			{Property: "tool_name", Op: ontology.FilterEq, Value: "http"},
		},
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Len(t, results[0].Outgoing, 1, "QueryEntities should include outgoing triples")
	assert.Equal(t, graph.CausedBy, results[0].Outgoing[0].Predicate)
}

func TestDeleteEntityProperties(t *testing.T) {
	svc, _ := newPropertyTestEnv(t)
	ctx := context.Background()

	require.NoError(t, svc.SetEntityProperty(ctx, "err:1", "ErrorPattern", "tool_name", "http"))
	require.NoError(t, svc.SetEntityProperty(ctx, "err:1", "ErrorPattern", "pattern", "timeout"))

	// Delete all properties.
	err := svc.DeleteEntityProperties(ctx, "err:1")
	require.NoError(t, err)

	// Properties should be empty.
	props, err := svc.GetEntityProperties(ctx, "err:1")
	require.NoError(t, err)
	assert.Empty(t, props)
}

func TestPropertyStore_NotInitialized(t *testing.T) {
	svc := newTestService(t) // no property store set
	ctx := context.Background()

	_, err := svc.GetEntityProperties(ctx, "x")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "property store not initialized")
}
