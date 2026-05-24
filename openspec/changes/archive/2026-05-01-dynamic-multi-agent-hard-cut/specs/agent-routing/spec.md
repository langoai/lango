## MODIFIED Requirements

### Requirement: Prompt override file consistency
All prompt override files (`IDENTITY.md`, `AGENT.md`) SHALL continue to forbid `[REJECT]` patterns. Non-planner prompt override files SHALL continue to include a `## Output Handling` section with `tool_output_get` guidance. The hard cut narrows only the built-in escalation path: embedded prompt override files for built-in teammates SHALL NOT require `transfer_to_agent("lango-orchestrator")` as the built-in escalation path, while remote/legacy transfer guidance may remain only where explicitly documented as compatibility behavior.

#### Scenario: No REJECT patterns
- **WHEN** any prompt override file is checked
- **THEN** it SHALL NOT contain the text `[REJECT]`

#### Scenario: Built-in AGENT.md files no longer encode built-in transfer escalation
- **WHEN** any embedded built-in `AGENT.md` file is checked
- **THEN** it SHALL NOT instruct the built-in teammate to call `transfer_to_agent("lango-orchestrator")`

#### Scenario: Output handling in non-planner overrides remains required
- **WHEN** a non-planner prompt override file is checked
- **THEN** it SHALL contain `## Output Handling` section with `tool_output_get` guidance
