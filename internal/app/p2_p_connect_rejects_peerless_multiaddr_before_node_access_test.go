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

func TestP2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess(t *testing.T) {
	t.Parallel()

	tools := buildP2PTools(&p2pComponents{
		sessions: p2PToolsMetadataAndMissingDependencyBranchesP2PSessions(t),
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

func TestP2PPaidInvokePaidToolInvokeErrorSkipsSpendRecord(t *testing.T) {
	originalRemoteAgent := newP2PPaidInvokeRemoteAgent
	originalNewUnsigned := newEIP3009Unsigned
	t.Cleanup(func() {
		newP2PPaidInvokeRemoteAgent = originalRemoteAgent
		newEIP3009Unsigned = originalNewUnsigned
	})

	did := p2PToolsMetadataAndMissingDependencyBranchesP2PPeerDID(t)
	sessions := p2PToolsMetadataAndMissingDependencyBranchesP2PSessions(t, did.ID)
	remote := &p2PToolsMetadataAndMissingDependencyBranchesP2PRemoteAgent{
		quote:     p2PToolsMetadataAndMissingDependencyBranchesP2PPaidQuote("0.42"),
		invokeErr: errors.New("seller unavailable"),
	}
	newP2PPaidInvokeRemoteAgent = func(string, *identity.DID, *handshake.Session, *p2pComponents) p2pPaidInvokeRemoteAgent {
		return remote
	}
	newEIP3009Unsigned = p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccessUnsignedAuth

	limiter := &p2PToolsMetadataAndMissingDependencyBranchesP2PLimiter{autoOK: true}
	wallet := p2PToolsMetadataAndMissingDependencyBranchesP2PWallet(t)
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

func TestP2PPaidInvokeRemoteErrorDefaultsMessageWhenEmpty(t *testing.T) {
	originalRemoteAgent := newP2PPaidInvokeRemoteAgent
	originalNewUnsigned := newEIP3009Unsigned
	t.Cleanup(func() {
		newP2PPaidInvokeRemoteAgent = originalRemoteAgent
		newEIP3009Unsigned = originalNewUnsigned
	})

	did := p2PToolsMetadataAndMissingDependencyBranchesP2PPeerDID(t)
	sessions := p2PToolsMetadataAndMissingDependencyBranchesP2PSessions(t, did.ID)
	remote := &p2PToolsMetadataAndMissingDependencyBranchesP2PRemoteAgent{
		quote:        p2PToolsMetadataAndMissingDependencyBranchesP2PPaidQuote("0.25"),
		paidResponse: &protocol.Response{Status: protocol.ResponseStatusError},
	}
	newP2PPaidInvokeRemoteAgent = func(string, *identity.DID, *handshake.Session, *p2pComponents) p2pPaidInvokeRemoteAgent {
		return remote
	}
	newEIP3009Unsigned = p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccessUnsignedAuth

	limiter := &p2PToolsMetadataAndMissingDependencyBranchesP2PLimiter{autoOK: true}
	tools := buildP2PPaidInvokeTool(&p2pComponents{sessions: sessions}, &paymentComponents{
		wallet:  p2PToolsMetadataAndMissingDependencyBranchesP2PWallet(t),
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

func TestPostAdjudicationTaskReaderMapsDispatcherSnapshots(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.May, 19, 9, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, time.May, 19, 9, 30, 0, 0, time.UTC)
	nextRetryAt := time.Date(2026, time.May, 19, 10, 0, 0, 0, time.UTC)
	dispatcher := &fakeAdjudicationBackgroundDispatcher{
		tasks: []background.TaskSnapshot{{
			ID:           "task-p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess9",
			StatusText:   "failed",
			RetryKey:     "tx-p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess9:release",
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
	assert.Equal(t, "task-p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess9", got[0].TaskID)
	assert.Equal(t, "failed", got[0].Status)
	assert.Equal(t, "tx-p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess9:release", got[0].RetryKey)
	assert.Equal(t, 3, got[0].AttemptCount)
	assert.Equal(t, nextRetryAt, got[0].NextRetryAt)
	assert.Equal(t, startedAt, got[0].StartedAt)
	assert.Equal(t, completedAt, got[0].CompletedAt)
}

func TestRetryPostAdjudicationReceiptOmitsDispatchWhenUnavailable(t *testing.T) {
	t.Parallel()

	got := newRetryPostAdjudicationExecutionReceipt(postadjudicationreplay.Result{
		CanonicalAdjudication: postadjudicationreplay.CanonicalAdjudicationSnapshot{
			TransactionReceipt: receipts.TransactionReceipt{
				TransactionReceiptID:       "tx-p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess9-retry",
				CurrentSubmissionReceiptID: "sub-p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess9-current",
				EscrowReference:            "escrow-p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess9",
				EscrowAdjudication:         receipts.EscrowAdjudicationRefund,
			},
			SubmissionReceipt: receipts.SubmissionReceipt{
				SubmissionReceiptID: "sub-p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess9-canonical",
			},
		},
	})

	assert.Equal(t, retryPostAdjudicationExecutionReceipt{
		TransactionReceiptID: "tx-p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess9-retry",
		SubmissionReceiptID:  "sub-p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess9-canonical",
		EscrowReference:      "escrow-p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccess9",
		Outcome:              "refund",
	}, got)
	assert.Nil(t, got.Dispatch)
}

func p2PConnectRejectsPeerlessMultiaddrBeforeNodeAccessUnsignedAuth(from, to common.Address, value *big.Int, _ time.Time) (*eip3009.UnsignedAuth, error) {
	return &eip3009.UnsignedAuth{
		From:        from,
		To:          to,
		Value:       new(big.Int).Set(value),
		ValidAfter:  big.NewInt(10),
		ValidBefore: big.NewInt(20),
		Nonce:       [32]byte{31: 49},
	}, nil
}
