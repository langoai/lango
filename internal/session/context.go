package session

import "context"

// sessionKeyCtxKey is the context key type for session keys.
type sessionKeyCtxKey struct{}

// runContextCtxKey is the context key type for structured run metadata.
type runContextCtxKey struct{}

// RunContext carries structured metadata for workflow/background run sessions.
type RunContext struct {
	SessionType string
	WorkflowID  string
	RunID       string
}

// WithSessionKey adds a session key to the context.
func WithSessionKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, sessionKeyCtxKey{}, key)
}

// SessionKeyFromContext extracts the session key from context.
func SessionKeyFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(sessionKeyCtxKey{}).(string); ok {
		return v
	}
	return ""
}

// WithRunContext adds structured run metadata to the context.
func WithRunContext(ctx context.Context, rc RunContext) context.Context {
	return context.WithValue(ctx, runContextCtxKey{}, rc)
}

// RunContextFromContext extracts structured run metadata from context.
func RunContextFromContext(ctx context.Context) *RunContext {
	v, ok := ctx.Value(runContextCtxKey{}).(RunContext)
	if !ok {
		return nil
	}
	return &v
}

// turnIDCtxKey is the context key type for turn IDs.
type turnIDCtxKey struct{}

// WithTurnID adds a turn ID to the context.
func WithTurnID(ctx context.Context, turnID string) context.Context {
	return context.WithValue(ctx, turnIDCtxKey{}, turnID)
}

// TurnIDFromContext extracts the turn ID from context.
func TurnIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(turnIDCtxKey{}).(string); ok {
		return v
	}
	return ""
}
