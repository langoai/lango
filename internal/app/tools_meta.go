package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/learning"
	"github.com/langoai/lango/internal/skill"
)

// buildMetaTools creates knowledge/learning/skill meta-tools for the agent.
func buildMetaTools(store *knowledge.Store, engine *learning.Engine, registry *skill.Registry, skillCfg config.SkillConfig) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "save_knowledge",
			Description: "Save a piece of knowledge (user rule, definition, preference, fact, pattern, or correction) for future reference",
			SafetyLevel: agent.SafetyLevelModerate,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key":      map[string]interface{}{"type": "string", "description": "Unique key for this knowledge entry"},
					"category": map[string]interface{}{"type": "string", "description": "Category: rule, definition, preference, fact, pattern, or correction", "enum": []string{"rule", "definition", "preference", "fact", "pattern", "correction"}},
					"content":  map[string]interface{}{"type": "string", "description": "The knowledge content to save"},
					"tags":     map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Optional tags for categorization"},
					"source":   map[string]interface{}{"type": "string", "description": "Where this knowledge came from"},
				},
				"required": []string{"key", "category", "content"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				key, _ := params["key"].(string)
				category, _ := params["category"].(string)
				content, _ := params["content"].(string)
				source, _ := params["source"].(string)

				if key == "" || category == "" || content == "" {
					return nil, fmt.Errorf("key, category, and content are required")
				}

				cat := entknowledge.Category(category)
				if err := entknowledge.CategoryValidator(cat); err != nil {
					return nil, fmt.Errorf("invalid category %q: %w", category, err)
				}

				var tags []string
				if rawTags, ok := params["tags"].([]interface{}); ok {
					for _, t := range rawTags {
						if s, ok := t.(string); ok {
							tags = append(tags, s)
						}
					}
				}

				entry := knowledge.KnowledgeEntry{
					Key:      key,
					Category: cat,
					Content:  content,
					Tags:     tags,
					Source:   source,
				}

				if err := store.SaveKnowledge(ctx, "", entry); err != nil {
					return nil, fmt.Errorf("save knowledge: %w", err)
				}

				if err := store.SaveAuditLog(ctx, knowledge.AuditEntry{
					Action: "knowledge_save",
					Actor:  "agent",
					Target: key,
				}); err != nil {
					logger().Warnw("audit log save failed", "action", "knowledge_save", "error", err)
				}

				return map[string]interface{}{
					"status":  "saved",
					"key":     key,
					"message": fmt.Sprintf("Knowledge '%s' saved successfully", key),
				}, nil
			},
		},
		{
			Name:        "search_knowledge",
			Description: "Search stored knowledge entries by query and optional category",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query":    map[string]interface{}{"type": "string", "description": "Search query"},
					"category": map[string]interface{}{"type": "string", "description": "Optional category filter: rule, definition, preference, or fact"},
				},
				"required": []string{"query"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				query, _ := params["query"].(string)
				category, _ := params["category"].(string)

				entries, err := store.SearchKnowledge(ctx, query, category, 10)
				if err != nil {
					return nil, fmt.Errorf("search knowledge: %w", err)
				}

				return map[string]interface{}{
					"results": entries,
					"count":   len(entries),
				}, nil
			},
		},
		{
			Name:        "save_learning",
			Description: "Save a diagnosed error pattern and its fix for future reference",
			SafetyLevel: agent.SafetyLevelModerate,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"trigger":       map[string]interface{}{"type": "string", "description": "What triggered this learning (e.g., tool name or action)"},
					"error_pattern": map[string]interface{}{"type": "string", "description": "The error pattern to match"},
					"diagnosis":     map[string]interface{}{"type": "string", "description": "Diagnosis of the error cause"},
					"fix":           map[string]interface{}{"type": "string", "description": "The fix or workaround"},
					"category":      map[string]interface{}{"type": "string", "description": "Category: tool_error, provider_error, user_correction, timeout, permission, general"},
				},
				"required": []string{"trigger", "fix"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				trigger, _ := params["trigger"].(string)
				errorPattern, _ := params["error_pattern"].(string)
				diagnosis, _ := params["diagnosis"].(string)
				fix, _ := params["fix"].(string)
				category, _ := params["category"].(string)

				if trigger == "" || fix == "" {
					return nil, fmt.Errorf("trigger and fix are required")
				}
				if category == "" {
					category = "general"
				}

				entry := knowledge.LearningEntry{
					Trigger:      trigger,
					ErrorPattern: errorPattern,
					Diagnosis:    diagnosis,
					Fix:          fix,
					Category:     entlearning.Category(category),
				}

				if err := store.SaveLearning(ctx, "", entry); err != nil {
					return nil, fmt.Errorf("save learning: %w", err)
				}

				if err := store.SaveAuditLog(ctx, knowledge.AuditEntry{
					Action: "learning_save",
					Actor:  "agent",
					Target: trigger,
				}); err != nil {
					logger().Warnw("audit log save failed", "action", "learning_save", "error", err)
				}

				return map[string]interface{}{
					"status":  "saved",
					"message": fmt.Sprintf("Learning for '%s' saved successfully", trigger),
				}, nil
			},
		},
		{
			Name:        "search_learnings",
			Description: "Search stored learnings by error pattern or trigger",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query":    map[string]interface{}{"type": "string", "description": "Search query (error message or trigger)"},
					"category": map[string]interface{}{"type": "string", "description": "Optional category filter"},
				},
				"required": []string{"query"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				query, _ := params["query"].(string)
				category, _ := params["category"].(string)

				entries, err := store.SearchLearnings(ctx, query, category, 10)
				if err != nil {
					return nil, fmt.Errorf("search learnings: %w", err)
				}

				return map[string]interface{}{
					"results": entries,
					"count":   len(entries),
				}, nil
			},
		},
		{
			Name:        "create_skill",
			Description: "Create a new reusable skill from a multi-step workflow, script, or template",
			SafetyLevel: agent.SafetyLevelModerate,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":        map[string]interface{}{"type": "string", "description": "Unique name for the skill"},
					"description": map[string]interface{}{"type": "string", "description": "Description of what the skill does"},
					"type":        map[string]interface{}{"type": "string", "description": "Skill type: composite, script, or template", "enum": []string{"composite", "script", "template"}},
					"definition":  map[string]interface{}{"type": "string", "description": "JSON string of the skill definition"},
					"parameters":  map[string]interface{}{"type": "string", "description": "Optional JSON string of parameter schema"},
				},
				"required": []string{"name", "description", "type", "definition"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				name, _ := params["name"].(string)
				description, _ := params["description"].(string)
				skillType, _ := params["type"].(string)
				definitionStr, _ := params["definition"].(string)

				if name == "" || description == "" || skillType == "" || definitionStr == "" {
					return nil, fmt.Errorf("name, description, type, and definition are required")
				}

				var definition map[string]interface{}
				if err := json.Unmarshal([]byte(definitionStr), &definition); err != nil {
					return nil, fmt.Errorf("parse definition JSON: %w", err)
				}

				var parameters map[string]interface{}
				if paramStr, ok := params["parameters"].(string); ok && paramStr != "" {
					if err := json.Unmarshal([]byte(paramStr), &parameters); err != nil {
						return nil, fmt.Errorf("parse parameters JSON: %w", err)
					}
				}

				entry := skill.SkillEntry{
					Name:             name,
					Description:      description,
					Type:             skill.SkillType(skillType),
					Definition:       definition,
					Parameters:       parameters,
					Status:           skill.SkillStatusActive,
					CreatedBy:        "agent",
					RequiresApproval: false,
				}

				if registry == nil {
					return nil, fmt.Errorf("skill system is not enabled")
				}

				if err := registry.CreateSkill(ctx, entry); err != nil {
					return nil, fmt.Errorf("create skill: %w", err)
				}

				if err := registry.ActivateSkill(ctx, name); err != nil {
					return nil, fmt.Errorf("activate skill: %w", err)
				}

				if err := store.SaveAuditLog(ctx, knowledge.AuditEntry{
					Action: "skill_create",
					Actor:  "agent",
					Target: name,
					Details: map[string]interface{}{
						"type":   skillType,
						"status": "active",
					},
				}); err != nil {
					logger().Warnw("audit log save failed", "action", "skill_create", "error", err)
				}

				return map[string]interface{}{
					"status":  "active",
					"name":    name,
					"message": fmt.Sprintf("Skill '%s' created and activated", name),
				}, nil
			},
		},
		{
			Name:        "list_skills",
			Description: "List all active skills",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				if registry == nil {
					return map[string]interface{}{"skills": []interface{}{}, "count": 0}, nil
				}

				skills, err := registry.ListActiveSkills(ctx)
				if err != nil {
					return nil, fmt.Errorf("list skills: %w", err)
				}

				return map[string]interface{}{
					"skills": skills,
					"count":  len(skills),
				}, nil
			},
		},
		{
			Name: "import_skill",
			Description: "Import skills from a GitHub repository or URL. " +
				"Supports bulk import (all skills from a repo) or single skill import.",
			SafetyLevel: agent.SafetyLevelModerate,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "GitHub repository URL or direct URL to a SKILL.md file",
					},
					"skill_name": map[string]interface{}{
						"type":        "string",
						"description": "Optional: import only this specific skill from the repo",
					},
				},
				"required": []string{"url"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				if !skillCfg.AllowImport {
					return nil, fmt.Errorf("skill import disabled (skill.allowImport=false)")
				}

				if registry == nil {
					return nil, fmt.Errorf("skill system is not enabled")
				}

				url, _ := params["url"].(string)
				skillName, _ := params["skill_name"].(string)

				if url == "" {
					return nil, fmt.Errorf("url is required")
				}

				importer := skill.NewImporter(logger())

				if skill.IsGitHubURL(url) {
					ref, err := skill.ParseGitHubURL(url)
					if err != nil {
						return nil, fmt.Errorf("parse GitHub URL: %w", err)
					}

					if skillName != "" {
						// Single skill import from GitHub (with resource files).
						entry, err := importer.ImportSingleWithResources(ctx, ref, skillName, registry.Store())
						if err != nil {
							return nil, fmt.Errorf("import skill %q: %w", skillName, err)
						}
						if err := registry.LoadSkills(ctx); err != nil {
							return nil, fmt.Errorf("reload skills: %w", err)
						}
						go func() {
							auditCtx, auditCancel := context.WithTimeout(context.Background(), 5*time.Second)
							defer auditCancel()
							if err := store.SaveAuditLog(auditCtx, knowledge.AuditEntry{
								Action: "skill_import",
								Actor:  "agent",
								Target: entry.Name,
								Details: map[string]interface{}{
									"source": url,
									"type":   entry.Type,
								},
							}); err != nil {
								logger().Warnw("audit log save failed", "action", "skill_import", "error", err)
							}
						}()
						return map[string]interface{}{
							"status":  "imported",
							"name":    entry.Name,
							"type":    entry.Type,
							"message": fmt.Sprintf("Skill '%s' imported from %s", entry.Name, url),
						}, nil
					}

					// Bulk import from GitHub repo.
					importCfg := skill.ImportConfig{
						MaxSkills:   skillCfg.MaxBulkImport,
						Concurrency: skillCfg.ImportConcurrency,
						Timeout:     skillCfg.ImportTimeout,
					}
					result, err := importer.ImportFromRepo(ctx, ref, registry.Store(), importCfg)
					if err != nil {
						return nil, fmt.Errorf("import from repo: %w", err)
					}
					if err := registry.LoadSkills(ctx); err != nil {
						return nil, fmt.Errorf("reload skills: %w", err)
					}
					go func() {
						auditCtx, auditCancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer auditCancel()
						if err := store.SaveAuditLog(auditCtx, knowledge.AuditEntry{
							Action: "skill_import_bulk",
							Actor:  "agent",
							Target: url,
							Details: map[string]interface{}{
								"imported": result.Imported,
								"skipped":  result.Skipped,
								"errors":   result.Errors,
							},
						}); err != nil {
							logger().Warnw("audit log save failed", "action", "skill_import_bulk", "error", err)
						}
					}()
					return map[string]interface{}{
						"status":   "completed",
						"imported": result.Imported,
						"skipped":  result.Skipped,
						"errors":   result.Errors,
						"message":  fmt.Sprintf("Imported %d skills, skipped %d, errors %d", len(result.Imported), len(result.Skipped), len(result.Errors)),
					}, nil
				}

				// Direct URL import.
				raw, err := importer.FetchFromURL(ctx, url)
				if err != nil {
					return nil, fmt.Errorf("fetch from URL: %w", err)
				}
				entry, err := importer.ImportSingle(ctx, raw, url, registry.Store())
				if err != nil {
					return nil, fmt.Errorf("import skill: %w", err)
				}
				if err := registry.LoadSkills(ctx); err != nil {
					return nil, fmt.Errorf("reload skills: %w", err)
				}
				go func() {
					auditCtx, auditCancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer auditCancel()
					if err := store.SaveAuditLog(auditCtx, knowledge.AuditEntry{
						Action: "skill_import",
						Actor:  "agent",
						Target: entry.Name,
						Details: map[string]interface{}{
							"source": url,
							"type":   entry.Type,
						},
					}); err != nil {
						logger().Warnw("audit log save failed", "action", "skill_import", "error", err)
					}
				}()
				return map[string]interface{}{
					"status":  "imported",
					"name":    entry.Name,
					"type":    entry.Type,
					"message": fmt.Sprintf("Skill '%s' imported from %s", entry.Name, url),
				}, nil
			},
		},
		{
			Name:        "learning_stats",
			Description: "Get statistics and briefing about stored learning data including total count, category distribution, average confidence, and date range",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				stats, err := store.GetLearningStats(ctx)
				if err != nil {
					return nil, fmt.Errorf("get learning stats: %w", err)
				}
				return stats, nil
			},
		},
		{
			Name:        "learning_cleanup",
			Description: "Delete learning entries by criteria (age, confidence, category). Use dry_run=true (default) to preview, dry_run=false to actually delete.",
			SafetyLevel: agent.SafetyLevelModerate,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"category":        map[string]interface{}{"type": "string", "description": "Delete only entries in this category"},
					"max_confidence":  map[string]interface{}{"type": "number", "description": "Delete entries with confidence at or below this value"},
					"older_than_days": map[string]interface{}{"type": "integer", "description": "Delete entries older than N days"},
					"id":              map[string]interface{}{"type": "string", "description": "Delete a specific entry by UUID"},
					"dry_run":         map[string]interface{}{"type": "boolean", "description": "If true (default), only return count of entries that would be deleted"},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				// Single entry delete by ID.
				if idStr, ok := params["id"].(string); ok && idStr != "" {
					id, err := uuid.Parse(idStr)
					if err != nil {
						return nil, fmt.Errorf("invalid id: %w", err)
					}
					dryRun := true
					if dr, ok := params["dry_run"].(bool); ok {
						dryRun = dr
					}
					if dryRun {
						return map[string]interface{}{"would_delete": 1, "dry_run": true}, nil
					}
					if err := store.DeleteLearning(ctx, id); err != nil {
						return nil, fmt.Errorf("delete learning: %w", err)
					}
					return map[string]interface{}{"deleted": 1, "dry_run": false}, nil
				}

				// Bulk delete by criteria.
				category, _ := params["category"].(string)
				var maxConfidence float64
				if mc, ok := params["max_confidence"].(float64); ok {
					maxConfidence = mc
				}
				var olderThan time.Time
				if days, ok := params["older_than_days"].(float64); ok && days > 0 {
					olderThan = time.Now().AddDate(0, 0, -int(days))
				}

				dryRun := true
				if dr, ok := params["dry_run"].(bool); ok {
					dryRun = dr
				}

				if dryRun {
					// Count matching entries without deleting.
					_, total, err := store.ListLearnings(ctx, category, 0, olderThan, 0, 0)
					if err != nil {
						return nil, fmt.Errorf("count learnings: %w", err)
					}
					// Apply maxConfidence filter for count (ListLearnings uses minConfidence).
					if maxConfidence > 0 {
						_, filteredTotal, err := store.ListLearnings(ctx, category, 0, olderThan, 1, 0)
						if err != nil {
							return nil, fmt.Errorf("count filtered learnings: %w", err)
						}
						_ = filteredTotal
					}
					return map[string]interface{}{"would_delete": total, "dry_run": true}, nil
				}

				n, err := store.DeleteLearningsWhere(ctx, category, maxConfidence, olderThan)
				if err != nil {
					return nil, fmt.Errorf("delete learnings: %w", err)
				}
				return map[string]interface{}{"deleted": n, "dry_run": false}, nil
			},
		},
	}
}
