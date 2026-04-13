package config

// ContextProfileName identifies a named context configuration preset.
type ContextProfileName string

const (
	// ContextProfileOff disables all context subsystems.
	ContextProfileOff ContextProfileName = "off"
	// ContextProfileLite enables knowledge and observational memory only.
	ContextProfileLite ContextProfileName = "lite"
	// ContextProfileBalanced enables knowledge, memory, and librarian.
	ContextProfileBalanced ContextProfileName = "balanced"
	// ContextProfileFull enables all context subsystems including graph.
	ContextProfileFull ContextProfileName = "full"
)

// ValidContextProfiles lists accepted profile names for validation.
var ValidContextProfiles = map[ContextProfileName]bool{
	ContextProfileOff:      true,
	ContextProfileLite:     true,
	ContextProfileBalanced: true,
	ContextProfileFull:     true,
}

// ApplyContextProfile sets context subsystem defaults based on the named profile.
// Fields explicitly set by the user (tracked in explicitKeys) are never overwritten.
// If cfg.ContextProfile is empty, this is a no-op.
func ApplyContextProfile(cfg *Config, explicitKeys map[string]bool) {
	if cfg.ContextProfile == "" {
		return
	}

	profile := cfg.ContextProfile

	type profileDefaults struct {
		knowledgeEnabled bool
		memoryEnabled    bool
		librarianEnabled bool
		graphEnabled     bool
	}

	presets := map[ContextProfileName]profileDefaults{
		ContextProfileOff:      {false, false, false, false},
		ContextProfileLite:     {true, true, false, false},
		ContextProfileBalanced: {true, true, true, false},
		ContextProfileFull:     {true, true, true, true},
	}

	defaults, ok := presets[profile]
	if !ok {
		return // invalid profile — will be caught by Validate()
	}

	if !explicitKeys["knowledge.enabled"] {
		cfg.Knowledge.Enabled = defaults.knowledgeEnabled
	}
	if !explicitKeys["observationalMemory.enabled"] {
		cfg.ObservationalMemory.Enabled = defaults.memoryEnabled
	}
	if !explicitKeys["librarian.enabled"] {
		cfg.Librarian.Enabled = defaults.librarianEnabled
	}
	if !explicitKeys["graph.enabled"] {
		cfg.Graph.Enabled = defaults.graphEnabled
	}
}
