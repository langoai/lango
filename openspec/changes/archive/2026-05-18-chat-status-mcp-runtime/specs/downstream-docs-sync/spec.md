## ADDED Requirements

### Requirement: TUI core docs describe local runtime status boundaries
Public TUI core documentation SHALL describe which configured features may still initialize in local interactive mode and how `/status` distinguishes configured intent from active runtime state.

#### Scenario: TUI core docs describe MCP runtime status truthfully
- **WHEN** a user reads `docs/cli/core.md`
- **THEN** the docs SHALL state that local TUI `/status` reports MCP as active when the local interactive bootstrap initialized MCP
