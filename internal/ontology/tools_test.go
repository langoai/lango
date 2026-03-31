package ontology_test

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ontology"
)

func newToolsTestEnv(t *testing.T) ontology.OntologyService {
	t.Helper()
	svc, _ := newPropertyTestEnv(t) // reuses property test env (includes truth + resolution + property store)
	return svc
}

func TestBuildTools_Count(t *testing.T) {
	svc := newToolsTestEnv(t)
	tools := ontology.BuildTools(svc, nil)
	assert.Len(t, tools, 13, "should have 13 ontology tools")
}

func TestBuildTools_Names(t *testing.T) {
	svc := newToolsTestEnv(t)
	tools := ontology.BuildTools(svc, nil)

	names := make(map[string]bool, len(tools))
	for _, t := range tools {
		names[t.Name] = true
	}

	expected := []string{
		"ontology_list_types", "ontology_describe_type",
		"ontology_query_entities", "ontology_get_entity",
		"ontology_assert_fact", "ontology_retract_fact",
		"ontology_list_conflicts", "ontology_resolve_conflict",
		"ontology_merge_entities", "ontology_facts_at",
		"ontology_import_json", "ontology_import_csv", "ontology_from_mcp",
	}
	for _, name := range expected {
		assert.True(t, names[name], "missing tool: %s", name)
	}
}

func TestOntologyListTypes(t *testing.T) {
	svc := newToolsTestEnv(t)
	tools := ontology.BuildTools(svc, nil)
	ctx := context.Background()

	handler := findHandler(tools, "ontology_list_types")
	require.NotNil(t, handler)

	result, err := handler(ctx, nil)
	require.NoError(t, err)

	m := result.(map[string]interface{})
	count := m["count"].(int)
	assert.GreaterOrEqual(t, count, 6, "should have at least 6 seeded types")
}

func TestOntologyDescribeType(t *testing.T) {
	svc := newToolsTestEnv(t)
	tools := ontology.BuildTools(svc, nil)
	ctx := context.Background()

	handler := findHandler(tools, "ontology_describe_type")
	result, err := handler(ctx, map[string]interface{}{"type_name": "ErrorPattern"})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, "ErrorPattern", m["name"])
	assert.NotNil(t, m["properties"])
	assert.NotNil(t, m["predicates"])
}

func TestOntologyAssertFact(t *testing.T) {
	svc := newToolsTestEnv(t)
	tools := ontology.BuildTools(svc, nil)
	ctx := context.Background()

	handler := findHandler(tools, "ontology_assert_fact")
	result, err := handler(ctx, map[string]interface{}{
		"subject":   "error:test",
		"predicate": "caused_by",
		"object":    "tool:test",
		"source":    "manual",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.True(t, m["stored"].(bool))
}

func TestOntologyRetractFact(t *testing.T) {
	svc := newToolsTestEnv(t)
	tools := ontology.BuildTools(svc, nil)
	ctx := context.Background()

	// First assert a fact.
	assertHandler := findHandler(tools, "ontology_assert_fact")
	_, err := assertHandler(ctx, map[string]interface{}{
		"subject": "err:r", "predicate": "related_to", "object": "err:s", "source": "manual",
	})
	require.NoError(t, err)

	// Then retract it.
	retractHandler := findHandler(tools, "ontology_retract_fact")
	result, err := retractHandler(ctx, map[string]interface{}{
		"subject": "err:r", "predicate": "related_to", "object": "err:s", "reason": "test",
	})
	require.NoError(t, err)
	assert.Equal(t, "retracted", result.(map[string]interface{})["status"])
}

func TestOntologyListConflicts(t *testing.T) {
	svc := newToolsTestEnv(t)
	tools := ontology.BuildTools(svc, nil)
	ctx := context.Background()

	handler := findHandler(tools, "ontology_list_conflicts")
	result, err := handler(ctx, nil)
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, 0, m["count"].(int))
}

func TestOntologyMergeEntities(t *testing.T) {
	svc := newToolsTestEnv(t)
	tools := ontology.BuildTools(svc, nil)
	ctx := context.Background()

	handler := findHandler(tools, "ontology_merge_entities")
	result, err := handler(ctx, map[string]interface{}{
		"canonical": "canon:1", "duplicate": "dup:1",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, "merged", m["status"])
}

func TestOntologyFactsAt(t *testing.T) {
	svc := newToolsTestEnv(t)
	tools := ontology.BuildTools(svc, nil)
	ctx := context.Background()

	// Assert a fact first.
	assertHandler := findHandler(tools, "ontology_assert_fact")
	_, err := assertHandler(ctx, map[string]interface{}{
		"subject": "node:fa", "predicate": "related_to", "object": "node:fb", "source": "manual",
	})
	require.NoError(t, err)

	handler := findHandler(tools, "ontology_facts_at")
	result, err := handler(ctx, map[string]interface{}{
		"subject":  "node:fa",
		"valid_at": "2030-01-01T00:00:00Z",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.GreaterOrEqual(t, m["count"].(int), 1)
}

func TestOntologyGetEntity(t *testing.T) {
	svc := newToolsTestEnv(t)
	tools := ontology.BuildTools(svc, nil)
	ctx := context.Background()

	// Set a property first.
	require.NoError(t, svc.SetEntityProperty(ctx, "tool:test_ge", "Tool", "name", "TestTool"))

	handler := findHandler(tools, "ontology_get_entity")
	result, err := handler(ctx, map[string]interface{}{"entity_id": "tool:test_ge"})
	require.NoError(t, err)

	entity := result.(*ontology.EntityResult)
	assert.Equal(t, "tool:test_ge", entity.EntityID)
	assert.Equal(t, "TestTool", entity.Properties["name"])
}

func TestOntologyQueryEntities(t *testing.T) {
	svc := newToolsTestEnv(t)
	tools := ontology.BuildTools(svc, nil)
	ctx := context.Background()

	require.NoError(t, svc.SetEntityProperty(ctx, "err:qe1", "ErrorPattern", "tool_name", "http"))
	require.NoError(t, svc.SetEntityProperty(ctx, "err:qe2", "ErrorPattern", "tool_name", "grpc"))

	handler := findHandler(tools, "ontology_query_entities")
	result, err := handler(ctx, map[string]interface{}{
		"type": "ErrorPattern",
		"filters": []interface{}{
			map[string]interface{}{"property": "tool_name", "op": "eq", "value": "http"},
		},
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, 1, m["count"].(int))
}

// --- Ingestion Tool Tests ---

func TestOntologyImportJSON(t *testing.T) {
	svc := newToolsTestEnv(t)
	tools := ontology.BuildTools(svc, nil)
	ctx := context.Background()

	data := `{
		"entities": [
			{
				"id": "err:ij1",
				"type": "ErrorPattern",
				"properties": {"tool_name": "http_client", "pattern": "timeout"},
				"relations": [
					{"predicate": "caused_by", "object": "tool:http", "object_type": "Tool"}
				]
			}
		]
	}`

	handler := findHandler(tools, "ontology_import_json")
	result, err := handler(ctx, map[string]interface{}{"data": data})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, 1, m["imported"].(int))
	assert.Equal(t, 0, m["errors"].(int))

	// Verify properties were stored.
	props, err := svc.GetEntityProperties(ctx, "err:ij1")
	require.NoError(t, err)
	assert.Equal(t, "http_client", props["tool_name"])
}

func TestOntologyImportJSON_InvalidType(t *testing.T) {
	svc := newToolsTestEnv(t)
	tools := ontology.BuildTools(svc, nil)
	ctx := context.Background()

	data := `{"entities": [{"id": "x", "type": "UnknownType", "properties": {"p": "v"}}]}`

	handler := findHandler(tools, "ontology_import_json")
	result, err := handler(ctx, map[string]interface{}{"data": data})
	require.NoError(t, err) // handler doesn't error — counts errors

	m := result.(map[string]interface{})
	assert.Equal(t, 0, m["imported"].(int))
	assert.Equal(t, 1, m["errors"].(int))
}

func TestOntologyImportCSV(t *testing.T) {
	svc := newToolsTestEnv(t)
	tools := ontology.BuildTools(svc, nil)
	ctx := context.Background()

	csvData := "entity_id,tool_name,pattern\nerr:csv1,http,timeout\nerr:csv2,grpc,deadline"

	handler := findHandler(tools, "ontology_import_csv")
	result, err := handler(ctx, map[string]interface{}{
		"data": csvData, "type": "ErrorPattern",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, 2, m["imported"].(int))

	props, err := svc.GetEntityProperties(ctx, "err:csv1")
	require.NoError(t, err)
	assert.Equal(t, "http", props["tool_name"])
}

func TestOntologyFromMCP(t *testing.T) {
	svc := newToolsTestEnv(t)
	tools := ontology.BuildTools(svc, nil)
	ctx := context.Background()

	handler := findHandler(tools, "ontology_from_mcp")
	result, err := handler(ctx, map[string]interface{}{
		"tool_name":   "weather_api",
		"result_json": `{"name": "weather_service"}`,
		"entity_type": "Tool",
		"predicate":   "related_to",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.NotEmpty(t, m["entity_id"])
	assert.GreaterOrEqual(t, m["properties_set"].(int), 1)
}

// --- Helpers ---

func findHandler(tools []*agent.Tool, name string) func(context.Context, map[string]interface{}) (interface{}, error) {
	for _, t := range tools {
		if t.Name == name {
			return t.Handler
		}
	}
	return nil
}
