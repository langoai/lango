package protocol

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/network"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/firewall"
	"github.com/langoai/lango/internal/p2p/handshake"
)

// ToolExecutor executes a tool by name with the given parameters.
// Uses the callback pattern to avoid import cycles with the agent package.
type ToolExecutor func(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error)

// ToolApprovalFunc asks the local owner for approval before executing a remote
// tool invocation. Returns true if approved, false if denied.
// Uses the callback pattern to avoid import cycles with the approval package.
type ToolApprovalFunc func(ctx context.Context, peerDID, toolName string, params map[string]interface{}) (bool, error)

// SecurityEventTracker records tool execution outcomes for security monitoring.
// Uses the callback pattern to avoid import cycles with the handshake package.
type SecurityEventTracker interface {
	RecordToolFailure(peerDID string)
	RecordToolSuccess(peerDID string)
}

// CardProvider returns the local agent card as a map.
type CardProvider func() map[string]interface{}

// PayGateChecker checks payment for a tool invocation.
type PayGateChecker interface {
	Check(peerDID, toolName string, payload map[string]interface{}) (PayGateResult, error)
}

// PayGate status values returned by PayGateChecker.Check.
const (
	payGateStatusFree            = "free"
	payGateStatusVerified        = "verified"
	payGateStatusPaymentRequired = "payment_required"
	payGateStatusInvalid         = "invalid"
	payGateStatusPostPayApproved = "postpay_approved"
)

// PayGateResult represents the payment check outcome.
type PayGateResult struct {
	Status       string                 // payGateStatusFree, payGateStatusVerified, payGateStatusPaymentRequired, payGateStatusInvalid, payGateStatusPostPayApproved
	Auth         interface{}            // the verified authorization (opaque to handler)
	PriceQuote   map[string]interface{} // price quote when payment required
	SettlementID string                 // deferred settlement ID for post-pay
}

// NegotiateHandler processes negotiation protocol messages.
type NegotiateHandler func(ctx context.Context, peerDID string, payload NegotiatePayload) (map[string]interface{}, error)

// TeamHandler processes team-related protocol messages.
type TeamHandler func(ctx context.Context, peerDID string, reqType RequestType, payload map[string]interface{}) (map[string]interface{}, error)

// Handler processes A2A-over-P2P messages on libp2p streams.
type Handler struct {
	sessions       *handshake.SessionStore
	firewall       *firewall.Firewall
	executor       ToolExecutor
	sandboxExec    ToolExecutor
	cardFn         CardProvider
	payGate        PayGateChecker
	approvalFn     ToolApprovalFunc
	securityEvents SecurityEventTracker
	eventBus       *eventbus.Bus
	negotiator      NegotiateHandler
	teamHandler     TeamHandler
	ontologyHandler OntologyHandler
	localDID        string
	logger          *zap.SugaredLogger
}

// HandlerConfig configures the protocol handler.
type HandlerConfig struct {
	Sessions *handshake.SessionStore
	Firewall *firewall.Firewall
	Executor ToolExecutor
	CardFn   CardProvider
	LocalDID string
	Logger   *zap.SugaredLogger
}

// NewHandler creates a new A2A-over-P2P protocol handler.
func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		sessions: cfg.Sessions,
		firewall: cfg.Firewall,
		executor: cfg.Executor,
		cardFn:   cfg.CardFn,
		localDID: cfg.LocalDID,
		logger:   cfg.Logger,
	}
}

// SetExecutor sets the tool executor callback.
func (h *Handler) SetExecutor(exec ToolExecutor) {
	h.executor = exec
}

// SetPayGate sets the payment gate checker for paid tool invocations.
func (h *Handler) SetPayGate(gate PayGateChecker) {
	h.payGate = gate
}

// SetApprovalFunc sets the owner approval callback for remote tool invocations.
func (h *Handler) SetApprovalFunc(fn ToolApprovalFunc) {
	h.approvalFn = fn
}

// SetSandboxExecutor sets an isolated executor for remote tool invocations.
// When set, tool calls from remote peers use this executor instead of the
// default in-process executor, preventing access to parent process memory.
func (h *Handler) SetSandboxExecutor(exec ToolExecutor) {
	h.sandboxExec = exec
}

// SetSecurityEvents sets the security event tracker for recording tool
// execution outcomes and triggering auto-invalidation on repeated failures.
func (h *Handler) SetSecurityEvents(tracker SecurityEventTracker) {
	h.securityEvents = tracker
}

// SetEventBus sets the event bus for post-execution event publishing.
func (h *Handler) SetEventBus(bus *eventbus.Bus) {
	h.eventBus = bus
}

// SetNegotiator sets the handler for negotiation protocol messages.
func (h *Handler) SetNegotiator(fn NegotiateHandler) {
	h.negotiator = fn
}

// SetTeamHandler sets the handler for team protocol messages.
func (h *Handler) SetTeamHandler(fn TeamHandler) {
	h.teamHandler = fn
}

// SetOntologyHandler sets the handler for ontology schema exchange messages.
func (h *Handler) SetOntologyHandler(oh OntologyHandler) {
	h.ontologyHandler = oh
}

// StreamHandler returns a libp2p stream handler for incoming A2A messages.
func (h *Handler) StreamHandler() network.StreamHandler {
	return func(s network.Stream) {
		defer s.Close()

		ctx := context.Background()

		var req Request
		if err := json.NewDecoder(s).Decode(&req); err != nil {
			h.sendError(s, "", fmt.Sprintf("decode request: %v", err))
			return
		}

		resp := h.handleRequest(ctx, s, &req)
		if err := json.NewEncoder(s).Encode(resp); err != nil {
			h.logger.Warnw("encode response", "error", err)
		}
	}
}

// handleRequest processes a single A2A request.
func (h *Handler) handleRequest(ctx context.Context, s network.Stream, req *Request) *Response {
	// Validate session token.
	peerDID := h.resolvePeerDID(s, req.SessionToken)
	if peerDID == "" {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusDenied,
			Error:     ErrInvalidSession.Error(),
			Timestamp: time.Now(),
		}
	}

	switch req.Type {
	case RequestAgentCard:
		return h.handleAgentCard(req)
	case RequestCapabilityQuery:
		return h.handleCapabilityQuery(req, peerDID)
	case RequestToolInvoke:
		return h.handleToolInvoke(ctx, req, peerDID)
	case RequestPriceQuery:
		return h.handlePriceQuery(ctx, req, peerDID)
	case RequestToolInvokePaid:
		return h.handleToolInvokePaid(ctx, req, peerDID)
	case RequestNegotiatePropose, RequestNegotiateRespond:
		return h.handleNegotiate(ctx, req, peerDID)
	case RequestTeamInvite, RequestTeamAccept, RequestTeamTask, RequestTeamResult, RequestTeamDisband:
		return h.handleTeamMessage(ctx, req, peerDID)
	case RequestSchemaQuery:
		return h.handleSchemaQuery(ctx, req, peerDID)
	case RequestSchemaPropose:
		return h.handleSchemaPropose(ctx, req, peerDID)
	default:
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     fmt.Sprintf("unknown request type: %s", req.Type),
			Timestamp: time.Now(),
		}
	}
}

// handleAgentCard returns the local agent card.
func (h *Handler) handleAgentCard(req *Request) *Response {
	if h.cardFn == nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     ErrAgentCardUnavailable.Error(),
			Timestamp: time.Now(),
		}
	}

	return &Response{
		RequestID: req.RequestID,
		Status:    ResponseStatusOK,
		Result:    h.cardFn(),
		Timestamp: time.Now(),
	}
}

// handleCapabilityQuery returns available capabilities.
func (h *Handler) handleCapabilityQuery(req *Request, peerDID string) *Response {
	// Return the agent card with capabilities.
	if h.cardFn != nil {
		card := h.cardFn()
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusOK,
			Result:    card,
			Timestamp: time.Now(),
		}
	}

	return &Response{
		RequestID: req.RequestID,
		Status:    ResponseStatusOK,
		Result:    map[string]interface{}{"capabilities": []string{}},
		Timestamp: time.Now(),
	}
}

// handleToolInvoke executes a tool and returns the result.
func (h *Handler) handleToolInvoke(ctx context.Context, req *Request, peerDID string) *Response {
	toolName, _ := req.Payload["toolName"].(string)
	if toolName == "" {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     ErrMissingToolName.Error(),
			Timestamp: time.Now(),
		}
	}

	// Firewall check.
	if h.firewall != nil {
		if err := h.firewall.FilterQuery(ctx, peerDID, toolName); err != nil {
			return &Response{
				RequestID: req.RequestID,
				Status:    ResponseStatusDenied,
				Error:     err.Error(),
				Timestamp: time.Now(),
			}
		}
	}

	// Owner approval check (default-deny when no approval handler is configured).
	params, _ := req.Payload["params"].(map[string]interface{})
	if params == nil {
		params = map[string]interface{}{}
	}

	if h.approvalFn == nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusDenied,
			Error:     ErrNoApprovalHandler.Error(),
			Timestamp: time.Now(),
		}
	}
	approved, err := h.approvalFn(ctx, peerDID, toolName, params)
	if err != nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     fmt.Sprintf("approval check: %v", err),
			Timestamp: time.Now(),
		}
	}
	if !approved {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusDenied,
			Error:     ErrDeniedByOwner.Error(),
			Timestamp: time.Now(),
		}
	}

	// Execute tool (prefer sandbox executor for process isolation).
	exec := h.executor
	if h.sandboxExec != nil {
		exec = h.sandboxExec
	}
	result, err := exec(ctx, toolName, params)
	if err != nil {
		if h.securityEvents != nil {
			h.securityEvents.RecordToolFailure(peerDID)
		}
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
	}

	if h.securityEvents != nil {
		h.securityEvents.RecordToolSuccess(peerDID)
	}

	// Sanitize response through firewall.
	if h.firewall != nil {
		result = h.firewall.SanitizeResponse(result)
	}

	// Generate ZK attestation if available.
	resp := &Response{
		RequestID: req.RequestID,
		Status:    ResponseStatusOK,
		Result:    result,
		Timestamp: time.Now(),
	}
	if h.firewall != nil {
		resultBytes, _ := json.Marshal(result)
		hash := sha256.Sum256(resultBytes)
		didHash := sha256.Sum256([]byte(h.localDID))
		ar, _ := h.firewall.AttestResponse(hash[:], didHash[:])
		if ar != nil {
			resp.Attestation = &AttestationData{
				Proof:        ar.Proof,
				PublicInputs: ar.PublicInputs,
				CircuitID:    ar.CircuitID,
				Scheme:       ar.Scheme,
			}
			resp.AttestationProof = ar.Proof // backward compat
		}
	}

	return resp
}

// handlePriceQuery returns pricing information for a tool.
func (h *Handler) handlePriceQuery(ctx context.Context, req *Request, peerDID string) *Response {
	toolName, _ := req.Payload["toolName"].(string)
	if toolName == "" {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     ErrMissingToolName.Error(),
			Timestamp: time.Now(),
		}
	}

	if h.payGate == nil {
		// No payment gate configured — everything is free.
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusOK,
			Result: map[string]interface{}{
				"toolName": toolName,
				"isFree":   true,
			},
			Timestamp: time.Now(),
		}
	}

	result, err := h.payGate.Check(peerDID, toolName, nil)
	if err != nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     fmt.Sprintf("price query %s: %v", toolName, err),
			Timestamp: time.Now(),
		}
	}

	if result.Status == payGateStatusFree {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusOK,
			Result: map[string]interface{}{
				"toolName": toolName,
				"isFree":   true,
			},
			Timestamp: time.Now(),
		}
	}

	return &Response{
		RequestID: req.RequestID,
		Status:    ResponseStatusOK,
		Result:    result.PriceQuote,
		Timestamp: time.Now(),
	}
}

// handleToolInvokePaid executes a paid tool invocation with payment verification.
func (h *Handler) handleToolInvokePaid(ctx context.Context, req *Request, peerDID string) *Response {
	toolName, _ := req.Payload["toolName"].(string)
	if toolName == "" {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     ErrMissingToolName.Error(),
			Timestamp: time.Now(),
		}
	}

	// 1. Firewall ACL check.
	if h.firewall != nil {
		if err := h.firewall.FilterQuery(ctx, peerDID, toolName); err != nil {
			return &Response{
				RequestID: req.RequestID,
				Status:    ResponseStatusDenied,
				Error:     err.Error(),
				Timestamp: time.Now(),
			}
		}
	}

	// 2. Payment gate check.
	var verifiedAuth interface{}
	var settlementID string
	if h.payGate != nil {
		pgResult, err := h.payGate.Check(peerDID, toolName, req.Payload)
		if err != nil {
			return &Response{
				RequestID: req.RequestID,
				Status:    ResponseStatusError,
				Error:     fmt.Sprintf("payment check %s: %v", toolName, err),
				Timestamp: time.Now(),
			}
		}

		switch pgResult.Status {
		case payGateStatusPaymentRequired:
			return &Response{
				RequestID: req.RequestID,
				Status:    ResponseStatusPaymentRequired,
				Result:    pgResult.PriceQuote,
				Timestamp: time.Now(),
			}
		case payGateStatusInvalid:
			return &Response{
				RequestID: req.RequestID,
				Status:    ResponseStatusError,
				Error:     ErrInvalidPaymentAuth.Error(),
				Timestamp: time.Now(),
			}
		case payGateStatusPostPayApproved:
			settlementID = pgResult.SettlementID
			// Continue to execution — payment deferred.
		case payGateStatusVerified:
			verifiedAuth = pgResult.Auth
			// Continue to execution.
		case payGateStatusFree:
			// Continue to execution.
		}
	}

	// 3. Owner approval check (default-deny when no approval handler is configured).
	params, _ := req.Payload["params"].(map[string]interface{})
	if params == nil {
		params = map[string]interface{}{}
	}

	if h.approvalFn == nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusDenied,
			Error:     ErrNoApprovalHandler.Error(),
			Timestamp: time.Now(),
		}
	}
	approved, err := h.approvalFn(ctx, peerDID, toolName, params)
	if err != nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     fmt.Sprintf("approval check: %v", err),
			Timestamp: time.Now(),
		}
	}
	if !approved {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusDenied,
			Error:     ErrDeniedByOwner.Error(),
			Timestamp: time.Now(),
		}
	}

	// 4. Execute tool (prefer sandbox executor for process isolation).
	paidExec := h.executor
	if h.sandboxExec != nil {
		paidExec = h.sandboxExec
	}
	if paidExec == nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     ErrExecutorNotConfigured.Error(),
			Timestamp: time.Now(),
		}
	}

	result, err := paidExec(ctx, toolName, params)
	if err != nil {
		if h.securityEvents != nil {
			h.securityEvents.RecordToolFailure(peerDID)
		}
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
	}

	if h.securityEvents != nil {
		h.securityEvents.RecordToolSuccess(peerDID)
	}

	// 4b. Publish settlement event for on-chain processing.
	if h.eventBus != nil && (verifiedAuth != nil || settlementID != "") {
		h.eventBus.Publish(eventbus.ToolExecutionPaidEvent{
			PeerDID:      peerDID,
			ToolName:     toolName,
			Auth:         verifiedAuth,
			SettlementID: settlementID,
		})
	}

	// 5. Sanitize response through firewall.
	if h.firewall != nil {
		result = h.firewall.SanitizeResponse(result)
	}

	// 6. ZK attestation.
	paidResp := &Response{
		RequestID: req.RequestID,
		Status:    ResponseStatusOK,
		Result:    result,
		Timestamp: time.Now(),
	}
	if h.firewall != nil {
		resultBytes, _ := json.Marshal(result)
		hash := sha256.Sum256(resultBytes)
		didHash := sha256.Sum256([]byte(h.localDID))
		ar, _ := h.firewall.AttestResponse(hash[:], didHash[:])
		if ar != nil {
			paidResp.Attestation = &AttestationData{
				Proof:        ar.Proof,
				PublicInputs: ar.PublicInputs,
				CircuitID:    ar.CircuitID,
				Scheme:       ar.Scheme,
			}
			paidResp.AttestationProof = ar.Proof // backward compat
		}
	}

	return paidResp
}

// resolvePeerDID validates the session token and returns the peer DID.
func (h *Handler) resolvePeerDID(s network.Stream, token string) string {
	if h.sessions == nil {
		return ""
	}

	// Check all active sessions for matching token.
	for _, sess := range h.sessions.ActiveSessions() {
		if h.sessions.Validate(sess.PeerDID, token) {
			return sess.PeerDID
		}
	}

	return ""
}

// sendError sends a quick error response on a stream.
func (h *Handler) sendError(s network.Stream, reqID, msg string) {
	resp := Response{
		RequestID: reqID,
		Status:    ResponseStatusError,
		Error:     msg,
		Timestamp: time.Now(),
	}
	_ = json.NewEncoder(s).Encode(resp)
}

// handleNegotiate processes negotiation protocol messages.
func (h *Handler) handleNegotiate(ctx context.Context, req *Request, peerDID string) *Response {
	if h.negotiator == nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     "negotiation not configured",
			Timestamp: time.Now(),
		}
	}

	var payload NegotiatePayload
	if raw, err := json.Marshal(req.Payload); err == nil {
		_ = json.Unmarshal(raw, &payload)
	}

	result, err := h.negotiator(ctx, peerDID, payload)
	if err != nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
	}

	return &Response{
		RequestID: req.RequestID,
		Status:    ResponseStatusOK,
		Result:    result,
		Timestamp: time.Now(),
	}
}

// handleTeamMessage routes team protocol messages to the team handler.
func (h *Handler) handleTeamMessage(ctx context.Context, req *Request, peerDID string) *Response {
	if h.teamHandler == nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     "team handler not configured",
			Timestamp: time.Now(),
		}
	}

	result, err := h.teamHandler(ctx, peerDID, req.Type, req.Payload)
	if err != nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
	}

	return &Response{
		RequestID: req.RequestID,
		Status:    ResponseStatusOK,
		Result:    result,
		Timestamp: time.Now(),
	}
}

// SendRequest sends an A2A request to a remote peer over a stream.
func SendRequest(ctx context.Context, s network.Stream, reqType RequestType, token string, payload map[string]interface{}) (*Response, error) {
	req := Request{
		Type:         reqType,
		SessionToken: token,
		RequestID:    uuid.New().String(),
		Payload:      payload,
	}

	if err := json.NewEncoder(s).Encode(req); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	var resp Response
	if err := json.NewDecoder(s).Decode(&resp); err != nil {
		return nil, fmt.Errorf("receive response: %w", err)
	}

	return &resp, nil
}

// handleSchemaQuery processes a schema_query request.
func (h *Handler) handleSchemaQuery(ctx context.Context, req *Request, peerDID string) *Response {
	if h.ontologyHandler == nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     "ontology handler not configured",
			Timestamp: time.Now(),
		}
	}

	raw, err := json.Marshal(req.Payload)
	if err != nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     fmt.Sprintf("marshal schema query payload: %v", err),
			Timestamp: time.Now(),
		}
	}

	var sqReq SchemaQueryRequest
	if err := json.Unmarshal(raw, &sqReq); err != nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     fmt.Sprintf("decode schema query: %v", err),
			Timestamp: time.Now(),
		}
	}

	sqResp, err := h.ontologyHandler.HandleSchemaQuery(ctx, peerDID, sqReq)
	if err != nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
	}

	result := map[string]interface{}{}
	respBytes, _ := json.Marshal(sqResp)
	_ = json.Unmarshal(respBytes, &result)

	return &Response{
		RequestID: req.RequestID,
		Status:    ResponseStatusOK,
		Result:    result,
		Timestamp: time.Now(),
	}
}

// handleSchemaPropose processes a schema_propose request.
func (h *Handler) handleSchemaPropose(ctx context.Context, req *Request, peerDID string) *Response {
	if h.ontologyHandler == nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     "ontology handler not configured",
			Timestamp: time.Now(),
		}
	}

	raw, err := json.Marshal(req.Payload)
	if err != nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     fmt.Sprintf("marshal schema propose payload: %v", err),
			Timestamp: time.Now(),
		}
	}

	var spReq SchemaProposeRequest
	if err := json.Unmarshal(raw, &spReq); err != nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     fmt.Sprintf("decode schema propose: %v", err),
			Timestamp: time.Now(),
		}
	}

	spResp, err := h.ontologyHandler.HandleSchemaPropose(ctx, peerDID, spReq)
	if err != nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    ResponseStatusError,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
	}

	result := map[string]interface{}{}
	respBytes, _ := json.Marshal(spResp)
	_ = json.Unmarshal(respBytes, &result)

	return &Response{
		RequestID: req.RequestID,
		Status:    ResponseStatusOK,
		Result:    result,
		Timestamp: time.Now(),
	}
}
