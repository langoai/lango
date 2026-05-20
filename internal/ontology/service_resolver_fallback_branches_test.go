package ontology_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/ontology"
	"github.com/langoai/lango/internal/testutil"
)

func TestServiceResolverFallbackBranches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	graphStore := testutil.NewMockGraphStore()
	require.NoError(t, graphStore.AddTriple(ctx, graph.Triple{
		Subject:   "raw-subject",
		Predicate: "knows",
		Object:    "raw-object",
	}))
	svc := ontology.NewService(newServiceSchemaHealthAndTypeUsageWithoutGovernanceRegistry(), graphStore)

	resolved, err := svc.Resolve(ctx, "raw-subject")
	require.NoError(t, err)
	require.Equal(t, "raw-subject", resolved)

	aliases, err := svc.Aliases(ctx, "raw-subject")
	require.NoError(t, err)
	require.Nil(t, aliases)

	triples, err := svc.QueryTriples(ctx, "raw-subject")
	require.NoError(t, err)
	require.Len(t, triples, 1)
	require.Equal(t, "raw-object", triples[0].Object)
}

func TestServiceVerifyP2PFactUpdatesOnlyUnverifiedTriple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	graphStore := testutil.NewMockGraphStore()
	require.NoError(t, graphStore.AddTriple(ctx, graph.Triple{
		Subject:   "agent:alice",
		Predicate: "status",
		Object:    "ready",
		Metadata:  map[string]string{"_p2p_verified": "false", "source": "peer"},
	}))
	svc := ontology.NewService(newServiceSchemaHealthAndTypeUsageWithoutGovernanceRegistry(), graphStore)

	require.NoError(t, svc.VerifyP2PFact(ctx, "agent:alice", "status", "ready"))

	triples, err := graphStore.QueryBySubject(ctx, "agent:alice")
	require.NoError(t, err)
	require.Len(t, triples, 1)
	require.Equal(t, "true", triples[0].Metadata["_p2p_verified"])
	require.Equal(t, "peer", triples[0].Metadata["source"])
}

func TestServiceVerifyP2PFactErrorBranches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	graphStore := testutil.NewMockGraphStore()
	svc := ontology.NewService(newServiceSchemaHealthAndTypeUsageWithoutGovernanceRegistry(), graphStore)

	err := svc.VerifyP2PFact(ctx, "agent:missing", "status", "ready")
	require.ErrorContains(t, err, "triple not found")

	queryErr := errors.New("graph query failed")
	graphStore.QueryErr = queryErr
	err = svc.VerifyP2PFact(ctx, "agent:missing", "status", "ready")
	require.ErrorIs(t, err, queryErr)
	require.ErrorContains(t, err, "query triples")
}

func TestServiceValidationAndPromotionErrorsPreserveVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	reg := newServiceSchemaHealthAndTypeUsageWithoutGovernanceRegistry()
	reg.predicates["active"] = ontology.PredicateDefinition{Name: "active", Status: ontology.SchemaActive}
	reg.predicates["old"] = ontology.PredicateDefinition{Name: "old", Status: ontology.SchemaDeprecated}
	svc := ontology.NewService(reg, nil)

	require.NoError(t, svc.ValidateTriple(ctx, graph.Triple{Predicate: "active"}))
	err := svc.ValidateTriple(ctx, graph.Triple{Predicate: "old"})
	require.ErrorContains(t, err, `unknown or deprecated predicate "old"`)

	err = svc.PromotePredicate(ctx, "active", ontology.SchemaActive, "missing governance")
	require.ErrorContains(t, err, "governance not enabled")
	version, versionErr := svc.SchemaVersion(ctx)
	require.NoError(t, versionErr)
	require.Zero(t, version)
}
