package ontologybridge

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ontology"
	"github.com/langoai/lango/internal/p2p/protocol"
)

// mockOntologyService implements the subset of OntologyService needed by Bridge tests.
type mockOntologyService struct {
	ontology.OntologyService
	exportResult *ontology.SchemaBundle
	importResult *ontology.ImportResult
	importErr    error
}

func (m *mockOntologyService) ExportSchema(_ context.Context) (*ontology.SchemaBundle, error) {
	return m.exportResult, nil
}

func (m *mockOntologyService) ImportSchema(_ context.Context, _ *ontology.SchemaBundle, _ ontology.ImportOptions) (*ontology.ImportResult, error) {
	return m.importResult, m.importErr
}

func TestBridge_HandleSchemaQuery(t *testing.T) {
	bundle := &ontology.SchemaBundle{
		Version: 1,
		Types: []ontology.SchemaTypeSlim{
			{Name: "TestType", Description: "test"},
		},
		Digest: "abc123",
	}
	svc := &mockOntologyService{exportResult: bundle}
	b := New(svc, nil, DefaultConfig())

	resp, err := b.HandleSchemaQuery(context.Background(), "did:lango:peer1", protocol.SchemaQueryRequest{})
	require.NoError(t, err)

	var decoded ontology.SchemaBundle
	require.NoError(t, json.Unmarshal(resp.Bundle, &decoded))
	assert.Equal(t, 1, len(decoded.Types))
	assert.Equal(t, "TestType", decoded.Types[0].Name)
}

func TestBridge_HandleSchemaQuery_FilteredTypes(t *testing.T) {
	bundle := &ontology.SchemaBundle{
		Version: 1,
		Types: []ontology.SchemaTypeSlim{
			{Name: "TypeA"},
			{Name: "TypeB"},
			{Name: "TypeC"},
		},
	}
	svc := &mockOntologyService{exportResult: bundle}
	b := New(svc, nil, DefaultConfig())

	resp, err := b.HandleSchemaQuery(context.Background(), "did:lango:peer1", protocol.SchemaQueryRequest{
		RequestedTypes: []string{"TypeA", "TypeC"},
	})
	require.NoError(t, err)

	var decoded ontology.SchemaBundle
	require.NoError(t, json.Unmarshal(resp.Bundle, &decoded))
	assert.Equal(t, 2, len(decoded.Types))
}

func TestBridge_HandleSchemaPropose_Shadow(t *testing.T) {
	svc := &mockOntologyService{
		importResult: &ontology.ImportResult{TypesAdded: 2, PredsAdded: 1},
	}
	b := New(svc, nil, DefaultConfig())

	bundle := ontology.SchemaBundle{
		Version: 1,
		Types:   []ontology.SchemaTypeSlim{{Name: "A"}, {Name: "B"}},
	}
	bundleJSON, _ := json.Marshal(bundle)

	resp, err := b.HandleSchemaPropose(context.Background(), "did:lango:peer1", protocol.SchemaProposeRequest{
		Bundle: bundleJSON,
	})
	require.NoError(t, err)
	assert.Equal(t, "accepted", resp.Action)
}

func TestBridge_HandleSchemaPropose_TooManyTypes(t *testing.T) {
	svc := &mockOntologyService{}
	cfg := DefaultConfig()
	cfg.MaxTypesPerImport = 2
	b := New(svc, nil, cfg)

	types := make([]ontology.SchemaTypeSlim, 5)
	for i := range types {
		types[i] = ontology.SchemaTypeSlim{Name: "T" + string(rune('A'+i))}
	}
	bundle := ontology.SchemaBundle{Version: 1, Types: types}
	bundleJSON, _ := json.Marshal(bundle)

	resp, err := b.HandleSchemaPropose(context.Background(), "did:lango:peer1", protocol.SchemaProposeRequest{Bundle: bundleJSON})
	require.NoError(t, err)
	assert.Equal(t, "rejected", resp.Action)
}

func TestBridge_HandleSchemaPropose_Disabled(t *testing.T) {
	svc := &mockOntologyService{}
	cfg := DefaultConfig()
	cfg.AutoImportMode = "disabled"
	b := New(svc, nil, cfg)

	bundleJSON, _ := json.Marshal(ontology.SchemaBundle{Version: 1})
	resp, err := b.HandleSchemaPropose(context.Background(), "did:lango:peer1", protocol.SchemaProposeRequest{Bundle: bundleJSON})
	require.NoError(t, err)
	assert.Equal(t, "rejected", resp.Action)
}
