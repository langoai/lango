package workflow

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeWorkflowCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func writeWorkflowFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sample.flow.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestWorkflowValidate_WritesTableOutputToCommandWriter(t *testing.T) {
	filePath := writeWorkflowFile(t, `
name: daily-report
schedule: "0 9 * * *"
steps:
  - id: collect
    agent: operator
    prompt: Collect the report
`)

	cmd := newValidateCmd()
	out, err := executeWorkflowCommand(t, cmd, filePath)

	require.NoError(t, err)
	assert.Contains(t, out, "Workflow")
	assert.Contains(t, out, "daily-report")
	assert.Contains(t, out, "Schedule:")
}

func TestWorkflowValidate_WritesJSONOutputToCommandWriter(t *testing.T) {
	filePath := writeWorkflowFile(t, `
name: daily-report
steps:
  - id: collect
    agent: operator
    prompt: Collect the report
`)

	cmd := newValidateCmd()
	out, err := executeWorkflowCommand(t, cmd, filePath, "--output", "json")

	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Equal(t, true, payload["valid"])
	assert.Equal(t, "daily-report", payload["name"])
}

func TestWorkflowValidate_InvalidOutputFailsBeforeParsing(t *testing.T) {
	cmd := newValidateCmd()
	out, err := executeWorkflowCommand(t, cmd, "sample.flow.yaml", "--output", "yaml")

	require.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}
