package session

import "context"

// sessionKeyCtxKey is the context key type for session keys.
type sessionKeyCtxKey struct{}

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
