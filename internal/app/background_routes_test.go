package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/gateway"
	"github.com/langoai/lango/internal/session"
)

func TestBackgroundRoutes_ListStatusAndResult(t *testing.T) {
	t.Parallel()

	runner := &backgroundRouteRunner{result: "finished output"}
	mgr := background.NewManager(runner, nil, 5, time.Minute, zap.NewNop().Sugar())
	cleanupBackgroundRouteManager(t, mgr)
	taskID, err := mgr.Submit(context.Background(), "summarize this", background.Origin{
		Channel: "test",
		Session: "sess-1",
	})
	require.NoError(t, err)
	waitForBackgroundStatus(t, mgr, taskID, background.Done)

	router := newBackgroundRouteTestRouter(&App{BackgroundManager: mgr}, nil)

	listRec := performBackgroundRouteRequest(router, http.MethodGet, "/api/bg/tasks")
	require.Equal(t, http.StatusOK, listRec.Code)

	var listResp struct {
		Tasks []map[string]any `json:"tasks"`
	}
	require.NoError(t, json.NewDecoder(listRec.Body).Decode(&listResp))
	require.Len(t, listResp.Tasks, 1)
	assert.Equal(t, taskID, listResp.Tasks[0]["id"])
	assert.Equal(t, "done", listResp.Tasks[0]["status"])
	assert.IsType(t, "", listResp.Tasks[0]["status"])
	assert.NotContains(t, listResp.Tasks[0], "status_text")
	assert.Equal(t, "test", listResp.Tasks[0]["originChannel"])
	assert.Equal(t, "sess-1", listResp.Tasks[0]["originSession"])

	statusRec := performBackgroundRouteRequest(router, http.MethodGet, "/api/bg/tasks/"+taskID)
	require.Equal(t, http.StatusOK, statusRec.Code)

	var statusResp struct {
		Task map[string]any `json:"task"`
	}
	require.NoError(t, json.NewDecoder(statusRec.Body).Decode(&statusResp))
	assert.Equal(t, taskID, statusResp.Task["id"])
	assert.Equal(t, "done", statusResp.Task["status"])
	assert.Equal(t, "summarize this", statusResp.Task["prompt"])

	resultRec := performBackgroundRouteRequest(router, http.MethodGet, "/api/bg/tasks/"+taskID+"/result")
	require.Equal(t, http.StatusOK, resultRec.Code)

	var resultResp struct {
		Result string `json:"result"`
	}
	require.NoError(t, json.NewDecoder(resultRec.Body).Decode(&resultResp))
	assert.Equal(t, "finished output", resultResp.Result)
}

func TestBackgroundRoutes_Cancel(t *testing.T) {
	t.Parallel()

	runner := newBlockingBackgroundRouteRunner()
	mgr := background.NewManager(runner, nil, 5, time.Minute, zap.NewNop().Sugar())
	cleanupBackgroundRouteManager(t, mgr)
	taskID, err := mgr.Submit(context.Background(), "keep running", background.Origin{})
	require.NoError(t, err)
	require.Eventually(t, runner.HasStarted, time.Second, 10*time.Millisecond)

	router := newBackgroundRouteTestRouter(&App{BackgroundManager: mgr}, nil)
	cancelRec := performBackgroundRouteRequest(router, http.MethodPost, "/api/bg/tasks/"+taskID+"/cancel")
	require.Equal(t, http.StatusOK, cancelRec.Code)

	var cancelResp struct {
		ID        string `json:"id"`
		Cancelled bool   `json:"cancelled"`
	}
	require.NoError(t, json.NewDecoder(cancelRec.Body).Decode(&cancelResp))
	assert.Equal(t, taskID, cancelResp.ID)
	assert.True(t, cancelResp.Cancelled)
	waitForBackgroundStatus(t, mgr, taskID, background.Cancelled)
}

func TestBackgroundRoutes_ErrorResponses(t *testing.T) {
	t.Parallel()

	unavailableRouter := newBackgroundRouteTestRouter(&App{}, nil)
	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/bg/tasks"},
		{method: http.MethodGet, path: "/api/bg/tasks/missing"},
		{method: http.MethodGet, path: "/api/bg/tasks/missing/result"},
		{method: http.MethodPost, path: "/api/bg/tasks/missing/cancel"},
	} {
		rec := performBackgroundRouteRequest(unavailableRouter, tc.method, tc.path)
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code, "%s %s", tc.method, tc.path)
	}

	runner := newBlockingBackgroundRouteRunner()
	mgr := background.NewManager(runner, nil, 5, time.Minute, zap.NewNop().Sugar())
	cleanupBackgroundRouteManager(t, mgr)
	taskID, err := mgr.Submit(context.Background(), "not done yet", background.Origin{})
	require.NoError(t, err)
	require.Eventually(t, runner.HasStarted, time.Second, 10*time.Millisecond)

	router := newBackgroundRouteTestRouter(&App{BackgroundManager: mgr}, nil)

	statusRec := performBackgroundRouteRequest(router, http.MethodGet, "/api/bg/tasks/missing")
	assert.Equal(t, http.StatusNotFound, statusRec.Code)

	missingResultRec := performBackgroundRouteRequest(router, http.MethodGet, "/api/bg/tasks/missing/result")
	assert.GreaterOrEqual(t, missingResultRec.Code, 400)

	notDoneResultRec := performBackgroundRouteRequest(router, http.MethodGet, "/api/bg/tasks/"+taskID+"/result")
	assert.GreaterOrEqual(t, notDoneResultRec.Code, 400)
}

func TestBackgroundRoutes_RequireAuthWhenConfigured(t *testing.T) {
	t.Parallel()

	mgr := background.NewManager(&backgroundRouteRunner{result: "ok"}, nil, 5, time.Minute, zap.NewNop().Sugar())
	cleanupBackgroundRouteManager(t, mgr)
	auth, err := gateway.NewAuthManager(config.AuthConfig{}, &backgroundRouteSessionStore{})
	require.NoError(t, err)

	router := newBackgroundRouteTestRouter(&App{BackgroundManager: mgr}, auth)
	rec := performBackgroundRouteRequest(router, http.MethodGet, "/api/bg/tasks")

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func newBackgroundRouteTestRouter(app *App, auth *gateway.AuthManager) chi.Router {
	router := chi.NewRouter()
	registerBackgroundRoutes(router, app, auth)
	return router
}

func performBackgroundRouteRequest(router http.Handler, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	router.ServeHTTP(rec, req)
	return rec
}

func waitForBackgroundStatus(t *testing.T, mgr *background.Manager, taskID string, status background.Status) {
	t.Helper()

	require.Eventually(t, func() bool {
		snap, err := mgr.Status(taskID)
		return err == nil && snap.Status == status
	}, time.Second, 10*time.Millisecond)
}

func cleanupBackgroundRouteManager(t *testing.T, mgr *background.Manager) {
	t.Helper()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = mgr.Shutdown(ctx)
	})
}

type backgroundRouteRunner struct {
	result string
}

func (r *backgroundRouteRunner) Run(context.Context, string, string) (string, error) {
	return r.result, nil
}

type blockingBackgroundRouteRunner struct {
	started chan struct{}
	once    sync.Once
}

func newBlockingBackgroundRouteRunner() *blockingBackgroundRouteRunner {
	return &blockingBackgroundRouteRunner{started: make(chan struct{})}
}

func (r *blockingBackgroundRouteRunner) Run(ctx context.Context, _, _ string) (string, error) {
	r.once.Do(func() { close(r.started) })
	<-ctx.Done()
	return "", ctx.Err()
}

func (r *blockingBackgroundRouteRunner) HasStarted() bool {
	select {
	case <-r.started:
		return true
	default:
		return false
	}
}

type backgroundRouteSessionStore struct {
	sessions map[string]*session.Session
}

func (s *backgroundRouteSessionStore) Create(sess *session.Session) error {
	if s.sessions == nil {
		s.sessions = make(map[string]*session.Session)
	}
	s.sessions[sess.Key] = sess
	return nil
}

func (s *backgroundRouteSessionStore) Get(key string) (*session.Session, error) {
	if s.sessions == nil {
		return nil, nil
	}
	return s.sessions[key], nil
}

func (s *backgroundRouteSessionStore) Update(sess *session.Session) error {
	return s.Create(sess)
}

func (s *backgroundRouteSessionStore) Delete(key string) error {
	delete(s.sessions, key)
	return nil
}

func (s *backgroundRouteSessionStore) AppendMessage(string, session.Message) error { return nil }
func (s *backgroundRouteSessionStore) AnnotateTimeout(string, string) error        { return nil }
func (s *backgroundRouteSessionStore) End(string) error                            { return nil }
func (s *backgroundRouteSessionStore) Close() error                                { return nil }
func (s *backgroundRouteSessionStore) ListSessions(context.Context) ([]session.SessionSummary, error) {
	return nil, nil
}
func (s *backgroundRouteSessionStore) GetSalt(string) ([]byte, error) { return nil, nil }
func (s *backgroundRouteSessionStore) SetSalt(string, []byte) error   { return nil }
