package skill

// SkillStep represents one step in a composite skill.
type SkillStep struct {
	Tool   string                 `json:"tool"`
	Params map[string]interface{} `json:"params"`
}

// BuildCompositeSkill creates a SkillEntry for a multi-step tool chain.
func BuildCompositeSkill(name, description string, steps []SkillStep, params map[string]interface{}) SkillEntry {
	stepDefs := make([]interface{}, 0, len(steps))
	for _, s := range steps {
		stepDefs = append(stepDefs, map[string]interface{}{
			"tool":   s.Tool,
			"params": s.Params,
		})
	}

	entry := SkillEntry{
		Name:        name,
		Description: description,
		Type:        "composite",
		Definition: map[string]interface{}{
			"steps": stepDefs,
		},
		RequiresApproval: true,
	}
	if params != nil {
		entry.Parameters = params
	}
	return entry
}

// BuildScriptSkill creates a SkillEntry for a shell script.
func BuildScriptSkill(name, description, script string, params map[string]interface{}) SkillEntry {
	entry := SkillEntry{
		Name:        name,
		Description: description,
		Type:        "script",
		Definition: map[string]interface{}{
			"script": script,
		},
		RequiresApproval: true,
	}
	if params != nil {
		entry.Parameters = params
	}
	return entry
}

// BuildTemplateSkill creates a SkillEntry for a template-based skill.
func BuildTemplateSkill(name, description, tmpl string, params map[string]interface{}) SkillEntry {
	entry := SkillEntry{
		Name:        name,
		Description: description,
		Type:        "template",
		Definition: map[string]interface{}{
			"template": tmpl,
		},
		RequiresApproval: true,
	}
	if params != nil {
		entry.Parameters = params
	}
	return entry
}
