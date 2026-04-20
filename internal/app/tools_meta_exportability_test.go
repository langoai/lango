package app

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/auditlog"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/knowledge"
)

func newExportabilityToolStore(t *testing.T) (*knowledge.Store, *ent.Client) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	return knowledge.NewStore(client, zap.NewNop().Sugar()), client
}

func newExportabilityToolConfig(enabled bool) *config.Config {
	cfg := config.DefaultConfig()
	cfg.Security.Exportability.Enabled = enabled
	return cfg
}

func TestBuildMetaTools_IncludesEvaluateExportability(t *testing.T) {
	store, _ := newExportabilityToolStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, newExportabilityToolConfig(true), nil)
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

func TestBuildMetaTools_EvaluateExportabilityCapabilityMetadata(t *testing.T) {
	store, _ := newExportabilityToolStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, newExportabilityToolConfig(true), nil)
	tool := findTool(tools, "evaluate_exportability")
	require.NotNil(t, tool)

	assert.Equal(t, "knowledge", tool.Capability.Category)
	assert.Equal(t, agent.ActivityWrite, tool.Capability.Activity)
	assert.False(t, tool.Capability.ReadOnly)
}

func TestEvaluateExportability_SavesAuditRow(t *testing.T) {
	store, client := newExportabilityToolStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, newExportabilityToolConfig(true), nil)
	saveTool := findTool(tools, "save_knowledge")
	require.NotNil(t, saveTool)
	evalTool := findTool(tools, "evaluate_exportability")
	require.NotNil(t, evalTool)

	ctx := context.Background()
	_, err := saveTool.Handler(ctx, map[string]interface{}{
		"key":          "public-source",
		"category":     "fact",
		"content":      "public content",
		"source":       "agent",
		"source_class": "public",
		"asset_label":  "knowledge/public-source",
	})
	require.NoError(t, err)

	got, err := evalTool.Handler(ctx, map[string]interface{}{
		"artifact_label": "artifact/public-doc",
		"source_keys":    []interface{}{"public-source"},
		"stage":          "final",
	})
	require.NoError(t, err)

	payload := got.(map[string]interface{})
	assert.Equal(t, "artifact/public-doc", payload["artifact_label"])
	assert.Equal(t, "final", payload["stage"])
	assert.Equal(t, "exportable", payload["state"])

	row, err := client.AuditLog.Query().
		Where(auditlog.ActionEQ(auditlog.ActionExportabilityDecision), auditlog.TargetEQ("artifact:artifact/public-doc")).
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, "exportability_decision", string(row.Action))
	assert.Equal(t, "agent", row.Actor)
	assert.Equal(t, "artifact:artifact/public-doc", row.Target)

	assert.Equal(t, "artifact/public-doc", row.Details["artifact_label"])
}

func TestEvaluateExportability_InvalidStage(t *testing.T) {
	store, _ := newExportabilityToolStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, newExportabilityToolConfig(true), nil)
	saveTool := findTool(tools, "save_knowledge")
	require.NotNil(t, saveTool)
	evalTool := findTool(tools, "evaluate_exportability")
	require.NotNil(t, evalTool)

	ctx := context.Background()
	_, err := saveTool.Handler(ctx, map[string]interface{}{
		"key":          "stage-source",
		"category":     "fact",
		"content":      "stage content",
		"source_class": "public",
	})
	require.NoError(t, err)

	_, err = evalTool.Handler(ctx, map[string]interface{}{
		"artifact_label": "artifact/stage-doc",
		"source_keys":    []interface{}{"stage-source"},
		"stage":          "review",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid stage")
}

func TestEvaluateExportability_MissingSourceKey(t *testing.T) {
	store, _ := newExportabilityToolStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, newExportabilityToolConfig(true), nil)
	evalTool := findTool(tools, "evaluate_exportability")
	require.NotNil(t, evalTool)

	_, err := evalTool.Handler(context.Background(), map[string]interface{}{
		"artifact_label": "artifact/missing-source",
		"source_keys":    []interface{}{"missing-source"},
		"stage":          "final",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load source knowledge")
}

func TestEvaluateExportability_PolicyDisabled(t *testing.T) {
	store, _ := newExportabilityToolStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, newExportabilityToolConfig(false), nil)
	saveTool := findTool(tools, "save_knowledge")
	require.NotNil(t, saveTool)
	evalTool := findTool(tools, "evaluate_exportability")
	require.NotNil(t, evalTool)

	ctx := context.Background()
	_, err := saveTool.Handler(ctx, map[string]interface{}{
		"key":          "disabled-source",
		"category":     "fact",
		"content":      "disabled content",
		"source_class": "public",
	})
	require.NoError(t, err)

	got, err := evalTool.Handler(ctx, map[string]interface{}{
		"artifact_label": "artifact/disabled-doc",
		"source_keys":    []interface{}{"disabled-source"},
		"stage":          "draft",
	})
	require.NoError(t, err)

	payload := got.(map[string]interface{})
	assert.Equal(t, "needs-human-review", payload["state"])
	assert.Equal(t, "review_policy_disabled", payload["policy_code"])
}
