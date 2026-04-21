package app

import (
	"context"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/langoai/lango/internal/p2p/handshake"
	"github.com/langoai/lango/internal/p2p/identity"
	corepayment "github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/receipts"
	toolpayment "github.com/langoai/lango/internal/tools/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestP2PPayment_DeniesWhenSettlementHintIsNotPrepay(t *testing.T) {
	t.Parallel()

	receiptStore := receipts.NewStore()
	sub, tx, err := receiptStore.CreateSubmissionReceipt(context.Background(), receipts.CreateSubmissionInput{
		TransactionID:       "tx-p2p-deny",
		ArtifactLabel:       "artifact",
		PayloadHash:         "hash",
		SourceLineageDigest: "lineage",
	})
	require.NoError(t, err)
	_, err = receiptStore.ApplyUpfrontPaymentApproval(context.Background(), tx.TransactionReceiptID, sub.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "approved",
		SuggestedMode: paymentapproval.ModeEscrow,
	})
	require.NoError(t, err)

	pk, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	did, err := identity.DIDFromPublicKey(ethcrypto.CompressPubkey(&pk.PublicKey))
	require.NoError(t, err)

	sessions, err := handshake.NewSessionStore(time.Hour)
	require.NoError(t, err)
	_, err = sessions.Create(did.ID, false)
	require.NoError(t, err)

	auditor := &fakeP2PAuditor{}
	pc := &paymentComponents{service: &corepayment.Service{}}
	p2pc := &p2pComponents{sessions: sessions}
	tools := buildP2PPaymentTool(p2pc, pc, receiptStore, auditor)
	require.Len(t, tools, 1)

	result, err := tools[0].Handler(context.Background(), map[string]interface{}{
		"peer_did":               did.ID,
		"transaction_receipt_id": tx.TransactionReceiptID,
		"amount":                 "0.50",
		"memo":                   "settlement mismatch",
	})
	require.NoError(t, err)

	denied, ok := result.(*toolpayment.PaymentExecutionDeniedResult)
	require.True(t, ok)
	assert.Equal(t, "execution_mode_mismatch", denied.Reason)
	assert.Contains(t, denied.Message, "prepay")

	_, events, err := receiptStore.GetSubmissionReceipt(context.Background(), sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, receipts.EventPaymentExecutionDenied, events[1].Type)
	assert.Equal(t, "execution_mode_mismatch", events[1].Reason)

	require.Len(t, auditor.entries, 1)
	assert.Equal(t, "denied", auditor.entries[0].Outcome)
	assert.Equal(t, "execution_mode_mismatch", auditor.entries[0].Reason)
}

type fakeP2PAuditor struct {
	entries []toolpayment.PaymentExecutionAuditEntry
}

func (f *fakeP2PAuditor) RecordPaymentExecution(_ context.Context, entry toolpayment.PaymentExecutionAuditEntry) error {
	f.entries = append(f.entries, entry)
	return nil
}
