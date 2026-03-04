package agentmemory

// Store is the interface for agent memory storage.
type Store interface {
	// Save upserts a memory entry (matched by agent_name + key).
	Save(entry *Entry) error

	// Get retrieves a specific entry by agent name and key.
	Get(agentName, key string) (*Entry, error)

	// Search finds entries matching criteria.
	Search(agentName string, opts SearchOptions) ([]*Entry, error)

	// SearchWithContext resolves entries with scope fallback:
	// instance (agent_name) > type (all agents of same type) > global.
	SearchWithContext(agentName string, query string, limit int) ([]*Entry, error)

	// Delete removes an entry.
	Delete(agentName, key string) error

	// IncrementUseCount bumps the use counter for an entry.
	IncrementUseCount(agentName, key string) error

	// Prune removes entries below a confidence threshold.
	Prune(agentName string, minConfidence float64) (int, error)

	// ListAgentNames returns the names of all agents that have stored memories.
	ListAgentNames() ([]string, error)

	// ListAll returns all entries for a given agent.
	ListAll(agentName string) ([]*Entry, error)
}

// SearchOptions configures a memory search query.
type SearchOptions struct {
	Query         string
	Scope         MemoryScope
	Kind          MemoryKind
	Tags          []string
	MinConfidence float64
	Limit         int
}
