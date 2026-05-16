## ADDED Requirements

### Requirement: README memory inventory keeps the `agent <name>` placeholder
The README internal CLI inventory SHALL describe the per-agent memory command with its current placeholder.

#### Scenario: Placeholder stays visible
- **WHEN** a maintainer updates the README internal tree inventory
- **THEN** it SHALL describe `lango memory list/status/clear/agents/agent <name>`
