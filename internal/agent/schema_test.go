package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchema_Build_SimpleString(t *testing.T) {
	t.Parallel()

	got := Schema().
		Str("command", "The shell command to execute").
		Required("command").
		Build()

	assert.Equal(t, "object", got["type"])

	props, ok := got["properties"].(map[string]interface{})
	assert.True(t, ok)

	cmd, ok := props["command"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "string", cmd["type"])
	assert.Equal(t, "The shell command to execute", cmd["description"])

	required, ok := got["required"].([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"command"}, required)
}

func TestSchema_Build_MultipleTypes(t *testing.T) {
	t.Parallel()

	got := Schema().
		Str("path", "File path").
		Int("offset", "Start line").
		Bool("fullPage", "Capture full page").
		Required("path").
		Build()

	props := got["properties"].(map[string]interface{})

	path := props["path"].(map[string]interface{})
	assert.Equal(t, "string", path["type"])

	offset := props["offset"].(map[string]interface{})
	assert.Equal(t, "integer", offset["type"])

	fp := props["fullPage"].(map[string]interface{})
	assert.Equal(t, "boolean", fp["type"])

	assert.Equal(t, []string{"path"}, got["required"])
}

func TestSchema_Build_Enum(t *testing.T) {
	t.Parallel()

	got := Schema().
		Enum("action", "The action", "click", "type", "eval").
		Required("action").
		Build()

	props := got["properties"].(map[string]interface{})
	action := props["action"].(map[string]interface{})

	assert.Equal(t, "string", action["type"])
	assert.Equal(t, []string{"click", "type", "eval"}, action["enum"])
}

func TestSchema_Build_NoRequired(t *testing.T) {
	t.Parallel()

	got := Schema().
		Bool("verbose", "Verbose output").
		Build()

	_, hasRequired := got["required"]
	assert.False(t, hasRequired, "should not have required key when none specified")
}

func TestSchema_Build_EmptySchema(t *testing.T) {
	t.Parallel()

	got := Schema().Build()

	assert.Equal(t, "object", got["type"])
	props := got["properties"].(map[string]interface{})
	assert.Empty(t, props)
}

func TestSchema_Build_MultipleRequired(t *testing.T) {
	t.Parallel()

	got := Schema().
		Str("path", "File path").
		Str("content", "Content").
		Int("startLine", "Start").
		Required("path", "content", "startLine").
		Build()

	required := got["required"].([]string)
	assert.Equal(t, []string{"path", "content", "startLine"}, required)
}
