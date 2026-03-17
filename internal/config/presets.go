package config

// PresetName represents a named configuration preset.
type PresetName string

const (
	PresetMinimal      PresetName = "minimal"
	PresetResearcher   PresetName = "researcher"
	PresetCollaborator PresetName = "collaborator"
	PresetFull         PresetName = "full"
)

// PresetInfo describes a preset for display.
type PresetInfo struct {
	Name PresetName
	Desc string
}

// AllPresets returns all available presets with descriptions.
func AllPresets() []PresetInfo {
	return []PresetInfo{
		{PresetMinimal, "Basic AI agent (quick start)"},
		{PresetResearcher, "Knowledge, RAG, Graph (research/analysis)"},
		{PresetCollaborator, "P2P team, payment, workspace (collaboration)"},
		{PresetFull, "All features enabled (power user)"},
	}
}

// IsValidPreset checks if the given name is a valid preset.
func IsValidPreset(name string) bool {
	for _, p := range AllPresets() {
		if string(p.Name) == name {
			return true
		}
	}
	return false
}

// PresetConfig returns a Config for the given preset name.
// Unknown names return DefaultConfig().
func PresetConfig(name string) *Config {
	cfg := DefaultConfig()

	switch PresetName(name) {
	case PresetMinimal:
		return cfg

	case PresetResearcher:
		cfg.Knowledge.Enabled = true
		cfg.ObservationalMemory.Enabled = true
		cfg.Graph.Enabled = true
		cfg.Embedding.Provider = "openai"
		cfg.Embedding.Model = "text-embedding-3-small"
		cfg.Librarian.Enabled = true
		return cfg

	case PresetCollaborator:
		cfg.P2P.Enabled = true
		cfg.Payment.Enabled = true
		cfg.Payment.Network.RPCURL = "https://sepolia.base.org"
		cfg.Economy.Enabled = true
		return cfg

	case PresetFull:
		cfg.Knowledge.Enabled = true
		cfg.ObservationalMemory.Enabled = true
		cfg.Graph.Enabled = true
		cfg.Embedding.Provider = "openai"
		cfg.Embedding.Model = "text-embedding-3-small"
		cfg.Librarian.Enabled = true
		cfg.Cron.Enabled = true
		cfg.Background.Enabled = true
		cfg.Workflow.Enabled = true
		cfg.MCP.Enabled = true
		cfg.AgentMemory.Enabled = true
		cfg.Agent.MultiAgent = true
		return cfg

	default:
		return cfg
	}
}
