package skill

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/agent"
)

// Registry manages skill lifecycle and converts file-based skills to executable tools.
type Registry struct {
	store     SkillStore
	executor  *Executor
	baseTools []*agent.Tool
	logger    *zap.SugaredLogger
	mu        sync.RWMutex
	loaded    []*agent.Tool
}

// NewRegistry creates a new skill registry.
func NewRegistry(store SkillStore, baseTools []*agent.Tool, logger *zap.SugaredLogger) *Registry {
	return &Registry{
		store:     store,
		executor:  NewExecutor(logger),
		baseTools: baseTools,
		logger:    logger,
	}
}

// LoadSkills loads active skills from the store and converts them to agent tools.
func (r *Registry) LoadSkills(ctx context.Context) error {
	skills, err := r.store.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("load active skills: %w", err)
	}

	tools := make([]*agent.Tool, 0, len(skills))
	for _, sk := range skills {
		tool := r.skillToTool(sk)
		tools = append(tools, tool)
	}

	r.mu.Lock()
	r.loaded = tools
	r.mu.Unlock()

	r.logger.Infof("loaded %d active skills", len(tools))
	return nil
}

// AllTools returns baseTools combined with loaded dynamic skills.
func (r *Registry) AllTools() []*agent.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*agent.Tool, 0, len(r.baseTools)+len(r.loaded))
	result = append(result, r.baseTools...)
	result = append(result, r.loaded...)
	return result
}

// CreateSkill validates and saves a new skill.
func (r *Registry) CreateSkill(ctx context.Context, entry SkillEntry) error {
	if entry.Name == "" {
		return fmt.Errorf("skill name is required")
	}
	if entry.Type != "composite" && entry.Type != "script" && entry.Type != "template" && entry.Type != "instruction" {
		return fmt.Errorf("skill type must be composite, script, template, or instruction")
	}
	if entry.Type != "instruction" && len(entry.Definition) == 0 {
		return fmt.Errorf("skill definition is required")
	}

	if entry.Type == "script" {
		scriptRaw, ok := entry.Definition["script"]
		if !ok {
			return fmt.Errorf("script skill must have 'script' in definition")
		}
		script, ok := scriptRaw.(string)
		if !ok {
			return fmt.Errorf("script skill 'script' must be a string")
		}
		if err := r.executor.ValidateScript(script); err != nil {
			return err
		}
	}

	return r.store.Save(ctx, entry)
}

// ActivateSkill activates a skill and reloads the skill tools.
func (r *Registry) ActivateSkill(ctx context.Context, name string) error {
	if err := r.store.Activate(ctx, name); err != nil {
		return err
	}
	return r.LoadSkills(ctx)
}

// LoadedSkills returns only the dynamically loaded skill tools (no base tools).
func (r *Registry) LoadedSkills() []*agent.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*agent.Tool, len(r.loaded))
	copy(result, r.loaded)
	return result
}

// GetSkillTool returns a specific loaded skill tool by name.
func (r *Registry) GetSkillTool(name string) (*agent.Tool, bool) {
	toolName := "skill_" + name

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, t := range r.loaded {
		if t.Name == toolName {
			return t, true
		}
	}
	return nil, false
}

// ListActiveSkills returns all active skill entries from the store.
func (r *Registry) ListActiveSkills(ctx context.Context) ([]SkillEntry, error) {
	return r.store.ListActive(ctx)
}

// Store returns the underlying SkillStore for direct access (e.g. by the importer).
func (r *Registry) Store() SkillStore {
	return r.store
}

func (r *Registry) skillToTool(sk SkillEntry) *agent.Tool {
	skillEntry := sk

	// instruction skills: Description is the agent's reasoning basis for when to invoke.
	// Handler returns the full reference document for agent context loading.
	if skillEntry.Type == "instruction" {
		desc := skillEntry.Description
		if desc == "" {
			desc = fmt.Sprintf("Reference guide for %s", skillEntry.Name)
		}

		params := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"topic": map[string]interface{}{
					"type":        "string",
					"description": "Optional: specific topic or section to focus on within the skill reference",
				},
			},
		}
		if skillEntry.Parameters != nil && len(skillEntry.Parameters) > 0 {
			params = skillEntry.Parameters
		}

		return &agent.Tool{
			Name:        "skill_" + skillEntry.Name,
			Description: desc,
			Parameters:  params,
			Handler: func(ctx context.Context, p map[string]interface{}) (interface{}, error) {
				content, _ := skillEntry.Definition["content"].(string)
				return map[string]interface{}{
					"skill":       skillEntry.Name,
					"type":        "instruction",
					"content":     content,
					"source":      skillEntry.Source,
					"description": skillEntry.Description,
				}, nil
			},
		}
	}

	params := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
	if skillEntry.Parameters != nil {
		params = skillEntry.Parameters
	}

	return &agent.Tool{
		Name:        "skill_" + skillEntry.Name,
		Description: skillEntry.Description,
		Parameters:  params,
		Handler: func(ctx context.Context, p map[string]interface{}) (interface{}, error) {
			return r.executor.Execute(ctx, skillEntry, p)
		},
	}
}
