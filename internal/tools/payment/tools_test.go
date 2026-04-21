package payment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/langoai/lango/internal/agent"
	corepayment "github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/paymentgate"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/x402"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBuildTools_BaseSet(t *testing.T) {
	t.Parallel()

	tools := BuildTools(nil, nil, nil, 84532, nil, nil, nil)

	require.Len(t, tools, 5, "base set: send, balance, history, limits, wallet_info")

	names := toolNames(tools)
	for _, name := range []string{"payment_send", "payment_balance", "payment_history", "payment_limits", "payment_wallet_info"} {
		assert.Contains(t, names, name)
	}

	// Conditional tools absent without secrets/interceptor.
	assert.NotContains(t, names, "payment_create_wallet")
	assert.NotContains(t, names, "payment_x402_fetch")
}

func TestBuildTools_SafetyLevels(t *testing.T) {
	t.Parallel()

	tools := BuildTools(nil, nil, nil, 84532, nil, nil, nil)

	levels := make(map[string]agent.SafetyLevel, len(tools))
	for _, tool := range tools {
		levels[tool.Name] = tool.SafetyLevel
	}

	assert.Equal(t, agent.SafetyLevelDangerous, levels["payment_send"], "send must be dangerous")

	for _, name := range []string{"payment_balance", "payment_history", "payment_limits", "payment_wallet_info"} {
		assert.Equal(t, agent.SafetyLevelSafe, levels[name], "%s must be safe", name)
	}
}

func TestBuildTools_ConditionalCreateWallet(t *testing.T) {
	t.Parallel()

	// Non-nil SecretsStore adds payment_create_wallet.
	secrets := &security.SecretsStore{}
	tools := BuildTools(nil, nil, secrets, 84532, nil, nil, nil)

	names := toolNames(tools)
	assert.Contains(t, names, "payment_create_wallet")
	assert.NotContains(t, names, "payment_x402_fetch")
}

func TestBuildTools_ConditionalX402(t *testing.T) {
	t.Parallel()

	// Interceptor with Enabled=true adds payment_x402_fetch.
	interceptor := x402.NewInterceptor(nil, nil, x402.Config{Enabled: true}, zap.NewNop().Sugar())
	tools := BuildTools(nil, nil, nil, 84532, interceptor, nil, nil)

	names := toolNames(tools)
	assert.Contains(t, names, "payment_x402_fetch")
}

func TestBuildTools_AllConditional(t *testing.T) {
	t.Parallel()

	secrets := &security.SecretsStore{}
	interceptor := x402.NewInterceptor(nil, nil, x402.Config{Enabled: true}, zap.NewNop().Sugar())
	tools := BuildTools(nil, nil, secrets, 84532, interceptor, nil, nil)

	require.Len(t, tools, 7, "all 7 tools with both secrets and interceptor")
}

func TestBuildTools_DisabledInterceptor(t *testing.T) {
	t.Parallel()

	// Interceptor with Enabled=false does NOT add payment_x402_fetch.
	interceptor := x402.NewInterceptor(nil, nil, x402.Config{Enabled: false}, zap.NewNop().Sugar())
	tools := BuildTools(nil, nil, nil, 84532, interceptor, nil, nil)

	names := toolNames(tools)
	assert.NotContains(t, names, "payment_x402_fetch")
	require.Len(t, tools, 5, "disabled interceptor = base set only")
}

func TestPaymentSend_DeniesWithoutTransactionReceiptID(t *testing.T) {
	t.Parallel()

	receiptStore := receipts.NewStore()
	sub, _, err := receiptStore.CreateSubmissionReceipt(context.Background(), receipts.CreateSubmissionInput{
		TransactionID:       "tx-missing-tx",
		ArtifactLabel:       "artifact",
		PayloadHash:         "hash",
		SourceLineageDigest: "lineage",
	})
	require.NoError(t, err)
	auditor := &fakeExecutionAuditor{}
	tools := BuildTools(&fakePaymentService{}, nil, nil, 84532, nil, receiptStore, auditor)

	sendTool := findTool(tools, "payment_send")
	require.NotNil(t, sendTool)

	result, err := sendTool.Handler(context.Background(), map[string]interface{}{
		"to":                    "0x1111111111111111111111111111111111111111",
		"submission_receipt_id": sub.SubmissionReceiptID,
		"amount":                "1.00",
		"purpose":               "missing tx receipt",
	})
	require.NoError(t, err)

	denied, ok := result.(*PaymentExecutionDeniedResult)
	require.True(t, ok)
	assert.Equal(t, "denied", denied.Status)
	assert.Equal(t, "missing_receipt", denied.Reason)
	assert.Contains(t, denied.Message, "transaction_receipt_id")

	require.Len(t, auditor.entries, 1)
	assert.Equal(t, "denied", auditor.entries[0].Outcome)
	assert.Equal(t, "missing_receipt", auditor.entries[0].Reason)
}

func TestPaymentSend_DeniesWithUnknownTransactionReceiptID(t *testing.T) {
	t.Parallel()

	receiptStore := receipts.NewStore()
	sub, _, err := receiptStore.CreateSubmissionReceipt(context.Background(), receipts.CreateSubmissionInput{
		TransactionID:       "tx-known-submission",
		ArtifactLabel:       "artifact",
		PayloadHash:         "hash",
		SourceLineageDigest: "lineage",
	})
	require.NoError(t, err)
	auditor := &fakeExecutionAuditor{}
	tools := BuildTools(&fakePaymentService{}, nil, nil, 84532, nil, receiptStore, auditor)

	sendTool := findTool(tools, "payment_send")
	require.NotNil(t, sendTool)

	result, err := sendTool.Handler(context.Background(), map[string]interface{}{
		"transaction_receipt_id": "missing-transaction",
		"submission_receipt_id":  sub.SubmissionReceiptID,
		"to":                     "0x1111111111111111111111111111111111111111",
		"amount":                 "1.00",
		"purpose":                "unknown tx receipt",
	})
	require.NoError(t, err)

	denied, ok := result.(*PaymentExecutionDeniedResult)
	require.True(t, ok)
	assert.Equal(t, "missing_receipt", denied.Reason)
	assert.Contains(t, denied.Message, "was not found")

	require.Len(t, auditor.entries, 1)
	assert.Equal(t, "denied", auditor.entries[0].Outcome)
	assert.Equal(t, "missing_receipt", auditor.entries[0].Reason)
}

func TestPaymentSend_DeniesWhenApprovalIsNotApproved(t *testing.T) {
	t.Parallel()

	receiptStore := receipts.NewStore()
	sub, tx, err := receiptStore.CreateSubmissionReceipt(context.Background(), receipts.CreateSubmissionInput{
		TransactionID:       "tx-deny-approval",
		ArtifactLabel:       "artifact",
		PayloadHash:         "hash",
		SourceLineageDigest: "lineage",
	})
	require.NoError(t, err)
	_, err = receiptStore.ApplyUpfrontPaymentApproval(context.Background(), tx.TransactionReceiptID, sub.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionReject,
		Reason:        "not approved",
		SuggestedMode: paymentapproval.ModeReject,
	})
	require.NoError(t, err)

	auditor := &fakeExecutionAuditor{}
	tools := BuildTools(&fakePaymentService{}, nil, nil, 84532, nil, receiptStore, auditor)
	sendTool := findTool(tools, "payment_send")
	require.NotNil(t, sendTool)

	result, err := sendTool.Handler(context.Background(), map[string]interface{}{
		"to":                     "0x1111111111111111111111111111111111111111",
		"transaction_receipt_id": tx.TransactionReceiptID,
		"submission_receipt_id":  sub.SubmissionReceiptID,
		"amount":                 "1.00",
		"purpose":                "approval denied",
	})
	require.NoError(t, err)

	denied, ok := result.(*PaymentExecutionDeniedResult)
	require.True(t, ok)
	assert.Equal(t, "approval_not_approved", denied.Reason)
	assert.Contains(t, denied.Message, "canonical payment approval")

	_, events, err := receiptStore.GetSubmissionReceipt(context.Background(), sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, receipts.EventPaymentExecutionDenied, events[1].Type)
	assert.Equal(t, "approval_not_approved", events[1].Reason)

	require.Len(t, auditor.entries, 1)
	assert.Equal(t, "denied", auditor.entries[0].Outcome)
	assert.Equal(t, "approval_not_approved", auditor.entries[0].Reason)
}

func TestPaymentSend_AllowPathRecordsSuccessEvents(t *testing.T) {
	t.Parallel()

	receiptStore := receipts.NewStore()
	sub, tx, err := receiptStore.CreateSubmissionReceipt(context.Background(), receipts.CreateSubmissionInput{
		TransactionID:       "tx-allow",
		ArtifactLabel:       "artifact",
		PayloadHash:         "hash",
		SourceLineageDigest: "lineage",
	})
	require.NoError(t, err)
	_, err = receiptStore.ApplyUpfrontPaymentApproval(context.Background(), tx.TransactionReceiptID, sub.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "approved",
		SuggestedMode: paymentapproval.ModePrepay,
	})
	require.NoError(t, err)

	auditor := &fakeExecutionAuditor{}
	fakeSvc := &fakePaymentService{
		receipt: &corepayment.PaymentReceipt{
			TxHash:      "0xabc",
			Status:      "confirmed",
			Amount:      "1.00",
			From:        "0x2222222222222222222222222222222222222222",
			To:          "0x1111111111111111111111111111111111111111",
			ChainID:     84532,
			GasUsed:     21000,
			BlockNumber: 42,
			Timestamp:   time.Unix(1700000000, 0).UTC(),
		},
	}
	tools := BuildTools(fakeSvc, nil, nil, 84532, nil, receiptStore, auditor)
	sendTool := findTool(tools, "payment_send")
	require.NotNil(t, sendTool)

	result, err := sendTool.Handler(context.Background(), map[string]interface{}{
		"to":                     "0x1111111111111111111111111111111111111111",
		"transaction_receipt_id": tx.TransactionReceiptID,
		"submission_receipt_id":  sub.SubmissionReceiptID,
		"amount":                 "1.00",
		"purpose":                "approved payment",
	})
	require.NoError(t, err)

	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "confirmed", payload["status"])
	assert.Equal(t, "0xabc", payload["txHash"])

	_, events, err := receiptStore.GetSubmissionReceipt(context.Background(), sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, receipts.EventPaymentExecutionAuthorized, events[1].Type)
	assert.Empty(t, events[1].Reason)

	require.Len(t, auditor.entries, 1)
	assert.Equal(t, "authorized", auditor.entries[0].Outcome)
	assert.Equal(t, tx.TransactionReceiptID, auditor.entries[0].TransactionReceiptID)
	assert.Equal(t, sub.SubmissionReceiptID, auditor.entries[0].SubmissionReceiptID)
}

func TestPaymentSend_FailsWhenAuditRecorderIsMissing(t *testing.T) {
	t.Parallel()

	receiptStore := receipts.NewStore()
	sub, tx, err := receiptStore.CreateSubmissionReceipt(context.Background(), receipts.CreateSubmissionInput{
		TransactionID:       "tx-no-audit",
		ArtifactLabel:       "artifact",
		PayloadHash:         "hash",
		SourceLineageDigest: "lineage",
	})
	require.NoError(t, err)
	_, err = receiptStore.ApplyUpfrontPaymentApproval(context.Background(), tx.TransactionReceiptID, sub.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "approved",
		SuggestedMode: paymentapproval.ModePrepay,
	})
	require.NoError(t, err)

	tools := BuildTools(&fakePaymentService{}, nil, nil, 84532, nil, receiptStore, nil)
	sendTool := findTool(tools, "payment_send")
	require.NotNil(t, sendTool)

	_, err = sendTool.Handler(context.Background(), map[string]interface{}{
		"to":                     "0x1111111111111111111111111111111111111111",
		"transaction_receipt_id": tx.TransactionReceiptID,
		"submission_receipt_id":  sub.SubmissionReceiptID,
		"amount":                 "1.00",
		"purpose":                "missing audit recorder",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "payment execution audit recorder is required")
}

func TestPaymentSend_RecordsAuthorizedEventOnCurrentSubmission(t *testing.T) {
	t.Parallel()

	receiptStore := receipts.NewStore()
	firstSub, tx, err := receiptStore.CreateSubmissionReceipt(context.Background(), receipts.CreateSubmissionInput{
		TransactionID:       "tx-multi-submission",
		ArtifactLabel:       "artifact-first",
		PayloadHash:         "hash-first",
		SourceLineageDigest: "lineage-first",
	})
	require.NoError(t, err)
	_, err = receiptStore.ApplyUpfrontPaymentApproval(context.Background(), tx.TransactionReceiptID, firstSub.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "approved",
		SuggestedMode: paymentapproval.ModePrepay,
	})
	require.NoError(t, err)

	secondSub, _, err := receiptStore.CreateSubmissionReceipt(context.Background(), receipts.CreateSubmissionInput{
		TransactionID:       "tx-multi-submission",
		ArtifactLabel:       "artifact-second",
		PayloadHash:         "hash-second",
		SourceLineageDigest: "lineage-second",
	})
	require.NoError(t, err)

	auditor := &fakeExecutionAuditor{}
	tools := BuildTools(&fakePaymentService{
		receipt: &corepayment.PaymentReceipt{
			TxHash:      "0xabc",
			Status:      "confirmed",
			Amount:      "1.00",
			From:        "0x2222222222222222222222222222222222222222",
			To:          "0x1111111111111111111111111111111111111111",
			ChainID:     84532,
			GasUsed:     21000,
			BlockNumber: 42,
			Timestamp:   time.Unix(1700000000, 0).UTC(),
		},
	}, nil, nil, 84532, nil, receiptStore, auditor)
	sendTool := findTool(tools, "payment_send")
	require.NotNil(t, sendTool)

	result, err := sendTool.Handler(context.Background(), map[string]interface{}{
		"to":                     "0x1111111111111111111111111111111111111111",
		"transaction_receipt_id": tx.TransactionReceiptID,
		"submission_receipt_id":  secondSub.SubmissionReceiptID,
		"amount":                 "1.00",
		"purpose":                "current canonical submission",
	})
	require.NoError(t, err)

	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "confirmed", payload["status"])

	_, firstEvents, err := receiptStore.GetSubmissionReceipt(context.Background(), firstSub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, firstEvents, 1)
	assert.Equal(t, receipts.EventPaymentApproval, firstEvents[0].Type)

	_, secondEvents, err := receiptStore.GetSubmissionReceipt(context.Background(), secondSub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, secondEvents, 1)
	assert.Equal(t, receipts.EventPaymentExecutionAuthorized, secondEvents[0].Type)
	assert.Equal(t, secondSub.SubmissionReceiptID, secondEvents[0].SubmissionReceiptID)

	require.Len(t, auditor.entries, 1)
	assert.Equal(t, "authorized", auditor.entries[0].Outcome)
	assert.Equal(t, tx.TransactionReceiptID, auditor.entries[0].TransactionReceiptID)
	assert.Equal(t, secondSub.SubmissionReceiptID, auditor.entries[0].SubmissionReceiptID)
}

func TestCheckDirectPaymentExecution_ReturnsErrorOnAuditWriteFailure(t *testing.T) {
	t.Parallel()

	allowedGate := fakeExecutionGate{result: paymentgate.Result{Decision: paymentgate.Allow}}
	trail := &fakeExecutionTrail{}
	auditor := &fakeExecutionAuditor{err: errors.New("audit write failed")}

	allowed, denied, err := CheckDirectPaymentExecution(context.Background(), "payment_send", "tx-1", "submission-1", &allowedGate, trail, auditor)
	require.Error(t, err)
	require.False(t, allowed)
	require.Nil(t, denied)
	assert.Contains(t, err.Error(), "record payment execution audit")
}

func TestCheckDirectPaymentExecution_ReturnsErrorOnTrailWriteFailure(t *testing.T) {
	t.Parallel()

	allowedGate := fakeExecutionGate{result: paymentgate.Result{Decision: paymentgate.Allow}}
	trail := &fakeExecutionTrail{allowErr: errors.New("trail write failed")}
	auditor := &fakeExecutionAuditor{}

	allowed, denied, err := CheckDirectPaymentExecution(context.Background(), "payment_send", "tx-1", "submission-1", &allowedGate, trail, auditor)
	require.Error(t, err)
	require.False(t, allowed)
	require.Nil(t, denied)
	assert.Contains(t, err.Error(), "record payment execution receipt trail")
}

// --- helpers ---

type fakeExecutionGate struct {
	result paymentgate.Result
	err    error
}

func (f *fakeExecutionGate) EvaluateDirectPayment(context.Context, paymentgate.Request) (paymentgate.Result, error) {
	if f.err != nil {
		return paymentgate.Result{}, f.err
	}
	return f.result, nil
}

type fakeExecutionTrail struct {
	allowErr error
	denyErr  error
}

func (f *fakeExecutionTrail) AppendPaymentExecutionAuthorized(context.Context, string) error {
	return f.allowErr
}

func (f *fakeExecutionTrail) AppendPaymentExecutionDenied(context.Context, string, string) error {
	return f.denyErr
}

type fakePaymentService struct {
	receipt *corepayment.PaymentReceipt
}

func (f *fakePaymentService) Send(context.Context, corepayment.PaymentRequest) (*corepayment.PaymentReceipt, error) {
	if f.receipt == nil {
		return &corepayment.PaymentReceipt{}, nil
	}
	return f.receipt, nil
}

func (f *fakePaymentService) Balance(context.Context) (string, error) {
	return "0.00", nil
}

func (f *fakePaymentService) History(context.Context, int) ([]corepayment.TransactionInfo, error) {
	return nil, nil
}

func (f *fakePaymentService) WalletAddress(context.Context) (string, error) {
	return "0x2222222222222222222222222222222222222222", nil
}

func (f *fakePaymentService) ChainID() int64 {
	return 84532
}

func (f *fakePaymentService) RecordX402Payment(context.Context, corepayment.X402PaymentRecord) error {
	return nil
}

type fakeExecutionAuditor struct {
	entries []PaymentExecutionAuditEntry
	err     error
}

func (f *fakeExecutionAuditor) RecordPaymentExecution(_ context.Context, entry PaymentExecutionAuditEntry) error {
	if f.err != nil {
		return f.err
	}
	f.entries = append(f.entries, entry)
	return nil
}

func toolNames(tools []*agent.Tool) map[string]bool {
	m := make(map[string]bool, len(tools))
	for _, tl := range tools {
		m[tl.Name] = true
	}
	return m
}

func findTool(tools []*agent.Tool, name string) *agent.Tool {
	for _, tl := range tools {
		if tl.Name == name {
			return tl
		}
	}
	return nil
}
