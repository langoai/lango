## ADDED Requirements

### Requirement: MCP settings form exists in TUI
The TUI settings editor SHALL provide an "MCP Servers" form accessible from the Infrastructure section of the settings menu.

#### Scenario: User navigates to MCP settings
- **WHEN** user opens the settings menu and selects "MCP Servers" from the Infrastructure section
- **THEN** the system SHALL display a form titled "MCP Servers Configuration" with 6 fields

### Requirement: MCP form exposes global configuration fields
The MCP form SHALL expose the following fields mapped to `config.MCPConfig`:
- `mcp_enabled` (InputBool) → `MCP.Enabled`
- `mcp_default_timeout` (InputText with duration validation) → `MCP.DefaultTimeout`
- `mcp_max_output_tokens` (InputInt with positive validation) → `MCP.MaxOutputTokens`
- `mcp_health_check_interval` (InputText with duration validation) → `MCP.HealthCheckInterval`
- `mcp_auto_reconnect` (InputBool) → `MCP.AutoReconnect`
- `mcp_max_reconnect_attempts` (InputInt with positive validation) → `MCP.MaxReconnectAttempts`

#### Scenario: Form displays current config values
- **WHEN** user opens the MCP form with existing config values
- **THEN** each field SHALL display the current value from `config.MCPConfig`

#### Scenario: Duration field validation rejects invalid input
- **WHEN** user enters an invalid duration string (e.g., "abc") in Default Timeout or Health Check Interval
- **THEN** the form SHALL display a validation error

#### Scenario: Integer field validation rejects non-positive values
- **WHEN** user enters 0 or a negative number in Max Output Tokens or Max Reconnect Attempts
- **THEN** the form SHALL display a validation error

### Requirement: MCP form saves to config state
When the user exits the MCP form (Esc), `UpdateConfigFromForm()` SHALL persist all 6 field values back to `ConfigState.Current.MCP`.

#### Scenario: Saving MCP settings updates config
- **WHEN** user modifies MCP fields and exits the form
- **THEN** `ConfigState.Current.MCP` SHALL reflect the updated values
