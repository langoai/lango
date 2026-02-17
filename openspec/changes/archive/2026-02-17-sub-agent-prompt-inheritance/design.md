## Context

In multi-agent mode, the orchestrator assembles its system prompt from the structured prompt builder (Safety, ConversationRules, Identity, ToolUsage sections loaded from `promptsDir`). However, sub-agents only receive the hard-coded `spec.Instruction` string from `agentSpecs` in `tools.go`. This creates an inconsistency: the orchestrator follows user-configured safety/conversation rules while sub-agents do not.

The prompt builder already supports directory-based loading (`LoadFromDir`) and section replacement (`Add` with same ID). The orchestration layer already has a callback injection pattern (`ToolAdapter`).

## Goals / Non-Goals

**Goals:**
- All sub-agents SHALL inherit shared prompt sections (Safety, ConversationRules) from the prompt builder
- Users SHALL be able to override or extend prompts per individual sub-agent via `agents/<name>/` directories
- The change MUST be backward compatible — nil `SubAgentPromptFunc` preserves existing behavior
- Sub-agent prompts MUST follow priority ordering: AgentIdentity (150) < Safety (200) < ConversationRules (300)

**Non-Goals:**
- Changing the orchestrator's own prompt assembly (already works correctly)
- Adding new prompt sections beyond what the builder already supports
- Per-agent model or tool customization (out of scope)
- Dynamic prompt reloading at runtime (requires restart)

## Decisions

### 1. Clone-based builder branching (over shared mutable builder)
Each sub-agent needs an independent builder that starts from the shared base but can diverge. `Clone()` creates a shallow copy of the sections slice, allowing independent `Add`/`Remove` without affecting the original. Alternative: re-building from scratch per agent — rejected because it duplicates file I/O and doesn't share cached sections.

### 2. Callback injection via SubAgentPromptFunc (over direct prompt import)
The orchestration package cannot import `prompt` or `config` without creating import cycles (orchestration ← app → prompt). A function type `SubAgentPromptFunc` follows the existing `ToolAdapter` pattern, keeping orchestration decoupled. Alternative: passing a pre-built map of agent→prompt — rejected because it requires knowing which agents will be created before `BuildAgentTree` runs.

### 3. SectionAgentIdentity at priority 150 (over reusing SectionIdentity at 100)
Sub-agents need their role description (`spec.Instruction`) separate from the global agent identity. Priority 150 places it after the global identity (100, removed for sub-agents) but before Safety (200). This ensures the agent's role context comes first in the final prompt.

### 4. Per-agent directory convention: `agents/<name>/` (over flat file naming)
A directory-per-agent structure (`agents/operator/IDENTITY.md`) is cleaner than flat naming (`operator_IDENTITY.md`) and naturally supports multiple override files per agent. It also mirrors the existing `prompts/` directory convention.

## Risks / Trade-offs

- **[Risk] Prompt size growth**: Sub-agents now receive Safety + ConversationRules + their own instruction, increasing token usage per sub-agent call → Mitigation: These sections are typically small (< 500 tokens combined) and the benefit of consistent safety rules outweighs the cost.
- **[Risk] File system overhead on startup**: Each sub-agent checks for an `agents/<name>/` directory → Mitigation: `os.ReadDir` on non-existent directories returns immediately; no measurable impact.
- **[Trade-off] Shared base excludes Identity and ToolUsage**: Sub-agents don't inherit `AGENTS.md` (global identity) or `TOOL_USAGE.md` because these are orchestrator-specific. If a user wants tool usage rules in a sub-agent, they must add a custom `.md` file in the agent's directory.
