package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/langowarny/lango/internal/agent"
	"github.com/langowarny/lango/internal/approval"
	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/embedding"
	"github.com/langowarny/lango/internal/graph"
	"github.com/langowarny/lango/internal/knowledge"
	"github.com/langowarny/lango/internal/learning"
	"github.com/langowarny/lango/internal/memory"
	"github.com/langowarny/lango/internal/payment"
	"github.com/langowarny/lango/internal/security"
	"github.com/langowarny/lango/internal/session"
	"github.com/langowarny/lango/internal/skill"
	"github.com/langowarny/lango/internal/supervisor"
	"github.com/langowarny/lango/internal/tools/browser"
	toolcrypto "github.com/langowarny/lango/internal/tools/crypto"
	"github.com/langowarny/lango/internal/tools/filesystem"
	toolpayment "github.com/langowarny/lango/internal/tools/payment"
	toolsecrets "github.com/langowarny/lango/internal/tools/secrets"
	"github.com/langowarny/lango/internal/wallet"
)

// buildTools creates the set of tools available to the agent.
// When browserSM is non-nil, browser tools are included.
func buildTools(sv *supervisor.Supervisor, fsCfg filesystem.Config, browserSM *browser.SessionManager) []*agent.Tool {
	var tools []*agent.Tool

	// Exec tools (delegated to Supervisor for security isolation)
	tools = append(tools, buildExecTools(sv)...)

	// Filesystem tools
	fsTool := filesystem.New(fsCfg)
	tools = append(tools, buildFilesystemTools(fsTool)...)

	// Browser tools (opt-in)
	if browserSM != nil {
		tools = append(tools, buildBrowserTools(browserSM)...)
	}

	return tools
}

func buildExecTools(sv *supervisor.Supervisor) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "exec",
			Description: "Execute shell commands",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The shell command to execute",
					},
				},
				"required": []string{"command"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				cmd, ok := params["command"].(string)
				if !ok {
					return nil, fmt.Errorf("missing command parameter")
				}
				return sv.ExecuteTool(ctx, cmd)
			},
		},
		{
			Name:        "exec_bg",
			Description: "Execute a shell command in the background",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The shell command to execute",
					},
				},
				"required": []string{"command"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				cmd, ok := params["command"].(string)
				if !ok {
					return nil, fmt.Errorf("missing command parameter")
				}
				return sv.StartBackground(cmd)
			},
		},
		{
			Name:        "exec_status",
			Description: "Check the status of a background process",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "The background process ID returned by exec_bg",
					},
				},
				"required": []string{"id"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				id, ok := params["id"].(string)
				if !ok {
					return nil, fmt.Errorf("missing id parameter")
				}
				return sv.GetBackgroundStatus(id)
			},
		},
		{
			Name:        "exec_stop",
			Description: "Stop a background process",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "The background process ID returned by exec_bg",
					},
				},
				"required": []string{"id"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				id, ok := params["id"].(string)
				if !ok {
					return nil, fmt.Errorf("missing id parameter")
				}
				return nil, sv.StopBackground(id)
			},
		},
	}
}

func buildFilesystemTools(fsTool *filesystem.Tool) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "fs_read",
			Description: "Read a file",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "The file path to read"},
				},
				"required": []string{"path"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				path, ok := params["path"].(string)
				if !ok {
					return nil, fmt.Errorf("missing path parameter")
				}
				return fsTool.Read(path)
			},
		},
		{
			Name:        "fs_list",
			Description: "List contents of a directory",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "The directory path to list"},
				},
				"required": []string{"path"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				path, _ := params["path"].(string)
				if path == "" {
					path = "."
				}
				return fsTool.ListDir(path)
			},
		},
		{
			Name:        "fs_write",
			Description: "Write content to a file",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":    map[string]interface{}{"type": "string", "description": "The file path to write to"},
					"content": map[string]interface{}{"type": "string", "description": "The content to write"},
				},
				"required": []string{"path", "content"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				path, _ := params["path"].(string)
				content, _ := params["content"].(string)
				if path == "" {
					return nil, fmt.Errorf("missing path parameter")
				}
				return nil, fsTool.Write(path, content)
			},
		},
		{
			Name:        "fs_edit",
			Description: "Edit a file by replacing a line range",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":      map[string]interface{}{"type": "string", "description": "The file path to edit"},
					"startLine": map[string]interface{}{"type": "integer", "description": "The starting line number (1-indexed)"},
					"endLine":   map[string]interface{}{"type": "integer", "description": "The ending line number (inclusive)"},
					"content":   map[string]interface{}{"type": "string", "description": "The new content for the specified range"},
				},
				"required": []string{"path", "startLine", "endLine", "content"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				path, _ := params["path"].(string)
				content, _ := params["content"].(string)
				if path == "" {
					return nil, fmt.Errorf("missing path parameter")
				}

				var startLine, endLine int
				if sl, ok := params["startLine"].(float64); ok {
					startLine = int(sl)
				} else if sl, ok := params["startLine"].(int); ok {
					startLine = sl
				}
				if el, ok := params["endLine"].(float64); ok {
					endLine = int(el)
				} else if el, ok := params["endLine"].(int); ok {
					endLine = el
				}

				return nil, fsTool.Edit(path, startLine, endLine, content)
			},
		},
		{
			Name:        "fs_mkdir",
			Description: "Create a directory",
			SafetyLevel: agent.SafetyLevelModerate,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "The directory path to create"},
				},
				"required": []string{"path"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				path, _ := params["path"].(string)
				if path == "" {
					return nil, fmt.Errorf("missing path parameter")
				}
				return nil, fsTool.Mkdir(path)
			},
		},
		{
			Name:        "fs_delete",
			Description: "Delete a file or directory",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "The path to delete"},
				},
				"required": []string{"path"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				path, _ := params["path"].(string)
				if path == "" {
					return nil, fmt.Errorf("missing path parameter")
				}
				return nil, fsTool.Delete(path)
			},
		},
	}
}

func buildBrowserTools(sm *browser.SessionManager) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "browser_navigate",
			Description: "Navigate the browser to a URL and return the page title, URL, and a text snippet",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "The URL to navigate to",
					},
				},
				"required": []string{"url"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				url, ok := params["url"].(string)
				if !ok || url == "" {
					return nil, fmt.Errorf("missing url parameter")
				}

				sessionID, err := sm.EnsureSession()
				if err != nil {
					return nil, err
				}

				if err := sm.Tool().Navigate(ctx, sessionID, url); err != nil {
					return nil, err
				}

				return sm.Tool().GetSnapshot(sessionID)
			},
		},
		{
			Name:        "browser_action",
			Description: "Perform an action on the current browser page: click, type, eval, get_text, get_element_info, or wait",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{
						"type":        "string",
						"description": "The action to perform",
						"enum":        []string{"click", "type", "eval", "get_text", "get_element_info", "wait"},
					},
					"selector": map[string]interface{}{
						"type":        "string",
						"description": "CSS selector for the target element (required for click, type, get_text, get_element_info, wait)",
					},
					"text": map[string]interface{}{
						"type":        "string",
						"description": "Text to type (required for type action) or JavaScript to evaluate (required for eval action)",
					},
					"timeout": map[string]interface{}{
						"type":        "integer",
						"description": "Timeout in seconds for wait action (default: 10)",
					},
				},
				"required": []string{"action"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				action, ok := params["action"].(string)
				if !ok || action == "" {
					return nil, fmt.Errorf("missing action parameter")
				}

				sessionID, err := sm.EnsureSession()
				if err != nil {
					return nil, err
				}

				selector, _ := params["selector"].(string)
				text, _ := params["text"].(string)

				switch action {
				case "click":
					if selector == "" {
						return nil, fmt.Errorf("selector required for click action")
					}
					return nil, sm.Tool().Click(ctx, sessionID, selector)

				case "type":
					if selector == "" {
						return nil, fmt.Errorf("selector required for type action")
					}
					if text == "" {
						return nil, fmt.Errorf("text required for type action")
					}
					return nil, sm.Tool().Type(ctx, sessionID, selector, text)

				case "eval":
					if text == "" {
						return nil, fmt.Errorf("text (JavaScript) required for eval action")
					}
					return sm.Tool().Eval(sessionID, text)

				case "get_text":
					if selector == "" {
						return nil, fmt.Errorf("selector required for get_text action")
					}
					return sm.Tool().GetText(sessionID, selector)

				case "get_element_info":
					if selector == "" {
						return nil, fmt.Errorf("selector required for get_element_info action")
					}
					return sm.Tool().GetElementInfo(sessionID, selector)

				case "wait":
					if selector == "" {
						return nil, fmt.Errorf("selector required for wait action")
					}
					timeout := 10 * time.Second
					if t, ok := params["timeout"].(float64); ok && t > 0 {
						timeout = time.Duration(t) * time.Second
					}
					return nil, sm.Tool().WaitForSelector(ctx, sessionID, selector, timeout)

				default:
					return nil, fmt.Errorf("unknown action: %s", action)
				}
			},
		},
		{
			Name:        "browser_screenshot",
			Description: "Capture a screenshot of the current browser page as base64 PNG",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"fullPage": map[string]interface{}{
						"type":        "boolean",
						"description": "Capture the full scrollable page (default: false)",
					},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				sessionID, err := sm.EnsureSession()
				if err != nil {
					return nil, err
				}

				fullPage, _ := params["fullPage"].(bool)
				return sm.Tool().Screenshot(sessionID, fullPage)
			},
		},
	}
}

// buildMetaTools creates knowledge/learning/skill meta-tools for the agent.
func buildMetaTools(store *knowledge.Store, engine *learning.Engine, registry *skill.Registry, autoApprove bool) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "save_knowledge",
			Description: "Save a piece of knowledge (user rule, definition, preference, or fact) for future reference",
			SafetyLevel: agent.SafetyLevelModerate,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"key":      map[string]interface{}{"type": "string", "description": "Unique key for this knowledge entry"},
					"category": map[string]interface{}{"type": "string", "description": "Category: rule, definition, preference, or fact", "enum": []string{"rule", "definition", "preference", "fact"}},
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
					Category: category,
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
					Category:     category,
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

				entry := knowledge.SkillEntry{
					Name:             name,
					Description:      description,
					Type:             skillType,
					Definition:       definition,
					Parameters:       parameters,
					CreatedBy:        "agent",
					RequiresApproval: !autoApprove,
				}

				if err := registry.CreateSkill(ctx, entry); err != nil {
					return nil, fmt.Errorf("create skill: %w", err)
				}

				status := "draft"
				if autoApprove {
					if err := registry.ActivateSkill(ctx, name); err != nil {
						return nil, fmt.Errorf("activate skill: %w", err)
					}
					status = "active"
				}

				if err := store.SaveAuditLog(ctx, knowledge.AuditEntry{
					Action: "skill_create",
					Actor:  "agent",
					Target: name,
					Details: map[string]interface{}{
						"type":   skillType,
						"status": status,
					},
				}); err != nil {
					logger().Warnw("audit log save failed", "action", "skill_create", "error", err)
				}

				return map[string]interface{}{
					"status":  status,
					"name":    name,
					"message": fmt.Sprintf("Skill '%s' created with status '%s'", name, status),
				}, nil
			},
		},
		{
			Name:        "list_skills",
			Description: "List all active skills with usage statistics",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				skills, err := store.ListActiveSkills(ctx)
				if err != nil {
					return nil, fmt.Errorf("list skills: %w", err)
				}

				return map[string]interface{}{
					"skills": skills,
					"count":  len(skills),
				}, nil
			},
		},
	}
}

// wrapWithLearning wraps a tool's handler to call the learning observer after each execution.
func wrapWithLearning(t *agent.Tool, observer learning.ToolResultObserver) *agent.Tool {
	original := t.Handler
	return &agent.Tool{
		Name:        t.Name,
		Description: t.Description,
		Parameters:  t.Parameters,
		SafetyLevel: t.SafetyLevel,
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			result, err := original(ctx, params)
			sessionKey := session.SessionKeyFromContext(ctx)
			observer.OnToolResult(ctx, sessionKey, t.Name, params, result, err)
			return result, err
		},
	}
}

// buildCryptoTools wraps crypto.Tool methods as agent tools.
func buildCryptoTools(crypto security.CryptoProvider, keys *security.KeyRegistry, refs *security.RefStore, scanner *agent.SecretScanner) []*agent.Tool {
	ct := toolcrypto.New(crypto, keys, refs, scanner)
	return []*agent.Tool{
		{
			Name:        "crypto_encrypt",
			Description: "Encrypt data using a registered key",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"data":  map[string]interface{}{"type": "string", "description": "The data to encrypt"},
					"keyId": map[string]interface{}{"type": "string", "description": "Key ID to use (default: default key)"},
				},
				"required": []string{"data"},
			},
			Handler: ct.Encrypt,
		},
		{
			Name:        "crypto_decrypt",
			Description: "Decrypt data using a registered key. Returns an opaque {{decrypt:id}} reference token. The decrypted value never enters the agent context.",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"ciphertext": map[string]interface{}{"type": "string", "description": "Base64-encoded ciphertext to decrypt"},
					"keyId":      map[string]interface{}{"type": "string", "description": "Key ID to use (default: default key)"},
				},
				"required": []string{"ciphertext"},
			},
			Handler: ct.Decrypt,
		},
		{
			Name:        "crypto_sign",
			Description: "Generate a digital signature for data",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"data":  map[string]interface{}{"type": "string", "description": "The data to sign"},
					"keyId": map[string]interface{}{"type": "string", "description": "Key ID to use"},
				},
				"required": []string{"data"},
			},
			Handler: ct.Sign,
		},
		{
			Name:        "crypto_hash",
			Description: "Compute a cryptographic hash of data",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"data":      map[string]interface{}{"type": "string", "description": "The data to hash"},
					"algorithm": map[string]interface{}{"type": "string", "description": "Hash algorithm: sha256 or sha512", "enum": []string{"sha256", "sha512"}},
				},
				"required": []string{"data"},
			},
			Handler: ct.Hash,
		},
		{
			Name:        "crypto_keys",
			Description: "List all registered cryptographic keys",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: ct.Keys,
		},
	}
}

// buildSecretsTools wraps secrets.Tool methods as agent tools.
func buildSecretsTools(secretsStore *security.SecretsStore, refs *security.RefStore, scanner *agent.SecretScanner) []*agent.Tool {
	st := toolsecrets.New(secretsStore, refs, scanner)
	return []*agent.Tool{
		{
			Name:        "secrets_store",
			Description: "Encrypt and store a secret value",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":  map[string]interface{}{"type": "string", "description": "Unique name for the secret"},
					"value": map[string]interface{}{"type": "string", "description": "The secret value to store"},
				},
				"required": []string{"name", "value"},
			},
			Handler: st.Store,
		},
		{
			Name:        "secrets_get",
			Description: "Retrieve a stored secret as a reference token. Returns an opaque {{secret:name}} token that is resolved at execution time by exec tools. The actual secret value never enters the agent context.",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{"type": "string", "description": "Name of the secret to retrieve"},
				},
				"required": []string{"name"},
			},
			Handler: st.Get,
		},
		{
			Name:        "secrets_list",
			Description: "List all stored secrets (metadata only, no values)",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: st.List,
		},
		{
			Name:        "secrets_delete",
			Description: "Delete a stored secret",
			SafetyLevel: agent.SafetyLevelDangerous,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{"type": "string", "description": "Name of the secret to delete"},
				},
				"required": []string{"name"},
			},
			Handler: st.Delete,
		},
	}
}

// needsApproval determines whether a tool requires approval based on the
// configured policy, explicit exemptions, and sensitive tool lists.
func needsApproval(t *agent.Tool, ic config.InterceptorConfig) bool {
	// ExemptTools always bypass approval.
	for _, name := range ic.ExemptTools {
		if name == t.Name {
			return false
		}
	}

	// SensitiveTools always require approval.
	for _, name := range ic.SensitiveTools {
		if name == t.Name {
			return true
		}
	}

	switch ic.ApprovalPolicy {
	case config.ApprovalPolicyAll:
		return true
	case config.ApprovalPolicyDangerous:
		return t.SafetyLevel.IsDangerous()
	case config.ApprovalPolicyConfigured:
		return false // only SensitiveTools (handled above)
	case config.ApprovalPolicyNone:
		return false
	default:
		return true // unknown policy → fail-safe
	}
}

// buildApprovalSummary returns a human-readable description of what a tool
// invocation will do, suitable for display in approval messages.
func buildApprovalSummary(toolName string, params map[string]interface{}) string {
	switch toolName {
	case "exec", "exec_bg":
		if cmd, ok := params["command"].(string); ok {
			return "Execute: " + truncate(cmd, 200)
		}
	case "fs_write":
		path, _ := params["path"].(string)
		content, _ := params["content"].(string)
		return fmt.Sprintf("Write to %s (%d bytes)", path, len(content))
	case "fs_edit":
		path, _ := params["path"].(string)
		return "Edit file: " + path
	case "fs_delete":
		path, _ := params["path"].(string)
		return "Delete: " + path
	case "browser_navigate":
		url, _ := params["url"].(string)
		return "Navigate to: " + truncate(url, 200)
	case "browser_action":
		action, _ := params["action"].(string)
		selector, _ := params["selector"].(string)
		if selector != "" {
			return fmt.Sprintf("Browser %s on: %s", action, truncate(selector, 100))
		}
		return "Browser action: " + action
	case "secrets_store":
		name, _ := params["name"].(string)
		return "Store secret: " + name
	case "secrets_get":
		name, _ := params["name"].(string)
		return "Retrieve secret: " + name
	case "secrets_delete":
		name, _ := params["name"].(string)
		return "Delete secret: " + name
	case "crypto_encrypt":
		return "Encrypt data"
	case "crypto_decrypt":
		return "Decrypt ciphertext"
	case "crypto_sign":
		return "Generate digital signature"
	case "payment_send":
		amount, _ := params["amount"].(string)
		to, _ := params["to"].(string)
		purpose, _ := params["purpose"].(string)
		return fmt.Sprintf("Send %s USDC to %s (%s)", amount, truncate(to, 12), truncate(purpose, 50))
	}
	return "Tool: " + toolName
}

// truncate shortens s to maxLen characters, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// wrapWithApproval wraps a tool to require approval based on the configured policy.
// Uses fail-closed: denies execution unless explicitly approved.
// The approval.Provider routes requests to the appropriate channel (Gateway, Telegram, Discord, Slack, TTY).
func wrapWithApproval(t *agent.Tool, ic config.InterceptorConfig, ap approval.Provider) *agent.Tool {
	if !needsApproval(t, ic) {
		return t
	}

	original := t.Handler
	return &agent.Tool{
		Name:        t.Name,
		Description: t.Description,
		Parameters:  t.Parameters,
		SafetyLevel: t.SafetyLevel,
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			req := approval.ApprovalRequest{
				ID:         fmt.Sprintf("req-%d", time.Now().UnixNano()),
				ToolName:   t.Name,
				SessionKey: session.SessionKeyFromContext(ctx),
				Params:     params,
				Summary:    buildApprovalSummary(t.Name, params),
				CreatedAt:  time.Now(),
			}
			approved, err := ap.RequestApproval(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("tool '%s' approval: %w", t.Name, err)
			}
			if !approved {
				sk := session.SessionKeyFromContext(ctx)
				if sk == "" {
					return nil, fmt.Errorf("tool '%s' execution denied: no approval channel available (session key missing)", t.Name)
				}
				return nil, fmt.Errorf("tool '%s' execution denied: user did not approve the action", t.Name)
			}
			return original(ctx, params)
		},
	}
}

// buildGraphTools creates tools for graph traversal and querying.
func buildGraphTools(gs graph.Store) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "graph_traverse",
			Description: "Traverse the knowledge graph from a start node using BFS. Returns related triples up to the specified depth.",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"start_node": map[string]interface{}{"type": "string", "description": "The node ID to start traversal from"},
					"max_depth":  map[string]interface{}{"type": "integer", "description": "Maximum traversal depth (default: 2)"},
					"predicates": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Filter by predicate types (empty = all)"},
				},
				"required": []string{"start_node"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				startNode, _ := params["start_node"].(string)
				if startNode == "" {
					return nil, fmt.Errorf("missing start_node parameter")
				}
				maxDepth := 2
				if d, ok := params["max_depth"].(float64); ok && d > 0 {
					maxDepth = int(d)
				}
				var predicates []string
				if raw, ok := params["predicates"].([]interface{}); ok {
					for _, p := range raw {
						if s, ok := p.(string); ok {
							predicates = append(predicates, s)
						}
					}
				}
				triples, err := gs.Traverse(ctx, startNode, maxDepth, predicates)
				if err != nil {
					return nil, fmt.Errorf("graph traverse: %w", err)
				}
				return map[string]interface{}{"triples": triples, "count": len(triples)}, nil
			},
		},
		{
			Name:        "graph_query",
			Description: "Query the knowledge graph by subject or object node. Returns matching triples.",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"subject":   map[string]interface{}{"type": "string", "description": "Subject node to query by"},
					"object":    map[string]interface{}{"type": "string", "description": "Object node to query by"},
					"predicate": map[string]interface{}{"type": "string", "description": "Optional predicate filter (used with subject)"},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				subject, _ := params["subject"].(string)
				object, _ := params["object"].(string)
				predicate, _ := params["predicate"].(string)

				if subject == "" && object == "" {
					return nil, fmt.Errorf("either subject or object is required")
				}

				var triples []graph.Triple
				var err error
				if subject != "" && predicate != "" {
					triples, err = gs.QueryBySubjectPredicate(ctx, subject, predicate)
				} else if subject != "" {
					triples, err = gs.QueryBySubject(ctx, subject)
				} else {
					triples, err = gs.QueryByObject(ctx, object)
				}
				if err != nil {
					return nil, fmt.Errorf("graph query: %w", err)
				}
				return map[string]interface{}{"triples": triples, "count": len(triples)}, nil
			},
		},
	}
}

// buildRAGTools creates tools for RAG retrieval.
func buildRAGTools(ragSvc *embedding.RAGService) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "rag_retrieve",
			Description: "Retrieve semantically similar content from the knowledge base using vector search.",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query":       map[string]interface{}{"type": "string", "description": "The search query"},
					"limit":       map[string]interface{}{"type": "integer", "description": "Maximum results to return (default: 5)"},
					"collections": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Filter by collections (e.g., knowledge, observation)"},
				},
				"required": []string{"query"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				query, _ := params["query"].(string)
				if query == "" {
					return nil, fmt.Errorf("missing query parameter")
				}
				limit := 5
				if l, ok := params["limit"].(float64); ok && l > 0 {
					limit = int(l)
				}
				var collections []string
				if raw, ok := params["collections"].([]interface{}); ok {
					for _, c := range raw {
						if s, ok := c.(string); ok {
							collections = append(collections, s)
						}
					}
				}
				sessionKey := session.SessionKeyFromContext(ctx)
				results, err := ragSvc.Retrieve(ctx, query, embedding.RetrieveOptions{
					Limit:       limit,
					Collections: collections,
					SessionKey:  sessionKey,
				})
				if err != nil {
					return nil, fmt.Errorf("rag retrieve: %w", err)
				}
				return map[string]interface{}{"results": results, "count": len(results)}, nil
			},
		},
	}
}

// buildMemoryAgentTools creates tools for observational memory management.
func buildMemoryAgentTools(ms *memory.Store) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "memory_list_observations",
			Description: "List observations for a session. Returns compressed notes from conversation history.",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"session_key": map[string]interface{}{"type": "string", "description": "Session key to list observations for (uses current session if empty)"},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				sessionKey, _ := params["session_key"].(string)
				if sessionKey == "" {
					sessionKey = session.SessionKeyFromContext(ctx)
				}
				observations, err := ms.ListObservations(ctx, sessionKey)
				if err != nil {
					return nil, fmt.Errorf("list observations: %w", err)
				}
				return map[string]interface{}{"observations": observations, "count": len(observations)}, nil
			},
		},
		{
			Name:        "memory_list_reflections",
			Description: "List reflections for a session. Reflections are condensed observations across time.",
			SafetyLevel: agent.SafetyLevelSafe,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"session_key": map[string]interface{}{"type": "string", "description": "Session key to list reflections for (uses current session if empty)"},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				sessionKey, _ := params["session_key"].(string)
				if sessionKey == "" {
					sessionKey = session.SessionKeyFromContext(ctx)
				}
				reflections, err := ms.ListReflections(ctx, sessionKey)
				if err != nil {
					return nil, fmt.Errorf("list reflections: %w", err)
				}
				return map[string]interface{}{"reflections": reflections, "count": len(reflections)}, nil
			},
		},
	}
}

// buildPaymentTools creates blockchain payment tools.
func buildPaymentTools(svc *payment.Service, limiter wallet.SpendingLimiter) []*agent.Tool {
	return toolpayment.BuildTools(svc, limiter)
}
