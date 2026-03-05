## MODIFIED Requirements

### Requirement: MCP Menu Entry
The TUI settings menu SHALL display two separate entries for MCP configuration under the Infrastructure section: "MCP Settings" for global MCP configuration and "MCP Server List" for per-server CRUD management.

#### Scenario: Menu displays both MCP entries
- **WHEN** user views the Infrastructure section in the settings menu
- **THEN** the menu shows "MCP Settings" (ID: `mcp`) with description "Global MCP server settings" and "MCP Server List" (ID: `mcp_servers`) with description "Add, edit, remove MCP servers"
