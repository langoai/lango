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
