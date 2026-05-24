package bg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
)

type stubRunner struct {
	result string
	delay  time.Duration
}

func (s stubRunner) Run(ctx context.Context, sessionKey string, prompt string) (string, error) {
	if s.delay > 0 {
		select {
		case <-time.After(s.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return s.result, nil
}

func executeBgCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

type fakeClient struct {
	tasks  []Task
	task   Task
	result string

	listErr   error
	statusErr error
	resultErr error
	cancelErr error

	statusID string
	resultID string
	cancelID string
}

func (f *fakeClient) List(context.Context) ([]Task, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.tasks, nil
}

func (f *fakeClient) Status(_ context.Context, id string) (Task, error) {
	f.statusID = id
	if f.statusErr != nil {
		return Task{}, f.statusErr
	}
	return f.task, nil
}

func (f *fakeClient) Result(_ context.Context, id string) (string, error) {
	f.resultID = id
	if f.resultErr != nil {
		return "", f.resultErr
	}
	return f.result, nil
}

func (f *fakeClient) Cancel(_ context.Context, id string) error {
	f.cancelID = id
	return f.cancelErr
}

func TestBgHelpExplainsInProcessManagerScope(t *testing.T) {
	cmd := NewBgCmd(NewInProcessClientProvider(func() (*background.Manager, error) {
		t.Fatal("help should not resolve the manager")
		return nil, nil
	}))

	out, err := executeBgCommand(t, cmd, "--help")

	require.NoError(t, err)
	assert.Contains(t, out, "gateway-backed or in-process background task client")
	assert.Contains(t, out, "embedded callers")
}

func TestGatewayBgHelpExplainsAddr(t *testing.T) {
	cmd := NewGatewayCmd(func() (*bootstrap.Result, error) {
		t.Fatal("help should not resolve gateway config")
		return nil, nil
	})

	out, err := executeBgCommand(t, cmd, "--help")

	require.NoError(t, err)
	assert.Contains(t, out, "--addr")
}

func TestGatewayBgListUsesAddrAndCommandOutput(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/bg/tasks", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]any{
			"tasks": []map[string]any{
				{
					"id":        "task-123456789",
					"status":    "done",
					"prompt":    "remote prompt",
					"startedAt": "2026-05-18 10:00:00",
					"duration":  "1s",
				},
			},
		})
		require.NoError(t, err)
	}))
	defer server.Close()

	cmd := NewGatewayCmd(nil)
	out, err := executeBgCommand(t, cmd, "list", "--addr", server.URL)

	require.NoError(t, err)
	assert.Equal(t, "/api/bg/tasks", gotPath)
	assert.Contains(t, out, "STATUS")
	assert.Contains(t, out, "remote prompt")
}

func TestGatewayBgCancelPostsToAddr(t *testing.T) {
	var gotMethod, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/bg/tasks/task-123/cancel", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]any{
			"id":        "task-123",
			"cancelled": true,
		})
		require.NoError(t, err)
	}))
	defer server.Close()

	cmd := NewGatewayCmd(nil)
	out, err := executeBgCommand(t, cmd, "cancel", "task-123", "--addr", server.URL)

	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/api/bg/tasks/task-123/cancel", gotPath)
	assert.Contains(t, out, "Task task-123 cancelled.")
}

func TestGatewayBgStatusReportsGatewayErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/bg/tasks/missing-task", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		err := json.NewEncoder(w).Encode(map[string]string{
			"error": "task status: task \"missing-task\" not found",
		})
		require.NoError(t, err)
	}))
	defer server.Close()

	cmd := NewGatewayCmd(nil)
	_, err := executeBgCommand(t, cmd, "status", "missing-task", "--addr", server.URL)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "task status: task \"missing-task\" not found")
}

func TestGatewayBgResultUsesEscapedPathAndJSONResponse(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/bg/tasks/task%20with%20space/result", r.URL.EscapedPath())
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]string{
			"result": "gateway result",
		})
		require.NoError(t, err)
	}))
	defer server.Close()

	cmd := NewGatewayCmd(nil)
	out, err := executeBgCommand(t, cmd, "result", "task with space", "--addr", server.URL)

	require.NoError(t, err)
	assert.Equal(t, "/api/bg/tasks/task%20with%20space/result", gotPath)
	assert.Equal(t, "gateway result\n", out)
}

func TestBgListEmptyTableOutput(t *testing.T) {
	client := &fakeClient{}
	cmd := NewBgCmd(func() (Client, error) { return client, nil })

	out, err := executeBgCommand(t, cmd, "list")

	require.NoError(t, err)
	assert.Equal(t, "No background tasks.\n", out)
}

func TestBgCommandsJSONOutput(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		client := &fakeClient{
			tasks: []Task{{
				ID:        "task-list",
				Status:    "done",
				Prompt:    "json list",
				StartedAt: "2026-05-18 10:00:00",
				Duration:  "1s",
			}},
		}
		cmd := NewBgCmd(func() (Client, error) { return client, nil })

		out, err := executeBgCommand(t, cmd, "list", "--output", "json")

		require.NoError(t, err)
		assert.JSONEq(t, `{
				"tasks": [{
					"id": "task-list",
					"status": "done",
					"prompt": "json list",
					"startedAt": "2026-05-18 10:00:00",
					"duration": "1s"
				}]
			}`, out)
	})

	t.Run("status", func(t *testing.T) {
		client := &fakeClient{
			task: Task{
				ID:            "task-status",
				Status:        "failed",
				Prompt:        "json status",
				OriginChannel: "telegram",
				OriginSession: "sess-1",
				Error:         "failed hard",
			},
		}
		cmd := NewBgCmd(func() (Client, error) { return client, nil })

		out, err := executeBgCommand(t, cmd, "status", "task-status", "--output", "json")

		require.NoError(t, err)
		assert.Equal(t, "task-status", client.statusID)
		assert.JSONEq(t, `{
				"task": {
					"id": "task-status",
					"status": "failed",
					"prompt": "json status",
					"originChannel": "telegram",
					"originSession": "sess-1",
					"error": "failed hard"
				}
			}`, out)
	})

	t.Run("cancel", func(t *testing.T) {
		client := &fakeClient{}
		cmd := NewBgCmd(func() (Client, error) { return client, nil })

		out, err := executeBgCommand(t, cmd, "cancel", "task-cancel", "--output", "json")

		require.NoError(t, err)
		assert.Equal(t, "task-cancel", client.cancelID)
		assert.JSONEq(t, `{"id":"task-cancel","cancelled":true}`, out)
	})

	t.Run("result", func(t *testing.T) {
		client := &fakeClient{result: "json result"}
		cmd := NewBgCmd(func() (Client, error) { return client, nil })

		out, err := executeBgCommand(t, cmd, "result", "task-result", "--output", "json")

		require.NoError(t, err)
		assert.Equal(t, "task-result", client.resultID)
		assert.JSONEq(t, `{"result":"json result"}`, out)
	})
}

func TestBgCommandErrorBranches(t *testing.T) {
	t.Run("client provider error", func(t *testing.T) {
		cmd := NewBgCmd(func() (Client, error) {
			return nil, errors.New("provider down")
		})

		_, err := executeBgCommand(t, cmd, "list")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "get background client: provider down")
	})

	t.Run("list error", func(t *testing.T) {
		client := &fakeClient{listErr: errors.New("list down")}
		cmd := NewBgCmd(func() (Client, error) { return client, nil })

		_, err := executeBgCommand(t, cmd, "list")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "list tasks: list down")
	})

	t.Run("status error", func(t *testing.T) {
		client := &fakeClient{statusErr: errors.New("status down")}
		cmd := NewBgCmd(func() (Client, error) { return client, nil })

		_, err := executeBgCommand(t, cmd, "status", "task-status")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "get status: status down")
	})

	t.Run("cancel error", func(t *testing.T) {
		client := &fakeClient{cancelErr: errors.New("cancel down")}
		cmd := NewBgCmd(func() (Client, error) { return client, nil })

		_, err := executeBgCommand(t, cmd, "cancel", "task-cancel")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cancel task: cancel down")
	})

	t.Run("result error", func(t *testing.T) {
		client := &fakeClient{resultErr: errors.New("result down")}
		cmd := NewBgCmd(func() (Client, error) { return client, nil })

		_, err := executeBgCommand(t, cmd, "result", "task-result")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "get result: result down")
	})

	t.Run("invalid output", func(t *testing.T) {
		client := &fakeClient{}
		cmd := NewBgCmd(func() (Client, error) { return client, nil })

		_, err := executeBgCommand(t, cmd, "list", "--output", "xml")

		require.Error(t, err)
		assert.Contains(t, err.Error(), `unknown output format "xml"`)
	})
}

func TestGatewayBgCommandResolvesConfiguredAddressAndLoaderErrors(t *testing.T) {
	t.Run("configured address", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/api/bg/tasks", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"tasks": []any{}}))
		}))
		defer server.Close()

		serverURL := strings.TrimPrefix(server.URL, "http://")
		host, port, ok := strings.Cut(serverURL, ":")
		require.True(t, ok)
		cfg := config.DefaultConfig()
		cfg.Server.Host = host
		_, err := fmt.Sscanf(port, "%d", &cfg.Server.Port)
		require.NoError(t, err)

		cmd := NewGatewayCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{Config: cfg}, nil
		})
		out, err := executeBgCommand(t, cmd, "list")

		require.NoError(t, err)
		assert.Equal(t, "No background tasks.\n", out)
	})

	t.Run("missing loader", func(t *testing.T) {
		cmd := NewGatewayCmd(nil)

		_, err := executeBgCommand(t, cmd, "list")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "gateway address is required")
	})

	t.Run("loader error", func(t *testing.T) {
		cmd := NewGatewayCmd(func() (*bootstrap.Result, error) {
			return nil, errors.New("boom")
		})

		_, err := executeBgCommand(t, cmd, "list")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "load config for gateway address: boom")
	})

	t.Run("nil config", func(t *testing.T) {
		cmd := NewGatewayCmd(func() (*bootstrap.Result, error) {
			return &bootstrap.Result{}, nil
		})

		_, err := executeBgCommand(t, cmd, "list")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "config is unavailable")
	})
}

func TestInProcessClientManagerErrorBranches(t *testing.T) {
	t.Run("nil manager provider", func(t *testing.T) {
		client, err := NewInProcessClientProvider(nil)()
		require.NoError(t, err)

		_, err = client.List(context.Background())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "background manager provider is not configured")
	})

	t.Run("manager provider error reaches every operation", func(t *testing.T) {
		client, err := NewInProcessClientProvider(func() (*background.Manager, error) {
			return nil, errors.New("manager down")
		})()
		require.NoError(t, err)

		_, listErr := client.List(context.Background())
		_, statusErr := client.Status(context.Background(), "task-status")
		_, resultErr := client.Result(context.Background(), "task-result")
		cancelErr := client.Cancel(context.Background(), "task-cancel")

		for _, err := range []error{listErr, statusErr, resultErr, cancelErr} {
			require.Error(t, err)
			assert.Contains(t, err.Error(), "manager down")
		}
	})
}

func TestBgList_WritesToCommandOutput(t *testing.T) {
	mgr := background.NewManager(stubRunner{result: "done"}, nil, 5, time.Minute, zap.NewNop().Sugar())
	_, err := mgr.Submit(context.Background(), "test prompt", background.Origin{Channel: "cli"})
	require.NoError(t, err)
	time.Sleep(20 * time.Millisecond)

	cmd := NewBgCmd(NewInProcessClientProvider(func() (*background.Manager, error) { return mgr, nil }))
	out, err := executeBgCommand(t, cmd, "list")

	require.NoError(t, err)
	assert.Contains(t, out, "STATUS")
	assert.Contains(t, out, "test prompt")
}

func TestBgListFormatsTableFields(t *testing.T) {
	longPrompt := strings.Repeat("a", 60)
	client := &fakeClient{
		tasks: []Task{{
			ID:     "123456789",
			Status: "queued",
			Prompt: longPrompt,
		}},
	}
	cmd := NewBgCmd(func() (Client, error) { return client, nil })

	out, err := executeBgCommand(t, cmd, "list")

	require.NoError(t, err)
	assert.Contains(t, out, "12345678")
	assert.Contains(t, out, strings.Repeat("a", 47)+"...")
	assert.Contains(t, out, "queued")
	assert.Contains(t, out, "-")
}

func TestBgStatus_WritesToCommandOutput(t *testing.T) {
	mgr := background.NewManager(stubRunner{result: "done"}, nil, 5, time.Minute, zap.NewNop().Sugar())
	id, err := mgr.Submit(context.Background(), "inspect me", background.Origin{Channel: "telegram", Session: "sess-1"})
	require.NoError(t, err)
	time.Sleep(20 * time.Millisecond)

	cmd := NewBgCmd(NewInProcessClientProvider(func() (*background.Manager, error) { return mgr, nil }))
	out, err := executeBgCommand(t, cmd, "status", id)

	require.NoError(t, err)
	assert.Contains(t, out, "ID:")
	assert.Contains(t, out, "inspect me")
	assert.Contains(t, out, "telegram")
}

func TestBgResultAndCancel_WriteToCommandOutput(t *testing.T) {
	doneMgr := background.NewManager(stubRunner{result: "hello world"}, nil, 5, time.Minute, zap.NewNop().Sugar())
	doneID, err := doneMgr.Submit(context.Background(), "done prompt", background.Origin{})
	require.NoError(t, err)
	time.Sleep(120 * time.Millisecond)

	resultCmd := NewBgCmd(NewInProcessClientProvider(func() (*background.Manager, error) { return doneMgr, nil }))
	resultOut, err := executeBgCommand(t, resultCmd, "result", doneID)
	require.NoError(t, err)
	assert.Contains(t, resultOut, "hello world")

	runningMgr := background.NewManager(stubRunner{delay: time.Second}, nil, 5, time.Minute, zap.NewNop().Sugar())
	cancelID, err := runningMgr.Submit(context.Background(), "cancel me", background.Origin{})
	require.NoError(t, err)
	time.Sleep(20 * time.Millisecond)

	cancelCmd := NewBgCmd(NewInProcessClientProvider(func() (*background.Manager, error) { return runningMgr, nil }))
	cancelOut, err := executeBgCommand(t, cancelCmd, "cancel", cancelID)
	require.NoError(t, err)
	assert.Contains(t, cancelOut, "cancelled")
	assert.Contains(t, cancelOut, cancelID)
}
