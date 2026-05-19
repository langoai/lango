package app

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/p2p/firewall"
	"github.com/langoai/lango/internal/p2p/handshake"
	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/payment/eip3009"
	"github.com/langoai/lango/internal/postadjudicationreplay"
	"github.com/langoai/lango/internal/receipts"
)

func TestWave49P2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess(t *testing.T) {
	t.Parallel()

	tools := buildP2PTools(&p2pComponents{
		sessions: wave28P2PSessions(t),
		fw:       firewall.New(nil, nil),
	})
	tool := findP2PTool(t, tools, "p2p_connect")

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"multiaddr": "/ip4/127.0.0.1/tcp/9000",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "parse peer addr")
}

func TestWave49P2PPaidInvokePaidToolInvokeErrorSkipsSpendRecord(t *testing.T) {
	originalRemoteAgent := newP2PPaidInvokeRemoteAgent
	originalNewUnsigned := newEIP3009Unsigned
	t.Cleanup(func() {
		newP2PPaidInvokeRemoteAgent = originalRemoteAgent
		newEIP3009Unsigned = originalNewUnsigned
	})

	did := wave28P2PPeerDID(t)
	sessions := wave28P2PSessions(t, did.ID)
	remote := &wave28P2PRemoteAgent{
		quote:     wave28P2PPaidQuote("0.42"),
		invokeErr: errors.New("seller unavailable"),
	}
	newP2PPaidInvokeRemoteAgent = func(string, *identity.DID, *handshake.Session, *p2pComponents) p2pPaidInvokeRemoteAgent {
		return remote
	}
	newEIP3009Unsigned = wave49UnsignedAuth

	limiter := &wave28P2PLimiter{autoOK: true}
	wallet := wave28P2PWallet(t)
	tools := buildP2PPaidInvokeTool(&p2pComponents{sessions: sessions}, &paymentComponents{
		wallet:  wallet,
		limiter: limiter,
		chainID: 84532,
	})
	require.Len(t, tools, 1)

	got, err := tools[0].Handler(context.Background(), map[string]interface{}{
		"peer_did":  did.ID,
		"tool_name": "paid_tool",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "paid tool invoke: seller unavailable")
	assert.True(t, remote.paidInvoked)
	assert.Empty(t, limiter.recorded)
	assert.Equal(t, 1, wallet.signTransactionCalls)
}

func TestWave49P2PPaidInvokeRemoteErrorDefaultsMessageWhenEmpty(t *testing.T) {
	originalRemoteAgent := newP2PPaidInvokeRemoteAgent
	originalNewUnsigned := newEIP3009Unsigned
	t.Cleanup(func() {
		newP2PPaidInvokeRemoteAgent = originalRemoteAgent
		newEIP3009Unsigned = originalNewUnsigned
	})

	did := wave28P2PPeerDID(t)
	sessions := wave28P2PSessions(t, did.ID)
	remote := &wave28P2PRemoteAgent{
		quote:        wave28P2PPaidQuote("0.25"),
		paidResponse: &protocol.Response{Status: protocol.ResponseStatusError},
	}
	newP2PPaidInvokeRemoteAgent = func(string, *identity.DID, *handshake.Session, *p2pComponents) p2pPaidInvokeRemoteAgent {
		return remote
	}
	newEIP3009Unsigned = wave49UnsignedAuth

	limiter := &wave28P2PLimiter{autoOK: true}
	tools := buildP2PPaidInvokeTool(&p2pComponents{sessions: sessions}, &paymentComponents{
		wallet:  wave28P2PWallet(t),
		limiter: limiter,
		chainID: 84532,
	})
	require.Len(t, tools, 1)

	got, err := tools[0].Handler(context.Background(), map[string]interface{}{
		"peer_did":  did.ID,
		"tool_name": "paid_tool",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "remote paid_tool: remote tool error")
	assert.True(t, remote.paidInvoked)
	assert.Empty(t, limiter.recorded)
}

func TestWave49PostAdjudicationTaskReaderMapsDispatcherSnapshots(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.May, 19, 9, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, time.May, 19, 9, 30, 0, 0, time.UTC)
	nextRetryAt := time.Date(2026, time.May, 19, 10, 0, 0, 0, time.UTC)
	dispatcher := &fakeAdjudicationBackgroundDispatcher{
		tasks: []background.TaskSnapshot{{
			ID:           "task-wave49",
			StatusText:   "failed",
			RetryKey:     "tx-wave49:release",
			AttemptCount: 3,
			NextRetryAt:  nextRetryAt,
			StartedAt:    startedAt,
			CompletedAt:  completedAt,
		}},
	}

	got, err := (postAdjudicationStatusBackgroundTaskReader{
		dispatcher: dispatcher,
	}).ListTaskSnapshots(context.Background())

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "task-wave49", got[0].TaskID)
	assert.Equal(t, "failed", got[0].Status)
	assert.Equal(t, "tx-wave49:release", got[0].RetryKey)
	assert.Equal(t, 3, got[0].AttemptCount)
	assert.Equal(t, nextRetryAt, got[0].NextRetryAt)
	assert.Equal(t, startedAt, got[0].StartedAt)
	assert.Equal(t, completedAt, got[0].CompletedAt)
}

func TestWave49RetryPostAdjudicationReceiptOmitsDispatchWhenUnavailable(t *testing.T) {
	t.Parallel()

	got := newRetryPostAdjudicationExecutionReceipt(postadjudicationreplay.Result{
		CanonicalAdjudication: postadjudicationreplay.CanonicalAdjudicationSnapshot{
			TransactionReceipt: receipts.TransactionReceipt{
				TransactionReceiptID:       "tx-wave49-retry",
				CurrentSubmissionReceiptID: "sub-wave49-current",
				EscrowReference:            "escrow-wave49",
				EscrowAdjudication:         receipts.EscrowAdjudicationRefund,
			},
			SubmissionReceipt: receipts.SubmissionReceipt{
				SubmissionReceiptID: "sub-wave49-canonical",
			},
		},
	})

	assert.Equal(t, retryPostAdjudicationExecutionReceipt{
		TransactionReceiptID: "tx-wave49-retry",
		SubmissionReceiptID:  "sub-wave49-canonical",
		EscrowReference:      "escrow-wave49",
		Outcome:              "refund",
	}, got)
	assert.Nil(t, got.Dispatch)
}

func wave49UnsignedAuth(from, to common.Address, value *big.Int, _ time.Time) (*eip3009.UnsignedAuth, error) {
	return &eip3009.UnsignedAuth{
		From:        from,
		To:          to,
		Value:       new(big.Int).Set(value),
		ValidAfter:  big.NewInt(10),
		ValidBefore: big.NewInt(20),
		Nonce:       [32]byte{31: 49},
	}, nil
}
