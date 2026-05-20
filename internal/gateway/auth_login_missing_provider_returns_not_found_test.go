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

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/session"
)

type authLoginMissingProviderReturnsNotFoundSessionStore struct {
	sessions   map[string]*session.Session
	deleteKeys []string
	deleteErr  error
}

func newAuthLoginMissingProviderReturnsNotFoundSessionStore() *authLoginMissingProviderReturnsNotFoundSessionStore {
	return &authLoginMissingProviderReturnsNotFoundSessionStore{sessions: make(map[string]*session.Session)}
}

func (s *authLoginMissingProviderReturnsNotFoundSessionStore) Create(sess *session.Session) error {
	s.sessions[sess.Key] = sess
	return nil
}

func (s *authLoginMissingProviderReturnsNotFoundSessionStore) Get(key string) (*session.Session, error) {
	return s.sessions[key], nil
}

func (s *authLoginMissingProviderReturnsNotFoundSessionStore) Update(sess *session.Session) error {
	s.sessions[sess.Key] = sess
	return nil
}

func (s *authLoginMissingProviderReturnsNotFoundSessionStore) Delete(key string) error {
	s.deleteKeys = append(s.deleteKeys, key)
	if s.deleteErr != nil {
		return s.deleteErr
	}
	delete(s.sessions, key)
	return nil
}

func (s *authLoginMissingProviderReturnsNotFoundSessionStore) AppendMessage(_ string, _ session.Message) error {
	return nil
}

func (s *authLoginMissingProviderReturnsNotFoundSessionStore) AnnotateTimeout(_ string, _ string) error {
	return nil
}

func (s *authLoginMissingProviderReturnsNotFoundSessionStore) End(_ string) error {
	return nil
}

func (s *authLoginMissingProviderReturnsNotFoundSessionStore) Close() error {
	return nil
}

func (s *authLoginMissingProviderReturnsNotFoundSessionStore) ListSessions(_ context.Context) ([]session.SessionSummary, error) {
	return nil, nil
}

func (s *authLoginMissingProviderReturnsNotFoundSessionStore) GetSalt(_ string) ([]byte, error) {
	return nil, nil
}

func (s *authLoginMissingProviderReturnsNotFoundSessionStore) SetSalt(_ string, _ []byte) error {
	return nil
}

type authLoginMissingProviderReturnsNotFoundRoundTripFunc func(*http.Request) (*http.Response, error)

func (f authLoginMissingProviderReturnsNotFoundRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestNewAuthManagerWithEmptyConfigRegistersNoProviders(t *testing.T) {
	t.Parallel()

	store := newAuthLoginMissingProviderReturnsNotFoundSessionStore()

	auth, err := NewAuthManager(config.AuthConfig{}, store)

	require.NoError(t, err)
	require.Same(t, store, auth.store)
	require.Empty(t, auth.providers)
}

func TestNewAuthManagerReturnsProviderCreationError(t *testing.T) {
	t.Parallel()

	auth, err := NewAuthManager(config.AuthConfig{
		Providers: map[string]config.OIDCProviderConfig{
			"broken": {
				IssuerURL: "://invalid-issuer",
				ClientID:  "client-id",
			},
		},
	}, newAuthLoginMissingProviderReturnsNotFoundSessionStore())

	require.Nil(t, auth)
	require.Error(t, err)
	require.Contains(t, err.Error(), "create provider broken")
	require.Contains(t, err.Error(), `query provider "://invalid-issuer"`)
}

func TestNewOIDCProviderReturnsInvalidIssuerURLError(t *testing.T) {
	t.Parallel()

	provider, err := NewOIDCProvider("broken", config.OIDCProviderConfig{
		IssuerURL: "://invalid-issuer",
		ClientID:  "client-id",
	})

	require.Nil(t, provider)
	require.Error(t, err)
	require.Contains(t, err.Error(), `query provider "://invalid-issuer"`)
}

func TestAuthRegisterRoutesDispatchesHandlers(t *testing.T) {
	t.Parallel()

	auth := &AuthManager{
		providers: make(map[string]*OIDCProvider),
		store:     newAuthLoginMissingProviderReturnsNotFoundSessionStore(),
	}
	router := chi.NewRouter()
	auth.RegisterRoutes(router)

	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, httptest.NewRequest(http.MethodGet, "/auth/login/missing", nil))

	require.Equal(t, http.StatusNotFound, loginRec.Code)
	require.Contains(t, loginRec.Body.String(), "provider not found")

	logoutRec := httptest.NewRecorder()
	logoutReq := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	logoutReq.Header.Set("X-Forwarded-Proto", "https")
	router.ServeHTTP(logoutRec, logoutReq)

	require.Equal(t, http.StatusOK, logoutRec.Code)
	require.JSONEq(t, `{"status":"logged_out"}`, logoutRec.Body.String())
	require.True(t, requireCookie(t, logoutRec.Result().Cookies(), "lango_session").Secure)
}

func TestAuthLoginMissingProviderReturnsNotFound(t *testing.T) {
	t.Parallel()

	auth := &AuthManager{
		providers: make(map[string]*OIDCProvider),
		store:     newAuthLoginMissingProviderReturnsNotFoundSessionStore(),
	}
	req := requestWithProvider(http.MethodGet, "/auth/login/missing", "missing")
	rec := httptest.NewRecorder()

	auth.handleLogin(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), "provider not found")
}

func TestAuthLoginSetsProviderStateCookieAndRedirects(t *testing.T) {
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
		store: newAuthLoginMissingProviderReturnsNotFoundSessionStore(),
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

func TestAuthCallbackMissingProviderReturnsNotFound(t *testing.T) {
	t.Parallel()

	auth := &AuthManager{
		providers: make(map[string]*OIDCProvider),
		store:     newAuthLoginMissingProviderReturnsNotFoundSessionStore(),
	}
	req := requestWithProvider(http.MethodGet, "/auth/callback/missing?state=s&code=c", "missing")
	req.AddCookie(&http.Cookie{Name: "oauth_state_missing", Value: "s"})
	rec := httptest.NewRecorder()

	auth.handleCallback(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), "provider not found")
}

func TestAuthCallbackRejectsMissingStateCookie(t *testing.T) {
	t.Parallel()

	auth := authManagerWithCallbackProvider(t)
	req := requestWithProvider(http.MethodGet, "/auth/callback/google?state=s&code=c", "google")
	rec := httptest.NewRecorder()

	auth.handleCallback(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "state cookie missing")
}

func TestAuthCallbackRejectsStateMismatch(t *testing.T) {
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

func TestAuthCallbackDeletesStateCookieBeforeTokenFailure(t *testing.T) {
	t.Parallel()

	var tokenRequests int
	transport := authLoginMissingProviderReturnsNotFoundRoundTripFunc(func(r *http.Request) (*http.Response, error) {
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

func TestAuthLogoutClearsCookieAndDeletesSessionWhenPresent(t *testing.T) {
	t.Parallel()

	store := newAuthLoginMissingProviderReturnsNotFoundSessionStore()
	require.NoError(t, store.Create(&session.Session{
		Key:       "sess_auth_callback",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}))
	auth := &AuthManager{store: store}
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.AddCookie(&http.Cookie{Name: "lango_session", Value: "sess_auth_callback"})
	rec := httptest.NewRecorder()

	auth.handleLogout(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"status":"logged_out"}`, rec.Body.String())
	require.Equal(t, []string{"sess_auth_callback"}, store.deleteKeys)
	require.Nil(t, store.sessions["sess_auth_callback"])

	cookie := requireCookie(t, rec.Result().Cookies(), "lango_session")
	require.Equal(t, "", cookie.Value)
	require.Equal(t, -1, cookie.MaxAge)
	require.True(t, cookie.HttpOnly)
	require.True(t, cookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}

func TestAuthLogoutStillClearsCookieWhenSessionDeleteFails(t *testing.T) {
	t.Parallel()

	store := newAuthLoginMissingProviderReturnsNotFoundSessionStore()
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

func TestAuthLogoutWithoutCookieDoesNotDeleteSession(t *testing.T) {
	t.Parallel()

	store := newAuthLoginMissingProviderReturnsNotFoundSessionStore()
	auth := &AuthManager{store: store}
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()

	auth.handleLogout(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Empty(t, store.deleteKeys)
	require.Equal(t, "", requireCookie(t, rec.Result().Cookies(), "lango_session").Value)
}

func TestAuthGenerateRandomStringProperties(t *testing.T) {
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
		store: newAuthLoginMissingProviderReturnsNotFoundSessionStore(),
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
