package ctxkeys

import (
	"context"
	"os/user"
	"strings"
)

// WithDefaultPrincipal injects a stable local principal when the context does
// not already carry one. Existing explicit principals are preserved.
func WithDefaultPrincipal(ctx context.Context, fallback string) context.Context {
	if strings.TrimSpace(PrincipalFromContext(ctx)) != "" {
		return ctx
	}
	if name := strings.TrimSpace(AgentNameFromContext(ctx)); name != "" {
		return WithPrincipal(ctx, name)
	}
	if current, err := user.Current(); err == nil {
		if username := strings.TrimSpace(current.Username); username != "" {
			return WithPrincipal(ctx, "operator:"+username)
		}
	}
	if fallback = strings.TrimSpace(fallback); fallback != "" {
		return WithPrincipal(ctx, fallback)
	}
	return ctx
}
