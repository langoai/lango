package agentrt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltinTeammateTypes(t *testing.T) {
	types := BuiltinTeammateTypes()

	for _, name := range []string{
		"operator",
		"navigator",
		"vault",
		"librarian",
		"automator",
		"planner",
		"chronicler",
		"ontologist",
	} {
		tt, ok := types[name]
		require.Truef(t, ok, "missing teammate type %q", name)
		assert.Equal(t, name, tt.Name)
	}

	assert.True(t, types["operator"].AllowsTool("fs_read"))
	assert.True(t, types["operator"].AllowsTool("exec"))
	assert.False(t, types["planner"].AllowsTool("exec"))
	assert.True(t, types["planner"].AllowsTool("agent_wait"))
	assert.False(t, types["planner"].AllowsTool("builtin_invoke"))
	assert.True(t, types["planner"].AllowsTool("tool_output_get"))
	assert.True(t, types["planner"].AllowsTool("builtin_list"))
	assert.True(t, types["planner"].AllowsTool("builtin_search"))
	assert.True(t, types["planner"].AllowsTool("builtin_health"))
	assert.True(t, types["librarian"].AllowsTool("graph_query"))
	assert.True(t, types["librarian"].AllowsTool("graph_traverse"))
	assert.True(t, types["librarian"].AllowsTool("librarian_pending_inquiries"))
	assert.True(t, types["librarian"].AllowsTool("librarian_dismiss_inquiry"))
}

func TestValidateAllowedToolsForTeammate(t *testing.T) {
	require.NoError(t, ValidateAllowedToolsForTeammate("operator", []string{"fs_read", "exec"}))

	err := ValidateAllowedToolsForTeammate("planner", []string{"exec"})
	require.Error(t, err)
	assert.Equal(t, `tool "exec" outside role maximum scope for teammate type "planner"`, err.Error())

	assert.NoError(t, ValidateAllowedToolsForTeammate("unknown-agent", []string{"exec"}))
}
