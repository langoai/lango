package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/exportability"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/learning"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/skill"
	"github.com/langoai/lango/internal/toolparam"
)

var (
	metaReceiptsStoreMu sync.Mutex
	metaReceiptsStore   *receipts.Store
	metaReceiptsFactory = receipts.NewStore
)

// buildMetaTools creates knowledge/learning/skill meta-tools for the agent.
// cfg is used for session-mode-aware skill filtering and view_skill path resolution;
// it may be nil for tests that don't exercise mode features.
func buildMetaTools(store *knowledge.Store, engine *learning.Engine, registry *skill.Registry, skillCfg config.SkillConfig, cfg *config.Config) []*agent.Tool {
	exportabilityEnabled := exportabilityPolicyEnabled(cfg)

	tools := []*agent.Tool{
		{
			Name:        "save_knowledge",
			Description: "Save knowledge (appends new version if content changes, skips duplicates). Categories: rule, definition, preference, fact, pattern, correction. Temporal tags (evergreen/current_state) are auto-assigned by analyzers",
			SafetyLevel: agent.SafetyLevelModerate,
			Capability: agent.ToolCapability{
				Category: "knowledge",
				Activity: agent.ActivityWrite,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key":          map[string]interface{}{"type": "string", "description": "Unique key for this knowledge entry"},
					"category":     map[string]interface{}{"type": "string", "description": "Category: rule, definition, preference, fact, pattern, or correction", "enum": []string{"rule", "definition", "preference", "fact", "pattern", "correction"}},
					"content":      map[string]interface{}{"type": "string", "description": "The knowledge content to save"},
					"tags":         map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Optional tags for categorization"},
					"source":       map[string]interface{}{"type": "string", "description": "Where this knowledge came from"},
					"source_class": map[string]interface{}{"type": "string", "description": "Exportability source class: public, user-exportable, or private-confidential", "enum": []string{"public", "user-exportable", "private-confidential"}},
					"asset_label":  map[string]interface{}{"type": "string", "description": "Asset label used for exportability evaluation"},
				},
				"required": []string{"key", "category", "content"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				key, err := toolparam.RequireString(params, "key")
				if err != nil {
					return nil, err
				}
				category, err := toolparam.RequireString(params, "category")
				if err != nil {
					return nil, err
				}
				content, err := toolparam.RequireString(params, "content")
				if err != nil {
					return nil, err
				}
				source := toolparam.OptionalString(params, "source", "knowledge")
				sourceClass, err := exportability.ParseSourceClass(toolparam.OptionalString(params, "source_class", string(exportability.DefaultSourceClass)))
				if err != nil {
					return nil, err
				}
				assetLabel := toolparam.OptionalString(params, "asset_label", key)

				cat := entknowledge.Category(category)
				if err := entknowledge.CategoryValidator(cat); err != nil {
					return nil, fmt.Errorf("invalid category %q: %w", category, err)
				}

				tags := toolparam.StringSlice(params, "tags")

				entry := knowledge.KnowledgeEntry{
					Key:         key,
					Category:    cat,
					Content:     content,
					Tags:        tags,
					Source:      source,
					SourceClass: string(sourceClass),
					AssetLabel:  assetLabel,
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

				// Read back to get the version number.
				saved, _ := store.GetKnowledge(ctx, key)
				version := 0
				if saved != nil {
					version = saved.Version
				}

				return map[string]interface{}{
					"status":  "saved",
					"key":     key,
					"version": version,
					"message": fmt.Sprintf("Knowledge '%s' saved (version %d)", key, version),
				}, nil
			},
		},
		{
			Name:        "evaluate_exportability",
			Description: "Evaluate exportability for a knowledge-backed artifact label from the latest source entries and emit a durable decision receipt",
			SafetyLevel: agent.SafetyLevelModerate,
			Capability: agent.ToolCapability{
				Category: "knowledge",
				Activity: agent.ActivityWrite,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"artifact_label": map[string]interface{}{"type": "string", "description": "Label for the artifact being evaluated"},
					"source_keys": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Knowledge keys that provide source lineage for the artifact",
					},
					"stage": map[string]interface{}{
						"type":        "string",
						"description": "Decision stage: draft or final",
						"enum":        []string{"draft", "final"},
					},
				},
				"required": []string{"artifact_label", "source_keys", "stage"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				if store == nil {
					return nil, fmt.Errorf("knowledge store is not available")
				}

				artifactLabel, err := toolparam.RequireString(params, "artifact_label")
				if err != nil {
					return nil, err
				}
				stageStr, err := toolparam.RequireString(params, "stage")
				if err != nil {
					return nil, err
				}
				var stage exportability.DecisionStage
				switch stageStr {
				case string(exportability.StageDraft):
					stage = exportability.StageDraft
				case string(exportability.StageFinal):
					stage = exportability.StageFinal
				default:
					return nil, fmt.Errorf("invalid stage %q: must be draft or final", stageStr)
				}

				sourceKeys := toolparam.StringSlice(params, "source_keys")
				if len(sourceKeys) == 0 {
					return nil, fmt.Errorf("source_keys must not be empty")
				}
				for _, key := range sourceKeys {
					if strings.TrimSpace(key) == "" {
						return nil, fmt.Errorf("source_keys must not contain empty values")
					}
				}

				return evaluateExportabilityArtifact(ctx, store, artifactLabel, sourceKeys, stage, exportabilityEnabled)
			},
		},
		{
			Name:        "approve_artifact_release",
			Description: "Evaluate an artifact release request and emit an audit-backed approval receipt",
			SafetyLevel: agent.SafetyLevelModerate,
			Capability: agent.ToolCapability{
				Category: "knowledge",
				Activity: agent.ActivityWrite,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"artifact_label":  map[string]interface{}{"type": "string", "description": "Label for the artifact being released"},
					"requested_scope": map[string]interface{}{"type": "string", "description": "Requested release scope or target label"},
					"exportability_state": map[string]interface{}{
						"type":        "string",
						"description": "Exportability receipt state: exportable, blocked, or needs-human-review",
						"enum":        []string{"exportable", "blocked", "needs-human-review"},
					},
					"override_requested": map[string]interface{}{"type": "boolean", "description": "Whether a blocked release override was requested"},
					"high_risk":          map[string]interface{}{"type": "boolean", "description": "Whether the release is high risk"},
				},
				"required": []string{"artifact_label", "requested_scope", "exportability_state"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				if store == nil {
					return nil, fmt.Errorf("knowledge store is not available")
				}

				artifactLabel, err := toolparam.RequireString(params, "artifact_label")
				if err != nil {
					return nil, err
				}
				requestedScope, err := toolparam.RequireString(params, "requested_scope")
				if err != nil {
					return nil, err
				}
				exportabilityStateStr, err := toolparam.RequireString(params, "exportability_state")
				if err != nil {
					return nil, err
				}

				overrideRequested := toolparam.OptionalBool(params, "override_requested", false)
				highRisk := toolparam.OptionalBool(params, "high_risk", false)

				var state exportability.DecisionState
				switch exportabilityStateStr {
				case string(exportability.StateExportable):
					state = exportability.StateExportable
				case string(exportability.StateBlocked):
					state = exportability.StateBlocked
				case string(exportability.StateNeedsHumanReview):
					state = exportability.StateNeedsHumanReview
				default:
					return nil, fmt.Errorf("invalid exportability_state %q", exportabilityStateStr)
				}

				outcome := evaluateArtifactReleaseApproval(artifactLabel, requestedScope, state, overrideRequested, highRisk)
				payload := newArtifactReleaseApprovalReceipt(artifactLabel, requestedScope, state, outcome, overrideRequested, highRisk)
				if err := store.SaveAuditLog(ctx, knowledge.AuditEntry{
					Action:  "artifact_release_approval",
					Actor:   "agent",
					Target:  "artifact:" + artifactLabel,
					Details: payload.Details(),
				}); err != nil {
					return nil, fmt.Errorf("save artifact release approval audit log: %w", err)
				}

				return payload, nil
			},
		},
		newDisputeReadyReceiptTool(),
		{
			Name:        "get_knowledge_history",
			Description: "Get version history for a knowledge entry. Returns all versions ordered newest first",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "knowledge",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key": map[string]interface{}{"type": "string", "description": "Knowledge entry key"},
				},
				"required": []string{"key"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				key, err := toolparam.RequireString(params, "key")
				if err != nil {
					return nil, err
				}

				history, err := store.GetKnowledgeHistory(ctx, key)
				if err != nil {
					return nil, fmt.Errorf("get knowledge history: %w", err)
				}

				versions := make([]map[string]interface{}, 0, len(history))
				for _, h := range history {
					versions = append(versions, map[string]interface{}{
						"version":    h.Version,
						"category":   string(h.Category),
						"content":    h.Content,
						"created_at": h.CreatedAt.Format(time.RFC3339),
					})
				}

				return map[string]interface{}{
					"key":      key,
					"versions": versions,
				}, nil
			},
		},
		{
			Name:        "search_knowledge",
			Description: "Search stored knowledge entries by query and optional category",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "knowledge",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query":    map[string]interface{}{"type": "string", "description": "Search query"},
					"category": map[string]interface{}{"type": "string", "description": "Optional category filter: rule, definition, preference, or fact"},
				},
				"required": []string{"query"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				query := toolparam.OptionalString(params, "query", "")
				category := toolparam.OptionalString(params, "category", "")

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
			Capability: agent.ToolCapability{
				Category: "knowledge",
				Activity: agent.ActivityWrite,
			},
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
				trigger, err := toolparam.RequireString(params, "trigger")
				if err != nil {
					return nil, err
				}
				fix, err := toolparam.RequireString(params, "fix")
				if err != nil {
					return nil, err
				}
				errorPattern := toolparam.OptionalString(params, "error_pattern", "")
				diagnosis := toolparam.OptionalString(params, "diagnosis", "")
				category := toolparam.OptionalString(params, "category", "general")

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
			Capability: agent.ToolCapability{
				Category:        "knowledge",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query":    map[string]interface{}{"type": "string", "description": "Search query (error message or trigger)"},
					"category": map[string]interface{}{"type": "string", "description": "Optional category filter"},
				},
				"required": []string{"query"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				query := toolparam.OptionalString(params, "query", "")
				category := toolparam.OptionalString(params, "category", "")

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
			Capability: agent.ToolCapability{
				Category: "skill",
				Activity: agent.ActivityManage,
			},
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
				name, err := toolparam.RequireString(params, "name")
				if err != nil {
					return nil, err
				}
				description, err := toolparam.RequireString(params, "description")
				if err != nil {
					return nil, err
				}
				skillType, err := toolparam.RequireString(params, "type")
				if err != nil {
					return nil, err
				}
				definitionStr, err := toolparam.RequireString(params, "definition")
				if err != nil {
					return nil, err
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
			Description: "List all active skills. Pass summary=true for token-efficient metadata only.",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "skill",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"summary": map[string]interface{}{
						"type":        "boolean",
						"description": "When true, return only {name, description, when_to_use} per skill.",
					},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				if registry == nil {
					return map[string]interface{}{"skills": []interface{}{}, "count": 0}, nil
				}

				skills, err := registry.ListActiveSkills(ctx)
				if err != nil {
					return nil, fmt.Errorf("list skills: %w", err)
				}

				// Apply session mode filter if active.
				modeName := session.ModeNameFromContext(ctx)
				if modeName != "" && cfg != nil {
					if mode, ok := cfg.LookupMode(modeName); ok && len(mode.Skills) > 0 {
						allow := make(map[string]bool, len(mode.Skills))
						for _, name := range mode.Skills {
							allow[name] = true
						}
						filtered := skills[:0]
						for _, s := range skills {
							if allow[s.Name] {
								filtered = append(filtered, s)
							}
						}
						skills = filtered
					}
				}

				summary, _ := params["summary"].(bool)
				if summary {
					out := make([]map[string]interface{}, 0, len(skills))
					for _, s := range skills {
						out = append(out, map[string]interface{}{
							"name":        s.Name,
							"description": s.Description,
							"when_to_use": s.WhenToUse,
						})
					}
					return map[string]interface{}{
						"skills": out,
						"count":  len(out),
					}, nil
				}

				return map[string]interface{}{
					"skills": skills,
					"count":  len(skills),
				}, nil
			},
		},
		{
			Name: "view_skill",
			Description: "Read the full content of an active skill. " +
				"With name only, returns the skill's SKILL.md. With name+path, returns the referenced supporting file " +
				"resolved relative to the skill directory.",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "skill",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "The active skill name to read.",
					},
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Optional supporting file path relative to the skill directory.",
					},
				},
				"required": []string{"name"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				if registry == nil {
					return nil, fmt.Errorf("skill registry not available")
				}
				name, _ := params["name"].(string)
				if strings.TrimSpace(name) == "" {
					return nil, fmt.Errorf("name is required")
				}
				path, _ := params["path"].(string)

				skills, err := registry.ListActiveSkills(ctx)
				if err != nil {
					return nil, fmt.Errorf("list skills: %w", err)
				}
				var found *skill.SkillEntry
				for i := range skills {
					if skills[i].Name == name {
						found = &skills[i]
						break
					}
				}
				if found == nil {
					return nil, fmt.Errorf("skill %q is not active", name)
				}

				skillsDir := skillCfg.SkillsDir
				if skillsDir == "" {
					return map[string]interface{}{
						"name":    name,
						"content": found.Context,
						"note":    "skills directory not configured; returning skill Context only",
					}, nil
				}

				var skillRoot string
				if found.SourcePack != "" {
					skillRoot = filepath.Join(skillsDir, "ext-"+found.SourcePack, name)
				} else {
					skillRoot = filepath.Join(skillsDir, name)
				}
				var target string
				if path == "" {
					target = filepath.Join(skillRoot, "SKILL.md")
				} else {
					// Resolve and verify the path stays inside the skill directory.
					cleaned := filepath.Clean(filepath.Join(skillRoot, path))
					absRoot, err := filepath.Abs(skillRoot)
					if err != nil {
						return nil, fmt.Errorf("resolve skill root: %w", err)
					}
					absTarget, err := filepath.Abs(cleaned)
					if err != nil {
						return nil, fmt.Errorf("resolve target: %w", err)
					}
					if !strings.HasPrefix(absTarget, absRoot+string(filepath.Separator)) && absTarget != absRoot {
						return nil, fmt.Errorf("path %q is outside the skill directory", path)
					}
					target = cleaned
				}

				content, err := os.ReadFile(target)
				if err != nil {
					return nil, fmt.Errorf("read %s: %w", filepath.Base(target), err)
				}
				return map[string]interface{}{
					"name":    name,
					"path":    target,
					"content": string(content),
				}, nil
			},
		},
		{
			Name: "import_skill",
			Description: "Import skills from a GitHub repository or URL. " +
				"Supports bulk import (all skills from a repo) or single skill import.",
			SafetyLevel: agent.SafetyLevelModerate,
			Capability: agent.ToolCapability{
				Category: "skill",
				Activity: agent.ActivityManage,
			},
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

				url, err := toolparam.RequireString(params, "url")
				if err != nil {
					return nil, err
				}
				skillName := toolparam.OptionalString(params, "skill_name", "")

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
			Capability: agent.ToolCapability{
				Category:        "knowledge",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
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
			Capability: agent.ToolCapability{
				Category: "knowledge",
				Activity: agent.ActivityManage,
			},
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

	return tools
}

func exportabilityPolicyEnabled(cfg *config.Config) bool {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	return cfg.Security.Exportability.Enabled
}

func evaluateExportabilityArtifact(ctx context.Context, store *knowledge.Store, artifactLabel string, sourceKeys []string, stage exportability.DecisionStage, enabled bool) (map[string]interface{}, error) {
	entries, err := store.GetKnowledgeByKeys(ctx, sourceKeys)
	if err != nil {
		return nil, fmt.Errorf("load source knowledge: %w", err)
	}

	refs, err := exportabilitySourceRefs(entries)
	if err != nil {
		return nil, err
	}

	receipt := exportability.Evaluate(exportability.Policy{Enabled: enabled}, stage, refs)
	payload := exportabilityReceiptPayload(artifactLabel, receipt)

	if err := store.SaveAuditLog(ctx, knowledge.AuditEntry{
		Action:  "exportability_decision",
		Actor:   "agent",
		Target:  "artifact:" + artifactLabel,
		Details: payload,
	}); err != nil {
		return nil, fmt.Errorf("save exportability decision audit log: %w", err)
	}

	return payload, nil
}

func exportabilitySourceRefs(entries []knowledge.KnowledgeEntry) ([]exportability.SourceRef, error) {
	refs := make([]exportability.SourceRef, 0, len(entries))
	for _, entry := range entries {
		class, err := exportability.ParseSourceClass(entry.SourceClass)
		if err != nil {
			return nil, fmt.Errorf("parse source class for %q: %w", entry.Key, err)
		}
		label := entry.AssetLabel
		if label == "" {
			label = entry.Key
		}
		refs = append(refs, exportability.SourceRef{
			AssetID:    entry.Key,
			AssetLabel: label,
			Class:      class,
		})
	}
	return refs, nil
}

func exportabilityReceiptPayload(artifactLabel string, receipt exportability.Receipt) map[string]interface{} {
	lineage := make([]map[string]interface{}, 0, len(receipt.Lineage))
	for _, row := range receipt.Lineage {
		lineage = append(lineage, map[string]interface{}{
			"asset_id":    row.AssetID,
			"asset_label": row.AssetLabel,
			"class":       string(row.Class),
			"rule":        row.Rule,
		})
	}

	return map[string]interface{}{
		"artifact_label": artifactLabel,
		"stage":          string(receipt.Stage),
		"state":          string(receipt.State),
		"policy_code":    receipt.PolicyCode,
		"explanation":    receipt.Explanation,
		"lineage":        lineage,
	}
}

func newDisputeReadyReceiptTool() *agent.Tool {
	return &agent.Tool{
		Name:        "create_dispute_ready_receipt",
		Description: "Create a lite dispute-ready submission receipt linked to a transaction receipt",
		SafetyLevel: agent.SafetyLevelModerate,
		Capability: agent.ToolCapability{
			Category: "knowledge",
			Activity: agent.ActivityWrite,
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"transaction_id":        map[string]interface{}{"type": "string", "description": "External transaction identifier"},
				"artifact_label":        map[string]interface{}{"type": "string", "description": "Label for the artifact being submitted"},
				"payload_hash":          map[string]interface{}{"type": "string", "description": "Payload content hash"},
				"source_lineage_digest": map[string]interface{}{"type": "string", "description": "Digest of the source lineage for the submission"},
			},
			"required": []string{"transaction_id", "artifact_label", "payload_hash", "source_lineage_digest"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			store := metaReceiptsStoreInstance()
			if store == nil {
				return nil, fmt.Errorf("receipts store dependency is not configured")
			}

			transactionID, err := toolparam.RequireString(params, "transaction_id")
			if err != nil {
				return nil, err
			}
			artifactLabel, err := toolparam.RequireString(params, "artifact_label")
			if err != nil {
				return nil, err
			}
			payloadHash, err := toolparam.RequireString(params, "payload_hash")
			if err != nil {
				return nil, err
			}
			sourceLineageDigest, err := toolparam.RequireString(params, "source_lineage_digest")
			if err != nil {
				return nil, err
			}

			submission, transaction, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
				TransactionID:       transactionID,
				ArtifactLabel:       artifactLabel,
				PayloadHash:         payloadHash,
				SourceLineageDigest: sourceLineageDigest,
			})
			if err != nil {
				return nil, fmt.Errorf("create dispute-ready receipt: %w", err)
			}

			return map[string]interface{}{
				"submission_receipt_id":         submission.SubmissionReceiptID,
				"transaction_receipt_id":        transaction.TransactionReceiptID,
				"current_submission_receipt_id": transaction.CurrentSubmissionReceiptID,
			}, nil
		},
	}
}

func metaReceiptsStoreInstance() *receipts.Store {
	metaReceiptsStoreMu.Lock()
	defer metaReceiptsStoreMu.Unlock()

	if metaReceiptsStore != nil {
		return metaReceiptsStore
	}
	if metaReceiptsFactory == nil {
		return nil
	}
	metaReceiptsStore = metaReceiptsFactory()
	return metaReceiptsStore
}
