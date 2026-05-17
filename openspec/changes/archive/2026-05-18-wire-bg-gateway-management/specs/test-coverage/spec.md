## MODIFIED Requirements

### Requirement: Background CLI boundary guards stay executable
Repository-level regressions in background CLI boundary messaging SHALL be enforced by executable tests.

#### Scenario: Root bg gateway client coverage stays executable
- **WHEN** root CLI bg commands are wired
- **THEN** executable tests SHALL fail if root `lango bg` falls back to the obsolete standalone in-memory boundary stub
- **AND** executable tests SHALL fail if `--addr` is ignored by the gateway-backed bg client

#### Scenario: Background gateway route coverage stays executable
- **WHEN** background gateway route tests run
- **THEN** they SHALL fail if list/status/result/cancel routes are not registered
- **AND** they SHALL fail if unavailable background managers do not return `503`
- **AND** they SHALL fail if task not-found responses are not reported as non-2xx errors

#### Scenario: Background automation docs guard checks gateway wording
- **WHEN** docs quality tests run
- **THEN** README, `docs/cli/index.md`, and `docs/automation/background.md` SHALL fail the test suite if they describe root `lango bg` as disconnected from gateway management after the gateway-backed client is implemented
- **AND** they SHALL be checked for the in-memory restart caveat, `--addr` override guidance, and auth-enabled gateway rejection caveat
