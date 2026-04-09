## Purpose

Capability spec for agent-prompting. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: System prompt construction
The system SHALL construct the system prompt using a structured `prompt.Builder` instead of a single string. The `ContextAwareModelAdapter` constructor SHALL accept a `*prompt.Builder` and call `Build()` to produce the base prompt string. Dynamic context injection (knowledge, memory, RAG) SHALL continue to append to the built prompt at runtime.

#### Scenario: Prepend system prompt to new session
- **WHEN** a new agent session is started
- **THEN** the system SHALL prepend a message with `role: system` containing the configured prompt to the conversation history

#### Scenario: Default identity prompt
- **WHEN** no custom system prompt is provided
- **THEN** a default prompt describing the agent's identity and tools SHALL be used

#### Scenario: Builder produces base prompt
- **WHEN** ContextAwareModelAdapter is created with a prompt.Builder
- **THEN** the basePrompt field SHALL equal the builder's Build() output

#### Scenario: Dynamic context still appended
- **WHEN** knowledge retrieval returns context during GenerateContent
- **THEN** the retrieved context SHALL be appended to the builder-produced base prompt

### Requirement: SAFETY prompt reflects PII detection scope
The SAFETY.md prompt SHALL enumerate specific PII categories (email, phone numbers, national IDs, financial account numbers) and mention 13 builtin patterns. The prompt SHALL reference Presidio NER-based detection as additional coverage when enabled.

#### Scenario: Agent responds to PII-related user question
- **WHEN** the agent processes SAFETY.md prompt during system prompt assembly
- **THEN** the agent understands it protects 13 builtin PII pattern categories
- **THEN** the agent can accurately inform users about PII protection coverage including Presidio

### Requirement: Tool selection priority in prompts
The TOOL_USAGE.md prompt SHALL include a "Tool Selection Priority" section that instructs agents to always prefer built-in tools over skills. The section SHALL state that skills wrapping `lango` CLI commands will fail due to passphrase authentication requirements in agent mode.

#### Scenario: Agent reads tool usage prompt
- **WHEN** the agent processes TOOL_USAGE.md during system prompt assembly
- **THEN** the prompt SHALL contain a "Tool Selection Priority" section before the "Exec Tool" section

#### Scenario: Agent encounters a skill with built-in equivalent
- **WHEN** a skill provides functionality already available as a built-in tool
- **THEN** the prompt guidance SHALL direct the agent to use the built-in tool instead

#### Scenario: Approval failure guidance for browser actions
- **WHEN** the agent processes browser guidance in TOOL_USAGE.md or navigator-specific instructions
- **THEN** it SHALL be instructed not to immediately reissue the same browser action after approval denial or expiry
- **AND** it SHALL instead explain the approval issue or choose a lower-risk alternative only when that alternative materially changes the action

#### Scenario: Browser search fallback guidance
- **WHEN** the agent processes the browser section of TOOL_USAGE.md
- **THEN** it SHALL be instructed to prefer `browser_search` for open-ended live web queries
- **AND** it SHALL also be instructed to fall back to `browser_navigate` with a search URL plus `browser_extract` in `search_results` mode when `browser_search` is unavailable
- **AND** it SHALL be instructed to continue with low-level `browser_action` or `eval` instead of stopping if equivalent browser tools are still available

#### Scenario: Bounded browser search guidance
- **WHEN** the agent processes browser guidance in TOOL_USAGE.md
- **THEN** it SHALL use imperative language: "ONCE", "EXACTLY once", "NEVER more than twice"
- **AND** it SHALL instruct to call `browser_search` ONCE and present results if `resultCount > 0` without searching again
- **AND** it SHALL allow reformulation EXACTLY once when `resultCount == 0` or results are clearly unrelated
- **AND** it SHALL state "NEVER call browser_search more than twice per request"
- **AND** it SHALL instruct to stop once the requested number of credible results has been collected

#### Scenario: Navigator bounded search protocol
- **WHEN** the navigator handles a topic-based live web request
- **THEN** its instruction SHALL use a "Search Workflow (MANDATORY)" section with imperative language
- **AND** it SHALL direct the agent to call `browser_search` ONCE with its best query
- **AND** it SHALL state that `resultCount > 0` means results are available and the agent MUST NOT search again
- **AND** it SHALL allow EXACTLY one reformulation when `resultCount == 0`
- **AND** it SHALL state "NEVER call browser_search more than twice per request. There are no exceptions."

### Requirement: Tool selection directive in agent identity
The AGENTS.md prompt SHALL include a tool selection directive stating that built-in tools MUST be preferred over skills, and skills are extensions for specialized use cases only.

#### Scenario: Agent reads identity prompt
- **WHEN** the agent processes AGENTS.md during system prompt assembly
- **THEN** the prompt SHALL contain a tool selection directive before the knowledge system description

### Requirement: Runtime skill priority note
The `AssemblePrompt()` method in `ContextRetriever` SHALL prepend a note to the "Available Skills" section advising agents to prefer built-in tools over skills.

#### Scenario: Skills section rendered with priority note
- **WHEN** the assembled prompt includes skill pattern items
- **THEN** the "Available Skills" section SHALL begin with a note stating to prefer built-in tools over skills
