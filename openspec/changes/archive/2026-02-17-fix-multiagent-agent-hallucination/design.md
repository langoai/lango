## Context

The multi-agent orchestrator currently has no tools of its own — all tools are partitioned exclusively to sub-agents. This forces the LLM to delegate every tool-requiring task. When the LLM cannot find a suitable sub-agent, it hallucinates agent names from tool name patterns (e.g., `browser_navigate` → `browser_agent`), causing ADK `transfer_to_agent` failures.

ADK's `llmagent.Config` supports both `Tools` and `SubAgents` simultaneously, allowing the orchestrator to call tools directly while still delegating complex tasks to sub-agents.

## Goals / Non-Goals

**Goals:**
- Eliminate agent name hallucination by giving the orchestrator direct tool access
- Reduce unnecessary delegation round-trips for simple single-tool tasks
- Maintain sub-agent specialization for complex multi-step workflows

**Non-Goals:**
- Changing the tool partitioning logic (sub-agents still get their partitioned tools)
- Adding new sub-agents or modifying sub-agent behavior
- Changing the single-agent mode code path

## Decisions

### Give orchestrator ALL tools directly

**Decision**: Pass the full `cfg.Tools` set (adapted to ADK tools) to the orchestrator's `Tools` field.

**Alternatives considered**:
1. **Prompt-only fix** — Add stronger instructions to prevent hallucination. Rejected: ADK already injects valid agent names in `transfer_to_agent` descriptions, but the LLM still hallucinates. Prompt engineering alone is not reliable.
2. **Remove sub-agents entirely** — Make multi-agent mode behave like single-agent. Rejected: Loses the benefit of specialized reasoning chains for complex tasks.
3. **Give orchestrator only "unpartitioned" tools** — Only tools not assigned to any sub-agent. Rejected: Would still force delegation for partitioned tools, not solving the core issue.

**Rationale**: Option 1 (all tools + sub-agents) is the simplest fix with maximum benefit. The orchestrator can handle simple tasks directly, and the LLM has no reason to hallucinate agent names when it can see and call tools itself.

### Update orchestrator instruction

**Decision**: Restructure the instruction to clearly separate "direct tool usage" from "sub-agent delegation" with explicit rules for when to use each.

**Rationale**: With tools available, the LLM needs guidance on when to call tools directly vs. delegate. Without this, it might never delegate (defeating the purpose of sub-agents).

## Risks / Trade-offs

- **[Risk] Duplicate tool execution** — Both orchestrator and sub-agent have the same tools, so the LLM could call a tool directly that a sub-agent is better suited for. → Mitigation: Instruction explicitly guides simple vs. complex task routing.
- **[Risk] Increased token usage** — Orchestrator tool descriptions add tokens to every request. → Mitigation: Acceptable trade-off; this is the same overhead as single-agent mode.
- **[Trade-off] Tool adaptation called twice** — Each tool is adapted once for the sub-agent and once for the orchestrator. → Acceptable: `adaptTools` is fast and only runs at startup.
