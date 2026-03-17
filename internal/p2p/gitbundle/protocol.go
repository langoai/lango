package gitbundle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"go.uber.org/zap"
)

// SessionValidator validates a session token and returns the peer DID.
type SessionValidator func(token string) (string, bool)

// Handler handles git protocol streams.
type Handler struct {
	service   *Service
	validator SessionValidator
	maxBundle int64
	logger    *zap.Logger
}

// HandlerConfig configures the git protocol handler.
type HandlerConfig struct {
	Service       *Service
	Validator     SessionValidator
	MaxBundleSize int64 // bytes, default 50MB
	Logger        *zap.Logger
}

// NewHandler creates a new git protocol stream handler.
func NewHandler(cfg HandlerConfig) *Handler {
	if cfg.MaxBundleSize <= 0 {
		cfg.MaxBundleSize = 50 * 1024 * 1024 // 50MB
	}
	return &Handler{
		service:   cfg.Service,
		validator: cfg.Validator,
		maxBundle: cfg.MaxBundleSize,
		logger:    cfg.Logger,
	}
}

// StreamHandler returns the libp2p stream handler function.
func (h *Handler) StreamHandler() network.StreamHandler {
	return func(s network.Stream) {
		defer s.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Read request using streaming decoder to avoid double-buffering.
		var req Request
		if err := json.NewDecoder(io.LimitReader(s, h.maxBundle+4096)).Decode(&req); err != nil {
			h.writeError(s, "decode request: "+err.Error())
			return
		}

		// Validate session.
		if h.validator != nil {
			peerDID, ok := h.validator(req.Token)
			if !ok {
				h.writeError(s, "invalid or expired session token")
				return
			}
			h.logger.Debug("git request authenticated",
				zap.String("peerDID", peerDID),
				zap.String("type", string(req.Type)),
				zap.String("workspace", req.WorkspaceID),
			)
		}

		// Dispatch request.
		switch req.Type {
		case RequestPushBundle:
			h.handlePushBundle(ctx, s, req)
		case RequestFetchByHash:
			h.handleFetchByHash(ctx, s, req)
		case RequestListCommits:
			h.handleListCommits(ctx, s, req)
		case RequestFindLeaves:
			h.handleFindLeaves(ctx, s, req)
		case RequestDiff:
			h.handleDiff(ctx, s, req)
		case RequestPushIncrementalBundle:
			h.handlePushIncrementalBundle(ctx, s, req)
		case RequestFetchIncremental:
			h.handleFetchIncremental(ctx, s, req)
		case RequestVerifyBundle:
			h.handleVerifyBundle(ctx, s, req)
		case RequestHasCommit:
			h.handleHasCommit(ctx, s, req)
		default:
			h.writeError(s, fmt.Sprintf("unknown request type: %s", req.Type))
		}
	}
}

func (h *Handler) handlePushBundle(ctx context.Context, s network.Stream, req Request) {
	var payload PushBundlePayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		h.writeError(s, "unmarshal push payload: "+err.Error())
		return
	}

	if int64(len(payload.Bundle)) > h.maxBundle {
		h.writeError(s, fmt.Sprintf("bundle too large: %d > %d", len(payload.Bundle), h.maxBundle))
		return
	}

	if err := h.service.ApplyBundle(ctx, req.WorkspaceID, payload.Bundle); err != nil {
		h.writeError(s, "apply bundle: "+err.Error())
		return
	}

	h.writeResponse(s, &PushBundleResponse{
		Applied: true,
		Message: "bundle applied successfully",
	})
}

func (h *Handler) handleFetchByHash(ctx context.Context, s network.Stream, req Request) {
	// For fetch, we create a bundle from the workspace repo and send it.
	bundle, hash, err := h.service.CreateBundle(ctx, req.WorkspaceID)
	if err != nil {
		h.writeError(s, "create bundle: "+err.Error())
		return
	}

	if bundle == nil {
		h.writeError(s, "empty repository")
		return
	}

	h.writeResponse(s, map[string]interface{}{
		"bundle":     bundle,
		"headCommit": hash,
	})
}

func (h *Handler) handleListCommits(ctx context.Context, s network.Stream, req Request) {
	var payload ListCommitsPayload
	if req.Payload != nil {
		_ = json.Unmarshal(req.Payload, &payload)
	}
	if payload.Limit <= 0 {
		payload.Limit = 20
	}

	commits, err := h.service.Log(ctx, req.WorkspaceID, payload.Limit)
	if err != nil {
		h.writeError(s, "list commits: "+err.Error())
		return
	}

	h.writeResponse(s, &ListCommitsResponse{Commits: commits})
}

func (h *Handler) handleFindLeaves(ctx context.Context, s network.Stream, req Request) {
	leaves, err := h.service.Leaves(ctx, req.WorkspaceID)
	if err != nil {
		h.writeError(s, "find leaves: "+err.Error())
		return
	}

	h.writeResponse(s, &FindLeavesResponse{Leaves: leaves})
}

func (h *Handler) handleDiff(ctx context.Context, s network.Stream, req Request) {
	var payload DiffPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		h.writeError(s, "unmarshal diff payload: "+err.Error())
		return
	}

	diff, err := h.service.Diff(ctx, req.WorkspaceID, payload.From, payload.To)
	if err != nil {
		h.writeError(s, "diff: "+err.Error())
		return
	}

	h.writeResponse(s, &DiffResponse{Diff: diff})
}

func (h *Handler) handlePushIncrementalBundle(ctx context.Context, s network.Stream, req Request) {
	var payload PushIncrementalBundlePayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		h.writeError(s, "unmarshal incremental push payload: "+err.Error())
		return
	}

	if int64(len(payload.Bundle)) > h.maxBundle {
		h.writeError(s, fmt.Sprintf("bundle too large: %d > %d", len(payload.Bundle), h.maxBundle))
		return
	}

	if err := h.service.SafeApplyBundle(ctx, req.WorkspaceID, payload.Bundle); err != nil {
		h.writeError(s, "safe apply bundle: "+err.Error())
		return
	}

	h.writeResponse(s, &PushBundleResponse{
		Applied: true,
		Message: "incremental bundle applied safely",
	})
}

func (h *Handler) handleFetchIncremental(ctx context.Context, s network.Stream, req Request) {
	var payload FetchIncrementalPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		h.writeError(s, "unmarshal fetch incremental payload: "+err.Error())
		return
	}

	bundle, hash, err := h.service.CreateIncrementalBundle(ctx, req.WorkspaceID, payload.BaseCommit)
	if err != nil {
		h.writeError(s, "create incremental bundle: "+err.Error())
		return
	}

	if bundle == nil {
		h.writeError(s, "empty repository")
		return
	}

	h.writeResponse(s, &FetchIncrementalResponse{
		Bundle:     bundle,
		HeadCommit: hash,
	})
}

func (h *Handler) handleVerifyBundle(ctx context.Context, s network.Stream, req Request) {
	var payload VerifyBundlePayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		h.writeError(s, "unmarshal verify payload: "+err.Error())
		return
	}

	if err := h.service.VerifyBundle(ctx, req.WorkspaceID, payload.Bundle); err != nil {
		h.writeResponse(s, &VerifyBundleResponse{
			Valid:   false,
			Message: err.Error(),
		})
		return
	}

	h.writeResponse(s, &VerifyBundleResponse{Valid: true})
}

func (h *Handler) handleHasCommit(ctx context.Context, s network.Stream, req Request) {
	var payload HasCommitPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		h.writeError(s, "unmarshal has_commit payload: "+err.Error())
		return
	}

	exists, err := h.service.HasCommit(ctx, req.WorkspaceID, payload.CommitHash)
	if err != nil {
		h.writeError(s, "has commit: "+err.Error())
		return
	}

	h.writeResponse(s, &HasCommitResponse{
		Exists: exists,
		Hash:   payload.CommitHash,
	})
}

func (h *Handler) writeResponse(s network.Stream, data interface{}) {
	resp := Response{Status: StatusOK}
	if data != nil {
		raw, err := json.Marshal(data)
		if err != nil {
			h.writeError(s, "marshal response: "+err.Error())
			return
		}
		resp.Data = raw
	}
	b, _ := json.Marshal(resp)
	_, _ = s.Write(b)
}

func (h *Handler) writeError(s network.Stream, msg string) {
	h.logger.Warn("git protocol error", zap.String("error", msg))
	resp := Response{Status: StatusError, Error: msg}
	b, _ := json.Marshal(resp)
	_, _ = s.Write(b)
}
