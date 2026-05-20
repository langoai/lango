package settlement

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type settlementRPCFake struct {
	mu sync.Mutex

	sendErrs []error
	sentRaw  []hexutil.Bytes
	onSend   func()

	receipts     []*types.Receipt
	receiptErrs  []error
	receiptCalls int
}

func (f *settlementRPCFake) SendRawTransaction(_ context.Context, raw hexutil.Bytes) (common.Hash, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.sentRaw = append(f.sentRaw, append(hexutil.Bytes(nil), raw...))
	call := len(f.sentRaw) - 1
	if f.onSend != nil {
		f.onSend()
	}
	if call < len(f.sendErrs) && f.sendErrs[call] != nil {
		return common.Hash{}, f.sendErrs[call]
	}
	return common.BytesToHash(raw), nil
}

func (f *settlementRPCFake) GetTransactionReceipt(_ context.Context, _ common.Hash) (*types.Receipt, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	call := f.receiptCalls
	f.receiptCalls++
	if call < len(f.receiptErrs) && f.receiptErrs[call] != nil {
		return nil, f.receiptErrs[call]
	}
	if call < len(f.receipts) {
		return f.receipts[call], nil
	}
	return nil, nil
}

func (f *settlementRPCFake) sendCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sentRaw)
}

func (f *settlementRPCFake) receiptCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.receiptCalls
}

func newSettlementRPCClient(t *testing.T, fake *settlementRPCFake) *ethclient.Client {
	t.Helper()

	server := rpc.NewServer()
	require.NoError(t, server.RegisterName("eth", fake))
	t.Cleanup(server.Stop)

	client := ethclient.NewClient(rpc.DialInProc(server))
	t.Cleanup(client.Close)
	return client
}

func newSettlementRPCTestService(t *testing.T, fake *settlementRPCFake, cfg Config) *Service {
	t.Helper()

	cfg.RPCClient = newSettlementRPCClient(t, fake)
	cfg.Logger = zap.NewNop().Sugar()
	return New(cfg)
}

func newSettlementRPCTestTx() *types.Transaction {
	return types.NewTx(&types.LegacyTx{
		Nonce:    7,
		To:       &common.Address{0x11},
		Value:    big.NewInt(0),
		Gas:      21_000,
		GasPrice: big.NewInt(1),
		Data:     []byte("settlement-test"),
	})
}

func TestSubmitWithRetryReturnsHashAfterTransientSendFailure(t *testing.T) {
	t.Parallel()

	fake := &settlementRPCFake{sendErrs: []error{errors.New("temporary submit failure")}}
	svc := newSettlementRPCTestService(t, fake, Config{
		MaxRetries: 2,
	})
	tx := newSettlementRPCTestTx()

	txHash, err := svc.submitWithRetry(context.Background(), tx)

	require.NoError(t, err)
	require.Equal(t, tx.Hash().Hex(), txHash)
	require.Equal(t, 2, fake.sendCount())
}

func TestSubmitWithRetryReturnsFinalSendErrorAfterRetries(t *testing.T) {
	t.Parallel()

	fake := &settlementRPCFake{sendErrs: []error{errors.New("still rejected")}}
	svc := newSettlementRPCTestService(t, fake, Config{
		MaxRetries: 1,
	})

	txHash, err := svc.submitWithRetry(context.Background(), newSettlementRPCTestTx())

	require.Error(t, err)
	require.Empty(t, txHash)
	require.Contains(t, err.Error(), "submit after 1 retries")
	require.Contains(t, err.Error(), "still rejected")
	require.Equal(t, 1, fake.sendCount())
}

func TestSubmitWithRetryReturnsContextErrorDuringBackoff(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fake := &settlementRPCFake{
		sendErrs: []error{errors.New("rejected")},
		onSend:   cancel,
	}
	svc := newSettlementRPCTestService(t, fake, Config{
		MaxRetries: 2,
	})

	txHash, err := svc.submitWithRetry(ctx, newSettlementRPCTestTx())

	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, txHash)
	require.Equal(t, 1, fake.sendCount())
}

func TestWaitForConfirmationReturnsNilForSuccessfulReceipt(t *testing.T) {
	t.Parallel()

	fake := &settlementRPCFake{
		receipts: []*types.Receipt{{
			Status:            types.ReceiptStatusSuccessful,
			CumulativeGasUsed: 21_000,
			Logs:              []*types.Log{},
			TxHash:            common.HexToHash("0x1234"),
			GasUsed:           21_000,
		}},
	}
	svc := newSettlementRPCTestService(t, fake, Config{
		ReceiptTimeout: time.Second,
	})

	err := svc.waitForConfirmation(context.Background(), common.HexToHash("0x1234"))

	require.NoError(t, err)
	require.Equal(t, 1, fake.receiptCallCount())
}

func TestWaitForConfirmationReturnsRevertErrorForFailedReceipt(t *testing.T) {
	t.Parallel()

	fake := &settlementRPCFake{
		receipts: []*types.Receipt{{
			Status:            types.ReceiptStatusFailed,
			CumulativeGasUsed: 21_000,
			Logs:              []*types.Log{},
			TxHash:            common.HexToHash("0x1234"),
			GasUsed:           21_000,
		}},
	}
	svc := newSettlementRPCTestService(t, fake, Config{
		ReceiptTimeout: time.Second,
	})

	err := svc.waitForConfirmation(context.Background(), common.HexToHash("0x1234"))

	require.Error(t, err)
	require.Contains(t, err.Error(), "tx reverted: status=0")
	require.Equal(t, 1, fake.receiptCallCount())
}

func TestWaitForConfirmationReturnsTimeoutWhenReceiptNeverAppears(t *testing.T) {
	t.Parallel()

	fake := &settlementRPCFake{}
	svc := newSettlementRPCTestService(t, fake, Config{
		ReceiptTimeout: time.Nanosecond,
	})

	err := svc.waitForConfirmation(context.Background(), common.HexToHash("0x1234"))

	require.Error(t, err)
	require.True(t, strings.HasPrefix(err.Error(), "receipt timeout after "), err.Error())
	require.Equal(t, 1, fake.receiptCallCount())
}

func TestWaitForConfirmationReturnsContextErrorWhenContextAlreadyCanceled(t *testing.T) {
	t.Parallel()

	fake := &settlementRPCFake{}
	svc := newSettlementRPCTestService(t, fake, Config{
		ReceiptTimeout: time.Second,
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := svc.waitForConfirmation(ctx, common.HexToHash("0x1234"))

	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 0, fake.receiptCallCount())
}
