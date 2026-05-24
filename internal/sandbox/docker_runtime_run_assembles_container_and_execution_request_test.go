package sandbox

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerRuntimeRunAssemblesContainerAndExecutionRequest(t *testing.T) {
	api := newDockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake(t)
	api.attachResponse = []byte(`{"output":{"answer":42,"ok":true}}`)

	rt := &DockerRuntime{cli: api.client}
	result, err := rt.Run(context.Background(), ContainerConfig{
		Image:          "lango-worker:test",
		ToolName:       "calc",
		NetworkMode:    "none",
		Params:         map[string]interface{}{"input": "hello", "limit": 7},
		MemoryLimitMB:  64,
		CPUQuotaUS:     25000,
		ReadOnlyRootfs: true,
		Timeout:        time.Second,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, float64(42), result.Output["answer"])
	assert.Equal(t, true, result.Output["ok"])

	snapshot := api.snapshot()
	require.NotNil(t, snapshot.createRequest.Config)
	require.NotNil(t, snapshot.createRequest.HostConfig)
	assert.Equal(t, "lango-worker:test", snapshot.createRequest.Image)
	assert.Equal(t, []string{"--sandbox-worker"}, []string(snapshot.createRequest.Cmd))
	assert.True(t, snapshot.createRequest.OpenStdin)
	assert.True(t, snapshot.createRequest.StdinOnce)
	assert.True(t, snapshot.createRequest.AttachStdin)
	assert.True(t, snapshot.createRequest.AttachStdout)
	assert.True(t, snapshot.createRequest.AttachStderr)
	assert.Equal(t, map[string]string{
		"lango.sandbox": "true",
		"lango.tool":    "calc",
	}, snapshot.createRequest.Labels)
	assert.Equal(t, container.NetworkMode("none"), snapshot.createRequest.HostConfig.NetworkMode)
	assert.Equal(t, int64(64*1024*1024), snapshot.createRequest.HostConfig.Memory)
	assert.Equal(t, int64(25000), snapshot.createRequest.HostConfig.CPUQuota)
	assert.True(t, snapshot.createRequest.HostConfig.ReadonlyRootfs)
	assert.Equal(t, "", snapshot.createRequest.HostConfig.Tmpfs["/tmp"])

	assert.Equal(t, "1", snapshot.attachQuery.Get("stream"))
	assert.Equal(t, "1", snapshot.attachQuery.Get("stdin"))
	assert.Equal(t, "1", snapshot.attachQuery.Get("stdout"))
	assert.Equal(t, "1", snapshot.attachQuery.Get("stderr"))
	assert.Equal(t, ExecutionRequest{
		Version:  1,
		ToolName: "calc",
		Params: map[string]interface{}{
			"input": "hello",
			"limit": float64(7),
		},
	}, snapshot.attachRequest)
	assert.Equal(t, 1, snapshot.startCalls)
	assert.Equal(t, 1, snapshot.waitCalls)
	assert.Equal(t, 1, snapshot.removeCalls)
}

func TestDockerRuntimeRunParsesDockerFramedToolError(t *testing.T) {
	api := newDockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake(t)
	api.attachResponse = dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerFrame([]byte(`{"output":{"partial":true},"error":"denied"}`))

	rt := &DockerRuntime{cli: api.client}
	result, err := rt.Run(context.Background(), ContainerConfig{
		Image:       "lango-worker:test",
		ToolName:    "danger",
		NetworkMode: "none",
		Params:      map[string]interface{}{"path": "/private"},
	})
	require.Error(t, err)
	require.NotNil(t, result)

	assert.Contains(t, err.Error(), `tool "danger": denied`)
	assert.Equal(t, true, result.Output["partial"])
	assert.Equal(t, 1, api.snapshot().removeCalls)
}

func TestDockerRuntimeRunReturnsOOMAndRemovesContainer(t *testing.T) {
	api := newDockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake(t)
	api.attachResponse = []byte(`{"output":{"ok":true}}`)
	api.waitStatusCode = 137

	rt := &DockerRuntime{cli: api.client}
	result, err := rt.Run(context.Background(), ContainerConfig{
		Image:       "lango-worker:test",
		ToolName:    "memory-hog",
		NetworkMode: "none",
	})
	require.ErrorIs(t, err, ErrContainerOOM)

	assert.Nil(t, result)
	assert.Equal(t, 1, api.snapshot().removeCalls)
}

func TestDockerRuntimeRunReportsCreateError(t *testing.T) {
	api := newDockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake(t)
	api.createStatus = http.StatusInternalServerError

	rt := &DockerRuntime{cli: api.client}
	result, err := rt.Run(context.Background(), ContainerConfig{
		Image:    "missing:test",
		ToolName: "noop",
	})
	require.Error(t, err)

	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "create container")
	assert.Equal(t, 0, api.snapshot().removeCalls)
}

func TestDockerRuntimeRunReportsMalformedResultAndRemovesContainer(t *testing.T) {
	api := newDockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake(t)
	api.attachResponse = []byte("not-json")

	rt := &DockerRuntime{cli: api.client}
	result, err := rt.Run(context.Background(), ContainerConfig{
		Image:    "lango-worker:test",
		ToolName: "bad-json",
	})
	require.Error(t, err)

	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unmarshal container result")
	assert.Contains(t, err.Error(), "raw: not-json")
	assert.Equal(t, 1, api.snapshot().removeCalls)
}

type dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake struct {
	t      *testing.T
	server *httptest.Server
	client *client.Client
	errs   chan error

	mu            sync.Mutex
	createRequest container.CreateRequest
	attachQuery   url.Values
	attachRequest ExecutionRequest
	startCalls    int
	waitCalls     int
	removeCalls   int

	containerID    string
	createStatus   int
	attachStatus   int
	startStatus    int
	waitStatus     int
	waitStatusCode int64
	attachResponse []byte
}

type dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPISnapshot struct {
	createRequest container.CreateRequest
	attachQuery   url.Values
	attachRequest ExecutionRequest
	startCalls    int
	waitCalls     int
	removeCalls   int
}

func newDockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake(t *testing.T) *dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake {
	t.Helper()

	api := &dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake{
		t:              t,
		errs:           make(chan error, 16),
		containerID:    "ragServiceRetrieveMergesSortsLimitsAndResolvesContent6-container",
		waitStatusCode: 0,
	}
	api.server = httptest.NewServer(http.HandlerFunc(api.handle))
	t.Cleanup(func() {
		api.close()
		api.assertNoAsyncErrors(t)
	})

	serverURL, err := url.Parse(api.server.URL)
	require.NoError(t, err)
	api.client, err = client.NewClientWithOpts(
		client.WithHost("tcp://"+serverURL.Host),
		client.WithVersion("1.47"),
	)
	require.NoError(t, err)

	return api
}

func (api *dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake) close() {
	if api.server != nil {
		api.server.Close()
	}
}

func (api *dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake) recordAsyncError(err error) {
	if err == nil {
		return
	}
	select {
	case api.errs <- err:
	default:
	}
}

func (api *dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake) assertNoAsyncErrors(t *testing.T) {
	t.Helper()

	for {
		select {
		case err := <-api.errs:
			require.NoError(t, err)
		default:
			return
		}
	}
}

func (api *dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake) snapshot() dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPISnapshot {
	api.mu.Lock()
	defer api.mu.Unlock()

	return dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPISnapshot{
		createRequest: api.createRequest,
		attachQuery:   api.attachQuery,
		attachRequest: api.attachRequest,
		startCalls:    api.startCalls,
		waitCalls:     api.waitCalls,
		removeCalls:   api.removeCalls,
	}
}

func (api *dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake) handle(w http.ResponseWriter, r *http.Request) {
	path := dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIPath(r.URL.Path)
	switch {
	case r.Method == http.MethodPost && path == "/containers/create":
		api.handleCreate(w, r)
	case r.Method == http.MethodPost && path == "/containers/"+api.containerID+"/attach":
		api.handleAttach(w, r)
	case r.Method == http.MethodPost && path == "/containers/"+api.containerID+"/start":
		api.handleStart(w, r)
	case r.Method == http.MethodPost && path == "/containers/"+api.containerID+"/wait":
		api.handleWait(w, r)
	case r.Method == http.MethodDelete && path == "/containers/"+api.containerID:
		api.handleRemove(w, r)
	default:
		http.Error(w, fmt.Sprintf("unexpected Docker API request %s %s", r.Method, r.URL.String()), http.StatusNotFound)
	}
}

func (api *dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake) handleCreate(w http.ResponseWriter, r *http.Request) {
	if api.createStatus != 0 {
		http.Error(w, "create failed", api.createStatus)
		return
	}

	var req container.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "decode create request: "+err.Error(), http.StatusBadRequest)
		return
	}
	api.mu.Lock()
	api.createRequest = req
	api.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	api.recordAsyncError(json.NewEncoder(w).Encode(container.CreateResponse{ID: api.containerID}))
}

func (api *dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake) handleAttach(w http.ResponseWriter, r *http.Request) {
	if api.attachStatus != 0 {
		http.Error(w, "attach failed", api.attachStatus)
		return
	}

	api.mu.Lock()
	api.attachQuery = r.URL.Query()
	api.mu.Unlock()

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "test server does not support hijacking", http.StatusInternalServerError)
		return
	}
	conn, rw, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, "hijack failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	go api.serveHijackedAttach(conn, rw)
}

func (api *dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake) serveHijackedAttach(conn net.Conn, rw *bufio.ReadWriter) {
	defer conn.Close()

	_, err := rw.WriteString("HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.raw-stream\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n")
	if err != nil {
		api.recordAsyncError(fmt.Errorf("write hijack response: %w", err))
		return
	}
	if err := rw.Flush(); err != nil {
		api.recordAsyncError(fmt.Errorf("flush hijack response: %w", err))
		return
	}

	rawRequest, err := io.ReadAll(rw)
	if !errors.Is(err, net.ErrClosed) {
		if err != nil {
			api.recordAsyncError(fmt.Errorf("read attach request: %w", err))
			return
		}
	}

	var req ExecutionRequest
	if err := json.Unmarshal(rawRequest, &req); err != nil {
		api.recordAsyncError(fmt.Errorf("decode attach request: %w", err))
		return
	}
	api.mu.Lock()
	api.attachRequest = req
	api.mu.Unlock()

	_, err = rw.Write(api.attachResponse)
	if err != nil {
		api.recordAsyncError(fmt.Errorf("write attach response: %w", err))
		return
	}
	if err := rw.Flush(); err != nil {
		api.recordAsyncError(fmt.Errorf("flush attach response: %w", err))
	}
}

func (api *dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake) handleStart(w http.ResponseWriter, _ *http.Request) {
	if api.startStatus != 0 {
		http.Error(w, "start failed", api.startStatus)
		return
	}

	api.mu.Lock()
	api.startCalls++
	api.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (api *dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake) handleWait(w http.ResponseWriter, _ *http.Request) {
	if api.waitStatus != 0 {
		http.Error(w, "wait failed", api.waitStatus)
		return
	}

	api.mu.Lock()
	api.waitCalls++
	statusCode := api.waitStatusCode
	api.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	api.recordAsyncError(json.NewEncoder(w).Encode(container.WaitResponse{StatusCode: statusCode}))
}

func (api *dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIFake) handleRemove(w http.ResponseWriter, _ *http.Request) {
	api.mu.Lock()
	api.removeCalls++
	api.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerAPIPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 2 && strings.HasPrefix(parts[1], "v") {
		return "/" + strings.Join(parts[2:], "/")
	}
	return path
}

func dockerRuntimeRunAssemblesContainerAndExecutionRequestDockerFrame(payload []byte) []byte {
	var frame bytes.Buffer
	frame.Write([]byte{1, 0, 0, 0, byte(len(payload) >> 24), byte(len(payload) >> 16), byte(len(payload) >> 8), byte(len(payload))})
	frame.Write(payload)
	return frame.Bytes()
}
