package ontology_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/ontology"
	"github.com/langoai/lango/internal/testutil"
)

func TestWave52ServicePredicateCacheAndVersionTrackPredicateMutations(t *testing.T) {
	reg := newWave26Registry()
	reg.predicates["active"] = ontology.PredicateDefinition{
		Name:   "active",
		Status: ontology.SchemaActive,
	}
	reg.predicates["shadow"] = ontology.PredicateDefinition{
		Name:   "shadow",
		Status: ontology.SchemaShadow,
	}
	reg.predicates["deprecated"] = ontology.PredicateDefinition{
		Name:   "deprecated",
		Status: ontology.SchemaDeprecated,
	}
	svc := ontology.NewService(reg, nil)
	ctx := context.Background()

	validator := svc.PredicateValidator()
	assert.True(t, validator("active"))
	assert.True(t, validator("shadow"))
	assert.False(t, validator("deprecated"))

	require.NoError(t, svc.RegisterPredicate(ctx, ontology.PredicateDefinition{
		Name:   "new_active",
		Status: ontology.SchemaActive,
	}))
	version, err := svc.SchemaVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, version)
	assert.True(t, validator("new_active"))

	require.NoError(t, svc.DeprecatePredicate(ctx, "new_active"))
	version, err = svc.SchemaVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, version)
	assert.False(t, validator("new_active"))
}

func TestWave52ServiceGovernedRegistrationProposesAndRateLimitsBeforeMutation(t *testing.T) {
	reg := newWave26Registry()
	svc := ontology.NewService(reg, nil)
	svc.SetGovernanceEngine(ontology.NewGovernanceEngine(ontology.GovernancePolicy{
		MaxNewPerDay: 1,
	}))
	ctx := context.Background()

	require.NoError(t, svc.RegisterType(ctx, ontology.ObjectType{
		Name:   "AgentCandidate",
		Status: ontology.SchemaActive,
	}))
	stored, err := reg.GetType(ctx, "AgentCandidate")
	require.NoError(t, err)
	assert.Equal(t, ontology.SchemaProposed, stored.Status)

	err = svc.RegisterPredicate(ctx, ontology.PredicateDefinition{
		Name:   "blocked_by_rate_limit",
		Status: ontology.SchemaActive,
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "daily schema proposal limit reached")
	assert.Zero(t, reg.calls["RegisterPredicate"])

	version, err := svc.SchemaVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, version)
	assert.False(t, svc.PredicateValidator()("blocked_by_rate_limit"))
}

func TestWave52ServiceGraphDelegationCanonicalizesOnlySuccessfulResolutions(t *testing.T) {
	store := testutil.NewMockGraphStore()
	reg := newWave26Registry()
	svc := ontology.NewService(reg, store)
	svc.SetEntityResolver(wave52Resolver{
		resolved: map[string]string{
			"raw-subject": "canonical-subject",
			"raw-query":   "canonical-query",
		},
		errs: map[string]error{
			"raw-object": errors.New("resolver miss"),
		},
	})
	ctx := context.Background()

	require.NoError(t, svc.StoreTriple(ctx, graph.Triple{
		Subject:   "raw-subject",
		Predicate: "uses",
		Object:    "raw-object",
	}))
	triples, err := store.QueryBySubject(ctx, "canonical-subject")
	require.NoError(t, err)
	require.Len(t, triples, 1)
	assert.Equal(t, "raw-object", triples[0].Object)

	require.NoError(t, store.AddTriple(ctx, graph.Triple{
		Subject:   "canonical-query",
		Predicate: "observes",
		Object:    "result",
	}))
	triples, err = svc.QueryTriples(ctx, "raw-query")
	require.NoError(t, err)
	require.Len(t, triples, 1)
	assert.Equal(t, "canonical-query", triples[0].Subject)
}

func TestWave52ServiceVerifyP2PFactWrapsGraphMutationErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("remove error", func(t *testing.T) {
		store := &wave52GraphStore{
			bySubject: map[string][]graph.Triple{"peer:subject": {newWave52UnverifiedP2PTriple()}},
			removeErr: errors.New("remove failed"),
		}
		svc := ontology.NewService(newWave26Registry(), store)

		err := svc.VerifyP2PFact(ctx, "peer:subject", "reports", "peer:object")
		require.Error(t, err)
		assert.ErrorContains(t, err, "remove unverified triple")
		assert.ErrorContains(t, err, "remove failed")
		require.Len(t, store.removed, 1)
		assert.Equal(t, newWave52VerifiedRemovalTriple(), store.removed[0])
		assert.Zero(t, store.addCalls)
	})

	t.Run("add error", func(t *testing.T) {
		store := &wave52GraphStore{
			bySubject: map[string][]graph.Triple{"peer:subject": {newWave52UnverifiedP2PTriple()}},
			addErr:    errors.New("add failed"),
		}
		svc := ontology.NewService(newWave26Registry(), store)

		err := svc.VerifyP2PFact(ctx, "peer:subject", "reports", "peer:object")
		require.Error(t, err)
		assert.ErrorContains(t, err, "add failed")
		assert.Equal(t, 1, store.removeCalls)
		assert.Equal(t, 1, store.addCalls)
		require.Len(t, store.removed, 1)
		assert.Equal(t, newWave52VerifiedRemovalTriple(), store.removed[0])
		require.Len(t, store.added, 1)
		assert.Equal(t, "true", store.added[0].Metadata["_p2p_verified"])
	})
}

func newWave52UnverifiedP2PTriple() graph.Triple {
	return graph.Triple{
		Subject:   "peer:subject",
		Predicate: "reports",
		Object:    "peer:object",
		Metadata:  map[string]string{"_p2p_verified": "false"},
	}
}

func newWave52VerifiedRemovalTriple() graph.Triple {
	return graph.Triple{
		Subject:   "peer:subject",
		Predicate: "reports",
		Object:    "peer:object",
	}
}

type wave52Resolver struct {
	resolved map[string]string
	errs     map[string]error
}

func (r wave52Resolver) Resolve(_ context.Context, rawID string) (string, error) {
	if err := r.errs[rawID]; err != nil {
		return "", err
	}
	if canonical, ok := r.resolved[rawID]; ok {
		return canonical, nil
	}
	return rawID, nil
}

func (wave52Resolver) RegisterAlias(context.Context, string, string, string) error { return nil }

func (wave52Resolver) DeclareSameAs(context.Context, string, string, string) error { return nil }

func (wave52Resolver) Merge(context.Context, string, string) (*ontology.MergeResult, error) {
	return nil, nil
}

func (wave52Resolver) Split(context.Context, string, string) error { return nil }

func (wave52Resolver) Aliases(context.Context, string) ([]string, error) { return nil, nil }

type wave52GraphStore struct {
	bySubject map[string][]graph.Triple
	addErr    error
	removeErr error

	addCalls    int
	removeCalls int
	added       []graph.Triple
	removed     []graph.Triple
}

func (s *wave52GraphStore) AddTriple(_ context.Context, triple graph.Triple) error {
	s.addCalls++
	s.added = append(s.added, triple)
	if s.addErr != nil {
		return s.addErr
	}
	s.bySubject[triple.Subject] = append(s.bySubject[triple.Subject], triple)
	return nil
}

func (s *wave52GraphStore) AddTriples(ctx context.Context, triples []graph.Triple) error {
	for _, triple := range triples {
		if err := s.AddTriple(ctx, triple); err != nil {
			return err
		}
	}
	return nil
}

func (s *wave52GraphStore) RemoveTriple(_ context.Context, triple graph.Triple) error {
	s.removeCalls++
	s.removed = append(s.removed, triple)
	return s.removeErr
}

func (s *wave52GraphStore) QueryBySubject(
	_ context.Context,
	subject string,
) ([]graph.Triple, error) {
	return append([]graph.Triple(nil), s.bySubject[subject]...), nil
}

func (*wave52GraphStore) QueryByObject(context.Context, string) ([]graph.Triple, error) {
	return nil, nil
}

func (*wave52GraphStore) QueryBySubjectPredicate(
	context.Context,
	string,
	string,
) ([]graph.Triple, error) {
	return nil, nil
}

func (*wave52GraphStore) Traverse(context.Context, string, int, []string) ([]graph.Triple, error) {
	return nil, nil
}

func (s *wave52GraphStore) Count(context.Context) (int, error) {
	count := 0
	for _, triples := range s.bySubject {
		count += len(triples)
	}
	return count, nil
}

func (*wave52GraphStore) PredicateStats(context.Context) (map[string]int, error) { return nil, nil }

func (*wave52GraphStore) AllTriples(context.Context) ([]graph.Triple, error) { return nil, nil }

func (s *wave52GraphStore) ClearAll(context.Context) error {
	s.bySubject = make(map[string][]graph.Triple)
	return nil
}

func (*wave52GraphStore) Close() error { return nil }
