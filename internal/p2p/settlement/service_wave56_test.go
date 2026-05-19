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
	"github.com/langoai/lango/internal/payment"
)

func TestServiceWave56_DefaultsTreatNonPositiveValuesAsUnset(t *testing.T) {
	t.Parallel()

	svc := New(Config{
		ReceiptTimeout: -time.Second,
		MaxRetries:     -7,
		Logger:         zap.NewNop().Sugar(),
	})

	require.Equal(t, 2*time.Minute, svc.timeout)
	require.Equal(t, 3, svc.maxRetries)
}

func TestServiceWave56_NewPreservesExplicitDependenciesAndConfig(t *testing.T) {
	t.Parallel()

	store := &wave56TxStore{}
	wallet := wave56Wallet{}
	usdc := common.HexToAddress("0x1111111111111111111111111111111111111111")

	svc := New(Config{
		Wallet:         wallet,
		TxStore:        store,
		ChainID:        84532,
		USDCAddr:       usdc,
		ReceiptTimeout: 45 * time.Second,
		MaxRetries:     2,
		Logger:         zap.NewNop().Sugar(),
	})

	require.Equal(t, wallet, svc.wallet)
	require.Same(t, store, svc.store)
	require.Equal(t, int64(84532), svc.chainID.Int64())
	require.Equal(t, usdc, svc.usdcAddr)
	require.Equal(t, 45*time.Second, svc.timeout)
	require.Equal(t, 2, svc.maxRetries)
}

type wave56TxStore struct {
	created []payment.TxRecord
}

func (s *wave56TxStore) Create(_ context.Context, record payment.TxRecord) (payment.TxRecord, error) {
	s.created = append(s.created, record)
	return record, nil
}

func (s *wave56TxStore) UpdateStatus(context.Context, uuid.UUID, paymenttx.Status, string, string) error {
	return nil
}

func (s *wave56TxStore) List(context.Context, int) ([]payment.TxRecord, error) {
	return append([]payment.TxRecord(nil), s.created...), nil
}

func (s *wave56TxStore) DailySpendSince(context.Context, time.Time) ([]string, error) {
	return nil, nil
}

type wave56Wallet struct{}

func (w wave56Wallet) Address(context.Context) (string, error) {
	return "0x1234567890abcdef1234567890abcdef12345678", nil
}

func (w wave56Wallet) Balance(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (w wave56Wallet) SignTransaction(context.Context, []byte) ([]byte, error) {
	return nil, errors.New("unexpected sign transaction")
}

func (w wave56Wallet) SignMessage(context.Context, []byte) ([]byte, error) {
	return nil, errors.New("unexpected sign message")
}

func (w wave56Wallet) PublicKey(context.Context) ([]byte, error) {
	return nil, nil
}

var _ payment.TxStore = (*wave56TxStore)(nil)
