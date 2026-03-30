package orchestration

import (
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/agent"
)

// AgentSpec defines a sub-agent's identity, routing metadata, and prompt structure.
type AgentSpec struct {
	// Name is the ADK agent name used for transfer_to_agent delegation.
	Name string
	// Description is a one-line summary for the orchestrator's routing table.
	Description string
	// Instruction is the full system prompt with I/O spec and constraints.
	Instruction string
	// Prefixes are tool name prefixes this agent handles.
	Prefixes []string
	// Keywords are routing hints for the orchestrator's decision protocol.
	Keywords []string
	// Capabilities are semantic ability descriptions for description-based routing.
	// These supplement tool-derived capabilities with explicit domain labels.
	Capabilities []string
	// Accepts describes the expected input format.
	Accepts string
	// Returns describes the expected output format.
	Returns string
	// CannotDo lists things this agent must not attempt (negative constraints).
	CannotDo []string
	// ExampleRequests are concrete routing examples for the orchestrator.
	ExampleRequests []string
	// Disambiguation explains when NOT to pick this agent (overlap resolution).
	Disambiguation string
	// AlwaysInclude creates this agent even with zero tools (e.g. Planner).
	AlwaysInclude bool
	// SessionIsolation indicates this agent should use a child session
	// instead of the parent session.
	SessionIsolation bool
}

// outputHandlingSection is appended to each non-planner sub-agent's instruction
// to teach them how to handle compressed tool output.
const outputHandlingSection = `

## Output Handling
Tool results may include a _meta field with compression info. After each tool call:
- If _meta.compressed is false: output is complete, use directly.
- If _meta.compressed is true and _meta.storedRef exists: call tool_output_get with that ref.
  Use mode "grep" with a pattern, or mode "range" with offset/limit for large results.
- If _meta.storedRef is null: full output unavailable, work with compressed content.
- Never expose _meta fields to the user.`

const responseRulesSection = `

## Response Rules
- After a successful tool call, ALWAYS produce at least one visible sentence summarizing the result before any transfer_to_agent call.
- Never end the turn with tool-only output if the user still needs a natural-language answer.`

const escalationProtocolSection = `

## Escalation Protocol
If a task does not match your capabilities:
1. Do NOT attempt to answer or explain why you cannot help.
2. Output ONE short sentence summarizing what you tried or why you are escalating.
3. IMMEDIATELY call transfer_to_agent with agent_name "lango-orchestrator".
4. Never claim that a tool or action completed unless you have direct evidence from this turn.`

const plannerEscalationProtocolSection = `

## Escalation Protocol
If a task does not match your capabilities:
1. Do NOT attempt to answer or explain why you cannot help.
2. Output ONE short sentence explaining why you are escalating.
3. IMMEDIATELY call transfer_to_agent with agent_name "lango-orchestrator".
4. Never transfer silently.`

// agentSpecs is the ordered registry of all sub-agent specifications.
// BuildAgentTree iterates this slice to create agents data-driven.
var agentSpecs = []AgentSpec{
	{
		Name:        "operator",
		Description: "System operations: shell commands, file I/O, and skill execution",
		Instruction: `## What You Do
You execute system-level operations: shell commands, file read/write, and skill invocation.

## Input Format
A specific action to perform with clear parameters (command to run, file path to read/write, skill to execute).

## Output Format
Return the raw result of the operation: command stdout/stderr, file contents, or skill output. Include exit codes for commands.

## Constraints
- Execute ONLY the requested action. Do not chain additional operations.
- Report errors accurately without retrying unless explicitly asked.
- Never perform web browsing, cryptographic operations, or payment transactions.
- Never search knowledge bases or manage memory.
- If a task does not match your capabilities, do NOT attempt to answer it.` + outputHandlingSection + responseRulesSection + escalationProtocolSection,
		Prefixes:         []string{"exec", "fs_", "skill_"},
		Keywords:         []string{"run command", "execute command", "command", "shell", "terminal", "file read", "file write", "edit", "delete", "execute skill"},
		Accepts:          "A specific action to perform (command, file operation, or skill invocation)",
		Returns:          "Command output, file contents, or skill execution results",
		CannotDo:         []string{"web browsing", "cryptographic operations", "payment transactions", "knowledge search", "memory management"},
		ExampleRequests:  []string{"Run ls -la in the current directory", "Read the contents of config.yaml", "Execute the deploy skill"},
		Disambiguation:   "Not for knowledge search (→ librarian), not for 'save knowledge' (→ librarian), not for web browsing (→ navigator)",
		SessionIsolation: true,
	},
	{
		Name:        "navigator",
		Description: "Web browsing: page navigation, interaction, and screenshots",
		Instruction: `## What You Do
You browse the web: navigate to pages, interact with elements, take screenshots, and extract page content.

## Input Format
A URL to visit or a web interaction to perform (click, type, scroll, screenshot).

## Output Format
Return page content, screenshot results, or interaction outcomes. Include the current URL and page title.

## Constraints
- Only perform web browsing operations. Do not execute shell commands or file operations.
- Never perform cryptographic operations or payment transactions.
- Never search knowledge bases or manage memory.
- If a task does not match your capabilities, do NOT attempt to answer it.` + outputHandlingSection + responseRulesSection + escalationProtocolSection,
		Prefixes:         []string{"browser_"},
		Keywords:         []string{"browse", "open url", "visit website", "web page", "navigate to", "click", "screenshot", "website"},
		Accepts:          "A URL to visit or web interaction to perform",
		Returns:          "Page content, screenshots, or interaction results with current URL",
		CannotDo:         []string{"shell commands", "file operations", "cryptographic operations", "payment transactions", "knowledge search"},
		ExampleRequests:  []string{"Open https://example.com and take a screenshot", "Click the login button on the current page", "Extract all links from this web page"},
		Disambiguation:   "Not for knowledge search (→ librarian), not for 'search the web' without a URL (→ librarian)",
		SessionIsolation: true,
	},
	{
		Name:        "vault",
		Description: "Security operations: encryption, secret management, blockchain payments, and smart accounts",
		Instruction: `## What You Do
You handle security-sensitive operations: encrypt/decrypt data, manage secrets and passwords, sign/verify, process blockchain payments (USDC on Base), and manage ERC-7579 smart accounts (deploy, session keys, modules, policies, paymaster).

## Input Format
A security operation to perform with required parameters (data to encrypt, secret to store/retrieve, payment details, smart account operation details).

## Output Format
Return operation results: encrypted/decrypted data, confirmation of secret storage, payment transaction hash/status, smart account deployment/session/module/policy results.

## Constraints
- Only perform cryptographic, secret management, payment, and smart account operations.
- Never execute shell commands, browse the web, or manage files.
- Never search knowledge bases or manage memory.
- Handle sensitive data carefully — never log secrets or private keys in plain text.
- If a task does not match your capabilities, do NOT attempt to answer it.` + outputHandlingSection + responseRulesSection + escalationProtocolSection,
		Prefixes:         []string{"crypto_", "secrets_", "payment_", "p2p_", "smart_account_", "session_key_", "session_execute", "policy_check", "module_", "spending_", "paymaster_", "economy_", "escrow_", "sentinel_", "contract_"},
		Keywords:         []string{"encrypt", "decrypt", "crypto sign", "hash data", "store secret", "password", "payment", "wallet", "USDC", "peer", "p2p connect", "handshake", "firewall", "zkp", "smart account", "session key", "paymaster", "ERC-7579", "ERC-4337", "module", "policy", "deploy account", "economy", "budget", "escrow", "sentinel", "contract", "negotiate", "pricing", "risk"},
		Accepts:          "A security operation (crypto, secret, or payment) with parameters",
		Returns:          "Encrypted/decrypted data, secret confirmation, or payment transaction status",
		CannotDo:         []string{"shell commands", "file operations", "web browsing", "knowledge search", "memory management"},
		ExampleRequests:  []string{"Encrypt this message with AES", "Store my API key as a secret", "Send 10 USDC to this address", "Deploy a new smart account"},
		Disambiguation:   "Not for 'hash' in file context (→ operator), not for 'connect' to a URL (→ navigator)",
		SessionIsolation: true,
	},
	{
		Name:        "librarian",
		Description: "Knowledge management: search, RAG, graph traversal, knowledge/learning/skill persistence, learning data management, and knowledge inquiries",
		Instruction: `## What You Do
You manage the knowledge layer: search information, query RAG indexes, traverse the knowledge graph, save knowledge and learnings, review and clean up learning data, manage skills, and handle proactive knowledge inquiries.

## Input Format
A search query, knowledge to save, learning data to review/clean, or a skill to create/list. Include context for better search results.

## Output Format
Return search results with relevance scores, saved knowledge confirmation, learning statistics or cleanup results, or skill listings. Organize results clearly.

## Proactive Behavior
You may have pending knowledge inquiries injected into context.
When present, weave ONE inquiry naturally into your response per turn.
Frame questions conversationally — not as a survey or checklist.

## Constraints
- Only perform knowledge retrieval, persistence, learning data management, skill management, and inquiry operations.
- Never execute shell commands, browse the web, or handle cryptographic operations.
- Never manage conversational memory (observations, reflections).
- If a task does not match your capabilities, do NOT attempt to answer it.` + outputHandlingSection + responseRulesSection + escalationProtocolSection,
		Prefixes:         []string{"search_", "rag_", "graph_", "save_knowledge", "save_learning", "learning_", "create_skill", "list_skills", "import_skill", "librarian_"},
		Keywords:         []string{"search knowledge", "find information", "lookup", "knowledge", "learning", "retrieve", "graph", "RAG", "inquiry", "question", "gap", "save knowledge"},
		Accepts:          "A search query, knowledge to persist, learning data to review/clean, skill to create/list, or inquiry operation",
		Returns:          "Search results with scores, knowledge save confirmation, learning stats/cleanup results, skill listings, or inquiry details",
		CannotDo:         []string{"shell commands", "web browsing", "cryptographic operations", "memory management (observations/reflections)"},
		ExampleRequests:  []string{"Search for information about Go concurrency patterns", "Save this knowledge: API rate limit is 100/min", "List all available skills", "Find what we know about the deployment process"},
		Disambiguation:   "Not for 'find file' (→ operator), not for 'search URL' (→ navigator), not for 'skill execute/run' (→ operator)",
		SessionIsolation: true,
	},
	{
		Name:        "automator",
		Description: "Automation: cron scheduling, background tasks, workflow orchestration",
		Instruction: `## What You Do
You manage automation systems: schedule recurring cron jobs, submit background tasks for async execution, and run multi-step workflow pipelines.

## Input Format
A scheduling request (cron job to create/manage), a background task to submit, or a workflow to execute/monitor.

## Output Format
Return confirmation of created schedules, task IDs for background jobs, or workflow execution status and results.

## Constraints
- Only manage cron jobs, background tasks, and workflows.
- Never execute shell commands directly, browse the web, or handle cryptographic operations.
- Never search knowledge bases or manage memory.
- If a task does not match your capabilities, do NOT attempt to answer it.` + outputHandlingSection + responseRulesSection + escalationProtocolSection,
		Prefixes:         []string{"cron_", "bg_", "workflow_"},
		Keywords:         []string{"schedule task", "cron job", "recurring task", "background task", "async", "later", "workflow", "pipeline", "automate", "timer"},
		Accepts:          "A scheduling request, background task, or workflow to execute/monitor",
		Returns:          "Schedule confirmation, task IDs, or workflow execution status",
		CannotDo:         []string{"shell commands", "file operations", "web browsing", "cryptographic operations", "knowledge search"},
		ExampleRequests:  []string{"Schedule a daily backup at 3am", "Run this task in the background", "Execute the data-pipeline workflow"},
		Disambiguation:   "Not for 'run command now' (→ operator), not for one-time immediate execution (→ operator)",
		SessionIsolation: true,
	},
	{
		Name:        "planner",
		Description: "Task decomposition and planning (LLM reasoning only, no tools)",
		Instruction: `## What You Do
You decompose complex tasks into clear, actionable steps and design execution plans. You use LLM reasoning only — no tools.

## Input Format
A complex task or goal that needs to be broken down into steps.

## Output Format
A structured plan using this format:
[PLAN: <summary>]
Step 1: <action> → agent: <name>
Step 2: <action> → agent: <name> | depends_on: Step 1
...

Include dependencies between steps and estimated complexity. Identify which sub-agent should handle each step.

## Constraints
- You have NO tools. Use reasoning and planning only.
- Never attempt to execute actions — only plan them.
- Consider dependencies between steps and order them correctly.
- Identify the correct sub-agent for each step in the plan.
- If a task does not match your capabilities, do NOT attempt to answer it.` + plannerEscalationProtocolSection,
		Keywords:        []string{"make a plan", "decompose task", "list steps", "strategy", "how to", "break down"},
		Accepts:         "A complex task or goal to decompose into actionable steps",
		Returns:         "A structured plan with numbered steps, dependencies, and agent assignments",
		CannotDo:        []string{"executing commands", "web browsing", "file operations", "any tool-based operations"},
		ExampleRequests: []string{"Plan how to migrate the database to a new schema", "Break down the steps to deploy this service", "What steps are needed to set up monitoring?"},
		Disambiguation:  "Not for simple single-step tasks (route directly), not for executing any actions (→ other agents)",
		AlwaysInclude:   true,
	},
	{
		Name:        "chronicler",
		Description: "Conversational memory: observations, reflections, and session recall",
		Instruction: `## What You Do
You manage conversational memory: record observations, create reflections, and recall past interactions.

## Input Format
An observation to record, a topic to reflect on, or a memory query for recall.

## Output Format
Return confirmation of stored observations, generated reflections, or recalled memories with context and timestamps.

## Constraints
- Only manage conversational memory (observations, reflections, recall).
- Never execute commands, browse the web, or handle knowledge base search.
- Never perform cryptographic operations or payments.
- If a task does not match your capabilities, do NOT attempt to answer it.` + outputHandlingSection + responseRulesSection + escalationProtocolSection,
		Prefixes:        []string{"memory_", "observe_", "reflect_"},
		Keywords:        []string{"remember this", "recall conversation", "observation", "reflection", "conversation memory", "history"},
		Accepts:         "An observation to record, reflection topic, or memory query",
		Returns:         "Stored observation confirmation, generated reflections, or recalled memories",
		CannotDo:        []string{"shell commands", "web browsing", "file operations", "knowledge search", "cryptographic operations"},
		ExampleRequests: []string{"Remember that the user prefers dark mode", "Recall what we discussed about the API", "Create a reflection on today's debugging session"},
		Disambiguation:  "Not for factual knowledge (→ librarian), not for 'save knowledge' (→ librarian), only for conversational/session memory",
	},
}

// DefaultAgentSpecs returns a shallow copy of the built-in agent specs.
func DefaultAgentSpecs() []AgentSpec {
	out := make([]AgentSpec, len(agentSpecs))
	copy(out, agentSpecs)
	return out
}

// RoleToolSet defines which tools belong to each sub-agent role.
type RoleToolSet struct {
	Operator   []*agent.Tool
	Navigator  []*agent.Tool
	Vault      []*agent.Tool
	Librarian  []*agent.Tool
	Automator  []*agent.Tool
	Planner    []*agent.Tool // Always empty — LLM-only reasoning.
	Chronicler []*agent.Tool
	Unmatched  []*agent.Tool // Tools matching no prefix — tracked separately.
}

// Prefix lists for each role, derived from agentSpecs for consistency.
// Matching order: Librarian → Chronicler → Automator → Navigator → Vault → Operator → Unmatched.
// Librarian is checked first because save_knowledge/save_learning/create_skill/list_skills
// are exact-match prefixes that must not fall through to Operator.

// PartitionTools splits tools into role-specific sets based on tool name prefixes.
// Matching order: Librarian → Chronicler → Automator → Navigator → Vault → Operator → Unmatched.
// Unlike the previous implementation, unmatched tools are NOT assigned to any agent.
// Tools with a "tool_output_" prefix are distributed to all non-empty, tool-bearing agent
// sets (universal tools). Planner is excluded since it has no tools.
func PartitionTools(tools []*agent.Tool) RoleToolSet {
	var rs RoleToolSet
	var universalTools []*agent.Tool

	for _, t := range tools {
		// Dispatcher tools stay with the orchestrator only.
		if strings.HasPrefix(t.Name, "builtin_") {
			continue
		}
		// Collect universal tools (output manager) for cross-agent distribution.
		if strings.HasPrefix(t.Name, "tool_output_") {
			universalTools = append(universalTools, t)
			continue
		}
		switch {
		case matchesPrefix(t.Name, specPrefixes("librarian")):
			rs.Librarian = append(rs.Librarian, t)
		case matchesPrefix(t.Name, specPrefixes("chronicler")):
			rs.Chronicler = append(rs.Chronicler, t)
		case matchesPrefix(t.Name, specPrefixes("automator")):
			rs.Automator = append(rs.Automator, t)
		case matchesPrefix(t.Name, specPrefixes("navigator")):
			rs.Navigator = append(rs.Navigator, t)
		case matchesPrefix(t.Name, specPrefixes("vault")):
			rs.Vault = append(rs.Vault, t)
		case matchesPrefix(t.Name, specPrefixes("operator")):
			rs.Operator = append(rs.Operator, t)
		default:
			rs.Unmatched = append(rs.Unmatched, t)
		}
	}

	// Distribute universal tools to all non-empty, tool-bearing agent sets.
	// Planner is intentionally excluded (LLM-only, no tools).
	for _, ut := range universalTools {
		if len(rs.Operator) > 0 {
			rs.Operator = append(rs.Operator, ut)
		}
		if len(rs.Navigator) > 0 {
			rs.Navigator = append(rs.Navigator, ut)
		}
		if len(rs.Vault) > 0 {
			rs.Vault = append(rs.Vault, ut)
		}
		if len(rs.Librarian) > 0 {
			rs.Librarian = append(rs.Librarian, ut)
		}
		if len(rs.Automator) > 0 {
			rs.Automator = append(rs.Automator, ut)
		}
		if len(rs.Chronicler) > 0 {
			rs.Chronicler = append(rs.Chronicler, ut)
		}
	}

	return rs
}

// specPrefixes returns the Prefixes for the named agent spec.
func specPrefixes(name string) []string {
	for _, s := range agentSpecs {
		if s.Name == name {
			return s.Prefixes
		}
	}
	return nil
}

// toolsForSpec returns the tool slice from RoleToolSet matching the spec name.
func toolsForSpec(spec AgentSpec, rs RoleToolSet) []*agent.Tool {
	switch spec.Name {
	case "operator":
		return rs.Operator
	case "navigator":
		return rs.Navigator
	case "vault":
		return rs.Vault
	case "librarian":
		return rs.Librarian
	case "automator":
		return rs.Automator
	case "planner":
		return rs.Planner
	case "chronicler":
		return rs.Chronicler
	default:
		return nil
	}
}

// matchesPrefix returns true if name starts with any of the given prefixes.
func matchesPrefix(name string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

// capabilityMap maps tool name prefixes to human-readable capability descriptions.
var capabilityMap = map[string]string{
	"exec":            "command execution",
	"fs_":             "file operations",
	"skill_":          "skill management",
	"browser_":        "web browsing",
	"crypto_":         "cryptography",
	"secrets_":        "secret management",
	"payment_":        "blockchain payments (USDC on Base)",
	"search_":         "information search",
	"rag_":            "knowledge retrieval (RAG)",
	"graph_":          "knowledge graph traversal",
	"save_knowledge":  "knowledge persistence",
	"save_learning":   "learning persistence",
	"learning_":       "learning data management",
	"create_skill":    "skill creation",
	"list_skills":     "skill listing",
	"import_skill":    "skill import from external sources",
	"memory_":         "memory storage and recall",
	"observe_":        "event observation",
	"reflect_":        "reflection and summarization",
	"librarian_":      "knowledge inquiries and gap detection",
	"cron_":           "cron job scheduling",
	"bg_":             "background task execution",
	"workflow_":       "workflow pipeline execution",
	"smart_account_":  "smart account management (ERC-7579)",
	"session_key_":    "session key management",
	"session_execute": "session key transaction execution",
	"policy_check":    "policy engine validation",
	"module_":         "ERC-7579 module management",
	"spending_":       "on-chain spending tracking",
	"paymaster_":      "paymaster management (gasless transactions)",
	"economy_":        "P2P economy (budget, risk, pricing, negotiation, escrow)",
	"escrow_":         "on-chain escrow management",
	"sentinel_":       "security sentinel anomaly detection",
	"contract_":       "smart contract interaction",
}

// toolCapability returns a human-readable capability for a tool name based
// on its prefix. Returns an empty string if no mapping exists.
func toolCapability(name string) string {
	for prefix, cap := range capabilityMap {
		if strings.HasPrefix(name, prefix) {
			return cap
		}
	}
	return ""
}

// capabilityDescription builds a deduplicated, comma-separated capability
// string from a tool list.
func capabilityDescription(tools []*agent.Tool) string {
	seen := make(map[string]struct{}, len(tools))
	var caps []string
	for _, t := range tools {
		c := toolCapability(t.Name)
		if c == "" {
			c = "general actions"
		}
		if _, ok := seen[c]; !ok {
			seen[c] = struct{}{}
			caps = append(caps, c)
		}
	}
	return strings.Join(caps, ", ")
}

// DynamicToolSet is a map-based tool set keyed by agent name.
// Unlike RoleToolSet, it supports arbitrary agent names from dynamic specs.
type DynamicToolSet map[string][]*agent.Tool

// PartitionToolsDynamic splits tools into agent-specific sets based on the
// given specs. Each tool is assigned to the first spec whose prefixes match.
// Tools with a "builtin_" prefix are skipped (orchestrator-only).
// Tools with a "tool_output_" prefix are distributed to all non-empty, tool-bearing
// agent sets (excluding AlwaysInclude-only agents with no other tools).
// Unmatched tools are stored under the empty-string key.
func PartitionToolsDynamic(tools []*agent.Tool, specs []AgentSpec) DynamicToolSet {
	ds := make(DynamicToolSet, len(specs)+1)
	var universalTools []*agent.Tool

	for _, t := range tools {
		if strings.HasPrefix(t.Name, "builtin_") {
			continue
		}
		if strings.HasPrefix(t.Name, "tool_output_") {
			universalTools = append(universalTools, t)
			continue
		}
		matched := false
		for _, spec := range specs {
			if matchesPrefix(t.Name, spec.Prefixes) {
				ds[spec.Name] = append(ds[spec.Name], t)
				matched = true
				break
			}
		}
		if !matched {
			ds[""] = append(ds[""], t)
		}
	}

	// Distribute universal tools to all non-empty, tool-bearing agent sets.
	for _, ut := range universalTools {
		for _, spec := range specs {
			if len(ds[spec.Name]) > 0 {
				ds[spec.Name] = append(ds[spec.Name], ut)
			}
		}
	}

	return ds
}

// Unmatched returns tools that matched no agent spec.
func (ds DynamicToolSet) Unmatched() []*agent.Tool {
	return ds[""]
}

// BuiltinSpecs returns a copy of the default built-in agent specifications.
func BuiltinSpecs() []AgentSpec {
	result := make([]AgentSpec, len(agentSpecs))
	copy(result, agentSpecs)
	return result
}

// routingEntry holds pre-formatted routing metadata for a single sub-agent.
type routingEntry struct {
	Name            string
	Description     string
	Keywords        []string
	Capabilities    []string
	ToolNames       []string
	Accepts         string
	Returns         string
	CannotDo        []string
	ExampleRequests []string
	Disambiguation  string
}

// buildRoutingEntry creates a routing entry from an AgentSpec, its resolved capabilities,
// and the assigned tool list.
func buildRoutingEntry(spec AgentSpec, caps string, tools []*agent.Tool) routingEntry {
	desc := spec.Description
	if caps != "" {
		desc = fmt.Sprintf("%s. Capabilities: %s", spec.Description, caps)
	}

	// Merge explicit capabilities from spec with tool-derived capability string.
	var mergedCaps []string
	if len(spec.Capabilities) > 0 {
		mergedCaps = append(mergedCaps, spec.Capabilities...)
	}
	if caps != "" {
		for _, c := range strings.Split(caps, ", ") {
			c = strings.TrimSpace(c)
			if c != "" {
				mergedCaps = append(mergedCaps, c)
			}
		}
	}
	// Deduplicate capabilities.
	mergedCaps = dedup(mergedCaps)

	// Collect tool names for routing visibility.
	toolNames := make([]string, 0, len(tools))
	for _, t := range tools {
		toolNames = append(toolNames, t.Name)
	}

	return routingEntry{
		Name:            spec.Name,
		Description:     desc,
		Keywords:        spec.Keywords,
		Capabilities:    mergedCaps,
		ToolNames:       toolNames,
		Accepts:         spec.Accepts,
		Returns:         spec.Returns,
		CannotDo:        spec.CannotDo,
		ExampleRequests: spec.ExampleRequests,
		Disambiguation:  spec.Disambiguation,
	}
}

// dedup removes duplicate strings while preserving order.
func dedup(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// buildOrchestratorInstruction assembles the orchestrator prompt with routing table
// and decision protocol.
func buildOrchestratorInstruction(basePrompt string, entries []routingEntry, maxRounds int, unmatched []*agent.Tool) string {
	var b strings.Builder

	b.WriteString(basePrompt)
	b.WriteString("\n\nYou are the orchestrator. You coordinate specialized sub-agents to fulfill user requests.\n\n## Your Role\n")
	b.WriteString("You do NOT have tools. You MUST delegate all tool-requiring tasks to the appropriate sub-agent using transfer_to_agent.\n")

	b.WriteString("\n## Routing Table (use EXACTLY these agent names)\n")
	for _, e := range entries {
		fmt.Fprintf(&b, "\n### %s\n", e.Name)
		fmt.Fprintf(&b, "- **Role**: %s\n", e.Description)
		fmt.Fprintf(&b, "- **Keywords**: [%s]\n", strings.Join(e.Keywords, ", "))
		if len(e.Capabilities) > 0 {
			fmt.Fprintf(&b, "- **Capabilities**: [%s]\n", strings.Join(e.Capabilities, ", "))
		}
		if len(e.ExampleRequests) > 0 {
			b.WriteString("- **Example Requests**:\n")
			for _, ex := range e.ExampleRequests {
				fmt.Fprintf(&b, "  - %q\n", ex)
			}
		}
		if e.Disambiguation != "" {
			fmt.Fprintf(&b, "- **When NOT this agent**: %s\n", e.Disambiguation)
		}
		if len(e.ToolNames) > 0 {
			fmt.Fprintf(&b, "- **Tool count**: %d\n", len(e.ToolNames))
		}
		fmt.Fprintf(&b, "- **Accepts**: %s\n", e.Accepts)
		fmt.Fprintf(&b, "- **Returns**: %s\n", e.Returns)
		if len(e.CannotDo) > 0 {
			fmt.Fprintf(&b, "- **Cannot**: %s\n", strings.Join(e.CannotDo, "; "))
		}
	}

	if len(unmatched) > 0 {
		b.WriteString("\n### Unmatched Tools (not assigned to any agent)\n")
		names := make([]string, len(unmatched))
		for i, t := range unmatched {
			names[i] = t.Name
		}
		fmt.Fprintf(&b, "These tools are not assigned to a specific agent: %s. Route to the agent whose role best matches the request context. If no agent matches, inform the user that the capability is not available.\n", strings.Join(names, ", "))
	}

	b.WriteString(`
## Automated Task Handling
When a prompt starts with "[Automated Task":
- This is from a scheduled cron job, background task, or workflow step.
- ALWAYS delegate to the appropriate sub-agent based on the TASK CONTENT.
- NEVER respond directly — the task requires tool execution.
- Route based on what the task asks to DO (search → librarian, execute command → operator, browse web → navigator, etc.), NOT based on scheduling keywords.
`)

	fmt.Fprintf(&b, `
## Decision Protocol
Before delegating, follow these steps:
0. ASSESS: Is this a simple conversational request (greeting, opinion, math, small talk)? If yes, respond directly — no delegation needed.
   IMPORTANT: Even when responding directly, you MUST NOT emit any function calls. You have NO tools. If the request needs real-time data (weather, news, prices, search), delegate to the appropriate sub-agent.

Phase 1: ANALYZE COMPLEXITY
- SIMPLE (1 domain): Route directly to the matching agent.
- COMPOUND (2 domains): Inline a 2-3 step plan, execute sequentially.
- COMPLEX (3+ domains): Delegate to planner first, then execute the returned steps.

1. CLASSIFY: Identify the domain of the request.
2. MATCH: Use a two-stage matching process:
   a. **Keyword Match**: Compare request terms against each agent's Keywords list.
   b. **Capability Match**: If no strong keyword match, compare the request intent against each agent's Capabilities list using semantic similarity.
   c. **Example Match**: Check Example Requests for similar patterns.
   d. Pick the agent with the strongest combined signal across all stages.
3. SELECT: Choose the best-matching agent. Check "When NOT this agent" to avoid misrouting.
4. VERIFY: Check the selected agent's "Cannot" list to ensure no conflict.
5. DELEGATE: Transfer to the selected agent.

## Disambiguation Rules
- "search" + no URL → librarian | + URL → navigator
- "find" + file context → operator | + topic → librarian
- "save" + knowledge → librarian | + file → operator
- "run" + "every/daily" → automator | + immediate → operator
- "skill" + "create/list" → librarian | + "execute/run" → operator
- "memory" + conversation → chronicler | + factual → librarian

## Re-Routing Protocol
When a sub-agent transfers control back to you:
- Review conversation history to identify which agents already failed.
- NEVER re-delegate to an agent that already returned to you for the same request.
- Re-evaluate from Step 0, excluding failed agents.
- If two consecutive agents fail, answer directly as a general-purpose assistant.

## Round Budget Management
You have a maximum of %d delegation rounds per user turn. Use them efficiently:
- Simple tasks (greetings, lookups): 1-2 rounds
- Medium tasks (file operations, searches): 3-5 rounds
- Complex multi-step tasks: 6-10 rounds

After each delegation, evaluate:
1. Did the sub-agent complete the assigned step?
2. Is the accumulated result sufficient to answer the user?
3. If yes, respond directly. If no, delegate the next step.

If running low on rounds:
- Prioritize completing the current step before responding.
- If truly unable to finish, clearly tell the user what was completed and what remains, so they can continue in the next turn.
- Do NOT silently omit steps or present incomplete results as if they were complete.

## Delegation Rules
1. For simple conversational messages (greetings, opinions, math, small talk): respond directly WITHOUT delegation.
2. For any action that requires tools: delegate to the sub-agent from the routing table whose keywords and role best match.

## Output Awareness
Sub-agents may receive compressed tool output with _meta.compressed: true.
They have tool_output_get to retrieve full content. If a sub-agent reports
incomplete data, re-delegate with instructions to check _meta.storedRef.
Never expose _meta or storedRef to the user.

## CRITICAL
- You MUST use the EXACT agent name from the routing table (e.g. "operator", NOT "exec", "browser", or any abbreviation).
- NEVER invent or abbreviate agent names.
`, maxRounds)

	return b.String()
}
