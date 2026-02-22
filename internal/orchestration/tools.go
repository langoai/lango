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
	// Accepts describes the expected input format.
	Accepts string
	// Returns describes the expected output format.
	Returns string
	// CannotDo lists things this agent must not attempt (negative constraints).
	CannotDo []string
	// AlwaysInclude creates this agent even with zero tools (e.g. Planner).
	AlwaysInclude bool
}

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
- If a task does not match your capabilities, REJECT it by responding:
  "[REJECT] This task requires <correct_agent>. I handle: shell commands, file I/O, skill execution."`,
		Prefixes: []string{"exec", "fs_", "skill_"},
		Keywords: []string{"run", "execute", "command", "shell", "file", "read", "write", "edit", "delete", "skill"},
		Accepts:  "A specific action to perform (command, file operation, or skill invocation)",
		Returns:  "Command output, file contents, or skill execution results",
		CannotDo: []string{"web browsing", "cryptographic operations", "payment transactions", "knowledge search", "memory management"},
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
- If a task does not match your capabilities, REJECT it by responding:
  "[REJECT] This task requires <correct_agent>. I handle: web browsing, page navigation, screenshots."`,
		Prefixes: []string{"browser_"},
		Keywords: []string{"browse", "web", "url", "page", "navigate", "click", "screenshot", "website"},
		Accepts:  "A URL to visit or web interaction to perform",
		Returns:  "Page content, screenshots, or interaction results with current URL",
		CannotDo: []string{"shell commands", "file operations", "cryptographic operations", "payment transactions", "knowledge search"},
	},
	{
		Name:        "vault",
		Description: "Security operations: encryption, secret management, and blockchain payments",
		Instruction: `## What You Do
You handle security-sensitive operations: encrypt/decrypt data, manage secrets and passwords, sign/verify, and process blockchain payments (USDC on Base).

## Input Format
A security operation to perform with required parameters (data to encrypt, secret to store/retrieve, payment details).

## Output Format
Return operation results: encrypted/decrypted data, confirmation of secret storage, payment transaction hash/status.

## Constraints
- Only perform cryptographic, secret management, and payment operations.
- Never execute shell commands, browse the web, or manage files.
- Never search knowledge bases or manage memory.
- Handle sensitive data carefully — never log secrets or private keys in plain text.
- If a task does not match your capabilities, REJECT it by responding:
  "[REJECT] This task requires <correct_agent>. I handle: encryption, secret management, blockchain payments."`,
		Prefixes: []string{"crypto_", "secrets_", "payment_", "p2p_"},
		Keywords: []string{"encrypt", "decrypt", "sign", "hash", "secret", "password", "payment", "wallet", "USDC", "peer", "p2p", "connect", "handshake", "firewall", "zkp"},
		Accepts:  "A security operation (crypto, secret, or payment) with parameters",
		Returns:  "Encrypted/decrypted data, secret confirmation, or payment transaction status",
		CannotDo: []string{"shell commands", "file operations", "web browsing", "knowledge search", "memory management"},
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
- If a task does not match your capabilities, REJECT it by responding:
  "[REJECT] This task requires <correct_agent>. I handle: search, RAG, graph traversal, knowledge/learning/skill management, inquiries."`,
		Prefixes: []string{"search_", "rag_", "graph_", "save_knowledge", "save_learning", "learning_", "create_skill", "list_skills", "import_skill", "librarian_"},
		Keywords: []string{"search", "find", "lookup", "knowledge", "learning", "retrieve", "graph", "RAG", "inquiry", "question", "gap"},
		Accepts:  "A search query, knowledge to persist, learning data to review/clean, skill to create/list, or inquiry operation",
		Returns:  "Search results with scores, knowledge save confirmation, learning stats/cleanup results, skill listings, or inquiry details",
		CannotDo: []string{"shell commands", "web browsing", "cryptographic operations", "memory management (observations/reflections)"},
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
- If a task does not match your capabilities, REJECT it by responding:
  "[REJECT] This task requires <correct_agent>. I handle: cron scheduling, background tasks, workflow pipelines."`,
		Prefixes: []string{"cron_", "bg_", "workflow_"},
		Keywords: []string{"schedule", "cron", "every", "recurring", "background",
			"async", "later", "workflow", "pipeline", "automate", "timer"},
		Accepts:  "A scheduling request, background task, or workflow to execute/monitor",
		Returns:  "Schedule confirmation, task IDs, or workflow execution status",
		CannotDo: []string{"shell commands", "file operations", "web browsing", "cryptographic operations", "knowledge search"},
	},
	{
		Name:        "planner",
		Description: "Task decomposition and planning (LLM reasoning only, no tools)",
		Instruction: `## What You Do
You decompose complex tasks into clear, actionable steps and design execution plans. You use LLM reasoning only — no tools.

## Input Format
A complex task or goal that needs to be broken down into steps.

## Output Format
A structured plan with numbered steps, dependencies between steps, and estimated complexity. Identify which sub-agent should handle each step.

## Constraints
- You have NO tools. Use reasoning and planning only.
- Never attempt to execute actions — only plan them.
- Consider dependencies between steps and order them correctly.
- Identify the correct sub-agent for each step in the plan.
- If a task does not match your capabilities, REJECT it by responding:
  "[REJECT] This task requires <correct_agent>. I handle: task decomposition and planning."`,
		Keywords:      []string{"plan", "decompose", "steps", "strategy", "how to", "break down"},
		Accepts:       "A complex task or goal to decompose into actionable steps",
		Returns:       "A structured plan with numbered steps, dependencies, and agent assignments",
		CannotDo:      []string{"executing commands", "web browsing", "file operations", "any tool-based operations"},
		AlwaysInclude: true,
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
- If a task does not match your capabilities, REJECT it by responding:
  "[REJECT] This task requires <correct_agent>. I handle: observations, reflections, memory recall."`,
		Prefixes: []string{"memory_", "observe_", "reflect_"},
		Keywords: []string{"remember", "recall", "observation", "reflection", "memory", "history"},
		Accepts:  "An observation to record, reflection topic, or memory query",
		Returns:  "Stored observation confirmation, generated reflections, or recalled memories",
		CannotDo: []string{"shell commands", "web browsing", "file operations", "knowledge search", "cryptographic operations"},
	},
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
func PartitionTools(tools []*agent.Tool) RoleToolSet {
	var rs RoleToolSet
	for _, t := range tools {
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
	"exec":           "command execution",
	"fs_":            "file operations",
	"skill_":         "skill management",
	"browser_":       "web browsing",
	"crypto_":        "cryptography",
	"secrets_":       "secret management",
	"payment_":       "blockchain payments (USDC on Base)",
	"search_":        "information search",
	"rag_":           "knowledge retrieval (RAG)",
	"graph_":         "knowledge graph traversal",
	"save_knowledge": "knowledge persistence",
	"save_learning":  "learning persistence",
	"learning_":      "learning data management",
	"create_skill":   "skill creation",
	"list_skills":    "skill listing",
	"import_skill":   "skill import from external sources",
	"memory_":        "memory storage and recall",
	"observe_":       "event observation",
	"reflect_":       "reflection and summarization",
	"librarian_":     "knowledge inquiries and gap detection",
	"cron_":          "cron job scheduling",
	"bg_":            "background task execution",
	"workflow_":      "workflow pipeline execution",
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

// routingEntry holds pre-formatted routing metadata for a single sub-agent.
type routingEntry struct {
	Name        string
	Description string
	Keywords    []string
	Accepts     string
	Returns     string
	CannotDo    []string
}

// buildRoutingEntry creates a routing entry from an AgentSpec and its resolved capabilities.
func buildRoutingEntry(spec AgentSpec, caps string) routingEntry {
	desc := spec.Description
	if caps != "" {
		desc = fmt.Sprintf("%s. Capabilities: %s", spec.Description, caps)
	}
	return routingEntry{
		Name:        spec.Name,
		Description: desc,
		Keywords:    spec.Keywords,
		Accepts:     spec.Accepts,
		Returns:     spec.Returns,
		CannotDo:    spec.CannotDo,
	}
}

// buildOrchestratorInstruction assembles the orchestrator prompt with routing table
// and decision protocol.
func buildOrchestratorInstruction(basePrompt string, entries []routingEntry, maxRounds int, unmatched []*agent.Tool) string {
	var b strings.Builder

	b.WriteString(basePrompt)
	b.WriteString(`

You are the orchestrator. You coordinate specialized sub-agents to fulfill user requests.

## Your Role
You do NOT have tools. You MUST delegate all tool-requiring tasks to the appropriate sub-agent using transfer_to_agent.

## Routing Table (use EXACTLY these agent names)
`)
	for _, e := range entries {
		b.WriteString(fmt.Sprintf("\n### %s\n", e.Name))
		b.WriteString(fmt.Sprintf("- **Role**: %s\n", e.Description))
		b.WriteString(fmt.Sprintf("- **Keywords**: [%s]\n", strings.Join(e.Keywords, ", ")))
		b.WriteString(fmt.Sprintf("- **Accepts**: %s\n", e.Accepts))
		b.WriteString(fmt.Sprintf("- **Returns**: %s\n", e.Returns))
		if len(e.CannotDo) > 0 {
			b.WriteString(fmt.Sprintf("- **Cannot**: %s\n", strings.Join(e.CannotDo, "; ")))
		}
	}

	if len(unmatched) > 0 {
		b.WriteString("\n### Unmatched Tools (not assigned to any agent)\n")
		names := make([]string, len(unmatched))
		for i, t := range unmatched {
			names[i] = t.Name
		}
		b.WriteString(fmt.Sprintf("The following tools are available but not assigned to a specific agent: %s. Handle requests for these tools directly or choose the closest matching agent.\n", strings.Join(names, ", ")))
	}

	b.WriteString(fmt.Sprintf(`
## Decision Protocol
Before delegating, follow these steps:
1. CLASSIFY: Identify the domain of the request.
2. MATCH: Compare keywords against the routing table.
3. SELECT: Choose the best-matching agent.
4. VERIFY: Check the selected agent's "Cannot" list to ensure no conflict.
5. DELEGATE: Transfer to the selected agent.

## Rejection Handling
If a sub-agent rejects a task with [REJECT], try the next most relevant agent or handle the request directly.

## Delegation Rules
1. For any action that requires tools: delegate to the sub-agent from the routing table whose keywords and role best match.
2. For simple conversational messages (greetings, opinions, general knowledge): respond directly without delegation.
3. Maximum %d delegation rounds per user turn.

## CRITICAL
- You MUST use the EXACT agent name from the routing table (e.g. "operator", NOT "exec", "browser", or any abbreviation).
- NEVER invent or abbreviate agent names.
`, maxRounds))

	return b.String()
}
