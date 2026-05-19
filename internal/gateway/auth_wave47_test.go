package gateway

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/langoai/lango/internal/session"
)

type wave47SessionStore struct {
	sessions   map[string]*session.Session
	deleteKeys []string
	deleteErr  error
}

func newWave47SessionStore() *wave47SessionStore {
	return &wave47SessionStore{sessions: make(map[string]*session.Session)}
}

func (s *wave47SessionStore) Create(sess *session.Session) error {
	s.sessions[sess.Key] = sess
	return nil
}

func (s *wave47SessionStore) Get(key string) (*session.Session, error) {
	return s.sessions[key], nil
}

func (s *wave47SessionStore) Update(sess *session.Session) error {
	s.sessions[sess.Key] = sess
	return nil
}

func (s *wave47SessionStore) Delete(key string) error {
	s.deleteKeys = append(s.deleteKeys, key)
	if s.deleteErr != nil {
		return s.deleteErr
	}
	delete(s.sessions, key)
	return nil
}

func (s *wave47SessionStore) AppendMessage(_ string, _ session.Message) error {
	return nil
}

func (s *wave47SessionStore) AnnotateTimeout(_ string, _ string) error {
	return nil
}

func (s *wave47SessionStore) End(_ string) error {
	return nil
}

func (s *wave47SessionStore) Close() error {
	return nil
}

func (s *wave47SessionStore) ListSessions(_ context.Context) ([]session.SessionSummary, error) {
	return nil, nil
}

func (s *wave47SessionStore) GetSalt(_ string) ([]byte, error) {
	return nil, nil
}

func (s *wave47SessionStore) SetSalt(_ string, _ []byte) error {
	return nil
}

type wave47RoundTripFunc func(*http.Request) (*http.Response, error)

func (f wave47RoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestAuthWave47LoginMissingProviderReturnsNotFound(t *testing.T) {
	t.Parallel()

	auth := &AuthManager{
		providers: make(map[string]*OIDCProvider),
		store:     newWave47SessionStore(),
	}
	req := requestWithProvider(http.MethodGet, "/auth/login/missing", "missing")
	rec := httptest.NewRecorder()

	auth.handleLogin(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), "provider not found")
}

func TestAuthWave47LoginSetsProviderStateCookieAndRedirects(t *testing.T) {
	t.Parallel()

	auth := &AuthManager{
		providers: map[string]*OIDCProvider{
			"google": {
				Name: "google",
				OAuthConfig: &oauth2.Config{
					ClientID:    "client-id",
					RedirectURL: "https://gateway.example/auth/callback/google",
					Endpoint: oauth2.Endpoint{
						AuthURL: "https://issuer.example/authorize",
					},
					Scopes: []string{"openid", "email"},
				},
			},
		},
		store: newWave47SessionStore(),
	}
	req := requestWithProvider(http.MethodGet, "/auth/login/google", "google")
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()

	auth.handleLogin(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
	location := rec.Header().Get("Location")
	require.Contains(t, location, "https://issuer.example/authorize")
	require.Contains(t, location, "client_id=client-id")
	require.Contains(t, location, "redirect_uri=https%3A%2F%2Fgateway.example")
	redirectURL, err := url.Parse(location)
	require.NoError(t, err)

	stateCookie := requireCookie(t, rec.Result().Cookies(), "oauth_state_google")
	require.NotEmpty(t, stateCookie.Value)
	require.True(t, stateCookie.HttpOnly)
	require.True(t, stateCookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, stateCookie.SameSite)
	require.WithinDuration(t, time.Now().Add(10*time.Minute), stateCookie.Expires, 5*time.Second)
	require.Equal(t, stateCookie.Value, redirectURL.Query().Get("state"))
}

func TestAuthWave47CallbackMissingProviderReturnsNotFound(t *testing.T) {
	t.Parallel()

	auth := &AuthManager{
		providers: make(map[string]*OIDCProvider),
		store:     newWave47SessionStore(),
	}
	req := requestWithProvider(http.MethodGet, "/auth/callback/missing?state=s&code=c", "missing")
	req.AddCookie(&http.Cookie{Name: "oauth_state_missing", Value: "s"})
	rec := httptest.NewRecorder()

	auth.handleCallback(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), "provider not found")
}

func TestAuthWave47CallbackRejectsMissingStateCookie(t *testing.T) {
	t.Parallel()

	auth := authManagerWithCallbackProvider(t)
	req := requestWithProvider(http.MethodGet, "/auth/callback/google?state=s&code=c", "google")
	rec := httptest.NewRecorder()

	auth.handleCallback(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "state cookie missing")
}

func TestAuthWave47CallbackRejectsStateMismatch(t *testing.T) {
	t.Parallel()

	auth := authManagerWithCallbackProvider(t)
	req := requestWithProvider(http.MethodGet, "/auth/callback/google?state=request&code=c", "google")
	req.AddCookie(&http.Cookie{Name: "oauth_state_google", Value: "cookie"})
	rec := httptest.NewRecorder()

	auth.handleCallback(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "state mismatch")
	require.Empty(t, rec.Result().Cookies())
}

func TestAuthWave47CallbackDeletesStateCookieBeforeTokenFailure(t *testing.T) {
	t.Parallel()

	var tokenRequests int
	transport := wave47RoundTripFunc(func(r *http.Request) (*http.Response, error) {
		tokenRequests++
		require.Equal(t, "https://issuer.example/token", r.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       http.NoBody,
		}, nil
	})
	auth := authManagerWithCallbackProvider(t)
	req := requestWithProvider(http.MethodGet, "/auth/callback/google?state=s&code=c", "google")
	req = req.WithContext(context.WithValue(req.Context(), oauth2.HTTPClient, &http.Client{
		Transport: transport,
	}))
	req.AddCookie(&http.Cookie{Name: "oauth_state_google", Value: "s"})
	rec := httptest.NewRecorder()

	auth.handleCallback(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Positive(t, tokenRequests)

	stateCookie := requireCookie(t, rec.Result().Cookies(), "oauth_state_google")
	require.Equal(t, "", stateCookie.Value)
	require.Equal(t, -1, stateCookie.MaxAge)
	require.True(t, stateCookie.HttpOnly)
	require.False(t, stateCookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, stateCookie.SameSite)
}

func TestAuthWave47LogoutClearsCookieAndDeletesSessionWhenPresent(t *testing.T) {
	t.Parallel()

	store := newWave47SessionStore()
	require.NoError(t, store.Create(&session.Session{
		Key:       "sess_wave47",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}))
	auth := &AuthManager{store: store}
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.AddCookie(&http.Cookie{Name: "lango_session", Value: "sess_wave47"})
	rec := httptest.NewRecorder()

	auth.handleLogout(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"status":"logged_out"}`, rec.Body.String())
	require.Equal(t, []string{"sess_wave47"}, store.deleteKeys)
	require.Nil(t, store.sessions["sess_wave47"])

	cookie := requireCookie(t, rec.Result().Cookies(), "lango_session")
	require.Equal(t, "", cookie.Value)
	require.Equal(t, -1, cookie.MaxAge)
	require.True(t, cookie.HttpOnly)
	require.True(t, cookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}

func TestAuthWave47LogoutStillClearsCookieWhenSessionDeleteFails(t *testing.T) {
	t.Parallel()

	store := newWave47SessionStore()
	store.deleteErr = errors.New("delete failed")
	auth := &AuthManager{store: store}
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "lango_session", Value: "sess_stale"})
	rec := httptest.NewRecorder()

	auth.handleLogout(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, []string{"sess_stale"}, store.deleteKeys)

	cookie := requireCookie(t, rec.Result().Cookies(), "lango_session")
	require.Equal(t, "", cookie.Value)
	require.Equal(t, -1, cookie.MaxAge)
}

func TestAuthWave47LogoutWithoutCookieDoesNotDeleteSession(t *testing.T) {
	t.Parallel()

	store := newWave47SessionStore()
	auth := &AuthManager{store: store}
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()

	auth.handleLogout(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Empty(t, store.deleteKeys)
	require.Equal(t, "", requireCookie(t, rec.Result().Cookies(), "lango_session").Value)
}

func TestAuthWave47GenerateRandomStringProperties(t *testing.T) {
	t.Parallel()

	first, err := generateRandomString(32)
	require.NoError(t, err)
	second, err := generateRandomString(32)
	require.NoError(t, err)

	require.NotEqual(t, first, second)
	require.Len(t, first, 44)
	decoded, err := base64.URLEncoding.DecodeString(first)
	require.NoError(t, err)
	require.Len(t, decoded, 32)
}

func authManagerWithCallbackProvider(t *testing.T) *AuthManager {
	t.Helper()

	return &AuthManager{
		providers: map[string]*OIDCProvider{
			"google": {
				Name: "google",
				OAuthConfig: &oauth2.Config{
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					RedirectURL:  "https://gateway.example/auth/callback/google",
					Endpoint: oauth2.Endpoint{
						AuthURL:  "https://issuer.example/authorize",
						TokenURL: "https://issuer.example/token",
					},
				},
			},
		},
		store: newWave47SessionStore(),
	}
}

func requestWithProvider(method string, target string, provider string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("provider", provider)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
	return req
}

func requireCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()

	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}

	var names []string
	for _, cookie := range cookies {
		names = append(names, cookie.Name)
	}
	require.Failf(t, "missing cookie", "cookie %q not found in [%s]", name, strings.Join(names, ", "))
	return nil
}
