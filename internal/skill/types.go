package skill

// SkillStatus represents the lifecycle status of a skill.
type SkillStatus string

const (
	SkillStatusDraft    SkillStatus = "draft"
	SkillStatusActive   SkillStatus = "active"
	SkillStatusDisabled SkillStatus = "disabled"
)

// Valid reports whether s is a known skill status.
func (s SkillStatus) Valid() bool {
	switch s {
	case SkillStatusDraft, SkillStatusActive, SkillStatusDisabled:
		return true
	}
	return false
}

// Values returns all known skill statuses.
func (s SkillStatus) Values() []SkillStatus {
	return []SkillStatus{SkillStatusDraft, SkillStatusActive, SkillStatusDisabled}
}

// SkillType represents the kind of skill definition.
type SkillType string

const (
	SkillTypeInstruction SkillType = "instruction"
	SkillTypeComposite   SkillType = "composite"
	SkillTypeScript      SkillType = "script"
	SkillTypeTemplate    SkillType = "template"
)

// Valid reports whether t is a known skill type.
func (t SkillType) Valid() bool {
	switch t {
	case SkillTypeInstruction, SkillTypeComposite, SkillTypeScript, SkillTypeTemplate:
		return true
	}
	return false
}

// Values returns all known skill types.
func (t SkillType) Values() []SkillType {
	return []SkillType{SkillTypeInstruction, SkillTypeComposite, SkillTypeScript, SkillTypeTemplate}
}

// SkillEntry is the domain type for skill CRUD operations.
// Replaces the former knowledge.SkillEntry, removing usage tracking fields.
type SkillEntry struct {
	Name             string
	Description      string
	Type             SkillType
	Definition       map[string]interface{}
	Parameters       map[string]interface{}
	Status           SkillStatus
	CreatedBy        string
	RequiresApproval bool
	Source       string   // import source URL (empty for locally created)
	AllowedTools []string // pre-approved tools (from "allowed-tools" frontmatter)
}
