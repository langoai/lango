package agentregistry

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed defaults/*/AGENT.md
var defaultAgents embed.FS

// EmbeddedStore loads agent definitions from the embedded defaults/ directory.
type EmbeddedStore struct{}

// NewEmbeddedStore creates an EmbeddedStore.
func NewEmbeddedStore() *EmbeddedStore {
	return &EmbeddedStore{}
}

// Load reads all embedded AGENT.md files and returns parsed definitions.
func (s *EmbeddedStore) Load() ([]*AgentDefinition, error) {
	var defs []*AgentDefinition

	entries, err := fs.ReadDir(defaultAgents, "defaults")
	if err != nil {
		return nil, fmt.Errorf("read embedded defaults: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := "defaults/" + entry.Name() + "/AGENT.md"
		data, err := defaultAgents.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read embedded %q: %w", path, err)
		}

		def, err := ParseAgentMD(data)
		if err != nil {
			return nil, fmt.Errorf("parse embedded %q: %w", path, err)
		}
		def.Source = SourceEmbedded
		defs = append(defs, def)
	}

	return defs, nil
}
