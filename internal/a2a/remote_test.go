package a2a

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/remoteagent"

	"github.com/langoai/lango/internal/config"
)

func TestLoadRemoteAgents_MissingAgentCardURLReturnsError(t *testing.T) {
	agents, err := LoadRemoteAgents([]config.RemoteAgentConfig{
		{Name: "weather-agent"},
	}, zap.NewNop().Sugar())

	require.Error(t, err)
	assert.Empty(t, agents)
	assert.Contains(t, err.Error(), "weather-agent")
	assert.Contains(t, err.Error(), "agentCardUrl")
}

func TestLoadRemoteAgents_MixedValidAndInvalidReturnsPartialAgentsAndError(t *testing.T) {
	agents, err := LoadRemoteAgents([]config.RemoteAgentConfig{
		{Name: "weather-agent", AgentCardURL: "https://weather.example/.well-known/agent.json"},
		{Name: "review-agent"},
	}, zap.NewNop().Sugar())

	require.Error(t, err)
	require.Len(t, agents, 1)
	assert.Equal(t, "weather-agent", agents[0].Name())
	assert.Contains(t, err.Error(), "review-agent")
	assert.Contains(t, err.Error(), "agentCardUrl")
}

func TestLoadRemoteAgents_RemoteConstructorErrorReturnsError(t *testing.T) {
	orig := newRemoteA2AFn
	newRemoteA2AFn = func(remoteagent.A2AConfig) (adkagent.Agent, error) {
		return nil, errors.New("constructor failed")
	}
	t.Cleanup(func() {
		newRemoteA2AFn = orig
	})

	agents, err := LoadRemoteAgents([]config.RemoteAgentConfig{
		{Name: "weather-agent", AgentCardURL: "https://weather.example/.well-known/agent.json"},
	}, zap.NewNop().Sugar())

	require.Error(t, err)
	assert.Empty(t, agents)
	assert.Contains(t, err.Error(), "weather-agent")
	assert.Contains(t, err.Error(), "constructor failed")
}
