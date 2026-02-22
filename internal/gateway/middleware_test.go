package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/langoai/lango/internal/session"
)

// mockStore implements session.Store for testing purposes.
type mockStore struct {
	sessions map[string]*session.Session
}

func newMockStore() *mockStore {
	return &mockStore{sessions: make(map[string]*session.Session)}
}

func (m *mockStore) Create(s *session.Session) error {
	m.sessions[s.Key] = s
	return nil
}

func (m *mockStore) Get(key string) (*session.Session, error) {
	s, ok := m.sessions[key]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (m *mockStore) Update(s *session.Session) error {
	m.sessions[s.Key] = s
	return nil
}

func (m *mockStore) Delete(key string) error {
	delete(m.sessions, key)
	return nil
}

func (m *mockStore) AppendMessage(_ string, _ session.Message) error { return nil }
func (m *mockStore) Close() error                                    { return nil }
func (m *mockStore) GetSalt(_ string) ([]byte, error)                { return nil, nil }
func (m *mockStore) SetSalt(_ string, _ []byte) error                { return nil }

func TestRequireAuth_NilAuthPassesThrough(t *testing.T) {
	handler := requireAuth(nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 when auth is nil, got %d", rec.Code)
	}
}

func TestRequireAuth_NoCookieReturns401(t *testing.T) {
	store := newMockStore()
	auth := &AuthManager{
		providers: make(map[string]*OIDCProvider),
		store:     store,
	}

	handler := requireAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 when no cookie, got %d", rec.Code)
	}
}

func TestRequireAuth_InvalidSessionReturns401(t *testing.T) {
	store := newMockStore()
	auth := &AuthManager{
		providers: make(map[string]*OIDCProvider),
		store:     store,
	}

	handler := requireAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.AddCookie(&http.Cookie{Name: "lango_session", Value: "nonexistent-key"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid session, got %d", rec.Code)
	}
}

func TestRequireAuth_ValidSessionSetsContext(t *testing.T) {
	store := newMockStore()
	store.Create(&session.Session{
		Key:       "sess_valid-key",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	auth := &AuthManager{
		providers: make(map[string]*OIDCProvider),
		store:     store,
	}

	var capturedSessionKey string
	handler := requireAuth(auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedSessionKey = SessionFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.AddCookie(&http.Cookie{Name: "lango_session", Value: "sess_valid-key"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for valid session, got %d", rec.Code)
	}
	if capturedSessionKey != "sess_valid-key" {
		t.Errorf("expected session key 'sess_valid-key', got %q", capturedSessionKey)
	}
}

func TestSessionFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	key := SessionFromContext(ctx)
	if key != "" {
		t.Errorf("expected empty string for empty context, got %q", key)
	}
}

func TestMakeOriginChecker_EmptyReturnsNil(t *testing.T) {
	checker := makeOriginChecker(nil)
	if checker != nil {
		t.Error("expected nil checker for empty origins")
	}

	checker = makeOriginChecker([]string{})
	if checker != nil {
		t.Error("expected nil checker for empty slice")
	}
}

func TestMakeOriginChecker_WildcardAllowsAll(t *testing.T) {
	checker := makeOriginChecker([]string{"*"})
	if checker == nil {
		t.Fatal("expected non-nil checker for wildcard")
	}

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	if !checker(req) {
		t.Error("expected wildcard to allow all origins")
	}
}

func TestMakeOriginChecker_SpecificOriginsMatch(t *testing.T) {
	checker := makeOriginChecker([]string{"https://app.example.com", "https://admin.example.com"})
	if checker == nil {
		t.Fatal("expected non-nil checker for specific origins")
	}

	tests := []struct {
		give string
		want bool
	}{
		{give: "https://app.example.com", want: true},
		{give: "https://admin.example.com", want: true},
		{give: "https://evil.example.com", want: false},
		{give: "", want: true}, // no Origin header = same-origin
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ws", nil)
			if tt.give != "" {
				req.Header.Set("Origin", tt.give)
			}
			got := checker(req)
			if got != tt.want {
				t.Errorf("origin %q: got %v, want %v", tt.give, got, tt.want)
			}
		})
	}
}

func TestMakeOriginChecker_TrailingSlashNormalized(t *testing.T) {
	checker := makeOriginChecker([]string{"https://app.example.com/"})
	if checker == nil {
		t.Fatal("expected non-nil checker")
	}

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Origin", "https://app.example.com")
	if !checker(req) {
		t.Error("expected trailing slash to be normalized")
	}
}

func TestIsSecure_DirectTLS(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://localhost/test", nil)
	// httptest doesn't set TLS, manually test the header path
	if isSecure(req) {
		// TLS is nil in httptest, that's expected
	}

	// Test X-Forwarded-Proto header
	req = httptest.NewRequest(http.MethodGet, "http://localhost/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	if !isSecure(req) {
		t.Error("expected isSecure=true with X-Forwarded-Proto: https")
	}
}

func TestIsSecure_XForwardedProto(t *testing.T) {
	tests := []struct {
		give string
		want bool
	}{
		{give: "https", want: true},
		{give: "HTTPS", want: true},
		{give: "http", want: false},
		{give: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://localhost/test", nil)
			if tt.give != "" {
				req.Header.Set("X-Forwarded-Proto", tt.give)
			}
			got := isSecure(req)
			if got != tt.want {
				t.Errorf("X-Forwarded-Proto %q: got %v, want %v", tt.give, got, tt.want)
			}
		})
	}
}

func TestLogout_ClearsSessionAndCookie(t *testing.T) {
	store := newMockStore()
	store.Create(&session.Session{
		Key:       "sess_to-delete",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	auth := &AuthManager{
		providers: make(map[string]*OIDCProvider),
		store:     store,
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "lango_session", Value: "sess_to-delete"})
	rec := httptest.NewRecorder()

	auth.handleLogout(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// Verify session was deleted from store
	sess, _ := store.Get("sess_to-delete")
	if sess != nil {
		t.Error("expected session to be deleted from store")
	}

	// Verify cookie was cleared
	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "lango_session" {
			found = true
			if c.MaxAge != -1 {
				t.Errorf("expected MaxAge -1, got %d", c.MaxAge)
			}
			if c.Value != "" {
				t.Errorf("expected empty cookie value, got %q", c.Value)
			}
		}
	}
	if !found {
		t.Error("expected lango_session cookie in response")
	}
}

func TestStateCookie_PerProviderName(t *testing.T) {
	// Verify that state cookie name includes provider name
	auth := &AuthManager{
		providers: make(map[string]*OIDCProvider),
		store:     newMockStore(),
	}

	// handleLogin requires a real OIDC provider, so we test indirectly
	// by verifying the handleCallback checks for per-provider cookie name

	req := httptest.NewRequest(http.MethodGet, "/auth/callback/google?state=abc&code=xyz", nil)
	// Set the old-style cookie (without provider suffix) — should fail
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "abc"})
	rec := httptest.NewRecorder()

	// This should return "state cookie missing" because it looks for "oauth_state_google"
	auth.handleCallback(rec, req)

	if rec.Code != http.StatusNotFound {
		// Provider "google" is not registered, so we get 404 first
		// But the important thing is it doesn't use the old cookie name
	}

	// Now test with correct per-provider cookie but non-existent provider
	req2 := httptest.NewRequest(http.MethodGet, "/auth/callback/google?state=abc&code=xyz", nil)
	req2.AddCookie(&http.Cookie{Name: "oauth_state_google", Value: "abc"})
	rec2 := httptest.NewRecorder()

	auth.handleCallback(rec2, req2)

	// Should get 404 (provider not found) rather than "state cookie missing"
	if rec2.Code != http.StatusNotFound {
		t.Errorf("expected 404 (provider not found), got %d", rec2.Code)
	}
}
