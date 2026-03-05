## MODIFIED Requirements

### Requirement: MCP configuration is editable via TUI
The MCP integration SHALL be configurable through both CLI commands and the TUI settings editor. Global settings (enabled, timeouts, reconnection) SHALL be available in the TUI settings form. Individual server management (add/remove/enable/disable) SHALL remain CLI-only via `lango mcp` subcommands.

#### Scenario: Global MCP settings accessible in TUI
- **WHEN** user opens TUI settings and navigates to Infrastructure > MCP Servers
- **THEN** the system SHALL display a form for editing global MCP configuration fields

#### Scenario: Server management remains CLI-only
- **WHEN** user needs to add, remove, enable, or disable individual MCP servers
- **THEN** user SHALL use `lango mcp add/remove/enable/disable` CLI commands
