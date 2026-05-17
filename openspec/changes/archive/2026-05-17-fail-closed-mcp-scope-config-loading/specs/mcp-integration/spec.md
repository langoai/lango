## MODIFIED Requirements

### Requirement: Configuration

MCP integration MUST support `mcp.enabled` boolean flag (default: false), named server configs under `mcp.servers.<name>`, and named server configs loaded from profile, user (`~/.lango/mcp.json`), and project (`.lango-mcp.json`) scopes. The system MUST merge configs in the order profile < user < project. Missing user or project scoped files SHALL be treated as absent optional configuration. Present scoped files that cannot be read, parsed, or validated SHALL return an actionable error that identifies the scope and path instead of silently falling back to lower-priority configuration.

#### Scenario: Configuration merges all supported scopes
- **WHEN** MCP configuration is loaded from profile, user, and project scopes
- **THEN** the system MUST merge those scopes in the documented order
- **AND** preserve per-server transport, timeout, safety, and auth settings

#### Scenario: Missing scoped MCP config files are optional
- **WHEN** the user or project MCP config file does not exist
- **THEN** the system SHALL continue loading configuration from the remaining scopes
- **AND** it SHALL NOT return an error for the missing optional file

#### Scenario: Invalid scoped MCP config fails visibly
- **WHEN** a user or project MCP config file exists but cannot be read, parsed, or validated
- **THEN** the system SHALL return an error
- **AND** the error SHALL identify the scope and file path
- **AND** the system SHALL NOT silently fall back to lower-priority MCP configuration

#### Scenario: MCP CLI write commands preserve invalid existing files
- **WHEN** an MCP CLI write command targets a scope whose config file exists but is invalid
- **THEN** the command SHALL return an actionable config load error
- **AND** it SHALL NOT overwrite that existing file with replacement content
