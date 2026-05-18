package settlement

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent/paymenttx"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/payment/eip3009"
)

type wave10TxStore struct {
	createErr error
	updateErr error
	created   []payment.TxRecord
	updates   []wave10StatusUpdate
}

type wave10StatusUpdate struct {
	id     uuid.UUID
	status paymenttx.Status
	txHash string
	errMsg string
}

func (s *wave10TxStore) Create(_ context.Context, record payment.TxRecord) (payment.TxRecord, error) {
	if s.createErr != nil {
		return payment.TxRecord{}, s.createErr
	}
	s.created = append(s.created, record)
	return record, nil
}

func (s *wave10TxStore) UpdateStatus(_ context.Context, id uuid.UUID, status paymenttx.Status, txHash, errMsg string) error {
	s.updates = append(s.updates, wave10StatusUpdate{
		id:     id,
		status: status,
		txHash: txHash,
		errMsg: errMsg,
	})
	return s.updateErr
}

func (s *wave10TxStore) List(context.Context, int) ([]payment.TxRecord, error) {
	return nil, nil
}

func (s *wave10TxStore) DailySpendSince(context.Context, time.Time) ([]string, error) {
	return nil, nil
}

type wave10Wallet struct {
	addressErr error
}

func (w wave10Wallet) Address(context.Context) (string, error) {
	if w.addressErr != nil {
		return "", w.addressErr
	}
	return "0x1234567890abcdef1234567890abcdef12345678", nil
}

func (w wave10Wallet) Balance(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (w wave10Wallet) SignTransaction(context.Context, []byte) ([]byte, error) {
	return nil, errors.New("unexpected sign")
}

func (w wave10Wallet) SignMessage(context.Context, []byte) ([]byte, error) {
	return nil, errors.New("unexpected sign message")
}

func (w wave10Wallet) PublicKey(context.Context) ([]byte, error) {
	return nil, nil
}

type wave10AsyncRecorder struct {
	successes chan string
	failures  chan string
}

func (r wave10AsyncRecorder) RecordSuccess(_ context.Context, peerDID string) error {
	r.successes <- peerDID
	return nil
}

func (r wave10AsyncRecorder) RecordFailure(_ context.Context, peerDID string) error {
	r.failures <- peerDID
	return nil
}

func TestSubscribeCloseWave10_WaitsForPublishedSettlementFailure(t *testing.T) {
	t.Parallel()

	rec := wave10AsyncRecorder{
		successes: make(chan string, 1),
		failures:  make(chan string, 1),
	}
	svc := New(Config{
		Logger:         zap.NewNop().Sugar(),
		ReceiptTimeout: time.Millisecond,
	})
	svc.SetReputationRecorder(rec)
	bus := eventbus.New()
	svc.Subscribe(bus)

	bus.Publish(eventbus.ToolExecutionPaidEvent{
		PeerDID:  "did:peer:wave10",
		ToolName: "paid-tool",
		Auth:     wave10Authorization(),
	})

	svc.Close()

	select {
	case got := <-rec.failures:
		assert.Equal(t, "did:peer:wave10", got)
	default:
		t.Fatal("Close returned before settlement failure was recorded")
	}
	select {
	case got := <-rec.successes:
		t.Fatalf("unexpected success recorded for %s", got)
	default:
	}
}

func TestHandleEventWave10_InvalidAuthDoesNotRecordReputation(t *testing.T) {
	t.Parallel()

	rec := wave10AsyncRecorder{
		successes: make(chan string, 1),
		failures:  make(chan string, 1),
	}
	svc := New(Config{Logger: zap.NewNop().Sugar()})
	svc.SetReputationRecorder(rec)

	svc.handleEvent(eventbus.ToolExecutionPaidEvent{
		PeerDID:  "did:peer:invalid-auth",
		ToolName: "paid-tool",
		Auth:     map[string]string{"not": "authorization"},
	})

	assert.Empty(t, rec.successes)
	assert.Empty(t, rec.failures)
}

func TestSettleWave10_ErrorAndUpdateBranches(t *testing.T) {
	t.Parallel()

	t.Run("missing store returns before creating record", func(t *testing.T) {
		t.Parallel()

		svc := New(Config{Logger: zap.NewNop().Sugar(), ChainID: 84532})
		err := svc.settle(context.Background(), wave10Authorization(), "did:peer:no-store", "tool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "payment tx store not configured")
	})

	t.Run("create error is returned without status update", func(t *testing.T) {
		t.Parallel()

		store := &wave10TxStore{createErr: errors.New("create failed")}
		svc := New(Config{Logger: zap.NewNop().Sugar(), TxStore: store, ChainID: 84532})
		err := svc.settle(context.Background(), wave10Authorization(), "did:peer:create", "tool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create payment record")
		assert.Empty(t, store.updates)
	})

	t.Run("build failure marks record failed", func(t *testing.T) {
		t.Parallel()

		store := &wave10TxStore{}
		svc := New(Config{
			Logger:  zap.NewNop().Sugar(),
			TxStore: store,
			Wallet:  wave10Wallet{addressErr: errors.New("wallet offline")},
			ChainID: 84532,
		})

		err := svc.settle(context.Background(), wave10Authorization(), "did:peer:build", "tool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "build/sign tx")
		require.Len(t, store.created, 1)
		require.Len(t, store.updates, 1)
		assert.Equal(t, store.created[0].ID, store.updates[0].id)
		assert.Equal(t, paymenttx.StatusFailed, store.updates[0].status)
		assert.Empty(t, store.updates[0].txHash)
		assert.Contains(t, store.updates[0].errMsg, "wallet offline")
		assert.Equal(t, paymenttx.PaymentMethodP2pSettlement, store.created[0].PaymentMethod)
		assert.Contains(t, store.created[0].Purpose, "tool")
		assert.Contains(t, store.created[0].Purpose, "did:peer:build")
	})

	t.Run("update status errors are swallowed", func(t *testing.T) {
		t.Parallel()

		store := &wave10TxStore{updateErr: errors.New("update failed")}
		svc := New(Config{
			Logger:  zap.NewNop().Sugar(),
			TxStore: store,
			Wallet:  wave10Wallet{addressErr: errors.New("wallet offline")},
			ChainID: 84532,
		})

		err := svc.settle(context.Background(), wave10Authorization(), "did:peer:update", "tool")
		require.Error(t, err)
		require.Len(t, store.updates, 1)
		assert.Equal(t, paymenttx.StatusFailed, store.updates[0].status)
	})
}

func wave10Authorization() *eip3009.Authorization {
	return &eip3009.Authorization{
		From:        common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		To:          common.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		Value:       big.NewInt(500000),
		ValidAfter:  big.NewInt(0),
		ValidBefore: big.NewInt(time.Now().Add(10 * time.Minute).Unix()),
		Nonce:       [32]byte{1},
		V:           27,
		R:           [32]byte{2},
		S:           [32]byte{3},
	}
}
