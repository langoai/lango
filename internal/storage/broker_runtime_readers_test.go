package storage

import (
	"context"
	"testing"
	"time"

	"github.com/langoai/lango/internal/p2p/reputation"
	"github.com/langoai/lango/internal/search"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/storagebroker"
	"github.com/stretchr/testify/require"
)

type runtimeReaderStubBroker struct {
}

func (b *runtimeReaderStubBroker) Health(context.Context) (storagebroker.HealthResult, error) {
	return storagebroker.HealthResult{Opened: true}, nil
}
func (b *runtimeReaderStubBroker) OpenDB(context.Context, storagebroker.OpenDBRequest) (storagebroker.OpenDBResult, error) {
	return storagebroker.OpenDBResult{Opened: true}, nil
}
func (b *runtimeReaderStubBroker) DBStatusSummary(context.Context, storagebroker.DBStatusSummaryRequest) (storagebroker.DBStatusSummaryResult, error) {
	return storagebroker.DBStatusSummaryResult{}, nil
}
func (b *runtimeReaderStubBroker) EncryptPayload(context.Context, []byte) (storagebroker.EncryptPayloadResult, error) {
	return storagebroker.EncryptPayloadResult{}, nil
}
func (b *runtimeReaderStubBroker) DecryptPayload(context.Context, []byte, []byte, int) (storagebroker.DecryptPayloadResult, error) {
	return storagebroker.DecryptPayloadResult{}, nil
}
func (b *runtimeReaderStubBroker) LoadSecurityState(context.Context) (storagebroker.LoadSecurityStateResult, error) {
	return storagebroker.LoadSecurityStateResult{}, nil
}
func (b *runtimeReaderStubBroker) StoreSalt(context.Context, []byte) error { return nil }
func (b *runtimeReaderStubBroker) StoreChecksum(context.Context, []byte) error {
	return nil
}
func (b *runtimeReaderStubBroker) ConfigLoad(context.Context, string) (storagebroker.ConfigLoadResult, error) {
	return storagebroker.ConfigLoadResult{}, nil
}
func (b *runtimeReaderStubBroker) ConfigLoadActive(context.Context) (storagebroker.ConfigLoadActiveResult, error) {
	return storagebroker.ConfigLoadActiveResult{}, nil
}
func (b *runtimeReaderStubBroker) ConfigSave(context.Context, string, any, map[string]bool) error {
	return nil
}
func (b *runtimeReaderStubBroker) ConfigSetActive(context.Context, string) error { return nil }
func (b *runtimeReaderStubBroker) ConfigList(context.Context) (storagebroker.ConfigListResult, error) {
	return storagebroker.ConfigListResult{}, nil
}
func (b *runtimeReaderStubBroker) ConfigDelete(context.Context, string) error { return nil }
func (b *runtimeReaderStubBroker) ConfigExists(context.Context, string) (storagebroker.ConfigExistsResult, error) {
	return storagebroker.ConfigExistsResult{}, nil
}
func (b *runtimeReaderStubBroker) SessionCreate(context.Context, *session.Session) error { return nil }
func (b *runtimeReaderStubBroker) SessionGet(context.Context, string) (*session.Session, error) {
	return nil, nil
}
func (b *runtimeReaderStubBroker) SessionUpdate(context.Context, *session.Session) error { return nil }
func (b *runtimeReaderStubBroker) SessionDelete(context.Context, string) error           { return nil }
func (b *runtimeReaderStubBroker) SessionAppendMessage(context.Context, string, session.Message) error {
	return nil
}
func (b *runtimeReaderStubBroker) SessionEnd(context.Context, string) error { return nil }
func (b *runtimeReaderStubBroker) SessionList(context.Context) ([]session.SessionSummary, error) {
	return nil, nil
}
func (b *runtimeReaderStubBroker) SessionGetSalt(context.Context, string) ([]byte, error) {
	return nil, nil
}
func (b *runtimeReaderStubBroker) SessionSetSalt(context.Context, string, []byte) error { return nil }
func (b *runtimeReaderStubBroker) RecallIndexSession(context.Context, string) error     { return nil }
func (b *runtimeReaderStubBroker) RecallProcessPending(context.Context) error           { return nil }
func (b *runtimeReaderStubBroker) RecallSearch(context.Context, string, int) ([]search.SearchResult, error) {
	return nil, nil
}
func (b *runtimeReaderStubBroker) RecallGetSummary(context.Context, string) (string, error) {
	return "", nil
}
func (b *runtimeReaderStubBroker) Close(context.Context) error { return nil }

func (b *runtimeReaderStubBroker) LearningHistory(context.Context, int) (storagebroker.LearningHistoryResult, error) {
	return storagebroker.LearningHistoryResult{
		Entries: []storagebroker.LearningHistoryRecord{{
			ID:         "l1",
			Trigger:    "timeout",
			Category:   "timeout",
			Diagnosis:  "diag",
			Fix:        "fix",
			Confidence: 0.7,
			CreatedAt:  time.Unix(1, 0),
		}},
	}, nil
}

func (b *runtimeReaderStubBroker) PendingInquiries(context.Context, int) (storagebroker.PendingInquiriesResult, error) {
	return storagebroker.PendingInquiriesResult{
		Entries: []storagebroker.PendingInquiryRecord{{
			ID:       "i1",
			Topic:    "topic",
			Question: "question",
			Priority: "high",
			Created:  time.Unix(2, 0),
		}},
	}, nil
}

func (b *runtimeReaderStubBroker) WorkflowRuns(context.Context, int) (storagebroker.WorkflowRunsResult, error) {
	return storagebroker.WorkflowRunsResult{
		Runs: []storagebroker.WorkflowRunRecord{{
			RunID:          "run-1",
			WorkflowName:   "wf",
			Status:         "completed",
			TotalSteps:     3,
			CompletedSteps: 3,
			StartedAt:      time.Unix(3, 0),
		}},
	}, nil
}

func (b *runtimeReaderStubBroker) Alerts(context.Context, time.Time) (storagebroker.AlertsResult, error) {
	return storagebroker.AlertsResult{
		Alerts: []storagebroker.AlertRecord{{
			ID:        "a1",
			Type:      "policy",
			Actor:     "system",
			Details:   map[string]interface{}{"k": "v"},
			Timestamp: time.Unix(4, 0),
		}},
	}, nil
}

func (b *runtimeReaderStubBroker) ReputationGet(context.Context, string) (storagebroker.ReputationGetResult, error) {
	return storagebroker.ReputationGetResult{
		PeerDID:             "did:lango:test",
		TrustScore:          0.9,
		SuccessfulExchanges: 2,
		FailedExchanges:     1,
		TimeoutCount:        0,
		FirstSeen:           time.Unix(5, 0),
		LastInteraction:     time.Unix(6, 0),
		Found:               true,
	}, nil
}
func (b *runtimeReaderStubBroker) PaymentHistory(context.Context, int) (storagebroker.PaymentHistoryResult, error) {
	return storagebroker.PaymentHistoryResult{}, nil
}
func (b *runtimeReaderStubBroker) PaymentUsage(context.Context) (storagebroker.PaymentUsageResult, error) {
	return storagebroker.PaymentUsageResult{DailySpent: "0"}, nil
}

func TestWithBrokerRuntimeReaders_WiresReaders(t *testing.T) {
	f := NewFacade(nil, nil, WithBrokerRuntimeReaders(&runtimeReaderStubBroker{}))

	learning, err := f.LearningHistory(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, learning, 1)
	require.Equal(t, "l1", learning[0].ID)

	inquiries, err := f.PendingInquiries(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, inquiries, 1)
	require.Equal(t, "i1", inquiries[0].ID)

	alerts, err := f.Alerts(context.Background(), time.Unix(0, 0))
	require.NoError(t, err)
	require.Len(t, alerts, 1)
	require.Equal(t, "a1", alerts[0].ID)

	details, err := f.ReputationDetails(context.Background(), "did:lango:test")
	require.NoError(t, err)
	require.Equal(t, &reputation.PeerDetails{
		PeerDID:             "did:lango:test",
		TrustScore:          0.9,
		SuccessfulExchanges: 2,
		FailedExchanges:     1,
		TimeoutCount:        0,
		FirstSeen:           time.Unix(5, 0),
		LastInteraction:     time.Unix(6, 0),
	}, details)

	reader := f.WorkflowStateStore(nil)
	require.NotNil(t, reader)
	runs, err := reader.ListRuns(context.Background(), 10)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	require.Equal(t, "run-1", runs[0].RunID)
	status, err := reader.GetRunStatus(context.Background(), "run-1")
	require.NoError(t, err)
	require.Equal(t, "run-1", status.RunID)
}
