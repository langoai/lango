package provenanceproto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"go.uber.org/zap"
)

// SessionValidator validates a session token and returns the peer DID.
type SessionValidator func(token string) (string, bool)

// BundleImporter verifies and stores a provenance bundle payload.
type BundleImporter func(ctx context.Context, peerDID string, data []byte) error

// Handler handles provenance protocol streams.
type Handler struct {
	validator SessionValidator
	importer  BundleImporter
	maxBundle int64
	logger    *zap.Logger
}

// HandlerConfig configures a provenance protocol handler.
type HandlerConfig struct {
	Validator     SessionValidator
	Importer      BundleImporter
	MaxBundleSize int64
	Logger        *zap.Logger
}

// NewHandler creates a new provenance protocol handler.
func NewHandler(cfg HandlerConfig) *Handler {
	if cfg.MaxBundleSize <= 0 {
		cfg.MaxBundleSize = 5 * 1024 * 1024
	}
	return &Handler{
		validator: cfg.Validator,
		importer:  cfg.Importer,
		maxBundle: cfg.MaxBundleSize,
		logger:    cfg.Logger,
	}
}

// StreamHandler returns the libp2p stream handler function.
func (h *Handler) StreamHandler() network.StreamHandler {
	return func(s network.Stream) {
		defer s.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		var req Request
		if err := json.NewDecoder(io.LimitReader(s, h.maxBundle+4096)).Decode(&req); err != nil {
			h.writeError(s, "decode request: "+err.Error())
			return
		}

		var peerDID string
		if h.validator != nil {
			var ok bool
			peerDID, ok = h.validator(req.Token)
			if !ok {
				h.writeError(s, "invalid or expired session token")
				return
			}
		}

		switch req.Type {
		case RequestPushBundle:
			h.handlePushBundle(ctx, s, peerDID, req)
		default:
			h.writeError(s, fmt.Sprintf("unknown request type: %s", req.Type))
		}
	}
}

func (h *Handler) handlePushBundle(ctx context.Context, s network.Stream, peerDID string, req Request) {
	var payload PushBundlePayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		h.writeError(s, "unmarshal bundle payload: "+err.Error())
		return
	}
	if int64(len(payload.Bundle)) > h.maxBundle {
		h.writeError(s, fmt.Sprintf("bundle too large: %d > %d", len(payload.Bundle), h.maxBundle))
		return
	}
	if h.importer == nil {
		h.writeError(s, "bundle importer not configured")
		return
	}
	if err := h.importer(ctx, peerDID, payload.Bundle); err != nil {
		h.writeError(s, "import provenance bundle: "+err.Error())
		return
	}

	if h.logger != nil {
		h.logger.Info("provenance bundle imported",
			zap.String("peerDID", peerDID),
			zap.Int("bundleSize", len(payload.Bundle)))
	}

	h.writeResponse(s, &PushBundleResponse{
		Stored:  true,
		Message: "provenance bundle imported",
	})
}

func (h *Handler) writeError(s network.Stream, message string) {
	_ = json.NewEncoder(s).Encode(&PushBundleResponse{
		Stored:  false,
		Message: message,
	})
}

func (h *Handler) writeResponse(s network.Stream, resp *PushBundleResponse) {
	_ = json.NewEncoder(s).Encode(resp)
}

// PushBundle sends a provenance bundle to a remote peer over the dedicated protocol.
func PushBundle(ctx context.Context, host host.Host, peerID peer.ID, token string, bundle []byte) (*PushBundleResponse, error) {
	stream, err := host.NewStream(ctx, peerID, protocol.ID(ProtocolID))
	if err != nil {
		return nil, fmt.Errorf("open provenance stream: %w", err)
	}
	defer stream.Close()

	payload, err := json.Marshal(PushBundlePayload{Bundle: bundle})
	if err != nil {
		return nil, fmt.Errorf("marshal provenance payload: %w", err)
	}
	req := Request{
		Type:    RequestPushBundle,
		Token:   token,
		Payload: payload,
	}
	if err := json.NewEncoder(stream).Encode(&req); err != nil {
		return nil, fmt.Errorf("send provenance request: %w", err)
	}

	var resp PushBundleResponse
	if err := json.NewDecoder(stream).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode provenance response: %w", err)
	}
	if !resp.Stored {
		return &resp, errors.New(resp.Message)
	}
	return &resp, nil
}
