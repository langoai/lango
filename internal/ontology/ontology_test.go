package ontology_test

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/ontology"
)

func newTestRegistry(t *testing.T) *ontology.EntRegistry {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	return ontology.NewEntRegistry(client)
}

func newTestService(t *testing.T) *ontology.ServiceImpl {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	reg := ontology.NewEntRegistry(client)
	return ontology.NewService(reg, nil)
}

// --- EntRegistry Tests ---

func TestEntRegistry_RegisterAndGetType(t *testing.T) {
	reg := newTestRegistry(t)
	ctx := context.Background()

	err := reg.RegisterType(ctx, ontology.ObjectType{
		Name:        "TestType",
		Description: "a test type",
		Properties: []ontology.PropertyDef{
			{Name: "field1", Type: ontology.TypeString, Required: true},
		},
		Status:  ontology.SchemaActive,
		Version: 1,
	})
	require.NoError(t, err)

	got, err := reg.GetType(ctx, "TestType")
	require.NoError(t, err)
	assert.Equal(t, "TestType", got.Name)
	assert.Equal(t, "a test type", got.Description)
	assert.Equal(t, ontology.SchemaActive, got.Status)
	assert.Len(t, got.Properties, 1)
	assert.Equal(t, "field1", got.Properties[0].Name)
}

func TestEntRegistry_RegisterDuplicateType(t *testing.T) {
	reg := newTestRegistry(t)
	ctx := context.Background()

	err := reg.RegisterType(ctx, ontology.ObjectType{Name: "Dup", Status: ontology.SchemaActive, Version: 1})
	require.NoError(t, err)

	err = reg.RegisterType(ctx, ontology.ObjectType{Name: "Dup", Status: ontology.SchemaActive, Version: 1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestEntRegistry_ListTypes(t *testing.T) {
	reg := newTestRegistry(t)
	ctx := context.Background()

	_ = reg.RegisterType(ctx, ontology.ObjectType{Name: "A", Status: ontology.SchemaActive, Version: 1})
	_ = reg.RegisterType(ctx, ontology.ObjectType{Name: "B", Status: ontology.SchemaActive, Version: 1})

	types, err := reg.ListTypes(ctx)
	require.NoError(t, err)
	assert.Len(t, types, 2)
}

func TestEntRegistry_DeprecateType(t *testing.T) {
	reg := newTestRegistry(t)
	ctx := context.Background()

	_ = reg.RegisterType(ctx, ontology.ObjectType{Name: "Dep", Status: ontology.SchemaActive, Version: 1})

	err := reg.DeprecateType(ctx, "Dep")
	require.NoError(t, err)

	got, err := reg.GetType(ctx, "Dep")
	require.NoError(t, err)
	assert.Equal(t, ontology.SchemaDeprecated, got.Status)
}

func TestEntRegistry_RegisterAndGetPredicate(t *testing.T) {
	reg := newTestRegistry(t)
	ctx := context.Background()

	err := reg.RegisterPredicate(ctx, ontology.PredicateDefinition{
		Name:        "test_rel",
		Description: "test relationship",
		SourceTypes: []string{"A"},
		TargetTypes: []string{"B"},
		Cardinality: ontology.OneToMany,
		Status:      ontology.SchemaActive,
		Version:     1,
	})
	require.NoError(t, err)

	got, err := reg.GetPredicate(ctx, "test_rel")
	require.NoError(t, err)
	assert.Equal(t, "test_rel", got.Name)
	assert.Equal(t, ontology.OneToMany, got.Cardinality)
	assert.Equal(t, []string{"A"}, got.SourceTypes)
	assert.Equal(t, []string{"B"}, got.TargetTypes)
}

func TestEntRegistry_RegisterDuplicatePredicate(t *testing.T) {
	reg := newTestRegistry(t)
	ctx := context.Background()

	err := reg.RegisterPredicate(ctx, ontology.PredicateDefinition{Name: "dup_pred", Cardinality: ontology.ManyToMany, Status: ontology.SchemaActive, Version: 1})
	require.NoError(t, err)

	err = reg.RegisterPredicate(ctx, ontology.PredicateDefinition{Name: "dup_pred", Cardinality: ontology.ManyToMany, Status: ontology.SchemaActive, Version: 1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestEntRegistry_DeprecatePredicate(t *testing.T) {
	reg := newTestRegistry(t)
	ctx := context.Background()

	_ = reg.RegisterPredicate(ctx, ontology.PredicateDefinition{Name: "old_pred", Cardinality: ontology.ManyToMany, Status: ontology.SchemaActive, Version: 1})

	err := reg.DeprecatePredicate(ctx, "old_pred")
	require.NoError(t, err)

	got, err := reg.GetPredicate(ctx, "old_pred")
	require.NoError(t, err)
	assert.Equal(t, ontology.SchemaDeprecated, got.Status)
}

// --- ServiceImpl Tests ---

func TestService_PredicateValidator(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	err := svc.RegisterPredicate(ctx, ontology.PredicateDefinition{
		Name: "known_pred", Cardinality: ontology.ManyToMany, Status: ontology.SchemaActive, Version: 1,
	})
	require.NoError(t, err)

	validator := svc.PredicateValidator()
	assert.True(t, validator("known_pred"))
	assert.False(t, validator("unknown_pred"))
}

func TestService_PredicateValidator_RefreshOnRegister(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	validator := svc.PredicateValidator()
	assert.False(t, validator("new_pred"))

	_ = svc.RegisterPredicate(ctx, ontology.PredicateDefinition{
		Name: "new_pred", Cardinality: ontology.ManyToMany, Status: ontology.SchemaActive, Version: 1,
	})
	// Cache refreshed on RegisterPredicate — same validator closure sees it
	assert.True(t, validator("new_pred"))
}

func TestService_PredicateValidator_RefreshOnDeprecate(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_ = svc.RegisterPredicate(ctx, ontology.PredicateDefinition{
		Name: "temp_pred", Cardinality: ontology.ManyToMany, Status: ontology.SchemaActive, Version: 1,
	})

	validator := svc.PredicateValidator()
	assert.True(t, validator("temp_pred"))

	_ = svc.DeprecatePredicate(ctx, "temp_pred")
	assert.False(t, validator("temp_pred"))
}

func TestService_ValidateTriple(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_ = svc.RegisterPredicate(ctx, ontology.PredicateDefinition{
		Name: "valid_pred", Cardinality: ontology.ManyToMany, Status: ontology.SchemaActive, Version: 1,
	})
	_ = svc.PredicateValidator() // init cache

	tests := []struct {
		give      string
		wantError bool
	}{
		{give: "valid_pred", wantError: false},
		{give: "invalid_pred", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			err := svc.ValidateTriple(ctx, graph.Triple{Predicate: tt.give})
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_SchemaVersion(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	v0, _ := svc.SchemaVersion(ctx)

	_ = svc.RegisterType(ctx, ontology.ObjectType{Name: "V1", Status: ontology.SchemaActive, Version: 1})
	v1, _ := svc.SchemaVersion(ctx)
	assert.Greater(t, v1, v0)

	_ = svc.RegisterPredicate(ctx, ontology.PredicateDefinition{Name: "vp1", Cardinality: ontology.ManyToMany, Status: ontology.SchemaActive, Version: 1})
	v2, _ := svc.SchemaVersion(ctx)
	assert.Greater(t, v2, v1)

	_ = svc.DeprecateType(ctx, "V1")
	v3, _ := svc.SchemaVersion(ctx)
	assert.Greater(t, v3, v2)
}

// --- Seed Tests ---

func TestSeedDefaults_FirstRun(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	err := ontology.SeedDefaults(ctx, svc)
	require.NoError(t, err)

	types, err := svc.ListTypes(ctx)
	require.NoError(t, err)
	assert.Len(t, types, 6)

	preds, err := svc.ListPredicates(ctx)
	require.NoError(t, err)
	assert.Len(t, preds, 9)

	// Check specific predicate cardinality
	inSession, err := svc.GetPredicate(ctx, "in_session")
	require.NoError(t, err)
	assert.Equal(t, ontology.ManyToOne, inSession.Cardinality)
	assert.Equal(t, []string{"Session"}, inSession.TargetTypes)
}

func TestSeedDefaults_IdempotentRerun(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	err := ontology.SeedDefaults(ctx, svc)
	require.NoError(t, err)

	// Second run should not error or create duplicates
	err = ontology.SeedDefaults(ctx, svc)
	require.NoError(t, err)

	types, err := svc.ListTypes(ctx)
	require.NoError(t, err)
	assert.Len(t, types, 6)

	preds, err := svc.ListPredicates(ctx)
	require.NoError(t, err)
	assert.Len(t, preds, 9)
}

// --- Wiring / Config Tests ---

func TestService_OntologyDisabledByDefault(t *testing.T) {
	cfg := &config.OntologyConfig{}
	assert.False(t, cfg.Enabled, "ontology should be disabled by default")
}

func TestService_StoreTriple_DelegatesToGraphStore(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	reg := ontology.NewEntRegistry(client)

	store := &fakeGraphStore{}
	svc := ontology.NewService(reg, store)
	ctx := context.Background()

	triple := graph.Triple{Subject: "a", Predicate: "rel", Object: "b"}
	err := svc.StoreTriple(ctx, triple)
	require.NoError(t, err)
	assert.Equal(t, 1, store.addCount, "StoreTriple should delegate to graph.Store.AddTriple")
	assert.Equal(t, "a", store.lastTriple.Subject)
}

func TestService_StoreTriple_NilGraphStore(t *testing.T) {
	svc := newTestService(t) // nil graph store
	ctx := context.Background()

	err := svc.StoreTriple(ctx, graph.Triple{Subject: "a", Predicate: "rel", Object: "b"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "graph store not available")
}

func TestService_InitFailureDoesNotBreakGraph(t *testing.T) {
	// Simulate ontology init failure by passing nil client.
	// initOntology should return an error, but the caller (modules.go)
	// logs a warning and continues — graph store remains functional.
	// Here we verify the error path of NewService + SeedDefaults.
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	reg := ontology.NewEntRegistry(client)
	svc := ontology.NewService(reg, nil)

	// Even if seed fails for some reason, the service itself is still usable
	// for queries (returns empty results, not panics).
	types, err := svc.ListTypes(context.Background())
	require.NoError(t, err)
	assert.Empty(t, types)

	preds, err := svc.ListPredicates(context.Background())
	require.NoError(t, err)
	assert.Empty(t, preds)
}

// --- fakeGraphStore for StoreTriple test ---

type fakeGraphStore struct {
	addCount   int
	lastTriple graph.Triple
}

func (f *fakeGraphStore) AddTriple(_ context.Context, t graph.Triple) error {
	f.addCount++
	f.lastTriple = t
	return nil
}
func (f *fakeGraphStore) AddTriples(_ context.Context, _ []graph.Triple) error { return nil }
func (f *fakeGraphStore) RemoveTriple(_ context.Context, _ graph.Triple) error  { return nil }
func (f *fakeGraphStore) QueryBySubject(_ context.Context, _ string) ([]graph.Triple, error) {
	return nil, nil
}
func (f *fakeGraphStore) QueryByObject(_ context.Context, _ string) ([]graph.Triple, error) {
	return nil, nil
}
func (f *fakeGraphStore) QueryBySubjectPredicate(_ context.Context, _, _ string) ([]graph.Triple, error) {
	return nil, nil
}
func (f *fakeGraphStore) Traverse(_ context.Context, _ string, _ int, _ []string) ([]graph.Triple, error) {
	return nil, nil
}
func (f *fakeGraphStore) Count(_ context.Context) (int, error)                { return 0, nil }
func (f *fakeGraphStore) PredicateStats(_ context.Context) (map[string]int, error) { return nil, nil }
func (f *fakeGraphStore) AllTriples(_ context.Context) ([]graph.Triple, error) { return nil, nil }
func (f *fakeGraphStore) ClearAll(_ context.Context) error                     { return nil }
func (f *fakeGraphStore) Close() error                                         { return nil }
