package settlement

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/payment/eip3009"
)

func TestNew_Defaults(t *testing.T) {
	svc := New(Config{
		Logger: zap.NewNop().Sugar(),
	})
	require.NotNil(t, svc)
	assert.Equal(t, 2*time.Minute, svc.timeout)
	assert.Equal(t, 3, svc.maxRetries)
}

func TestNew_CustomConfig(t *testing.T) {
	svc := New(Config{
		ReceiptTimeout: 5 * time.Minute,
		MaxRetries:     5,
		Logger:         zap.NewNop().Sugar(),
	})
	assert.Equal(t, 5*time.Minute, svc.timeout)
	assert.Equal(t, 5, svc.maxRetries)
}

func TestSubscribe_RegistersHandler(t *testing.T) {
	svc := New(Config{
		Logger: zap.NewNop().Sugar(),
	})
	bus := eventbus.New()

	// Should not panic.
	svc.Subscribe(bus)
}

func TestHandleEvent_NilAuth(t *testing.T) {
	svc := New(Config{
		Logger: zap.NewNop().Sugar(),
	})

	// handleEvent with nil auth should not panic.
	svc.handleEvent(eventbus.ToolExecutionPaidEvent{
		PeerDID:  "did:peer:test",
		ToolName: "test-tool",
		Auth:     nil,
	})
}

func TestHandleEvent_WrongAuthType(t *testing.T) {
	svc := New(Config{
		Logger: zap.NewNop().Sugar(),
	})

	// handleEvent with wrong auth type should not panic.
	svc.handleEvent(eventbus.ToolExecutionPaidEvent{
		PeerDID:  "did:peer:test",
		ToolName: "test-tool",
		Auth:     "not-an-authorization",
	})
}

type mockRepRecorder struct {
	successes int
	failures  int
}

func (m *mockRepRecorder) RecordSuccess(_ context.Context, _ string) error {
	m.successes++
	return nil
}

func (m *mockRepRecorder) RecordFailure(_ context.Context, _ string) error {
	m.failures++
	return nil
}

func TestHandleEvent_FailureRecordsReputation(t *testing.T) {
	rec := &mockRepRecorder{}
	svc := New(Config{
		Logger: zap.NewNop().Sugar(),
	})
	svc.SetReputationRecorder(rec)

	// This will fail because there's no RPC client, which triggers failure path.
	auth := &eip3009.Authorization{
		From:        common.HexToAddress("0xaaaa"),
		To:          common.HexToAddress("0xbbbb"),
		Value:       big.NewInt(500000),
		ValidAfter:  big.NewInt(0),
		ValidBefore: big.NewInt(time.Now().Add(10 * time.Minute).Unix()),
	}

	svc.handleEvent(eventbus.ToolExecutionPaidEvent{
		PeerDID:  "did:peer:test",
		ToolName: "test-tool",
		Auth:     auth,
	})

	assert.Equal(t, 1, rec.failures)
	assert.Equal(t, 0, rec.successes)
}
