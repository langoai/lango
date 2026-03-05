## ADDED Requirements

### Requirement: MCP Server List View
The TUI settings editor SHALL provide a list view displaying all configured MCP servers sorted alphabetically by name. Each item SHALL show the server name, transport type, and enabled/disabled status.

#### Scenario: Display servers with details
- **WHEN** user navigates to "MCP Server List" in the settings menu
- **THEN** the system displays all servers from `cfg.MCP.Servers` as `name (transport) [enabled/disabled]`, sorted by name

#### Scenario: Empty server list
- **WHEN** no MCP servers are configured
- **THEN** the list shows only the "+ Add New MCP Server" action item

### Requirement: Add New MCP Server
The TUI SHALL allow adding a new MCP server via a form accessible from the server list. The form SHALL include a server name field (editable only for new servers) and all MCPServerConfig fields.

#### Scenario: Add new server via form
- **WHEN** user selects "+ Add New MCP Server" from the list
- **THEN** the system opens a form titled "Add New MCP Server" with an editable name field and all server configuration fields

#### Scenario: Save new server
- **WHEN** user completes the form and presses Esc
- **THEN** the system creates a new entry in `cfg.MCP.Servers[name]` with the form values and returns to the server list

### Requirement: Edit Existing MCP Server
The TUI SHALL allow editing an existing MCP server by selecting it from the list. The form SHALL pre-populate all fields from the existing configuration.

#### Scenario: Edit existing server
- **WHEN** user selects an existing server from the list
- **THEN** the system opens a form titled "Edit MCP Server: <name>" with all fields pre-populated from the server configuration

### Requirement: Delete MCP Server
The TUI SHALL allow deleting an MCP server by pressing "d" on the selected item in the list.

#### Scenario: Delete server
- **WHEN** user presses "d" on a server in the list
- **THEN** the system removes the server from `cfg.MCP.Servers` and refreshes the list

### Requirement: Transport-Conditional Fields
The server form SHALL conditionally show fields based on the selected transport type. stdio transport SHALL show command and args fields. http and sse transports SHALL show url and headers fields.

#### Scenario: stdio transport fields
- **WHEN** transport is set to "stdio"
- **THEN** the form shows Command and Args fields but hides URL and Headers fields

#### Scenario: http/sse transport fields
- **WHEN** transport is set to "http" or "sse"
- **THEN** the form shows URL and Headers fields but hides Command and Args fields

### Requirement: Map and Slice Field Serialization
Environment variables, headers, and args SHALL be serialized as comma-separated values in text input fields. Maps SHALL use `KEY=VAL,KEY=VAL` format. Slices SHALL use `val1,val2,val3` format.

#### Scenario: Parse environment variables
- **WHEN** user enters `API_KEY=secret,DEBUG=true` in the Environment field
- **THEN** the system stores `{"API_KEY": "secret", "DEBUG": "true"}` in `srv.Env`

#### Scenario: Parse args
- **WHEN** user enters `-y,@anthropic-ai/mcp-server` in the Args field
- **THEN** the system stores `["-y", "@anthropic-ai/mcp-server"]` in `srv.Args`
