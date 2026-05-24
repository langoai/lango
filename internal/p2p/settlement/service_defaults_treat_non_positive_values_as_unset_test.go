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

func TestServiceDefaultsTreatNonPositiveValuesAsUnset(t *testing.T) {
	t.Parallel()

	svc := New(Config{
		ReceiptTimeout: -time.Second,
		MaxRetries:     -7,
		Logger:         zap.NewNop().Sugar(),
	})

	require.Equal(t, 2*time.Minute, svc.timeout)
	require.Equal(t, 3, svc.maxRetries)
}

func TestServiceNewPreservesExplicitDependenciesAndConfig(t *testing.T) {
	t.Parallel()

	store := &serviceDefaultsTreatNonPositiveValuesAsUnsetTxStore{}
	wallet := serviceDefaultsTreatNonPositiveValuesAsUnsetWallet{}
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

type serviceDefaultsTreatNonPositiveValuesAsUnsetTxStore struct {
	created []payment.TxRecord
}

func (s *serviceDefaultsTreatNonPositiveValuesAsUnsetTxStore) Create(_ context.Context, record payment.TxRecord) (payment.TxRecord, error) {
	s.created = append(s.created, record)
	return record, nil
}

func (s *serviceDefaultsTreatNonPositiveValuesAsUnsetTxStore) UpdateStatus(context.Context, uuid.UUID, paymenttx.Status, string, string) error {
	return nil
}

func (s *serviceDefaultsTreatNonPositiveValuesAsUnsetTxStore) List(context.Context, int) ([]payment.TxRecord, error) {
	return append([]payment.TxRecord(nil), s.created...), nil
}

func (s *serviceDefaultsTreatNonPositiveValuesAsUnsetTxStore) DailySpendSince(context.Context, time.Time) ([]string, error) {
	return nil, nil
}

type serviceDefaultsTreatNonPositiveValuesAsUnsetWallet struct{}

func (w serviceDefaultsTreatNonPositiveValuesAsUnsetWallet) Address(context.Context) (string, error) {
	return "0x1234567890abcdef1234567890abcdef12345678", nil
}

func (w serviceDefaultsTreatNonPositiveValuesAsUnsetWallet) Balance(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (w serviceDefaultsTreatNonPositiveValuesAsUnsetWallet) SignTransaction(context.Context, []byte) ([]byte, error) {
	return nil, errors.New("unexpected sign transaction")
}

func (w serviceDefaultsTreatNonPositiveValuesAsUnsetWallet) SignMessage(context.Context, []byte) ([]byte, error) {
	return nil, errors.New("unexpected sign message")
}

func (w serviceDefaultsTreatNonPositiveValuesAsUnsetWallet) PublicKey(context.Context) ([]byte, error) {
	return nil, nil
}

var _ payment.TxStore = (*serviceDefaultsTreatNonPositiveValuesAsUnsetTxStore)(nil)
