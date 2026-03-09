package paymaster

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

// CircleProvider implements PaymasterProvider using Circle's Paymaster API.
type CircleProvider struct {
	url        string
	httpClient *http.Client
	reqID      atomic.Int64
}

// NewCircleProvider creates a Circle paymaster provider.
func NewCircleProvider(rpcURL string) *CircleProvider {
	return &CircleProvider{
		url:        rpcURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *CircleProvider) Type() string { return "circle" }

func (c *CircleProvider) SponsorUserOp(ctx context.Context, req *SponsorRequest) (*SponsorResult, error) {
	opMap := userOpToMap(req.UserOp)

	params := []interface{}{
		opMap,
		req.EntryPoint.Hex(),
	}

	raw, err := c.call(ctx, "pm_sponsorUserOperation", params)
	if err != nil {
		return nil, fmt.Errorf("circle sponsor: %w", err)
	}

	return parseSponsorResponse(raw)
}

func (c *CircleProvider) call(ctx context.Context, method string, params []interface{}) (json.RawMessage, error) {
	if params == nil {
		params = make([]interface{}, 0)
	}

	reqID := int(c.reqID.Add(1))
	rpcReq := jsonrpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      reqID,
	}

	body, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("paymaster RPC call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paymaster HTTP %d: %s: %w", resp.StatusCode, string(respBody), ErrPaymasterRejected)
	}

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("paymaster RPC error %d: %s: %w", rpcResp.Error.Code, rpcResp.Error.Message, ErrPaymasterRejected)
	}

	return rpcResp.Result, nil
}

// shared JSON-RPC types
type jsonrpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
	ID      int             `json:"id"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// userOpToMap converts UserOpData to JSON-RPC hex-encoded format.
func userOpToMap(op *UserOpData) map[string]interface{} {
	return map[string]interface{}{
		"sender":               op.Sender.Hex(),
		"nonce":                encodeBigInt(op.Nonce),
		"initCode":             hexutil.Encode(op.InitCode),
		"callData":             hexutil.Encode(op.CallData),
		"callGasLimit":         encodeBigInt(op.CallGasLimit),
		"verificationGasLimit": encodeBigInt(op.VerificationGasLimit),
		"preVerificationGas":   encodeBigInt(op.PreVerificationGas),
		"maxFeePerGas":         encodeBigInt(op.MaxFeePerGas),
		"maxPriorityFeePerGas": encodeBigInt(op.MaxPriorityFeePerGas),
		"paymasterAndData":     hexutil.Encode(op.PaymasterAndData),
		"signature":            hexutil.Encode(op.Signature),
	}
}

func encodeBigInt(n *big.Int) string {
	if n == nil {
		return "0x0"
	}
	return hexutil.EncodeBig(n)
}

// parseSponsorResponse parses the paymaster sponsorship response.
func parseSponsorResponse(raw json.RawMessage) (*SponsorResult, error) {
	var resp struct {
		PaymasterAndData     string `json:"paymasterAndData"`
		CallGasLimit         string `json:"callGasLimit,omitempty"`
		VerificationGasLimit string `json:"verificationGasLimit,omitempty"`
		PreVerificationGas   string `json:"preVerificationGas,omitempty"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode sponsor response: %w", err)
	}

	pmData := common.FromHex(resp.PaymasterAndData)
	if len(pmData) == 0 {
		return nil, fmt.Errorf("empty paymasterAndData: %w", ErrPaymasterRejected)
	}

	result := &SponsorResult{
		PaymasterAndData: pmData,
	}

	// Parse optional gas overrides
	var overrides GasOverrides
	hasOverrides := false

	if resp.CallGasLimit != "" {
		v, err := hexutil.DecodeBig(resp.CallGasLimit)
		if err == nil {
			overrides.CallGasLimit = v
			hasOverrides = true
		}
	}
	if resp.VerificationGasLimit != "" {
		v, err := hexutil.DecodeBig(resp.VerificationGasLimit)
		if err == nil {
			overrides.VerificationGasLimit = v
			hasOverrides = true
		}
	}
	if resp.PreVerificationGas != "" {
		v, err := hexutil.DecodeBig(resp.PreVerificationGas)
		if err == nil {
			overrides.PreVerificationGas = v
			hasOverrides = true
		}
	}

	if hasOverrides {
		result.GasOverrides = &overrides
	}

	return result, nil
}
