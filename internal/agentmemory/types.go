package agentmemory

import "time"

// MemoryScope defines the visibility of a memory entry.
type MemoryScope string

const (
	ScopeInstance MemoryScope = "instance" // specific to one agent instance
	ScopeType     MemoryScope = "type"     // shared across agents of same type
	ScopeGlobal   MemoryScope = "global"   // shared across all agents
)

// MemoryKind categorizes memory entries.
type MemoryKind string

const (
	KindPattern    MemoryKind = "pattern"    // learned tool usage patterns
	KindPreference MemoryKind = "preference" // user/agent preferences
	KindFact       MemoryKind = "fact"       // discovered facts
	KindSkill      MemoryKind = "skill"      // learned capabilities
)

// Entry represents a single agent memory entry.
type Entry struct {
	ID         string      `json:"id"`
	AgentName  string      `json:"agent_name"`
	Scope      MemoryScope `json:"scope"`
	Kind       MemoryKind  `json:"kind"`
	Key        string      `json:"key"`
	Content    string      `json:"content"`
	Confidence float64     `json:"confidence"` // 0.0-1.0
	UseCount   int         `json:"use_count"`
	Tags       []string    `json:"tags,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}
