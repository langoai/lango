package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunWorkerWithIO_DecodeFailure(t *testing.T) {
	var out bytes.Buffer

	exitCode := RunWorkerWithIO(nil, bytes.NewBufferString(`{`), &out)

	assert.Equal(t, 1, exitCode)
	result := decodeWorkerResult(t, &out)
	assert.Contains(t, result.Error, "decode request")
	assert.Empty(t, result.Output)
}

func TestRunWorkerWithIO_UnregisteredTool(t *testing.T) {
	var out bytes.Buffer
	in := encodeWorkerRequest(t, ExecutionRequest{ToolName: "missing-tool"})

	exitCode := RunWorkerWithIO(nil, in, &out)

	assert.Equal(t, 1, exitCode)
	result := decodeWorkerResult(t, &out)
	assert.Contains(t, result.Error, "not registered")
	assert.Empty(t, result.Output)
}

func TestRunWorkerWithIO_HandlerError(t *testing.T) {
	var out bytes.Buffer
	in := encodeWorkerRequest(t, ExecutionRequest{ToolName: "failing-tool"})
	registry := ToolRegistry{
		"failing-tool": func(context.Context, map[string]interface{}) (interface{}, error) {
			return nil, errors.New("tool failed")
		},
	}

	exitCode := RunWorkerWithIO(registry, in, &out)

	assert.Equal(t, 0, exitCode)
	result := decodeWorkerResult(t, &out)
	assert.Equal(t, "tool failed", result.Error)
	assert.Empty(t, result.Output)
}

func TestRunWorkerWithIO_ScalarSuccess(t *testing.T) {
	var out bytes.Buffer
	in := encodeWorkerRequest(t, ExecutionRequest{ToolName: "scalar-tool"})
	registry := ToolRegistry{
		"scalar-tool": func(context.Context, map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}

	exitCode := RunWorkerWithIO(registry, in, &out)

	assert.Equal(t, 0, exitCode)
	result := decodeWorkerResult(t, &out)
	assert.Empty(t, result.Error)
	assert.Equal(t, map[string]interface{}{"result": "ok"}, result.Output)
}

func TestRunWorkerWithIO_MapSuccess(t *testing.T) {
	var out bytes.Buffer
	in := encodeWorkerRequest(t, ExecutionRequest{
		ToolName: "map-tool",
		Params:   map[string]interface{}{"input": "hello"},
	})
	registry := ToolRegistry{
		"map-tool": func(_ context.Context, params map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"echo": params["input"]}, nil
		},
	}

	exitCode := RunWorkerWithIO(registry, in, &out)

	assert.Equal(t, 0, exitCode)
	result := decodeWorkerResult(t, &out)
	assert.Empty(t, result.Error)
	assert.Equal(t, map[string]interface{}{"echo": "hello"}, result.Output)
}

func TestRunWorker_UsesPublicStdioSeams(t *testing.T) {
	restoreWorkerStdioSeams(t)

	var out bytes.Buffer
	workerStdin = encodeWorkerRequest(t, ExecutionRequest{
		ToolName: "map-tool",
		Params:   map[string]interface{}{"input": "hello"},
	})
	workerStdout = &out
	registry := ToolRegistry{
		"map-tool": func(_ context.Context, params map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"echo": params["input"]}, nil
		},
	}

	exitCode := RunWorker(registry)

	assert.Equal(t, 0, exitCode)
	result := decodeWorkerResult(t, &out)
	assert.Empty(t, result.Error)
	assert.Equal(t, map[string]interface{}{"echo": "hello"}, result.Output)
}

func restoreWorkerStdioSeams(t *testing.T) {
	t.Helper()

	originalStdin := workerStdin
	originalStdout := workerStdout
	t.Cleanup(func() {
		workerStdin = originalStdin
		workerStdout = originalStdout
	})
}

func encodeWorkerRequest(t *testing.T, req ExecutionRequest) *bytes.Buffer {
	t.Helper()

	var in bytes.Buffer
	require.NoError(t, json.NewEncoder(&in).Encode(req))
	return &in
}

func decodeWorkerResult(t *testing.T, out *bytes.Buffer) ExecutionResult {
	t.Helper()

	var result ExecutionResult
	decoder := json.NewDecoder(out)
	require.NoError(t, decoder.Decode(&result))
	require.ErrorIs(t, decoder.Decode(&ExecutionResult{}), io.EOF, "expected one result")
	return result
}
