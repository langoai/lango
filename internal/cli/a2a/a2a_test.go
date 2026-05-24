package a2a

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/testutil"
)

func executeA2ACmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestA2ACard_TableWritesToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.A2A.Enabled = true
	cfg.A2A.BaseURL = "http://localhost:18789"
	cfg.A2A.AgentName = "lango"
	cfg.A2A.AgentDescription = "AI assistant with tools"
	cfg.A2A.RemoteAgents = []config.RemoteAgentConfig{{
		Name:         "weather-agent",
		AgentCardURL: "http://weather/.well-known/agent.json",
	}}
	cmd := NewA2ACmd(testutil.FakeCfgLoader(cfg))

	out, err := executeA2ACmd(t, cmd, "card")

	require.NoError(t, err)
	assert.Contains(t, out, "A2A Agent Card")
	assert.Contains(t, out, "weather-agent")
}

func TestA2ACard_JSONWritesToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.A2A.Enabled = true
	cfg.A2A.AgentName = "lango"
	cmd := NewA2ACmd(testutil.FakeCfgLoader(cfg))

	out, err := executeA2ACmd(t, cmd, "card", "--output", "json")

	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Equal(t, true, payload["enabled"])
	assert.Equal(t, "lango", payload["agent_name"])
}

func TestA2ACheck_TableAndJSONWriteToCommandOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"name":"weather-agent",
			"description":"Provides weather data",
			"url":"http://weather-svc:8080",
			"did":"did:lango:abc",
			"capabilities":["weather"],
			"skills":[{"id":"get-weather","name":"Get Weather","tags":["weather"]}]
		}`))
	}))
	defer server.Close()

	cmd := NewA2ACmd(testutil.FakeCfgLoader(config.DefaultConfig()))
	tableOut, err := executeA2ACmd(t, cmd, "check", server.URL)
	require.NoError(t, err)
	assert.Contains(t, tableOut, "Remote Agent Card")
	assert.Contains(t, tableOut, "weather-agent")

	jsonOut, err := executeA2ACmd(t, cmd, "check", server.URL, "--output", "json")
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(jsonOut), &payload))
	assert.Equal(t, "weather-agent", payload["name"])
}

func TestA2ACard_InvalidOutputFailsBeforeConfigLoad(t *testing.T) {
	cmd := NewA2ACmd(func() (*config.Config, error) {
		t.Fatal("config loader should not be called for invalid output")
		return nil, nil
	})

	_, err := executeA2ACmd(t, cmd, "card", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestA2ACheck_InvalidOutputFailsBeforeFetch(t *testing.T) {
	cmd := NewA2ACmd(testutil.FakeCfgLoader(config.DefaultConfig()))

	_, err := executeA2ACmd(t, cmd, "check", "https://agent.example.com", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}
