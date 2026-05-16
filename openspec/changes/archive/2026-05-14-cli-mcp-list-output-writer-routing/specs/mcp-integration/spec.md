## MODIFIED Requirements

### Requirement: CLI manages MCP server lifecycle
The CLI MUST support listing, adding, removing, inspecting, testing, enabling, and disabling servers.

#### Scenario: MCP list output uses the command writer
- **WHEN** an operator runs `lango mcp list`
- **THEN** the command SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
