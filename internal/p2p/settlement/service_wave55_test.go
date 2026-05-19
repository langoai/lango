package settlement

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent/paymenttx"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/payment/eip3009"
)

type wave55TxStore struct {
	createErr error
	created   []payment.TxRecord
	updates   []wave55StatusUpdate
}

type wave55StatusUpdate struct {
	id     uuid.UUID
	status paymenttx.Status
	txHash string
	errMsg string
}

func (s *wave55TxStore) Create(_ context.Context, record payment.TxRecord) (payment.TxRecord, error) {
	if s.createErr != nil {
		return payment.TxRecord{}, s.createErr
	}
	s.created = append(s.created, record)
	return record, nil
}

func (s *wave55TxStore) UpdateStatus(_ context.Context, id uuid.UUID, status paymenttx.Status, txHash, errMsg string) error {
	s.updates = append(s.updates, wave55StatusUpdate{id: id, status: status, txHash: txHash, errMsg: errMsg})
	return nil
}

func (s *wave55TxStore) List(context.Context, int) ([]payment.TxRecord, error) {
	return append([]payment.TxRecord(nil), s.created...), nil
}

func (s *wave55TxStore) DailySpendSince(context.Context, time.Time) ([]string, error) {
	return nil, nil
}

type wave55Wallet struct {
	addressErr error
}

func (w wave55Wallet) Address(context.Context) (string, error) {
	if w.addressErr != nil {
		return "", w.addressErr
	}
	return "0x1234567890abcdef1234567890abcdef12345678", nil
}

func (w wave55Wallet) Balance(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (w wave55Wallet) SignTransaction(context.Context, []byte) ([]byte, error) {
	return nil, errors.New("unexpected sign transaction")
}

func (w wave55Wallet) SignMessage(context.Context, []byte) ([]byte, error) {
	return nil, errors.New("unexpected sign message")
}

func (w wave55Wallet) PublicKey(context.Context) ([]byte, error) {
	return nil, nil
}

type wave55ReputationRecorder struct {
	successes []string
	failures  []string
}

func (r *wave55ReputationRecorder) RecordSuccess(_ context.Context, peerDID string) error {
	r.successes = append(r.successes, peerDID)
	return nil
}

func (r *wave55ReputationRecorder) RecordFailure(_ context.Context, peerDID string) error {
	r.failures = append(r.failures, peerDID)
	return nil
}

func TestSettleWave55_LocalStoreRecordsDeterministicPaymentBeforeBuildFailure(t *testing.T) {
	t.Parallel()

	store := &wave55TxStore{}
	auth := wave55Authorization()
	svc := New(Config{
		Logger:  zap.NewNop().Sugar(),
		TxStore: store,
		Wallet:  wave55Wallet{addressErr: errors.New("wallet unavailable")},
		ChainID: 84532,
	})

	err := svc.settle(context.Background(), auth, "did:peer:wave55", "summarize")
	require.Error(t, err)
	require.Contains(t, err.Error(), "build/sign tx")
	require.Len(t, store.created, 1)
	require.Len(t, store.updates, 1)

	record := store.created[0]
	require.Equal(t, auth.From.Hex(), record.FromAddress)
	require.Equal(t, auth.To.Hex(), record.ToAddress)
	require.Equal(t, auth.Value.String(), record.Amount)
	require.Equal(t, int64(84532), record.ChainID)
	require.Equal(t, paymenttx.StatusPending, record.Status)
	require.Equal(t, paymenttx.PaymentMethodP2pSettlement, record.PaymentMethod)
	require.Contains(t, record.Purpose, "summarize")
	require.Contains(t, record.Purpose, "did:peer:wave55")

	update := store.updates[0]
	require.Equal(t, record.ID, update.id)
	require.Equal(t, paymenttx.StatusFailed, update.status)
	require.Empty(t, update.txHash)
	require.Contains(t, update.errMsg, "wallet unavailable")
}

func TestHandleEventWave55_ValidationAndNilDependencyBranches(t *testing.T) {
	t.Parallel()

	t.Run("invalid auth does not touch reputation", func(t *testing.T) {
		t.Parallel()

		rec := &wave55ReputationRecorder{}
		svc := New(Config{Logger: zap.NewNop().Sugar()})
		svc.SetReputationRecorder(rec)

		svc.handleEvent(eventbus.ToolExecutionPaidEvent{
			PeerDID:  "did:peer:invalid",
			ToolName: "summarize",
			Auth:     map[string]string{"from": "not an eip3009 authorization"},
		})

		require.Empty(t, rec.successes)
		require.Empty(t, rec.failures)
	})

	t.Run("missing store records settlement failure", func(t *testing.T) {
		t.Parallel()

		rec := &wave55ReputationRecorder{}
		svc := New(Config{Logger: zap.NewNop().Sugar()})
		svc.SetReputationRecorder(rec)

		svc.handleEvent(eventbus.ToolExecutionPaidEvent{
			PeerDID:  "did:peer:no-store",
			ToolName: "summarize",
			Auth:     wave55Authorization(),
		})

		require.Empty(t, rec.successes)
		require.Equal(t, []string{"did:peer:no-store"}, rec.failures)
	})
}

func wave55Authorization() *eip3009.Authorization {
	return &eip3009.Authorization{
		From:        common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		To:          common.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		Value:       big.NewInt(123456),
		ValidAfter:  big.NewInt(0),
		ValidBefore: big.NewInt(1893456000),
		Nonce:       [32]byte{0x55},
		V:           27,
		R:           [32]byte{0x01},
		S:           [32]byte{0x02},
	}
}
