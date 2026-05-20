package contract

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const callerAbiAndLocalReadErrorsCallerABI = `[
	{
		"type":"function",
		"name":"balanceOf",
		"stateMutability":"view",
		"inputs":[{"name":"account","type":"address"}],
		"outputs":[{"name":"balance","type":"uint256"}]
	},
	{
		"type":"function",
		"name":"transfer",
		"stateMutability":"nonpayable",
		"inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],
		"outputs":[{"name":"ok","type":"bool"}]
	}
]`

func TestCallerABIAndLocalReadErrors(t *testing.T) {
	t.Parallel()

	address := common.HexToAddress("0x1111111111111111111111111111111111111111")

	tests := []struct {
		name      string
		req       ContractCallRequest
		wantError string
	}{
		{
			name: "invalid ABI fails before RPC",
			req: ContractCallRequest{
				ChainID: 1,
				Address: address,
				ABI:     "not json",
				Method:  "balanceOf",
			},
			wantError: "parse ABI",
		},
		{
			name: "missing method fails before RPC",
			req: ContractCallRequest{
				ChainID: 1,
				Address: address,
				ABI:     callerAbiAndLocalReadErrorsCallerABI,
				Method:  "allowance",
			},
			wantError: `method "allowance" not found in ABI`,
		},
		{
			name: "argument type mismatch fails before RPC",
			req: ContractCallRequest{
				ChainID: 1,
				Address: address,
				ABI:     callerAbiAndLocalReadErrorsCallerABI,
				Method:  "balanceOf",
				Args:    []interface{}{"not an address"},
			},
			wantError: `pack args for "balanceOf"`,
		},
		{
			name: "missing argument fails before RPC",
			req: ContractCallRequest{
				ChainID: 1,
				Address: address,
				ABI:     callerAbiAndLocalReadErrorsCallerABI,
				Method:  "balanceOf",
				Args:    []interface{}{},
			},
			wantError: `pack args for "balanceOf"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			caller := NewCaller(nil, nil, 1, NewABICache())
			result, err := caller.Read(context.Background(), tt.req)

			require.Nil(t, result)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantError)
		})
	}
}

func TestCallerReadUnpacksDeterministicABIResult(t *testing.T) {
	t.Parallel()

	parsed, err := ParseABI(callerAbiAndLocalReadErrorsCallerABI)
	require.NoError(t, err)
	output, err := parsed.Methods["balanceOf"].Outputs.Pack(big.NewInt(12345))
	require.NoError(t, err)
	input, err := parsed.Pack(
		"balanceOf",
		common.HexToAddress("0x7777777777777777777777777777777777777777"),
	)
	require.NoError(t, err)

	contractAddress := common.HexToAddress("0x6666666666666666666666666666666666666666")
	caller := callerAbiAndLocalReadErrorsCallerWithRPC(t, &callerAbiAndLocalReadErrorsContractAPI{
		expectedTo:   contractAddress,
		expectedData: input,
		output:       output,
	})
	result, err := caller.Read(context.Background(), ContractCallRequest{
		ChainID: 1,
		Address: contractAddress,
		ABI:     callerAbiAndLocalReadErrorsCallerABI,
		Method:  "balanceOf",
		Args: []interface{}{
			common.HexToAddress("0x7777777777777777777777777777777777777777"),
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Data, 1)
	assert.Equal(t, big.NewInt(12345), result.Data[0])
}

func TestCallerWriteLocalErrors(t *testing.T) {
	t.Parallel()

	address := common.HexToAddress("0x3333333333333333333333333333333333333333")
	to := common.HexToAddress("0x4444444444444444444444444444444444444444")

	tests := []struct {
		name      string
		req       ContractCallRequest
		wallet    *callerAbiAndLocalReadErrorsWallet
		wantError string
	}{
		{
			name: "invalid ABI fails before wallet and RPC",
			req: ContractCallRequest{
				ChainID: 1,
				Address: address,
				ABI:     "not json",
				Method:  "transfer",
			},
			wallet:    &callerAbiAndLocalReadErrorsWallet{},
			wantError: "parse ABI",
		},
		{
			name: "missing method fails before wallet and RPC",
			req: ContractCallRequest{
				ChainID: 1,
				Address: address,
				ABI:     callerAbiAndLocalReadErrorsCallerABI,
				Method:  "approve",
			},
			wallet:    &callerAbiAndLocalReadErrorsWallet{},
			wantError: `method "approve" not found in ABI`,
		},
		{
			name: "argument type mismatch fails before wallet and RPC",
			req: ContractCallRequest{
				ChainID: 1,
				Address: address,
				ABI:     callerAbiAndLocalReadErrorsCallerABI,
				Method:  "transfer",
				Args:    []interface{}{to, "not a uint256"},
			},
			wallet:    &callerAbiAndLocalReadErrorsWallet{},
			wantError: `pack args for "transfer"`,
		},
		{
			name: "wallet address error is wrapped before nonce lookup",
			req: ContractCallRequest{
				ChainID: 1,
				Address: address,
				ABI:     callerAbiAndLocalReadErrorsCallerABI,
				Method:  "transfer",
				Args:    []interface{}{to, big.NewInt(10)},
			},
			wallet: &callerAbiAndLocalReadErrorsWallet{
				addressErr: errors.New("wallet locked"),
			},
			wantError: "get wallet address: wallet locked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			caller := NewCaller(nil, tt.wallet, 1, NewABICache())
			result, err := caller.Write(context.Background(), tt.req)

			require.Nil(t, result)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantError)
		})
	}
}

func TestExtractRevertReason(t *testing.T) {
	t.Parallel()

	errorData := callerAbiAndLocalReadErrorsSolidityErrorData(t, "Caller is not the owner")

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "solidity Error string from hex string data",
			err:  callerAbiAndLocalReadErrorsDataError{data: errorData},
			want: "Caller is not the owner",
		},
		{
			name: "byte slice data is ignored by the current extractor",
			err: callerAbiAndLocalReadErrorsDataError{
				data: common.FromHex(
					"0x4e487b710000000000000000000000000000000000000000000000000000000000000012",
				),
			},
			want: "",
		},
		{
			name: "unsupported error data returns empty reason",
			err:  callerAbiAndLocalReadErrorsDataError{data: 12345},
			want: "",
		},
		{
			name: "plain errors return empty reason",
			err:  errors.New("execution reverted"),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractRevertReason(tt.err)
			if tt.want == "" {
				assert.Empty(t, got)
			} else {
				assert.Contains(t, got, tt.want)
			}
		})
	}
}

func TestWaitForReceiptTimeoutAndContext(t *testing.T) {
	t.Parallel()

	t.Run("times out without waiting for polling delay", func(t *testing.T) {
		t.Parallel()

		txHash := common.HexToHash("0xdef")
		caller := callerAbiAndLocalReadErrorsCallerWithRPC(t, &callerAbiAndLocalReadErrorsReceiptAPI{})
		caller.timeout = time.Millisecond

		got, err := caller.waitForReceipt(context.Background(), txHash)

		require.Nil(t, got)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrReceiptTimeout)
		assert.Contains(t, err.Error(), txHash.Hex())
	})

	t.Run("returns context cancellation", func(t *testing.T) {
		t.Parallel()

		txHash := common.HexToHash("0x123")
		caller := callerAbiAndLocalReadErrorsCallerWithRPC(t, &callerAbiAndLocalReadErrorsReceiptAPI{})
		caller.timeout = time.Second
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		got, err := caller.waitForReceipt(ctx, txHash)

		require.Nil(t, got)
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

type callerAbiAndLocalReadErrorsWallet struct {
	addressErr error
}

func (w *callerAbiAndLocalReadErrorsWallet) Address(context.Context) (string, error) {
	if w.addressErr != nil {
		return "", w.addressErr
	}
	return "0x5555555555555555555555555555555555555555", nil
}

func (w *callerAbiAndLocalReadErrorsWallet) Balance(context.Context) (*big.Int, error) {
	return nil, errors.New("not implemented")
}

func (w *callerAbiAndLocalReadErrorsWallet) SignTransaction(context.Context, []byte) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (w *callerAbiAndLocalReadErrorsWallet) SignMessage(context.Context, []byte) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (w *callerAbiAndLocalReadErrorsWallet) PublicKey(context.Context) ([]byte, error) {
	return nil, errors.New("not implemented")
}

type callerAbiAndLocalReadErrorsDataError struct {
	data interface{}
}

func (e callerAbiAndLocalReadErrorsDataError) Error() string {
	return "execution reverted"
}

func (e callerAbiAndLocalReadErrorsDataError) ErrorData() interface{} {
	return e.data
}

type callerAbiAndLocalReadErrorsReceiptAPI struct {
	receipt *types.Receipt
}

func (api *callerAbiAndLocalReadErrorsReceiptAPI) GetTransactionReceipt(
	context.Context,
	common.Hash,
) (*types.Receipt, error) {
	return api.receipt, nil
}

type callerAbiAndLocalReadErrorsContractAPI struct {
	expectedTo   common.Address
	expectedData []byte
	output       hexutil.Bytes
}

func (api *callerAbiAndLocalReadErrorsContractAPI) Call(
	_ context.Context,
	msg map[string]interface{},
	block string,
) (hexutil.Bytes, error) {
	to, err := callerAbiAndLocalReadErrorsRPCAddress(msg["to"])
	if err != nil {
		return nil, err
	}
	if to != api.expectedTo.Hex() {
		return nil, fmt.Errorf("unexpected call target: got %s want %s", to, api.expectedTo.Hex())
	}

	dataValue, ok := msg["data"]
	if !ok {
		dataValue = msg["input"]
	}
	data, err := callerAbiAndLocalReadErrorsRPCBytes(dataValue)
	if err != nil {
		return nil, err
	}
	if common.Bytes2Hex(data) != common.Bytes2Hex(api.expectedData) {
		return nil, fmt.Errorf("unexpected calldata: got %s want %s", common.Bytes2Hex(data), common.Bytes2Hex(api.expectedData))
	}

	if block != "latest" {
		return nil, fmt.Errorf("unexpected block: got %s want latest", block)
	}

	return api.output, nil
}

func callerAbiAndLocalReadErrorsRPCAddress(value interface{}) (string, error) {
	switch v := value.(type) {
	case common.Address:
		return v.Hex(), nil
	case string:
		return common.HexToAddress(v).Hex(), nil
	default:
		return "", fmt.Errorf("unexpected RPC address type %T", value)
	}
}

func callerAbiAndLocalReadErrorsRPCBytes(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case hexutil.Bytes:
		return []byte(v), nil
	case []byte:
		return v, nil
	case string:
		return common.FromHex(v), nil
	default:
		return nil, fmt.Errorf("unexpected RPC bytes type %T", value)
	}
}

func callerAbiAndLocalReadErrorsCallerWithRPC(t *testing.T, api interface{}) *Caller {
	t.Helper()

	server := rpc.NewServer()
	require.NoError(t, server.RegisterName("eth", api))
	t.Cleanup(server.Stop)

	client := rpc.DialInProc(server)
	t.Cleanup(client.Close)

	return NewCaller(ethclient.NewClient(client), nil, 1, NewABICache())
}

func callerAbiAndLocalReadErrorsSolidityErrorData(t *testing.T, reason string) string {
	t.Helper()

	stringType, err := abi.NewType("string", "", nil)
	require.NoError(t, err)
	encoded, err := abi.Arguments{{Type: stringType}}.Pack(reason)
	require.NoError(t, err)

	return "0x08c379a0" + common.Bytes2Hex(encoded)
}
