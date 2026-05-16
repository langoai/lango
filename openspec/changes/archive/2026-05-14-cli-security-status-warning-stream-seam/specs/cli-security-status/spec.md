## MODIFIED Requirements

### Requirement: Security status output routing
`lango security status` SHALL route human-readable and JSON output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: Non-interactive status warning path is capturable
- **WHEN** non-interactive passphrase acquisition fails with an unexpected warning-worthy error during security status DB inspection
- **THEN** the warning SHALL be capturable through an injected warning writer instead of requiring process-global stderr interception
