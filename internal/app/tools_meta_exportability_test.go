package app

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/knowledge"
	"go.uber.org/zap"
)

func newExportabilityToolStore(t *testing.T) *knowledge.Store {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	return knowledge.NewStore(client, zap.NewNop().Sugar())
}

func TestBuildMetaTools_IncludesEvaluateExportability(t *testing.T) {
	store := newExportabilityToolStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, config.DefaultConfig())
	tool := findTool(tools, "evaluate_exportability")
	require.NotNil(t, tool)

	props, _ := tool.Parameters["properties"].(map[string]interface{})
	_, hasArtifactLabel := props["artifact_label"]
	_, hasSourceKeys := props["source_keys"]
	_, hasStage := props["stage"]
	assert.True(t, hasArtifactLabel)
	assert.True(t, hasSourceKeys)
	assert.True(t, hasStage)

	required, _ := tool.Parameters["required"].([]string)
	assert.Contains(t, required, "artifact_label")
	assert.Contains(t, required, "source_keys")
	assert.Contains(t, required, "stage")
}

func TestEvaluateExportability_BlocksPrivateKnowledgeSource(t *testing.T) {
	store := newExportabilityToolStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, config.DefaultConfig())
	saveTool := findTool(tools, "save_knowledge")
	require.NotNil(t, saveTool)
	evalTool := findTool(tools, "evaluate_exportability")
	require.NotNil(t, evalTool)

	ctx := context.Background()
	_, err := saveTool.Handler(ctx, map[string]interface{}{
		"key":          "private-source",
		"category":     "fact",
		"content":      "classified source content",
		"source":       "agent",
		"source_class": "private-confidential",
		"asset_label":  "knowledge/private-source",
	})
	require.NoError(t, err)

	got, err := evalTool.Handler(ctx, map[string]interface{}{
		"artifact_label": "artifact/private-doc",
		"source_keys":    []interface{}{"private-source"},
		"stage":          "final",
	})
	require.NoError(t, err)

	payload := got.(map[string]interface{})
	assert.Equal(t, "artifact/private-doc", payload["artifact_label"])
	assert.Equal(t, "final", payload["stage"])
	assert.Equal(t, "blocked", payload["state"])
	assert.Equal(t, "blocked_private_source", payload["policy_code"])
	assert.NotEmpty(t, payload["explanation"])

	lineage, ok := payload["lineage"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, lineage, 1)
	assert.Equal(t, "private-source", lineage[0]["asset_id"])
	assert.Equal(t, "knowledge/private-source", lineage[0]["asset_label"])
	assert.Equal(t, "private-confidential", lineage[0]["class"])
}
