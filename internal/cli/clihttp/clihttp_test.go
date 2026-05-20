package clihttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
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

func TestResolveGatewayAddrFormatsConfiguredIPv6Host(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()
	cfg.Server.Host = "::1"
	cfg.Server.Port = 18789

	got := ResolveGatewayAddr("", cfg)

	assert.Equal(t, "http://[::1]:18789", got)
}

func TestPostJSONContextPostsPayloadAndDecodesResponse(t *testing.T) {
	t.Parallel()

	type capturedRequest struct {
		method      string
		contentType string
		path        string
		payload     map[string]string
	}
	requests := make(chan capturedRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		requests <- capturedRequest{
			method:      r.Method,
			contentType: r.Header.Get("Content-Type"),
			path:        r.URL.Path,
			payload:     payload,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true,
		})
	}))
	defer server.Close()

	var out map[string]bool
	err := PostJSONContext(context.Background(), server.URL, "/v1/tasks", map[string]string{
		"task": "index",
	}, &out)

	require.NoError(t, err)
	assert.True(t, out["ok"])

	got := <-requests
	assert.Equal(t, http.MethodPost, got.method)
	assert.Equal(t, "application/json", got.contentType)
	assert.Equal(t, "/v1/tasks", got.path)
	assert.Equal(t, "index", got.payload["task"])
}

func TestPostJSONContextHandlesMarshalBuildStatusAndNilOutputBranches(t *testing.T) {
	t.Parallel()

	t.Run("marshal error", func(t *testing.T) {
		err := PostJSON("http://127.0.0.1", "/bad", map[string]any{"fn": func() {}}, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "marshal request")
	})

	t.Run("build request error", func(t *testing.T) {
		err := PostJSONContext(context.Background(), "://bad-url", "/path", map[string]string{"ok": "true"}, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "build gateway request")
	})

	t.Run("message error body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "gateway is warming up",
			})
		}))
		defer server.Close()

		err := PostJSON(server.URL, "/tasks", map[string]string{"task": "index"}, nil)

		require.Error(t, err)
		assert.EqualError(t, err, "gateway is warming up")
	})

	t.Run("plain status error body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("not-json"))
		}))
		defer server.Close()

		err := PostJSON(server.URL, "/tasks", map[string]string{"task": "index"}, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "gateway returned status 500")
	})

	t.Run("nil output skips decode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte("not-json"))
		}))
		defer server.Close()

		err := PostJSON(server.URL, "/tasks", map[string]string{"task": "index"}, nil)

		require.NoError(t, err)
	})
}

func TestResolveTableOrJSONOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		flag    string
		want    string
		wantErr string
	}{
		{name: "default", want: "table"},
		{name: "table trims and folds", flag: " TABLE ", want: "table"},
		{name: "json", flag: "json", want: "json"},
		{name: "unknown", flag: "yaml", wantErr: `unknown output format "yaml"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("output", tt.flag, "")

			got, err := ResolveTableOrJSONOutput(cmd)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Empty(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrintJSONWritesIndentedOutputAndPropagatesErrors(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := PrintJSON(&buf, map[string]string{"status": "ok"})

	require.NoError(t, err)
	assert.Equal(t, "{\n  \"status\": \"ok\"\n}\n", buf.String())

	err = PrintJSON(errorWriter{}, map[string]string{"status": "ok"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
