package cockpit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/postadjudicationstatus"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/toolcatalog"
)

func TestDeadLetterToolBridge_ReadyRequiresBothTools(t *testing.T) {
	catalog := toolcatalog.New()
	bridge := NewDeadLetterToolBridge(catalog)
	assert.False(t, bridge.Ready())

	catalog.RegisterCategory(toolcatalog.Category{Name: "knowledge", Enabled: true})
	catalog.Register("knowledge", []*agent.Tool{
		{Name: "list_dead_lettered_post_adjudication_executions", Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"entries": []postadjudicationstatus.DeadLetterBacklogEntry{}}, nil
		}},
	})
	assert.False(t, bridge.Ready())

	catalog.Register("knowledge", []*agent.Tool{
		{Name: "get_post_adjudication_execution_status", Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
			return postadjudicationstatus.TransactionStatus{}, nil
		}},
	})
	assert.True(t, bridge.Ready())
	assert.False(t, bridge.CanRetry())

	catalog.Register("knowledge", []*agent.Tool{
		{Name: "retry_post_adjudication_execution", Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"status": "queued"}, nil
		}},
	})
	assert.True(t, bridge.CanRetry())
}

func TestDeadLetterToolBridge_ListAndDetail(t *testing.T) {
	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{Name: "knowledge", Enabled: true})

	wantEntries := []postadjudicationstatus.DeadLetterBacklogEntry{
		{
			TransactionReceiptID: "tx-1",
			SubmissionReceiptID:  "sub-1",
			Adjudication:         "release",
			LatestRetryAttempt:   3,
		},
	}
	wantDetail := postadjudicationstatus.TransactionStatus{
		CanonicalSnapshot: postadjudicationstatus.CanonicalSnapshot{
			TransactionReceipt: receipts.TransactionReceipt{TransactionReceiptID: "tx-1"},
			SubmissionReceipt:  receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1"},
		},
		Adjudication: "release",
	}

	catalog.Register("knowledge", []*agent.Tool{
		{
			Name: "list_dead_lettered_post_adjudication_executions",
			Handler: func(_ context.Context, params map[string]interface{}) (interface{}, error) {
				assert.Equal(t, "tx-1", params["query"])
				assert.Equal(t, "release", params["adjudication"])
				assert.Equal(t, "manual-retry-requested", params["latest_status_subtype"])
				assert.Equal(t, "manual-retry", params["latest_status_subtype_family"])
				assert.Equal(t, "dead-letter", params["any_match_family"])
				assert.Equal(t, "operator:alice", params["manual_replay_actor"])
				assert.Equal(t, "2026-04-24T11:00:00Z", params["dead_lettered_after"])
				assert.Equal(t, "2026-04-24T13:00:00Z", params["dead_lettered_before"])
				assert.Equal(t, "worker exhausted", params["dead_letter_reason_query"])
				assert.Equal(t, "dispatch-7", params["latest_dispatch_reference"])
				return map[string]interface{}{
					"entries": wantEntries,
					"count":   1,
					"total":   1,
				}, nil
			},
		},
		{
			Name: "get_post_adjudication_execution_status",
			Handler: func(_ context.Context, params map[string]interface{}) (interface{}, error) {
				require.Equal(t, "tx-1", params["transaction_receipt_id"])
				return wantDetail, nil
			},
		},
	})

	bridge := NewDeadLetterToolBridge(catalog)
	gotEntries, err := bridge.List(context.Background(), DeadLetterListOptions{
		Query:                     "tx-1",
		Adjudication:              "release",
		LatestStatusSubtype:       "manual-retry-requested",
		LatestStatusSubtypeFamily: "manual-retry",
		AnyMatchFamily:            "dead-letter",
		ManualReplayActor:         "operator:alice",
		DeadLetteredAfter:         "2026-04-24T11:00:00Z",
		DeadLetteredBefore:        "2026-04-24T13:00:00Z",
		DeadLetterReasonQuery:     "worker exhausted",
		LatestDispatchReference:   "dispatch-7",
	})
	require.NoError(t, err)
	assert.Equal(t, wantEntries, gotEntries)

	gotDetail, err := bridge.Detail(context.Background(), "tx-1")
	require.NoError(t, err)
	assert.Equal(t, wantDetail, gotDetail)
}

func TestDeadLetterToolBridge_Retry(t *testing.T) {
	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{Name: "knowledge", Enabled: true})

	called := false
	catalog.Register("knowledge", []*agent.Tool{
		{
			Name: "list_dead_lettered_post_adjudication_executions",
			Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
				return map[string]interface{}{"entries": []postadjudicationstatus.DeadLetterBacklogEntry{}}, nil
			},
		},
		{
			Name: "get_post_adjudication_execution_status",
			Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
				return postadjudicationstatus.TransactionStatus{}, nil
			},
		},
		{
			Name: "retry_post_adjudication_execution",
			Handler: func(_ context.Context, params map[string]interface{}) (interface{}, error) {
				called = true
				require.Equal(t, "tx-7", params["transaction_receipt_id"])
				return map[string]interface{}{"status": "queued"}, nil
			},
		},
	})

	bridge := NewDeadLetterToolBridge(catalog)
	require.True(t, bridge.CanRetry())
	require.NoError(t, bridge.Retry(context.Background(), "tx-7"))
	assert.True(t, called)
}

func TestDeadLetterToolBridge_ListOmitsAdjudicationWhenAll(t *testing.T) {
	catalog := toolcatalog.New()
	catalog.RegisterCategory(toolcatalog.Category{Name: "knowledge", Enabled: true})
	catalog.Register("knowledge", []*agent.Tool{
		{
			Name: "list_dead_lettered_post_adjudication_executions",
			Handler: func(_ context.Context, params map[string]interface{}) (interface{}, error) {
				assert.Equal(t, "needle", params["query"])
				_, hasAdjudication := params["adjudication"]
				assert.False(t, hasAdjudication)
				_, hasSubtype := params["latest_status_subtype"]
				assert.False(t, hasSubtype)
				_, hasFamily := params["latest_status_subtype_family"]
				assert.False(t, hasFamily)
				_, hasAnyMatchFamily := params["any_match_family"]
				assert.False(t, hasAnyMatchFamily)
				_, hasActor := params["manual_replay_actor"]
				assert.False(t, hasActor)
				_, hasAfter := params["dead_lettered_after"]
				assert.False(t, hasAfter)
				_, hasBefore := params["dead_lettered_before"]
				assert.False(t, hasBefore)
				_, hasReason := params["dead_letter_reason_query"]
				assert.False(t, hasReason)
				_, hasDispatch := params["latest_dispatch_reference"]
				assert.False(t, hasDispatch)
				return map[string]interface{}{
					"entries": []postadjudicationstatus.DeadLetterBacklogEntry{},
				}, nil
			},
		},
		{
			Name: "get_post_adjudication_execution_status",
			Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
				return postadjudicationstatus.TransactionStatus{}, nil
			},
		},
	})

	bridge := NewDeadLetterToolBridge(catalog)
	_, err := bridge.List(context.Background(), DeadLetterListOptions{
		Query:                     "needle",
		Adjudication:              "all",
		LatestStatusSubtype:       "all",
		LatestStatusSubtypeFamily: "all",
		AnyMatchFamily:            "all",
		ManualReplayActor:         "",
		DeadLetteredAfter:         "",
		DeadLetteredBefore:        "",
		DeadLetterReasonQuery:     "",
		LatestDispatchReference:   "",
	})
	require.NoError(t, err)
}
