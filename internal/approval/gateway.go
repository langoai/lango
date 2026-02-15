package approval

import (
	"context"
	"fmt"
)

// GatewayApprover abstracts the gateway.Server methods needed for approval.
// gateway.Server already satisfies this interface — no gateway changes needed.
type GatewayApprover interface {
	HasCompanions() bool
	RequestApproval(ctx context.Context, message string) (bool, error)
}

// GatewayProvider routes approval requests to companion apps via WebSocket.
type GatewayProvider struct {
	gw GatewayApprover
}

var _ Provider = (*GatewayProvider)(nil)

// NewGatewayProvider creates a GatewayProvider backed by the given gateway.
func NewGatewayProvider(gw GatewayApprover) *GatewayProvider {
	return &GatewayProvider{gw: gw}
}

// RequestApproval sends the approval request to connected companions.
func (g *GatewayProvider) RequestApproval(ctx context.Context, req ApprovalRequest) (bool, error) {
	msg := fmt.Sprintf("Tool '%s' requires approval", req.ToolName)
	return g.gw.RequestApproval(ctx, msg)
}

// CanHandle returns true when at least one companion is connected.
func (g *GatewayProvider) CanHandle(_ string) bool {
	return g.gw.HasCompanions()
}
