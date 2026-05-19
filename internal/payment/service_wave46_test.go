package payment

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ent/paymenttx"
)

func TestWave46ServiceBalanceUsesInProcessRPC(t *testing.T) {
	t.Parallel()

	walletAddr := "0x1111111111111111111111111111111111111111"
	usdcAddr := "0x2222222222222222222222222222222222222222"
	api := &wave46EthAPI{balance: big.NewInt(123_456_789)}
	client := wave46EthClient(t, api)
	svc := &Service{
		wallet:    &wave46Wallet{address: walletAddr},
		builder:   NewTxBuilder(client, 8453, usdcAddr),
		rpcClient: client,
	}

	balance, err := svc.Balance(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "123.456789", balance)
	assert.Equal(t, 1, api.callCount())
	assert.Equal(t, common.HexToAddress(usdcAddr), api.lastCallToValue())
	data := api.lastCallDataValue()
	require.Len(t, data, 4+32)
	assert.Equal(t, BalanceOfSelector, data[:4])
	assert.Equal(t, common.HexToAddress(walletAddr), common.BytesToAddress(data[4+12:4+32]))
}

func TestWave46ServiceBalanceWrapsCallContractError(t *testing.T) {
	t.Parallel()

	callErr := errors.New("rpc unavailable")
	client := wave46EthClient(t, &wave46EthAPI{callErr: callErr})
	svc := &Service{
		wallet:    &wave46Wallet{address: "0x1111111111111111111111111111111111111111"},
		builder:   NewTxBuilder(client, 8453, "0x2222222222222222222222222222222222222222"),
		rpcClient: client,
	}

	balance, err := svc.Balance(context.Background())

	require.Error(t, err)
	assert.Empty(t, balance)
	assert.Contains(t, err.Error(), "query USDC balance")
	assert.Contains(t, err.Error(), callErr.Error())
}

func TestWave46ServiceSendCompletesWithInProcessRPC(t *testing.T) {
	t.Parallel()

	wallet := wave46KeyWallet(t)
	store := newWave46TxStore()
	limiter := &wave46Limiter{}
	api := &wave46EthAPI{
		receipt: &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			GasUsed:     44_000,
			BlockNumber: big.NewInt(123),
		},
	}
	client := wave46EthClient(t, api)
	svc := &Service{
		wallet:         wallet,
		limiter:        limiter,
		builder:        NewTxBuilder(client, 8453, "0x2222222222222222222222222222222222222222"),
		store:          store,
		rpcClient:      client,
		chainID:        8453,
		receiptTimeout: time.Second,
		maxRetries:     1,
	}

	receipt, err := svc.Send(context.Background(), PaymentRequest{
		To:         "0x3333333333333333333333333333333333333333",
		Amount:     "1.25",
		Purpose:    "wave46 success",
		SessionKey: "session-wave46",
		X402URL:    "https://example.test/paid",
	})

	require.NoError(t, err)
	require.NotNil(t, receipt)
	assert.Equal(t, string(paymenttx.StatusConfirmed), receipt.Status)
	assert.NotEmpty(t, receipt.TxHash)
	assert.Equal(t, "1.25", receipt.Amount)
	assert.Equal(t, wallet.address, receipt.From)
	assert.Equal(t, "0x3333333333333333333333333333333333333333", receipt.To)
	assert.Equal(t, int64(8453), receipt.ChainID)
	assert.Equal(t, uint64(44_000), receipt.GasUsed)
	assert.Equal(t, uint64(123), receipt.BlockNumber)
	assert.Equal(t, 1, api.sendCount())
	assert.Equal(t, 1, api.receiptCount())
	assert.Equal(t, 1, limiter.recordCount())
	assert.Equal(t, 1, wallet.signCount())
	assert.Equal(t, []paymenttx.Status{paymenttx.StatusSubmitted, paymenttx.StatusConfirmed}, store.statuses())

	records := store.recordsSnapshot()
	require.Len(t, records, 1)
	for _, record := range records {
		assert.Equal(t, paymenttx.StatusConfirmed, record.Status)
		assert.Equal(t, receipt.TxHash, record.TxHash)
		assert.Empty(t, record.ErrorMessage)
		assert.Equal(t, "wave46 success", record.Purpose)
		assert.Equal(t, "session-wave46", record.SessionKey)
		assert.Equal(t, "https://example.test/paid", record.X402URL)
	}
}

func TestWave46ServiceSendRecordsSignFailure(t *testing.T) {
	t.Parallel()

	signErr := errors.New("sign denied")
	wallet := wave46KeyWallet(t)
	wallet.signErr = signErr
	store := newWave46TxStore()
	client := wave46EthClient(t, &wave46EthAPI{})
	svc := &Service{
		wallet:         wallet,
		limiter:        &wave46Limiter{},
		builder:        NewTxBuilder(client, 8453, "0x2222222222222222222222222222222222222222"),
		store:          store,
		rpcClient:      client,
		chainID:        8453,
		receiptTimeout: time.Second,
		maxRetries:     1,
	}

	receipt, err := svc.Send(context.Background(), PaymentRequest{
		To:     "0x3333333333333333333333333333333333333333",
		Amount: "0.50",
	})

	require.Error(t, err)
	assert.Nil(t, receipt)
	assert.Contains(t, err.Error(), "sign transaction")
	assert.ErrorIs(t, err, signErr)
	assert.Equal(t, []paymenttx.Status{paymenttx.StatusFailed}, store.statuses())
	records := store.recordsSnapshot()
	require.Len(t, records, 1)
	for _, record := range records {
		assert.Equal(t, paymenttx.StatusFailed, record.Status)
		assert.Equal(t, "sign denied", record.ErrorMessage)
	}
}

func TestWave46ServiceSendRecordsFailedConfirmationBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		api            *wave46EthAPI
		receiptTimeout time.Duration
		wantErr        string
		wantRecordErr  string
	}{
		{
			name: "reverted receipt",
			api: &wave46EthAPI{
				receipt: &types.Receipt{
					Status:      types.ReceiptStatusFailed,
					GasUsed:     21_000,
					BlockNumber: big.NewInt(9),
				},
			},
			receiptTimeout: time.Second,
			wantErr:        "reverted",
			wantRecordErr:  "reverted",
		},
		{
			name: "receipt timeout",
			api: &wave46EthAPI{
				receiptErr: ethereum.NotFound,
			},
			receiptTimeout: time.Nanosecond,
			wantErr:        "confirm transaction",
			wantRecordErr:  "receipt timeout",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := newWave46TxStore()
			client := wave46EthClient(t, tt.api)
			svc := &Service{
				wallet:         wave46KeyWallet(t),
				limiter:        &wave46Limiter{},
				builder:        NewTxBuilder(client, 8453, "0x2222222222222222222222222222222222222222"),
				store:          store,
				rpcClient:      client,
				chainID:        8453,
				receiptTimeout: tt.receiptTimeout,
				maxRetries:     1,
			}

			receipt, err := svc.Send(context.Background(), PaymentRequest{
				To:     "0x3333333333333333333333333333333333333333",
				Amount: "0.75",
			})

			require.Error(t, err)
			assert.Nil(t, receipt)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Equal(t, 1, tt.api.sendCount())
			assert.Equal(t, 1, tt.api.receiptCount())
			assert.Equal(t, []paymenttx.Status{paymenttx.StatusSubmitted, paymenttx.StatusFailed}, store.statuses())

			records := store.recordsSnapshot()
			require.Len(t, records, 1)
			for _, record := range records {
				assert.Equal(t, paymenttx.StatusFailed, record.Status)
				assert.Contains(t, record.ErrorMessage, tt.wantRecordErr)
				assert.NotEmpty(t, record.TxHash)
			}
		})
	}
}

func TestWave46SubmitWithRetryReturnsContextErrorWithoutSleep(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	api := &wave46EthAPI{
		sendErrs: []error{errors.New("temporary rpc failure")},
		onSend: func(int) {
			cancel()
		},
	}
	svc := &Service{
		rpcClient:  wave46EthClient(t, api),
		maxRetries: 3,
	}

	txHash, err := svc.submitWithRetry(ctx, wave46SignedTx(t))

	require.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, txHash)
	assert.Equal(t, 1, api.sendCount())
}

func TestWave46WaitForConfirmationBranchesWithoutLongPoll(t *testing.T) {
	t.Parallel()

	txHash := common.HexToHash("0x1234")

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		api := &wave46EthAPI{
			receipt: &types.Receipt{
				Status:      types.ReceiptStatusSuccessful,
				GasUsed:     55_000,
				BlockNumber: big.NewInt(77),
			},
		}
		svc := &Service{
			rpcClient:      wave46EthClient(t, api),
			receiptTimeout: time.Second,
		}

		receipt, err := svc.waitForConfirmation(context.Background(), txHash)

		require.NoError(t, err)
		require.NotNil(t, receipt)
		assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
		assert.Equal(t, uint64(55_000), receipt.GasUsed)
		assert.Equal(t, uint64(77), receipt.BlockNumber.Uint64())
		assert.Equal(t, 1, api.receiptCount())
	})

	t.Run("context deadline", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Nanosecond))
		defer cancel()
		api := &wave46EthAPI{receiptErr: ethereum.NotFound}
		svc := &Service{
			rpcClient:      wave46EthClient(t, api),
			receiptTimeout: time.Second,
		}

		receipt, err := svc.waitForConfirmation(ctx, txHash)

		require.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Nil(t, receipt)
		assert.Equal(t, 0, api.receiptCount())
	})
}

type wave46Wallet struct {
	address string
	key     *ecdsa.PrivateKey
	signErr error

	mu    sync.Mutex
	signs int
}

func (w *wave46Wallet) Address(context.Context) (string, error) {
	return w.address, nil
}

func (w *wave46Wallet) Balance(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (w *wave46Wallet) SignTransaction(_ context.Context, rawTx []byte) ([]byte, error) {
	w.mu.Lock()
	w.signs++
	w.mu.Unlock()

	if w.signErr != nil {
		return nil, w.signErr
	}
	if w.key == nil {
		return nil, errors.New("missing signing key")
	}
	return crypto.Sign(rawTx, w.key)
}

func (w *wave46Wallet) SignMessage(context.Context, []byte) ([]byte, error) {
	return nil, nil
}

func (w *wave46Wallet) PublicKey(context.Context) ([]byte, error) {
	return nil, nil
}

func (w *wave46Wallet) signCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.signs
}

type wave46Limiter struct {
	mu      sync.Mutex
	records int
}

func (l *wave46Limiter) Check(context.Context, *big.Int) error {
	return nil
}

func (l *wave46Limiter) Record(context.Context, *big.Int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.records++
	return nil
}

func (l *wave46Limiter) DailySpent(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (l *wave46Limiter) DailyRemaining(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (l *wave46Limiter) IsAutoApprovable(context.Context, *big.Int) (bool, error) {
	return false, nil
}

func (l *wave46Limiter) recordCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.records
}

type wave46TxStore struct {
	mu      sync.Mutex
	records map[uuid.UUID]TxRecord
	updates []paymenttx.Status
}

func newWave46TxStore() *wave46TxStore {
	return &wave46TxStore{records: make(map[uuid.UUID]TxRecord)}
}

func (s *wave46TxStore) Create(_ context.Context, record TxRecord) (TxRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}
	record.CreatedAt = time.Now()
	s.records[record.ID] = record
	return record, nil
}

func (s *wave46TxStore) UpdateStatus(_ context.Context, id uuid.UUID, status paymenttx.Status, txHash, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[id]
	if !ok {
		return fmt.Errorf("missing tx record %s", id)
	}
	record.Status = status
	if txHash != "" {
		record.TxHash = txHash
	}
	if errMsg != "" {
		record.ErrorMessage = errMsg
	}
	s.records[id] = record
	s.updates = append(s.updates, status)
	return nil
}

func (s *wave46TxStore) List(context.Context, int) ([]TxRecord, error) {
	return nil, nil
}

func (s *wave46TxStore) DailySpendSince(context.Context, time.Time) ([]string, error) {
	return nil, nil
}

func (s *wave46TxStore) statuses() []paymenttx.Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]paymenttx.Status(nil), s.updates...)
}

func (s *wave46TxStore) recordsSnapshot() map[uuid.UUID]TxRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make(map[uuid.UUID]TxRecord, len(s.records))
	for id, record := range s.records {
		out[id] = record
	}
	return out
}

type wave46EthAPI struct {
	mu sync.Mutex

	balance      *big.Int
	callErr      error
	calls        int
	lastCallTo   common.Address
	lastCallData []byte

	nonce uint64
	gas   uint64

	sendErrs []error
	onSend   func(int)
	sends    int

	receipt    *types.Receipt
	receiptErr error
	receipts   int
}

func (api *wave46EthAPI) Call(_ context.Context, msg map[string]interface{}, _ string) (hexutil.Bytes, error) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.calls++
	if api.callErr != nil {
		return nil, api.callErr
	}

	to, err := wave46RPCAddress(msg["to"])
	if err != nil {
		return nil, err
	}
	data, err := wave46RPCBytes(msg["data"])
	if err != nil && msg["data"] == nil {
		data, err = wave46RPCBytes(msg["input"])
	}
	if err != nil {
		return nil, err
	}
	api.lastCallTo = to
	api.lastCallData = append([]byte(nil), data...)

	out := make([]byte, 32)
	if api.balance != nil {
		api.balance.FillBytes(out)
	}
	return out, nil
}

func (api *wave46EthAPI) GetTransactionCount(context.Context, common.Address, string) (hexutil.Uint64, error) {
	api.mu.Lock()
	defer api.mu.Unlock()
	return hexutil.Uint64(api.nonce), nil
}

func (api *wave46EthAPI) EstimateGas(context.Context, map[string]interface{}) (hexutil.Uint64, error) {
	api.mu.Lock()
	defer api.mu.Unlock()
	if api.gas == 0 {
		return hexutil.Uint64(50_000), nil
	}
	return hexutil.Uint64(api.gas), nil
}

func (api *wave46EthAPI) GetBlockByNumber(context.Context, string, bool) (*types.Header, error) {
	return &types.Header{
		Number:     big.NewInt(1),
		Difficulty: big.NewInt(0),
		GasLimit:   30_000_000,
		BaseFee:    big.NewInt(DefaultBaseFeeWei),
	}, nil
}

func (api *wave46EthAPI) SendRawTransaction(context.Context, hexutil.Bytes) (common.Hash, error) {
	api.mu.Lock()
	api.sends++
	sendNumber := api.sends
	var err error
	if len(api.sendErrs) > 0 {
		err = api.sendErrs[0]
		api.sendErrs = api.sendErrs[1:]
	}
	onSend := api.onSend
	api.mu.Unlock()

	if onSend != nil {
		onSend(sendNumber)
	}
	if err != nil {
		return common.Hash{}, err
	}
	return common.HexToHash("0xbeef"), nil
}

func (api *wave46EthAPI) GetTransactionReceipt(_ context.Context, hash common.Hash) (map[string]interface{}, error) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.receipts++
	if api.receiptErr != nil {
		return nil, api.receiptErr
	}
	if api.receipt == nil {
		return nil, nil
	}

	blockNumber := api.receipt.BlockNumber
	if blockNumber == nil {
		blockNumber = big.NewInt(1)
	}
	return map[string]interface{}{
		"transactionHash":   hash,
		"blockHash":         common.HexToHash("0xbeef"),
		"blockNumber":       (*hexutil.Big)(blockNumber),
		"transactionIndex":  hexutil.Uint64(0),
		"from":              common.HexToAddress("0x1111111111111111111111111111111111111111"),
		"to":                common.HexToAddress("0x2222222222222222222222222222222222222222"),
		"gasUsed":           hexutil.Uint64(api.receipt.GasUsed),
		"cumulativeGasUsed": hexutil.Uint64(api.receipt.GasUsed),
		"contractAddress":   nil,
		"logs":              []*types.Log{},
		"logsBloom":         types.Bloom{},
		"status":            hexutil.Uint64(api.receipt.Status),
		"type":              hexutil.Uint64(2),
	}, nil
}

func (api *wave46EthAPI) callCount() int {
	api.mu.Lock()
	defer api.mu.Unlock()
	return api.calls
}

func (api *wave46EthAPI) lastCallToValue() common.Address {
	api.mu.Lock()
	defer api.mu.Unlock()
	return api.lastCallTo
}

func (api *wave46EthAPI) lastCallDataValue() []byte {
	api.mu.Lock()
	defer api.mu.Unlock()
	return append([]byte(nil), api.lastCallData...)
}

func (api *wave46EthAPI) sendCount() int {
	api.mu.Lock()
	defer api.mu.Unlock()
	return api.sends
}

func (api *wave46EthAPI) receiptCount() int {
	api.mu.Lock()
	defer api.mu.Unlock()
	return api.receipts
}

func wave46EthClient(t *testing.T, api *wave46EthAPI) *ethclient.Client {
	t.Helper()

	server := rpc.NewServer()
	require.NoError(t, server.RegisterName("eth", api))
	t.Cleanup(server.Stop)

	client := rpc.DialInProc(server)
	t.Cleanup(client.Close)

	return ethclient.NewClient(client)
}

func wave46KeyWallet(t *testing.T) *wave46Wallet {
	t.Helper()

	key, err := crypto.HexToECDSA("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)
	return &wave46Wallet{
		address: crypto.PubkeyToAddress(key.PublicKey).Hex(),
		key:     key,
	}
}

func wave46SignedTx(t *testing.T) *types.Transaction {
	t.Helper()

	key, err := crypto.HexToECDSA("4c0883a69102937d6231471b5dbb6204fe5129617082791686b7f9cd8eadf3b4")
	require.NoError(t, err)
	to := common.HexToAddress("0x3333333333333333333333333333333333333333")
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(8453),
		Nonce:     1,
		GasTipCap: big.NewInt(1),
		GasFeeCap: big.NewInt(2),
		Gas:       21_000,
		To:        &to,
		Value:     big.NewInt(0),
	})
	signer := types.LatestSignerForChainID(big.NewInt(8453))
	sig, err := crypto.Sign(signer.Hash(tx).Bytes(), key)
	require.NoError(t, err)
	signed, err := tx.WithSignature(signer, sig)
	require.NoError(t, err)
	return signed
}

func wave46RPCAddress(value interface{}) (common.Address, error) {
	switch v := value.(type) {
	case common.Address:
		return v, nil
	case string:
		return common.HexToAddress(v), nil
	default:
		return common.Address{}, fmt.Errorf("unexpected RPC address type %T", value)
	}
}

func wave46RPCBytes(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case hexutil.Bytes:
		return append([]byte(nil), v...), nil
	case []byte:
		return append([]byte(nil), v...), nil
	case string:
		return common.FromHex(v), nil
	default:
		return nil, fmt.Errorf("unexpected RPC bytes type %T", value)
	}
}
