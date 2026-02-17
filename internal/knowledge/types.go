package knowledge

// ContextLayer represents the 6 context layers in the self-learning architecture.
type ContextLayer int

const (
	LayerToolRegistry      ContextLayer = iota + 1
	LayerUserKnowledge                  // User rules, preferences, definitions, facts
	LayerSkillPatterns                  // Known working tool chains and workflows
	LayerExternalKnowledge              // Docs, wiki, MCP integration
	LayerAgentLearnings                 // Error patterns, discovered fixes
	LayerRuntimeContext                 // Session history, tool results, env state
	LayerObservations                   // Compressed conversation observations
	LayerReflections                    // Condensed observation reflections
)

// ContextItem represents a single item from any context layer.
type ContextItem struct {
	Layer    ContextLayer
	Key      string
	Content  string
	Score    float64
	Source   string
	Category string
}

// RetrievalRequest specifies what context to retrieve.
type RetrievalRequest struct {
	Query       string
	SessionKey  string
	Tags        []string
	Layers      []ContextLayer // nil means all layers
	MaxPerLayer int            // 0 uses config default
}

// RetrievalResult contains retrieved context items grouped by layer.
type RetrievalResult struct {
	Items      map[ContextLayer][]ContextItem
	TotalItems int
}

// KnowledgeEntry is the domain type for knowledge CRUD operations.
type KnowledgeEntry struct {
	Key      string
	Category string
	Content  string
	Tags     []string
	Source   string
}

// LearningEntry is the domain type for learning CRUD operations.
type LearningEntry struct {
	Trigger      string
	ErrorPattern string
	Diagnosis    string
	Fix          string
	Category     string
	Tags         []string
}

// AuditEntry is the domain type for audit log writes.
type AuditEntry struct {
	SessionKey string
	Action     string
	Actor      string
	Target     string
	Details    map[string]interface{}
}

// ExternalRefEntry is the domain type for external reference CRUD operations.
type ExternalRefEntry struct {
	Name     string
	RefType  string
	Location string
	Summary  string
	Metadata map[string]interface{}
}

// ToolDescriptor describes a single tool available to the agent.
type ToolDescriptor struct {
	Name        string
	Description string
}

// RuntimeContext holds the current session and system state.
type RuntimeContext struct {
	SessionKey        string
	ChannelType       string // "telegram", "discord", "slack", "direct"
	ActiveToolCount   int
	EncryptionEnabled bool
	KnowledgeEnabled  bool
	MemoryEnabled     bool
}
