package ontology_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/ontology"
)

func TestBuildToolsIncludesDynamicToolsAndMetadata(t *testing.T) {
	svc := &buildToolsIncludesDynamicToolsAndMetadataToolService{}
	reg := ontology.NewActionRegistry()
	require.NoError(t, reg.Register(&ontology.ActionType{
		Name:         "approve_fact",
		Description:  "Approve a candidate fact",
		RequiredPerm: ontology.PermWrite,
		ParamSchema: map[string]string{
			"entity_id": "Entity identifier",
			"status":    "Target status",
		},
	}))
	require.NoError(t, reg.Register(&ontology.ActionType{
		Name:        "refresh_usage",
		Description: "Refresh usage counters",
	}))

	tools := ontology.BuildTools(svc, reg)
	toolNames := make(map[string]int, len(tools))
	for _, tool := range tools {
		toolNames[tool.Name]++
	}

	listActions := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_list_actions")
	assert.Equal(t, agent.SafetyLevelSafe, listActions.SafetyLevel)
	assert.True(t, listActions.Capability.ReadOnly)
	assert.True(t, listActions.Capability.ConcurrencySafe)
	assert.Equal(t, agent.ActivityQuery, listActions.Capability.Activity)

	actionTool := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_action_approve_fact")
	assert.Equal(t, agent.SafetyLevelModerate, actionTool.SafetyLevel)
	assert.False(t, actionTool.Capability.ReadOnly)
	assert.Equal(t, agent.ActivityExecute, actionTool.Capability.Activity)
	assert.Equal(t, "Approve a candidate fact", actionTool.Description)

	required, ok := actionTool.Parameters["required"].([]string)
	require.True(t, ok)
	assert.ElementsMatch(t, []string{"entity_id", "status"}, required)

	props, ok := actionTool.Parameters["properties"].(map[string]interface{})
	require.True(t, ok)
	statusProp, ok := props["status"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", statusProp["type"])
	assert.Equal(t, "Target status", statusProp["description"])

	for _, name := range []string{
		"ontology_list_types",
		"ontology_describe_type",
		"ontology_query_entities",
		"ontology_get_entity",
		"ontology_assert_fact",
		"ontology_retract_fact",
		"ontology_list_conflicts",
		"ontology_resolve_conflict",
		"ontology_merge_entities",
		"ontology_facts_at",
		"ontology_import_json",
		"ontology_import_csv",
		"ontology_from_mcp",
		"ontology_promote_type",
		"ontology_promote_predicate",
		"ontology_schema_health",
		"ontology_type_usage",
		"ontology_action_refresh_usage",
	} {
		requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, name)
		assert.Equal(t, 1, toolNames[name], "tool should be registered exactly once")
	}
	assert.Equal(t, 1, toolNames["ontology_list_actions"], "tool should be registered exactly once")
	assert.Equal(t, 1, toolNames["ontology_action_approve_fact"], "tool should be registered exactly once")
}

func TestListActionsReturnsServicePayloadAndPropagatesErrors(t *testing.T) {
	ctx := context.Background()
	svc := &buildToolsIncludesDynamicToolsAndMetadataToolService{
		listActionsResult: []ontology.ActionSummary{
			{
				Name:         "promote",
				Description:  "Promote candidate",
				RequiredPerm: ontology.PermWrite,
				ParamSchema:  map[string]string{"id": "Identifier"},
			},
		},
	}
	tool := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, ontology.BuildTools(svc, ontology.NewActionRegistry()), "ontology_list_actions")

	result, err := tool.Handler(ctx, nil)
	require.NoError(t, err)

	payload := result.(map[string]interface{})
	assert.Equal(t, 1, payload["count"])
	assert.Equal(t, svc.listActionsResult, payload["actions"])
	assert.Equal(t, 1, svc.listActionsCalls)

	wantErr := errors.New("list actions unavailable")
	svc.listActionsErr = wantErr
	result, err = tool.Handler(ctx, nil)
	require.ErrorIs(t, err, wantErr)
	assert.Nil(t, result)
}

func TestDynamicActionToolValidatesParametersExecutesAndPropagatesErrors(t *testing.T) {
	ctx := context.Background()
	logID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	svc := &buildToolsIncludesDynamicToolsAndMetadataToolService{
		executeActionResult: &ontology.ActionResult{
			LogID:  logID,
			Status: ontology.ActionCompleted,
			Effects: &ontology.ActionEffects{
				PropertiesSet: []ontology.PropertyEffect{
					{EntityID: "entity:1", Property: "status", NewValue: "active"},
				},
			},
		},
	}
	reg := ontology.NewActionRegistry()
	require.NoError(t, reg.Register(&ontology.ActionType{
		Name:        "approve_fact",
		Description: "Approve a candidate fact",
		ParamSchema: map[string]string{
			"entity_id": "Entity identifier",
			"status":    "Target status",
		},
	}))
	tool := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, ontology.BuildTools(svc, reg), "ontology_action_approve_fact")

	result, err := tool.Handler(ctx, map[string]interface{}{"entity_id": "entity:1"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "missing status parameter")
	assert.Nil(t, result)
	assert.Zero(t, svc.executeActionCalls, "parameter validation must short-circuit service execution")

	result, err = tool.Handler(ctx, map[string]interface{}{
		"entity_id": "entity:1",
		"status":    "active",
		"attempt":   3,
	})
	require.NoError(t, err)

	assert.Equal(t, 1, svc.executeActionCalls)
	assert.Equal(t, "approve_fact", svc.lastActionName)
	assert.Equal(t, map[string]string{
		"entity_id": "entity:1",
		"status":    "active",
		"attempt":   "3",
	}, svc.lastActionParams)

	payload := result.(map[string]interface{})
	assert.Equal(t, logID.String(), payload["logID"])
	assert.Equal(t, "completed", payload["status"])
	assert.Equal(t, svc.executeActionResult.Effects, payload["effects"])
	assert.Equal(t, "", payload["error"])

	wantErr := errors.New("executor rejected action")
	svc.executeActionErr = wantErr
	result, err = tool.Handler(ctx, map[string]interface{}{
		"entity_id": "entity:1",
		"status":    "active",
	})
	require.ErrorIs(t, err, wantErr)
	assert.Nil(t, result)
}

func TestQueryEntitiesBuildsQueryFiltersPayloadAndPropagatesErrors(t *testing.T) {
	ctx := context.Background()
	svc := &buildToolsIncludesDynamicToolsAndMetadataToolService{
		queryEntitiesResult: []ontology.EntityResult{
			{
				EntityID:   "err:1",
				EntityType: "ErrorPattern",
				Properties: map[string]string{"tool_name": "http"},
				Outgoing: []ontology.ResultTriple{
					{Subject: "err:1", Predicate: "related_to", Object: "tool:http"},
					{
						Subject:   "err:1",
						Predicate: "caused_by",
						Object:    "peer:unverified",
						Metadata:  map[string]string{"_p2p_verified": "false"},
					},
				},
			},
		},
	}
	tool := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, ontology.BuildTools(svc, nil), "ontology_query_entities")

	result, err := tool.Handler(ctx, map[string]interface{}{
		"type":               "ErrorPattern",
		"limit":              float64(7),
		"exclude_unverified": true,
		"filters": []interface{}{
			map[string]interface{}{"property": "tool_name", "op": "eq", "value": "http"},
			map[string]interface{}{"property": "pattern", "op": "contains", "value": "timeout"},
			"ignored",
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, svc.queryEntitiesCalls)
	assert.Equal(t, "ErrorPattern", svc.lastQuery.EntityType)
	assert.Equal(t, 7, svc.lastQuery.Limit)
	require.Len(t, svc.lastQuery.Filters, 2)
	assert.Equal(t, ontology.PropertyFilter{Property: "tool_name", Op: ontology.FilterEq, Value: "http"}, svc.lastQuery.Filters[0])
	assert.Equal(t, ontology.PropertyFilter{Property: "pattern", Op: ontology.FilterContains, Value: "timeout"}, svc.lastQuery.Filters[1])

	payload := result.(map[string]interface{})
	assert.Equal(t, 1, payload["count"])
	entities := payload["entities"].([]ontology.EntityResult)
	require.Len(t, entities, 1)
	assert.Equal(t, []ontology.ResultTriple{{Subject: "err:1", Predicate: "related_to", Object: "tool:http"}}, entities[0].Outgoing)

	result, err = tool.Handler(ctx, map[string]interface{}{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "missing type parameter")
	assert.Nil(t, result)

	wantErr := errors.New("query store unavailable")
	svc.queryEntitiesErr = wantErr
	result, err = tool.Handler(ctx, map[string]interface{}{"type": "ErrorPattern"})
	require.ErrorIs(t, err, wantErr)
	assert.ErrorContains(t, err, "query entities")
	assert.Nil(t, result)
}

func TestGetEntityAndFactsAtValidationFilteringAndErrorBranches(t *testing.T) {
	ctx := context.Background()
	svc := &buildToolsIncludesDynamicToolsAndMetadataToolService{
		getEntityResult: &ontology.EntityResult{
			EntityID: "entity:1",
			Outgoing: []ontology.ResultTriple{
				{Subject: "entity:1", Predicate: "related_to", Object: "local:1"},
				{Subject: "entity:1", Predicate: "related_to", Object: "peer:1", Metadata: map[string]string{"_p2p_verified": "false"}},
			},
			Incoming: []ontology.ResultTriple{
				{Subject: "local:2", Predicate: "related_to", Object: "entity:1"},
				{Subject: "peer:2", Predicate: "related_to", Object: "entity:1", Metadata: map[string]string{"_p2p_verified": "false"}},
			},
		},
		factsAtResult: []graph.Triple{
			{Subject: "entity:1", Predicate: "related_to", Object: "local:1"},
			{Subject: "entity:1", Predicate: "related_to", Object: "peer:1", Metadata: map[string]string{"_p2p_verified": "false"}},
		},
	}
	tools := ontology.BuildTools(svc, nil)

	getEntity := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_get_entity")
	result, err := getEntity.Handler(ctx, map[string]interface{}{"entity_id": "entity:1"})
	require.NoError(t, err)
	entity := result.(*ontology.EntityResult)
	assert.Equal(t, "entity:1", svc.lastEntityID)
	assert.Len(t, entity.Outgoing, 1)
	assert.Len(t, entity.Incoming, 1)

	result, err = getEntity.Handler(ctx, map[string]interface{}{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "missing entity_id parameter")
	assert.Nil(t, result)

	factsAt := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_facts_at")
	result, err = factsAt.Handler(ctx, map[string]interface{}{
		"subject":  "entity:1",
		"valid_at": "2026-05-18T10:30:00Z",
	})
	require.NoError(t, err)
	assert.Equal(t, "entity:1", svc.lastFactsSubject)
	assert.Equal(t, time.Date(2026, 5, 18, 10, 30, 0, 0, time.UTC), svc.lastFactsAt)
	factsPayload := result.(map[string]interface{})
	assert.Equal(t, 1, factsPayload["count"])
	assert.Equal(t, []graph.Triple{{Subject: "entity:1", Predicate: "related_to", Object: "local:1"}}, factsPayload["facts"])

	result, err = factsAt.Handler(ctx, map[string]interface{}{
		"subject":  "entity:1",
		"valid_at": "not-a-time",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid valid_at format")
	assert.Nil(t, result)

	wantErr := errors.New("truth store unavailable")
	svc.factsAtErr = wantErr
	result, err = factsAt.Handler(ctx, map[string]interface{}{
		"subject":  "entity:1",
		"valid_at": "2026-05-18T10:30:00Z",
	})
	require.ErrorIs(t, err, wantErr)
	assert.ErrorContains(t, err, "facts at")
	assert.Nil(t, result)
}

func TestIngestionToolsValidateMapSideEffectsAndReportPartialErrors(t *testing.T) {
	ctx := context.Background()
	svc := &buildToolsIncludesDynamicToolsAndMetadataToolService{
		setEntityPropertyErrByKey: map[string]error{
			"bad:1.pattern":    errors.New("invalid property"),
			"err:csv2.pattern": errors.New("invalid csv property"),
		},
		assertFactErrBySubject: map[string]error{
			"err:2": errors.New("invalid predicate"),
		},
	}
	tools := ontology.BuildTools(svc, nil)

	importJSON := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_import_json")
	result, err := importJSON.Handler(ctx, map[string]interface{}{
		"data": `{
			"entities": [
				{"id": "err:1", "type": "ErrorPattern", "properties": {"tool_name": "http"}, "relations": [{"predicate": "caused_by", "object": "tool:http", "object_type": "Tool"}]},
				{"id": "bad:1", "type": "ErrorPattern", "properties": {"pattern": "bad"}},
				{"id": "err:2", "type": "ErrorPattern", "properties": {"tool_name": "grpc"}, "relations": [{"predicate": "caused_by", "object": "tool:grpc"}]}
			]
		}`,
	})
	require.NoError(t, err)

	jsonPayload := result.(map[string]interface{})
	assert.Equal(t, 2, jsonPayload["imported"])
	assert.Equal(t, 2, jsonPayload["errors"])
	assert.Equal(t, 3, jsonPayload["total"])
	assert.Contains(t, svc.propertyCalls, buildToolsIncludesDynamicToolsAndMetadataPropertyCall{entityID: "err:1", entityType: "ErrorPattern", property: "tool_name", value: "http"})
	assert.Contains(t, svc.assertFactCalls, buildToolsIncludesDynamicToolsAndMetadataAssertionCall{subject: "err:1", subjectType: "ErrorPattern", predicate: "caused_by", object: "tool:http", objectType: "Tool", source: "import", confidence: 0.9})

	result, err = importJSON.Handler(ctx, map[string]interface{}{"data": `{"entities":`})
	require.Error(t, err)
	assert.ErrorContains(t, err, "import json parse")
	assert.Nil(t, result)

	importCSV := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_import_csv")
	result, err = importCSV.Handler(ctx, map[string]interface{}{
		"type": "ErrorPattern",
		"data": "entity_id,tool_name,pattern\nerr:csv1,http,timeout\nerr:csv2,grpc,bad",
	})
	require.NoError(t, err)
	csvPayload := result.(map[string]interface{})
	assert.Equal(t, 1, csvPayload["imported"])
	assert.Equal(t, 1, csvPayload["errors"])
	assert.Equal(t, 2, csvPayload["total"])
	assert.Contains(t, svc.propertyCalls, buildToolsIncludesDynamicToolsAndMetadataPropertyCall{entityID: "err:csv1", entityType: "ErrorPattern", property: "tool_name", value: "http"})
	assert.Contains(t, svc.propertyCalls, buildToolsIncludesDynamicToolsAndMetadataPropertyCall{entityID: "err:csv1", entityType: "ErrorPattern", property: "pattern", value: "timeout"})

	result, err = importCSV.Handler(ctx, map[string]interface{}{"type": "ErrorPattern", "data": "entity_id\nerr:1"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "need at least entity_id column")
	assert.Nil(t, result)
}

func TestFromMCPStoresStringFieldsLinksToolAndReportsFactError(t *testing.T) {
	ctx := context.Background()
	svc := &buildToolsIncludesDynamicToolsAndMetadataToolService{}
	tool := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, ontology.BuildTools(svc, nil), "ontology_from_mcp")

	result, err := tool.Handler(ctx, map[string]interface{}{
		"tool_name":   "weather_api",
		"result_json": `{"id":"weather:1","name":"weather service","temperature":72,"status":"ok"}`,
		"entity_type": "Tool",
		"predicate":   "related_to",
	})
	require.NoError(t, err)

	payload := result.(map[string]interface{})
	assert.Equal(t, "weather:1", payload["entity_id"])
	assert.Equal(t, 3, payload["properties_set"], "only string fields, including id, are stored as properties")
	assert.Equal(t, true, payload["fact_asserted"])
	assert.Contains(t, svc.propertyCalls, buildToolsIncludesDynamicToolsAndMetadataPropertyCall{entityID: "weather:1", entityType: "Tool", property: "name", value: "weather service"})
	assert.Contains(t, svc.propertyCalls, buildToolsIncludesDynamicToolsAndMetadataPropertyCall{entityID: "weather:1", entityType: "Tool", property: "status", value: "ok"})
	assert.NotContains(t, svc.propertyCalls, buildToolsIncludesDynamicToolsAndMetadataPropertyCall{entityID: "weather:1", entityType: "Tool", property: "temperature", value: "72"})
	assert.Contains(t, svc.assertFactCalls, buildToolsIncludesDynamicToolsAndMetadataAssertionCall{subject: "weather:1", subjectType: "Tool", predicate: "related_to", object: "tool:weather_api", objectType: "Tool", source: "mcp", confidence: 0.7})

	svc.assertFactErrBySubject = map[string]error{"mcp:weather_api:Tool": errors.New("predicate not active")}
	result, err = tool.Handler(ctx, map[string]interface{}{
		"tool_name":   "weather_api",
		"result_json": `{"name":"weather service"}`,
		"entity_type": "Tool",
		"predicate":   "related_to",
	})
	require.NoError(t, err)
	payload = result.(map[string]interface{})
	assert.Equal(t, "mcp:weather_api:Tool", payload["entity_id"])
	assert.Equal(t, "predicate not active", payload["fact_error"])
	assert.NotContains(t, payload, "fact_asserted")

	result, err = tool.Handler(ctx, map[string]interface{}{
		"tool_name":   "weather_api",
		"result_json": `{"name":`,
		"entity_type": "Tool",
		"predicate":   "related_to",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "from mcp json parse")
	assert.Nil(t, result)
}

func TestGovernanceToolsValidateInputsPayloadsAndPropagateErrors(t *testing.T) {
	ctx := context.Background()
	typeUsage := &ontology.TypeUsageInfo{TypeName: "Agent", Status: ontology.SchemaActive, Version: 2}
	health := &ontology.SchemaHealthReport{
		Types:      map[ontology.SchemaStatus]int{ontology.SchemaActive: 1},
		Predicates: map[ontology.SchemaStatus]int{ontology.SchemaShadow: 2},
	}
	svc := &buildToolsIncludesDynamicToolsAndMetadataToolService{
		typeUsageResult:    typeUsage,
		schemaHealthResult: health,
	}
	tools := ontology.BuildTools(svc, nil)

	promoteType := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_promote_type")
	result, err := promoteType.Handler(ctx, map[string]interface{}{
		"type_name":     "Agent",
		"target_status": "active",
		"reason":        "approved",
	})
	require.NoError(t, err)
	assert.Equal(t, buildToolsIncludesDynamicToolsAndMetadataPromotionCall{name: "Agent", target: ontology.SchemaActive, reason: "approved"}, svc.promoteTypeCalls[0])
	assert.Equal(t, map[string]interface{}{"status": "promoted", "type": "Agent", "newStatus": "active"}, result)

	result, err = promoteType.Handler(ctx, map[string]interface{}{"target_status": "active", "reason": "approved"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "missing type_name parameter")
	assert.Nil(t, result)

	promotePredicate := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_promote_predicate")
	result, err = promotePredicate.Handler(ctx, map[string]interface{}{
		"predicate_name": "related_to",
		"target_status":  "shadow",
		"reason":         "candidate",
	})
	require.NoError(t, err)
	assert.Equal(t, buildToolsIncludesDynamicToolsAndMetadataPromotionCall{name: "related_to", target: ontology.SchemaShadow, reason: "candidate"}, svc.promotePredicateCalls[0])
	assert.Equal(t, map[string]interface{}{"status": "promoted", "predicate": "related_to", "newStatus": "shadow"}, result)

	usageTool := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_type_usage")
	result, err = usageTool.Handler(ctx, map[string]interface{}{"type_name": "Agent"})
	require.NoError(t, err)
	assert.Equal(t, "Agent", svc.lastTypeUsageName)
	assert.Equal(t, typeUsage, result)

	healthTool := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_schema_health")
	result, err = healthTool.Handler(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, health, result)

	wantErr := errors.New("governance unavailable")
	svc.promoteTypeErr = wantErr
	result, err = promoteType.Handler(ctx, map[string]interface{}{
		"type_name":     "Agent",
		"target_status": "active",
		"reason":        "approved",
	})
	require.ErrorIs(t, err, wantErr)
	assert.Nil(t, result)
}

func TestStaticOntologyToolsResidualValidationAndErrorBranches(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("ontology unavailable")
	conflictID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	svc := &buildToolsIncludesDynamicToolsAndMetadataToolService{
		listTypesResult: []ontology.ObjectType{{
			Name:        "Agent",
			Description: "Agent type",
			Status:      ontology.SchemaActive,
			Properties:  []ontology.PropertyDef{{Name: "name", Type: ontology.TypeString}},
		}},
		getTypeResult: &ontology.ObjectType{
			Name:        "Agent",
			Description: "Agent type",
			Status:      ontology.SchemaActive,
			Properties:  []ontology.PropertyDef{{Name: "name", Type: ontology.TypeString}},
		},
		listPredicatesResult: []ontology.PredicateDefinition{{
			Name:        "uses",
			Description: "Agent uses tool",
			SourceTypes: []string{"Agent"},
			TargetTypes: []string{"Tool"},
			Cardinality: ontology.ManyToMany,
		}},
		assertFactResult: &ontology.AssertionResult{
			Stored:     false,
			Message:    "conflict detected",
			ConflictID: &conflictID,
		},
		mergeResult: &ontology.MergeResult{TriplesUpdated: 2, AliasesCreated: 1},
	}
	tools := ontology.BuildTools(svc, nil)

	listTypes := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_list_types")
	result, err := listTypes.Handler(ctx, nil)
	require.NoError(t, err)
	payload := result.(map[string]interface{})
	assert.Equal(t, 1, payload["count"])
	svc.listTypesErr = wantErr
	result, err = listTypes.Handler(ctx, nil)
	require.ErrorIs(t, err, wantErr)
	assert.ErrorContains(t, err, "list types")
	assert.Nil(t, result)
	svc.listTypesErr = nil

	describe := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_describe_type")
	result, err = describe.Handler(ctx, map[string]interface{}{"type_name": "Agent"})
	require.NoError(t, err)
	describePayload := result.(map[string]interface{})
	assert.Equal(t, "Agent", describePayload["name"])
	assert.Len(t, describePayload["predicates"], 1)
	result, err = describe.Handler(ctx, map[string]interface{}{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "missing type_name parameter")
	assert.Nil(t, result)
	svc.getTypeErr = wantErr
	result, err = describe.Handler(ctx, map[string]interface{}{"type_name": "Agent"})
	require.ErrorIs(t, err, wantErr)
	assert.ErrorContains(t, err, "describe type")
	assert.Nil(t, result)
	svc.getTypeErr = nil

	assertFact := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_assert_fact")
	result, err = assertFact.Handler(ctx, map[string]interface{}{
		"subject":    "agent:1",
		"predicate":  "uses",
		"object":     "tool:1",
		"confidence": 0.25,
	})
	require.NoError(t, err)
	assertPayload := result.(map[string]interface{})
	assert.Equal(t, false, assertPayload["stored"])
	assert.Equal(t, conflictID.String(), assertPayload["conflict_id"])
	assert.Equal(t, 0.25, svc.assertFactCalls[len(svc.assertFactCalls)-1].confidence)
	result, err = assertFact.Handler(ctx, map[string]interface{}{"subject": "agent:1", "object": "tool:1"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "missing predicate parameter")
	assert.Nil(t, result)
	svc.assertFactErrBySubject = map[string]error{"agent:1": wantErr}
	result, err = assertFact.Handler(ctx, map[string]interface{}{
		"subject":   "agent:1",
		"predicate": "uses",
		"object":    "tool:1",
	})
	require.ErrorIs(t, err, wantErr)
	assert.ErrorContains(t, err, "assert fact")
	assert.Nil(t, result)
	svc.assertFactErrBySubject = nil

	retract := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_retract_fact")
	result, err = retract.Handler(ctx, map[string]interface{}{"subject": "agent:1", "predicate": "uses", "object": "tool:1"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "missing reason parameter")
	assert.Nil(t, result)
	svc.retractFactErr = wantErr
	result, err = retract.Handler(ctx, map[string]interface{}{
		"subject":   "agent:1",
		"predicate": "uses",
		"object":    "tool:1",
		"reason":    "duplicate",
	})
	require.ErrorIs(t, err, wantErr)
	assert.ErrorContains(t, err, "retract fact")
	assert.Nil(t, result)
	svc.retractFactErr = nil

	conflicts := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_list_conflicts")
	svc.openConflictsErr = wantErr
	result, err = conflicts.Handler(ctx, nil)
	require.ErrorIs(t, err, wantErr)
	assert.ErrorContains(t, err, "list conflicts")
	assert.Nil(t, result)
	svc.openConflictsErr = nil

	merge := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_merge_entities")
	result, err = merge.Handler(ctx, map[string]interface{}{"canonical": "agent:1"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "missing duplicate parameter")
	assert.Nil(t, result)
	svc.mergeEntitiesErr = wantErr
	result, err = merge.Handler(ctx, map[string]interface{}{"canonical": "agent:1", "duplicate": "agent:old"})
	require.ErrorIs(t, err, wantErr)
	assert.ErrorContains(t, err, "merge entities")
	assert.Nil(t, result)
	svc.mergeEntitiesErr = nil
	result, err = merge.Handler(ctx, map[string]interface{}{"canonical": "agent:1", "duplicate": "agent:old"})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"status": "merged", "triples_updated": 2, "aliases_created": 1}, result)
}

func TestIngestionAndGovernanceResidualErrorBranches(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("governance unavailable")
	svc := &buildToolsIncludesDynamicToolsAndMetadataToolService{}
	tools := ontology.BuildTools(svc, nil)

	importCSV := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_import_csv")
	result, err := importCSV.Handler(ctx, map[string]interface{}{"data": "entity_id,name\nagent:1,Alice"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "missing type parameter")
	assert.Nil(t, result)
	result, err = importCSV.Handler(ctx, map[string]interface{}{"type": "Agent", "data": "\"unterminated"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "import csv parse")
	assert.Nil(t, result)
	result, err = importCSV.Handler(ctx, map[string]interface{}{"type": "Agent", "data": "entity_id,name\nagent:1"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "import csv parse")
	assert.Nil(t, result)

	fromMCP := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_from_mcp")
	result, err = fromMCP.Handler(ctx, map[string]interface{}{"tool_name": "weather"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "missing result_json parameter")
	assert.Nil(t, result)

	promotePredicate := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_promote_predicate")
	svc.promotePredicateErr = wantErr
	result, err = promotePredicate.Handler(ctx, map[string]interface{}{
		"predicate_name": "uses",
		"target_status":  "active",
		"reason":         "approved",
	})
	require.ErrorIs(t, err, wantErr)
	assert.Nil(t, result)

	schemaHealth := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_schema_health")
	svc.schemaHealthErr = wantErr
	result, err = schemaHealth.Handler(ctx, nil)
	require.ErrorIs(t, err, wantErr)
	assert.Nil(t, result)

	typeUsage := requireBuildToolsIncludesDynamicToolsAndMetadataTool(t, tools, "ontology_type_usage")
	svc.typeUsageErr = wantErr
	result, err = typeUsage.Handler(ctx, map[string]interface{}{"type_name": "Agent"})
	require.ErrorIs(t, err, wantErr)
	assert.Nil(t, result)
}

func requireBuildToolsIncludesDynamicToolsAndMetadataTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	require.FailNowf(t, "missing ontology tool", "tool %q not found", name)
	return nil
}

type buildToolsIncludesDynamicToolsAndMetadataToolService struct {
	ontology.OntologyService

	listActionsResult []ontology.ActionSummary
	listActionsErr    error
	listActionsCalls  int

	listTypesResult []ontology.ObjectType
	listTypesErr    error

	getTypeResult        *ontology.ObjectType
	getTypeErr           error
	listPredicatesResult []ontology.PredicateDefinition
	listPredicatesErr    error

	executeActionResult *ontology.ActionResult
	executeActionErr    error
	executeActionCalls  int
	lastActionName      string
	lastActionParams    map[string]string

	queryEntitiesResult []ontology.EntityResult
	queryEntitiesErr    error
	queryEntitiesCalls  int
	lastQuery           ontology.PropertyQuery

	getEntityResult *ontology.EntityResult
	getEntityErr    error
	lastEntityID    string

	factsAtResult    []graph.Triple
	factsAtErr       error
	lastFactsSubject string
	lastFactsAt      time.Time

	setEntityPropertyErrByKey map[string]error
	propertyCalls             []buildToolsIncludesDynamicToolsAndMetadataPropertyCall

	assertFactErrBySubject map[string]error
	assertFactCalls        []buildToolsIncludesDynamicToolsAndMetadataAssertionCall
	assertFactResult       *ontology.AssertionResult

	retractFactErr error

	openConflictsErr error

	mergeResult      *ontology.MergeResult
	mergeEntitiesErr error

	promoteTypeErr        error
	promoteTypeCalls      []buildToolsIncludesDynamicToolsAndMetadataPromotionCall
	promotePredicateErr   error
	promotePredicateCalls []buildToolsIncludesDynamicToolsAndMetadataPromotionCall

	schemaHealthResult *ontology.SchemaHealthReport
	schemaHealthErr    error

	typeUsageResult   *ontology.TypeUsageInfo
	typeUsageErr      error
	lastTypeUsageName string
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) ListTypes(context.Context) ([]ontology.ObjectType, error) {
	if s.listTypesErr != nil {
		return nil, s.listTypesErr
	}
	return append([]ontology.ObjectType(nil), s.listTypesResult...), nil
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) GetType(context.Context, string) (*ontology.ObjectType, error) {
	if s.getTypeErr != nil {
		return nil, s.getTypeErr
	}
	return s.getTypeResult, nil
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) ListPredicates(context.Context) ([]ontology.PredicateDefinition, error) {
	if s.listPredicatesErr != nil {
		return nil, s.listPredicatesErr
	}
	return append([]ontology.PredicateDefinition(nil), s.listPredicatesResult...), nil
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) ListActions(context.Context) ([]ontology.ActionSummary, error) {
	s.listActionsCalls++
	if s.listActionsErr != nil {
		return nil, s.listActionsErr
	}
	return s.listActionsResult, nil
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) ExecuteAction(_ context.Context, actionName string, params map[string]string) (*ontology.ActionResult, error) {
	s.executeActionCalls++
	s.lastActionName = actionName
	s.lastActionParams = make(map[string]string, len(params))
	for k, v := range params {
		s.lastActionParams[k] = v
	}
	if s.executeActionErr != nil {
		return nil, s.executeActionErr
	}
	return s.executeActionResult, nil
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) QueryEntities(_ context.Context, q ontology.PropertyQuery) ([]ontology.EntityResult, error) {
	s.queryEntitiesCalls++
	s.lastQuery = q
	if s.queryEntitiesErr != nil {
		return nil, s.queryEntitiesErr
	}
	return append([]ontology.EntityResult(nil), s.queryEntitiesResult...), nil
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) GetEntity(_ context.Context, entityID string) (*ontology.EntityResult, error) {
	s.lastEntityID = entityID
	if s.getEntityErr != nil {
		return nil, s.getEntityErr
	}
	return s.getEntityResult, nil
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) FactsAt(_ context.Context, subject string, validAt time.Time) ([]graph.Triple, error) {
	s.lastFactsSubject = subject
	s.lastFactsAt = validAt
	if s.factsAtErr != nil {
		return nil, s.factsAtErr
	}
	return append([]graph.Triple(nil), s.factsAtResult...), nil
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) SetEntityProperty(_ context.Context, entityID, entityType, property, value string) error {
	s.propertyCalls = append(s.propertyCalls, buildToolsIncludesDynamicToolsAndMetadataPropertyCall{
		entityID:   entityID,
		entityType: entityType,
		property:   property,
		value:      value,
	})
	if s.setEntityPropertyErrByKey != nil {
		if err := s.setEntityPropertyErrByKey[entityID+"."+property]; err != nil {
			return err
		}
	}
	return nil
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) AssertFact(_ context.Context, input ontology.AssertionInput) (*ontology.AssertionResult, error) {
	s.assertFactCalls = append(s.assertFactCalls, buildToolsIncludesDynamicToolsAndMetadataAssertionCall{
		subject:     input.Triple.Subject,
		subjectType: input.Triple.SubjectType,
		predicate:   input.Triple.Predicate,
		object:      input.Triple.Object,
		objectType:  input.Triple.ObjectType,
		source:      input.Source,
		confidence:  input.Confidence,
	})
	if s.assertFactErrBySubject != nil {
		if err := s.assertFactErrBySubject[input.Triple.Subject]; err != nil {
			return nil, err
		}
	}
	if s.assertFactResult != nil {
		return s.assertFactResult, nil
	}
	return &ontology.AssertionResult{Stored: true, Message: "stored"}, nil
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) RetractFact(context.Context, string, string, string, string) error {
	return s.retractFactErr
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) OpenConflicts(context.Context) ([]ontology.Conflict, error) {
	if s.openConflictsErr != nil {
		return nil, s.openConflictsErr
	}
	return nil, nil
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) MergeEntities(context.Context, string, string) (*ontology.MergeResult, error) {
	if s.mergeEntitiesErr != nil {
		return nil, s.mergeEntitiesErr
	}
	return s.mergeResult, nil
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) PromoteType(_ context.Context, typeName string, targetStatus ontology.SchemaStatus, reason string) error {
	s.promoteTypeCalls = append(s.promoteTypeCalls, buildToolsIncludesDynamicToolsAndMetadataPromotionCall{
		name:   typeName,
		target: targetStatus,
		reason: reason,
	})
	return s.promoteTypeErr
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) PromotePredicate(_ context.Context, predName string, targetStatus ontology.SchemaStatus, reason string) error {
	s.promotePredicateCalls = append(s.promotePredicateCalls, buildToolsIncludesDynamicToolsAndMetadataPromotionCall{
		name:   predName,
		target: targetStatus,
		reason: reason,
	})
	return s.promotePredicateErr
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) SchemaHealth(context.Context) (*ontology.SchemaHealthReport, error) {
	if s.schemaHealthErr != nil {
		return nil, s.schemaHealthErr
	}
	return s.schemaHealthResult, nil
}

func (s *buildToolsIncludesDynamicToolsAndMetadataToolService) TypeUsage(_ context.Context, typeName string) (*ontology.TypeUsageInfo, error) {
	s.lastTypeUsageName = typeName
	if s.typeUsageErr != nil {
		return nil, s.typeUsageErr
	}
	return s.typeUsageResult, nil
}

type buildToolsIncludesDynamicToolsAndMetadataPropertyCall struct {
	entityID   string
	entityType string
	property   string
	value      string
}

type buildToolsIncludesDynamicToolsAndMetadataAssertionCall struct {
	subject     string
	subjectType string
	predicate   string
	object      string
	objectType  string
	source      string
	confidence  float64
}

type buildToolsIncludesDynamicToolsAndMetadataPromotionCall struct {
	name   string
	target ontology.SchemaStatus
	reason string
}
