package paymaster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

// AlchemyProvider implements PaymasterProvider using Alchemy's Gas Manager API.
// Uses the combined alchemy_requestGasAndPaymasterAndData endpoint.
type AlchemyProvider struct {
	url        string
	policyID   string
	httpClient *http.Client
	reqID      atomic.Int64
}

// NewAlchemyProvider creates an Alchemy paymaster provider.
func NewAlchemyProvider(rpcURL, policyID string) *AlchemyProvider {
	return &AlchemyProvider{
		url:        rpcURL,
		policyID:   policyID,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (a *AlchemyProvider) Type() string { return "alchemy" }

func (a *AlchemyProvider) SponsorUserOp(ctx context.Context, req *SponsorRequest) (*SponsorResult, error) {
	opMap := userOpToMap(req.UserOp)

	params := []interface{}{
		map[string]interface{}{
			"policyId":      a.policyID,
			"entryPoint":    req.EntryPoint.Hex(),
			"userOperation": opMap,
		},
	}

	raw, err := a.call(ctx, "alchemy_requestGasAndPaymasterAndData", params)
	if err != nil {
		return nil, fmt.Errorf("alchemy sponsor: %w", err)
	}

	return parseSponsorResponse(raw)
}

func (a *AlchemyProvider) call(ctx context.Context, method string, params []interface{}) (json.RawMessage, error) {
	if params == nil {
		params = make([]interface{}, 0)
	}

	reqID := int(a.reqID.Add(1))
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

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(httpReq)
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
