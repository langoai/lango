package ontologybridge

import (
	"context"
	"encoding/json"
	"fmt"
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

// --- Regression Tests (Stage 3 Review) ---

// mockTrustScorer implements TrustScorer for testing.
type mockTrustScorer struct {
	scores map[string]float64
}

func (m *mockTrustScorer) GetScore(_ context.Context, peerDID string) (float64, error) {
	score, ok := m.scores[peerDID]
	if !ok {
		return 0, fmt.Errorf("unknown peer")
	}
	return score, nil
}

func TestBridge_HandleSchemaQuery_TrustRejected(t *testing.T) {
	// Regression: Finding 4 — low-trust peer must be rejected
	bundle := &ontology.SchemaBundle{Version: 1, Types: []ontology.SchemaTypeSlim{{Name: "T"}}}
	svc := &mockOntologyService{exportResult: bundle}
	cfg := DefaultConfig()
	cfg.MinTrustForSchema = 0.5

	var rep TrustScorer = &mockTrustScorer{scores: map[string]float64{"did:lango:untrusted": 0.2}}
	b := New(svc, nil, cfg)
	b.SetReputation(rep)

	_, err := b.HandleSchemaQuery(context.Background(), "did:lango:untrusted", protocol.SchemaQueryRequest{})
	assert.Error(t, err, "low-trust peer should be rejected")
	assert.Contains(t, err.Error(), "trust")
}

func TestBridge_HandleSchemaQuery_TrustAccepted(t *testing.T) {
	bundle := &ontology.SchemaBundle{Version: 1, Types: []ontology.SchemaTypeSlim{{Name: "T"}}}
	svc := &mockOntologyService{exportResult: bundle}
	cfg := DefaultConfig()
	cfg.MinTrustForSchema = 0.5

	var rep TrustScorer = &mockTrustScorer{scores: map[string]float64{"did:lango:trusted": 0.8}}
	b := New(svc, nil, cfg)
	b.SetReputation(rep)

	resp, err := b.HandleSchemaQuery(context.Background(), "did:lango:trusted", protocol.SchemaQueryRequest{})
	require.NoError(t, err)
	assert.NotNil(t, resp.Bundle)
}
