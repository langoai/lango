## MODIFIED Requirements

### Requirement: Slash commands
The TUI SHALL support slash commands: `/help`, `/clear`, `/new`, `/model`, `/status`, `/exit`, `/quit`, `/mode`, `/cost`. The `/mode` command SHALL accept a mode name argument and update the session's mode accordingly; without an argument, it SHALL print the current mode and available modes. The `/cost` command SHALL print the session's cumulative token usage and estimated cost.

#### Scenario: /status reports active local MCP runtime
- **WHEN** the chat model is constructed with MCP configured and an active MCP runtime snapshot
- **THEN** `/status` SHALL report MCP as active in TUI mode
- **AND** it SHALL NOT report MCP as configured but not active in TUI mode

#### Scenario: /status keeps configured-only MCP distinct
- **WHEN** the chat model is constructed with MCP configured but no active MCP runtime snapshot
- **THEN** `/status` SHALL report MCP as configured without an active MCP runtime
