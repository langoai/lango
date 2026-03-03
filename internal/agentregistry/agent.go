package agentregistry

// AgentSource indicates where an agent definition originated.
type AgentSource int

const (
	SourceBuiltin  AgentSource = iota // Hardcoded in Go
	SourceEmbedded                    // From embed.FS (defaults/)
	SourceUser                        // From ~/.lango/agents/
	SourceRemote                      // From P2P network
)

// AgentStatus controls whether an agent is active.
type AgentStatus string

const (
	StatusActive   AgentStatus = "active"
	StatusDisabled AgentStatus = "disabled"
	StatusDraft    AgentStatus = "draft"
)

// AgentDefinition is the parsed representation of an AGENT.md file.
type AgentDefinition struct {
	Name             string      `yaml:"name"`
	Description      string      `yaml:"description"`
	Instruction      string      `yaml:"-"` // markdown body
	Status           AgentStatus `yaml:"status"`
	Prefixes         []string    `yaml:"prefixes,omitempty"`
	Keywords         []string    `yaml:"keywords,omitempty"`
	Capabilities     []string    `yaml:"capabilities,omitempty"`
	Accepts          string      `yaml:"accepts,omitempty"`
	Returns          string      `yaml:"returns,omitempty"`
	CannotDo         []string    `yaml:"cannot_do,omitempty"`
	AlwaysInclude    bool        `yaml:"always_include,omitempty"`
	SessionIsolation bool        `yaml:"session_isolation,omitempty"`
	Source           AgentSource `yaml:"-"`
}
