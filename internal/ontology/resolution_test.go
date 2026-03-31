package ontology_test

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/ontology"
	"github.com/langoai/lango/internal/testutil"
)

// newResolutionTestEnv creates a test environment with truth maintenance + entity resolution.
func newResolutionTestEnv(t *testing.T) (ontology.OntologyService, *testutil.MockGraphStore) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	reg := ontology.NewEntRegistry(client)
	gs := testutil.NewMockGraphStore()
	svc := ontology.NewService(reg, gs)

	require.NoError(t, ontology.SeedDefaults(context.Background(), svc))

	// Wire truth maintenance.
	cs := ontology.NewConflictStore(client)
	tm := ontology.NewTruthMaintainer(svc, gs, cs)
	svc.SetTruthMaintainer(tm)

	// Wire entity resolution.
	as := ontology.NewAliasStore(client)
	er := ontology.NewEntityResolver(as, gs, tm)
	svc.SetEntityResolver(er)

	return svc, gs
}

func TestEntityResolver_Resolve_NoAlias(t *testing.T) {
	svc, _ := newResolutionTestEnv(t)
	ctx := context.Background()

	canonical, err := svc.Resolve(ctx, "error:timeout")
	require.NoError(t, err)
	assert.Equal(t, "error:timeout", canonical, "no alias = identity")
}

func TestEntityResolver_RegisterAlias_And_Resolve(t *testing.T) {
	svc, _ := newResolutionTestEnv(t)
	ctx := context.Background()

	err := svc.DeclareSameAs(ctx, "error:api_timeout", "error:timeout", "manual")
	require.NoError(t, err)

	canonical, err := svc.Resolve(ctx, "error:api_timeout")
	require.NoError(t, err)
	assert.Equal(t, "error:timeout", canonical, "should resolve to canonical")

	// Original should still resolve to itself.
	canonical, err = svc.Resolve(ctx, "error:timeout")
	require.NoError(t, err)
	assert.Equal(t, "error:timeout", canonical)
}

func TestEntityResolver_DeclareSameAs(t *testing.T) {
	svc, _ := newResolutionTestEnv(t)
	ctx := context.Background()

	// Convention: second arg (node_b) is canonical.
	err := svc.DeclareSameAs(ctx, "node_a", "node_b", "manual")
	require.NoError(t, err)

	// node_a should resolve to node_b (canonical).
	canonA, err := svc.Resolve(ctx, "node_a")
	require.NoError(t, err)
	assert.Equal(t, "node_b", canonA, "first arg should resolve to second arg (canonical)")

	// node_b should resolve to itself (already canonical).
	canonB, err := svc.Resolve(ctx, "node_b")
	require.NoError(t, err)
	assert.Equal(t, "node_b", canonB, "canonical should resolve to itself")
}

func TestEntityResolver_DeclareSameAs_AlreadySame(t *testing.T) {
	svc, _ := newResolutionTestEnv(t)
	ctx := context.Background()

	// First declaration.
	err := svc.DeclareSameAs(ctx, "a", "b", "manual")
	require.NoError(t, err)

	// Second declaration of same pair — should be no-op.
	err = svc.DeclareSameAs(ctx, "a", "b", "manual")
	require.NoError(t, err)
}

func TestEntityResolver_Merge_TriplesUpdated(t *testing.T) {
	svc, gs := newResolutionTestEnv(t)
	ctx := context.Background()

	// Set up triples for the duplicate entity.
	require.NoError(t, gs.AddTriple(ctx, graph.Triple{
		Subject: "error:api_timeout", Predicate: graph.CausedBy, Object: "tool:http",
		Metadata: map[string]string{
			ontology.MetaSource:    "graph_engine",
			ontology.MetaValidFrom: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		},
	}))
	require.NoError(t, gs.AddTriple(ctx, graph.Triple{
		Subject: "fix:retry", Predicate: graph.ResolvedBy, Object: "error:api_timeout",
		Metadata: map[string]string{
			ontology.MetaSource:    "graph_engine",
			ontology.MetaValidFrom: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		},
	}))

	// Merge duplicate into canonical.
	result, err := svc.MergeEntities(ctx, "error:timeout", "error:api_timeout")
	require.NoError(t, err)
	assert.Equal(t, 2, result.TriplesUpdated, "should move outgoing + incoming")
	assert.Equal(t, 1, result.AliasesCreated)

	// Verify canonical has the replicated triples.
	triples, err := gs.QueryBySubject(ctx, "error:timeout")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(triples), 1, "canonical should have outgoing triples")
}

func TestEntityResolver_Merge_RetractFact(t *testing.T) {
	svc, gs := newResolutionTestEnv(t)
	ctx := context.Background()

	// Set up a triple with temporal metadata (so RetractFact can find it).
	require.NoError(t, gs.AddTriple(ctx, graph.Triple{
		Subject: "dup:x", Predicate: graph.RelatedTo, Object: "other:y",
		Metadata: map[string]string{
			ontology.MetaSource:    "graph_engine",
			ontology.MetaValidFrom: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		},
	}))

	_, err := svc.MergeEntities(ctx, "canon:x", "dup:x")
	require.NoError(t, err)

	// After merge, original triple should have ValidTo set (retracted).
	triples, err := gs.QueryBySubject(ctx, "dup:x")
	require.NoError(t, err)
	for _, tr := range triples {
		if tr.Predicate == graph.RelatedTo && tr.Object == "other:y" {
			assert.NotEmpty(t, tr.Metadata[ontology.MetaValidTo], "original should be retracted")
		}
	}
}

func TestEntityResolver_Split(t *testing.T) {
	svc, _ := newResolutionTestEnv(t)
	ctx := context.Background()

	// Create alias.
	err := svc.DeclareSameAs(ctx, "split_me", "canonical", "manual")
	require.NoError(t, err)

	// Verify alias exists.
	canon, _ := svc.Resolve(ctx, "split_me")
	assert.Equal(t, "canonical", canon)

	// Split.
	err = svc.SplitEntity(ctx, "canonical", "split_me")
	require.NoError(t, err)

	// After split, alias should be gone.
	canon, _ = svc.Resolve(ctx, "split_me")
	assert.Equal(t, "split_me", canon, "after split, should resolve to self")
}

func TestEntityResolver_Aliases(t *testing.T) {
	svc, _ := newResolutionTestEnv(t)
	ctx := context.Background()

	err := svc.DeclareSameAs(ctx, "alias1", "canonical", "manual")
	require.NoError(t, err)
	err = svc.DeclareSameAs(ctx, "alias2", "canonical", "manual")
	require.NoError(t, err)

	aliases, err := svc.Aliases(ctx, "canonical")
	require.NoError(t, err)
	assert.Len(t, aliases, 2)
	assert.Contains(t, aliases, "alias1")
	assert.Contains(t, aliases, "alias2")
}

func TestStoreTriple_WithResolver(t *testing.T) {
	svc, gs := newResolutionTestEnv(t)
	ctx := context.Background()

	// Register alias.
	err := svc.DeclareSameAs(ctx, "error:api_timeout", "error:timeout", "manual")
	require.NoError(t, err)

	// StoreTriple with the aliased name — should be stored with canonical.
	err = svc.StoreTriple(ctx, graph.Triple{
		Subject: "error:api_timeout", Predicate: graph.CausedBy, Object: "tool:http",
	})
	require.NoError(t, err)

	// Query by canonical — should find the triple.
	triples, err := gs.QueryBySubject(ctx, "error:timeout")
	require.NoError(t, err)
	assert.Len(t, triples, 1, "StoreTriple should canonicalize subject before storing")
	assert.Equal(t, graph.CausedBy, triples[0].Predicate)
}

func TestStoreTriple_WithResolver_ObjectCanonicalization(t *testing.T) {
	svc, gs := newResolutionTestEnv(t)
	ctx := context.Background()

	// Register alias for the object side.
	err := svc.DeclareSameAs(ctx, "tool:http_client", "tool:http", "manual")
	require.NoError(t, err)

	// StoreTriple with aliased object — should be stored with canonical object.
	err = svc.StoreTriple(ctx, graph.Triple{
		Subject: "error:timeout", Predicate: graph.CausedBy, Object: "tool:http_client",
	})
	require.NoError(t, err)

	// Query — object should be canonical.
	triples, err := gs.QueryBySubject(ctx, "error:timeout")
	require.NoError(t, err)
	require.Len(t, triples, 1)
	assert.Equal(t, "tool:http", triples[0].Object, "StoreTriple should canonicalize object before storing")
}

func TestQueryTriples_WithResolver(t *testing.T) {
	svc, gs := newResolutionTestEnv(t)
	ctx := context.Background()

	// Store a triple under canonical name.
	require.NoError(t, gs.AddTriple(ctx, graph.Triple{
		Subject: "error:timeout", Predicate: graph.CausedBy, Object: "tool:http",
	}))

	// Register alias.
	err := svc.DeclareSameAs(ctx, "error:api_timeout", "error:timeout", "manual")
	require.NoError(t, err)

	// QueryTriples with alias — should find canonical's triples.
	triples, err := svc.QueryTriples(ctx, "error:api_timeout")
	require.NoError(t, err)
	assert.Len(t, triples, 1, "QueryTriples should resolve alias before querying")
}

func TestEntityResolver_Merge_TransitiveAliases(t *testing.T) {
	svc, _ := newResolutionTestEnv(t)
	ctx := context.Background()

	// Create alias chain: old_name → intermediate.
	err := svc.DeclareSameAs(ctx, "old_name", "intermediate", "manual")
	require.NoError(t, err)

	// Now merge intermediate into final_canonical.
	_, err = svc.MergeEntities(ctx, "final_canonical", "intermediate")
	require.NoError(t, err)

	// old_name should now resolve to final_canonical (not intermediate).
	canon, err := svc.Resolve(ctx, "old_name")
	require.NoError(t, err)
	assert.Equal(t, "final_canonical", canon, "transitive alias should be updated to new canonical")

	// intermediate should also resolve to final_canonical.
	canon, err = svc.Resolve(ctx, "intermediate")
	require.NoError(t, err)
	assert.Equal(t, "final_canonical", canon)
}

func TestAliasStore_CRUD(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	as := ontology.NewAliasStore(client)

	// Resolve unknown.
	id, err := as.Resolve(ctx, "unknown")
	require.NoError(t, err)
	assert.Equal(t, "unknown", id)

	// Register.
	err = as.Register(ctx, "raw1", "canon1", "manual")
	require.NoError(t, err)

	// Resolve.
	id, err = as.Resolve(ctx, "raw1")
	require.NoError(t, err)
	assert.Equal(t, "canon1", id)

	// Update (re-register with different canonical).
	err = as.Register(ctx, "raw1", "canon2", "correction")
	require.NoError(t, err)
	id, err = as.Resolve(ctx, "raw1")
	require.NoError(t, err)
	assert.Equal(t, "canon2", id)

	// ListByCanonical.
	err = as.Register(ctx, "raw2", "canon2", "manual")
	require.NoError(t, err)
	aliases, err := as.ListByCanonical(ctx, "canon2")
	require.NoError(t, err)
	assert.Len(t, aliases, 2)

	// Remove.
	err = as.Remove(ctx, "raw1")
	require.NoError(t, err)
	id, err = as.Resolve(ctx, "raw1")
	require.NoError(t, err)
	assert.Equal(t, "raw1", id, "after remove, should resolve to self")
}
