package approval

import (
	"context"
	"fmt"
	"sync"
)

// CompositeProvider routes approval requests to the appropriate provider
// based on session key prefix. Falls back to TTY, then denies (fail-closed).
type CompositeProvider struct {
	mu          sync.RWMutex
	providers   []Provider
	ttyFallback Provider
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

// RequestApproval routes the request to the first provider whose CanHandle
// returns true. If none match, falls back to TTY. If TTY is unavailable, denies.
func (c *CompositeProvider) RequestApproval(ctx context.Context, req ApprovalRequest) (ApprovalResponse, error) {
	c.mu.RLock()
	providers := make([]Provider, len(c.providers))
	copy(providers, c.providers)
	tty := c.ttyFallback
	c.mu.RUnlock()

	for _, p := range providers {
		if p.CanHandle(req.SessionKey) {
			return p.RequestApproval(ctx, req)
		}
	}

	// TTY fallback
	if tty != nil {
		return tty.RequestApproval(ctx, req)
	}

	// Fail-closed: no provider available
	return ApprovalResponse{}, fmt.Errorf("no approval provider for session %q", req.SessionKey)
}

// CanHandle always returns true; CompositeProvider accepts all requests
// and routes internally.
func (c *CompositeProvider) CanHandle(_ string) bool {
	return true
}
