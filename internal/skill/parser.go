package skill

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/langoai/lango/internal/mdparse"
)

// buildSkillBody renders the body portion of a SKILL.md (below the frontmatter).
func buildSkillBody(entry *SkillEntry) (string, error) {
	var buf bytes.Buffer

	switch entry.Type {
	case "script":
		script, _ := entry.Definition["script"].(string)
		buf.WriteString("```sh\n")
		buf.WriteString(script)
		if !strings.HasSuffix(script, "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString("```\n")

	case "template":
		tmpl, _ := entry.Definition["template"].(string)
		buf.WriteString("```template\n")
		buf.WriteString(tmpl)
		if !strings.HasSuffix(tmpl, "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString("```\n")

	case "composite":
		steps, _ := entry.Definition["steps"].([]interface{})
		for i, step := range steps {
			stepJSON, err := json.MarshalIndent(step, "", "  ")
			if err != nil {
				return "", fmt.Errorf("marshal step %d: %w", i, err)
			}
			fmt.Fprintf(&buf, "### Step %d\n\n", i+1)
			buf.WriteString("```json\n")
			buf.Write(stepJSON)
			buf.WriteString("\n```\n\n")
		}

	case "instruction":
		content, _ := entry.Definition["content"].(string)
		buf.WriteString(content)
		if content != "" && !strings.HasSuffix(content, "\n") {
			buf.WriteString("\n")
		}
	}

	if len(entry.Parameters) > 0 {
		paramJSON, err := json.MarshalIndent(entry.Parameters, "", "  ")
		if err != nil {
			return "", fmt.Errorf("marshal parameters: %w", err)
		}
		buf.WriteString("\n## Parameters\n\n```json\n")
		buf.Write(paramJSON)
		buf.WriteString("\n```\n")
	}

	return buf.String(), nil
}

// frontmatter is the YAML frontmatter structure of a SKILL.md file.
type frontmatter struct {
	Name             string `yaml:"name"`
	Description      string `yaml:"description"`
	Type             string `yaml:"type"`
	Status           string `yaml:"status"`
	CreatedBy        string `yaml:"created_by"`
	RequiresApproval bool   `yaml:"requires_approval"`
	Source           string `yaml:"source,omitempty"`
	AllowedTools     string `yaml:"allowed-tools,omitempty"`
}

var _codeBlockRe = regexp.MustCompile("(?s)```(\\w+)?\\s*\n(.*?)```")

// ParseSkillMD parses a SKILL.md file into a SkillEntry.
// Format: YAML frontmatter (between --- delimiters) + markdown body with code blocks.
func ParseSkillMD(content []byte) (*SkillEntry, error) {
	fm, body, err := splitFrontmatter(content)
	if err != nil {
		return nil, err
	}

	var meta frontmatter
	if err := yaml.Unmarshal(fm, &meta); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	if meta.Name == "" {
		return nil, fmt.Errorf("skill name is required in frontmatter")
	}
	if meta.Type == "" {
		meta.Type = "instruction"
	}
	if meta.Status == "" {
		meta.Status = "active"
	}

	var allowedTools []string
	if meta.AllowedTools != "" {
		allowedTools = strings.Fields(meta.AllowedTools)
	}

	entry := &SkillEntry{
		Name:             meta.Name,
		Description:      meta.Description,
		Type:             SkillType(meta.Type),
		Status:           SkillStatus(meta.Status),
		CreatedBy:        meta.CreatedBy,
		RequiresApproval: meta.RequiresApproval,
		Source:           meta.Source,
		AllowedTools:     allowedTools,
	}

	definition, params, err := parseBody(meta.Type, body)
	if err != nil {
		return nil, fmt.Errorf("parse body for skill %q: %w", meta.Name, err)
	}
	entry.Definition = definition
	entry.Parameters = params

	return entry, nil
}

// RenderSkillMD renders a SkillEntry to SKILL.md format.
func RenderSkillMD(entry *SkillEntry) ([]byte, error) {
	status := entry.Status
	if status == "" {
		status = "draft"
	}

	var allowedToolsStr string
	if len(entry.AllowedTools) > 0 {
		allowedToolsStr = strings.Join(entry.AllowedTools, " ")
	}

	meta := frontmatter{
		Name:             entry.Name,
		Description:      entry.Description,
		Type:             string(entry.Type),
		Status:           string(status),
		CreatedBy:        entry.CreatedBy,
		RequiresApproval: entry.RequiresApproval,
		Source:           entry.Source,
		AllowedTools:     allowedToolsStr,
	}

	body, err := buildSkillBody(entry)
	if err != nil {
		return nil, err
	}

	result, err := mdparse.RenderFrontmatter(meta, body)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}
	return result, nil
}

// splitFrontmatter delegates to mdparse.SplitFrontmatter.
var splitFrontmatter = mdparse.SplitFrontmatter

// parseBody extracts Definition and Parameters from the markdown body.
func parseBody(skillType, body string) (definition map[string]interface{}, params map[string]interface{}, err error) {
	definition = make(map[string]interface{})

	blocks := _codeBlockRe.FindAllStringSubmatch(body, -1)

	switch skillType {
	case "script":
		for _, block := range blocks {
			lang := strings.ToLower(block[1])
			if lang == "sh" || lang == "bash" || lang == "" {
				definition["script"] = strings.TrimSpace(block[2])
				break
			}
		}

	case "template":
		for _, block := range blocks {
			lang := strings.ToLower(block[1])
			if lang == "template" || lang == "" {
				definition["template"] = strings.TrimSpace(block[2])
				break
			}
		}

	case "composite":
		var steps []interface{}
		for _, block := range blocks {
			lang := strings.ToLower(block[1])
			if lang == "json" {
				var step map[string]interface{}
				if err := json.Unmarshal([]byte(block[2]), &step); err != nil {
					continue
				}
				steps = append(steps, step)
			}
		}
		if len(steps) > 0 {
			definition["steps"] = steps
		}

	case "instruction":
		// Store the entire body as content (markdown reference document).
		// Strip the ## Parameters section if present — it is parsed separately below.
		content := body
		if idx := strings.Index(content, "## Parameters"); idx >= 0 {
			content = strings.TrimSpace(content[:idx])
		}
		definition["content"] = content
	}

	// Extract parameters section (last json block after ## Parameters header)
	paramIdx := strings.Index(body, "## Parameters")
	if paramIdx >= 0 {
		paramBody := body[paramIdx:]
		paramBlocks := _codeBlockRe.FindAllStringSubmatch(paramBody, -1)
		for _, block := range paramBlocks {
			lang := strings.ToLower(block[1])
			if lang == "json" {
				var p map[string]interface{}
				if err := json.Unmarshal([]byte(block[2]), &p); err == nil {
					params = p
					break
				}
			}
		}
	}

	return definition, params, nil
}
