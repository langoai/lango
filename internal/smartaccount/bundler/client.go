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

// GetNonce retrieves the nonce for an account from the EntryPoint
// contract. Uses eth_getTransactionCount as a fallback nonce source.
func (c *Client) GetNonce(
	ctx context.Context,
	account common.Address,
) (*big.Int, error) {
	raw, err := c.call(
		ctx,
		"eth_getTransactionCount",
		[]interface{}{account.Hex(), "latest"},
	)
	if err != nil {
		return nil, fmt.Errorf("get nonce: %w", err)
	}

	var hexNonce string
	if err := json.Unmarshal(raw, &hexNonce); err != nil {
		return nil, fmt.Errorf("decode nonce: %w", err)
	}

	nonce, err := hexutil.DecodeBig(hexNonce)
	if err != nil {
		return nil, fmt.Errorf("parse nonce: %w", err)
	}
	return nonce, nil
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
