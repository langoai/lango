package app

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/auditlog"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/langoai/lango/internal/knowledge"
)

func newApprovalFlowToolStore(t *testing.T) (*knowledge.Store, *ent.Client) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	return knowledge.NewStore(client, zap.NewNop().Sugar()), client
}

func TestBuildMetaTools_IncludesApproveArtifactRelease(t *testing.T) {
	store, _ := newApprovalFlowToolStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, nil)
	tool := findTool(tools, "approve_artifact_release")
	require.NotNil(t, tool)

	assert.Equal(t, "knowledge", tool.Capability.Category)
	assert.Equal(t, "approve_artifact_release", tool.Name)
	assert.False(t, tool.Capability.ReadOnly)

	props, _ := tool.Parameters["properties"].(map[string]interface{})
	_, hasArtifactLabel := props["artifact_label"]
	_, hasRequestedScope := props["requested_scope"]
	_, hasExportabilityState := props["exportability_state"]
	_, hasOverrideRequested := props["override_requested"]
	_, hasHighRisk := props["high_risk"]
	assert.True(t, hasArtifactLabel)
	assert.True(t, hasRequestedScope)
	assert.True(t, hasExportabilityState)
	assert.True(t, hasOverrideRequested)
	assert.True(t, hasHighRisk)

	required, _ := tool.Parameters["required"].([]string)
	assert.Contains(t, required, "artifact_label")
	assert.Contains(t, required, "requested_scope")
	assert.Contains(t, required, "exportability_state")
	assert.Contains(t, required, "override_requested")
	assert.Contains(t, required, "high_risk")
}

func TestApproveArtifactRelease_EscalatesNeedsHumanReview(t *testing.T) {
	store, client := newApprovalFlowToolStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, nil)
	tool := findTool(tools, "approve_artifact_release")
	require.NotNil(t, tool)

	ctx := context.Background()
	got, err := tool.Handler(ctx, map[string]interface{}{
		"artifact_label":      "artifact/needs-review",
		"requested_scope":     "artifact/needs-review",
		"exportability_state": "needs-human-review",
		"override_requested":  false,
		"high_risk":           false,
	})
	require.NoError(t, err)

	payload := got.(map[string]interface{})
	assert.Equal(t, "escalate", payload["decision"])
	assert.Equal(t, "policy_issue", payload["issue"])
	assert.Equal(t, "review", payload["settlement_hint"])
	assert.Contains(t, payload["reason"].(string), "human escalation")

	row, err := client.AuditLog.Query().
		Where(auditlog.ActionEQ(auditlog.ActionArtifactReleaseApproval), auditlog.TargetEQ("artifact:artifact/needs-review")).
		Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, "artifact_release_approval", string(row.Action))
	assert.Equal(t, "artifact:artifact/needs-review", row.Target)
	assert.Equal(t, "needs-human-review", row.Details["exportability_state"])
	assert.Equal(t, "escalate", row.Details["decision"])
}

func TestApproveArtifactRelease_SavesAuditRow(t *testing.T) {
	store, client := newApprovalFlowToolStore(t)
	tools := buildMetaTools(store, nil, nil, config.SkillConfig{}, nil)
	tool := findTool(tools, "approve_artifact_release")
	require.NotNil(t, tool)

	ctx := context.Background()
	_, err := tool.Handler(ctx, map[string]interface{}{
		"artifact_label":      "artifact/final-memo",
		"requested_scope":     "artifact/final-memo",
		"exportability_state": "exportable",
		"override_requested":  false,
		"high_risk":           false,
	})
	require.NoError(t, err)

	count, err := client.AuditLog.Query().
		Where(auditlog.ActionEQ(auditlog.ActionArtifactReleaseApproval)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
