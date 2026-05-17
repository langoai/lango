## ADDED Requirements

### Requirement: TUI runtime status coverage stays executable
Repository-level regressions in chat slash-command runtime status rendering SHALL be enforced by executable tests.

#### Scenario: TUI status MCP runtime coverage stays executable
- **WHEN** the chat slash-command status surface can receive an active MCP runtime snapshot
- **THEN** executable chat tests SHALL fail if `/status` still labels MCP as configured but inactive in TUI mode
- **AND** executable chat tests SHALL fail if configured-only MCP is indistinguishable from active MCP
