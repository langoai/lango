package approval

import "context"

type approvalTargetCtxKey struct{}

// WithApprovalTarget sets an explicit approval routing target in the context.
// This overrides the session key for approval routing, allowing automation
// systems (cron, background) to route approval requests to the originating channel.
func WithApprovalTarget(ctx context.Context, target string) context.Context {
	return context.WithValue(ctx, approvalTargetCtxKey{}, target)
}

// ApprovalTargetFromContext retrieves the approval routing target from the context.
// Returns empty string if no target is set.
func ApprovalTargetFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(approvalTargetCtxKey{}).(string); ok {
		return v
	}
	return ""
}
