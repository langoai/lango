## MODIFIED Requirements

### Requirement: Navigator bounded search protocol

The navigator agent's instruction SHALL use imperative language for the search workflow. The instruction SHALL direct the agent to call `browser_search` ONCE, present results if `resultCount > 0` without searching again, allow EXACTLY one reformulation when `resultCount == 0`, and NEVER call `browser_search` more than twice per request.

#### Scenario: Navigator instruction uses mandatory language

- **WHEN** the navigator agent receives its system instruction
- **THEN** the search workflow section SHALL use "MUST", "NEVER", "Do NOT" instead of "Prefer", "before considering", "You may"

#### Scenario: Navigator instruction caps searches at two

- **WHEN** the navigator agent processes its search workflow rules
- **THEN** the instruction SHALL state "NEVER call browser_search more than twice per request" with no exceptions

### Requirement: TOOL_USAGE browser search guidance

The shared `TOOL_USAGE.md` browser section SHALL use imperative language consistent with the navigator instruction. It SHALL state that `browser_search` must be called ONCE, reformulation is allowed EXACTLY once when `resultCount == 0`, and searching more than twice per request is prohibited.

#### Scenario: TOOL_USAGE uses imperative search guidance

- **WHEN** an agent reads the Browser Tool section of TOOL_USAGE.md
- **THEN** the guidance SHALL use "ONCE", "EXACTLY once", "NEVER more than twice" instead of "Prefer", "before considering"
