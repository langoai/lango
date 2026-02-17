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
)

// PromptSection produces a titled block of text for the system prompt.
type PromptSection interface {
	ID() SectionID
	Priority() int // Lower = first. Identity=100, Safety=200, ...
	Render() string // Empty string = omitted
}
