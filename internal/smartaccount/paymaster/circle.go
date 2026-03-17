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

	"github.com/langoai/lango/internal/smartaccount/paymaster/permit"
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

// CirclePermitProvider implements PaymasterProvider using Circle's on-chain
// permissionless paymaster with EIP-2612 permit mode. No API key required.
type CirclePermitProvider struct {
	paymasterAddr   common.Address
	usdcAddr        common.Address
	chainID         int64
	signer          permit.PermitSigner
	ethCaller       permit.EthCaller
	verificationGas *big.Int
	postOpGas       *big.Int
}

// NewCirclePermitProvider creates a Circle paymaster provider in permit mode.
func NewCirclePermitProvider(
	paymasterAddr common.Address,
	usdcAddr common.Address,
	chainID int64,
	signer permit.PermitSigner,
	ethCaller permit.EthCaller,
) *CirclePermitProvider {
	return &CirclePermitProvider{
		paymasterAddr:   paymasterAddr,
		usdcAddr:        usdcAddr,
		chainID:         chainID,
		signer:          signer,
		ethCaller:       ethCaller,
		verificationGas: big.NewInt(150000),
		postOpGas:       big.NewInt(80000),
	}
}

func (c *CirclePermitProvider) Type() string { return "circle-permit" }

// SponsorUserOp builds PaymasterAndData for Circle's on-chain permit paymaster.
//
// When stub=true, returns a dummy PaymasterAndData with correct length for gas
// estimation. When stub=false, queries the permit nonce, signs an EIP-2612
// permit, and assembles the full paymasterData.
//
// PaymasterAndData layout (v0.7 packed):
//
//	paymaster(20) + verificationGas(16) + postOpGas(16) + paymasterData(variable)
//
// paymasterData layout (permit mode):
//
//	mode(1) + token(20) + amount(32) + signature(65) = 118 bytes
func (c *CirclePermitProvider) SponsorUserOp(ctx context.Context, req *SponsorRequest) (*SponsorResult, error) {
	// Build the fixed prefix: paymaster(20) + verGas(16) + postOpGas(16) = 52 bytes.
	prefix := make([]byte, 52)
	copy(prefix[:20], c.paymasterAddr.Bytes())
	c.verificationGas.FillBytes(prefix[20:36])
	c.postOpGas.FillBytes(prefix[36:52])

	if req.Stub {
		// Stub: prefix + zero paymasterData of correct length (118 bytes).
		stub := make([]byte, 52+118)
		copy(stub[:52], prefix)
		return &SponsorResult{PaymasterAndData: stub}, nil
	}

	// Get the owner address.
	ownerHex, err := c.signer.Address(ctx)
	if err != nil {
		return nil, fmt.Errorf("get signer address: %w", err)
	}
	owner := common.HexToAddress(ownerHex)

	// Query the permit nonce.
	nonce, err := permit.GetPermitNonce(ctx, c.ethCaller, c.usdcAddr, owner)
	if err != nil {
		return nil, fmt.Errorf("get permit nonce: %w", err)
	}

	// Use a generous approval amount to cover gas cost.
	// 10 USDC (6 decimals) should be more than enough for any single UserOp.
	amount := big.NewInt(10_000_000)

	// Deadline: 10 minutes from now.
	deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())

	// Sign the EIP-2612 permit.
	v, r, s, err := permit.Sign(
		ctx, c.signer,
		owner, c.paymasterAddr, amount, nonce, deadline,
		c.chainID, c.usdcAddr,
	)
	if err != nil {
		return nil, fmt.Errorf("sign permit: %w", err)
	}

	// Build paymasterData: mode(1) + token(20) + amount(32) + sig(65) = 118 bytes.
	pmData := make([]byte, 118)
	pmData[0] = 0x01 // permit mode
	copy(pmData[1:21], c.usdcAddr.Bytes())
	amount.FillBytes(pmData[21:53])

	// Pack signature: r(32) + s(32) + v(1).
	copy(pmData[53:85], r[:])
	copy(pmData[85:117], s[:])
	pmData[117] = v

	// Assemble full PaymasterAndData: prefix(52) + pmData(118) = 170 bytes.
	full := make([]byte, 52+118)
	copy(full[:52], prefix)
	copy(full[52:], pmData)

	return &SponsorResult{PaymasterAndData: full}, nil
}
