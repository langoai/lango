## MODIFIED Requirements

### Requirement: Gateway-backed CLI address resolution coverage stays executable
Repository-level regressions in gateway-backed CLI address normalization SHALL be enforced by executable tests.

#### Scenario: Explicit gateway trailing slash normalization remains covered
- **WHEN** the shared CLI gateway address resolver receives an explicit address with trailing slashes
- **THEN** executable tests SHALL fail if it returns the trailing slash or causes double-slash gateway request paths

#### Scenario: Status explicit gateway display remains covered
- **WHEN** `lango status --addr <url>` probes a custom gateway
- **THEN** executable tests SHALL fail if the status output reports the configured gateway instead of the normalized explicit probe target
