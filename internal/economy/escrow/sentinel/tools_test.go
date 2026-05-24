package sentinel_test

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/economy/escrow/sentinel"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTools(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	engine := sentinel.New(bus, sentinel.DefaultSentinelConfig())
	tools := sentinel.BuildTools(engine)

	assert.Len(t, tools, 4)

	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = tool.Name
	}

	wantNames := []string{
		"sentinel_status",
		"sentinel_alerts",
		"sentinel_config",
		"sentinel_acknowledge",
	}
	for _, name := range wantNames {
		assert.Contains(t, names, name)
	}
}

func TestBuildTools_SafetyLevels(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	engine := sentinel.New(bus, sentinel.DefaultSentinelConfig())
	tools := sentinel.BuildTools(engine)

	toolMap := make(map[string]*agent.Tool, len(tools))
	for _, tool := range tools {
		toolMap[tool.Name] = tool
	}

	tests := []struct {
		give     string
		wantSafe bool
	}{
		{give: "sentinel_status", wantSafe: true},
		{give: "sentinel_alerts", wantSafe: true},
		{give: "sentinel_config", wantSafe: true},
		{give: "sentinel_acknowledge", wantSafe: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			tool, ok := toolMap[tt.give]
			require.True(t, ok, "tool %q not found", tt.give)
			isSafe := tool.SafetyLevel == agent.SafetyLevelSafe
			assert.Equal(t, tt.wantSafe, isSafe)
		})
	}
}

func TestStatusTool_Handler(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	engine := sentinel.New(bus, sentinel.DefaultSentinelConfig())
	tools := sentinel.BuildTools(engine)

	var statusTool *agent.Tool
	for _, tool := range tools {
		if tool.Name == "sentinel_status" {
			statusTool = tool
			break
		}
	}
	require.NotNil(t, statusTool)

	result, err := statusTool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)

	m, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, false, m["running"])
	assert.Equal(t, 0, m["totalAlerts"])
	assert.Equal(t, 0, m["activeAlerts"])
}

func TestConfigTool_Handler(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	engine := sentinel.New(bus, sentinel.DefaultSentinelConfig())
	tools := sentinel.BuildTools(engine)

	var cfgTool *agent.Tool
	for _, tool := range tools {
		if tool.Name == "sentinel_config" {
			cfgTool = tool
			break
		}
	}
	require.NotNil(t, cfgTool)

	result, err := cfgTool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)

	m, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, m["rapidCreationWindow"])
	assert.NotEmpty(t, m["disputeWindow"])
	assert.NotEmpty(t, m["washTradeWindow"])
}

func TestAlertsTool_EmptyList(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	engine := sentinel.New(bus, sentinel.DefaultSentinelConfig())
	tools := sentinel.BuildTools(engine)

	var alertsTool *agent.Tool
	for _, tool := range tools {
		if tool.Name == "sentinel_alerts" {
			alertsTool = tool
			break
		}
	}
	require.NotNil(t, alertsTool)

	result, err := alertsTool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)

	m, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, m["count"])
}

func TestAcknowledgeTool_RequiresAlertID(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	engine := sentinel.New(bus, sentinel.DefaultSentinelConfig())
	tools := sentinel.BuildTools(engine)

	var ackTool *agent.Tool
	for _, tool := range tools {
		if tool.Name == "sentinel_acknowledge" {
			ackTool = tool
			break
		}
	}
	require.NotNil(t, ackTool)

	result, err := ackTool.Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "missing alertId parameter")
}

func TestAcknowledgeTool_CapabilityMetadataMatchesMutation(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	engine := sentinel.New(bus, sentinel.DefaultSentinelConfig())
	tools := sentinel.BuildTools(engine)

	var ackTool *agent.Tool
	for _, tool := range tools {
		if tool.Name == "sentinel_acknowledge" {
			ackTool = tool
			break
		}
	}
	require.NotNil(t, ackTool)
	assert.Equal(t, agent.ActivityManage, ackTool.Capability.Activity)
	assert.False(t, ackTool.Capability.ReadOnly)
}
