package a2a

import (
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/remoteagent"

	"github.com/langoai/lango/internal/config"
)

// LoadRemoteAgents creates ADK agents from the configured remote A2A agent list.
// Each remote agent can be used as a sub-agent in the orchestrator.
func LoadRemoteAgents(remotes []config.RemoteAgentConfig, logger *zap.SugaredLogger) ([]agent.Agent, error) {
	if len(remotes) == 0 {
		return nil, nil
	}

	agents := make([]agent.Agent, 0, len(remotes))

	for _, rc := range remotes {
		if rc.AgentCardURL == "" {
			logger.Warnw("remote agent missing card URL, skipping", "name", rc.Name)
			continue
		}

		a2aCfg := remoteagent.A2AConfig{
			Name:            rc.Name,
			Description:     fmt.Sprintf("Remote A2A agent: %s", rc.Name),
			AgentCardSource: rc.AgentCardURL,
		}

		remoteAgent, err := remoteagent.NewA2A(a2aCfg)
		if err != nil {
			logger.Warnw("load remote agent",
				"name", rc.Name,
				"url", rc.AgentCardURL,
				"error", err,
			)
			continue
		}

		agents = append(agents, remoteAgent)
		logger.Infow("remote A2A agent loaded",
			"name", rc.Name,
			"url", rc.AgentCardURL,
		)
	}

	return agents, nil
}
