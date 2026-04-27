package app

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/langoai/lango/internal/ent/paymenttx"
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
		SuggestedMode: paymentapproval.ModePrepay,
	})
	require.NoError(t, err)
	otherSub, _, err := receiptStore.CreateSubmissionReceipt(context.Background(), receipts.CreateSubmissionInput{
		TransactionID:       "tx-p2p-deny",
		ArtifactLabel:       "artifact-second",
		PayloadHash:         "hash-second",
		SourceLineageDigest: "lineage-second",
	})
	require.NoError(t, err)
	_, err = receiptStore.ApplyUpfrontPaymentApproval(context.Background(), tx.TransactionReceiptID, otherSub.SubmissionReceiptID, paymentapproval.Outcome{
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
	pc, cleanup := newTestP2PPaymentComponents(t)
	t.Cleanup(cleanup)
	p2pc := &p2pComponents{sessions: sessions}
	tools := buildP2PPaymentTool(p2pc, pc, receiptStore, auditor)
	require.Len(t, tools, 1)

	result, err := tools[0].Handler(context.Background(), map[string]interface{}{
		"peer_did":               did.ID,
		"transaction_receipt_id": tx.TransactionReceiptID,
		"submission_receipt_id":  otherSub.SubmissionReceiptID,
		"amount":                 "0.50",
		"memo":                   "settlement mismatch",
	})
	require.NoError(t, err)

	denied, ok := result.(*toolpayment.PaymentExecutionDeniedResult)
	require.True(t, ok)
	assert.Equal(t, "execution_mode_mismatch", denied.Reason)
	assert.Contains(t, denied.Message, "prepay")

	_, events, err := receiptStore.GetSubmissionReceipt(context.Background(), otherSub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, receipts.EventPaymentExecutionDenied, events[1].Type)
	assert.Equal(t, "execution_mode_mismatch", events[1].Reason)

	_, otherEvents, err := receiptStore.GetSubmissionReceipt(context.Background(), sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, otherEvents, 1)
	assert.Equal(t, receipts.EventPaymentApproval, otherEvents[0].Type)

	require.Len(t, auditor.entries, 1)
	assert.Equal(t, "denied", auditor.entries[0].Outcome)
	assert.Equal(t, "execution_mode_mismatch", auditor.entries[0].Reason)
}

func TestP2PPayment_AllowsAndAppendsAuthorizedTrailWhenSubmissionReceiptIDIsOmitted(t *testing.T) {
	t.Parallel()

	receiptStore := receipts.NewStore()
	sub, tx, err := receiptStore.CreateSubmissionReceipt(context.Background(), receipts.CreateSubmissionInput{
		TransactionID:       "tx-p2p-allow",
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

	pk, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	did, err := identity.DIDFromPublicKey(ethcrypto.CompressPubkey(&pk.PublicKey))
	require.NoError(t, err)

	sessions, err := handshake.NewSessionStore(time.Hour)
	require.NoError(t, err)
	_, err = sessions.Create(did.ID, false)
	require.NoError(t, err)

	auditor := &fakeP2PAuditor{}
	pc, cleanup := newTestP2PPaymentComponents(t)
	t.Cleanup(cleanup)
	p2pc := &p2pComponents{sessions: sessions}
	tools := buildP2PPaymentTool(p2pc, pc, receiptStore, auditor)
	require.Len(t, tools, 1)

	result, err := tools[0].Handler(context.Background(), map[string]interface{}{
		"peer_did":               did.ID,
		"transaction_receipt_id": tx.TransactionReceiptID,
		"amount":                 "0.50",
		"memo":                   "authorized payment",
	})
	require.NoError(t, err)

	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "confirmed", payload["status"])
	assert.Equal(t, did.ID, payload["peerDID"])

	_, events, err := receiptStore.GetSubmissionReceipt(context.Background(), sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, receipts.EventPaymentExecutionAuthorized, events[1].Type)
	assert.Equal(t, "payment_execution", events[1].Source)
	assert.Equal(t, "authorized", events[1].Subtype)
	assert.Equal(t, sub.SubmissionReceiptID, events[1].SubmissionReceiptID)

	require.Len(t, auditor.entries, 1)
	assert.Equal(t, "authorized", auditor.entries[0].Outcome)
	assert.Equal(t, tx.TransactionReceiptID, auditor.entries[0].TransactionReceiptID)
	assert.Equal(t, sub.SubmissionReceiptID, auditor.entries[0].SubmissionReceiptID)
}

func TestP2PPayment_FailsWhenAuditRecorderIsMissing(t *testing.T) {
	t.Parallel()

	receiptStore := receipts.NewStore()
	sub, tx, err := receiptStore.CreateSubmissionReceipt(context.Background(), receipts.CreateSubmissionInput{
		TransactionID:       "tx-p2p-no-audit",
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

	pk, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	did, err := identity.DIDFromPublicKey(ethcrypto.CompressPubkey(&pk.PublicKey))
	require.NoError(t, err)

	sessions, err := handshake.NewSessionStore(time.Hour)
	require.NoError(t, err)
	_, err = sessions.Create(did.ID, false)
	require.NoError(t, err)

	pc, cleanup := newTestP2PPaymentComponents(t)
	t.Cleanup(cleanup)
	p2pc := &p2pComponents{sessions: sessions}
	tools := buildP2PPaymentTool(p2pc, pc, receiptStore, nil)
	require.Len(t, tools, 1)

	_, err = tools[0].Handler(context.Background(), map[string]interface{}{
		"peer_did":               did.ID,
		"transaction_receipt_id": tx.TransactionReceiptID,
		"submission_receipt_id":  sub.SubmissionReceiptID,
		"amount":                 "0.50",
		"memo":                   "missing audit recorder",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "payment execution audit recorder is required")
}

func TestP2PPayment_FailsWhenReceiptTrailIsMissing(t *testing.T) {
	t.Parallel()

	pk, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	did, err := identity.DIDFromPublicKey(ethcrypto.CompressPubkey(&pk.PublicKey))
	require.NoError(t, err)

	sessions, err := handshake.NewSessionStore(time.Hour)
	require.NoError(t, err)
	_, err = sessions.Create(did.ID, false)
	require.NoError(t, err)

	pc, cleanup := newTestP2PPaymentComponents(t)
	t.Cleanup(cleanup)
	p2pc := &p2pComponents{sessions: sessions}
	tools := buildP2PPaymentTool(p2pc, pc, nil, &fakeP2PAuditor{})
	require.Len(t, tools, 1)

	_, err = tools[0].Handler(context.Background(), map[string]interface{}{
		"peer_did":               did.ID,
		"transaction_receipt_id": "tx-no-trail",
		"amount":                 "0.50",
		"memo":                   "missing receipt trail",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "payment execution receipt trail is required")
}

type fakeP2PAuditor struct {
	entries []toolpayment.PaymentExecutionAuditEntry
}

func (f *fakeP2PAuditor) RecordPaymentExecution(_ context.Context, entry toolpayment.PaymentExecutionAuditEntry) error {
	f.entries = append(f.entries, entry)
	return nil
}

type p2pTestWallet struct {
	key     *ecdsa.PrivateKey
	address string
}

func newP2PTestWallet(t *testing.T) *p2pTestWallet {
	t.Helper()

	key, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	return &p2pTestWallet{
		key:     key,
		address: ethcrypto.PubkeyToAddress(key.PublicKey).Hex(),
	}
}

func (w *p2pTestWallet) Address(context.Context) (string, error) {
	return w.address, nil
}

func (w *p2pTestWallet) Balance(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (w *p2pTestWallet) SignTransaction(_ context.Context, rawTx []byte) ([]byte, error) {
	return ethcrypto.Sign(rawTx, w.key)
}

func (w *p2pTestWallet) SignMessage(context.Context, []byte) ([]byte, error) {
	return nil, nil
}

func (w *p2pTestWallet) PublicKey(context.Context) ([]byte, error) {
	return ethcrypto.CompressPubkey(&w.key.PublicKey), nil
}

type p2pTestLimiter struct{}

func (p2pTestLimiter) Check(context.Context, *big.Int) error            { return nil }
func (p2pTestLimiter) Record(context.Context, *big.Int) error           { return nil }
func (p2pTestLimiter) DailySpent(context.Context) (*big.Int, error)     { return big.NewInt(0), nil }
func (p2pTestLimiter) DailyRemaining(context.Context) (*big.Int, error) { return big.NewInt(0), nil }
func (p2pTestLimiter) IsAutoApprovable(context.Context, *big.Int) (bool, error) {
	return false, nil
}

type p2pTestTxStore struct {
	mu      sync.Mutex
	records map[uuid.UUID]corepayment.TxRecord
}

func newP2PTestTxStore() *p2pTestTxStore {
	return &p2pTestTxStore{records: make(map[uuid.UUID]corepayment.TxRecord)}
}

func (s *p2pTestTxStore) Create(_ context.Context, record corepayment.TxRecord) (corepayment.TxRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}
	s.records[record.ID] = record
	return record, nil
}

func (s *p2pTestTxStore) UpdateStatus(_ context.Context, id uuid.UUID, status paymenttx.Status, txHash, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := s.records[id]
	record.Status = status
	if txHash != "" {
		record.TxHash = txHash
	}
	if errMsg != "" {
		record.ErrorMessage = errMsg
	}
	s.records[id] = record
	return nil
}

func (s *p2pTestTxStore) List(context.Context, int) ([]corepayment.TxRecord, error) {
	return nil, nil
}

func (s *p2pTestTxStore) DailySpendSince(context.Context, time.Time) ([]string, error) {
	return nil, nil
}

type p2pTestRPCServer struct {
	mu      sync.Mutex
	receipt *types.Receipt
}

func newP2PTestRPCServer() *httptest.Server {
	server := &p2pTestRPCServer{}
	return httptest.NewServer(http.HandlerFunc(server.handle))
}

func (s *p2pTestRPCServer) handle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		JSONRPC string            `json:"jsonrpc"`
		ID      any               `json:"id"`
		Method  string            `json:"method"`
		Params  []json.RawMessage `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRPCError(w, req.ID, -32700, err.Error())
		return
	}

	switch req.Method {
	case "eth_getTransactionCount":
		writeRPCResult(w, req.ID, "0x0")
	case "eth_estimateGas":
		writeRPCResult(w, req.ID, "0x5208")
	case "eth_getBlockByNumber":
		header := &types.Header{
			ParentHash:  ethcommon.Hash{},
			UncleHash:   ethcommon.Hash{},
			Root:        ethcommon.Hash{},
			TxHash:      ethcommon.Hash{},
			ReceiptHash: ethcommon.Hash{},
			Bloom:       types.Bloom{},
			Difficulty:  big.NewInt(0),
			Number:      big.NewInt(1),
			GasLimit:    30_000_000,
			GasUsed:     0,
			Time:        1,
			Extra:       []byte{},
			BaseFee:     big.NewInt(1_000_000_000),
		}
		payload, err := json.Marshal(header)
		if err != nil {
			writeRPCError(w, req.ID, -32000, err.Error())
			return
		}
		var result map[string]any
		if err := json.Unmarshal(payload, &result); err != nil {
			writeRPCError(w, req.ID, -32000, err.Error())
			return
		}
		writeRPCResult(w, req.ID, result)
	case "eth_sendRawTransaction":
		var rawHex string
		if err := json.Unmarshal(req.Params[0], &rawHex); err != nil {
			writeRPCError(w, req.ID, -32602, err.Error())
			return
		}
		raw, err := hexutil.Decode(rawHex)
		if err != nil {
			writeRPCError(w, req.ID, -32602, err.Error())
			return
		}
		var tx types.Transaction
		if err := tx.UnmarshalBinary(raw); err != nil {
			writeRPCError(w, req.ID, -32000, err.Error())
			return
		}

		receipt := &types.Receipt{
			Status:            types.ReceiptStatusSuccessful,
			CumulativeGasUsed: 21000,
			Bloom:             types.Bloom{},
			Logs:              []*types.Log{},
			TxHash:            tx.Hash(),
			GasUsed:           21000,
			BlockNumber:       big.NewInt(42),
		}

		s.mu.Lock()
		s.receipt = receipt
		s.mu.Unlock()

		writeRPCResult(w, req.ID, tx.Hash().Hex())
	case "eth_getTransactionReceipt":
		s.mu.Lock()
		receipt := s.receipt
		s.mu.Unlock()
		if receipt == nil {
			writeRPCError(w, req.ID, -32000, "not found")
			return
		}
		writeRPCResult(w, req.ID, receipt)
	default:
		writeRPCError(w, req.ID, -32601, "method not found")
	}
}

func writeRPCResult(w http.ResponseWriter, id any, result any) {
	_ = json.NewEncoder(w).Encode(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	})
}

func writeRPCError(w http.ResponseWriter, id any, code int, message string) {
	_ = json.NewEncoder(w).Encode(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}

func newTestP2PPaymentComponents(t *testing.T) (*paymentComponents, func()) {
	t.Helper()

	server := newP2PTestRPCServer()
	client, err := ethclient.Dial(server.URL)
	require.NoError(t, err)

	w := newP2PTestWallet(t)
	store := newP2PTestTxStore()
	builder := corepayment.NewTxBuilder(client, 84532, "0x4200000000000000000000000000000000000006")
	svc := corepayment.NewService(w, p2pTestLimiter{}, builder, store, client, 84532)

	return &paymentComponents{service: svc}, func() {
		server.Close()
		client.Close()
	}
}
