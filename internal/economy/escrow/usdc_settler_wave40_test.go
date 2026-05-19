package escrow

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/payment"
)

func TestWave40USDCSettlerResolveAddressUsesFallbackAndResolverOption(t *testing.T) {
	t.Parallel()

	key, err := crypto.HexToECDSA("4c0883a69102937d6231471b5dbb6204fe5129617082791686b7f9cd8eadf3b4")
	require.NoError(t, err)
	compressed := crypto.CompressPubkey(&key.PublicKey)
	did := "did:lango:" + common.Bytes2Hex(compressed)
	wantFallback := crypto.PubkeyToAddress(key.PublicKey)

	settler := NewUSDCSettler(nil, nil, nil, 8453)
	gotFallback, err := settler.resolveAddress(did)
	require.NoError(t, err)
	assert.Equal(t, wantFallback, gotFallback)

	wantResolved := common.HexToAddress("0x2222222222222222222222222222222222222222")
	resolver := &wave40Resolver{addresses: map[string]common.Address{
		"did:lango:v2:agent": wantResolved,
	}}
	settler = NewUSDCSettler(nil, nil, nil, 8453, WithAddressResolver(resolver))
	gotResolved, err := settler.resolveAddress("did:lango:v2:agent")
	require.NoError(t, err)
	assert.Equal(t, wantResolved, gotResolved)
	assert.Equal(t, []string{"did:lango:v2:agent"}, resolver.calls)
}

func TestWave40USDCSettlerLockHandlesWalletAndBalanceBranches(t *testing.T) {
	t.Parallel()

	walletAddr := "0x1111111111111111111111111111111111111111"
	usdcAddr := "0x2222222222222222222222222222222222222222"
	ctx := context.Background()

	t.Run("wallet address error stops before RPC", func(t *testing.T) {
		t.Parallel()

		api := &wave40EthAPI{balance: big.NewInt(100)}
		client := wave40EthClient(t, api)
		settler := wave40Settler(client, usdcAddr, &wave40Wallet{addressErr: errors.New("wallet offline")})

		err := settler.Lock(ctx, "buyer", big.NewInt(10))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "get agent wallet address")
		assert.Equal(t, 0, api.callCount())
	})

	t.Run("insufficient balance returns deterministic error", func(t *testing.T) {
		t.Parallel()

		api := &wave40EthAPI{balance: big.NewInt(99)}
		client := wave40EthClient(t, api)
		settler := wave40Settler(client, usdcAddr, &wave40Wallet{address: walletAddr})

		err := settler.Lock(ctx, "buyer", big.NewInt(100))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient USDC balance")
		assert.Equal(t, 1, api.callCount())
		assert.Equal(t, common.HexToAddress(usdcAddr), api.lastCallTo)
		require.Len(t, api.lastCallData, 4+32)
		assert.Equal(t, walletAddr, common.BytesToAddress(api.lastCallData[4+12:4+32]).Hex())
	})

	t.Run("sufficient balance passes", func(t *testing.T) {
		t.Parallel()

		api := &wave40EthAPI{balance: big.NewInt(100)}
		client := wave40EthClient(t, api)
		settler := wave40Settler(client, usdcAddr, &wave40Wallet{address: walletAddr})

		require.NoError(t, settler.Lock(ctx, "buyer", big.NewInt(100)))
		assert.Equal(t, 1, api.callCount())
	})
}

func TestWave40USDCSettlerReleaseAndRefundResolveErrors(t *testing.T) {
	t.Parallel()

	settler := NewUSDCSettler(&wave40Wallet{address: "0x1111111111111111111111111111111111111111"}, nil, nil, 8453)
	err := settler.Release(context.Background(), "not-a-did", big.NewInt(1))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolve seller address")
	assert.ErrorIs(t, err, ErrInvalidDID)

	resolverErr := errors.New("bundle lookup failed")
	settler = NewUSDCSettler(
		&wave40Wallet{address: "0x1111111111111111111111111111111111111111"},
		nil,
		nil,
		8453,
		WithAddressResolver(&wave40Resolver{err: resolverErr}),
	)

	err = settler.Refund(context.Background(), "did:lango:v2:buyer", big.NewInt(1))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolve buyer address")
	assert.ErrorIs(t, err, resolverErr)
}

func TestWave40USDCSettlerSignTxHandlesWalletErrorAndSuccess(t *testing.T) {
	t.Parallel()

	tx := wave40UnsignedTx()
	settler := NewUSDCSettler(&wave40Wallet{signErr: errors.New("sign denied")}, nil, nil, 8453)

	_, err := settler.signTx(context.Background(), tx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sign")

	key, err := crypto.HexToECDSA("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)
	wallet := &wave40Wallet{key: key}
	settler = NewUSDCSettler(wallet, nil, nil, 8453)

	signedTx, err := settler.signTx(context.Background(), tx)
	require.NoError(t, err)

	sender, err := types.Sender(types.LatestSignerForChainID(big.NewInt(8453)), signedTx)
	require.NoError(t, err)
	assert.Equal(t, crypto.PubkeyToAddress(key.PublicKey), sender)
	require.Len(t, wallet.signedHashes, 1)
	assert.Len(t, wallet.signedHashes[0], 32)
}

func TestWave40USDCSettlerSubmitWithRetryReturnsContextErrorWithoutLongSleep(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	api := &wave40EthAPI{
		sendErr: errors.New("temporary rpc failure"),
		onSend:  cancel,
	}
	client := wave40EthClient(t, api)
	settler := NewUSDCSettler(&wave40Wallet{}, nil, client, 8453, WithMaxRetries(3))

	_, err := settler.submitWithRetry(ctx, wave40UnsignedTx())

	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, api.sendCount())
}

func TestWave40USDCSettlerWaitForConfirmationBranches(t *testing.T) {
	t.Parallel()

	txHash := common.HexToHash("0x1234")

	t.Run("success receipt", func(t *testing.T) {
		t.Parallel()

		api := &wave40EthAPI{receipt: &types.Receipt{Status: types.ReceiptStatusSuccessful}}
		settler := NewUSDCSettler(nil, nil, wave40EthClient(t, api), 8453, WithReceiptTimeout(time.Second))

		require.NoError(t, settler.waitForConfirmation(context.Background(), txHash))
		assert.Equal(t, 1, api.receiptCount())
	})

	t.Run("reverted receipt", func(t *testing.T) {
		t.Parallel()

		api := &wave40EthAPI{receipt: &types.Receipt{Status: types.ReceiptStatusFailed}}
		settler := NewUSDCSettler(nil, nil, wave40EthClient(t, api), 8453, WithReceiptTimeout(time.Second))

		err := settler.waitForConfirmation(context.Background(), txHash)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tx reverted")
		assert.Equal(t, 1, api.receiptCount())
	})

	t.Run("timeout without long poll", func(t *testing.T) {
		t.Parallel()

		api := &wave40EthAPI{receiptErr: ethereum.NotFound}
		settler := NewUSDCSettler(nil, nil, wave40EthClient(t, api), 8453, WithReceiptTimeout(time.Nanosecond))

		err := settler.waitForConfirmation(context.Background(), txHash)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "receipt timeout")
		assert.Equal(t, 1, api.receiptCount())
	})
}

type wave40Wallet struct {
	address      string
	addressErr   error
	key          *ecdsa.PrivateKey
	signErr      error
	signedHashes [][]byte
}

func (w *wave40Wallet) Address(context.Context) (string, error) {
	if w.addressErr != nil {
		return "", w.addressErr
	}
	return w.address, nil
}

func (w *wave40Wallet) Balance(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (w *wave40Wallet) SignTransaction(_ context.Context, rawTx []byte) ([]byte, error) {
	if w.signErr != nil {
		return nil, w.signErr
	}
	if w.key == nil {
		return nil, errors.New("missing private key")
	}
	w.signedHashes = append(w.signedHashes, append([]byte(nil), rawTx...))
	return crypto.Sign(rawTx, w.key)
}

func (w *wave40Wallet) SignMessage(context.Context, []byte) ([]byte, error) {
	return nil, nil
}

func (w *wave40Wallet) PublicKey(context.Context) ([]byte, error) {
	return nil, nil
}

type wave40Resolver struct {
	addresses map[string]common.Address
	err       error
	calls     []string
}

func (r *wave40Resolver) ResolveAddress(did string) (common.Address, error) {
	r.calls = append(r.calls, did)
	if r.err != nil {
		return common.Address{}, r.err
	}
	addr, ok := r.addresses[did]
	if !ok {
		return common.Address{}, fmt.Errorf("%s: %w", did, ErrInvalidDID)
	}
	return addr, nil
}

type wave40EthAPI struct {
	mu sync.Mutex

	balance *big.Int

	calls        int
	lastCallTo   common.Address
	lastCallData []byte

	sendErr error
	onSend  func()
	sends   int

	receipt    *types.Receipt
	receiptErr error
	receipts   int
}

func (api *wave40EthAPI) Call(_ context.Context, msg map[string]interface{}, _ string) (hexutil.Bytes, error) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.calls++
	to, err := wave40RPCAddress(msg["to"])
	if err != nil {
		return nil, err
	}
	data, err := wave40RPCBytes(msg["data"])
	if err != nil && msg["data"] == nil {
		data, err = wave40RPCBytes(msg["input"])
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

func (api *wave40EthAPI) SendRawTransaction(context.Context, hexutil.Bytes) (common.Hash, error) {
	api.mu.Lock()
	api.sends++
	onSend := api.onSend
	if api.sendErr != nil {
		api.mu.Unlock()
		if onSend != nil {
			onSend()
		}
		return common.Hash{}, api.sendErr
	}
	api.mu.Unlock()
	if onSend != nil {
		onSend()
	}
	return common.HexToHash("0xbeef"), nil
}

func (api *wave40EthAPI) GetTransactionReceipt(_ context.Context, hash common.Hash) (map[string]interface{}, error) {
	api.mu.Lock()
	defer api.mu.Unlock()

	api.receipts++
	if api.receiptErr != nil {
		return nil, api.receiptErr
	}
	if api.receipt == nil {
		return nil, nil
	}
	return map[string]interface{}{
		"transactionHash":   hash,
		"blockHash":         common.HexToHash("0xbeef"),
		"blockNumber":       (*hexutil.Big)(big.NewInt(1)),
		"transactionIndex":  hexutil.Uint64(0),
		"from":              common.HexToAddress("0x1111111111111111111111111111111111111111"),
		"to":                common.HexToAddress("0x2222222222222222222222222222222222222222"),
		"gasUsed":           hexutil.Uint64(21_000),
		"cumulativeGasUsed": hexutil.Uint64(21_000),
		"contractAddress":   nil,
		"logs":              []*types.Log{},
		"logsBloom":         types.Bloom{},
		"status":            hexutil.Uint64(api.receipt.Status),
		"type":              hexutil.Uint64(2),
	}, nil
}

func (api *wave40EthAPI) callCount() int {
	api.mu.Lock()
	defer api.mu.Unlock()
	return api.calls
}

func (api *wave40EthAPI) sendCount() int {
	api.mu.Lock()
	defer api.mu.Unlock()
	return api.sends
}

func (api *wave40EthAPI) receiptCount() int {
	api.mu.Lock()
	defer api.mu.Unlock()
	return api.receipts
}

func wave40EthClient(t *testing.T, api *wave40EthAPI) *ethclient.Client {
	t.Helper()

	server := rpc.NewServer()
	require.NoError(t, server.RegisterName("eth", api))
	t.Cleanup(server.Stop)

	client := rpc.DialInProc(server)
	t.Cleanup(client.Close)

	return ethclient.NewClient(client)
}

func wave40Settler(client *ethclient.Client, usdcAddr string, wallet *wave40Wallet) *USDCSettler {
	txb := payment.NewTxBuilder(client, 8453, usdcAddr)
	return NewUSDCSettler(wallet, txb, client, 8453, WithReceiptTimeout(time.Second), WithMaxRetries(1))
}

func wave40UnsignedTx() *types.Transaction {
	to := common.HexToAddress("0x3333333333333333333333333333333333333333")
	return types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(8453),
		Nonce:     1,
		GasTipCap: big.NewInt(1),
		GasFeeCap: big.NewInt(2),
		Gas:       21_000,
		To:        &to,
		Value:     big.NewInt(0),
	})
}

func wave40RPCAddress(value interface{}) (common.Address, error) {
	switch v := value.(type) {
	case common.Address:
		return v, nil
	case string:
		return common.HexToAddress(v), nil
	default:
		return common.Address{}, fmt.Errorf("unexpected RPC address type %T", value)
	}
}

func wave40RPCBytes(value interface{}) ([]byte, error) {
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
