## MODIFIED Requirements

### Requirement: Default values
The configuration system SHALL apply sensible defaults for all non-credential fields. The minimum viable configuration SHALL require only: `agent.provider`, `providers.<name>.type`, `providers.<name>.apiKey`, and one channel's `enabled: true` + token. All other fields SHALL have defaults:
- `server.host`: `"localhost"`
- `server.port`: `18789`
- `server.httpEnabled`: `true`
- `server.wsEnabled`: `true`
- `session.databasePath`: `"~/.lango/data.db"`
- `session.maxHistoryTurns`: `100`
- `logging.level`: `"info"`
- `logging.format`: `"console"`
- `agent.maxTokens`: `4096`
- `agent.temperature`: `0.7`
- `tools.exec.defaultTimeout`: `30s`
- `tools.exec.allowBackground`: `true`
- `tools.filesystem.maxReadSize`: `1048576` (1MB)
- `tools.browser.headless`: `true`
- `tools.browser.sessionTimeout`: `5m`
- `librarian.enabled`: `false`
- `librarian.observationThreshold`: `2`
- `librarian.inquiryCooldownTurns`: `3`
- `librarian.maxPendingInquiries`: `2`
- `librarian.autoSaveConfidence`: `"high"`

#### Scenario: Missing optional field
- **WHEN** a configuration field is not specified
- **THEN** the system SHALL use the default value listed above
- **THEN** no error or warning SHALL be emitted for missing optional fields

#### Scenario: Minimal configuration startup
- **WHEN** config contains only `agent.provider`, one provider entry with `type` and `apiKey`, and one channel with `enabled: true` and token
- **THEN** the application SHALL start successfully with all defaults applied

#### Scenario: Librarian defaults applied
- **WHEN** the `librarian` section is omitted from configuration
- **THEN** the system SHALL apply default values: enabled=false, observationThreshold=2, inquiryCooldownTurns=3, maxPendingInquiries=2, autoSaveConfidence="high"
