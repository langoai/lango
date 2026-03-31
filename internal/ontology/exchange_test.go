package ontology_test

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/ontology"
	"github.com/langoai/lango/internal/testutil"
)

func newExchangeTestEnv(t *testing.T) *ontology.ServiceImpl {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })

	reg := ontology.NewEntRegistry(client)
	gs := testutil.NewMockGraphStore()
	svc := ontology.NewService(reg, gs)
	require.NoError(t, ontology.SeedDefaults(context.Background(), svc))

	ps := ontology.NewPropertyStore(client)
	svc.SetPropertyStore(ps)

	return svc
}

func TestExportSchema_SeedDefaults(t *testing.T) {
	svc := newExchangeTestEnv(t)
	ctx := context.Background()

	bundle, err := svc.ExportSchema(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, bundle.Version)
	assert.Equal(t, "local", bundle.ExportedBy)
	assert.GreaterOrEqual(t, len(bundle.Types), 6, "should export at least 6 seed types")
	assert.GreaterOrEqual(t, len(bundle.Predicates), 9, "should export at least 9 seed predicates")
	assert.NotEmpty(t, bundle.Digest)
}

func TestExportSchema_DigestStability(t *testing.T) {
	svc := newExchangeTestEnv(t)
	ctx := context.Background()

	bundle1, err := svc.ExportSchema(ctx)
	require.NoError(t, err)

	bundle2, err := svc.ExportSchema(ctx)
	require.NoError(t, err)

	assert.Equal(t, bundle1.Digest, bundle2.Digest, "digest should be identical for same schema")
}

func TestExportSchema_SlimTypes(t *testing.T) {
	svc := newExchangeTestEnv(t)
	ctx := context.Background()

	bundle, err := svc.ExportSchema(ctx)
	require.NoError(t, err)

	// Slim types should have Name but no UUID/status fields (verified by struct type)
	for _, st := range bundle.Types {
		assert.NotEmpty(t, st.Name)
	}
	for _, sp := range bundle.Predicates {
		assert.NotEmpty(t, sp.Name)
	}
}

func TestComputeDigest_OrderIndependent(t *testing.T) {
	types := []ontology.SchemaTypeSlim{
		{Name: "A", Description: "first"},
		{Name: "B", Description: "second"},
	}
	preds := []ontology.SchemaPredicateSlim{
		{Name: "x", Cardinality: "many_to_many"},
		{Name: "y", Cardinality: "one_to_one"},
	}

	d1 := ontology.ComputeDigest(types, preds)

	// Reverse order
	typesRev := []ontology.SchemaTypeSlim{types[1], types[0]}
	predsRev := []ontology.SchemaPredicateSlim{preds[1], preds[0]}

	d2 := ontology.ComputeDigest(typesRev, predsRev)

	assert.Equal(t, d1, d2, "digest should be order-independent")
}

func TestImportSchema_Shadow(t *testing.T) {
	svc := newExchangeTestEnv(t)
	ctx := context.Background()

	bundle := &ontology.SchemaBundle{
		Version: 1,
		Types: []ontology.SchemaTypeSlim{
			{Name: "ImportedType", Description: "from peer"},
		},
		Predicates: []ontology.SchemaPredicateSlim{
			{Name: "imported_pred", Cardinality: "many_to_many"},
		},
	}

	result, err := svc.ImportSchema(ctx, bundle, ontology.ImportOptions{Mode: ontology.ImportShadow})
	require.NoError(t, err)

	assert.Equal(t, 1, result.TypesAdded)
	assert.Equal(t, 1, result.PredsAdded)

	// Verify imported type has shadow status
	imported, err := svc.GetType(ctx, "ImportedType")
	require.NoError(t, err)
	assert.Equal(t, ontology.SchemaShadow, imported.Status)
}

func TestImportSchema_Governed(t *testing.T) {
	svc := newExchangeTestEnv(t)
	ctx := context.Background()

	// Enable governance
	svc.SetGovernanceEngine(ontology.NewGovernanceEngine(ontology.GovernancePolicy{MaxNewPerDay: 100}))

	bundle := &ontology.SchemaBundle{
		Version: 1,
		Types: []ontology.SchemaTypeSlim{
			{Name: "GovernedType", Description: "needs review"},
		},
	}

	result, err := svc.ImportSchema(ctx, bundle, ontology.ImportOptions{Mode: ontology.ImportGoverned})
	require.NoError(t, err)
	assert.Equal(t, 1, result.TypesAdded)

	imported, err := svc.GetType(ctx, "GovernedType")
	require.NoError(t, err)
	assert.Equal(t, ontology.SchemaProposed, imported.Status)
}

func TestImportSchema_DryRun(t *testing.T) {
	svc := newExchangeTestEnv(t)
	ctx := context.Background()

	bundle := &ontology.SchemaBundle{
		Version: 1,
		Types: []ontology.SchemaTypeSlim{
			{Name: "DryRunType", Description: "should not persist"},
		},
	}

	result, err := svc.ImportSchema(ctx, bundle, ontology.ImportOptions{Mode: ontology.ImportDryRun})
	require.NoError(t, err)
	assert.Equal(t, 1, result.TypesAdded)

	// Should NOT actually exist
	_, err = svc.GetType(ctx, "DryRunType")
	assert.Error(t, err)
}

func TestImportSchema_NameConflict(t *testing.T) {
	svc := newExchangeTestEnv(t)
	ctx := context.Background()

	// "ErrorPattern" is a seed type — import with different properties
	bundle := &ontology.SchemaBundle{
		Version: 1,
		Types: []ontology.SchemaTypeSlim{
			{Name: "ErrorPattern", Description: "DIFFERENT description from seed"},
		},
	}

	result, err := svc.ImportSchema(ctx, bundle, ontology.ImportOptions{Mode: ontology.ImportShadow})
	require.NoError(t, err)
	assert.Equal(t, 0, result.TypesAdded)
	assert.Contains(t, result.TypesConflicting, "ErrorPattern")
}

func TestImportSchema_SkipsIdentical(t *testing.T) {
	svc := newExchangeTestEnv(t)
	ctx := context.Background()

	// Export then re-import the same schema
	bundle, err := svc.ExportSchema(ctx)
	require.NoError(t, err)

	result, err := svc.ImportSchema(ctx, bundle, ontology.ImportOptions{Mode: ontology.ImportShadow})
	require.NoError(t, err)

	assert.Equal(t, 0, result.TypesAdded)
	assert.Equal(t, len(bundle.Types), result.TypesSkipped)
	assert.Empty(t, result.TypesConflicting)
}

func TestRoundtrip_PreservesSemantics(t *testing.T) {
	original := ontology.ObjectType{
		Name:        "TestType",
		Description: "test desc",
		Properties: []ontology.PropertyDef{
			{Name: "field1", Type: ontology.TypeString, Required: true},
			{Name: "field2", Type: ontology.TypeInt, Required: false},
		},
		Extends: "BaseType",
		Status:  ontology.SchemaActive,
		Version: 3,
	}

	slim := ontology.TypeToSlim(original)
	restored := ontology.SlimToType(slim, ontology.SchemaShadow)

	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.Description, restored.Description)
	assert.Equal(t, original.Extends, restored.Extends)
	assert.Equal(t, len(original.Properties), len(restored.Properties))
	for i := range original.Properties {
		assert.Equal(t, original.Properties[i].Name, restored.Properties[i].Name)
		assert.Equal(t, original.Properties[i].Type, restored.Properties[i].Type)
		assert.Equal(t, original.Properties[i].Required, restored.Properties[i].Required)
	}
	// Status and Version are NOT preserved (local-only)
	assert.Equal(t, ontology.SchemaShadow, restored.Status)
	assert.Equal(t, 1, restored.Version)
}
