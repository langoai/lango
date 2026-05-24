package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/disputehold"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/escrowadjudication"
	"github.com/langoai/lango/internal/escrowrefund"
	"github.com/langoai/lango/internal/escrowrelease"
	"github.com/langoai/lango/internal/partialsettlementexecution"
	"github.com/langoai/lango/internal/postadjudicationreplay"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/settlementexecution"
)

type fakeBuildMetaToolsWithRuntimesRuntimeToolCompositionEscrowExecutionRuntime struct {
	createErr error
	fundErr   error
}

func (f fakeBuildMetaToolsWithRuntimesRuntimeToolCompositionEscrowExecutionRuntime) Create(_ context.Context, _ escrow.CreateRequest) (*escrow.EscrowEntry, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &escrow.EscrowEntry{ID: "escrow-buildCatalogFromEntriesRegistersCategoriesAndTools1"}, nil
}

func (f fakeBuildMetaToolsWithRuntimesRuntimeToolCompositionEscrowExecutionRuntime) Fund(_ context.Context, id string) (*escrow.EscrowEntry, error) {
	if f.fundErr != nil {
		return nil, f.fundErr
	}
	return &escrow.EscrowEntry{ID: id}, nil
}

func TestBuildMetaToolsWithRuntimes_RuntimeToolComposition(t *testing.T) {
	t.Parallel()

	tools := buildMetaToolsWithRuntimes(
		nil,
		nil,
		nil,
		config.SkillConfig{},
		nil,
		receipts.NewStore(),
		fakeBuildMetaToolsWithRuntimesRuntimeToolCompositionEscrowExecutionRuntime{},
		&fakeSettlementExecutionRuntime{},
		nil,
		&fakeDisputeHoldRuntime{},
		&fakeEscrowReleaseRuntime{},
		&fakeEscrowRefundRuntime{},
		&fakeAdjudicationBackgroundDispatcher{},
	)

	assert.NotNil(t, findTool(tools, "execute_settlement"))
	assert.NotNil(t, findTool(tools, "execute_partial_settlement"), "settlement runtime should adapt to partial settlement")
	assert.NotNil(t, findTool(tools, "hold_escrow_for_dispute"))
	assert.NotNil(t, findTool(tools, "release_escrow_settlement"))
	assert.NotNil(t, findTool(tools, "refund_escrow_settlement"))
	assert.NotNil(t, findTool(tools, "execute_escrow_recommendation"))
	assert.NotNil(t, findTool(tools, "retry_post_adjudication_execution"))
	assert.NotNil(t, findTool(tools, "get_post_adjudication_execution_status"))
}

func TestReplayDispatcherAdapter_ErrorBranches(t *testing.T) {
	t.Parallel()

	req := postadjudicationreplay.BackgroundDispatchRequest{Prompt: "retry"}

	got, err := (replayDispatcherAdapter{}).Dispatch(context.Background(), req)
	require.Error(t, err)
	assert.Equal(t, postadjudicationreplay.BackgroundDispatchReceipt{}, got)
	assert.ErrorContains(t, err, "background manager is not configured")

	submitErr := errors.New("queue unavailable")
	got, err = (replayDispatcherAdapter{
		dispatcher: &fakeAdjudicationBackgroundDispatcher{err: submitErr},
	}).Dispatch(context.Background(), req)
	require.ErrorIs(t, err, submitErr)
	assert.Equal(t, postadjudicationreplay.BackgroundDispatchReceipt{}, got)
}

func TestRuntimeBackedToolHandlers_RejectNilReceiptStoreBeforeParams(t *testing.T) {
	t.Parallel()

	tools := buildMetaToolsWithRuntimes(
		nil,
		nil,
		nil,
		config.SkillConfig{},
		nil,
		nil,
		fakeBuildMetaToolsWithRuntimesRuntimeToolCompositionEscrowExecutionRuntime{},
		&fakeSettlementExecutionRuntime{},
		&fakePartialSettlementExecutionRuntime{},
		&fakeDisputeHoldRuntime{},
		&fakeEscrowReleaseRuntime{},
		&fakeEscrowRefundRuntime{},
	)

	for _, name := range []string{
		"execute_settlement",
		"execute_partial_settlement",
		"hold_escrow_for_dispute",
		"release_escrow_settlement",
		"refund_escrow_settlement",
		"execute_escrow_recommendation",
	} {
		t.Run(name, func(t *testing.T) {
			tool := findTool(tools, name)
			require.NotNil(t, tool)

			got, err := tool.Handler(context.Background(), map[string]interface{}{})
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, "receipts store dependency is not configured")
		})
	}
}

func TestCreateDisputeReadyReceipt_ReturnsLinkedReceiptIDs(t *testing.T) {
	t.Parallel()

	store := receipts.NewStore()
	got, err := createDisputeReadyReceipt(context.Background(), store, receipts.CreateSubmissionInput{
		TransactionID:       "tx-buildCatalogFromEntriesRegistersCategoriesAndTools1-dispute-ready",
		ArtifactLabel:       "artifact/buildCatalogFromEntriesRegistersCategoriesAndTools1",
		PayloadHash:         "hash-buildCatalogFromEntriesRegistersCategoriesAndTools1",
		SourceLineageDigest: "lineage-buildCatalogFromEntriesRegistersCategoriesAndTools1",
	})
	require.NoError(t, err)

	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	submissionReceiptID := payload["submission_receipt_id"].(string)
	transactionReceiptID := payload["transaction_receipt_id"].(string)
	currentSubmissionReceiptID := payload["current_submission_receipt_id"].(string)
	assert.NotEmpty(t, submissionReceiptID)
	assert.NotEmpty(t, transactionReceiptID)
	assert.Equal(t, submissionReceiptID, currentSubmissionReceiptID)

	transaction, err := store.GetTransactionReceipt(context.Background(), transactionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, submissionReceiptID, transaction.CurrentSubmissionReceiptID)
}

func TestReceiptConstructors_MapBranchSpecificFields(t *testing.T) {
	t.Parallel()

	settlementReceipt := newExecuteSettlementReceipt(settlementexecution.Result{
		TransactionReceiptID:        "tx-settle",
		SubmissionReceiptID:         "sub-settle",
		SettlementProgressionStatus: receipts.SettlementProgressionSettled,
		ResolvedAmount:              "1.25",
		RuntimeReference:            "payment-settle",
	})
	assert.Equal(t, executeSettlementReceipt{
		TransactionReceiptID:        "tx-settle",
		SubmissionReceiptID:         "sub-settle",
		SettlementProgressionStatus: "settled",
		ResolvedAmount:              "1.25",
		RuntimeReference:            "payment-settle",
	}, settlementReceipt)

	partialReceipt := newExecutePartialSettlementReceipt(partialsettlementexecution.Result{
		TransactionReceiptID:        "tx-partial",
		SubmissionReceiptID:         "sub-partial",
		SettlementProgressionStatus: receipts.SettlementProgressionPartiallySettled,
		ExecutedAmount:              "0.40",
		RemainingAmount:             "0.60",
		RuntimeReference:            "payment-partial",
	})
	assert.Equal(t, executePartialSettlementReceipt{
		TransactionReceiptID:        "tx-partial",
		SubmissionReceiptID:         "sub-partial",
		SettlementProgressionStatus: "partially-settled",
		ExecutedAmount:              "0.40",
		RemainingAmount:             "0.60",
		RuntimeReference:            "payment-partial",
	}, partialReceipt)

	holdReceipt := newHoldEscrowForDisputeReceipt(disputehold.Result{
		TransactionReceiptID:        "tx-hold",
		SubmissionReceiptID:         "sub-hold",
		SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
		DisputeLifecycleStatus:      receipts.DisputeLifecycleHoldActive,
		EscrowReference:             "escrow-hold",
		RuntimeReference:            "hold-runtime",
	})
	assert.Equal(t, holdEscrowForDisputeReceipt{
		TransactionReceiptID:        "tx-hold",
		SubmissionReceiptID:         "sub-hold",
		SettlementProgressionStatus: "review-needed",
		DisputeLifecycleStatus:      "hold-active",
		EscrowReference:             "escrow-hold",
		RuntimeReference:            "hold-runtime",
	}, holdReceipt)

	releaseReceipt := newReleaseEscrowSettlementReceipt(escrowrelease.Result{
		TransactionReceiptID:        "tx-release",
		SubmissionReceiptID:         "sub-release",
		SettlementProgressionStatus: receipts.SettlementProgressionSettled,
		ResolvedAmount:              "1.00",
		RuntimeReference:            "release-runtime",
	})
	assert.Equal(t, releaseEscrowSettlementReceipt{
		TransactionReceiptID:        "tx-release",
		SubmissionReceiptID:         "sub-release",
		SettlementProgressionStatus: "settled",
		ResolvedAmount:              "1.00",
		RuntimeReference:            "release-runtime",
	}, releaseReceipt)

	refundReceipt := newRefundEscrowSettlementReceipt(escrowrefund.Result{
		TransactionReceiptID:        "tx-refund",
		SubmissionReceiptID:         "sub-refund",
		SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
		ResolvedAmount:              "0.75",
		RuntimeReference:            "refund-runtime",
	})
	assert.Equal(t, refundEscrowSettlementReceipt{
		TransactionReceiptID:        "tx-refund",
		SubmissionReceiptID:         "sub-refund",
		SettlementProgressionStatus: "review-needed",
		ResolvedAmount:              "0.75",
		RuntimeReference:            "refund-runtime",
	}, refundReceipt)

	adjudication := escrowadjudication.Result{
		TransactionReceiptID:        "tx-adjudicate",
		SubmissionReceiptID:         "sub-adjudicate",
		SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
		DisputeLifecycleStatus:      receipts.DisputeLifecycleHoldActive,
		EscrowReference:             "escrow-adjudicate",
		Outcome:                     escrowadjudication.OutcomeRefund,
	}
	assert.Equal(t, adjudicateEscrowDisputeReceipt{
		TransactionReceiptID:        "tx-adjudicate",
		SubmissionReceiptID:         "sub-adjudicate",
		SettlementProgressionStatus: "review-needed",
		DisputeLifecycleStatus:      "hold-active",
		EscrowReference:             "escrow-adjudicate",
		Outcome:                     "refund",
	}, newAdjudicateEscrowDisputeReceipt(adjudication))
	assert.Equal(t, &adjudicateEscrowDisputeBackgroundDispatchReceipt{
		Status:               "queued",
		TransactionReceiptID: "tx-adjudicate",
		SubmissionReceiptID:  "sub-adjudicate",
		EscrowReference:      "escrow-adjudicate",
		Outcome:              "refund",
		DispatchReference:    "task-buildCatalogFromEntriesRegistersCategoriesAndTools1",
	}, newAdjudicateEscrowDisputeBackgroundDispatchReceipt(adjudication, "task-buildCatalogFromEntriesRegistersCategoriesAndTools1"))

	retryReceipt := newRetryPostAdjudicationExecutionReceipt(postadjudicationreplay.Result{
		CanonicalAdjudication: postadjudicationreplay.CanonicalAdjudicationSnapshot{
			TransactionReceipt: receipts.TransactionReceipt{
				TransactionReceiptID:       "tx-retry",
				CurrentSubmissionReceiptID: "sub-retry",
				EscrowReference:            "escrow-retry",
				EscrowAdjudication:         receipts.EscrowAdjudicationRelease,
			},
			SubmissionReceipt: receipts.SubmissionReceipt{
				SubmissionReceiptID: "sub-retry",
			},
		},
		BackgroundDispatchReceipt: &postadjudicationreplay.BackgroundDispatchReceipt{
			Status:            "queued",
			DispatchReference: "task-retry",
		},
	})
	require.NotNil(t, retryReceipt.Dispatch)
	assert.Equal(t, retryPostAdjudicationExecutionReceipt{
		TransactionReceiptID: "tx-retry",
		SubmissionReceiptID:  "sub-retry",
		EscrowReference:      "escrow-retry",
		Outcome:              "release",
		Dispatch: &retryPostAdjudicationDispatchReceipt{
			Status:            "queued",
			DispatchReference: "task-retry",
		},
	}, retryReceipt)

	assert.Equal(t, &adjudicateEscrowDisputeExecutionReceipt{
		Branch:                      "release",
		Status:                      "settled-target",
		SettlementProgressionStatus: "settled",
		ResolvedAmount:              "1.00",
		RuntimeReference:            "release-runtime",
	}, newAdjudicationNestedExecutionReceiptFromRelease(escrowrelease.Result{
		Status:                      escrowrelease.ResultStatusSettledTarget,
		SettlementProgressionStatus: receipts.SettlementProgressionSettled,
		ResolvedAmount:              "1.00",
		RuntimeReference:            "release-runtime",
	}))
	assert.Equal(t, &adjudicateEscrowDisputeExecutionReceipt{
		Branch:                      "refund",
		Status:                      "refund-executed",
		SettlementProgressionStatus: "review-needed",
		ResolvedAmount:              "0.75",
		RuntimeReference:            "refund-runtime",
	}, newAdjudicationNestedExecutionReceiptFromRefund(escrowrefund.Result{
		Status:                      escrowrefund.ResultStatusRefundExecuted,
		SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
		ResolvedAmount:              "0.75",
		RuntimeReference:            "refund-runtime",
	}))
}
