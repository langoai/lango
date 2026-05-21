package settlement

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent/paymenttx"
	"github.com/langoai/lango/internal/eventbus"
)

type settlementRPCFake struct {
	mu sync.Mutex

	sendErrs []error
	sentRaw  []hexutil.Bytes
	onSend   func()

	receipts     []*types.Receipt
	receiptErrs  []error
	receiptCalls int

	nonce       hexutil.Uint64
	nonceErr    error
	estimateGas hexutil.Uint64
	gasErr      error
	header      *types.Header
	headerErr   error
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

func (f *settlementRPCFake) GetTransactionCount(_ context.Context, _ common.Address, _ string) (hexutil.Uint64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.nonceErr != nil {
		return 0, f.nonceErr
	}
	return f.nonce, nil
}

func (f *settlementRPCFake) EstimateGas(_ context.Context, _ map[string]interface{}) (hexutil.Uint64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.gasErr != nil {
		return 0, f.gasErr
	}
	if f.estimateGas == 0 {
		return hexutil.Uint64(75_000), nil
	}
	return f.estimateGas, nil
}

func (f *settlementRPCFake) GetBlockByNumber(_ context.Context, _ string, _ bool) (*types.Header, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.headerErr != nil {
		return nil, f.headerErr
	}
	if f.header != nil {
		return f.header, nil
	}
	return newSettlementRPCHeader(big.NewInt(1_000_000_000)), nil
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

func newSettlementRPCHeader(baseFee *big.Int) *types.Header {
	return &types.Header{
		Number:     big.NewInt(1),
		Difficulty: big.NewInt(0),
		GasLimit:   30_000_000,
		Time:       1,
		BaseFee:    baseFee,
	}
}

type settlementRPCSigningWallet struct {
	key        *ecdsa.PrivateKey
	addressErr error
	signErr    error
}

func newSettlementRPCSigningWallet(t *testing.T) settlementRPCSigningWallet {
	t.Helper()

	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	return settlementRPCSigningWallet{key: key}
}

func (w settlementRPCSigningWallet) Address(context.Context) (string, error) {
	if w.addressErr != nil {
		return "", w.addressErr
	}
	return crypto.PubkeyToAddress(w.key.PublicKey).Hex(), nil
}

func (w settlementRPCSigningWallet) Balance(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (w settlementRPCSigningWallet) SignTransaction(_ context.Context, digest []byte) ([]byte, error) {
	if w.signErr != nil {
		return nil, w.signErr
	}
	sig, err := crypto.Sign(digest, w.key)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (w settlementRPCSigningWallet) SignMessage(context.Context, []byte) ([]byte, error) {
	return nil, errors.New("unexpected sign message")
}

func (w settlementRPCSigningWallet) PublicKey(context.Context) ([]byte, error) {
	return crypto.FromECDSAPub(&w.key.PublicKey), nil
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

func TestBuildAndSignTxCoversRPCAndSignatureBranches(t *testing.T) {
	t.Parallel()

	t.Run("success builds signed dynamic fee transaction", func(t *testing.T) {
		t.Parallel()

		fake := &settlementRPCFake{
			nonce:       11,
			estimateGas: 91_000,
			header:      newSettlementRPCHeader(big.NewInt(2_000_000_000)),
		}
		wallet := newSettlementRPCSigningWallet(t)
		svc := newSettlementRPCTestService(t, fake, Config{
			Wallet:   wallet,
			ChainID:  84532,
			USDCAddr: common.HexToAddress("0x3333333333333333333333333333333333333333"),
		})

		tx, err := svc.buildAndSignTx(context.Background(), subscribeCloseWaitsForPublishedSettlementFailureAuthorization())

		require.NoError(t, err)
		require.Equal(t, uint64(11), tx.Nonce())
		require.Equal(t, uint64(91_000), tx.Gas())
		require.Equal(t, svc.usdcAddr, *tx.To())
		require.NotEmpty(t, tx.Data())
		require.Equal(t, big.NewInt(5_500_000_000), tx.GasFeeCap())
	})

	t.Run("nonce error is wrapped", func(t *testing.T) {
		t.Parallel()

		fake := &settlementRPCFake{nonceErr: errors.New("nonce unavailable")}
		wallet := newSettlementRPCSigningWallet(t)
		svc := newSettlementRPCTestService(t, fake, Config{Wallet: wallet, ChainID: 84532})

		tx, err := svc.buildAndSignTx(context.Background(), subscribeCloseWaitsForPublishedSettlementFailureAuthorization())

		require.Error(t, err)
		require.Nil(t, tx)
		require.Contains(t, err.Error(), "get nonce")
	})

	t.Run("estimate gas error is wrapped", func(t *testing.T) {
		t.Parallel()

		fake := &settlementRPCFake{gasErr: errors.New("gas unavailable")}
		wallet := newSettlementRPCSigningWallet(t)
		svc := newSettlementRPCTestService(t, fake, Config{Wallet: wallet, ChainID: 84532})

		tx, err := svc.buildAndSignTx(context.Background(), subscribeCloseWaitsForPublishedSettlementFailureAuthorization())

		require.Error(t, err)
		require.Nil(t, tx)
		require.Contains(t, err.Error(), "estimate gas")
	})

	t.Run("header error is wrapped", func(t *testing.T) {
		t.Parallel()

		fake := &settlementRPCFake{headerErr: errors.New("header unavailable")}
		wallet := newSettlementRPCSigningWallet(t)
		svc := newSettlementRPCTestService(t, fake, Config{Wallet: wallet, ChainID: 84532})

		tx, err := svc.buildAndSignTx(context.Background(), subscribeCloseWaitsForPublishedSettlementFailureAuthorization())

		require.Error(t, err)
		require.Nil(t, tx)
		require.Contains(t, err.Error(), "get block header")
	})

	t.Run("nil base fee uses fallback", func(t *testing.T) {
		t.Parallel()

		fake := &settlementRPCFake{header: newSettlementRPCHeader(nil)}
		wallet := newSettlementRPCSigningWallet(t)
		svc := newSettlementRPCTestService(t, fake, Config{Wallet: wallet, ChainID: 84532})

		tx, err := svc.buildAndSignTx(context.Background(), subscribeCloseWaitsForPublishedSettlementFailureAuthorization())

		require.NoError(t, err)
		require.Equal(t, big.NewInt(3_500_000_000), tx.GasFeeCap())
	})

	t.Run("sign error is wrapped", func(t *testing.T) {
		t.Parallel()

		fake := &settlementRPCFake{}
		wallet := newSettlementRPCSigningWallet(t)
		wallet.signErr = errors.New("signer unavailable")
		svc := newSettlementRPCTestService(t, fake, Config{Wallet: wallet, ChainID: 84532})

		tx, err := svc.buildAndSignTx(context.Background(), subscribeCloseWaitsForPublishedSettlementFailureAuthorization())

		require.Error(t, err)
		require.Nil(t, tx)
		require.Contains(t, err.Error(), "sign tx")
	})

}

func TestSettleRecordsSubmittedAndConfirmedOnSuccessfulLifecycle(t *testing.T) {
	t.Parallel()

	store := &subscribeCloseWaitsForPublishedSettlementFailureTxStore{}
	fake := &settlementRPCFake{
		receipts: []*types.Receipt{{
			Status:            types.ReceiptStatusSuccessful,
			CumulativeGasUsed: 75_000,
			Logs:              []*types.Log{},
			TxHash:            common.HexToHash("0x1234"),
			GasUsed:           75_000,
		}},
	}
	svc := newSettlementRPCTestService(t, fake, Config{
		Wallet:         newSettlementRPCSigningWallet(t),
		TxStore:        store,
		ChainID:        84532,
		USDCAddr:       common.HexToAddress("0x3333333333333333333333333333333333333333"),
		ReceiptTimeout: time.Second,
	})

	err := svc.settle(context.Background(), subscribeCloseWaitsForPublishedSettlementFailureAuthorization(), "did:peer:ok", "paid-tool")

	require.NoError(t, err)
	require.Len(t, store.created, 1)
	require.Len(t, store.updates, 2)
	require.Equal(t, paymenttx.StatusSubmitted, store.updates[0].status)
	require.NotEmpty(t, store.updates[0].txHash)
	require.Empty(t, store.updates[0].errMsg)
	require.Equal(t, paymenttx.StatusConfirmed, store.updates[1].status)
	require.Equal(t, store.updates[0].txHash, store.updates[1].txHash)
	require.Empty(t, store.updates[1].errMsg)
	require.Equal(t, 1, fake.sendCount())
	require.Equal(t, 1, fake.receiptCallCount())
}

func TestSettleMarksFailedForSubmitAndConfirmationFailures(t *testing.T) {
	t.Parallel()

	t.Run("submit failure records failed status without tx hash", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		store := &subscribeCloseWaitsForPublishedSettlementFailureTxStore{}
		fake := &settlementRPCFake{
			sendErrs: []error{errors.New("submit rejected")},
			onSend:   cancel,
		}
		svc := newSettlementRPCTestService(t, fake, Config{
			Wallet:   newSettlementRPCSigningWallet(t),
			TxStore:  store,
			ChainID:  84532,
			USDCAddr: common.HexToAddress("0x3333333333333333333333333333333333333333"),
		})

		err := svc.settle(ctx, subscribeCloseWaitsForPublishedSettlementFailureAuthorization(), "did:peer:submit", "paid-tool")

		require.Error(t, err)
		require.Contains(t, err.Error(), "submit tx")
		require.Len(t, store.updates, 1)
		require.Equal(t, paymenttx.StatusFailed, store.updates[0].status)
		require.Empty(t, store.updates[0].txHash)
		require.Contains(t, store.updates[0].errMsg, "context canceled")
		require.Equal(t, 1, fake.sendCount())
	})

	t.Run("failed receipt records failed status with tx hash", func(t *testing.T) {
		t.Parallel()

		store := &subscribeCloseWaitsForPublishedSettlementFailureTxStore{}
		fake := &settlementRPCFake{
			receipts: []*types.Receipt{{
				Status:            types.ReceiptStatusFailed,
				CumulativeGasUsed: 75_000,
				Logs:              []*types.Log{},
				GasUsed:           75_000,
			}},
		}
		svc := newSettlementRPCTestService(t, fake, Config{
			Wallet:         newSettlementRPCSigningWallet(t),
			TxStore:        store,
			ChainID:        84532,
			USDCAddr:       common.HexToAddress("0x3333333333333333333333333333333333333333"),
			ReceiptTimeout: time.Second,
		})

		err := svc.settle(context.Background(), subscribeCloseWaitsForPublishedSettlementFailureAuthorization(), "did:peer:confirm", "paid-tool")

		require.Error(t, err)
		require.Contains(t, err.Error(), "wait confirmation")
		require.Len(t, store.updates, 2)
		require.Equal(t, paymenttx.StatusSubmitted, store.updates[0].status)
		require.Equal(t, paymenttx.StatusFailed, store.updates[1].status)
		require.Equal(t, store.updates[0].txHash, store.updates[1].txHash)
		require.Contains(t, store.updates[1].errMsg, "tx reverted")
		require.Equal(t, 1, fake.sendCount())
		require.Equal(t, 1, fake.receiptCallCount())
	})
}

func TestHandleEventSuccessRecordsReputation(t *testing.T) {
	t.Parallel()

	rec := &settleLocalStoreRecordsDeterministicPaymentBeforeBuildFailureReputationRecorder{}
	store := &subscribeCloseWaitsForPublishedSettlementFailureTxStore{}
	fake := &settlementRPCFake{
		receipts: []*types.Receipt{{
			Status:            types.ReceiptStatusSuccessful,
			CumulativeGasUsed: 75_000,
			Logs:              []*types.Log{},
			GasUsed:           75_000,
		}},
	}
	svc := newSettlementRPCTestService(t, fake, Config{
		Wallet:         newSettlementRPCSigningWallet(t),
		TxStore:        store,
		ChainID:        84532,
		USDCAddr:       common.HexToAddress("0x3333333333333333333333333333333333333333"),
		ReceiptTimeout: time.Second,
	})
	svc.SetReputationRecorder(rec)

	svc.handleEvent(eventbus.ToolExecutionPaidEvent{
		PeerDID:  "did:peer:success",
		ToolName: "paid-tool",
		Auth:     subscribeCloseWaitsForPublishedSettlementFailureAuthorization(),
	})

	require.Equal(t, []string{"did:peer:success"}, rec.successes)
	require.Empty(t, rec.failures)
	require.Len(t, store.updates, 2)
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
