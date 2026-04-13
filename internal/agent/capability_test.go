package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExposurePolicy_String(t *testing.T) {
	tests := []struct {
		give ExposurePolicy
		want string
	}{
		{give: ExposureDefault, want: "default"},
		{give: ExposureAlwaysVisible, want: "always_visible"},
		{give: ExposureDeferred, want: "deferred"},
		{give: ExposureHidden, want: "hidden"},
		{give: ExposurePolicy(99), want: "default"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.give.String())
		})
	}
}

func TestExposurePolicy_IsVisible(t *testing.T) {
	tests := []struct {
		give ExposurePolicy
		want bool
	}{
		{give: ExposureDefault, want: true},
		{give: ExposureAlwaysVisible, want: true},
		{give: ExposureDeferred, want: false},
		{give: ExposureHidden, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.give.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.give.IsVisible())
		})
	}
}

func TestToolCapability_ZeroValue(t *testing.T) {
	var cap ToolCapability

	assert.Nil(t, cap.Aliases)
	assert.Empty(t, cap.Category)
	assert.Nil(t, cap.SearchHints)
	assert.Equal(t, ExposureDefault, cap.Exposure)
	assert.False(t, cap.ReadOnly)
	assert.False(t, cap.ConcurrencySafe)
	assert.Empty(t, string(cap.Activity))
	assert.Nil(t, cap.RequiredCapabilities)

	// Zero-value Exposure is visible (backward compatible).
	assert.True(t, cap.Exposure.IsVisible())
}

func TestTool_CapabilityField(t *testing.T) {
	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		SafetyLevel: SafetyLevelSafe,
		Capability: ToolCapability{
			Aliases:         []string{"tt", "test"},
			Category:        "testing",
			SearchHints:     []string{"unit", "test"},
			Exposure:        ExposureDeferred,
			ReadOnly:        true,
			ConcurrencySafe: true,
			Activity:        ActivityQuery,
		},
	}

	assert.Equal(t, "test_tool", tool.Name)
	assert.Equal(t, []string{"tt", "test"}, tool.Capability.Aliases)
	assert.Equal(t, "testing", tool.Capability.Category)
	assert.True(t, tool.Capability.ReadOnly)
	assert.True(t, tool.Capability.ConcurrencySafe)
	assert.Equal(t, ActivityQuery, tool.Capability.Activity)
	assert.False(t, tool.Capability.Exposure.IsVisible())
}

func TestTool_ZeroCapabilityBackwardCompat(t *testing.T) {
	// Existing code creates Tool without Capability — must still work.
	tool := Tool{
		Name:        "legacy_tool",
		Description: "A legacy tool",
		SafetyLevel: SafetyLevelSafe,
	}

	assert.True(t, tool.Capability.Exposure.IsVisible())
	assert.False(t, tool.Capability.ReadOnly)
	assert.False(t, tool.Capability.ConcurrencySafe)
}
