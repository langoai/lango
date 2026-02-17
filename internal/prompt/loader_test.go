package prompt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromDir_OverridesKnownFile(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("Custom identity"), 0644)
	require.NoError(t, err)

	b := LoadFromDir(dir, nil)
	result := b.Build()
	assert.Contains(t, result, "Custom identity")
	assert.NotContains(t, result, "You are Lango")
}

func TestLoadFromDir_AddsCustomSection(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "MY_RULES.md"), []byte("Custom rule content"), 0644)
	require.NoError(t, err)

	b := LoadFromDir(dir, nil)
	result := b.Build()
	assert.Contains(t, result, "Custom rule content")
	assert.True(t, b.Has("custom_my_rules"))
}

func TestLoadFromDir_IgnoresEmptyFiles(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(""), 0644)
	require.NoError(t, err)

	b := LoadFromDir(dir, nil)
	result := b.Build()
	// Default identity should remain since the file was empty
	assert.Contains(t, result, "You are Lango")
}

func TestLoadFromDir_IgnoresNonMdFiles(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("not a prompt"), 0644)
	require.NoError(t, err)

	b := LoadFromDir(dir, nil)
	result := b.Build()
	assert.NotContains(t, result, "not a prompt")
}

func TestLoadFromDir_NonExistentDir(t *testing.T) {
	b := LoadFromDir("/nonexistent/path", nil)
	result := b.Build()
	// Should fall back to defaults
	assert.Contains(t, result, "You are Lango")
}

func TestLoadFromDir_OverridesMultipleSections(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("My agent"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SAFETY.md"), []byte("My safety rules"), 0644))

	b := LoadFromDir(dir, nil)
	result := b.Build()
	assert.Contains(t, result, "My agent")
	assert.Contains(t, result, "My safety rules")
	assert.NotContains(t, result, "You are Lango")
	// Conversation rules should still be default
	assert.Contains(t, result, "Answer only the current question")
}

// --- LoadAgentFromDir tests ---

func TestLoadAgentFromDir_OverridesIdentity(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "IDENTITY.md"), []byte("Custom agent identity"), 0644))

	base := NewBuilder()
	base.Add(NewStaticSection(SectionAgentIdentity, 150, "", "Default identity"))
	base.Add(NewStaticSection(SectionSafety, 200, "Safety Guidelines", "Shared safety"))

	b := LoadAgentFromDir(base, dir, nil)
	result := b.Build()
	assert.Contains(t, result, "Custom agent identity")
	assert.NotContains(t, result, "Default identity")
	assert.Contains(t, result, "Shared safety")
}

func TestLoadAgentFromDir_OverridesSafety(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SAFETY.md"), []byte("Agent-specific safety"), 0644))

	base := NewBuilder()
	base.Add(NewStaticSection(SectionSafety, 200, "Safety Guidelines", "Shared safety"))

	b := LoadAgentFromDir(base, dir, nil)
	result := b.Build()
	assert.Contains(t, result, "Agent-specific safety")
	assert.NotContains(t, result, "Shared safety")
}

func TestLoadAgentFromDir_AddsCustomSection(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "MY_RULES.md"), []byte("Custom rules"), 0644))

	base := NewBuilder()
	base.Add(NewStaticSection(SectionSafety, 200, "Safety Guidelines", "Shared safety"))

	b := LoadAgentFromDir(base, dir, nil)
	result := b.Build()
	assert.Contains(t, result, "Custom rules")
	assert.Contains(t, result, "Shared safety")
	assert.True(t, b.Has("custom_my_rules"))
}

func TestLoadAgentFromDir_NonExistentDir(t *testing.T) {
	base := NewBuilder()
	base.Add(NewStaticSection(SectionSafety, 200, "Safety Guidelines", "Shared safety"))

	b := LoadAgentFromDir(base, "/nonexistent/agent/path", nil)
	result := b.Build()
	assert.Contains(t, result, "Shared safety")
}

func TestLoadAgentFromDir_DoesNotMutateBase(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "IDENTITY.md"), []byte("Override"), 0644))

	base := NewBuilder()
	base.Add(NewStaticSection(SectionAgentIdentity, 150, "", "Original"))

	_ = LoadAgentFromDir(base, dir, nil)

	// base should be unchanged
	result := base.Build()
	assert.Contains(t, result, "Original")
	assert.NotContains(t, result, "Override")
}

func TestLoadAgentFromDir_IgnoresEmptyFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "IDENTITY.md"), []byte("   "), 0644))

	base := NewBuilder()
	base.Add(NewStaticSection(SectionAgentIdentity, 150, "", "Original"))

	b := LoadAgentFromDir(base, dir, nil)
	result := b.Build()
	assert.Contains(t, result, "Original")
}

func TestLoadFromDir_CustomSectionPriorityAfterDefaults(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "EXTRA.md"), []byte("extra content"), 0644))

	b := LoadFromDir(dir, nil)
	result := b.Build()
	// Default sections should appear before custom
	idxTool := len(result) // fallback
	for i := range result {
		if i+len("Tool Usage Guidelines") <= len(result) && result[i:i+len("Tool Usage Guidelines")] == "Tool Usage Guidelines" {
			idxTool = i
			break
		}
	}
	idxCustom := len(result)
	for i := range result {
		if i+len("extra content") <= len(result) && result[i:i+len("extra content")] == "extra content" {
			idxCustom = i
			break
		}
	}
	assert.Less(t, idxTool, idxCustom, "default sections should appear before custom sections")
}
