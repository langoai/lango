package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/storage"
	"github.com/langoai/lango/internal/testutil"
	"github.com/langoai/lango/internal/turntrace"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeAgentCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func traceBootLoader(t *testing.T, seed func(store turntrace.Store)) func() (*bootstrap.Result, error) {
	t.Helper()
	return func() (*bootstrap.Result, error) {
		client := testutil.TestEntClient(t)
		facade := storage.NewFacade(nil, nil, storage.WithEntClient(client))
		store := turntrace.NewEntStore(client)
		if seed != nil {
			seed(store)
		}
		return &bootstrap.Result{Storage: facade}, nil
	}
}

func TestAgentStatus_DynamicRuntimeDoesNotRequireBuiltinSubAgents(t *testing.T) {
	originalLoader := loadAgentRegistryCounts
	loadAgentRegistryCounts = func(_ *config.Config) (int, int, int, error) {
		return 0, 0, 0, nil
	}
	t.Cleanup(func() {
		loadAgentRegistryCounts = originalLoader
	})

	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.MultiAgent = true
		cfg.Background.Enabled = true
		cfg.Agent.Provider = "anthropic"
		cfg.Agent.Model = "claude-4"
		return cfg, nil
	})

	output, err := executeAgentCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	assert.Contains(t, output, `"teammate_runtime": "dynamic-v1"`)
	assert.Contains(t, output, `"builtin": 0`)
	assert.Contains(t, output, `"active": 0`)
}

func TestStatusCmd_TableHintsBackgroundRequirementForDynamicRuntime(t *testing.T) {
	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.MultiAgent = true
		cfg.Background.Enabled = false
		cfg.Agent.Provider = "anthropic"
		cfg.Agent.Model = "claude-4"
		return cfg, nil
	})

	output, err := executeAgentCmd(t, cmd)
	require.NoError(t, err)

	assert.Contains(t, output, "Enable background.enabled to report dynamic-v1 teammate runtime.")
}

func TestStatusCmd_JSONOmitsTeammateRuntimeWithoutAutomation(t *testing.T) {
	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.MultiAgent = true
		cfg.Cron.Enabled = false
		cfg.Background.Enabled = false
		cfg.Workflow.Enabled = false
		cfg.Agent.Provider = "anthropic"
		cfg.Agent.Model = "claude-4"
		return cfg, nil
	})

	output, err := executeAgentCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	assert.NotContains(t, output, `"teammate_runtime"`)
}

func TestStatusCmd_JSONOmitsTeammateRuntimeInSingleAgentMode(t *testing.T) {
	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.MultiAgent = false
		cfg.Background.Enabled = true
		cfg.Agent.Provider = "anthropic"
		cfg.Agent.Model = "claude-4"
		return cfg, nil
	})

	output, err := executeAgentCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	assert.NotContains(t, output, `"teammate_runtime"`)
}

func TestStatusCmd_TableIncludesTeammateRuntimeWhenBackgroundEnabled(t *testing.T) {
	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.MultiAgent = true
		cfg.Background.Enabled = true
		cfg.Agent.Provider = "anthropic"
		cfg.Agent.Model = "claude-4"
		return cfg, nil
	})

	output, err := executeAgentCmd(t, cmd)
	require.NoError(t, err)

	assert.Contains(t, output, "Teammate Runtime:  dynamic-v1")
}

func TestStatusCmd_JSONOmitsTeammateRuntimeWithoutBackgroundSubmitter(t *testing.T) {
	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.MultiAgent = true
		cfg.Cron.Enabled = true
		cfg.Background.Enabled = false
		cfg.Workflow.Enabled = true
		cfg.Agent.Provider = "anthropic"
		cfg.Agent.Model = "claude-4"
		return cfg, nil
	})

	output, err := executeAgentCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	assert.NotContains(t, output, `"teammate_runtime"`)
}

func TestToolsCmd_WritesTableOutputToCommandWriter(t *testing.T) {
	cmd := newToolsCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Tools.Browser.Enabled = false
		return cfg, nil
	})

	out, err := executeAgentCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "CATEGORY")
	assert.Contains(t, out, "browser")
}

func TestToolsCmd_WritesJSONOutputToCommandWriter(t *testing.T) {
	cmd := newToolsCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Tools.Browser.Enabled = true
		return cfg, nil
	})

	out, err := executeAgentCmd(t, cmd, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"name": "browser"`)
	assert.Contains(t, out, `"enabled": true`)
}

func TestToolsCmd_WritesEmptyFilterStateToCommandWriter(t *testing.T) {
	cmd := newToolsCmd(func() (*config.Config, error) {
		return config.DefaultConfig(), nil
	})

	out, err := executeAgentCmd(t, cmd, "--category", "missing-category")
	require.NoError(t, err)
	assert.Contains(t, out, `No tool category "missing-category" found.`)
}

func TestListCmd_WritesTableOutputToCommandWriter(t *testing.T) {
	cmd := newListCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.AgentsDir = ""
		return cfg, nil
	})

	out, err := executeAgentCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "planner")
}

func TestListCmd_WritesJSONOutputToCommandWriter(t *testing.T) {
	cmd := newListCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		return cfg, nil
	})

	out, err := executeAgentCmd(t, cmd, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"type": "local"`)
	assert.Contains(t, out, `"source": "embedded"`)
}

func TestListCmd_SurfaceUserAgentRegistryLoadError(t *testing.T) {
	agentsDir := t.TempDir()
	badAgentDir := filepath.Join(agentsDir, "broken")
	badAgentPath := filepath.Join(badAgentDir, "AGENT.md")
	require.NoError(t, os.MkdirAll(badAgentDir, 0o755))
	require.NoError(t, os.WriteFile(badAgentPath, []byte("missing frontmatter"), 0o644))

	cmd := newListCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.AgentsDir = agentsDir
		return cfg, nil
	})

	_, err := executeAgentCmd(t, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load user agents")
	assert.Contains(t, err.Error(), badAgentPath)
}

func TestListCmd_MissingUserAgentsDirIsOptional(t *testing.T) {
	missingAgentsDir := filepath.Join(t.TempDir(), "missing")

	cmd := newListCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.AgentsDir = missingAgentsDir
		return cfg, nil
	})

	out, err := executeAgentCmd(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, out, "planner")
}

func TestStatusCmd_SurfaceUserAgentRegistryLoadError(t *testing.T) {
	agentsDir := t.TempDir()
	badAgentDir := filepath.Join(agentsDir, "broken")
	badAgentPath := filepath.Join(badAgentDir, "AGENT.md")
	require.NoError(t, os.MkdirAll(badAgentDir, 0o755))
	require.NoError(t, os.WriteFile(badAgentPath, []byte("missing frontmatter"), 0o644))

	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.AgentsDir = agentsDir
		return cfg, nil
	})

	_, err := executeAgentCmd(t, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load user agents")
	assert.Contains(t, err.Error(), badAgentPath)
}

func TestStatusCmd_SurfaceEmbeddedRegistryLoadError(t *testing.T) {
	originalLoader := loadAgentRegistryCounts
	loadAgentRegistryCounts = func(_ *config.Config) (int, int, int, error) {
		return 0, 0, 0, errors.New("load embedded agents: embedded registry unavailable")
	}
	t.Cleanup(func() {
		loadAgentRegistryCounts = originalLoader
	})

	cmd := newStatusCmd(func() (*config.Config, error) {
		return config.DefaultConfig(), nil
	})

	_, err := executeAgentCmd(t, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load embedded agents")
	assert.Contains(t, err.Error(), "embedded registry unavailable")
}

func TestStatusCmd_MissingUserAgentsDirIsOptional(t *testing.T) {
	missingAgentsDir := filepath.Join(t.TempDir(), "missing")

	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.AgentsDir = missingAgentsDir
		return cfg, nil
	})

	out, err := executeAgentCmd(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, out, "Builtin Agents:")
	assert.Contains(t, out, "User Agents:")
}

func TestTraceMetricsCmd_WritesEmptyStateToCommandWriter(t *testing.T) {
	cmd := newTraceMetricsCmd(traceBootLoader(t, nil))

	out, err := executeAgentCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "No agent metrics found.")
}

func TestNewAgentCmd_NestsTraceMetricsUnderTrace(t *testing.T) {
	cmd := NewAgentCmd(
		func() (*config.Config, error) { return config.DefaultConfig(), nil },
		traceBootLoader(t, nil),
	)

	traceCmd, _, err := cmd.Find([]string{"trace"})
	require.NoError(t, err)
	require.NotNil(t, traceCmd)

	metricsCmd, _, err := cmd.Find([]string{"trace", "metrics"})
	require.NoError(t, err)
	require.NotNil(t, metricsCmd)
	assert.Equal(t, "metrics", metricsCmd.Name())

	rootMetricsCmd, _, err := cmd.Find([]string{"metrics"})
	require.Error(t, err)
	require.NotNil(t, rootMetricsCmd)
	assert.Equal(t, "agent", rootMetricsCmd.Name())
	assert.Contains(t, err.Error(), `unknown command "metrics" for "agent"`)
}

func TestTraceMetricsCmd_WritesTableOutputToCommandWriter(t *testing.T) {
	now := time.Now()
	cmd := newTraceMetricsCmd(traceBootLoader(t, func(store turntrace.Store) {
		require.NoError(t, store.CreateTrace(context.Background(), turntrace.Trace{
			TraceID:    "t1",
			SessionKey: "s1",
			Entrypoint: "chat",
			Outcome:    turntrace.OutcomeRunning,
			StartedAt:  now,
		}))
		require.NoError(t, store.AppendEvent(context.Background(), turntrace.Event{
			TraceID:   "t1",
			Seq:       1,
			EventType: turntrace.EventToolCall,
			AgentName: "operator",
			CreatedAt: now.Add(100 * time.Millisecond),
		}))
		require.NoError(t, store.FinishTrace(context.Background(), "t1", turntrace.OutcomeSuccess, "ok", "", "", "", now.Add(time.Second)))
	}))

	out, err := executeAgentCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "AGENT")
	assert.Contains(t, out, "operator")
}

func TestTraceMetricsCmd_WritesJSONOutputToCommandWriter(t *testing.T) {
	now := time.Now()
	cmd := newTraceMetricsCmd(traceBootLoader(t, func(store turntrace.Store) {
		require.NoError(t, store.CreateTrace(context.Background(), turntrace.Trace{
			TraceID:    "t1",
			SessionKey: "s1",
			Entrypoint: "chat",
			Outcome:    turntrace.OutcomeRunning,
			StartedAt:  now,
		}))
		require.NoError(t, store.AppendEvent(context.Background(), turntrace.Event{
			TraceID:   "t1",
			Seq:       1,
			EventType: turntrace.EventToolCall,
			AgentName: "operator",
			CreatedAt: now.Add(100 * time.Millisecond),
		}))
		require.NoError(t, store.FinishTrace(context.Background(), "t1", turntrace.OutcomeSuccess, "ok", "", "", "", now.Add(time.Second)))
	}))

	out, err := executeAgentCmd(t, cmd, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"operator"`)
	assert.Contains(t, out, `"success_count"`)
}

func TestTraceListCmd_WritesEmptyStateToCommandWriter(t *testing.T) {
	cmd := newTraceListCmd(traceBootLoader(t, nil))

	out, err := executeAgentCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "No traces found.")
}

func TestTraceListCmd_WritesTableOutputToCommandWriter(t *testing.T) {
	now := time.Now()
	cmd := newTraceListCmd(traceBootLoader(t, func(store turntrace.Store) {
		require.NoError(t, store.CreateTrace(context.Background(), turntrace.Trace{
			TraceID:    "trace-1",
			SessionKey: "session-1",
			Entrypoint: "chat",
			Outcome:    turntrace.OutcomeSuccess,
			StartedAt:  now,
		}))
		require.NoError(t, store.FinishTrace(context.Background(), "trace-1", turntrace.OutcomeSuccess, "ok", "", "", "", now.Add(time.Second)))
	}))

	out, err := executeAgentCmd(t, cmd, "--session", "session-1")
	require.NoError(t, err)
	assert.Contains(t, out, "TRACE ID")
	assert.Contains(t, out, "trace-1")
	assert.Contains(t, out, "success")
}

func TestTraceListCmd_WritesJSONOutputToCommandWriter(t *testing.T) {
	now := time.Now()
	cmd := newTraceListCmd(traceBootLoader(t, func(store turntrace.Store) {
		require.NoError(t, store.CreateTrace(context.Background(), turntrace.Trace{
			TraceID:    "trace-1",
			SessionKey: "session-1",
			Entrypoint: "chat",
			Outcome:    turntrace.OutcomeSuccess,
			StartedAt:  now,
		}))
		require.NoError(t, store.FinishTrace(context.Background(), "trace-1", turntrace.OutcomeSuccess, "ok", "", "", "", now.Add(time.Second)))
	}))

	out, err := executeAgentCmd(t, cmd, "--session", "session-1", "--output", "json")
	require.NoError(t, err)
	var payload []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	require.Len(t, payload, 1)
	assert.Equal(t, "trace-1", payload[0]["TraceID"])
	assert.Equal(t, "success", payload[0]["Outcome"])
}

func TestTraceShowCmd_WritesEmptyStateToCommandWriter(t *testing.T) {
	cmd := newTraceDetailCmd(traceBootLoader(t, nil))

	out, err := executeAgentCmd(t, cmd, "missing-trace")
	require.NoError(t, err)
	assert.Contains(t, out, "No events found for trace missing-trace")
}

func TestTraceShowCmd_WritesTableOutputToCommandWriter(t *testing.T) {
	now := time.Now()
	cmd := newTraceDetailCmd(traceBootLoader(t, func(store turntrace.Store) {
		require.NoError(t, store.CreateTrace(context.Background(), turntrace.Trace{
			TraceID:    "trace-1",
			SessionKey: "session-1",
			Entrypoint: "chat",
			Outcome:    turntrace.OutcomeRunning,
			StartedAt:  now,
		}))
		require.NoError(t, store.AppendEvent(context.Background(), turntrace.Event{
			TraceID:     "trace-1",
			Seq:         1,
			EventType:   turntrace.EventToolCall,
			AgentName:   "operator",
			ToolName:    "exec",
			PayloadJSON: `{"command":"pwd"}`,
			CreatedAt:   now.Add(50 * time.Millisecond),
		}))
	}))

	out, err := executeAgentCmd(t, cmd, "trace-1")
	require.NoError(t, err)
	assert.Contains(t, out, "Trace: trace-1 (1 events)")
	assert.Contains(t, out, "SEQ")
	assert.Contains(t, out, "operator")
}

func TestTraceShowCmd_WritesJSONOutputToCommandWriter(t *testing.T) {
	now := time.Now()
	cmd := newTraceDetailCmd(traceBootLoader(t, func(store turntrace.Store) {
		require.NoError(t, store.CreateTrace(context.Background(), turntrace.Trace{
			TraceID:    "trace-1",
			SessionKey: "session-1",
			Entrypoint: "chat",
			Outcome:    turntrace.OutcomeRunning,
			StartedAt:  now,
		}))
		require.NoError(t, store.AppendEvent(context.Background(), turntrace.Event{
			TraceID:     "trace-1",
			Seq:         1,
			EventType:   turntrace.EventToolCall,
			AgentName:   "operator",
			ToolName:    "exec",
			PayloadJSON: `{"command":"pwd"}`,
			CreatedAt:   now.Add(50 * time.Millisecond),
		}))
	}))

	out, err := executeAgentCmd(t, cmd, "trace-1", "--output", "json")
	require.NoError(t, err)
	var payload []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	require.Len(t, payload, 1)
	assert.Equal(t, "trace-1", payload[0]["TraceID"])
	assert.Equal(t, "exec", payload[0]["ToolName"])
}

func TestGraphCmd_WritesEmptyStateToCommandWriter(t *testing.T) {
	cmd := newGraphCmd(traceBootLoader(t, nil))

	out, err := executeAgentCmd(t, cmd, "session-1")
	require.NoError(t, err)
	assert.Contains(t, out, "No delegation data found for this session.")
}

func TestGraphCmd_WritesTextOutputToCommandWriter(t *testing.T) {
	now := time.Now()
	cmd := newGraphCmd(traceBootLoader(t, func(store turntrace.Store) {
		require.NoError(t, store.CreateTrace(context.Background(), turntrace.Trace{
			TraceID:    "trace-1",
			SessionKey: "session-1",
			Entrypoint: "chat",
			Outcome:    turntrace.OutcomeRunning,
			StartedAt:  now,
		}))
		require.NoError(t, store.AppendEvent(context.Background(), turntrace.Event{
			TraceID:     "trace-1",
			Seq:         1,
			EventType:   turntrace.EventDelegation,
			AgentName:   "planner",
			PayloadJSON: `{"to":"operator"}`,
			CreatedAt:   now.Add(50 * time.Millisecond),
		}))
		require.NoError(t, store.AppendEvent(context.Background(), turntrace.Event{
			TraceID:   "trace-1",
			Seq:       2,
			EventType: turntrace.EventToolCall,
			AgentName: "operator",
			ToolName:  "exec",
			CreatedAt: now.Add(100 * time.Millisecond),
		}))
	}))

	out, err := executeAgentCmd(t, cmd, "session-1")
	require.NoError(t, err)
	assert.Contains(t, out, "Delegation graph for session: session-1")
	assert.Contains(t, out, "AGENT")
	assert.Contains(t, out, "planner")
	assert.Contains(t, out, "Edges (1):")
}

func TestGraphCmd_WritesJSONOutputToCommandWriter(t *testing.T) {
	now := time.Now()
	cmd := newGraphCmd(traceBootLoader(t, func(store turntrace.Store) {
		require.NoError(t, store.CreateTrace(context.Background(), turntrace.Trace{
			TraceID:    "trace-1",
			SessionKey: "session-1",
			Entrypoint: "chat",
			Outcome:    turntrace.OutcomeRunning,
			StartedAt:  now,
		}))
		require.NoError(t, store.AppendEvent(context.Background(), turntrace.Event{
			TraceID:     "trace-1",
			Seq:         1,
			EventType:   turntrace.EventDelegation,
			AgentName:   "planner",
			PayloadJSON: `{"to":"operator"}`,
			CreatedAt:   now.Add(50 * time.Millisecond),
		}))
	}))

	out, err := executeAgentCmd(t, cmd, "session-1", "--output", "json")
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Equal(t, "session-1", payload["session_key"])
	edges, ok := payload["edges"].([]any)
	require.True(t, ok)
	require.Len(t, edges, 1)
	edge, ok := edges[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "planner", edge["from"])
	assert.Equal(t, "operator", edge["to"])
}

func TestAgentDiagnosticsCommands_InvalidOutputFailFast(t *testing.T) {
	traceList := newTraceListCmd(traceBootLoader(t, nil))
	_, err := executeAgentCmd(t, traceList, "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)

	traceShow := newTraceDetailCmd(traceBootLoader(t, nil))
	_, err = executeAgentCmd(t, traceShow, "trace-1", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)

	graphCmd := newGraphCmd(traceBootLoader(t, nil))
	_, err = executeAgentCmd(t, graphCmd, "session-1", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)

	metricsCmd := newTraceMetricsCmd(traceBootLoader(t, nil))
	_, err = executeAgentCmd(t, metricsCmd, "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}
