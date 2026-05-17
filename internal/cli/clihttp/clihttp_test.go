package clihttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchJSONReturnsGatewayErrorBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]string{
			"error": "task result is not available until task is done",
		}))
	}))
	defer server.Close()

	var out map[string]any
	err := FetchJSON(server.URL, "/tasks/1/result", &out)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "task result is not available until task is done")
}

func TestFetchJSONContextCancelsRequest(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-release:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true}`))
		}
	}))
	defer server.Close()
	defer close(release)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var out map[string]any
	err := FetchJSONContext(ctx, server.URL, "/slow", &out)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "connect to gateway")
}

func TestResolveGatewayAddrTrimsExplicitTrailingSlashes(t *testing.T) {
	t.Parallel()

	got := ResolveGatewayAddr("  http://127.0.0.1:18789///  ", nil)

	assert.Equal(t, "http://127.0.0.1:18789", got)
}
