package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/receipts"
)

func TestBuildMetaTools_IncludesCreateDisputeReadyReceipt(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore())
	tool := findTool(tools, "create_dispute_ready_receipt")
	require.NotNil(t, tool)

	assert.Equal(t, "knowledge", tool.Capability.Category)
	assert.Equal(t, agent.ActivityWrite, tool.Capability.Activity)
	assert.False(t, tool.Capability.ReadOnly)

	props, _ := tool.Parameters["properties"].(map[string]interface{})
	_, hasTransactionID := props["transaction_id"]
	_, hasArtifactLabel := props["artifact_label"]
	_, hasPayloadHash := props["payload_hash"]
	_, hasSourceLineageDigest := props["source_lineage_digest"]
	assert.True(t, hasTransactionID)
	assert.True(t, hasArtifactLabel)
	assert.True(t, hasPayloadHash)
	assert.True(t, hasSourceLineageDigest)

	required, _ := tool.Parameters["required"].([]string)
	assert.Contains(t, required, "transaction_id")
	assert.Contains(t, required, "artifact_label")
	assert.Contains(t, required, "payload_hash")
	assert.Contains(t, required, "source_lineage_digest")
}

func TestBuildMetaTools_IncludesCreateDisputeReadyReceiptWithoutStore(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, nil)
	tool := findTool(tools, "create_dispute_ready_receipt")
	require.NotNil(t, tool)
}

func TestCreateDisputeReadyReceipt_ValidationErrors(t *testing.T) {
	store := receipts.NewStore()
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store)
	tool := findTool(tools, "create_dispute_ready_receipt")
	require.NotNil(t, tool)

	cases := []struct {
		name      string
		params    map[string]interface{}
		wantError string
	}{
		{
			name: "missing transaction_id",
			params: map[string]interface{}{
				"artifact_label":        "artifact/one",
				"payload_hash":          "hash-1",
				"source_lineage_digest": "lineage-1",
			},
			wantError: "missing transaction_id parameter",
		},
		{
			name: "empty transaction_id",
			params: map[string]interface{}{
				"transaction_id":        "",
				"artifact_label":        "artifact/one",
				"payload_hash":          "hash-1",
				"source_lineage_digest": "lineage-1",
			},
			wantError: "missing transaction_id parameter",
		},
		{
			name: "missing artifact_label",
			params: map[string]interface{}{
				"transaction_id":        "tx-1",
				"payload_hash":          "hash-1",
				"source_lineage_digest": "lineage-1",
			},
			wantError: "missing artifact_label parameter",
		},
		{
			name: "empty artifact_label",
			params: map[string]interface{}{
				"transaction_id":        "tx-1",
				"artifact_label":        "",
				"payload_hash":          "hash-1",
				"source_lineage_digest": "lineage-1",
			},
			wantError: "missing artifact_label parameter",
		},
		{
			name: "missing payload_hash",
			params: map[string]interface{}{
				"transaction_id":        "tx-1",
				"artifact_label":        "artifact/one",
				"source_lineage_digest": "lineage-1",
			},
			wantError: "missing payload_hash parameter",
		},
		{
			name: "empty payload_hash",
			params: map[string]interface{}{
				"transaction_id":        "tx-1",
				"artifact_label":        "artifact/one",
				"payload_hash":          "",
				"source_lineage_digest": "lineage-1",
			},
			wantError: "missing payload_hash parameter",
		},
		{
			name: "missing source_lineage_digest",
			params: map[string]interface{}{
				"transaction_id": "tx-1",
				"artifact_label": "artifact/one",
				"payload_hash":   "hash-1",
			},
			wantError: "missing source_lineage_digest parameter",
		},
		{
			name: "empty source_lineage_digest",
			params: map[string]interface{}{
				"transaction_id":        "tx-1",
				"artifact_label":        "artifact/one",
				"payload_hash":          "hash-1",
				"source_lineage_digest": "",
			},
			wantError: "missing source_lineage_digest parameter",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := tool.Handler(context.Background(), tt.params)
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantError)
		})
	}
}

func TestCreateDisputeReadyReceipt_CreatesAndReusesTransactionIDs(t *testing.T) {
	store := receipts.NewStore()
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store)
	tool := findTool(tools, "create_dispute_ready_receipt")
	require.NotNil(t, tool)

	ctx := context.Background()

	first, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_id":        "tx-123",
		"artifact_label":        "artifact/one",
		"payload_hash":          "hash-1",
		"source_lineage_digest": "lineage-1",
	})
	require.NoError(t, err)

	firstPayload, ok := first.(map[string]interface{})
	require.True(t, ok)
	require.NotEmpty(t, firstPayload["submission_receipt_id"])
	require.NotEmpty(t, firstPayload["transaction_receipt_id"])
	require.NotEmpty(t, firstPayload["current_submission_receipt_id"])

	second, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_id":        "tx-123",
		"artifact_label":        "artifact/two",
		"payload_hash":          "hash-2",
		"source_lineage_digest": "lineage-2",
	})
	require.NoError(t, err)

	secondPayload, ok := second.(map[string]interface{})
	require.True(t, ok)
	require.NotEmpty(t, secondPayload["submission_receipt_id"])
	require.NotEmpty(t, secondPayload["transaction_receipt_id"])
	require.NotEmpty(t, secondPayload["current_submission_receipt_id"])

	assert.NotEqual(t, firstPayload["submission_receipt_id"], secondPayload["submission_receipt_id"])
	assert.Equal(t, firstPayload["transaction_receipt_id"], secondPayload["transaction_receipt_id"])
	assert.Equal(t, secondPayload["submission_receipt_id"], secondPayload["current_submission_receipt_id"])
}

func TestCreateDisputeReadyReceipt_ReportsMissingReceiptsDependency(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, nil)
	tool := findTool(tools, "create_dispute_ready_receipt")
	require.NotNil(t, tool)

	_, err := tool.Handler(context.Background(), map[string]interface{}{
		"transaction_id":        "tx-missing",
		"artifact_label":        "artifact/missing",
		"payload_hash":          "hash-missing",
		"source_lineage_digest": "lineage-missing",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "receipts store dependency is not configured")
}
