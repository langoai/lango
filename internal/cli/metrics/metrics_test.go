package metrics

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func executeMetricsCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func configForMetricsServer(t *testing.T, rawURL string) *config.Config {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)
	host, portText, err := net.SplitHostPort(parsed.Host)
	require.NoError(t, err)
	port, err := strconv.Atoi(portText)
	require.NoError(t, err)

	cfg := config.DefaultConfig()
	cfg.Server.Host = host
	cfg.Server.Port = port
	return cfg
}

func TestMetricsSummary_UsesConfiguredGatewayWhenAddrOmitted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/metrics", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"uptime":         "configured",
			"toolExecutions": 9,
		}))
	}))
	defer srv.Close()

	cmd := NewMetricsCmdWithConfig(func() (*config.Config, error) {
		return configForMetricsServer(t, srv.URL), nil
	})
	out, err := executeMetricsCmd(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, out, "configured")
	assert.Contains(t, out, "Tool Executions:  9")
}

func TestMetricsSummary_ExplicitAddrSkipsConfigLoader(t *testing.T) {
	var configLoads atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/metrics", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"uptime":         "explicit",
			"toolExecutions": 4,
		}))
	}))
	defer srv.Close()

	cmd := NewMetricsCmdWithConfig(func() (*config.Config, error) {
		configLoads.Add(1)
		return config.DefaultConfig(), nil
	})
	out, err := executeMetricsCmd(t, cmd, "--addr", srv.URL)

	require.NoError(t, err)
	assert.Contains(t, out, "explicit")
	assert.Equal(t, int32(0), configLoads.Load())
}

func TestMetricsSubcommands_UseConfiguredGatewayWhenAddrOmitted(t *testing.T) {
	tests := []struct {
		name string
		args []string
		path string
		body map[string]any
		want string
	}{
		{
			name: "sessions",
			args: []string{"sessions"},
			path: "/metrics/sessions",
			body: map[string]any{"sessions": []map[string]any{{
				"sessionKey":   "configured-session",
				"inputTokens":  7,
				"outputTokens": 5,
				"totalTokens":  12,
				"requestCount": 2,
			}}},
			want: "configured-session",
		},
		{
			name: "tools",
			args: []string{"tools"},
			path: "/metrics/tools",
			body: map[string]any{"tools": []map[string]any{{
				"name":        "configured_tool",
				"count":       3,
				"errors":      0,
				"avgDuration": "1ms",
				"errorRate":   0,
			}}},
			want: "configured_tool",
		},
		{
			name: "agents",
			args: []string{"agents"},
			path: "/metrics/agents",
			body: map[string]any{"agents": []map[string]any{{
				"name":         "configured-agent",
				"inputTokens":  11,
				"outputTokens": 13,
				"toolCalls":    2,
			}}},
			want: "configured-agent",
		},
		{
			name: "history",
			args: []string{"history", "--days", "2"},
			path: "/metrics/history",
			body: map[string]any{
				"records": []map[string]any{{
					"provider":    "configured-provider",
					"model":       "configured-model",
					"sessionKey":  "configured-history",
					"agentName":   "configured-agent",
					"inputTokens": 1,
					"timestamp":   "2026-05-14T01:02:03Z",
				}},
				"total": map[string]any{"inputTokens": 1, "outputTokens": 0, "recordCount": 1},
			},
			want: "configured-provider",
		},
		{
			name: "policy",
			args: []string{"policy"},
			path: "/metrics/policy",
			body: map[string]any{"blocks": 1, "observes": 2, "byReason": map[string]int64{
				"configured_reason": 1,
			}},
			want: "configured_reason",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.path, r.URL.Path)
				require.NoError(t, json.NewEncoder(w).Encode(tt.body))
			}))
			defer srv.Close()

			cmd := NewMetricsCmdWithConfig(func() (*config.Config, error) {
				return configForMetricsServer(t, srv.URL), nil
			})
			out, err := executeMetricsCmd(t, cmd, tt.args...)

			require.NoError(t, err)
			assert.Contains(t, out, tt.want)
		})
	}
}

func TestMetricsSubcommands_ExplicitAddrSkipsConfigLoader(t *testing.T) {
	tests := []struct {
		name string
		args []string
		path string
		body map[string]any
	}{
		{name: "sessions", args: []string{"sessions"}, path: "/metrics/sessions", body: map[string]any{"sessions": []any{}}},
		{name: "tools", args: []string{"tools"}, path: "/metrics/tools", body: map[string]any{"tools": []any{}}},
		{name: "agents", args: []string{"agents"}, path: "/metrics/agents", body: map[string]any{"agents": []any{}}},
		{name: "history", args: []string{"history"}, path: "/metrics/history", body: map[string]any{"records": []any{}, "total": map[string]any{}}},
		{name: "policy", args: []string{"policy"}, path: "/metrics/policy", body: map[string]any{"blocks": 0, "observes": 0, "byReason": map[string]int64{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configLoads atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.path, r.URL.Path)
				require.NoError(t, json.NewEncoder(w).Encode(tt.body))
			}))
			defer srv.Close()

			cmd := NewMetricsCmdWithConfig(func() (*config.Config, error) {
				configLoads.Add(1)
				return config.DefaultConfig(), nil
			})
			args := append(append([]string{}, tt.args...), "--addr", srv.URL)
			_, err := executeMetricsCmd(t, cmd, args...)

			require.NoError(t, err)
			assert.Equal(t, int32(0), configLoads.Load())
		})
	}
}

func TestMetricsPolicy_WritesTableToCommandWriter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/metrics/policy", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"blocks":   3,
			"observes": 12,
			"byReason": map[string]int64{
				"destructive_command":  1,
				"network_exfiltration": 5,
			},
		}))
	}))
	defer srv.Close()

	cmd := newPolicyCmd()
	cmd.Flags().String("output", "table", "")
	cmd.Flags().String("addr", defaultGatewayAddr, "")

	out, err := executeMetricsCmd(t, cmd, "--addr", srv.URL)
	require.NoError(t, err)
	assert.Contains(t, out, "=== Policy Decisions ===")
	assert.Contains(t, out, "Blocks:    3")
	assert.Contains(t, out, "REASON")
	assert.Contains(t, out, "network_exfiltration")
}

func TestMetricsPolicy_WritesJSONToCommandWriter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/metrics/policy", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"blocks":   1,
			"observes": 2,
			"byReason": map[string]int64{
				"catastrophic_pattern": 1,
			},
		}))
	}))
	defer srv.Close()

	cmd := newPolicyCmd()
	cmd.Flags().String("output", "table", "")
	cmd.Flags().String("addr", defaultGatewayAddr, "")

	out, err := executeMetricsCmd(t, cmd, "--addr", srv.URL, "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(1), decoded["blocks"])
	assert.Equal(t, float64(2), decoded["observes"])
}

func TestMetricsSummary_WritesTableToCommandWriter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/metrics", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"uptime": "2h15m30s",
			"tokenUsage": map[string]any{
				"inputTokens":  145200,
				"outputTokens": 52800,
			},
			"toolExecutions": 342,
		}))
	}))
	defer srv.Close()

	cmd := NewMetricsCmd()
	out, err := executeMetricsCmd(t, cmd, "--addr", srv.URL)
	require.NoError(t, err)
	assert.Contains(t, out, "=== System Metrics ===")
	assert.Contains(t, out, "Uptime:           2h15m30s")
	assert.Contains(t, out, "Total Input:      145200 tokens")
	assert.Contains(t, out, "Tool Executions:  342")
}

func TestMetricsSummary_WritesJSONToCommandWriter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/metrics", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"uptime": "2h15m30s",
			"tokenUsage": map[string]any{
				"inputTokens":  145200,
				"outputTokens": 52800,
			},
			"toolExecutions": 342,
		}))
	}))
	defer srv.Close()

	cmd := NewMetricsCmd()
	out, err := executeMetricsCmd(t, cmd, "--addr", srv.URL, "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "2h15m30s", decoded["uptime"])
	assert.Equal(t, float64(342), decoded["toolExecutions"])
}

func TestMetricsSummary_InvalidOutputRejectsBeforeGatewayCall(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cmd := NewMetricsCmd()
	out, err := executeMetricsCmd(t, cmd, "--addr", srv.URL, "--output", "yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
	assert.Empty(t, out)
	assert.Equal(t, int32(0), hits.Load())
}

func TestMetricsSessions_WritesTableToCommandWriter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/metrics/sessions", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"sessions": []map[string]any{
				{
					"sessionKey":   "abc123def456ghij78901xyz",
					"inputTokens":  45200,
					"outputTokens": 12800,
					"totalTokens":  58000,
					"requestCount": 24,
				},
			},
		}))
	}))
	defer srv.Close()

	cmd := NewMetricsCmd()
	out, err := executeMetricsCmd(t, cmd, "sessions", "--addr", srv.URL)
	require.NoError(t, err)
	assert.Contains(t, out, "SESSION")
	assert.Contains(t, out, "45200")
	assert.Contains(t, out, "58000")
}

func TestMetricsSessions_WritesJSONToCommandWriter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/metrics/sessions", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"sessions": []map[string]any{
				{"sessionKey": "abc", "inputTokens": 1, "outputTokens": 2, "totalTokens": 3, "requestCount": 4},
			},
		}))
	}))
	defer srv.Close()

	cmd := NewMetricsCmd()
	out, err := executeMetricsCmd(t, cmd, "sessions", "--addr", srv.URL, "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	sessions, ok := decoded["sessions"].([]any)
	require.True(t, ok)
	require.Len(t, sessions, 1)
}

func TestMetricsTools_WritesEmptyStateToCommandWriter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/metrics/tools", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"tools": []any{}}))
	}))
	defer srv.Close()

	cmd := NewMetricsCmd()
	out, err := executeMetricsCmd(t, cmd, "tools", "--addr", srv.URL)
	require.NoError(t, err)
	assert.Contains(t, out, "No tool execution data available.")
}

func TestMetricsAgents_WritesEmptyStateToCommandWriter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/metrics/agents", r.URL.Path)
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"agents": []any{}}))
	}))
	defer srv.Close()

	cmd := NewMetricsCmd()
	out, err := executeMetricsCmd(t, cmd, "agents", "--addr", srv.URL)
	require.NoError(t, err)
	assert.Contains(t, out, "No agent data available.")
}

func TestMetricsHistory_WritesTableToCommandWriter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/metrics/history", r.URL.Path)
		require.Equal(t, "3", r.URL.Query().Get("days"))
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"records": []map[string]any{
				{
					"provider":     "openai",
					"model":        "gpt-4.1",
					"sessionKey":   "abc",
					"agentName":    "operator",
					"inputTokens":  12,
					"outputTokens": 8,
					"timestamp":    "2026-05-14T01:02:03Z",
				},
			},
			"total": map[string]any{
				"inputTokens":  12,
				"outputTokens": 8,
				"recordCount":  1,
			},
		}))
	}))
	defer srv.Close()

	cmd := NewMetricsCmd()
	out, err := executeMetricsCmd(t, cmd, "history", "--addr", srv.URL, "--days", "3")
	require.NoError(t, err)
	assert.Contains(t, out, "Token usage history (last 3 days)")
	assert.Contains(t, out, "openai")
	assert.Contains(t, out, "gpt-4.1")
}
