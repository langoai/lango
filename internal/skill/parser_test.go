package skill

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSkillMD_Script(t *testing.T) {
	t.Parallel()

	content := `---
name: serve
description: Start the lango server
type: script
status: active
---

` + "```sh\nlango serve\n```\n"

	entry, err := ParseSkillMD([]byte(content))
	require.NoError(t, err)

	assert.Equal(t, "serve", entry.Name)
	assert.Equal(t, SkillTypeScript, entry.Type)
	assert.Equal(t, SkillStatusActive, entry.Status)

	script, ok := entry.Definition["script"].(string)
	require.True(t, ok, "Definition[\"script\"] not a string")
	assert.Equal(t, "lango serve", script)
}

func TestParseSkillMD_Template(t *testing.T) {
	t.Parallel()

	content := `---
name: greet
description: Greet someone
type: template
status: active
---

` + "```template\nHello {{.Name}}!\n```\n"

	entry, err := ParseSkillMD([]byte(content))
	require.NoError(t, err)

	assert.Equal(t, SkillTypeTemplate, entry.Type)

	tmpl, ok := entry.Definition["template"].(string)
	require.True(t, ok, "Definition[\"template\"] not a string")
	assert.Equal(t, "Hello {{.Name}}!", tmpl)
}

func TestParseSkillMD_Composite(t *testing.T) {
	t.Parallel()

	content := `---
name: deploy
description: Deploy workflow
type: composite
status: active
---

### Step 1

` + "```json\n{\"tool\": \"exec\", \"params\": {\"command\": \"build\"}}\n```\n\n" +
		"### Step 2\n\n```json\n{\"tool\": \"exec\", \"params\": {\"command\": \"deploy\"}}\n```\n"

	entry, err := ParseSkillMD([]byte(content))
	require.NoError(t, err)

	assert.Equal(t, SkillTypeComposite, entry.Type)

	steps, ok := entry.Definition["steps"].([]interface{})
	require.True(t, ok, "Definition[\"steps\"] not a []interface{}")
	assert.Len(t, steps, 2)
}

func TestParseSkillMD_WithParameters(t *testing.T) {
	t.Parallel()

	content := `---
name: greet
description: Greet someone
type: template
status: active
---

` + "```template\nHello {{.Name}}!\n```\n\n" +
		"## Parameters\n\n```json\n{\"type\": \"object\", \"properties\": {\"Name\": {\"type\": \"string\"}}}\n```\n"

	entry, err := ParseSkillMD([]byte(content))
	require.NoError(t, err)

	require.NotNil(t, entry.Parameters)
	assert.Contains(t, entry.Parameters, "type")
}

func TestParseSkillMD_MissingFrontmatter(t *testing.T) {
	t.Parallel()

	content := "no frontmatter here"
	_, err := ParseSkillMD([]byte(content))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "frontmatter")
}

func TestParseSkillMD_MissingName(t *testing.T) {
	t.Parallel()

	content := "---\ndescription: test\ntype: script\n---\n\n```sh\necho hi\n```\n"
	_, err := ParseSkillMD([]byte(content))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestParseSkillMD_Instruction(t *testing.T) {
	t.Parallel()

	content := `---
name: obsidian-markdown
description: Obsidian-flavored Markdown reference guide
---

# Obsidian Markdown

Use **bold** and *italic* in Obsidian.

## Links

Use [[wikilinks]] for internal links.`

	entry, err := ParseSkillMD([]byte(content))
	require.NoError(t, err)

	assert.Equal(t, "obsidian-markdown", entry.Name)
	// No explicit type → defaults to "instruction".
	assert.Equal(t, SkillTypeInstruction, entry.Type)
	assert.Equal(t, SkillStatusActive, entry.Status)

	body, ok := entry.Definition["content"].(string)
	require.True(t, ok, "Definition[\"content\"] not a string")
	assert.Contains(t, body, "[[wikilinks]]")
}

func TestRenderSkillMD_Instruction(t *testing.T) {
	t.Parallel()

	original := &SkillEntry{
		Name:        "guide-skill",
		Description: "A guide",
		Type:        "instruction",
		Status:      "active",
		Definition:  map[string]interface{}{"content": "# Guide\n\nSome instructions."},
		Source:      "https://github.com/owner/repo",
	}

	rendered, err := RenderSkillMD(original)
	require.NoError(t, err)

	parsed, err := ParseSkillMD(rendered)
	require.NoError(t, err)

	assert.Equal(t, SkillTypeInstruction, parsed.Type)
	assert.Equal(t, "https://github.com/owner/repo", parsed.Source)
	content, _ := parsed.Definition["content"].(string)
	assert.Contains(t, content, "Some instructions.")
}

func TestParseSkillMD_WithSource(t *testing.T) {
	t.Parallel()

	content := `---
name: imported-skill
description: An imported skill
type: instruction
source: https://github.com/owner/repo
---

Reference content here.`

	entry, err := ParseSkillMD([]byte(content))
	require.NoError(t, err)

	assert.Equal(t, "https://github.com/owner/repo", entry.Source)

	// Render and re-parse to test roundtrip.
	rendered, err := RenderSkillMD(entry)
	require.NoError(t, err)

	reparsed, err := ParseSkillMD(rendered)
	require.NoError(t, err)
	assert.Equal(t, entry.Source, reparsed.Source)
}

func TestParseSkillMD_AllowedTools(t *testing.T) {
	t.Parallel()

	content := `---
name: deploy-skill
description: Deployment skill
type: composite
status: active
allowed-tools: exec fs_write fs_read
---

### Step 1

` + "```json\n{\"tool\": \"exec\", \"params\": {\"command\": \"deploy\"}}\n```\n"

	entry, err := ParseSkillMD([]byte(content))
	require.NoError(t, err)

	require.Len(t, entry.AllowedTools, 3)
	assert.Equal(t, []string{"exec", "fs_write", "fs_read"}, entry.AllowedTools)
}

func TestRenderSkillMD_AllowedTools_Roundtrip(t *testing.T) {
	t.Parallel()

	original := &SkillEntry{
		Name:         "deploy-skill",
		Description:  "Deployment skill",
		Type:         "script",
		Status:       "active",
		Definition:   map[string]interface{}{"script": "echo deploy"},
		AllowedTools: []string{"exec", "fs_write"},
	}

	rendered, err := RenderSkillMD(original)
	require.NoError(t, err)

	parsed, err := ParseSkillMD(rendered)
	require.NoError(t, err)

	require.Len(t, parsed.AllowedTools, 2)
	assert.Equal(t, []string{"exec", "fs_write"}, parsed.AllowedTools)
}

func TestParseSkillMD_NoAllowedTools(t *testing.T) {
	t.Parallel()

	content := `---
name: basic-skill
description: Basic skill
type: script
status: active
---

` + "```sh\necho hello\n```\n"

	entry, err := ParseSkillMD([]byte(content))
	require.NoError(t, err)

	assert.Empty(t, entry.AllowedTools)
}

func TestRenderSkillMD_Roundtrip(t *testing.T) {
	t.Parallel()

	original := &SkillEntry{
		Name:        "test-skill",
		Description: "A test skill",
		Type:        "script",
		Status:      "active",
		CreatedBy:   "agent",
		Definition:  map[string]interface{}{"script": "echo hello"},
	}

	rendered, err := RenderSkillMD(original)
	require.NoError(t, err)

	parsed, err := ParseSkillMD(rendered)
	require.NoError(t, err)

	assert.Equal(t, original.Name, parsed.Name)
	assert.Equal(t, original.Description, parsed.Description)
	assert.Equal(t, original.Type, parsed.Type)
	assert.Equal(t, original.Status, parsed.Status)

	script, _ := parsed.Definition["script"].(string)
	assert.Equal(t, "echo hello", script)
}

func TestSkillTypeFork_Valid(t *testing.T) {
	t.Parallel()

	assert.True(t, SkillTypeFork.Valid())
	assert.Contains(t, SkillTypeFork.Values(), SkillTypeFork)
}

func TestParseSkillMD_V2Fields(t *testing.T) {
	t.Parallel()

	content := `---
name: refactor-skill
description: Refactoring assistant
type: fork
status: active
when_to_use: When the user asks for a multi-file refactor
paths: "src/**/*.go internal/**/*.go"
context: Always check tests after changes
model: claude-opus-4
effort: high
agent: core-developer
hooks:
  pre: go vet ./...
  post: go test ./...
---

# Refactor Guide

Follow the refactoring plan carefully.`

	entry, err := ParseSkillMD([]byte(content))
	require.NoError(t, err)

	assert.Equal(t, "refactor-skill", entry.Name)
	assert.Equal(t, SkillTypeFork, entry.Type)
	assert.Equal(t, "When the user asks for a multi-file refactor", entry.WhenToUse)
	assert.Equal(t, []string{"src/**/*.go", "internal/**/*.go"}, entry.Paths)
	assert.Equal(t, "Always check tests after changes", entry.Context)
	assert.Equal(t, "claude-opus-4", entry.Model)
	assert.Equal(t, "high", entry.Effort)
	assert.Equal(t, "core-developer", entry.Agent)
	require.Len(t, entry.Hooks, 2)
	assert.Equal(t, "go vet ./...", entry.Hooks["pre"])
	assert.Equal(t, "go test ./...", entry.Hooks["post"])
}

func TestRenderSkillMD_V2Fields_Roundtrip(t *testing.T) {
	t.Parallel()

	original := &SkillEntry{
		Name:        "refactor-skill",
		Description: "Refactoring assistant",
		Type:        SkillTypeFork,
		Status:      SkillStatusActive,
		Definition:  map[string]interface{}{"content": "# Refactor Guide"},
		WhenToUse:   "When user asks for refactoring",
		Paths:       []string{"src/**/*.go", "internal/**/*.go"},
		Context:     "Check tests after changes",
		Model:       "claude-opus-4",
		Effort:      "high",
		Agent:       "core-developer",
		Hooks:       map[string]string{"pre": "go vet ./...", "post": "go test ./..."},
	}

	rendered, err := RenderSkillMD(original)
	require.NoError(t, err)

	parsed, err := ParseSkillMD(rendered)
	require.NoError(t, err)

	assert.Equal(t, original.Name, parsed.Name)
	assert.Equal(t, original.Type, parsed.Type)
	assert.Equal(t, original.WhenToUse, parsed.WhenToUse)
	assert.Equal(t, original.Paths, parsed.Paths)
	assert.Equal(t, original.Context, parsed.Context)
	assert.Equal(t, original.Model, parsed.Model)
	assert.Equal(t, original.Effort, parsed.Effort)
	assert.Equal(t, original.Agent, parsed.Agent)
	assert.Equal(t, original.Hooks, parsed.Hooks)
}

func TestParseSkillMD_V2Fields_Empty(t *testing.T) {
	t.Parallel()

	// Existing SKILL.md without new fields must parse correctly (backward compat).
	content := `---
name: legacy-skill
description: Old skill
type: script
status: active
---

` + "```sh\necho legacy\n```\n"

	entry, err := ParseSkillMD([]byte(content))
	require.NoError(t, err)

	assert.Equal(t, "legacy-skill", entry.Name)
	assert.Empty(t, entry.WhenToUse)
	assert.Empty(t, entry.Paths)
	assert.Empty(t, entry.Context)
	assert.Empty(t, entry.Model)
	assert.Empty(t, entry.Effort)
	assert.Empty(t, entry.Agent)
	assert.Empty(t, entry.Hooks)
}

func TestRenderSkillMD_V2Fields_OmitEmpty(t *testing.T) {
	t.Parallel()

	// When v2 fields are empty, they should not appear in rendered output.
	original := &SkillEntry{
		Name:        "minimal-skill",
		Description: "Minimal",
		Type:        SkillTypeScript,
		Status:      SkillStatusActive,
		Definition:  map[string]interface{}{"script": "echo hi"},
	}

	rendered, err := RenderSkillMD(original)
	require.NoError(t, err)

	renderedStr := string(rendered)
	assert.NotContains(t, renderedStr, "when_to_use")
	assert.NotContains(t, renderedStr, "paths")
	assert.NotContains(t, renderedStr, "context")
	assert.NotContains(t, renderedStr, "model")
	assert.NotContains(t, renderedStr, "effort")
	assert.NotContains(t, renderedStr, "agent")
	assert.NotContains(t, renderedStr, "hooks")

	// Also verify it still parses correctly.
	parsed, err := ParseSkillMD(rendered)
	require.NoError(t, err)
	assert.Equal(t, "minimal-skill", parsed.Name)
}

func TestRenderSkillMD_V2Fields_PartialFill(t *testing.T) {
	t.Parallel()

	// Only some v2 fields set — others should be omitted.
	original := &SkillEntry{
		Name:        "partial-skill",
		Description: "Partial v2 fields",
		Type:        SkillTypeInstruction,
		Status:      SkillStatusActive,
		Definition:  map[string]interface{}{"content": "Some content"},
		WhenToUse:   "When doing partial work",
		Effort:      "medium",
	}

	rendered, err := RenderSkillMD(original)
	require.NoError(t, err)

	renderedStr := string(rendered)
	assert.Contains(t, renderedStr, "when_to_use")
	assert.Contains(t, renderedStr, "effort")
	assert.NotContains(t, renderedStr, "paths")
	assert.NotContains(t, renderedStr, "model")
	assert.NotContains(t, renderedStr, "agent")
	assert.NotContains(t, renderedStr, "hooks")

	parsed, err := ParseSkillMD(rendered)
	require.NoError(t, err)
	assert.Equal(t, "When doing partial work", parsed.WhenToUse)
	assert.Equal(t, "medium", parsed.Effort)
	assert.Empty(t, parsed.Paths)
	assert.Empty(t, parsed.Hooks)
}
