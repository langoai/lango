package workflow

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

func newWorkflowToolEngine(t *testing.T) *Engine {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:workflow-tools?mode=memory&_fk=1")
	t.Cleanup(func() { client.Close() })
	state := NewStateStore(client, zap.NewNop().Sugar())
	runner := &mockAgentRunner{result: "ok"}
	return NewEngine(runner, state, nil, 1, 0, zap.NewNop().Sugar())
}

func findWorkflowTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("tool %s not found", name)
	return nil
}

func TestWorkflowRun_RequiresFileOrInlineYAML(t *testing.T) {
	t.Parallel()

	engine := newWorkflowToolEngine(t)
	tool := findWorkflowTool(t, BuildTools(engine, t.TempDir(), nil), "workflow_run")

	got, err := tool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "either file_path or yaml_content is required")
}

func TestWorkflowRun_RejectsFileAndInlineYAMLTogether(t *testing.T) {
	t.Parallel()

	engine := newWorkflowToolEngine(t)
	tool := findWorkflowTool(t, BuildTools(engine, t.TempDir(), nil), "workflow_run")

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"file_path":    "workflow.yaml",
		"yaml_content": "name: sample\nsteps:\n  - id: first\n    agent: executor\n    prompt: hello\n",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "file_path and yaml_content are mutually exclusive")
}

func TestWorkflowRun_ReturnsAsyncRunReceiptForInlineYAML(t *testing.T) {
	t.Parallel()

	engine := newWorkflowToolEngine(t)
	tool := findWorkflowTool(t, BuildTools(engine, t.TempDir(), nil), "workflow_run")

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"yaml_content": "name: sample\nsteps:\n  - id: first\n    agent: executor\n    prompt: hello\n",
	})
	require.NoError(t, err)

	payload, ok := got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "running", payload["status"])
	assert.NotEmpty(t, payload["run_id"])
	assert.Contains(t, payload["message"], "Workflow 'sample' started")
	_, hasResults := payload["results"]
	assert.False(t, hasResults)
}

func TestWorkflowTools_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	engine := newWorkflowToolEngine(t)
	tools := BuildTools(engine, t.TempDir(), nil)

	testCases := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name:    "status requires run id",
			tool:    "workflow_status",
			params:  map[string]interface{}{},
			wantErr: "missing run_id parameter",
		},
		{
			name:    "cancel requires run id",
			tool:    "workflow_cancel",
			params:  map[string]interface{}{},
			wantErr: "missing run_id parameter",
		},
		{
			name:    "save requires name",
			tool:    "workflow_save",
			params:  map[string]interface{}{"yaml_content": "name: sample\nsteps:\n  - id: first\n    agent: executor\n    prompt: hello\n"},
			wantErr: "missing name parameter",
		},
		{
			name:    "save requires yaml content",
			tool:    "workflow_save",
			params:  map[string]interface{}{"name": "sample"},
			wantErr: "missing yaml_content parameter",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := findWorkflowTool(t, tools, tc.tool).Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}
