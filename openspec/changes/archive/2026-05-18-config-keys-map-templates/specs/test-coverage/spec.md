## MODIFIED Requirements

### Requirement: Config CLI behavior coverage stays executable
Repository-level regressions in config get/set/keys behavior SHALL be enforced by executable tests.

#### Scenario: Dynamic config key templates remain listed
- **WHEN** the repository still supports map-backed config set paths for providers, MCP server env/header values, and auth providers
- **THEN** executable config CLI tests SHALL fail if `collectKeys` or `lango config keys <prefix>` omits the corresponding dynamic templates
- **AND** the tests SHALL fail if dynamic templates include unsupported `time.Duration` leaves such as `mcp.servers.<name>.timeout`
