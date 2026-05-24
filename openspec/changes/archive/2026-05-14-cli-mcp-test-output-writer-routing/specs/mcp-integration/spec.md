## MODIFIED Requirements

### Requirement: CLI manages MCP server lifecycle
The CLI MUST support listing, adding, removing, inspecting, testing, enabling, and disabling servers.

#### Scenario: MCP test output uses the command writer
- **WHEN** an operator runs `lango mcp test <name>`
- **THEN** the command SHALL write the full diagnostic output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
