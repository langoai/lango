package bg

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/background"
	"github.com/langoai/lango/internal/bootstrap"
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
