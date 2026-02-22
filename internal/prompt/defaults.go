package prompt

import (
	"github.com/langoai/lango/prompts"
)

const (
	fallbackIdentity          = "You are Lango, a powerful AI assistant."
	fallbackSafety            = "Never expose secrets. Confirm before destructive operations."
	fallbackConversationRules = "Focus on the current question. Do not repeat previous answers."
	fallbackToolUsage         = "Prefer read operations before writes. Report errors clearly."
)

// defaultContent reads an embedded prompt file, falling back to a minimal
// string if the read fails (should not happen with correctly embedded files).
func defaultContent(filename, fallback string) string {
	data, err := prompts.FS.ReadFile(filename)
	if err != nil {
		return fallback
	}
	return string(data)
}

// DefaultBuilder returns a Builder pre-loaded with the four built-in sections.
func DefaultBuilder() *Builder {
	b := NewBuilder()
	b.Add(NewStaticSection(SectionIdentity, 100, "",
		defaultContent("AGENTS.md", fallbackIdentity)))
	b.Add(NewStaticSection(SectionSafety, 200, "Safety Guidelines",
		defaultContent("SAFETY.md", fallbackSafety)))
	b.Add(NewStaticSection(SectionConversationRules, 300, "Conversation Rules",
		defaultContent("CONVERSATION_RULES.md", fallbackConversationRules)))
	b.Add(NewStaticSection(SectionToolUsage, 400, "Tool Usage Guidelines",
		defaultContent("TOOL_USAGE.md", fallbackToolUsage)))
	return b
}
