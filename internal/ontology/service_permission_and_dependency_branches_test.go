package ontology_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/ontology"
)

func TestServicePermissionDeniedShortCircuitsWrapperBranches(t *testing.T) {
	ctx := ctxkeys.WithPrincipal(context.Background(), "blocked")

	tests := []struct {
		name string
		run  func(*ontology.ServiceImpl) error
	}{
		{
			name: "GetType",
			run: func(svc *ontology.ServiceImpl) error {
				_, err := svc.GetType(ctx, "Agent")
				return err
			},
		},
		{
			name: "ListTypes",
			run: func(svc *ontology.ServiceImpl) error {
				_, err := svc.ListTypes(ctx)
				return err
			},
		},
		{
			name: "GetPredicate",
			run: func(svc *ontology.ServiceImpl) error {
				_, err := svc.GetPredicate(ctx, "uses")
				return err
			},
		},
		{
			name: "ListPredicates",
			run: func(svc *ontology.ServiceImpl) error {
				_, err := svc.ListPredicates(ctx)
				return err
			},
		},
		{
			name: "RegisterType",
			run: func(svc *ontology.ServiceImpl) error {
				return svc.RegisterType(ctx, ontology.ObjectType{Name: "Agent"})
			},
		},
		{
			name: "RegisterPredicate",
			run: func(svc *ontology.ServiceImpl) error {
				return svc.RegisterPredicate(ctx, ontology.PredicateDefinition{Name: "uses"})
			},
		},
		{
			name: "DeprecateType",
			run: func(svc *ontology.ServiceImpl) error {
				return svc.DeprecateType(ctx, "Agent")
			},
		},
		{
			name: "DeprecatePredicate",
			run: func(svc *ontology.ServiceImpl) error {
				return svc.DeprecatePredicate(ctx, "uses")
			},
		},
		{
			name: "ValidateTriple",
			run: func(svc *ontology.ServiceImpl) error {
				return svc.ValidateTriple(ctx, graph.Triple{Subject: "s", Predicate: "uses", Object: "o"})
			},
		},
		{
			name: "RetractFact",
			run: func(svc *ontology.ServiceImpl) error {
				return svc.RetractFact(ctx, "s", "uses", "o", "denied")
			},
		},
		{
			name: "ConflictSet",
			run: func(svc *ontology.ServiceImpl) error {
				_, err := svc.ConflictSet(ctx, "s", "uses")
				return err
			},
		},
		{
			name: "ResolveConflict",
			run: func(svc *ontology.ServiceImpl) error {
				return svc.ResolveConflict(ctx, uuid.New(), "winner", "denied")
			},
		},
		{
			name: "FactsAt",
			run: func(svc *ontology.ServiceImpl) error {
				_, err := svc.FactsAt(ctx, "s", time.Unix(1, 0))
				return err
			},
		},
		{
			name: "OpenConflicts",
			run: func(svc *ontology.ServiceImpl) error {
				_, err := svc.OpenConflicts(ctx)
				return err
			},
		},
		{
			name: "Resolve",
			run: func(svc *ontology.ServiceImpl) error {
				_, err := svc.Resolve(ctx, "alias")
				return err
			},
		},
		{
			name: "MergeEntities",
			run: func(svc *ontology.ServiceImpl) error {
				_, err := svc.MergeEntities(ctx, "canonical", "duplicate")
				return err
			},
		},
		{
			name: "SplitEntity",
			run: func(svc *ontology.ServiceImpl) error {
				return svc.SplitEntity(ctx, "canonical", "alias")
			},
		},
		{
			name: "Aliases",
			run: func(svc *ontology.ServiceImpl) error {
				_, err := svc.Aliases(ctx, "canonical")
				return err
			},
		},
		{
			name: "GetEntityProperties",
			run: func(svc *ontology.ServiceImpl) error {
				_, err := svc.GetEntityProperties(ctx, "entity:1")
				return err
			},
		},
		{
			name: "QueryEntities",
			run: func(svc *ontology.ServiceImpl) error {
				_, err := svc.QueryEntities(ctx, ontology.PropertyQuery{EntityType: "Agent"})
				return err
			},
		},
		{
			name: "GetEntity",
			run: func(svc *ontology.ServiceImpl) error {
				_, err := svc.GetEntity(ctx, "entity:1")
				return err
			},
		},
		{
			name: "DeleteEntityProperties",
			run: func(svc *ontology.ServiceImpl) error {
				return svc.DeleteEntityProperties(ctx, "entity:1")
			},
		},
		{
			name: "PromotePredicate",
			run: func(svc *ontology.ServiceImpl) error {
				return svc.PromotePredicate(ctx, "uses", ontology.SchemaShadow, "denied")
			},
		},
		{
			name: "AssertP2PFact",
			run: func(svc *ontology.ServiceImpl) error {
				_, err := svc.AssertP2PFact(ctx, ontology.P2PFactInput{})
				return err
			},
		},
		{
			name: "VerifyP2PFact",
			run: func(svc *ontology.ServiceImpl) error {
				return svc.VerifyP2PFact(ctx, "s", "uses", "o")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := newServiceSchemaHealthAndTypeUsageWithoutGovernanceRegistry()
			reg.types["Agent"] = ontology.ObjectType{Name: "Agent", Status: ontology.SchemaActive}
			reg.predicates["uses"] = ontology.PredicateDefinition{Name: "uses", Status: ontology.SchemaActive}
			graphStore := &servicePermissionDeniedFailOnCallGraphStore{}
			truth := &servicePermissionDeniedFailOnCallTruthMaintainer{}
			resolver := &servicePermissionDeniedFailOnCallResolver{}
			svc := ontology.NewService(reg, graphStore)
			svc.SetTruthMaintainer(truth)
			svc.SetEntityResolver(resolver)
			svc.SetACLPolicy(serviceSchemaHealthAndTypeUsageWithoutGovernanceDenyACL{})
			reg.resetCalls()

			err := tt.run(svc)
			require.ErrorIs(t, err, ontology.ErrPermissionDenied)
			assert.Empty(t, reg.calls, "permission denial should not delegate to the registry")
			assert.Zero(t, truth.calls, "permission denial should not delegate to truth maintenance")
			assert.Zero(t, resolver.calls, "permission denial should not delegate to entity resolution")
			assert.Zero(t, graphStore.calls, "permission denial should not delegate to graph storage")
		})
	}
}

func TestServiceMissingDependencyWrapperBranches(t *testing.T) {
	svc := ontology.NewService(newServiceSchemaHealthAndTypeUsageWithoutGovernanceRegistry(), nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		run     func(*testing.T) error
		wantErr string
	}{
		{
			name: "RetractFactWithoutTruthMaintainer",
			run: func(*testing.T) error {
				return svc.RetractFact(ctx, "s", "uses", "o", "missing truth")
			},
			wantErr: "truth maintenance not initialized",
		},
		{
			name: "ConflictSetWithoutTruthMaintainer",
			run: func(t *testing.T) error {
				got, err := svc.ConflictSet(ctx, "s", "uses")
				assert.Nil(t, got)
				return err
			},
			wantErr: "truth maintenance not initialized",
		},
		{
			name: "ResolveConflictWithoutTruthMaintainer",
			run: func(*testing.T) error {
				return svc.ResolveConflict(ctx, uuid.New(), "winner", "missing truth")
			},
			wantErr: "truth maintenance not initialized",
		},
		{
			name: "FactsAtWithoutTruthMaintainer",
			run: func(t *testing.T) error {
				got, err := svc.FactsAt(ctx, "s", time.Unix(1, 0))
				assert.Nil(t, got)
				return err
			},
			wantErr: "truth maintenance not initialized",
		},
		{
			name: "OpenConflictsWithoutTruthMaintainer",
			run: func(t *testing.T) error {
				got, err := svc.OpenConflicts(ctx)
				assert.Nil(t, got)
				return err
			},
			wantErr: "truth maintenance not initialized",
		},
		{
			name: "MergeEntitiesWithoutResolver",
			run: func(t *testing.T) error {
				got, err := svc.MergeEntities(ctx, "canonical", "duplicate")
				assert.Nil(t, got)
				return err
			},
			wantErr: "entity resolver not initialized",
		},
		{
			name: "SplitEntityWithoutResolver",
			run: func(*testing.T) error {
				return svc.SplitEntity(ctx, "canonical", "alias")
			},
			wantErr: "entity resolver not initialized",
		},
		{
			name: "GetEntityPropertiesWithoutPropertyStore",
			run: func(t *testing.T) error {
				got, err := svc.GetEntityProperties(ctx, "entity:1")
				assert.Nil(t, got)
				return err
			},
			wantErr: "property store not initialized",
		},
		{
			name: "QueryEntitiesWithoutPropertyStore",
			run: func(t *testing.T) error {
				got, err := svc.QueryEntities(ctx, ontology.PropertyQuery{EntityType: "Agent"})
				assert.Nil(t, got)
				return err
			},
			wantErr: "property store not initialized",
		},
		{
			name: "GetEntityWithoutPropertyStore",
			run: func(t *testing.T) error {
				got, err := svc.GetEntity(ctx, "entity:1")
				assert.Nil(t, got)
				return err
			},
			wantErr: "property store not initialized",
		},
		{
			name: "DeleteEntityPropertiesWithoutPropertyStore",
			run: func(*testing.T) error {
				return svc.DeleteEntityProperties(ctx, "entity:1")
			},
			wantErr: "property store not initialized",
		},
		{
			name: "AssertP2PFactWithoutTruthMaintainer",
			run: func(t *testing.T) error {
				got, err := svc.AssertP2PFact(ctx, ontology.P2PFactInput{})
				assert.Nil(t, got)
				return err
			},
			wantErr: "truth maintenance not initialized",
		},
		{
			name: "VerifyP2PFactWithoutGraphStore",
			run: func(*testing.T) error {
				return svc.VerifyP2PFact(ctx, "s", "uses", "o")
			},
			wantErr: "graph store not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run(t)
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}

	aliases, err := svc.Aliases(ctx, "canonical")
	require.NoError(t, err)
	assert.Nil(t, aliases)
}

type servicePermissionDeniedFailOnCallTruthMaintainer struct {
	calls int
}

func (tm *servicePermissionDeniedFailOnCallTruthMaintainer) AssertFact(context.Context, ontology.AssertionInput) (*ontology.AssertionResult, error) {
	tm.calls++
	return nil, nil
}

func (tm *servicePermissionDeniedFailOnCallTruthMaintainer) RetractFact(context.Context, string, string, string, string) error {
	tm.calls++
	return nil
}

func (tm *servicePermissionDeniedFailOnCallTruthMaintainer) ConflictSet(context.Context, string, string) ([]ontology.Conflict, error) {
	tm.calls++
	return nil, nil
}

func (tm *servicePermissionDeniedFailOnCallTruthMaintainer) ResolveConflict(context.Context, uuid.UUID, string, string) error {
	tm.calls++
	return nil
}

func (tm *servicePermissionDeniedFailOnCallTruthMaintainer) FactsAt(context.Context, string, time.Time) ([]graph.Triple, error) {
	tm.calls++
	return nil, nil
}

func (tm *servicePermissionDeniedFailOnCallTruthMaintainer) OpenConflicts(context.Context) ([]ontology.Conflict, error) {
	tm.calls++
	return nil, nil
}

type servicePermissionDeniedFailOnCallResolver struct {
	calls int
}

func (r *servicePermissionDeniedFailOnCallResolver) Resolve(context.Context, string) (string, error) {
	r.calls++
	return "", nil
}

func (r *servicePermissionDeniedFailOnCallResolver) RegisterAlias(context.Context, string, string, string) error {
	r.calls++
	return nil
}

func (r *servicePermissionDeniedFailOnCallResolver) DeclareSameAs(context.Context, string, string, string) error {
	r.calls++
	return nil
}

func (r *servicePermissionDeniedFailOnCallResolver) Merge(context.Context, string, string) (*ontology.MergeResult, error) {
	r.calls++
	return nil, nil
}

func (r *servicePermissionDeniedFailOnCallResolver) Split(context.Context, string, string) error {
	r.calls++
	return nil
}

func (r *servicePermissionDeniedFailOnCallResolver) Aliases(context.Context, string) ([]string, error) {
	r.calls++
	return nil, nil
}

type servicePermissionDeniedFailOnCallGraphStore struct {
	calls int
}

func (g *servicePermissionDeniedFailOnCallGraphStore) AddTriple(context.Context, graph.Triple) error {
	g.calls++
	return nil
}

func (g *servicePermissionDeniedFailOnCallGraphStore) AddTriples(context.Context, []graph.Triple) error {
	g.calls++
	return nil
}

func (g *servicePermissionDeniedFailOnCallGraphStore) RemoveTriple(context.Context, graph.Triple) error {
	g.calls++
	return nil
}

func (g *servicePermissionDeniedFailOnCallGraphStore) QueryBySubject(context.Context, string) ([]graph.Triple, error) {
	g.calls++
	return nil, nil
}

func (g *servicePermissionDeniedFailOnCallGraphStore) QueryByObject(context.Context, string) ([]graph.Triple, error) {
	g.calls++
	return nil, nil
}

func (g *servicePermissionDeniedFailOnCallGraphStore) QueryBySubjectPredicate(context.Context, string, string) ([]graph.Triple, error) {
	g.calls++
	return nil, nil
}

func (g *servicePermissionDeniedFailOnCallGraphStore) Traverse(context.Context, string, int, []string) ([]graph.Triple, error) {
	g.calls++
	return nil, nil
}

func (g *servicePermissionDeniedFailOnCallGraphStore) Count(context.Context) (int, error) {
	g.calls++
	return 0, nil
}

func (g *servicePermissionDeniedFailOnCallGraphStore) PredicateStats(context.Context) (map[string]int, error) {
	g.calls++
	return nil, nil
}

func (g *servicePermissionDeniedFailOnCallGraphStore) AllTriples(context.Context) ([]graph.Triple, error) {
	g.calls++
	return nil, nil
}

func (g *servicePermissionDeniedFailOnCallGraphStore) ClearAll(context.Context) error {
	g.calls++
	return nil
}

func (g *servicePermissionDeniedFailOnCallGraphStore) Close() error {
	g.calls++
	return nil
}
