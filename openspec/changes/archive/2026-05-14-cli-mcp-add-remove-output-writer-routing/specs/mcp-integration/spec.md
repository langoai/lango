## MODIFIED Requirements

### Requirement: CLI MCP management commands
The CLI MUST provide `lango mcp add <name>` and `lango mcp remove <name>` to manage MCP server configuration entries across supported scopes.

#### Scenario: MCP add/remove output uses the command writer
- **WHEN** `lango mcp add` or `lango mcp remove` renders human-readable confirmation output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
