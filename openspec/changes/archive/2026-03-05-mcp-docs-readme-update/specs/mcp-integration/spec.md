## MODIFIED Requirements

### Requirement: MCP documentation coverage
The MCP Plugin System SHALL have complete documentation coverage across README.md and docs/cli/ matching all other documented features.

#### Scenario: README Features list includes MCP
- **WHEN** a user reads the README.md Features section
- **THEN** MCP Integration is listed with description of stdio/HTTP/SSE transport, auto-discovery, health checks, and multi-scope config

#### Scenario: README CLI Commands section includes MCP
- **WHEN** a user reads the README.md CLI Commands section
- **THEN** all 7 `lango mcp` subcommands (list, add, remove, get, test, enable, disable) are listed with descriptions

#### Scenario: README Architecture diagram includes MCP
- **WHEN** a user reads the README.md Architecture section
- **THEN** `mcp/` appears in both the cli/ tree and the internal/ tree

#### Scenario: docs/cli/index.md Quick Reference includes MCP
- **WHEN** a user reads the CLI Quick Reference table in docs/cli/index.md
- **THEN** an "MCP Servers" section lists all 7 subcommands

#### Scenario: docs/cli/mcp.md exists with full reference
- **WHEN** a user reads docs/cli/mcp.md
- **THEN** each subcommand has argument tables, flag tables, and usage examples matching the actual CLI implementation
