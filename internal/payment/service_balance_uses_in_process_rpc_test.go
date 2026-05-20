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

func TestServiceBalanceUsesInProcessRPC(t *testing.T) {
	t.Parallel()

	walletAddr := "0x1111111111111111111111111111111111111111"
	usdcAddr := "0x2222222222222222222222222222222222222222"
	api := &serviceBalanceUsesInProcessRpcEthAPI{balance: big.NewInt(123_456_789)}
	client := serviceBalanceUsesInProcessRpcEthClient(t, api)
	svc := &Service{
		wallet:    &serviceBalanceUsesInProcessRpcWallet{address: walletAddr},
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

func TestServiceBalanceWrapsCallContractError(t *testing.T) {
	t.Parallel()

	callErr := errors.New("rpc unavailable")
	client := serviceBalanceUsesInProcessRpcEthClient(t, &serviceBalanceUsesInProcessRpcEthAPI{callErr: callErr})
	svc := &Service{
		wallet:    &serviceBalanceUsesInProcessRpcWallet{address: "0x1111111111111111111111111111111111111111"},
		builder:   NewTxBuilder(client, 8453, "0x2222222222222222222222222222222222222222"),
		rpcClient: client,
	}

	balance, err := svc.Balance(context.Background())

	require.Error(t, err)
	assert.Empty(t, balance)
	assert.Contains(t, err.Error(), "query USDC balance")
	assert.Contains(t, err.Error(), callErr.Error())
}

func TestServiceSendCompletesWithInProcessRPC(t *testing.T) {
	t.Parallel()

	wallet := serviceBalanceUsesInProcessRpcKeyWallet(t)
	store := newServiceBalanceUsesInProcessRpcTxStore()
	limiter := &serviceBalanceUsesInProcessRpcLimiter{}
	api := &serviceBalanceUsesInProcessRpcEthAPI{
		receipt: &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			GasUsed:     44_000,
			BlockNumber: big.NewInt(123),
		},
	}
	client := serviceBalanceUsesInProcessRpcEthClient(t, api)
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
		Purpose:    "runChatStartErrorStopsBeforeSessionAndTui6 success",
		SessionKey: "session-runChatStartErrorStopsBeforeSessionAndTui6",
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
		assert.Equal(t, "runChatStartErrorStopsBeforeSessionAndTui6 success", record.Purpose)
		assert.Equal(t, "session-runChatStartErrorStopsBeforeSessionAndTui6", record.SessionKey)
		assert.Equal(t, "https://example.test/paid", record.X402URL)
	}
}

func TestServiceSendRecordsSignFailure(t *testing.T) {
	t.Parallel()

	signErr := errors.New("sign denied")
	wallet := serviceBalanceUsesInProcessRpcKeyWallet(t)
	wallet.signErr = signErr
	store := newServiceBalanceUsesInProcessRpcTxStore()
	client := serviceBalanceUsesInProcessRpcEthClient(t, &serviceBalanceUsesInProcessRpcEthAPI{})
	svc := &Service{
		wallet:         wallet,
		limiter:        &serviceBalanceUsesInProcessRpcLimiter{},
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

func TestServiceSendRecordsFailedConfirmationBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		api            *serviceBalanceUsesInProcessRpcEthAPI
		receiptTimeout time.Duration
		wantErr        string
		wantRecordErr  string
	}{
		{
			name: "reverted receipt",
			api: &serviceBalanceUsesInProcessRpcEthAPI{
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
			api: &serviceBalanceUsesInProcessRpcEthAPI{
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

			store := newServiceBalanceUsesInProcessRpcTxStore()
			client := serviceBalanceUsesInProcessRpcEthClient(t, tt.api)
			svc := &Service{
				wallet:         serviceBalanceUsesInProcessRpcKeyWallet(t),
				limiter:        &serviceBalanceUsesInProcessRpcLimiter{},
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

func TestSubmitWithRetryReturnsContextErrorWithoutSleep(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	api := &serviceBalanceUsesInProcessRpcEthAPI{
		sendErrs: []error{errors.New("temporary rpc failure")},
		onSend: func(int) {
			cancel()
		},
	}
	svc := &Service{
		rpcClient:  serviceBalanceUsesInProcessRpcEthClient(t, api),
		maxRetries: 3,
	}

	txHash, err := svc.submitWithRetry(ctx, serviceBalanceUsesInProcessRpcSignedTx(t))

	require.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, txHash)
	assert.Equal(t, 1, api.sendCount())
}

func TestWaitForConfirmationBranchesWithoutLongPoll(t *testing.T) {
	t.Parallel()

	txHash := common.HexToHash("0x1234")

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		api := &serviceBalanceUsesInProcessRpcEthAPI{
			receipt: &types.Receipt{
				Status:      types.ReceiptStatusSuccessful,
				GasUsed:     55_000,
				BlockNumber: big.NewInt(77),
			},
		}
		svc := &Service{
			rpcClient:      serviceBalanceUsesInProcessRpcEthClient(t, api),
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
		api := &serviceBalanceUsesInProcessRpcEthAPI{receiptErr: ethereum.NotFound}
		svc := &Service{
			rpcClient:      serviceBalanceUsesInProcessRpcEthClient(t, api),
			receiptTimeout: time.Second,
		}

		receipt, err := svc.waitForConfirmation(ctx, txHash)

		require.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Nil(t, receipt)
		assert.Equal(t, 0, api.receiptCount())
	})
}

type serviceBalanceUsesInProcessRpcWallet struct {
	address string
	key     *ecdsa.PrivateKey
	signErr error

	mu    sync.Mutex
	signs int
}

func (w *serviceBalanceUsesInProcessRpcWallet) Address(context.Context) (string, error) {
	return w.address, nil
}

func (w *serviceBalanceUsesInProcessRpcWallet) Balance(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (w *serviceBalanceUsesInProcessRpcWallet) SignTransaction(_ context.Context, rawTx []byte) ([]byte, error) {
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

func (w *serviceBalanceUsesInProcessRpcWallet) SignMessage(context.Context, []byte) ([]byte, error) {
	return nil, nil
}

func (w *serviceBalanceUsesInProcessRpcWallet) PublicKey(context.Context) ([]byte, error) {
	return nil, nil
}

func (w *serviceBalanceUsesInProcessRpcWallet) signCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.signs
}

type serviceBalanceUsesInProcessRpcLimiter struct {
	mu      sync.Mutex
	records int
}

func (l *serviceBalanceUsesInProcessRpcLimiter) Check(context.Context, *big.Int) error {
	return nil
}

func (l *serviceBalanceUsesInProcessRpcLimiter) Record(context.Context, *big.Int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.records++
	return nil
}

func (l *serviceBalanceUsesInProcessRpcLimiter) DailySpent(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (l *serviceBalanceUsesInProcessRpcLimiter) DailyRemaining(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (l *serviceBalanceUsesInProcessRpcLimiter) IsAutoApprovable(context.Context, *big.Int) (bool, error) {
	return false, nil
}

func (l *serviceBalanceUsesInProcessRpcLimiter) recordCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.records
}

type serviceBalanceUsesInProcessRpcTxStore struct {
	mu      sync.Mutex
	records map[uuid.UUID]TxRecord
	updates []paymenttx.Status
}

func newServiceBalanceUsesInProcessRpcTxStore() *serviceBalanceUsesInProcessRpcTxStore {
	return &serviceBalanceUsesInProcessRpcTxStore{records: make(map[uuid.UUID]TxRecord)}
}

func (s *serviceBalanceUsesInProcessRpcTxStore) Create(_ context.Context, record TxRecord) (TxRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}
	record.CreatedAt = time.Now()
	s.records[record.ID] = record
	return record, nil
}

func (s *serviceBalanceUsesInProcessRpcTxStore) UpdateStatus(_ context.Context, id uuid.UUID, status paymenttx.Status, txHash, errMsg string) error {
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

func (s *serviceBalanceUsesInProcessRpcTxStore) List(context.Context, int) ([]TxRecord, error) {
	return nil, nil
}

func (s *serviceBalanceUsesInProcessRpcTxStore) DailySpendSince(context.Context, time.Time) ([]string, error) {
	return nil, nil
}

func (s *serviceBalanceUsesInProcessRpcTxStore) statuses() []paymenttx.Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]paymenttx.Status(nil), s.updates...)
}

func (s *serviceBalanceUsesInProcessRpcTxStore) recordsSnapshot() map[uuid.UUID]TxRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make(map[uuid.UUID]TxRecord, len(s.records))
	for id, record := range s.records {
		out[id] = record
	}
	return out
}

type serviceBalanceUsesInProcessRpcEthAPI struct {
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

func (api *serviceBalanceUsesInProcessRpcEthAPI) Call(_ context.Context, msg map[string]interface{}, _ string) (hexutil.Bytes, error) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.calls++
	if api.callErr != nil {
		return nil, api.callErr
	}

	to, err := serviceBalanceUsesInProcessRpcRPCAddress(msg["to"])
	if err != nil {
		return nil, err
	}
	data, err := serviceBalanceUsesInProcessRpcRPCBytes(msg["data"])
	if err != nil && msg["data"] == nil {
		data, err = serviceBalanceUsesInProcessRpcRPCBytes(msg["input"])
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

func (api *serviceBalanceUsesInProcessRpcEthAPI) GetTransactionCount(context.Context, common.Address, string) (hexutil.Uint64, error) {
	api.mu.Lock()
	defer api.mu.Unlock()
	return hexutil.Uint64(api.nonce), nil
}

func (api *serviceBalanceUsesInProcessRpcEthAPI) EstimateGas(context.Context, map[string]interface{}) (hexutil.Uint64, error) {
	api.mu.Lock()
	defer api.mu.Unlock()
	if api.gas == 0 {
		return hexutil.Uint64(50_000), nil
	}
	return hexutil.Uint64(api.gas), nil
}

func (api *serviceBalanceUsesInProcessRpcEthAPI) GetBlockByNumber(context.Context, string, bool) (*types.Header, error) {
	return &types.Header{
		Number:     big.NewInt(1),
		Difficulty: big.NewInt(0),
		GasLimit:   30_000_000,
		BaseFee:    big.NewInt(DefaultBaseFeeWei),
	}, nil
}

func (api *serviceBalanceUsesInProcessRpcEthAPI) SendRawTransaction(context.Context, hexutil.Bytes) (common.Hash, error) {
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

func (api *serviceBalanceUsesInProcessRpcEthAPI) GetTransactionReceipt(_ context.Context, hash common.Hash) (map[string]interface{}, error) {
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

func (api *serviceBalanceUsesInProcessRpcEthAPI) callCount() int {
	api.mu.Lock()
	defer api.mu.Unlock()
	return api.calls
}

func (api *serviceBalanceUsesInProcessRpcEthAPI) lastCallToValue() common.Address {
	api.mu.Lock()
	defer api.mu.Unlock()
	return api.lastCallTo
}

func (api *serviceBalanceUsesInProcessRpcEthAPI) lastCallDataValue() []byte {
	api.mu.Lock()
	defer api.mu.Unlock()
	return append([]byte(nil), api.lastCallData...)
}

func (api *serviceBalanceUsesInProcessRpcEthAPI) sendCount() int {
	api.mu.Lock()
	defer api.mu.Unlock()
	return api.sends
}

func (api *serviceBalanceUsesInProcessRpcEthAPI) receiptCount() int {
	api.mu.Lock()
	defer api.mu.Unlock()
	return api.receipts
}

func serviceBalanceUsesInProcessRpcEthClient(t *testing.T, api *serviceBalanceUsesInProcessRpcEthAPI) *ethclient.Client {
	t.Helper()

	server := rpc.NewServer()
	require.NoError(t, server.RegisterName("eth", api))
	t.Cleanup(server.Stop)

	client := rpc.DialInProc(server)
	t.Cleanup(client.Close)

	return ethclient.NewClient(client)
}

func serviceBalanceUsesInProcessRpcKeyWallet(t *testing.T) *serviceBalanceUsesInProcessRpcWallet {
	t.Helper()

	key, err := crypto.HexToECDSA("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)
	return &serviceBalanceUsesInProcessRpcWallet{
		address: crypto.PubkeyToAddress(key.PublicKey).Hex(),
		key:     key,
	}
}

func serviceBalanceUsesInProcessRpcSignedTx(t *testing.T) *types.Transaction {
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

func serviceBalanceUsesInProcessRpcRPCAddress(value interface{}) (common.Address, error) {
	switch v := value.(type) {
	case common.Address:
		return v, nil
	case string:
		return common.HexToAddress(v), nil
	default:
		return common.Address{}, fmt.Errorf("unexpected RPC address type %T", value)
	}
}

func serviceBalanceUsesInProcessRpcRPCBytes(value interface{}) ([]byte, error) {
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
