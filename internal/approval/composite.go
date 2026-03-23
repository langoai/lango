package approval

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/langoai/lango/internal/logging"
)

// CompositeProvider routes approval requests to the appropriate provider
// based on session key prefix. Falls back to TTY, then denies (fail-closed).
// P2P sessions ("p2p:..." keys) use a dedicated fallback and are never
// routed to HeadlessProvider to prevent remote peers from auto-approving.
type CompositeProvider struct {
	mu          sync.RWMutex
	providers   []Provider
	ttyFallback Provider
	p2pFallback Provider
}

type providerNamer interface {
	Name() string
}

// NewCompositeProvider creates a new CompositeProvider.
func NewCompositeProvider() *CompositeProvider {
	return &CompositeProvider{}
}

// Register appends a provider to the routing chain.
func (c *CompositeProvider) Register(p Provider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.providers = append(c.providers, p)
}

// SetTTYFallback sets the TTY provider used when no other provider matches.
func (c *CompositeProvider) SetTTYFallback(p Provider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ttyFallback = p
}

// SetP2PFallback sets a dedicated approval provider for P2P sessions.
// P2P sessions are never routed to the TTY fallback when it is a
// HeadlessProvider, preventing remote peers from auto-approving.
func (c *CompositeProvider) SetP2PFallback(p Provider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.p2pFallback = p
}

// RequestApproval routes the request to the first provider whose CanHandle
// returns true. P2P sessions ("p2p:..." keys) use a dedicated fallback
// instead of the TTY fallback to ensure HeadlessProvider never auto-approves
// remote peer requests. If no provider matches, denies (fail-closed).
func (c *CompositeProvider) RequestApproval(ctx context.Context, req ApprovalRequest) (ApprovalResponse, error) {
	c.mu.RLock()
	providers := make([]Provider, len(c.providers))
	copy(providers, c.providers)
	tty := c.ttyFallback
	p2p := c.p2pFallback
	c.mu.RUnlock()

	for _, p := range providers {
		if p.CanHandle(req.SessionKey) {
			provider := providerName(p)
			hash := ""
			if _, h, err := TurnApprovalKey(req.ToolName, req.Params); err == nil {
				hash = h
			}
			logging.SubsystemSugar("approval").Debugw("approval requested",
				"session", req.SessionKey,
				"request_id", req.ID,
				"tool", req.ToolName,
				"summary", req.Summary,
				"params_hash", hash,
				"outcome", "requested",
				"grant_scope", "none",
				"provider", provider,
			)
			resp, err := p.RequestApproval(ctx, req)
			if err != nil {
				return ApprovalResponse{}, err
			}
			if resp.Provider == "" {
				resp.Provider = provider
			}
			return resp, nil
		}
	}

	// P2P sessions: use dedicated fallback, NEVER HeadlessProvider.
	if strings.HasPrefix(req.SessionKey, "p2p:") {
		if p2p != nil {
			return p2p.RequestApproval(ctx, req)
		}
		return ApprovalResponse{}, WrapError(
			ErrUnavailable,
			"p2p_fallback",
			req.ID,
			fmt.Sprintf(
				"no approval provider for P2P session %q (headless auto-approve is not allowed for remote peers)",
				req.SessionKey,
			),
		)
	}

	// TTY fallback (non-P2P only).
	if tty != nil {
		return tty.RequestApproval(ctx, req)
	}

	// Fail-closed: no provider available.
	return ApprovalResponse{}, WrapError(
		ErrUnavailable,
		"composite",
		req.ID,
		fmt.Sprintf("no approval provider for session %q", req.SessionKey),
	)
}

// CanHandle always returns true; CompositeProvider accepts all requests
// and routes internally.
func (c *CompositeProvider) CanHandle(_ string) bool {
	return true
}

func providerName(p Provider) string {
	if named, ok := p.(providerNamer); ok {
		return named.Name()
	}
	return "unknown"
}
