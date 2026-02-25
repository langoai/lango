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

// PayGateResult represents the payment check outcome.
type PayGateResult struct {
	Status     string                 // "free", "verified", "payment_required", "invalid"
	Auth       interface{}            // the verified authorization (opaque to handler)
	PriceQuote map[string]interface{} // price quote when payment required
}

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
	localDID       string
	logger         *zap.SugaredLogger
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
			Status:    "denied",
			Error:     "invalid or expired session token",
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
	default:
		return &Response{
			RequestID: req.RequestID,
			Status:    "error",
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
			Status:    "error",
			Error:     "agent card not available",
			Timestamp: time.Now(),
		}
	}

	return &Response{
		RequestID: req.RequestID,
		Status:    "ok",
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
			Status:    "ok",
			Result:    card,
			Timestamp: time.Now(),
		}
	}

	return &Response{
		RequestID: req.RequestID,
		Status:    "ok",
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
			Status:    "error",
			Error:     "missing toolName in payload",
			Timestamp: time.Now(),
		}
	}

	// Firewall check.
	if h.firewall != nil {
		if err := h.firewall.FilterQuery(ctx, peerDID, toolName); err != nil {
			return &Response{
				RequestID: req.RequestID,
				Status:    "denied",
				Error:     err.Error(),
				Timestamp: time.Now(),
			}
		}
	}

	// Owner approval check.
	params, _ := req.Payload["params"].(map[string]interface{})
	if params == nil {
		params = map[string]interface{}{}
	}

	if h.approvalFn != nil {
		approved, err := h.approvalFn(ctx, peerDID, toolName, params)
		if err != nil {
			return &Response{
				RequestID: req.RequestID,
				Status:    "error",
				Error:     fmt.Sprintf("approval check: %v", err),
				Timestamp: time.Now(),
			}
		}
		if !approved {
			return &Response{
				RequestID: req.RequestID,
				Status:    "denied",
				Error:     "tool invocation denied by owner",
				Timestamp: time.Now(),
			}
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
			Status:    "error",
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
		Status:    "ok",
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
			Status:    "error",
			Error:     "missing toolName in payload",
			Timestamp: time.Now(),
		}
	}

	if h.payGate == nil {
		// No payment gate configured — everything is free.
		return &Response{
			RequestID: req.RequestID,
			Status:    "ok",
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
			Status:    "error",
			Error:     fmt.Sprintf("price query %s: %v", toolName, err),
			Timestamp: time.Now(),
		}
	}

	if result.Status == "free" {
		return &Response{
			RequestID: req.RequestID,
			Status:    "ok",
			Result: map[string]interface{}{
				"toolName": toolName,
				"isFree":   true,
			},
			Timestamp: time.Now(),
		}
	}

	return &Response{
		RequestID: req.RequestID,
		Status:    "ok",
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
			Status:    "error",
			Error:     "missing toolName in payload",
			Timestamp: time.Now(),
		}
	}

	// 1. Firewall ACL check.
	if h.firewall != nil {
		if err := h.firewall.FilterQuery(ctx, peerDID, toolName); err != nil {
			return &Response{
				RequestID: req.RequestID,
				Status:    "denied",
				Error:     err.Error(),
				Timestamp: time.Now(),
			}
		}
	}

	// 2. Payment gate check.
	if h.payGate != nil {
		result, err := h.payGate.Check(peerDID, toolName, req.Payload)
		if err != nil {
			return &Response{
				RequestID: req.RequestID,
				Status:    "error",
				Error:     fmt.Sprintf("payment check %s: %v", toolName, err),
				Timestamp: time.Now(),
			}
		}

		switch result.Status {
		case "payment_required":
			return &Response{
				RequestID: req.RequestID,
				Status:    StatusPaymentRequired,
				Result:    result.PriceQuote,
				Timestamp: time.Now(),
			}
		case "invalid":
			return &Response{
				RequestID: req.RequestID,
				Status:    "error",
				Error:     "invalid payment authorization",
				Timestamp: time.Now(),
			}
		case "verified", "free":
			// Continue to execution.
		}
	}

	// 3. Owner approval check.
	params, _ := req.Payload["params"].(map[string]interface{})
	if params == nil {
		params = map[string]interface{}{}
	}

	if h.approvalFn != nil {
		approved, err := h.approvalFn(ctx, peerDID, toolName, params)
		if err != nil {
			return &Response{
				RequestID: req.RequestID,
				Status:    "error",
				Error:     fmt.Sprintf("approval check: %v", err),
				Timestamp: time.Now(),
			}
		}
		if !approved {
			return &Response{
				RequestID: req.RequestID,
				Status:    "denied",
				Error:     "tool invocation denied by owner",
				Timestamp: time.Now(),
			}
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
			Status:    "error",
			Error:     "tool executor not configured",
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
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
	}

	if h.securityEvents != nil {
		h.securityEvents.RecordToolSuccess(peerDID)
	}

	// 5. Sanitize response through firewall.
	if h.firewall != nil {
		result = h.firewall.SanitizeResponse(result)
	}

	// 6. ZK attestation.
	paidResp := &Response{
		RequestID: req.RequestID,
		Status:    "ok",
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
		Status:    "error",
		Error:     msg,
		Timestamp: time.Now(),
	}
	_ = json.NewEncoder(s).Encode(resp)
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
