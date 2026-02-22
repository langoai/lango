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

// CardProvider returns the local agent card as a map.
type CardProvider func() map[string]interface{}

// Handler processes A2A-over-P2P messages on libp2p streams.
type Handler struct {
	sessions *handshake.SessionStore
	firewall *firewall.Firewall
	executor ToolExecutor
	cardFn   CardProvider
	localDID string
	logger   *zap.SugaredLogger
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
		if err := h.firewall.FilterQuery(peerDID, toolName); err != nil {
			return &Response{
				RequestID: req.RequestID,
				Status:    "denied",
				Error:     err.Error(),
				Timestamp: time.Now(),
			}
		}
	}

	// Execute tool.
	params, _ := req.Payload["params"].(map[string]interface{})
	if params == nil {
		params = map[string]interface{}{}
	}

	result, err := h.executor(ctx, toolName, params)
	if err != nil {
		return &Response{
			RequestID: req.RequestID,
			Status:    "error",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
	}

	// Sanitize response through firewall.
	if h.firewall != nil {
		result = h.firewall.SanitizeResponse(result)
	}

	// Generate ZK attestation if available.
	var attestation []byte
	if h.firewall != nil {
		resultBytes, _ := json.Marshal(result)
		hash := sha256.Sum256(resultBytes)
		didHash := sha256.Sum256([]byte(h.localDID))
		attestation, _ = h.firewall.AttestResponse(hash[:], didHash[:])
	}

	return &Response{
		RequestID:        req.RequestID,
		Status:           "ok",
		Result:           result,
		AttestationProof: attestation,
		Timestamp:        time.Now(),
	}
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
