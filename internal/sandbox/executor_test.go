package sandbox

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInProcessExecutor_Execute(t *testing.T) {
	tests := []struct {
		give       string
		giveParams map[string]interface{}
		wantResult map[string]interface{}
		wantErr    bool
	}{
		{
			give:       "echo",
			giveParams: map[string]interface{}{"msg": "hello"},
			wantResult: map[string]interface{}{"msg": "hello"},
		},
		{
			give:       "empty",
			giveParams: nil,
			wantResult: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			exec := NewInProcessExecutor(func(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
				if params == nil {
					return map[string]interface{}{}, nil
				}
				return params, nil
			})

			result, err := exec.Execute(context.Background(), tt.give, tt.giveParams)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}

func TestExecutionRequest_JSON(t *testing.T) {
	req := ExecutionRequest{
		ToolName: "search",
		Params:   map[string]interface{}{"query": "test", "limit": float64(10)},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded ExecutionRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, req.ToolName, decoded.ToolName)
	assert.Equal(t, req.Params["query"], decoded.Params["query"])
	assert.Equal(t, req.Params["limit"], decoded.Params["limit"])
}

func TestExecutionResult_JSON(t *testing.T) {
	tests := []struct {
		give     string
		giveData ExecutionResult
	}{
		{
			give: "success",
			giveData: ExecutionResult{
				Output: map[string]interface{}{"status": "ok", "count": float64(42)},
			},
		},
		{
			give: "error",
			giveData: ExecutionResult{
				Error: "tool not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			data, err := json.Marshal(tt.giveData)
			require.NoError(t, err)

			var decoded ExecutionResult
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Equal(t, tt.giveData.Output, decoded.Output)
			assert.Equal(t, tt.giveData.Error, decoded.Error)
		})
	}
}

func TestSubprocessExecutor_Timeout(t *testing.T) {
	exec := NewSubprocessExecutor(Config{
		TimeoutPerTool: 1 * time.Millisecond,
	})

	// Use a context that is already expired to force immediate timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	time.Sleep(5 * time.Millisecond) // ensure context is expired

	_, err := exec.Execute(ctx, "slow-tool", map[string]interface{}{})
	require.Error(t, err)
	// The error should indicate timeout or context cancellation.
	assert.Contains(t, err.Error(), "timed out")
}

func TestIsWorkerMode(t *testing.T) {
	// IsWorkerMode checks os.Args, which we cannot safely mutate in parallel tests.
	// Just verify the function exists and returns false in normal test mode.
	assert.False(t, IsWorkerMode())
}

func TestCleanEnv(t *testing.T) {
	env := cleanEnv()
	// Should contain at most PATH and HOME.
	assert.LessOrEqual(t, len(env), 2)
	for _, e := range env {
		assert.True(t, len(e) > 0)
		// Each entry should be either PATH= or HOME=.
		assert.Regexp(t, `^(PATH|HOME)=`, e)
	}
}

func TestConfig_Defaults(t *testing.T) {
	cfg := Config{}
	assert.False(t, cfg.Enabled)
	assert.Equal(t, time.Duration(0), cfg.TimeoutPerTool)
	assert.Equal(t, 0, cfg.MaxMemoryMB)
}
