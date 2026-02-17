package skill

import (
	"strings"
	"testing"
)

func TestParseSkillMD_Script(t *testing.T) {
	content := `---
name: serve
description: Start the lango server
type: script
status: active
---

` + "```sh\nlango serve\n```\n"

	entry, err := ParseSkillMD([]byte(content))
	if err != nil {
		t.Fatalf("ParseSkillMD: %v", err)
	}

	if entry.Name != "serve" {
		t.Errorf("Name = %q, want %q", entry.Name, "serve")
	}
	if entry.Type != "script" {
		t.Errorf("Type = %q, want %q", entry.Type, "script")
	}
	if entry.Status != "active" {
		t.Errorf("Status = %q, want %q", entry.Status, "active")
	}

	script, ok := entry.Definition["script"].(string)
	if !ok {
		t.Fatal("Definition[\"script\"] not a string")
	}
	if script != "lango serve" {
		t.Errorf("script = %q, want %q", script, "lango serve")
	}
}

func TestParseSkillMD_Template(t *testing.T) {
	content := `---
name: greet
description: Greet someone
type: template
status: active
---

` + "```template\nHello {{.Name}}!\n```\n"

	entry, err := ParseSkillMD([]byte(content))
	if err != nil {
		t.Fatalf("ParseSkillMD: %v", err)
	}

	if entry.Type != "template" {
		t.Errorf("Type = %q, want %q", entry.Type, "template")
	}

	tmpl, ok := entry.Definition["template"].(string)
	if !ok {
		t.Fatal("Definition[\"template\"] not a string")
	}
	if tmpl != "Hello {{.Name}}!" {
		t.Errorf("template = %q, want %q", tmpl, "Hello {{.Name}}!")
	}
}

func TestParseSkillMD_Composite(t *testing.T) {
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
	if err != nil {
		t.Fatalf("ParseSkillMD: %v", err)
	}

	if entry.Type != "composite" {
		t.Errorf("Type = %q, want %q", entry.Type, "composite")
	}

	steps, ok := entry.Definition["steps"].([]interface{})
	if !ok {
		t.Fatal("Definition[\"steps\"] not a []interface{}")
	}
	if len(steps) != 2 {
		t.Fatalf("len(steps) = %d, want 2", len(steps))
	}
}

func TestParseSkillMD_WithParameters(t *testing.T) {
	content := `---
name: greet
description: Greet someone
type: template
status: active
---

` + "```template\nHello {{.Name}}!\n```\n\n" +
		"## Parameters\n\n```json\n{\"type\": \"object\", \"properties\": {\"Name\": {\"type\": \"string\"}}}\n```\n"

	entry, err := ParseSkillMD([]byte(content))
	if err != nil {
		t.Fatalf("ParseSkillMD: %v", err)
	}

	if entry.Parameters == nil {
		t.Fatal("Parameters is nil, want non-nil")
	}
	if _, ok := entry.Parameters["type"]; !ok {
		t.Error("Parameters missing 'type' key")
	}
}

func TestParseSkillMD_MissingFrontmatter(t *testing.T) {
	content := "no frontmatter here"
	_, err := ParseSkillMD([]byte(content))
	if err == nil {
		t.Fatal("expected error for missing frontmatter")
	}
	if !strings.Contains(err.Error(), "frontmatter") {
		t.Errorf("error = %q, want to contain 'frontmatter'", err.Error())
	}
}

func TestParseSkillMD_MissingName(t *testing.T) {
	content := "---\ndescription: test\ntype: script\n---\n\n```sh\necho hi\n```\n"
	_, err := ParseSkillMD([]byte(content))
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("error = %q, want to contain 'name is required'", err.Error())
	}
}

func TestRenderSkillMD_Roundtrip(t *testing.T) {
	original := &SkillEntry{
		Name:        "test-skill",
		Description: "A test skill",
		Type:        "script",
		Status:      "active",
		CreatedBy:   "agent",
		Definition:  map[string]interface{}{"script": "echo hello"},
	}

	rendered, err := RenderSkillMD(original)
	if err != nil {
		t.Fatalf("RenderSkillMD: %v", err)
	}

	parsed, err := ParseSkillMD(rendered)
	if err != nil {
		t.Fatalf("ParseSkillMD (roundtrip): %v", err)
	}

	if parsed.Name != original.Name {
		t.Errorf("Name = %q, want %q", parsed.Name, original.Name)
	}
	if parsed.Description != original.Description {
		t.Errorf("Description = %q, want %q", parsed.Description, original.Description)
	}
	if parsed.Type != original.Type {
		t.Errorf("Type = %q, want %q", parsed.Type, original.Type)
	}
	if parsed.Status != original.Status {
		t.Errorf("Status = %q, want %q", parsed.Status, original.Status)
	}

	script, _ := parsed.Definition["script"].(string)
	if script != "echo hello" {
		t.Errorf("script = %q, want %q", script, "echo hello")
	}
}
