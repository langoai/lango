package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolBuilder_FullChain(t *testing.T) {
	handler := func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
		return "ok", nil
	}
	params := map[string]interface{}{
		"path": map[string]interface{}{
			"type":        "string",
			"description": "file path",
		},
	}

	tool := NewTool("fs_read", "Read a file from disk").
		Safety(SafetyLevelSafe).
		Params(params).
		Handler(handler).
		Aliases("read", "cat").
		Category("filesystem").
		Hints("file", "open", "load").
		Deferred().
		ReadOnly().
		ConcurrencySafe().
		Activity(ActivityRead).
		Requires("fs_access").
		Build()

	require.NotNil(t, tool)
	assert.Equal(t, "fs_read", tool.Name)
	assert.Equal(t, "Read a file from disk", tool.Description)
	assert.Equal(t, SafetyLevelSafe, tool.SafetyLevel)
	assert.Equal(t, params, tool.Parameters)
	assert.NotNil(t, tool.Handler)

	assert.Equal(t, []string{"read", "cat"}, tool.Capability.Aliases)
	assert.Equal(t, "filesystem", tool.Capability.Category)
	assert.Equal(t, []string{"file", "open", "load"}, tool.Capability.SearchHints)
	assert.Equal(t, ExposureDeferred, tool.Capability.Exposure)
	assert.True(t, tool.Capability.ReadOnly)
	assert.True(t, tool.Capability.ConcurrencySafe)
	assert.Equal(t, ActivityRead, tool.Capability.Activity)
	assert.Equal(t, []string{"fs_access"}, tool.Capability.RequiredCapabilities)
}

func TestToolBuilder_ZeroValue(t *testing.T) {
	tool := NewTool("noop", "Does nothing").Build()

	require.NotNil(t, tool)
	assert.Equal(t, "noop", tool.Name)
	assert.Equal(t, "Does nothing", tool.Description)

	// Zero-value SafetyLevel (0) is treated as dangerous.
	assert.Equal(t, SafetyLevel(0), tool.SafetyLevel)
	assert.True(t, tool.SafetyLevel.IsDangerous())

	assert.Nil(t, tool.Parameters)
	assert.Nil(t, tool.Handler)

	// Capability zero values are backward compatible.
	assert.Nil(t, tool.Capability.Aliases)
	assert.Empty(t, tool.Capability.Category)
	assert.Nil(t, tool.Capability.SearchHints)
	assert.Equal(t, ExposureDefault, tool.Capability.Exposure)
	assert.True(t, tool.Capability.Exposure.IsVisible())
	assert.False(t, tool.Capability.ReadOnly)
	assert.False(t, tool.Capability.ConcurrencySafe)
	assert.Empty(t, string(tool.Capability.Activity))
	assert.Nil(t, tool.Capability.RequiredCapabilities)
}

func TestToolBuilder_IndividualSetters(t *testing.T) {
	tests := []struct {
		give string
		want func(t *testing.T, tool *Tool)
	}{
		{
			give: "Safety",
			want: func(t *testing.T, tool *Tool) {
				got := NewTool("t", "d").Safety(SafetyLevelModerate).Build()
				assert.Equal(t, SafetyLevelModerate, got.SafetyLevel)
			},
		},
		{
			give: "Params",
			want: func(t *testing.T, tool *Tool) {
				p := map[string]interface{}{"key": "value"}
				got := NewTool("t", "d").Params(p).Build()
				assert.Equal(t, p, got.Parameters)
			},
		},
		{
			give: "Handler",
			want: func(t *testing.T, tool *Tool) {
				h := func(_ context.Context, _ map[string]interface{}) (interface{}, error) {
					return 42, nil
				}
				got := NewTool("t", "d").Handler(h).Build()
				assert.NotNil(t, got.Handler)
				result, err := got.Handler(context.Background(), nil)
				require.NoError(t, err)
				assert.Equal(t, 42, result)
			},
		},
		{
			give: "Aliases",
			want: func(t *testing.T, tool *Tool) {
				got := NewTool("t", "d").Aliases("a1", "a2").Build()
				assert.Equal(t, []string{"a1", "a2"}, got.Capability.Aliases)
			},
		},
		{
			give: "Category",
			want: func(t *testing.T, tool *Tool) {
				got := NewTool("t", "d").Category("crypto").Build()
				assert.Equal(t, "crypto", got.Capability.Category)
			},
		},
		{
			give: "Hints",
			want: func(t *testing.T, tool *Tool) {
				got := NewTool("t", "d").Hints("encrypt", "decrypt").Build()
				assert.Equal(t, []string{"encrypt", "decrypt"}, got.Capability.SearchHints)
			},
		},
		{
			give: "Deferred",
			want: func(t *testing.T, tool *Tool) {
				got := NewTool("t", "d").Deferred().Build()
				assert.Equal(t, ExposureDeferred, got.Capability.Exposure)
				assert.False(t, got.Capability.Exposure.IsVisible())
			},
		},
		{
			give: "Hidden",
			want: func(t *testing.T, tool *Tool) {
				got := NewTool("t", "d").Hidden().Build()
				assert.Equal(t, ExposureHidden, got.Capability.Exposure)
				assert.False(t, got.Capability.Exposure.IsVisible())
			},
		},
		{
			give: "ReadOnly",
			want: func(t *testing.T, tool *Tool) {
				got := NewTool("t", "d").ReadOnly().Build()
				assert.True(t, got.Capability.ReadOnly)
			},
		},
		{
			give: "ConcurrencySafe",
			want: func(t *testing.T, tool *Tool) {
				got := NewTool("t", "d").ConcurrencySafe().Build()
				assert.True(t, got.Capability.ConcurrencySafe)
			},
		},
		{
			give: "Activity",
			want: func(t *testing.T, tool *Tool) {
				got := NewTool("t", "d").Activity(ActivityWrite).Build()
				assert.Equal(t, ActivityWrite, got.Capability.Activity)
			},
		},
		{
			give: "Requires",
			want: func(t *testing.T, tool *Tool) {
				got := NewTool("t", "d").Requires("payment", "encryption").Build()
				assert.Equal(t, []string{"payment", "encryption"}, got.Capability.RequiredCapabilities)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			tt.want(t, nil)
		})
	}
}

func TestToolBuilder_BuildReturnsCopy(t *testing.T) {
	builder := NewTool("orig", "original description").
		Safety(SafetyLevelSafe).
		Category("test")

	tool1 := builder.Build()
	tool2 := builder.Build()

	// Mutating tool1 must not affect tool2.
	tool1.Name = "mutated"
	tool1.Capability.Category = "mutated"

	assert.Equal(t, "orig", tool2.Name)
	assert.Equal(t, "test", tool2.Capability.Category)
}

func TestToolBuilder_OverwritePrevious(t *testing.T) {
	tool := NewTool("t", "d").
		Safety(SafetyLevelSafe).
		Safety(SafetyLevelDangerous).
		Category("first").
		Category("second").
		Deferred().
		Hidden().
		Build()

	assert.Equal(t, SafetyLevelDangerous, tool.SafetyLevel)
	assert.Equal(t, "second", tool.Capability.Category)
	assert.Equal(t, ExposureHidden, tool.Capability.Exposure)
}
