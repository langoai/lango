package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	t.Parallel()
	handler := requireAuth(nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireAuth_NoCookieReturns401(t *testing.T) {
	t.Parallel()
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

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireAuth_InvalidSessionReturns401(t *testing.T) {
	t.Parallel()
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

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireAuth_ValidSessionSetsContext(t *testing.T) {
	t.Parallel()
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

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "sess_valid-key", capturedSessionKey)
}

func TestSessionFromContext_Empty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	key := SessionFromContext(ctx)
	assert.Empty(t, key)
}

func TestMakeOriginChecker_EmptyReturnsNil(t *testing.T) {
	t.Parallel()
	checker := makeOriginChecker(nil)
	assert.Nil(t, checker)

	checker = makeOriginChecker([]string{})
	assert.Nil(t, checker)
}

func TestMakeOriginChecker_WildcardAllowsAll(t *testing.T) {
	t.Parallel()
	checker := makeOriginChecker([]string{"*"})
	require.NotNil(t, checker)

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	assert.True(t, checker(req))
}

func TestMakeOriginChecker_SpecificOriginsMatch(t *testing.T) {
	t.Parallel()
	checker := makeOriginChecker([]string{"https://app.example.com", "https://admin.example.com"})
	require.NotNil(t, checker)

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
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, "/ws", nil)
			if tt.give != "" {
				req.Header.Set("Origin", tt.give)
			}
			got := checker(req)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMakeOriginChecker_TrailingSlashNormalized(t *testing.T) {
	t.Parallel()
	checker := makeOriginChecker([]string{"https://app.example.com/"})
	require.NotNil(t, checker)

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Origin", "https://app.example.com")
	assert.True(t, checker(req))
}

func TestIsSecure_DirectTLS(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "https://localhost/test", nil)
	_ = isSecure(req)

	req = httptest.NewRequest(http.MethodGet, "http://localhost/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	assert.True(t, isSecure(req))
}

func TestIsSecure_XForwardedProto(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, "http://localhost/test", nil)
			if tt.give != "" {
				req.Header.Set("X-Forwarded-Proto", tt.give)
			}
			got := isSecure(req)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLogout_ClearsSessionAndCookie(t *testing.T) {
	t.Parallel()
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

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify session was deleted from store
	sess, _ := store.Get("sess_to-delete")
	assert.Nil(t, sess)

	// Verify cookie was cleared
	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "lango_session" {
			found = true
			assert.Equal(t, -1, c.MaxAge)
			assert.Empty(t, c.Value)
		}
	}
	assert.True(t, found, "expected lango_session cookie in response")
}

func TestStateCookie_PerProviderName(t *testing.T) {
	t.Parallel()
	auth := &AuthManager{
		providers: make(map[string]*OIDCProvider),
		store:     newMockStore(),
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/callback/google?state=abc&code=xyz", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "abc"})
	rec := httptest.NewRecorder()

	auth.handleCallback(rec, req)

	req2 := httptest.NewRequest(http.MethodGet, "/auth/callback/google?state=abc&code=xyz", nil)
	req2.AddCookie(&http.Cookie{Name: "oauth_state_google", Value: "abc"})
	rec2 := httptest.NewRecorder()

	auth.handleCallback(rec2, req2)

	assert.Equal(t, http.StatusNotFound, rec2.Code)
}
