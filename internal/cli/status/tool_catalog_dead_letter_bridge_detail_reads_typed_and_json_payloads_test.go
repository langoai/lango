package status

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/postadjudicationstatus"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/toolcatalog"
)

func TestToolCatalogDeadLetterBridgeDetailReadsTypedAndJSONPayloads(t *testing.T) {
	tests := []struct {
		name    string
		raw     interface{}
		wantTx  string
		wantCan bool
	}{
		{
			name: "typed status",
			raw: postadjudicationstatus.TransactionStatus{
				CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
					TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-typed"},
				},
				CanRetry: true,
			},
			wantTx:  "tx-typed",
			wantCan: true,
		},
		{
			name: "json shaped status",
			raw: map[string]interface{}{
				"canonical_snapshot": map[string]interface{}{
					"transaction_receipt": map[string]interface{}{
						"transaction_receipt_id": "tx-json",
					},
				},
				"retry_dead_letter_summary": map[string]interface{}{
					"latest_status_subtype": "dead-lettered",
				},
				"can_retry": true,
			},
			wantTx:  "tx-json",
			wantCan: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			catalog := toolcatalog.New()
			var gotParams map[string]interface{}
			catalog.Register("status", []*agent.Tool{{
				Name: "get_post_adjudication_execution_status",
				Handler: func(_ context.Context, params map[string]interface{}) (interface{}, error) {
					gotParams = params
					return tt.raw, nil
				},
			}})

			bridge := NewToolCatalogDeadLetterBridge(catalog)
			got, err := bridge.Detail(context.Background(), "  tx-input  ")

			require.NoError(t, err)
			assert.Equal(t, "tx-input", gotParams["transaction_receipt_id"])
			assert.Equal(t, tt.wantTx, got.CanonicalSnapshot.TransactionReceipt.TransactionReceiptID)
			assert.Equal(t, tt.wantCan, got.CanRetry)
		})
	}
}

func TestToolCatalogDeadLetterBridgeDetailReportsUnavailableAndInvalidPayloads(t *testing.T) {
	tests := []struct {
		name    string
		bridge  *toolCatalogDeadLetterBridge
		wantErr string
	}{
		{
			name:    "nil catalog",
			bridge:  &toolCatalogDeadLetterBridge{},
			wantErr: "dead-letter tool catalog is not configured",
		},
		{
			name:    "missing detail tool",
			bridge:  &toolCatalogDeadLetterBridge{catalog: toolcatalog.New()},
			wantErr: "dead-letter detail tool is not available",
		},
		{
			name: "handler error",
			bridge: func() *toolCatalogDeadLetterBridge {
				catalog := toolcatalog.New()
				catalog.Register("status", []*agent.Tool{{
					Name: "get_post_adjudication_execution_status",
					Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
						return nil, errors.New("detail backend down")
					},
				}})
				return &toolCatalogDeadLetterBridge{catalog: catalog}
			}(),
			wantErr: "detail backend down",
		},
		{
			name: "marshal failure",
			bridge: func() *toolCatalogDeadLetterBridge {
				catalog := toolcatalog.New()
				catalog.Register("status", []*agent.Tool{{
					Name: "get_post_adjudication_execution_status",
					Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
						return map[string]interface{}{"bad": make(chan int)}, nil
					},
				}})
				return &toolCatalogDeadLetterBridge{catalog: catalog}
			}(),
			wantErr: "unsupported type",
		},
		{
			name: "unmarshal failure",
			bridge: func() *toolCatalogDeadLetterBridge {
				catalog := toolcatalog.New()
				catalog.Register("status", []*agent.Tool{{
					Name: "get_post_adjudication_execution_status",
					Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
						return map[string]interface{}{"can_retry": "not-a-bool"}, nil
					},
				}})
				return &toolCatalogDeadLetterBridge{catalog: catalog}
			}(),
			wantErr: "cannot unmarshal string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.bridge.Detail(context.Background(), "tx-1")

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestToolCatalogDeadLetterBridgeRetryTrimsIDAndPreservesPrincipal(t *testing.T) {
	catalog := toolcatalog.New()
	var gotParams map[string]interface{}
	var gotPrincipal string
	catalog.Register("status", []*agent.Tool{{
		Name: "retry_post_adjudication_execution",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			gotParams = params
			gotPrincipal = ctxkeys.PrincipalFromContext(ctx)
			return map[string]interface{}{"result": "accepted"}, nil
		},
	}})

	bridge := NewToolCatalogDeadLetterBridge(catalog)
	ctx := ctxkeys.WithPrincipal(context.Background(), "operator:contextHelpersRoundTripAgentAndChildSession1")
	err := bridge.Retry(ctx, "  tx-retry  ")

	require.NoError(t, err)
	assert.Equal(t, "tx-retry", gotParams["transaction_receipt_id"])
	assert.Equal(t, "operator:contextHelpersRoundTripAgentAndChildSession1", gotPrincipal)
}

func TestToolCatalogDeadLetterBridgeRetryReportsUnavailableAndHandlerErrors(t *testing.T) {
	handlerErr := errors.New("retry queue down")
	tests := []struct {
		name    string
		bridge  *toolCatalogDeadLetterBridge
		wantErr string
	}{
		{
			name:    "nil catalog",
			bridge:  &toolCatalogDeadLetterBridge{},
			wantErr: "dead-letter tool catalog is not configured",
		},
		{
			name:    "missing retry tool",
			bridge:  &toolCatalogDeadLetterBridge{catalog: toolcatalog.New()},
			wantErr: "dead-letter retry tool is not available",
		},
		{
			name: "handler error",
			bridge: func() *toolCatalogDeadLetterBridge {
				catalog := toolcatalog.New()
				catalog.Register("status", []*agent.Tool{{
					Name: "retry_post_adjudication_execution",
					Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
						return nil, handlerErr
					},
				}})
				return &toolCatalogDeadLetterBridge{catalog: catalog}
			}(),
			wantErr: "retry queue down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bridge.Retry(context.Background(), "tx-1")

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
			if tt.name == "handler error" {
				assert.ErrorIs(t, err, handlerErr)
			}
		})
	}
}

func TestOptionalIntAcceptsNumericToolPayloadTypes(t *testing.T) {
	payload := map[string]interface{}{
		"int":     3,
		"int32":   int32(4),
		"int64":   int64(5),
		"float64": float64(6),
		"string":  "7",
	}

	assert.Equal(t, 3, optionalInt(payload, "int"))
	assert.Equal(t, 4, optionalInt(payload, "int32"))
	assert.Equal(t, 5, optionalInt(payload, "int64"))
	assert.Equal(t, 6, optionalInt(payload, "float64"))
	assert.Equal(t, 0, optionalInt(payload, "string"))
	assert.Equal(t, 0, optionalInt(payload, "missing"))
}

func TestDeadLetterRetryFollowUpChangedDetectsStatusAndTaskTransitions(t *testing.T) {
	baseline := postadjudicationstatus.TransactionStatus{
		IsDeadLettered: true,
		CanRetry:       true,
		RetryDeadLetterSummary: postadjudicationstatus.RetryDeadLetterSummary{
			LatestStatusSubtype:       "dead-lettered",
			LatestStatusSubtypeFamily: "dead-letter",
			LatestRetryAttempt:        3,
			LatestDispatchReference:   "dispatch-1",
		},
		LatestBackgroundTask: &postadjudicationstatus.BackgroundTaskBridge{
			TaskID:       "task-1",
			Status:       "queued",
			AttemptCount: 1,
			NextRetryAt:  "2026-05-19T00:00:00Z",
		},
	}
	matchingFollowUp := &deadLetterRetryFollowUp{
		IsDeadLettered:            true,
		CanRetry:                  true,
		LatestStatusSubtype:       "dead-lettered",
		LatestStatusSubtypeFamily: "dead-letter",
		LatestRetryAttempt:        3,
		LatestDispatchReference:   "dispatch-1",
		BackgroundTask: &postadjudicationstatus.BackgroundTaskBridge{
			TaskID:       "task-1",
			Status:       "queued",
			AttemptCount: 1,
			NextRetryAt:  "2026-05-19T00:00:00Z",
		},
	}

	tests := []struct {
		name     string
		mutate   func(*deadLetterRetryFollowUp)
		wantDiff bool
	}{
		{name: "nil follow-up", mutate: nil, wantDiff: false},
		{name: "unchanged", mutate: func(*deadLetterRetryFollowUp) {}, wantDiff: false},
		{name: "dead-letter flag changed", mutate: func(f *deadLetterRetryFollowUp) { f.IsDeadLettered = false }, wantDiff: true},
		{name: "can retry changed", mutate: func(f *deadLetterRetryFollowUp) { f.CanRetry = false }, wantDiff: true},
		{name: "subtype changed", mutate: func(f *deadLetterRetryFollowUp) { f.LatestStatusSubtype = "retry-scheduled" }, wantDiff: true},
		{name: "family changed", mutate: func(f *deadLetterRetryFollowUp) { f.LatestStatusSubtypeFamily = "retry" }, wantDiff: true},
		{name: "attempt changed", mutate: func(f *deadLetterRetryFollowUp) { f.LatestRetryAttempt = 4 }, wantDiff: true},
		{name: "dispatch changed", mutate: func(f *deadLetterRetryFollowUp) { f.LatestDispatchReference = "dispatch-2" }, wantDiff: true},
		{name: "background task removed", mutate: func(f *deadLetterRetryFollowUp) { f.BackgroundTask = nil }, wantDiff: true},
		{name: "background task status changed", mutate: func(f *deadLetterRetryFollowUp) { f.BackgroundTask.Status = "running" }, wantDiff: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mutate == nil {
				assert.False(t, deadLetterRetryFollowUpChanged(baseline, nil))
				return
			}
			followUp := *matchingFollowUp
			task := *matchingFollowUp.BackgroundTask
			followUp.BackgroundTask = &task
			tt.mutate(&followUp)

			assert.Equal(t, tt.wantDiff, deadLetterRetryFollowUpChanged(baseline, &followUp))
		})
	}
}

func TestCollectDeadLetterRetryFollowUpHandlesErrorCancelAndTimeout(t *testing.T) {
	baseline := postadjudicationstatus.TransactionStatus{
		IsDeadLettered: true,
		CanRetry:       true,
		RetryDeadLetterSummary: postadjudicationstatus.RetryDeadLetterSummary{
			LatestStatusSubtype:       "dead-lettered",
			LatestStatusSubtypeFamily: "dead-letter",
		},
	}

	t.Run("immediate detail error has no poll count", func(t *testing.T) {
		bridge := &fakeDeadLetterBridge{detailErr: errors.New("detail unavailable")}

		followUp, pollCount, timedOut, err := collectDeadLetterRetryFollowUp(
			context.Background(),
			bridge,
			"tx-1",
			baseline,
			false,
			time.Millisecond,
			time.Millisecond,
		)

		require.Error(t, err)
		assert.Nil(t, followUp)
		assert.Equal(t, 0, pollCount)
		assert.False(t, timedOut)
	})

	t.Run("canceled wait returns last unchanged follow-up", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		bridge := &toolCatalogDeadLetterBridgeDetailReadsTypedAndJsonPayloadsCancelAfterFirstDetailBridge{
			detail: baseline,
			cancel: cancel,
		}

		followUp, pollCount, timedOut, err := collectDeadLetterRetryFollowUp(
			ctx,
			bridge,
			"tx-1",
			baseline,
			true,
			time.Hour,
			time.Hour,
		)

		require.ErrorIs(t, err, context.Canceled)
		require.NotNil(t, followUp)
		assert.Equal(t, 1, pollCount)
		assert.False(t, timedOut)
	})

	t.Run("wait timeout reports unchanged follow-up", func(t *testing.T) {
		bridge := &fakeDeadLetterBridge{detail: baseline}

		followUp, pollCount, timedOut, err := collectDeadLetterRetryFollowUp(
			context.Background(),
			bridge,
			"tx-1",
			baseline,
			true,
			time.Millisecond,
			2*time.Millisecond,
		)

		require.NoError(t, err)
		require.NotNil(t, followUp)
		assert.GreaterOrEqual(t, pollCount, 1)
		assert.True(t, timedOut)
	})
}

type toolCatalogDeadLetterBridgeDetailReadsTypedAndJsonPayloadsCancelAfterFirstDetailBridge struct {
	detail postadjudicationstatus.TransactionStatus
	cancel context.CancelFunc
	calls  int
}

func (b *toolCatalogDeadLetterBridgeDetailReadsTypedAndJsonPayloadsCancelAfterFirstDetailBridge) List(context.Context, DeadLetterListOptions) (DeadLetterListPage, error) {
	return DeadLetterListPage{}, nil
}

func (b *toolCatalogDeadLetterBridgeDetailReadsTypedAndJsonPayloadsCancelAfterFirstDetailBridge) Detail(ctx context.Context, transactionReceiptID string) (postadjudicationstatus.TransactionStatus, error) {
	b.calls++
	if err := ctx.Err(); err != nil {
		return postadjudicationstatus.TransactionStatus{}, err
	}
	if b.calls == 1 {
		b.cancel()
	}
	return b.detail, nil
}

func (b *toolCatalogDeadLetterBridgeDetailReadsTypedAndJsonPayloadsCancelAfterFirstDetailBridge) Retry(context.Context, string) error {
	return nil
}

func TestDeadLetterRetryCmdJSONIncludesSanitizedFollowUpError(t *testing.T) {
	bridge := &fakeDeadLetterBridge{
		detailSeq: []postadjudicationstatus.TransactionStatus{{
			CanRetry:       true,
			IsDeadLettered: true,
		}},
		detailErrSeq: []error{
			nil,
			errors.New("follow-up \x1b[31munavailable\n"),
		},
	}
	cmd := newDeadLetterCmd(func() (DeadLetterBridge, func(), error) {
		return bridge, func() {}, nil
	})

	out, err := executeCommand(t, cmd, "retry", "tx-1", "--yes", "--output", "json")

	require.NoError(t, err)
	var got deadLetterRetryResult
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Equal(t, "accepted", got.Result)
	assert.Equal(t, "follow-up unavailable", got.FollowUpError)
	assert.Equal(t, 0, got.PollCount)
	assert.False(t, got.TimedOut)
	assert.NotContains(t, out, "\u001b")
	assert.NotContains(t, out, "\\u001b")
}

func TestSanitizeDeadLetterBacklogEntryCoversNestedBreakdown(t *testing.T) {
	got := sanitizeDeadLetterBacklogEntryJSON(postadjudicationstatus.DeadLetterBacklogEntry{
		TransactionGlobalAnyMatchFamilies: []string{"retry\x1b[31m\n", "  ", "dead-letter"},
		TransactionGlobalDominantFamily:   "dead-\x1b[31mletter\n",
		SubmissionBreakdown: []postadjudicationstatus.SubmissionBreakdownItem{{
			SubmissionReceiptID: "sub-\x1b[31m1\n",
			AnyMatchFamilies:    []string{"manual-\x1b[31mretry\n", "  "},
		}},
	})

	assert.Equal(t, []string{"retry", "dead-letter"}, got.TransactionGlobalAnyMatchFamilies)
	assert.Equal(t, "dead-letter", got.TransactionGlobalDominantFamily)
	require.Len(t, got.SubmissionBreakdown, 1)
	assert.Equal(t, "sub-1", got.SubmissionBreakdown[0].SubmissionReceiptID)
	assert.Equal(t, []string{"manual-retry"}, got.SubmissionBreakdown[0].AnyMatchFamilies)
}

func TestSanitizeTransactionReceiptCoversEscrowInputAndEnums(t *testing.T) {
	got := sanitizeTransactionReceiptJSON(receipts.TransactionReceipt{
		RequestedScope:                  "scope-\x1b[31m1\n",
		PriceContext:                    "price-\x1b[31mctx\n",
		TrustContext:                    "trust-\x1b[31mctx\n",
		KnowledgeExchangeRuntimeStatus:  receipts.KnowledgeExchangeRuntimeStatus("ready\x1b[31m\n"),
		SettlementProgressionStatus:     receipts.SettlementProgressionStatus("blocked\x1b[31m\n"),
		SettlementProgressionReasonCode: receipts.SettlementProgressionReasonCode("policy\x1b[31m\n"),
		PartialSettlementHint:           "partial-\x1b[31mhint\n",
		DisputeLifecycleStatus:          receipts.DisputeLifecycleStatus("open\x1b[31m\n"),
		CurrentSubmissionReceiptID:      "sub-\x1b[31mcurrent\n",
		CanonicalApprovalStatus:         receipts.ApprovalStatus("approved\x1b[31m\n"),
		CanonicalSettlementStatus:       receipts.SettlementStatus("settled\x1b[31m\n"),
		CurrentPaymentApprovalStatus:    receipts.PaymentApprovalStatus("approved\x1b[31m\n"),
		CanonicalSettlementHint:         "release-\x1b[31mfunds\n",
		EscrowExecutionStatus:           receipts.EscrowExecutionStatus("submitted\x1b[31m\n"),
		EscrowReference:                 "escrow-\x1b[31m1\n",
		EscrowAdjudication:              receipts.EscrowAdjudicationDecision("release\x1b[31m\n"),
		EscrowExecutionInput: &receipts.EscrowExecutionInput{
			BuyerDID:  "did:\x1b[31mbuyer\n",
			SellerDID: "did:\x1b[31mseller\n",
			Amount:    "10.\x1b[31m00\n",
			Milestones: []receipts.EscrowMilestoneInput{{
				Description: "draft\x1b[31m delivered\n",
				Amount:      "5.\x1b[31m00\n",
			}},
		},
	})

	assert.Equal(t, "scope-1", got.RequestedScope)
	assert.Equal(t, receipts.KnowledgeExchangeRuntimeStatus("ready"), got.KnowledgeExchangeRuntimeStatus)
	assert.Equal(t, receipts.SettlementProgressionStatus("blocked"), got.SettlementProgressionStatus)
	assert.Equal(t, receipts.SettlementProgressionReasonCode("policy"), got.SettlementProgressionReasonCode)
	assert.Equal(t, receipts.DisputeLifecycleStatus("open"), got.DisputeLifecycleStatus)
	assert.Equal(t, receipts.ApprovalStatus("approved"), got.CanonicalApprovalStatus)
	assert.Equal(t, receipts.SettlementStatus("settled"), got.CanonicalSettlementStatus)
	assert.Equal(t, receipts.PaymentApprovalStatus("approved"), got.CurrentPaymentApprovalStatus)
	assert.Equal(t, receipts.EscrowExecutionStatus("submitted"), got.EscrowExecutionStatus)
	assert.Equal(t, receipts.EscrowAdjudicationDecision("release"), got.EscrowAdjudication)
	require.NotNil(t, got.EscrowExecutionInput)
	assert.Equal(t, "did:buyer", got.EscrowExecutionInput.BuyerDID)
	assert.Equal(t, "did:seller", got.EscrowExecutionInput.SellerDID)
	assert.Equal(t, "10.00", got.EscrowExecutionInput.Amount)
	require.Len(t, got.EscrowExecutionInput.Milestones, 1)
	assert.Equal(t, "draft delivered", got.EscrowExecutionInput.Milestones[0].Description)
	assert.Equal(t, "5.00", got.EscrowExecutionInput.Milestones[0].Amount)
}

func TestJSONAndCLIErrorHelpersHandleNilAndBlankMessages(t *testing.T) {
	var out bytes.Buffer
	require.NoError(t, printJSONErrorTo(&out, nil))
	assert.Empty(t, out.String())

	blankErr := errors.New("\x1b[31m\n\t")
	safeErr := sanitizeCLIError(blankErr)
	require.Error(t, safeErr)
	assert.Equal(t, "status command failed", safeErr.Error())
	assert.ErrorIs(t, safeErr, blankErr)
	assert.Nil(t, sanitizeCLIError(nil))
}
