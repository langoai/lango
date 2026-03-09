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

// PimlicoProvider implements PaymasterProvider using Pimlico's Paymaster API.
type PimlicoProvider struct {
	url        string
	policyID   string
	httpClient *http.Client
	reqID      atomic.Int64
}

// NewPimlicoProvider creates a Pimlico paymaster provider.
func NewPimlicoProvider(rpcURL, policyID string) *PimlicoProvider {
	return &PimlicoProvider{
		url:        rpcURL,
		policyID:   policyID,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (p *PimlicoProvider) Type() string { return "pimlico" }

func (p *PimlicoProvider) SponsorUserOp(ctx context.Context, req *SponsorRequest) (*SponsorResult, error) {
	opMap := userOpToMap(req.UserOp)

	params := []interface{}{
		opMap,
		req.EntryPoint.Hex(),
	}

	// Add sponsorship policy context if configured.
	if p.policyID != "" {
		params = append(params, map[string]interface{}{
			"sponsorshipPolicyId": p.policyID,
		})
	}

	raw, err := p.call(ctx, "pm_sponsorUserOperation", params)
	if err != nil {
		return nil, fmt.Errorf("pimlico sponsor: %w", err)
	}

	return parseSponsorResponse(raw)
}

func (p *PimlicoProvider) call(ctx context.Context, method string, params []interface{}) (json.RawMessage, error) {
	if params == nil {
		params = make([]interface{}, 0)
	}

	reqID := int(p.reqID.Add(1))
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

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
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
