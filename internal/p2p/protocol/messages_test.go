package protocol

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseStatus_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give ResponseStatus
		want bool
	}{
		{give: ResponseStatusOK, want: true},
		{give: ResponseStatusError, want: true},
		{give: ResponseStatusDenied, want: true},
		{give: ResponseStatusPaymentRequired, want: true},
		{give: ResponseStatus(""), want: false},
		{give: ResponseStatus("unknown"), want: false},
		{give: ResponseStatus("OK"), want: false},
	}

	for _, tt := range tests {
		t.Run(string(tt.give), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.give.Valid())
		})
	}
}

func TestRequestType_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give RequestType
		want string
	}{
		{give: RequestToolInvoke, want: "tool_invoke"},
		{give: RequestCapabilityQuery, want: "capability_query"},
		{give: RequestAgentCard, want: "agent_card"},
		{give: RequestPriceQuery, want: "price_query"},
		{give: RequestToolInvokePaid, want: "tool_invoke_paid"},
		{give: RequestContextShare, want: "context_share"},
		{give: RequestNegotiatePropose, want: "negotiate_propose"},
		{give: RequestNegotiateRespond, want: "negotiate_respond"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, string(tt.give))
		})
	}
}

func TestProtocolID(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "/lango/a2a/1.0.0", ProtocolID)
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give error
		want string
	}{
		{give: ErrMissingToolName, want: "missing toolName in payload"},
		{give: ErrAgentCardUnavailable, want: "agent card not available"},
		{give: ErrNoApprovalHandler, want: "no approval handler configured for remote tool invocation"},
		{give: ErrDeniedByOwner, want: "tool invocation denied by owner"},
		{give: ErrExecutorNotConfigured, want: "tool executor not configured"},
		{give: ErrInvalidSession, want: "invalid or expired session token"},
		{give: ErrInvalidPaymentAuth, want: "invalid payment authorization"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			assert.EqualError(t, tt.give, tt.want)
		})
	}
}

func TestRequest_JSON(t *testing.T) {
	t.Parallel()

	give := Request{
		Type:         RequestToolInvoke,
		SessionToken: "tok-123",
		RequestID:    "req-1",
		Payload:      map[string]interface{}{"toolName": "echo"},
	}

	data, err := json.Marshal(give)
	require.NoError(t, err)

	var got Request
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, give.Type, got.Type)
	assert.Equal(t, give.SessionToken, got.SessionToken)
	assert.Equal(t, give.RequestID, got.RequestID)
	assert.Equal(t, "echo", got.Payload["toolName"])
}

func TestRequest_JSON_OmitEmptyPayload(t *testing.T) {
	t.Parallel()

	give := Request{
		Type:         RequestCapabilityQuery,
		SessionToken: "tok-456",
		RequestID:    "req-2",
	}

	data, err := json.Marshal(give)
	require.NoError(t, err)

	// payload should be omitted from JSON when nil.
	assert.NotContains(t, string(data), "payload")
}

func TestResponse_JSON(t *testing.T) {
	t.Parallel()

	now := time.Now().Truncate(time.Second)
	give := Response{
		RequestID: "req-1",
		Status:    ResponseStatusOK,
		Result:    map[string]interface{}{"output": "hello"},
		Timestamp: now,
	}

	data, err := json.Marshal(give)
	require.NoError(t, err)

	var got Response
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, give.RequestID, got.RequestID)
	assert.Equal(t, ResponseStatusOK, got.Status)
	assert.Equal(t, "hello", got.Result["output"])
}

func TestResponse_JSON_WithAttestation(t *testing.T) {
	t.Parallel()

	give := Response{
		RequestID: "req-attest",
		Status:    ResponseStatusOK,
		Attestation: &AttestationData{
			Proof:        []byte{0x01, 0x02},
			PublicInputs: []byte{0x03, 0x04},
			CircuitID:    "cap-v1",
			Scheme:       "plonk",
		},
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(give)
	require.NoError(t, err)

	var got Response
	require.NoError(t, json.Unmarshal(data, &got))

	require.NotNil(t, got.Attestation)
	assert.Equal(t, "cap-v1", got.Attestation.CircuitID)
	assert.Equal(t, "plonk", got.Attestation.Scheme)
	assert.Equal(t, []byte{0x01, 0x02}, got.Attestation.Proof)
	assert.Equal(t, []byte{0x03, 0x04}, got.Attestation.PublicInputs)
}

func TestResponse_JSON_ErrorOmitEmpty(t *testing.T) {
	t.Parallel()

	give := Response{
		RequestID: "req-ok",
		Status:    ResponseStatusOK,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(give)
	require.NoError(t, err)

	// error, result, attestationProof, and attestation should be omitted.
	raw := string(data)
	assert.NotContains(t, raw, `"error"`)
	assert.NotContains(t, raw, `"result"`)
	assert.NotContains(t, raw, `"attestation"`)
	assert.NotContains(t, raw, `"attestationProof"`)
}

func TestToolInvokePayload_JSON(t *testing.T) {
	t.Parallel()

	give := ToolInvokePayload{
		ToolName: "web_search",
		Params:   map[string]interface{}{"query": "lango"},
	}

	data, err := json.Marshal(give)
	require.NoError(t, err)

	var got ToolInvokePayload
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, "web_search", got.ToolName)
	assert.Equal(t, "lango", got.Params["query"])
}

func TestCapabilityQueryPayload_JSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give     CapabilityQueryPayload
		wantJSON string
	}{
		{
			give:     CapabilityQueryPayload{Filter: "web_"},
			wantJSON: `{"filter":"web_"}`,
		},
		{
			give:     CapabilityQueryPayload{},
			wantJSON: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.wantJSON, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.give)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))
		})
	}
}

func TestPriceQuoteResult_JSON(t *testing.T) {
	t.Parallel()

	give := PriceQuoteResult{
		ToolName:     "translate",
		Price:        "1.50",
		Currency:     "USDC",
		USDCContract: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
		ChainID:      1,
		SellerAddr:   "0x1234",
		QuoteExpiry:  1700000000,
		IsFree:       false,
	}

	data, err := json.Marshal(give)
	require.NoError(t, err)

	var got PriceQuoteResult
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, give.ToolName, got.ToolName)
	assert.Equal(t, give.Price, got.Price)
	assert.Equal(t, give.Currency, got.Currency)
	assert.Equal(t, give.USDCContract, got.USDCContract)
	assert.Equal(t, give.ChainID, got.ChainID)
	assert.Equal(t, give.SellerAddr, got.SellerAddr)
	assert.Equal(t, give.QuoteExpiry, got.QuoteExpiry)
	assert.False(t, got.IsFree)
}

func TestPriceQuoteResult_JSON_Free(t *testing.T) {
	t.Parallel()

	give := PriceQuoteResult{
		ToolName: "free_tool",
		IsFree:   true,
	}

	data, err := json.Marshal(give)
	require.NoError(t, err)

	var got PriceQuoteResult
	require.NoError(t, json.Unmarshal(data, &got))

	assert.True(t, got.IsFree)
	assert.Equal(t, "free_tool", got.ToolName)
}

func TestPaidInvokePayload_JSON(t *testing.T) {
	t.Parallel()

	give := PaidInvokePayload{
		ToolName:    "premium_search",
		Params:      map[string]interface{}{"query": "test"},
		PaymentAuth: map[string]interface{}{"txHash": "0xabc"},
	}

	data, err := json.Marshal(give)
	require.NoError(t, err)

	var got PaidInvokePayload
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, "premium_search", got.ToolName)
	assert.Equal(t, "test", got.Params["query"])
	assert.Equal(t, "0xabc", got.PaymentAuth["txHash"])
}

func TestPaidInvokePayload_JSON_OmitEmptyPaymentAuth(t *testing.T) {
	t.Parallel()

	give := PaidInvokePayload{
		ToolName: "tool",
		Params:   map[string]interface{}{},
	}

	data, err := json.Marshal(give)
	require.NoError(t, err)

	assert.NotContains(t, string(data), "paymentAuth")
}

func TestContextSharePayload_JSON(t *testing.T) {
	t.Parallel()

	give := ContextSharePayload{
		TeamID:  "team-1",
		Context: map[string]interface{}{"key": "value"},
	}

	data, err := json.Marshal(give)
	require.NoError(t, err)

	var got ContextSharePayload
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, "team-1", got.TeamID)
	assert.Equal(t, "value", got.Context["key"])
}

func TestNegotiatePayload_JSON(t *testing.T) {
	t.Parallel()

	give := NegotiatePayload{
		SessionID: "sess-1",
		Action:    "propose",
		ToolName:  "translate",
		Price:     "2.00",
		Reason:    "initial offer",
	}

	data, err := json.Marshal(give)
	require.NoError(t, err)

	var got NegotiatePayload
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, "sess-1", got.SessionID)
	assert.Equal(t, "propose", got.Action)
	assert.Equal(t, "translate", got.ToolName)
	assert.Equal(t, "2.00", got.Price)
	assert.Equal(t, "initial offer", got.Reason)
}

func TestNegotiatePayload_JSON_OmitEmpty(t *testing.T) {
	t.Parallel()

	give := NegotiatePayload{
		Action: "accept",
	}

	data, err := json.Marshal(give)
	require.NoError(t, err)

	raw := string(data)
	assert.NotContains(t, raw, "sessionId")
	assert.NotContains(t, raw, "toolName")
	assert.NotContains(t, raw, "price")
	assert.NotContains(t, raw, "reason")
	assert.Contains(t, raw, `"action":"accept"`)
}
