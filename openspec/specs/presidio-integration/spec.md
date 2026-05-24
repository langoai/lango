# presidio-integration Specification

## Purpose
Define Microsoft Presidio integration for supplemental PII detection, scoring, and category mapping.
## Requirements
### Requirement: Presidio HTTP client
PresidioDetector SHALL call the Microsoft Presidio analyzer's POST /analyze endpoint to detect PII entities.

#### Scenario: Successful detection
- **WHEN** Presidio returns entity results with scores above threshold
- **THEN** PresidioDetector SHALL map each result to a PIIMatch with pattern name "presidio:<entity_type>"

#### Scenario: Entity category mapping
- **WHEN** Presidio returns entity type "EMAIL_ADDRESS"
- **THEN** the match category SHALL be "contact"

#### Scenario: Unknown entity type
- **WHEN** Presidio returns an unrecognized entity type
- **THEN** the match category SHALL default to "identity"

### Requirement: Graceful degradation
PresidioDetector SHALL return nil (no matches) on any error, ensuring regex-based detection continues as fallback.

#### Scenario: Server unreachable
- **WHEN** Presidio endpoint is not reachable
- **THEN** Detect SHALL return nil without panicking

#### Scenario: Server error response
- **WHEN** Presidio returns HTTP 500
- **THEN** Detect SHALL return nil

#### Scenario: Invalid response JSON
- **WHEN** Presidio returns invalid JSON
- **THEN** Detect SHALL return nil

### Requirement: Functional options
PresidioDetector SHALL support configuration via functional options: WithPresidioThreshold, WithPresidioLanguage, WithPresidioTimeout.

#### Scenario: Custom language
- **WHEN** WithPresidioLanguage("ko") is applied
- **THEN** the analyze request SHALL include language="ko"

#### Scenario: Custom threshold
- **WHEN** WithPresidioThreshold(0.9) is applied
- **THEN** the analyze request SHALL include score_threshold=0.9

### Requirement: Health check
PresidioDetector SHALL provide a HealthCheck(ctx) error method that verifies the /health endpoint returns HTTP 200.

#### Scenario: Healthy service
- **WHEN** Presidio /health returns 200
- **THEN** HealthCheck SHALL return nil

#### Scenario: Unhealthy service
- **WHEN** Presidio /health returns non-200
- **THEN** HealthCheck SHALL return an error
