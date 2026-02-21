package prompt

// SectionID identifies a prompt section.
type SectionID string

const (
	SectionIdentity          SectionID = "identity"
	SectionAgentIdentity     SectionID = "agent_identity"
	SectionSafety            SectionID = "safety"
	SectionConversationRules SectionID = "conversation_rules"
	SectionToolUsage         SectionID = "tool_usage"
	SectionCustom            SectionID = "custom"
	SectionAutomation        SectionID = "automation"
)

// Valid reports whether s is a known section ID.
func (s SectionID) Valid() bool {
	switch s {
	case SectionIdentity, SectionAgentIdentity, SectionSafety, SectionConversationRules, SectionToolUsage, SectionCustom, SectionAutomation:
		return true
	}
	return false
}

// Values returns all known section IDs.
func (s SectionID) Values() []SectionID {
	return []SectionID{SectionIdentity, SectionAgentIdentity, SectionSafety, SectionConversationRules, SectionToolUsage, SectionCustom, SectionAutomation}
}

// PromptSection produces a titled block of text for the system prompt.
type PromptSection interface {
	ID() SectionID
	Priority() int // Lower = first. Identity=100, Safety=200, ...
	Render() string // Empty string = omitted
}
