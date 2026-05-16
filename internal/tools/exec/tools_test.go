package exec

import (
	"bytes"
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/toolparam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubExecutor struct{}

func (stubExecutor) ExecuteTool(context.Context, string) (string, error) {
	return "ok", nil
}

func (stubExecutor) StartBackground(string) (string, error) {
	return "bg-1", nil
}

func (stubExecutor) GetBackgroundStatus(string) (map[string]interface{}, error) {
	return map[string]interface{}{"status": "running"}, nil
}

func (stubExecutor) StopBackground(string) error {
	return nil
}

func TestBuildTools_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	tools := BuildTools(stubExecutor{})

	testCases := []struct {
		name       string
		toolName   string
		params     map[string]interface{}
		wantResult interface{}
		wantErr    *toolparam.ErrMissingParam
	}{
		{
			name:       "exec requires command",
			toolName:   "exec",
			params:     map[string]interface{}{},
			wantResult: nil,
			wantErr:    &toolparam.ErrMissingParam{Name: "command"},
		},
		{
			name:       "exec background requires command",
			toolName:   "exec_bg",
			params:     map[string]interface{}{},
			wantResult: nil,
			wantErr:    &toolparam.ErrMissingParam{Name: "command"},
		},
		{
			name:       "exec status requires id",
			toolName:   "exec_status",
			params:     map[string]interface{}{},
			wantResult: nil,
			wantErr:    &toolparam.ErrMissingParam{Name: "id"},
		},
		{
			name:       "exec stop requires id",
			toolName:   "exec_stop",
			params:     map[string]interface{}{},
			wantResult: nil,
			wantErr:    &toolparam.ErrMissingParam{Name: "id"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tool := findExecTool(t, tools, tc.toolName)
			result, err := tool.Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Equal(t, tc.wantResult, result)
			var got *toolparam.ErrMissingParam
			assert.ErrorAs(t, err, &got)
			require.NotNil(t, got)
			assert.Equal(t, tc.wantErr.Name, got.Name)
		})
	}
}

func TestWarnFallbackOnce_WritesOnlyOneWarning(t *testing.T) {
	origWriter := execWarningWriter
	t.Cleanup(func() { execWarningWriter = origWriter })

	var errBuf bytes.Buffer
	execWarningWriter = &errBuf

	tool := New(Config{})
	tool.warnFallbackOnce("first reason")
	tool.warnFallbackOnce("second reason")

	assert.Contains(t, errBuf.String(), "lango: WARNING — sandbox fallback active (reason: first reason); commands run unsandboxed")
	assert.NotContains(t, errBuf.String(), "second reason")
}

func findExecTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}
