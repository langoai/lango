package skill

// SkillEntry is the domain type for skill CRUD operations.
// Replaces the former knowledge.SkillEntry, removing usage tracking fields.
type SkillEntry struct {
	Name             string
	Description      string
	Type             string // composite, script, template
	Definition       map[string]interface{}
	Parameters       map[string]interface{}
	Status           string // active, draft, disabled
	CreatedBy        string
	RequiresApproval bool
}
