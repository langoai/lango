package a2a

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/adk/agent"
	remoteagent "google.golang.org/adk/agent/remoteagent/v2"

	"github.com/langoai/lango/internal/config"
)

var newRemoteA2AFn = remoteagent.NewA2A

// LoadRemoteAgents creates ADK agents from the configured remote A2A agent list.
// Each remote agent can be used as a sub-agent in the orchestrator.
func LoadRemoteAgents(remotes []config.RemoteAgentConfig, logger *zap.SugaredLogger) ([]agent.Agent, error) {
	if len(remotes) == 0 {
		return nil, nil
	}

	agents := make([]agent.Agent, 0, len(remotes))
	var loadErrs []error

	for _, rc := range remotes {
		if rc.AgentCardURL == "" {
			err := fmt.Errorf("remote agent %q missing agentCardUrl", rc.Name)
			loadErrs = append(loadErrs, err)
			logger.Warnw("remote agent missing card URL, skipping", "name", rc.Name)
			continue
		}

		a2aCfg := remoteagent.A2AConfig{
			Name:              rc.Name,
			Description:       fmt.Sprintf("Remote A2A agent: %s", rc.Name),
			AgentCardProvider: remoteagent.NewAgentCardProvider(rc.AgentCardURL),
		}

		remoteAgent, err := newRemoteA2AFn(a2aCfg)
		if err != nil {
			loadErrs = append(loadErrs, fmt.Errorf("remote agent %q: %w", rc.Name, err))
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

	return agents, errors.Join(loadErrs...)
}
