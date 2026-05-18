## ADDED Requirements

### Requirement: Slack HTTP client timeout
The Slack channel SHALL use a bounded default HTTP client for Slack REST/API operations when no custom HTTP client is provided. Custom HTTP clients supplied through `Config.HTTPClient` SHALL be preserved unchanged.

#### Scenario: Default Slack HTTP client is bounded
- **WHEN** the Slack channel is created without `Config.HTTPClient`
- **THEN** the Slack SDK client SHALL use an HTTP client with a finite timeout
- **AND** the timeout SHALL be at least 10 seconds

#### Scenario: Custom Slack HTTP client is preserved
- **WHEN** the Slack channel is created with `Config.HTTPClient`
- **THEN** the provided HTTP client SHALL be used unchanged
