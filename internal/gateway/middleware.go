package gateway

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is an unexported type for context keys in this package.
type contextKey int

const (
	// sessionContextKey stores the authenticated session key in request context.
	sessionContextKey contextKey = iota
)

// SessionFromContext extracts the authenticated session key from the request context.
// Returns empty string if no session is present.
func SessionFromContext(ctx context.Context) string {
	v, _ := ctx.Value(sessionContextKey).(string)
	return v
}

// requireAuth returns chi middleware that validates the lango_session cookie
// against the AuthManager's session store. If auth is nil (no OIDC configured),
// all requests pass through unchanged (development/local mode).
func requireAuth(auth *AuthManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// No auth configured — pass through (dev mode)
			if auth == nil {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie("lango_session")
			if err != nil || cookie.Value == "" {
				http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
				return
			}

			sess, err := auth.store.Get(cookie.Value)
			if err != nil || sess == nil {
				http.Error(w, `{"error":"invalid or expired session"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), sessionContextKey, sess.Key)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// makeOriginChecker builds a CheckOrigin function for gorilla/websocket.Upgrader.
// - Empty list: returns nil (gorilla default behavior = same-origin check).
// - Single "*" entry: allows all origins.
// - Otherwise: matches the request Origin header against the allowed list.
func makeOriginChecker(allowedOrigins []string) func(*http.Request) bool {
	if len(allowedOrigins) == 0 {
		return nil
	}

	for _, o := range allowedOrigins {
		if o == "*" {
			return func(_ *http.Request) bool { return true }
		}
	}

	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[strings.TrimRight(o, "/")] = struct{}{}
	}

	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // no Origin header — same-origin request
		}
		_, ok := allowed[strings.TrimRight(origin, "/")]
		return ok
	}
}

// isSecure reports whether the request was made over HTTPS, either directly
// (r.TLS != nil) or via a reverse proxy (X-Forwarded-Proto: https).
func isSecure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}
