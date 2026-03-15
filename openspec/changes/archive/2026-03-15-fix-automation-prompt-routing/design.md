## Context

The orchestrator's Decision Protocol begins with "Step 0: ASSESS" which checks whether a prompt is a simple conversational request (greeting, general knowledge, etc.) and responds directly if so. Cron, background, and workflow prompts are user-authored task strings like "search for latest news" which look conversational to the orchestrator, causing it to answer from general knowledge instead of delegating to a sub-agent with tools.

## Goals / Non-Goals

**Goals:**
- Ensure all automated prompts (cron, background, workflow) are always delegated to the correct sub-agent
- Preserve the orchestrator's ability to handle genuine conversational requests directly
- Minimal, non-breaking change — prompt enrichment only, no structural changes

**Non-Goals:**
- Changing the orchestrator's Decision Protocol logic or step ordering
- Modifying the sub-agent routing table or agent specs
- Adding new configuration flags or user-facing settings

## Decisions

### Decision 1: Prefix-based signal over metadata channel
**Choice**: Prepend a textual `[Automated Task]` prefix to prompts rather than passing metadata via context or a separate field.

**Rationale**: The AgentRunner interface accepts a plain `string` prompt. Adding a metadata field would require changing the interface across 3 packages (cron, background, workflow) and the orchestration layer. A prefix-based approach is zero-cost in terms of API surface and works immediately with the existing LLM instruction parsing.

**Alternative considered**: Context-based metadata (e.g., `context.WithValue`). Rejected because the orchestrator processes the prompt string via LLM, not programmatic metadata.

### Decision 2: Per-package constant over shared utility
**Choice**: Each package (cron, background, workflow) defines its own `automationPrefix` constant rather than importing from a shared package.

**Rationale**: Avoids introducing a new shared package or import dependency for a single constant. The three packages are independent and should remain so per the project's import cycle avoidance pattern.

### Decision 3: Orchestrator instruction section placement
**Choice**: "Automated Task Handling" section is placed immediately before the Decision Protocol so the LLM encounters it first.

**Rationale**: LLM instruction following is order-sensitive. Placing the override rule before Step 0 ensures the orchestrator checks for the `[Automated Task]` prefix before applying the general ASSESS heuristic.

## Risks / Trade-offs

- **[Risk]** User manually crafts a prompt starting with `[Automated Task` → orchestrator forces delegation even for conversational intent → **Mitigation**: The prefix is an internal implementation detail not exposed in any user-facing API; users interact via cron_add/bg_submit/workflow tools which construct prompts programmatically.
- **[Trade-off]** Prompt length increases by ~90 characters per automation call → Acceptable given typical prompt sizes of hundreds to thousands of characters.
