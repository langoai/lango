package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/receipts"
)

func TestBuildMetaTools_IncludesKnowledgeExchangeRuntimeTools(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore())

	openTool := findTool(tools, "open_knowledge_exchange_transaction")
	require.NotNil(t, openTool)
	assert.Equal(t, "knowledge", openTool.Capability.Category)
	assert.Equal(t, agent.ActivityWrite, openTool.Capability.Activity)

	openProps, _ := openTool.Parameters["properties"].(map[string]interface{})
	_, hasTransactionID := openProps["transaction_id"]
	_, hasCounterparty := openProps["counterparty"]
	_, hasRequestedScope := openProps["requested_scope"]
	_, hasPriceContext := openProps["price_context"]
	_, hasTrustContext := openProps["trust_context"]
	assert.True(t, hasTransactionID)
	assert.True(t, hasCounterparty)
	assert.True(t, hasRequestedScope)
	assert.True(t, hasPriceContext)
	assert.True(t, hasTrustContext)
	assert.Equal(t, []string{"transaction_id", "counterparty", "requested_scope"}, openTool.Parameters["required"])

	selectTool := findTool(tools, "select_knowledge_exchange_path")
	require.NotNil(t, selectTool)
	assert.Equal(t, "knowledge", selectTool.Capability.Category)
	assert.Equal(t, agent.ActivityWrite, selectTool.Capability.Activity)

	selectProps, _ := selectTool.Parameters["properties"].(map[string]interface{})
	_, hasTransactionReceiptID := selectProps["transaction_receipt_id"]
	assert.True(t, hasTransactionReceiptID)
	assert.Equal(t, []string{"transaction_receipt_id"}, selectTool.Parameters["required"])
}

func TestOpenKnowledgeExchangeTransaction_ReturnsStableReceiptPayload(t *testing.T) {
	store := receipts.NewStore()
	tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store), "open_knowledge_exchange_transaction")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"transaction_id":  "deal-tool-open-1",
		"counterparty":    "did:lango:peer-1",
		"requested_scope": "artifact/research-note",
		"price_context":   "quote:0.50-usdc",
		"trust_context":   "trust:0.71",
	})
	require.NoError(t, err)

	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, payload["transaction_receipt_id"])
	assert.Equal(t, string(receipts.RuntimeStatusOpened), payload["runtime_status"])

	stored, err := store.GetTransactionReceipt(context.Background(), payload["transaction_receipt_id"].(string))
	require.NoError(t, err)
	assert.Equal(t, "did:lango:peer-1", stored.Counterparty)
	assert.Equal(t, "artifact/research-note", stored.RequestedScope)
	assert.Equal(t, "quote:0.50-usdc", stored.PriceContext)
	assert.Equal(t, "trust:0.71", stored.TrustContext)
	assert.Equal(t, receipts.RuntimeStatusOpened, stored.KnowledgeExchangeRuntimeStatus)
}

func TestOpenKnowledgeExchangeTransaction_RequiresReceiptStore(t *testing.T) {
	tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, nil), "open_knowledge_exchange_transaction")
	require.NotNil(t, tool)

	_, err := tool.Handler(context.Background(), map[string]interface{}{
		"transaction_id":  "deal-tool-open-missing-store",
		"counterparty":    "did:lango:peer-2",
		"requested_scope": "artifact/research-note",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "receipts store dependency is not configured")
}

func TestOpenKnowledgeExchangeTransaction_SelectKnowledgeExchangePath_ReturnsBranchSelection(t *testing.T) {
	ctx := context.Background()
	store := receipts.NewStore()
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store)
	selectTool := findTool(tools, "select_knowledge_exchange_path")
	require.NotNil(t, selectTool)

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-tool-select-1",
		Counterparty:   "did:lango:peer-3",
		RequestedScope: "artifact/design-draft",
		PriceContext:   "quote:1.00-usdc",
		TrustContext:   "trust:0.92",
	})
	require.NoError(t, err)

	submission, _, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "deal-tool-select-1",
		ArtifactLabel:       "artifact/design-draft-v1",
		PayloadHash:         "hash-deal-tool-select-1",
		SourceLineageDigest: "lineage-deal-tool-select-1",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "approved",
		SuggestedMode: paymentapproval.ModePrepay,
	})
	require.NoError(t, err)

	got, err := selectTool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.NoError(t, err)

	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, tx.TransactionReceiptID, payload["transaction_receipt_id"])
	assert.Equal(t, submission.SubmissionReceiptID, payload["current_submission_receipt_id"])
	assert.Equal(t, "prepay", payload["selected_path"])

	stored, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, receipts.RuntimeStatusPaymentApproved, stored.KnowledgeExchangeRuntimeStatus)
}
