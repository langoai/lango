package prompt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultBuilder_ContainsAllSections(t *testing.T) {
	t.Parallel()

	b := DefaultBuilder()
	assert.True(t, b.Has(SectionIdentity))
	assert.True(t, b.Has(SectionSafety))
	assert.True(t, b.Has(SectionConversationRules))
	assert.True(t, b.Has(SectionOutputPrinciples))
	assert.True(t, b.Has(SectionToolUsage))
}

func TestDefaultBuilder_IncludesConversationRules(t *testing.T) {
	t.Parallel()

	result := DefaultBuilder().Build()
	assert.Contains(t, result, "Answer only the current question")
	assert.Contains(t, result, "Do not repeat")
}

func TestDefaultBuilder_IncludesIdentity(t *testing.T) {
	t.Parallel()

	result := DefaultBuilder().Build()
	assert.Contains(t, result, "You are Lango")
}

func TestDefaultBuilder_SectionOrder(t *testing.T) {
	t.Parallel()

	result := DefaultBuilder().Build()
	idxIdentity := strings.Index(result, "You are Lango")
	idxSafety := strings.Index(result, "Safety Guidelines")
	idxConversation := strings.Index(result, "Conversation Rules")
	idxOutput := strings.Index(result, "Output Principles")
	idxTool := strings.Index(result, "Tool Usage Guidelines")

	assert.Less(t, idxIdentity, idxSafety, "Identity should come before Safety")
	assert.Less(t, idxSafety, idxConversation, "Safety should come before Conversation Rules")
	assert.Less(t, idxConversation, idxOutput, "Conversation Rules should come before Output Principles")
	assert.Less(t, idxOutput, idxTool, "Output Principles should come before Tool Usage")
}

func TestDefaultBuilder_UsesEmbeddedContent(t *testing.T) {
	t.Parallel()

	result := DefaultBuilder().Build()
	// Verify embedded content is loaded (not fallbacks)
	assert.Contains(t, result, "twenty-four tool categories")
	assert.Contains(t, result, "Never expose secrets")
	assert.Contains(t, result, "Exec Tool")
}
