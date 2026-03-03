package agentregistry

// Store is the interface for loading agent definitions from a source.
type Store interface {
	// Load returns all agent definitions from this store.
	Load() ([]*AgentDefinition, error)
}
