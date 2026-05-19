package contract

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWave50CallerWriteSubmitsSignedTransactionAndReturnsReceipt(t *testing.T) {
	t.Parallel()

	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	from := crypto.PubkeyToAddress(key.PublicKey)
	to := common.HexToAddress("0x5050505050505050505050505050505050505050")
	recipient := common.HexToAddress("0x5151515151515151515151515151515151515151")
	value := big.NewInt(7)

	parsed, err := ParseABI(wave25CallerABI)
	require.NoError(t, err)
	expectedData, err := parsed.Pack("transfer", recipient, big.NewInt(42))
	require.NoError(t, err)

	api := &wave50WriteAPI{
		expectedFrom:  from,
		expectedTo:    to,
		expectedData:  expectedData,
		expectedValue: value,
		nonce:         9,
		gasLimit:      54321,
		header:        &types.Header{Number: big.NewInt(100), Difficulty: big.NewInt(0), BaseFee: big.NewInt(1000)},
		receipt: &types.Receipt{
			Status:            types.ReceiptStatusSuccessful,
			CumulativeGasUsed: 33333,
			GasUsed:           33333,
			Logs:              []*types.Log{},
		},
	}
	caller := wave50CallerWithRPC(t, api, &wave50SigningWallet{key: key})
	caller.maxRetries = 1
	caller.timeout = 50 * time.Millisecond

	result, err := caller.Write(context.Background(), ContractCallRequest{
		ChainID: 1,
		Address: to,
		ABI:     wave25CallerABI,
		Method:  "transfer",
		Args:    []interface{}{recipient, big.NewInt(42)},
		Value:   value,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, api.sentTransactions, 1)
	sent := api.sentTransactions[0]
	assert.Equal(t, sent.Hash().Hex(), result.TxHash)
	assert.Equal(t, uint64(33333), result.GasUsed)
	assert.Equal(t, uint64(9), sent.Nonce())
	assert.Equal(t, uint64(54321), sent.Gas())
	assert.Equal(t, value, sent.Value())
	require.NotNil(t, sent.To())
	assert.Equal(t, to, *sent.To())
	assert.Equal(t, expectedData, sent.Data())
	assert.Equal(t, []string{"pending"}, api.nonceBlocks)
	assert.Equal(t, []string{"latest"}, api.headerBlocks)
	assert.Equal(t, []common.Hash{sent.Hash()}, api.receiptHashes)
}

func TestWave50CallerWriteReplaysGasEstimateFailureForRevertReason(t *testing.T) {
	t.Parallel()

	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	from := crypto.PubkeyToAddress(key.PublicKey)
	to := common.HexToAddress("0x5252525252525252525252525252525252525252")
	recipient := common.HexToAddress("0x5353535353535353535353535353535353535353")
	estimateErr := errors.New("execution reverted")

	api := &wave50WriteAPI{
		expectedFrom: from,
		expectedTo:   to,
		nonce:        1,
		estimateErr:  estimateErr,
		callErr:      wave25DataError{data: wave25SolidityErrorData(t, "transfer is paused")},
	}
	caller := wave50CallerWithRPC(t, api, &wave50SigningWallet{key: key})

	result, err := caller.Write(context.Background(), ContractCallRequest{
		ChainID: 1,
		Address: to,
		ABI:     wave25CallerABI,
		Method:  "transfer",
		Args:    []interface{}{recipient, big.NewInt(1)},
	})

	require.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "estimate gas (revert: transfer is paused)")
	assert.Contains(t, err.Error(), estimateErr.Error())
	assert.Equal(t, []string{"latest"}, api.callBlocks)
	assert.Empty(t, api.sentTransactions)
}

func TestWave50ReplayForRevertReasonUsesProvidedBlockNumber(t *testing.T) {
	t.Parallel()

	api := &wave50WriteAPI{
		callErr: wave25DataError{data: wave25SolidityErrorData(t, "block scoped revert")},
	}
	caller := wave50CallerWithRPC(t, api, &wave50SigningWallet{})

	reason := caller.replayForRevertReason(
		context.Background(),
		common.HexToAddress("0x5454545454545454545454545454545454545454"),
		common.HexToAddress("0x5555555555555555555555555555555555555555"),
		[]byte{0xaa, 0xbb},
		big.NewInt(3),
		big.NewInt(42),
	)

	assert.Contains(t, reason, "block scoped revert")
	assert.Equal(t, []string{"0x2a"}, api.callBlocks)
}

type wave50SigningWallet struct {
	key *ecdsa.PrivateKey
	err error
}

func (w *wave50SigningWallet) Address(context.Context) (string, error) {
	if w.err != nil {
		return "", w.err
	}
	if w.key == nil {
		return common.Address{}.Hex(), nil
	}
	return crypto.PubkeyToAddress(w.key.PublicKey).Hex(), nil
}

func (w *wave50SigningWallet) Balance(context.Context) (*big.Int, error) {
	return nil, errors.New("not implemented")
}

func (w *wave50SigningWallet) SignTransaction(_ context.Context, digest []byte) ([]byte, error) {
	if w.err != nil {
		return nil, w.err
	}
	if w.key == nil {
		return nil, errors.New("missing signing key")
	}
	return crypto.Sign(digest, w.key)
}

func (w *wave50SigningWallet) SignMessage(context.Context, []byte) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (w *wave50SigningWallet) PublicKey(context.Context) ([]byte, error) {
	return nil, errors.New("not implemented")
}

type wave50WriteAPI struct {
	expectedFrom  common.Address
	expectedTo    common.Address
	expectedData  []byte
	expectedValue *big.Int
	nonce         uint64
	gasLimit      uint64
	estimateErr   error
	header        *types.Header
	headerErr     error
	callErr       error
	receipt       *types.Receipt
	receiptErr    error

	nonceBlocks      []string
	headerBlocks     []string
	callBlocks       []string
	receiptHashes    []common.Hash
	sentTransactions []*types.Transaction
}

func (api *wave50WriteAPI) GetTransactionCount(
	_ context.Context,
	address common.Address,
	block string,
) (hexutil.Uint64, error) {
	if api.expectedFrom != (common.Address{}) && address != api.expectedFrom {
		return 0, fmt.Errorf("unexpected nonce address: got %s want %s", address.Hex(), api.expectedFrom.Hex())
	}
	api.nonceBlocks = append(api.nonceBlocks, block)
	return hexutil.Uint64(api.nonce), nil
}

func (api *wave50WriteAPI) EstimateGas(
	_ context.Context,
	msg map[string]interface{},
) (hexutil.Uint64, error) {
	if err := api.assertCallMessage(msg); err != nil {
		return 0, err
	}
	if api.estimateErr != nil {
		return 0, api.estimateErr
	}
	return hexutil.Uint64(api.gasLimit), nil
}

func (api *wave50WriteAPI) GetBlockByNumber(
	_ context.Context,
	block string,
	_ bool,
) (*types.Header, error) {
	api.headerBlocks = append(api.headerBlocks, block)
	if api.headerErr != nil {
		return nil, api.headerErr
	}
	return api.header, nil
}

func (api *wave50WriteAPI) SendRawTransaction(
	_ context.Context,
	raw hexutil.Bytes,
) (common.Hash, error) {
	var tx types.Transaction
	if err := tx.UnmarshalBinary(raw); err != nil {
		return common.Hash{}, err
	}
	api.sentTransactions = append(api.sentTransactions, &tx)
	return tx.Hash(), nil
}

func (api *wave50WriteAPI) GetTransactionReceipt(
	_ context.Context,
	hash common.Hash,
) (*types.Receipt, error) {
	api.receiptHashes = append(api.receiptHashes, hash)
	if api.receiptErr != nil {
		return nil, api.receiptErr
	}
	if api.receipt == nil {
		return nil, nil
	}
	receipt := *api.receipt
	receipt.TxHash = hash
	receipt.BlockNumber = big.NewInt(100)
	receipt.BlockHash = common.HexToHash("0x5050")
	return &receipt, nil
}

func (api *wave50WriteAPI) Call(
	_ context.Context,
	msg map[string]interface{},
	block string,
) (hexutil.Bytes, error) {
	api.callBlocks = append(api.callBlocks, block)
	if err := api.assertCallMessage(msg); err != nil {
		return nil, err
	}
	return nil, api.callErr
}

func (api *wave50WriteAPI) assertCallMessage(msg map[string]interface{}) error {
	if api.expectedFrom != (common.Address{}) {
		from, err := wave25RPCAddress(msg["from"])
		if err != nil {
			return err
		}
		if from != api.expectedFrom.Hex() {
			return fmt.Errorf("unexpected from: got %s want %s", from, api.expectedFrom.Hex())
		}
	}
	if api.expectedTo != (common.Address{}) {
		to, err := wave25RPCAddress(msg["to"])
		if err != nil {
			return err
		}
		if to != api.expectedTo.Hex() {
			return fmt.Errorf("unexpected to: got %s want %s", to, api.expectedTo.Hex())
		}
	}
	if len(api.expectedData) > 0 {
		dataValue, ok := msg["data"]
		if !ok {
			dataValue = msg["input"]
		}
		data, err := wave25RPCBytes(dataValue)
		if err != nil {
			return err
		}
		if common.Bytes2Hex(data) != common.Bytes2Hex(api.expectedData) {
			return fmt.Errorf("unexpected data: got %s want %s", common.Bytes2Hex(data), common.Bytes2Hex(api.expectedData))
		}
	}
	if api.expectedValue != nil {
		value, err := wave50RPCBigInt(msg["value"])
		if err != nil {
			return err
		}
		if value.Cmp(api.expectedValue) != 0 {
			return fmt.Errorf("unexpected value: got %s want %s", value, api.expectedValue)
		}
	}
	return nil
}

func wave50RPCBigInt(value interface{}) (*big.Int, error) {
	switch v := value.(type) {
	case nil:
		return new(big.Int), nil
	case *big.Int:
		return new(big.Int).Set(v), nil
	case hexutil.Big:
		return (*big.Int)(&v), nil
	case string:
		n, ok := new(big.Int).SetString(v, 0)
		if !ok {
			return nil, fmt.Errorf("invalid big integer %q", v)
		}
		return n, nil
	default:
		return nil, fmt.Errorf("unexpected RPC integer type %T", value)
	}
}

func wave50CallerWithRPC(t *testing.T, api interface{}, wallet *wave50SigningWallet) *Caller {
	t.Helper()

	server := rpc.NewServer()
	require.NoError(t, server.RegisterName("eth", api))
	t.Cleanup(server.Stop)

	client := rpc.DialInProc(server)
	t.Cleanup(client.Close)

	return NewCaller(ethclient.NewClient(client), wallet, 1, NewABICache())
}
