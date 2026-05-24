package config

import (
	"strings"

	"github.com/langoai/lango/internal/types"
)

// AgentSetupStatus captures whether the current agent-facing profile is ready
// for a normal interactive chat flow.
type AgentSetupStatus struct {
	ProviderID   string
	Model        string
	ProviderType types.ProviderType

	MissingProvider     bool
	MissingModel        bool
	MissingProviderMap  bool
	MissingProviderType bool
	MissingAPIKey       bool
}

// EvaluateAgentSetup inspects the active agent/provider settings and returns a
// normalized readiness snapshot used by UI and validation surfaces.
func EvaluateAgentSetup(cfg *Config) AgentSetupStatus {
	if cfg == nil {
		return AgentSetupStatus{
			MissingProvider: true,
			MissingModel:    true,
		}
	}

	status := AgentSetupStatus{
		ProviderID: strings.TrimSpace(cfg.Agent.Provider),
		Model:      strings.TrimSpace(cfg.Agent.Model),
	}

	if status.ProviderID == "" {
		status.MissingProvider = true
	}
	if status.Model == "" {
		status.MissingModel = true
	}
	if status.MissingProvider {
		return status
	}

	providerCfg, ok := cfg.Providers[status.ProviderID]
	if !ok {
		status.MissingProviderMap = true
		return status
	}

	status.ProviderType = providerCfg.Type
	if strings.TrimSpace(string(providerCfg.Type)) == "" {
		status.MissingProviderType = true
		return status
	}

	if strings.EqualFold(strings.TrimSpace(string(providerCfg.Type)), "ollama") {
		return status
	}

	if strings.TrimSpace(providerCfg.APIKey) == "" {
		status.MissingAPIKey = true
	}

	return status
}

// Ready reports whether the profile has a usable provider/model path for a
// normal interactive turn.
func (s AgentSetupStatus) Ready() bool {
	return !s.MissingProvider &&
		!s.MissingModel &&
		!s.MissingProviderMap &&
		!s.MissingProviderType &&
		!s.MissingAPIKey
}
