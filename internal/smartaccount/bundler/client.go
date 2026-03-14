package bundler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Client communicates with an ERC-4337 bundler via JSON-RPC.
type Client struct {
	url        string
	httpClient *http.Client
	entryPoint common.Address
	reqID      atomic.Int64
}

// NewClient creates a bundler client.
func NewClient(
	bundlerURL string,
	entryPoint common.Address,
) *Client {
	return &Client{
		url:        bundlerURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		entryPoint: entryPoint,
	}
}

// SendUserOperation submits a UserOp to the bundler.
func (c *Client) SendUserOperation(
	ctx context.Context,
	op *UserOperation,
) (*UserOpResult, error) {
	if op == nil {
		return nil, fmt.Errorf(
			"send user operation: %w", ErrInvalidUserOp,
		)
	}

	opMap := userOpToMap(op)
	raw, err := c.call(
		ctx,
		"eth_sendUserOperation",
		[]interface{}{opMap, c.entryPoint},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"send user operation: %w", err,
		)
	}

	var hashHex string
	if err := json.Unmarshal(raw, &hashHex); err != nil {
		return nil, fmt.Errorf(
			"decode user op hash: %w", err,
		)
	}

	return &UserOpResult{
		UserOpHash: common.HexToHash(hashHex),
		Success:    true,
	}, nil
}

// EstimateGas estimates gas for a UserOp.
func (c *Client) EstimateGas(
	ctx context.Context,
	op *UserOperation,
) (*GasEstimate, error) {
	if op == nil {
		return nil, fmt.Errorf(
			"estimate gas: %w", ErrInvalidUserOp,
		)
	}

	opMap := userOpToMap(op)
	raw, err := c.call(
		ctx,
		"eth_estimateUserOperationGas",
		[]interface{}{opMap, c.entryPoint},
	)
	if err != nil {
		return nil, fmt.Errorf("estimate gas: %w", err)
	}

	var result struct {
		CallGasLimit         string `json:"callGasLimit"`
		VerificationGasLimit string `json:"verificationGasLimit"`
		PreVerificationGas   string `json:"preVerificationGas"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode gas estimate: %w", err)
	}

	callGas, err := hexutil.DecodeBig(result.CallGasLimit)
	if err != nil {
		return nil, fmt.Errorf(
			"decode callGasLimit: %w", err,
		)
	}
	verificationGas, err := hexutil.DecodeBig(
		result.VerificationGasLimit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"decode verificationGasLimit: %w", err,
		)
	}
	preVerificationGas, err := hexutil.DecodeBig(
		result.PreVerificationGas,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"decode preVerificationGas: %w", err,
		)
	}

	return &GasEstimate{
		CallGasLimit:         callGas,
		VerificationGasLimit: verificationGas,
		PreVerificationGas:   preVerificationGas,
	}, nil
}

// GetUserOperationReceipt gets the receipt for a UserOp hash.
func (c *Client) GetUserOperationReceipt(
	ctx context.Context,
	hash common.Hash,
) (*UserOpResult, error) {
	raw, err := c.call(
		ctx,
		"eth_getUserOperationReceipt",
		[]interface{}{hash.Hex()},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get user op receipt: %w", err,
		)
	}

	var receipt struct {
		UserOpHash string `json:"userOpHash"`
		TxHash     string `json:"transactionHash"`
		Success    bool   `json:"success"`
		ActualGas  string `json:"actualGasUsed"`
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return nil, fmt.Errorf(
			"decode user op receipt: %w", err,
		)
	}

	var gasUsed uint64
	if receipt.ActualGas != "" {
		gas, err := hexutil.DecodeUint64(receipt.ActualGas)
		if err == nil {
			gasUsed = gas
		}
	}

	return &UserOpResult{
		UserOpHash: common.HexToHash(receipt.UserOpHash),
		TxHash:     common.HexToHash(receipt.TxHash),
		Success:    receipt.Success,
		GasUsed:    gasUsed,
	}, nil
}

// getNonceSelector is the function selector for
// EntryPoint.getNonce(address,uint192) → 0x35567e1a.
var getNonceSelector = common.FromHex("0x35567e1a")

// GetNonce retrieves the nonce for an account from the EntryPoint
// contract via eth_call to EntryPoint.getNonce(address, key=0).
func (c *Client) GetNonce(
	ctx context.Context,
	account common.Address,
) (*big.Int, error) {
	// ABI-encode: getNonce(address sender, uint192 key)
	// selector (4 bytes) + address padded to 32 + key padded to 32
	calldata := make([]byte, 0, 68)
	calldata = append(calldata, getNonceSelector...)
	// Left-pad address to 32 bytes.
	addrPadded := make([]byte, 32)
	copy(addrPadded[12:], account.Bytes())
	calldata = append(calldata, addrPadded...)
	// key = 0 (32 zero bytes for sequential nonce).
	calldata = append(calldata, make([]byte, 32)...)

	callMsg := map[string]interface{}{
		"to":   c.entryPoint.Hex(),
		"data": hexutil.Encode(calldata),
	}

	raw, err := c.call(
		ctx,
		"eth_call",
		[]interface{}{callMsg, "latest"},
	)
	if err != nil {
		return nil, fmt.Errorf("get entrypoint nonce: %w", err)
	}

	var hexResult string
	if err := json.Unmarshal(raw, &hexResult); err != nil {
		return nil, fmt.Errorf("decode nonce result: %w", err)
	}

	// eth_call returns ABI-encoded uint256 (0-padded to 32 bytes).
	// Use hexutil.Decode (accepts leading zeros) instead of DecodeBig.
	resultBytes, err := hexutil.Decode(hexResult)
	if err != nil {
		return nil, fmt.Errorf("parse nonce: %w", err)
	}
	return new(big.Int).SetBytes(resultBytes), nil
}

// defaultMaxPriorityFeeWei is the fallback priority fee (1.5 gwei)
// when eth_maxPriorityFeePerGas is not supported.
const defaultMaxPriorityFeeWei = 1_500_000_000

// baseFeeMultiplier doubles the base fee for safety margin.
const baseFeeMultiplier = 2

// GetGasFees retrieves EIP-1559 gas fee parameters from the network.
// Uses eth_maxPriorityFeePerGas for priority fee and the latest block
// header for base fee. Falls back to defaults if RPC calls fail.
func (c *Client) GetGasFees(
	ctx context.Context,
) (*GasFees, error) {
	// Get priority fee from RPC.
	priorityFee := big.NewInt(defaultMaxPriorityFeeWei)
	raw, err := c.call(
		ctx,
		"eth_maxPriorityFeePerGas",
		nil,
	)
	if err == nil {
		var hexFee string
		if jsonErr := json.Unmarshal(raw, &hexFee); jsonErr == nil {
			if decoded, decErr := hexutil.DecodeBig(hexFee); decErr == nil {
				priorityFee = decoded
			}
		}
	}
	// If eth_maxPriorityFeePerGas fails, use the default — not an error.

	// Get base fee from latest block.
	raw, err = c.call(
		ctx,
		"eth_getBlockByNumber",
		[]interface{}{"latest", false},
	)
	if err != nil {
		return nil, fmt.Errorf("get latest block: %w", err)
	}

	var block struct {
		BaseFeePerGas string `json:"baseFeePerGas"`
	}
	if err := json.Unmarshal(raw, &block); err != nil {
		return nil, fmt.Errorf("decode block: %w", err)
	}

	baseFee := big.NewInt(1_000_000_000) // 1 gwei default
	if block.BaseFeePerGas != "" {
		if decoded, decErr := hexutil.DecodeBig(
			block.BaseFeePerGas,
		); decErr == nil {
			baseFee = decoded
		}
	}

	// maxFeePerGas = 2 * baseFee + priorityFee
	maxFee := new(big.Int).Add(
		new(big.Int).Mul(
			baseFee, big.NewInt(baseFeeMultiplier),
		),
		priorityFee,
	)

	return &GasFees{
		MaxFeePerGas:         maxFee,
		MaxPriorityFeePerGas: priorityFee,
	}, nil
}

// SupportedEntryPoints returns supported entry point addresses.
func (c *Client) SupportedEntryPoints(
	ctx context.Context,
) ([]common.Address, error) {
	raw, err := c.call(
		ctx, "eth_supportedEntryPoints", nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get supported entry points: %w", err,
		)
	}

	var hexAddrs []string
	if err := json.Unmarshal(raw, &hexAddrs); err != nil {
		return nil, fmt.Errorf(
			"decode entry points: %w", err,
		)
	}

	addrs := make([]common.Address, len(hexAddrs))
	for i, h := range hexAddrs {
		addrs[i] = common.HexToAddress(h)
	}
	return addrs, nil
}

// call makes a JSON-RPC call.
func (c *Client) call(
	ctx context.Context,
	method string,
	params []interface{},
) (json.RawMessage, error) {
	if params == nil {
		params = make([]interface{}, 0)
	}

	reqID := int(c.reqID.Add(1))
	req := jsonrpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      reqID,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx, http.MethodPost, c.url, bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("bundler RPC call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"bundler HTTP %d: %s: %w",
			resp.StatusCode, string(respBody), ErrBundlerError,
		)
	}

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf(
			"bundler RPC error %d: %s: %w",
			rpcResp.Error.Code,
			rpcResp.Error.Message,
			ErrBundlerError,
		)
	}

	return rpcResp.Result, nil
}

// userOpToMap converts a UserOp to the JSON-RPC hex-encoded
// format expected by ERC-4337 bundlers.
func userOpToMap(
	op *UserOperation,
) map[string]interface{} {
	m := map[string]interface{}{
		"sender":   op.Sender.Hex(),
		"nonce":    encodeBigInt(op.Nonce),
		"initCode": hexutil.Encode(op.InitCode),
		"callData": hexutil.Encode(op.CallData),
		"callGasLimit": encodeBigInt(
			op.CallGasLimit,
		),
		"verificationGasLimit": encodeBigInt(
			op.VerificationGasLimit,
		),
		"preVerificationGas": encodeBigInt(
			op.PreVerificationGas,
		),
		"maxFeePerGas": encodeBigInt(
			op.MaxFeePerGas,
		),
		"maxPriorityFeePerGas": encodeBigInt(
			op.MaxPriorityFeePerGas,
		),
		"paymasterAndData": hexutil.Encode(
			op.PaymasterAndData,
		),
		"signature": hexutil.Encode(op.Signature),
	}
	return m
}

// encodeBigInt encodes a *big.Int to a hex string,
// defaulting to "0x0" if nil.
func encodeBigInt(n *big.Int) string {
	if n == nil {
		return "0x0"
	}
	return hexutil.EncodeBig(n)
}
