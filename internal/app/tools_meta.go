package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/approvalflow"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/disputehold"
	"github.com/langoai/lango/internal/economy/escrow"
	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/escrowexecution"
	"github.com/langoai/lango/internal/escrowrefund"
	"github.com/langoai/lango/internal/escrowrelease"
	"github.com/langoai/lango/internal/exportability"
	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/knowledgeruntime"
	"github.com/langoai/lango/internal/learning"
	"github.com/langoai/lango/internal/partialsettlementexecution"
	"github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/settlementexecution"
	"github.com/langoai/lango/internal/settlementprogression"
	"github.com/langoai/lango/internal/skill"
	"github.com/langoai/lango/internal/toolparam"
)

// buildMetaTools creates knowledge/learning/skill meta-tools for the agent.
// cfg is used for session-mode-aware skill filtering and view_skill path resolution;
// it may be nil for tests that don't exercise mode features.
func buildMetaTools(store *knowledge.Store, engine *learning.Engine, registry *skill.Registry, skillCfg config.SkillConfig, cfg *config.Config, receiptStore *receipts.Store) []*agent.Tool {
	return buildMetaToolsWithRuntimes(store, engine, registry, skillCfg, cfg, receiptStore, nil, nil, nil, nil, nil, nil)
}

func buildMetaToolsWithEscrow(
	store *knowledge.Store,
	engine *learning.Engine,
	registry *skill.Registry,
	skillCfg config.SkillConfig,
	cfg *config.Config,
	receiptStore *receipts.Store,
	escrowRuntime escrowExecutionRuntime,
) []*agent.Tool {
	var escrowDisputeHoldRuntime escrowDisputeHoldExecutionRuntime
	var escrowReleaseRuntime escrowReleaseExecutionRuntime
	var escrowRefundRuntime escrowRefundExecutionRuntime
	if engine, ok := escrowRuntime.(*escrow.Engine); ok && engine != nil {
		escrowDisputeHoldRuntime = engineEscrowDisputeHoldRuntime{engine: engine}
		escrowReleaseRuntime = engineEscrowReleaseRuntime{engine: engine}
		escrowRefundRuntime = engineEscrowRefundRuntime{engine: engine}
	}

	return buildMetaToolsWithRuntimes(store, engine, registry, skillCfg, cfg, receiptStore, escrowRuntime, nil, nil, escrowDisputeHoldRuntime, escrowReleaseRuntime, escrowRefundRuntime)
}

func buildMetaToolsWithRuntimes(
	store *knowledge.Store,
	engine *learning.Engine,
	registry *skill.Registry,
	skillCfg config.SkillConfig,
	cfg *config.Config,
	receiptStore *receipts.Store,
	escrowRuntime escrowExecutionRuntime,
	settlementRuntime settlementExecutionRuntime,
	partialSettlementRuntime partialSettlementExecutionRuntime,
	escrowDisputeHoldRuntime escrowDisputeHoldExecutionRuntime,
	escrowReleaseRuntime escrowReleaseExecutionRuntime,
	escrowRefundRuntime escrowRefundExecutionRuntime,
) []*agent.Tool {
	exportabilityEnabled := exportabilityPolicyEnabled(cfg)
	effectivePartialRuntime := partialSettlementRuntime
	if effectivePartialRuntime == nil && settlementRuntime != nil {
		effectivePartialRuntime = settlementPartialSettlementRuntime{runtime: settlementRuntime}
	}

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
		newDisputeReadyReceiptTool(receiptStore),
		newOpenKnowledgeExchangeTransactionTool(receiptStore),
		newSelectKnowledgeExchangePathTool(receiptStore),
		{
			Name:        "approve_upfront_payment",
			Description: "Evaluate an upfront payment request and update the linked transaction receipt with canonical payment approval state",
			SafetyLevel: agent.SafetyLevelModerate,
			Capability: agent.ToolCapability{
				Category: "knowledge",
				Activity: agent.ActivityWrite,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"transaction_receipt_id": map[string]interface{}{"type": "string", "description": "Linked transaction receipt identifier"},
					"submission_receipt_id":  map[string]interface{}{"type": "string", "description": "Explicit submission receipt identifier for this approval"},
					"amount":                 map[string]interface{}{"type": "string", "description": "Requested upfront payment amount"},
					"trust_score":            map[string]interface{}{"type": "number", "description": "Trust score used for upfront payment policy evaluation"},
					"user_max_prepay":        map[string]interface{}{"type": "string", "description": "Maximum prepay amount allowed by policy"},
					"remaining_budget":       map[string]interface{}{"type": "string", "description": "Remaining budget available for the transaction"},
					"escrow_buyer_did":       map[string]interface{}{"type": "string", "description": "Buyer DID for escrow-backed execution"},
					"escrow_seller_did":      map[string]interface{}{"type": "string", "description": "Seller DID for escrow-backed execution"},
					"escrow_reason":          map[string]interface{}{"type": "string", "description": "Reason to store for escrow-backed execution"},
					"escrow_task_id":         map[string]interface{}{"type": "string", "description": "Optional task identifier for escrow-backed execution"},
					"escrow_milestones": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"description": map[string]interface{}{"type": "string", "description": "Milestone description"},
								"amount":      map[string]interface{}{"type": "string", "description": "Milestone amount"},
							},
						},
						"description": "Optional escrow milestone breakdown",
					},
				},
				"required": []string{"transaction_receipt_id", "submission_receipt_id", "amount", "trust_score", "user_max_prepay", "remaining_budget"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				return approveUpfrontPayment(ctx, receiptStore, params, paymentapproval.EvaluateUpfrontPayment)
			},
		},
		newApplySettlementProgressionTool(receiptStore),
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

	if settlementTool := newExecuteSettlementTool(receiptStore, settlementRuntime); settlementTool != nil {
		tools = append(tools, settlementTool)
	}
	if partialSettlementTool := newExecutePartialSettlementTool(receiptStore, effectivePartialRuntime); partialSettlementTool != nil {
		tools = append(tools, partialSettlementTool)
	}
	if disputeHoldTool := newHoldEscrowForDisputeTool(receiptStore, escrowDisputeHoldRuntime); disputeHoldTool != nil {
		tools = append(tools, disputeHoldTool)
	}
	if escrowReleaseTool := newReleaseEscrowSettlementTool(receiptStore, escrowReleaseRuntime); escrowReleaseTool != nil {
		tools = append(tools, escrowReleaseTool)
	}
	if escrowRefundTool := newRefundEscrowSettlementTool(receiptStore, escrowRefundRuntime); escrowRefundTool != nil {
		tools = append(tools, escrowRefundTool)
	}
	if escrowTool := newExecuteEscrowRecommendationTool(receiptStore, escrowRuntime); escrowTool != nil {
		tools = append(tools, escrowTool)
	}

	return tools
}

type escrowExecutionRuntime interface {
	Create(context.Context, escrow.CreateRequest) (*escrow.EscrowEntry, error)
	Fund(context.Context, string) (*escrow.EscrowEntry, error)
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

func approveUpfrontPayment(ctx context.Context, receiptStore *receipts.Store, params map[string]interface{}, evaluate func(paymentapproval.Input) paymentapproval.Outcome) (interface{}, error) {
	if receiptStore == nil {
		return nil, fmt.Errorf("receipts store dependency is not configured")
	}

	transactionReceiptID, err := toolparam.RequireString(params, "transaction_receipt_id")
	if err != nil {
		return nil, err
	}
	submissionReceiptID, err := toolparam.RequireString(params, "submission_receipt_id")
	if err != nil {
		return nil, err
	}
	amount, err := toolparam.RequireString(params, "amount")
	if err != nil {
		return nil, err
	}
	trustScore, err := toolparam.RequireFloat64(params, "trust_score")
	if err != nil {
		return nil, err
	}
	userMaxPrepay, err := toolparam.RequireString(params, "user_max_prepay")
	if err != nil {
		return nil, err
	}
	remainingBudget, err := toolparam.RequireString(params, "remaining_budget")
	if err != nil {
		return nil, err
	}

	outcome := evaluate(paymentapproval.Input{
		Amount: amount,
		Trust: paymentapproval.TrustInput{
			Score: trustScore,
		},
		Budget: paymentapproval.BudgetPolicyContext{
			UserMaxPrepay:   userMaxPrepay,
			RemainingBudget: remainingBudget,
		},
	})

	var escrowInput receipts.EscrowExecutionInput
	if outcome.SuggestedMode == paymentapproval.ModeEscrow {
		escrowInput, err = parseEscrowExecutionInput(params, amount)
		if err != nil {
			return nil, err
		}
		transaction, err := receiptStore.GetTransactionReceipt(ctx, transactionReceiptID)
		if err != nil {
			return nil, fmt.Errorf("load transaction receipt for escrow approval: %w", err)
		}
		if transaction.CurrentSubmissionReceiptID != submissionReceiptID {
			return nil, fmt.Errorf(
				"submission receipt %q is not current for transaction receipt %q",
				submissionReceiptID,
				transactionReceiptID,
			)
		}
	}

	updatedTx, err := receiptStore.ApplyUpfrontPaymentApproval(ctx, transactionReceiptID, submissionReceiptID, outcome)
	if err != nil {
		return nil, fmt.Errorf("apply upfront payment approval: %w", err)
	}
	if outcome.SuggestedMode == paymentapproval.ModeEscrow {
		updatedTx, err = receiptStore.BindEscrowExecutionInput(ctx, transactionReceiptID, submissionReceiptID, escrowInput)
		if err != nil {
			return nil, fmt.Errorf("bind escrow execution input: %w", err)
		}
	}

	return newUpfrontPaymentApprovalReceipt(
		transactionReceiptID,
		submissionReceiptID,
		amount,
		trustScore,
		userMaxPrepay,
		remainingBudget,
		outcome,
		updatedTx,
	), nil
}

func parseEscrowExecutionInput(params map[string]interface{}, amount string) (receipts.EscrowExecutionInput, error) {
	buyerDID, err := toolparam.RequireString(params, "escrow_buyer_did")
	if err != nil {
		return receipts.EscrowExecutionInput{}, err
	}
	sellerDID, err := toolparam.RequireString(params, "escrow_seller_did")
	if err != nil {
		return receipts.EscrowExecutionInput{}, err
	}
	reason, err := toolparam.RequireString(params, "escrow_reason")
	if err != nil {
		return receipts.EscrowExecutionInput{}, err
	}
	taskID := toolparam.OptionalString(params, "escrow_task_id", "")

	var milestones []receipts.EscrowMilestoneInput
	if raw, ok := params["escrow_milestones"]; ok {
		rawMilestones, ok := raw.([]interface{})
		if !ok {
			return receipts.EscrowExecutionInput{}, fmt.Errorf("escrow_milestones must be an array")
		}
		milestones = make([]receipts.EscrowMilestoneInput, 0, len(rawMilestones))
		for i, rawMilestone := range rawMilestones {
			milestoneMap, ok := rawMilestone.(map[string]interface{})
			if !ok {
				return receipts.EscrowExecutionInput{}, fmt.Errorf("escrow_milestones[%d] must be an object", i)
			}
			description, err := toolparam.RequireString(milestoneMap, "description")
			if err != nil {
				return receipts.EscrowExecutionInput{}, fmt.Errorf("escrow_milestones[%d]: %w", i, err)
			}
			milestoneAmount, err := toolparam.RequireString(milestoneMap, "amount")
			if err != nil {
				return receipts.EscrowExecutionInput{}, fmt.Errorf("escrow_milestones[%d]: %w", i, err)
			}
			milestones = append(milestones, receipts.EscrowMilestoneInput{
				Description: description,
				Amount:      milestoneAmount,
			})
		}
	}

	return receipts.EscrowExecutionInput{
		BuyerDID:   buyerDID,
		SellerDID:  sellerDID,
		Amount:     amount,
		Reason:     reason,
		TaskID:     taskID,
		Milestones: milestones,
	}, nil
}

func newExecuteEscrowRecommendationTool(receiptStore *receipts.Store, runtime escrowExecutionRuntime) *agent.Tool {
	if runtime == nil {
		return nil
	}

	return &agent.Tool{
		Name:        "execute_escrow_recommendation",
		Description: "Execute a previously approved escrow recommendation for a transaction receipt and persist canonical escrow evidence",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category: "knowledge",
			Activity: agent.ActivityWrite,
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"transaction_receipt_id": map[string]interface{}{"type": "string", "description": "Approved transaction receipt identifier with bound escrow execution input"},
			},
			"required": []string{"transaction_receipt_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if receiptStore == nil {
				return nil, fmt.Errorf("receipts store dependency is not configured")
			}

			transactionReceiptID, err := toolparam.RequireString(params, "transaction_receipt_id")
			if err != nil {
				return nil, err
			}

			result, err := escrowexecution.NewService(receiptStore, runtime).ExecuteRecommendation(ctx, escrowexecution.Request{
				TransactionReceiptID: transactionReceiptID,
			})
			if err != nil {
				return nil, err
			}

			return executeEscrowRecommendationReceipt{
				TransactionReceiptID:  result.TransactionReceiptID,
				SubmissionReceiptID:   result.SubmissionReceiptID,
				EscrowReference:       result.EscrowReference,
				EscrowExecutionStatus: string(result.EscrowExecutionStatus),
			}, nil
		},
	}
}

func newDisputeReadyReceiptTool(receiptStore *receipts.Store) *agent.Tool {
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
			if receiptStore == nil {
				return nil, fmt.Errorf("receipts store dependency is not configured")
			}

			input, err := parseDisputeReadyReceiptInput(params)
			if err != nil {
				return nil, err
			}

			return createDisputeReadyReceipt(ctx, receiptStore, input)
		},
	}
}

func newOpenKnowledgeExchangeTransactionTool(receiptStore *receipts.Store) *agent.Tool {
	return &agent.Tool{
		Name:        "open_knowledge_exchange_transaction",
		Description: "Open a receipts-backed knowledge exchange transaction and persist canonical runtime-open state",
		SafetyLevel: agent.SafetyLevelModerate,
		Capability: agent.ToolCapability{
			Category: "knowledge",
			Activity: agent.ActivityWrite,
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"transaction_id":  map[string]interface{}{"type": "string", "description": "External transaction identifier"},
				"counterparty":    map[string]interface{}{"type": "string", "description": "Counterparty DID or stable participant identifier"},
				"requested_scope": map[string]interface{}{"type": "string", "description": "Requested knowledge exchange scope"},
				"price_context":   map[string]interface{}{"type": "string", "description": "Price context to bind at open time"},
				"trust_context":   map[string]interface{}{"type": "string", "description": "Trust context to bind at open time"},
			},
			"required": []string{"transaction_id", "counterparty", "requested_scope", "price_context", "trust_context"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if receiptStore == nil {
				return nil, fmt.Errorf("receipts store dependency is not configured")
			}

			transactionID, err := toolparam.RequireString(params, "transaction_id")
			if err != nil {
				return nil, err
			}
			counterparty, err := toolparam.RequireString(params, "counterparty")
			if err != nil {
				return nil, err
			}
			requestedScope, err := toolparam.RequireString(params, "requested_scope")
			if err != nil {
				return nil, err
			}
			priceContext, err := toolparam.RequireString(params, "price_context")
			if err != nil {
				return nil, err
			}
			trustContext, err := toolparam.RequireString(params, "trust_context")
			if err != nil {
				return nil, err
			}

			result, err := knowledgeruntime.NewService(receiptStore).OpenTransaction(ctx, knowledgeruntime.OpenTransactionRequest{
				TransactionID:  transactionID,
				Counterparty:   counterparty,
				RequestedScope: requestedScope,
				PriceContext:   priceContext,
				TrustContext:   trustContext,
			})
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"transaction_id":                    transactionID,
				"transaction_receipt_id":            result.TransactionReceiptID,
				"knowledge_exchange_runtime_status": string(result.RuntimeStatus),
			}, nil
		},
	}
}

func newSelectKnowledgeExchangePathTool(receiptStore *receipts.Store) *agent.Tool {
	return &agent.Tool{
		Name:        "select_knowledge_exchange_path",
		Description: "Select the receipts-backed knowledge exchange execution path from canonical approval state",
		SafetyLevel: agent.SafetyLevelModerate,
		Capability: agent.ToolCapability{
			Category: "knowledge",
			Activity: agent.ActivityWrite,
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"transaction_receipt_id": map[string]interface{}{"type": "string", "description": "Transaction receipt identifier to evaluate"},
			},
			"required": []string{"transaction_receipt_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if receiptStore == nil {
				return nil, fmt.Errorf("receipts store dependency is not configured")
			}

			transactionReceiptID, err := toolparam.RequireString(params, "transaction_receipt_id")
			if err != nil {
				return nil, err
			}

			result, err := knowledgeruntime.NewService(receiptStore).SelectExecutionPath(ctx, transactionReceiptID)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"transaction_receipt_id": result.TransactionReceiptID,
				"branch":                 string(result.Branch),
			}, nil
		},
	}
}

func newApplySettlementProgressionTool(receiptStore *receipts.Store) *agent.Tool {
	return &agent.Tool{
		Name:        "apply_settlement_progression",
		Description: "Apply a release approval decision to the linked transaction receipt and return canonical settlement progression state",
		SafetyLevel: agent.SafetyLevelModerate,
		Capability: agent.ToolCapability{
			Category: "knowledge",
			Activity: agent.ActivityWrite,
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"transaction_receipt_id": map[string]interface{}{"type": "string", "description": "Transaction receipt identifier to update"},
				"outcome": map[string]interface{}{
					"type":        "string",
					"description": "Release approval decision",
					"enum":        []string{"approve", "reject", "request-revision", "escalate"},
				},
				"reason":       map[string]interface{}{"type": "string", "description": "Optional human-readable reason for the progression update"},
				"partial_hint": map[string]interface{}{"type": "string", "description": "Optional partial settlement hint"},
			},
			"required": []string{"transaction_receipt_id", "outcome"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if receiptStore == nil {
				return nil, fmt.Errorf("receipts store dependency is not configured")
			}

			transactionReceiptID, err := toolparam.RequireString(params, "transaction_receipt_id")
			if err != nil {
				return nil, err
			}
			transactionReceiptID = strings.TrimSpace(transactionReceiptID)
			if transactionReceiptID == "" {
				return nil, &toolparam.ErrMissingParam{Name: "transaction_receipt_id"}
			}

			outcome, err := toolparam.RequireString(params, "outcome")
			if err != nil {
				return nil, err
			}
			outcome = strings.TrimSpace(outcome)
			if outcome == "" {
				return nil, &toolparam.ErrMissingParam{Name: "outcome"}
			}

			result, err := settlementprogression.NewService(receiptStore).ApplyReleaseOutcome(ctx, settlementprogression.ApplyReleaseOutcomeRequest{
				TransactionReceiptID: transactionReceiptID,
				Outcome: settlementprogression.ReleaseOutcome{
					Decision: approvalflow.Decision(outcome),
					Reason:   toolparam.OptionalString(params, "reason", ""),
				},
				PartialHint: toolparam.OptionalString(params, "partial_hint", ""),
			})
			if err != nil {
				return nil, err
			}

			return newApplySettlementProgressionReceipt(result), nil
		},
	}
}

type settlementExecutionRuntime interface {
	ExecuteSettlement(context.Context, settlementexecution.DirectPaymentRequest) (settlementexecution.DirectPaymentResult, error)
}

type paymentSettlementRuntime struct {
	service *payment.Service
}

func (p paymentSettlementRuntime) ExecuteSettlement(ctx context.Context, req settlementexecution.DirectPaymentRequest) (settlementexecution.DirectPaymentResult, error) {
	receipt, err := p.service.Send(ctx, payment.PaymentRequest{
		To:      req.Counterparty,
		Amount:  req.Amount,
		Purpose: "final settlement",
	})
	if err != nil {
		return settlementexecution.DirectPaymentResult{}, err
	}
	return settlementexecution.DirectPaymentResult{Reference: receipt.TxHash}, nil
}

type partialSettlementExecutionRuntime interface {
	ExecuteSettlement(context.Context, partialsettlementexecution.DirectPaymentRequest) (partialsettlementexecution.DirectPaymentResult, error)
}

type paymentPartialSettlementRuntime struct {
	service *payment.Service
}

func (p paymentPartialSettlementRuntime) ExecuteSettlement(ctx context.Context, req partialsettlementexecution.DirectPaymentRequest) (partialsettlementexecution.DirectPaymentResult, error) {
	receipt, err := p.service.Send(ctx, payment.PaymentRequest{
		To:      req.Counterparty,
		Amount:  req.Amount,
		Purpose: "partial settlement",
	})
	if err != nil {
		return partialsettlementexecution.DirectPaymentResult{}, err
	}
	return partialsettlementexecution.DirectPaymentResult{Reference: receipt.TxHash}, nil
}

type settlementPartialSettlementRuntime struct {
	runtime settlementExecutionRuntime
}

func (s settlementPartialSettlementRuntime) ExecuteSettlement(ctx context.Context, req partialsettlementexecution.DirectPaymentRequest) (partialsettlementexecution.DirectPaymentResult, error) {
	if s.runtime == nil {
		return partialsettlementexecution.DirectPaymentResult{}, fmt.Errorf("direct payment runtime is required")
	}

	result, err := s.runtime.ExecuteSettlement(ctx, settlementexecution.DirectPaymentRequest{
		TransactionReceiptID: req.TransactionReceiptID,
		SubmissionReceiptID:  req.SubmissionReceiptID,
		Counterparty:         req.Counterparty,
		Amount:               req.Amount,
	})
	if err != nil {
		return partialsettlementexecution.DirectPaymentResult{}, err
	}

	return partialsettlementexecution.DirectPaymentResult{Reference: result.Reference}, nil
}

type escrowDisputeHoldExecutionRuntime interface {
	Hold(context.Context, disputehold.EscrowHoldRequest) (disputehold.HoldResult, error)
}

type engineEscrowDisputeHoldRuntime struct {
	engine *escrow.Engine
}

func (e engineEscrowDisputeHoldRuntime) Hold(_ context.Context, req disputehold.EscrowHoldRequest) (disputehold.HoldResult, error) {
	if e.engine == nil {
		return disputehold.HoldResult{}, fmt.Errorf("escrow engine is required")
	}
	if strings.TrimSpace(req.EscrowReference) == "" {
		return disputehold.HoldResult{}, fmt.Errorf("escrow reference is required")
	}
	return disputehold.HoldResult{Reference: req.EscrowReference}, nil
}

type escrowReleaseExecutionRuntime interface {
	Release(context.Context, escrowrelease.ReleaseRequest) (escrowrelease.ReleaseResult, error)
}

type engineEscrowReleaseRuntime struct {
	engine *escrow.Engine
}

func (e engineEscrowReleaseRuntime) Release(ctx context.Context, req escrowrelease.ReleaseRequest) (escrowrelease.ReleaseResult, error) {
	if e.engine == nil {
		return escrowrelease.ReleaseResult{}, fmt.Errorf("escrow engine is required")
	}
	entry, err := e.engine.Release(ctx, req.EscrowReference)
	if err != nil {
		return escrowrelease.ReleaseResult{}, err
	}
	if entry == nil {
		return escrowrelease.ReleaseResult{}, fmt.Errorf("escrow release returned nil entry")
	}
	return escrowrelease.ReleaseResult{Reference: entry.ID}, nil
}

type escrowRefundExecutionRuntime interface {
	Refund(context.Context, escrowrefund.RefundRequest) (escrowrefund.RefundResult, error)
}

type engineEscrowRefundRuntime struct {
	engine *escrow.Engine
}

func (e engineEscrowRefundRuntime) Refund(ctx context.Context, req escrowrefund.RefundRequest) (escrowrefund.RefundResult, error) {
	if e.engine == nil {
		return escrowrefund.RefundResult{}, fmt.Errorf("escrow engine is required")
	}
	entry, err := e.engine.Refund(ctx, req.EscrowReference)
	if err != nil {
		return escrowrefund.RefundResult{}, err
	}
	if entry == nil {
		return escrowrefund.RefundResult{}, fmt.Errorf("escrow refund returned nil entry")
	}
	return escrowrefund.RefundResult{Reference: entry.ID}, nil
}

func newExecuteSettlementTool(receiptStore *receipts.Store, runtime settlementExecutionRuntime) *agent.Tool {
	if runtime == nil {
		return nil
	}

	return &agent.Tool{
		Name:        "execute_settlement",
		Description: "Execute direct final settlement for an approved transaction receipt and return canonical settlement state",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category: "knowledge",
			Activity: agent.ActivityWrite,
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"transaction_receipt_id": map[string]interface{}{"type": "string", "description": "Approved transaction receipt identifier to settle"},
			},
			"required": []string{"transaction_receipt_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if receiptStore == nil {
				return nil, fmt.Errorf("receipts store dependency is not configured")
			}

			transactionReceiptID, err := toolparam.RequireString(params, "transaction_receipt_id")
			if err != nil {
				return nil, err
			}
			transactionReceiptID = strings.TrimSpace(transactionReceiptID)
			if transactionReceiptID == "" {
				return nil, &toolparam.ErrMissingParam{Name: "transaction_receipt_id"}
			}

			result, err := settlementexecution.NewService(receiptStore, runtime).Execute(ctx, settlementexecution.ExecuteRequest{
				TransactionReceiptID: transactionReceiptID,
			})
			if err != nil {
				return nil, err
			}

			return newExecuteSettlementReceipt(result), nil
		},
	}
}

func newExecutePartialSettlementTool(receiptStore *receipts.Store, runtime partialSettlementExecutionRuntime) *agent.Tool {
	if runtime == nil {
		return nil
	}

	return &agent.Tool{
		Name:        "execute_partial_settlement",
		Description: "Execute direct partial settlement for an approved transaction receipt and return canonical partial settlement state",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category: "knowledge",
			Activity: agent.ActivityWrite,
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"transaction_receipt_id": map[string]interface{}{"type": "string", "description": "Approved transaction receipt identifier to partially settle"},
			},
			"required": []string{"transaction_receipt_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if receiptStore == nil {
				return nil, fmt.Errorf("receipts store dependency is not configured")
			}

			transactionReceiptID, err := toolparam.RequireString(params, "transaction_receipt_id")
			if err != nil {
				return nil, err
			}
			transactionReceiptID = strings.TrimSpace(transactionReceiptID)
			if transactionReceiptID == "" {
				return nil, &toolparam.ErrMissingParam{Name: "transaction_receipt_id"}
			}

			result, err := partialsettlementexecution.NewService(receiptStore, runtime).Execute(ctx, partialsettlementexecution.ExecuteRequest{
				TransactionReceiptID: transactionReceiptID,
			})
			if err != nil {
				return nil, err
			}

			return newExecutePartialSettlementReceipt(result), nil
		},
	}
}

func newHoldEscrowForDisputeTool(receiptStore *receipts.Store, runtime escrowDisputeHoldExecutionRuntime) *agent.Tool {
	if runtime == nil {
		return nil
	}

	return &agent.Tool{
		Name:        "hold_escrow_for_dispute",
		Description: "Record dispute hold evidence for a funded dispute-ready escrow and return canonical hold state",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category: "knowledge",
			Activity: agent.ActivityWrite,
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"transaction_receipt_id": map[string]interface{}{"type": "string", "description": "Funded dispute-ready transaction receipt identifier to hold"},
			},
			"required": []string{"transaction_receipt_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if receiptStore == nil {
				return nil, fmt.Errorf("receipts store dependency is not configured")
			}

			transactionReceiptID, err := toolparam.RequireString(params, "transaction_receipt_id")
			if err != nil {
				return nil, err
			}
			transactionReceiptID = strings.TrimSpace(transactionReceiptID)
			if transactionReceiptID == "" {
				return nil, &toolparam.ErrMissingParam{Name: "transaction_receipt_id"}
			}

			result, err := disputehold.NewService(receiptStore, runtime).Execute(ctx, disputehold.ExecuteRequest{
				TransactionReceiptID: transactionReceiptID,
			})
			if err != nil {
				return nil, err
			}

			return newHoldEscrowForDisputeReceipt(result), nil
		},
	}
}

func newReleaseEscrowSettlementTool(receiptStore *receipts.Store, runtime escrowReleaseExecutionRuntime) *agent.Tool {
	if runtime == nil {
		return nil
	}

	return &agent.Tool{
		Name:        "release_escrow_settlement",
		Description: "Release a funded escrow for an approved transaction receipt and return canonical settlement state",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category: "knowledge",
			Activity: agent.ActivityWrite,
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"transaction_receipt_id": map[string]interface{}{"type": "string", "description": "Funded transaction receipt identifier to release"},
			},
			"required": []string{"transaction_receipt_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if receiptStore == nil {
				return nil, fmt.Errorf("receipts store dependency is not configured")
			}

			transactionReceiptID, err := toolparam.RequireString(params, "transaction_receipt_id")
			if err != nil {
				return nil, err
			}
			transactionReceiptID = strings.TrimSpace(transactionReceiptID)
			if transactionReceiptID == "" {
				return nil, &toolparam.ErrMissingParam{Name: "transaction_receipt_id"}
			}

			result, err := escrowrelease.NewService(receiptStore, runtime).Execute(ctx, escrowrelease.ExecuteRequest{
				TransactionReceiptID: transactionReceiptID,
			})
			if err != nil {
				return nil, err
			}

			return newReleaseEscrowSettlementReceipt(result), nil
		},
	}
}

func newRefundEscrowSettlementTool(receiptStore *receipts.Store, runtime escrowRefundExecutionRuntime) *agent.Tool {
	if runtime == nil {
		return nil
	}

	return &agent.Tool{
		Name:        "refund_escrow_settlement",
		Description: "Refund a funded escrow from the settlement review path and return canonical refund state",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category: "knowledge",
			Activity: agent.ActivityWrite,
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"transaction_receipt_id": map[string]interface{}{"type": "string", "description": "Funded review-path transaction receipt identifier to refund"},
			},
			"required": []string{"transaction_receipt_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if receiptStore == nil {
				return nil, fmt.Errorf("receipts store dependency is not configured")
			}

			transactionReceiptID, err := toolparam.RequireString(params, "transaction_receipt_id")
			if err != nil {
				return nil, err
			}
			transactionReceiptID = strings.TrimSpace(transactionReceiptID)
			if transactionReceiptID == "" {
				return nil, &toolparam.ErrMissingParam{Name: "transaction_receipt_id"}
			}

			result, err := escrowrefund.NewService(receiptStore, runtime).Execute(ctx, escrowrefund.ExecuteRequest{
				TransactionReceiptID: transactionReceiptID,
			})
			if err != nil {
				return nil, err
			}

			return newRefundEscrowSettlementReceipt(result), nil
		},
	}
}

func parseDisputeReadyReceiptInput(params map[string]interface{}) (receipts.CreateSubmissionInput, error) {
	transactionID, err := toolparam.RequireString(params, "transaction_id")
	if err != nil {
		return receipts.CreateSubmissionInput{}, err
	}
	artifactLabel, err := toolparam.RequireString(params, "artifact_label")
	if err != nil {
		return receipts.CreateSubmissionInput{}, err
	}
	payloadHash, err := toolparam.RequireString(params, "payload_hash")
	if err != nil {
		return receipts.CreateSubmissionInput{}, err
	}
	sourceLineageDigest, err := toolparam.RequireString(params, "source_lineage_digest")
	if err != nil {
		return receipts.CreateSubmissionInput{}, err
	}

	return receipts.CreateSubmissionInput{
		TransactionID:       transactionID,
		ArtifactLabel:       artifactLabel,
		PayloadHash:         payloadHash,
		SourceLineageDigest: sourceLineageDigest,
	}, nil
}

func createDisputeReadyReceipt(ctx context.Context, receiptStore *receipts.Store, input receipts.CreateSubmissionInput) (interface{}, error) {
	submission, transaction, err := receiptStore.CreateSubmissionReceipt(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("create dispute-ready receipt: %w", err)
	}

	return map[string]interface{}{
		"submission_receipt_id":         submission.SubmissionReceiptID,
		"transaction_receipt_id":        transaction.TransactionReceiptID,
		"current_submission_receipt_id": transaction.CurrentSubmissionReceiptID,
	}, nil
}

type upfrontPaymentApprovalReceipt struct {
	TransactionReceiptID         string  `json:"transaction_receipt_id"`
	SubmissionReceiptID          string  `json:"submission_receipt_id"`
	Amount                       string  `json:"amount"`
	TrustScore                   float64 `json:"trust_score"`
	UserMaxPrepay                string  `json:"user_max_prepay"`
	RemainingBudget              string  `json:"remaining_budget"`
	Decision                     string  `json:"decision"`
	Reason                       string  `json:"reason"`
	PolicyCode                   string  `json:"policy_code,omitempty"`
	SuggestedMode                string  `json:"suggested_mode"`
	AmountClass                  string  `json:"amount_class,omitempty"`
	RiskClass                    string  `json:"risk_class,omitempty"`
	FailureDetail                string  `json:"failure_detail,omitempty"`
	CurrentPaymentApprovalStatus string  `json:"current_payment_approval_status"`
	CanonicalDecision            string  `json:"canonical_decision,omitempty"`
	CanonicalSettlementHint      string  `json:"canonical_settlement_hint,omitempty"`
	EscrowExecutionStatus        string  `json:"escrow_execution_status,omitempty"`
}

type executeEscrowRecommendationReceipt struct {
	TransactionReceiptID  string `json:"transaction_receipt_id"`
	SubmissionReceiptID   string `json:"submission_receipt_id"`
	EscrowReference       string `json:"escrow_reference,omitempty"`
	EscrowExecutionStatus string `json:"escrow_execution_status"`
}

type applySettlementProgressionReceipt struct {
	TransactionReceiptID            string `json:"transaction_receipt_id"`
	SettlementProgressionStatus     string `json:"settlement_progression_status"`
	SettlementProgressionReasonCode string `json:"settlement_progression_reason_code,omitempty"`
	SettlementProgressionReason     string `json:"settlement_progression_reason,omitempty"`
	PartialHint                     string `json:"partial_hint,omitempty"`
}

type executeSettlementReceipt struct {
	TransactionReceiptID        string `json:"transaction_receipt_id"`
	SubmissionReceiptID         string `json:"submission_receipt_id,omitempty"`
	SettlementProgressionStatus string `json:"settlement_progression_status"`
	ResolvedAmount              string `json:"resolved_amount,omitempty"`
	RuntimeReference            string `json:"runtime_reference,omitempty"`
}

type executePartialSettlementReceipt struct {
	TransactionReceiptID        string `json:"transaction_receipt_id"`
	SubmissionReceiptID         string `json:"submission_receipt_id,omitempty"`
	SettlementProgressionStatus string `json:"settlement_progression_status"`
	ExecutedAmount              string `json:"executed_amount,omitempty"`
	RemainingAmount             string `json:"remaining_amount,omitempty"`
	RuntimeReference            string `json:"runtime_reference,omitempty"`
}

type releaseEscrowSettlementReceipt struct {
	TransactionReceiptID        string `json:"transaction_receipt_id"`
	SubmissionReceiptID         string `json:"submission_receipt_id,omitempty"`
	SettlementProgressionStatus string `json:"settlement_progression_status"`
	ResolvedAmount              string `json:"resolved_amount,omitempty"`
	RuntimeReference            string `json:"runtime_reference,omitempty"`
}

type refundEscrowSettlementReceipt struct {
	TransactionReceiptID        string `json:"transaction_receipt_id"`
	SubmissionReceiptID         string `json:"submission_receipt_id,omitempty"`
	SettlementProgressionStatus string `json:"settlement_progression_status"`
	ResolvedAmount              string `json:"resolved_amount,omitempty"`
	RuntimeReference            string `json:"runtime_reference,omitempty"`
}

type holdEscrowForDisputeReceipt struct {
	TransactionReceiptID        string `json:"transaction_receipt_id"`
	SubmissionReceiptID         string `json:"submission_receipt_id,omitempty"`
	SettlementProgressionStatus string `json:"settlement_progression_status"`
	EscrowReference             string `json:"escrow_reference,omitempty"`
	RuntimeReference            string `json:"runtime_reference,omitempty"`
}

func newUpfrontPaymentApprovalReceipt(
	transactionReceiptID string,
	submissionReceiptID string,
	amount string,
	trustScore float64,
	userMaxPrepay string,
	remainingBudget string,
	outcome paymentapproval.Outcome,
	updatedTx receipts.TransactionReceipt,
) upfrontPaymentApprovalReceipt {
	return upfrontPaymentApprovalReceipt{
		TransactionReceiptID:         transactionReceiptID,
		SubmissionReceiptID:          submissionReceiptID,
		Amount:                       amount,
		TrustScore:                   trustScore,
		UserMaxPrepay:                userMaxPrepay,
		RemainingBudget:              remainingBudget,
		Decision:                     string(outcome.Decision),
		Reason:                       outcome.Reason,
		PolicyCode:                   outcome.PolicyCode,
		SuggestedMode:                string(outcome.SuggestedMode),
		AmountClass:                  string(outcome.AmountClass),
		RiskClass:                    string(outcome.RiskClass),
		FailureDetail:                outcome.FailureDetail,
		CurrentPaymentApprovalStatus: string(updatedTx.CurrentPaymentApprovalStatus),
		CanonicalDecision:            updatedTx.CanonicalDecision,
		CanonicalSettlementHint:      updatedTx.CanonicalSettlementHint,
		EscrowExecutionStatus:        string(updatedTx.EscrowExecutionStatus),
	}
}

func newApplySettlementProgressionReceipt(result settlementprogression.ApplyReleaseOutcomeResult) applySettlementProgressionReceipt {
	return applySettlementProgressionReceipt{
		TransactionReceiptID:            result.Transaction.TransactionReceiptID,
		SettlementProgressionStatus:     string(result.Transaction.SettlementProgressionStatus),
		SettlementProgressionReasonCode: string(result.Transaction.SettlementProgressionReasonCode),
		SettlementProgressionReason:     result.Transaction.SettlementProgressionReason,
		PartialHint:                     result.Transaction.PartialSettlementHint,
	}
}

func newExecuteSettlementReceipt(result settlementexecution.Result) executeSettlementReceipt {
	return executeSettlementReceipt{
		TransactionReceiptID:        result.TransactionReceiptID,
		SubmissionReceiptID:         result.SubmissionReceiptID,
		SettlementProgressionStatus: string(result.SettlementProgressionStatus),
		ResolvedAmount:              result.ResolvedAmount,
		RuntimeReference:            result.RuntimeReference,
	}
}

func newExecutePartialSettlementReceipt(result partialsettlementexecution.Result) executePartialSettlementReceipt {
	return executePartialSettlementReceipt{
		TransactionReceiptID:        result.TransactionReceiptID,
		SubmissionReceiptID:         result.SubmissionReceiptID,
		SettlementProgressionStatus: string(result.SettlementProgressionStatus),
		ExecutedAmount:              result.ExecutedAmount,
		RemainingAmount:             result.RemainingAmount,
		RuntimeReference:            result.RuntimeReference,
	}
}

func newReleaseEscrowSettlementReceipt(result escrowrelease.Result) releaseEscrowSettlementReceipt {
	return releaseEscrowSettlementReceipt{
		TransactionReceiptID:        result.TransactionReceiptID,
		SubmissionReceiptID:         result.SubmissionReceiptID,
		SettlementProgressionStatus: string(result.SettlementProgressionStatus),
		ResolvedAmount:              result.ResolvedAmount,
		RuntimeReference:            result.RuntimeReference,
	}
}

func newRefundEscrowSettlementReceipt(result escrowrefund.Result) refundEscrowSettlementReceipt {
	return refundEscrowSettlementReceipt{
		TransactionReceiptID:        result.TransactionReceiptID,
		SubmissionReceiptID:         result.SubmissionReceiptID,
		SettlementProgressionStatus: string(result.SettlementProgressionStatus),
		ResolvedAmount:              result.ResolvedAmount,
		RuntimeReference:            result.RuntimeReference,
	}
}

func newHoldEscrowForDisputeReceipt(result disputehold.Result) holdEscrowForDisputeReceipt {
	return holdEscrowForDisputeReceipt{
		TransactionReceiptID:        result.TransactionReceiptID,
		SubmissionReceiptID:         result.SubmissionReceiptID,
		SettlementProgressionStatus: string(result.SettlementProgressionStatus),
		EscrowReference:             result.EscrowReference,
		RuntimeReference:            result.RuntimeReference,
	}
}
