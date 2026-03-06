// Package protocol implements the A2A-over-P2P message exchange protocol.
package protocol

import (
	"errors"
	"time"
)

// ProtocolID is the libp2p protocol identifier for A2A messages.
const ProtocolID = "/lango/a2a/1.0.0"

// RequestType identifies the type of A2A request.
type RequestType string

const (
	// RequestToolInvoke invokes a tool on the remote agent.
	RequestToolInvoke RequestType = "tool_invoke"

	// RequestCapabilityQuery queries the capabilities of the remote agent.
	RequestCapabilityQuery RequestType = "capability_query"

	// RequestAgentCard requests the agent card of the remote agent.
	RequestAgentCard RequestType = "agent_card"

	// RequestPriceQuery queries the pricing for a tool on the remote agent.
	RequestPriceQuery RequestType = "price_query"

	// RequestToolInvokePaid invokes a paid tool on the remote agent.
	RequestToolInvokePaid RequestType = "tool_invoke_paid"

	// RequestContextShare shares scoped context with a team member.
	RequestContextShare RequestType = "context_share"

	// RequestNegotiatePropose proposes a negotiation session.
	RequestNegotiatePropose RequestType = "negotiate_propose"

	// RequestNegotiateRespond responds to a negotiation (counter/accept/reject).
	RequestNegotiateRespond RequestType = "negotiate_respond"
)

// ResponseStatus identifies the status of an A2A response.
type ResponseStatus string

const (
	// ResponseStatusOK indicates a successful response.
	ResponseStatusOK ResponseStatus = "ok"

	// ResponseStatusError indicates an error response.
	ResponseStatusError ResponseStatus = "error"

	// ResponseStatusDenied indicates the request was denied.
	ResponseStatusDenied ResponseStatus = "denied"

	// ResponseStatusPaymentRequired indicates payment is needed.
	ResponseStatusPaymentRequired ResponseStatus = "payment_required"
)

// Valid reports whether s is a known response status.
func (s ResponseStatus) Valid() bool {
	switch s {
	case ResponseStatusOK, ResponseStatusError, ResponseStatusDenied, ResponseStatusPaymentRequired:
		return true
	}
	return false
}

// Sentinel errors for protocol-level failures.
var (
	ErrMissingToolName       = errors.New("missing toolName in payload")
	ErrAgentCardUnavailable  = errors.New("agent card not available")
	ErrNoApprovalHandler     = errors.New("no approval handler configured for remote tool invocation")
	ErrDeniedByOwner         = errors.New("tool invocation denied by owner")
	ErrExecutorNotConfigured = errors.New("tool executor not configured")
	ErrInvalidSession        = errors.New("invalid or expired session token")
	ErrInvalidPaymentAuth    = errors.New("invalid payment authorization")
)

// Request is a P2P A2A request message.
type Request struct {
	Type         RequestType            `json:"type"`
	SessionToken string                 `json:"sessionToken"`
	RequestID    string                 `json:"requestId"`
	Payload      map[string]interface{} `json:"payload,omitempty"`
}

// AttestationData holds structured ZK attestation proof with metadata.
type AttestationData struct {
	Proof        []byte `json:"proof"`
	PublicInputs []byte `json:"publicInputs"`
	CircuitID    string `json:"circuitId"`
	Scheme       string `json:"scheme"`
}

// Response is a P2P A2A response message.
type Response struct {
	RequestID        string                 `json:"requestId"`
	Status           ResponseStatus         `json:"status"` // ResponseStatusOK, ResponseStatusError, ResponseStatusDenied
	Result           map[string]interface{} `json:"result,omitempty"`
	Error            string                 `json:"error,omitempty"`
	AttestationProof []byte                 `json:"attestationProof,omitempty"` // Deprecated: use Attestation
	Attestation      *AttestationData       `json:"attestation,omitempty"`
	Timestamp        time.Time              `json:"timestamp"`
}

// ToolInvokePayload is the payload for a tool invocation request.
type ToolInvokePayload struct {
	ToolName string                 `json:"toolName"`
	Params   map[string]interface{} `json:"params"`
}

// CapabilityQueryPayload is the payload for a capability query.
type CapabilityQueryPayload struct {
	Filter string `json:"filter,omitempty"` // optional tool name prefix filter
}

// PriceQuoteResult is returned when querying tool pricing.
type PriceQuoteResult struct {
	ToolName     string `json:"toolName"`
	Price        string `json:"price"`
	Currency     string `json:"currency"`
	USDCContract string `json:"usdcContract"`
	ChainID      int64  `json:"chainId"`
	SellerAddr   string `json:"sellerAddr"`
	QuoteExpiry  int64  `json:"quoteExpiry"`
	IsFree       bool   `json:"isFree"`
}

// PaidInvokePayload is the payload for a paid tool invocation.
type PaidInvokePayload struct {
	ToolName    string                 `json:"toolName"`
	Params      map[string]interface{} `json:"params"`
	PaymentAuth map[string]interface{} `json:"paymentAuth,omitempty"`
}

// ContextSharePayload is the payload for sharing scoped context with a team member.
type ContextSharePayload struct {
	TeamID  string                 `json:"teamId"`
	Context map[string]interface{} `json:"context"`
}

// NegotiatePayload is the payload for negotiation messages.
type NegotiatePayload struct {
	SessionID string `json:"sessionId,omitempty"` // empty for initial proposal
	Action    string `json:"action"`              // "propose", "counter", "accept", "reject"
	ToolName  string `json:"toolName,omitempty"`
	Price     string `json:"price,omitempty"`     // USDC amount as string
	Reason    string `json:"reason,omitempty"`
}
