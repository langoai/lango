package cliboot

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/search"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/storagebroker"
)

func TestBootResultEnablesStorageBroker(t *testing.T) {
	oldVersion := Version
	oldBootstrapRun := bootstrapRun
	t.Cleanup(func() {
		Version = oldVersion
		bootstrapRun = oldBootstrapRun
	})

	Version = "test-version"
	var captured bootstrap.Options
	bootstrapRun = func(opts bootstrap.Options) (*bootstrap.Result, error) {
		captured = opts
		return &bootstrap.Result{}, nil
	}

	result, err := BootResult()

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "test-version", captured.Version)
	require.True(t, captured.StartStorageBroker)
}

func TestConfigEnablesStorageBrokerAndClosesBootstrapResult(t *testing.T) {
	oldVersion := Version
	oldBootstrapRun := bootstrapRun
	t.Cleanup(func() {
		Version = oldVersion
		bootstrapRun = oldBootstrapRun
	})

	Version = "test-version"
	broker := &closeTrackingBroker{}
	var captured bootstrap.Options
	bootstrapRun = func(opts bootstrap.Options) (*bootstrap.Result, error) {
		captured = opts
		return &bootstrap.Result{
			Config: config.DefaultConfig(),
			Broker: broker,
		}, nil
	}

	cfg, err := Config()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "test-version", captured.Version)
	require.True(t, captured.StartStorageBroker)
	require.True(t, broker.closed)
}

type closeTrackingBroker struct {
	closed bool
}

func (b *closeTrackingBroker) Health(context.Context) (storagebroker.HealthResult, error) {
	return storagebroker.HealthResult{}, nil
}

func (b *closeTrackingBroker) OpenDB(context.Context, storagebroker.OpenDBRequest) (storagebroker.OpenDBResult, error) {
	return storagebroker.OpenDBResult{}, nil
}

func (b *closeTrackingBroker) DBStatusSummary(context.Context, storagebroker.DBStatusSummaryRequest) (storagebroker.DBStatusSummaryResult, error) {
	return storagebroker.DBStatusSummaryResult{}, nil
}

func (b *closeTrackingBroker) EncryptPayload(context.Context, []byte) (storagebroker.EncryptPayloadResult, error) {
	return storagebroker.EncryptPayloadResult{}, nil
}

func (b *closeTrackingBroker) DecryptPayload(context.Context, []byte, []byte, int) (storagebroker.DecryptPayloadResult, error) {
	return storagebroker.DecryptPayloadResult{}, nil
}

func (b *closeTrackingBroker) LoadSecurityState(context.Context) (storagebroker.LoadSecurityStateResult, error) {
	return storagebroker.LoadSecurityStateResult{}, nil
}

func (b *closeTrackingBroker) StoreSalt(context.Context, []byte) error {
	return nil
}

func (b *closeTrackingBroker) StoreChecksum(context.Context, []byte) error {
	return nil
}

func (b *closeTrackingBroker) ConfigLoad(context.Context, string) (storagebroker.ConfigLoadResult, error) {
	return storagebroker.ConfigLoadResult{}, nil
}

func (b *closeTrackingBroker) ConfigLoadActive(context.Context) (storagebroker.ConfigLoadActiveResult, error) {
	raw, _ := json.Marshal(config.DefaultConfig())
	return storagebroker.ConfigLoadActiveResult{Name: "default", Config: raw}, nil
}

func (b *closeTrackingBroker) ConfigSave(context.Context, string, any, map[string]bool) error {
	return nil
}

func (b *closeTrackingBroker) ConfigSetActive(context.Context, string) error {
	return nil
}

func (b *closeTrackingBroker) ConfigList(context.Context) (storagebroker.ConfigListResult, error) {
	return storagebroker.ConfigListResult{}, nil
}

func (b *closeTrackingBroker) ConfigDelete(context.Context, string) error {
	return nil
}

func (b *closeTrackingBroker) ConfigExists(context.Context, string) (storagebroker.ConfigExistsResult, error) {
	return storagebroker.ConfigExistsResult{}, nil
}

func (b *closeTrackingBroker) SessionCreate(context.Context, *session.Session) error {
	return nil
}

func (b *closeTrackingBroker) SessionGet(context.Context, string) (*session.Session, error) {
	return nil, nil
}

func (b *closeTrackingBroker) SessionUpdate(context.Context, *session.Session) error {
	return nil
}

func (b *closeTrackingBroker) SessionDelete(context.Context, string) error {
	return nil
}

func (b *closeTrackingBroker) SessionAppendMessage(context.Context, string, session.Message) error {
	return nil
}

func (b *closeTrackingBroker) SessionEnd(context.Context, string) error {
	return nil
}

func (b *closeTrackingBroker) SessionList(context.Context) ([]session.SessionSummary, error) {
	return nil, nil
}

func (b *closeTrackingBroker) SessionGetSalt(context.Context, string) ([]byte, error) {
	return nil, nil
}

func (b *closeTrackingBroker) SessionSetSalt(context.Context, string, []byte) error {
	return nil
}

func (b *closeTrackingBroker) RecallIndexSession(context.Context, string) error {
	return nil
}

func (b *closeTrackingBroker) RecallProcessPending(context.Context) error {
	return nil
}

func (b *closeTrackingBroker) RecallSearch(context.Context, string, int) ([]search.SearchResult, error) {
	return nil, nil
}

func (b *closeTrackingBroker) RecallGetSummary(context.Context, string) (string, error) {
	return "", nil
}

func (b *closeTrackingBroker) LearningHistory(context.Context, int) (storagebroker.LearningHistoryResult, error) {
	return storagebroker.LearningHistoryResult{}, nil
}

func (b *closeTrackingBroker) PendingInquiries(context.Context, int) (storagebroker.PendingInquiriesResult, error) {
	return storagebroker.PendingInquiriesResult{}, nil
}

func (b *closeTrackingBroker) WorkflowRuns(context.Context, int) (storagebroker.WorkflowRunsResult, error) {
	return storagebroker.WorkflowRunsResult{}, nil
}

func (b *closeTrackingBroker) Alerts(context.Context, time.Time) (storagebroker.AlertsResult, error) {
	return storagebroker.AlertsResult{}, nil
}

func (b *closeTrackingBroker) ReputationGet(context.Context, string) (storagebroker.ReputationGetResult, error) {
	return storagebroker.ReputationGetResult{}, nil
}

func (b *closeTrackingBroker) PaymentHistory(context.Context, int) (storagebroker.PaymentHistoryResult, error) {
	return storagebroker.PaymentHistoryResult{}, nil
}

func (b *closeTrackingBroker) PaymentUsage(context.Context) (storagebroker.PaymentUsageResult, error) {
	return storagebroker.PaymentUsageResult{}, nil
}

func (b *closeTrackingBroker) Close(context.Context) error {
	b.closed = true
	return nil
}
