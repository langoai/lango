package config

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	"github.com/langoai/lango/internal/types"
)

// contextRelatedKeys lists config keys tracked for explicit override detection.
var contextRelatedKeys = []string{
	"knowledge.enabled",
	"observationalMemory.enabled",
	"retrieval.enabled",
	"librarian.enabled",
	"graph.enabled",
	"embedding.provider",
}

// ContextRelatedKeys returns a copy of the config keys tracked for explicit override detection.
func ContextRelatedKeys() []string {
	out := make([]string, len(contextRelatedKeys))
	copy(out, contextRelatedKeys)
	return out
}

// AutoEnabledSet records which context subsystems were auto-enabled.
type AutoEnabledSet struct {
	Knowledge bool `json:"knowledge,omitempty"`
	Memory    bool `json:"memory,omitempty"`
	Retrieval bool `json:"retrieval,omitempty"`
	Embedding bool `json:"embedding,omitempty"`
}

// collectExplicitKeys reads the raw config file and checks which of the given
// dotted keys are present. Uses raw JSON parsing (no defaults, no viper) to
// detect only keys the user actually wrote in their config file.
// Returns nil if the file cannot be read.
func collectExplicitKeys(configPath string, keys []string) map[string]bool {
	if configPath == "" {
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}

	result := make(map[string]bool, len(keys))
	for _, key := range keys {
		if nestedKeyExists(raw, key) {
			result[key] = true
		}
	}
	return result
}

// nestedKeyExists checks if a dotted key path exists in a nested map.
func nestedKeyExists(m map[string]interface{}, key string) bool {
	parts := strings.SplitN(key, ".", 2)
	val, ok := m[parts[0]]
	if !ok {
		return false
	}
	if len(parts) == 1 {
		return true
	}
	sub, ok := val.(map[string]interface{})
	if !ok {
		return false
	}
	return nestedKeyExists(sub, parts[1])
}

// ResolveContextAutoEnable sets Enabled for context subsystems that are not
// explicitly configured by the user when their config-level dependencies are
// detectable. This is a shared resolver: both config.Load() and bootstrap call
// it so that app runtime and CLI diagnostics see the same resolved values.
//
// explicitKeys: keys the user explicitly set. nil means "nothing explicit"
// (legacy profiles, new profiles) → auto-enable all detectable features.
func ResolveContextAutoEnable(cfg *Config, explicitKeys map[string]bool) AutoEnabledSet {
	var set AutoEnabledSet
	hasDBPath := cfg.Session.DatabasePath != ""

	// Knowledge: auto-enable if DB path configured and not explicitly disabled.
	if !explicitKeys["knowledge.enabled"] && !cfg.Knowledge.Enabled {
		if hasDBPath {
			cfg.Knowledge.Enabled = true
			set.Knowledge = true
		}
	}

	// Observational Memory: same conditions as knowledge.
	if !explicitKeys["observationalMemory.enabled"] && !cfg.ObservationalMemory.Enabled {
		if hasDBPath {
			cfg.ObservationalMemory.Enabled = true
			set.Memory = true
		}
	}

	// Retrieval: auto-enable if knowledge will be enabled and not explicitly disabled.
	if !explicitKeys["retrieval.enabled"] && !cfg.Retrieval.Enabled {
		if cfg.Knowledge.Enabled {
			cfg.Retrieval.Enabled = true
			set.Retrieval = true
		}
	}

	// Embedding: auto-detect provider if not explicitly configured.
	if !explicitKeys["embedding.provider"] && cfg.Embedding.Provider == "" {
		if probed := cfg.ProbeEmbeddingProvider(); probed != "" {
			cfg.Embedding.Provider = probed
			set.Embedding = true
		}
	}

	return set
}

// ProbeEmbeddingProvider scans the configured providers map for an
// embedding-capable provider and returns its key. Returns "" if no suitable
// provider is found.
//
// Policy (conservative, cost-aware):
//   - Local-first: Ollama always preferred if configured
//   - Single-remote-only: if exactly one remote embedding-capable provider → use it
//   - Multiple-remote → no auto-select (require explicit embedding.provider)
func (c *Config) ProbeEmbeddingProvider() string {
	if c.Embedding.Provider != "" {
		return c.Embedding.Provider
	}

	var localKey string
	var remoteKeys []string

	for key, pCfg := range c.Providers {
		bt := ProviderTypeToEmbeddingType[pCfg.Type]
		if bt == "" {
			continue // not embedding-capable (e.g., anthropic)
		}
		if pCfg.Type == types.ProviderOllama {
			localKey = key
		} else {
			remoteKeys = append(remoteKeys, key)
		}
	}

	// Local-first: prefer Ollama.
	if localKey != "" {
		return localKey
	}

	// Single-remote-only: auto-select only if exactly one remote provider.
	if len(remoteKeys) == 1 {
		return remoteKeys[0]
	}

	// Multiple remote or none → no auto-select.
	return ""
}

// PresetExplicitKeys returns the set of config keys that a preset explicitly sets.
// This allows callers to record preset-set fields as explicit, preventing
// auto-enable from overriding preset intentions.
func PresetExplicitKeys(name string) map[string]bool {
	keys := make(map[string]bool)
	switch PresetName(name) {
	case PresetResearcher:
		keys["knowledge.enabled"] = true
		keys["observationalMemory.enabled"] = true
		keys["graph.enabled"] = true
		keys["embedding.provider"] = true
		keys["librarian.enabled"] = true
	case PresetFull:
		keys["knowledge.enabled"] = true
		keys["observationalMemory.enabled"] = true
		keys["graph.enabled"] = true
		keys["embedding.provider"] = true
		keys["librarian.enabled"] = true
	case PresetCollaborator:
		// No context-related keys set by this preset.
	case PresetMinimal:
		// No keys set — all auto-enable eligible.
	}

	// Sort for determinism in tests (map iteration order).
	_ = sort.Strings
	return keys
}
